import type { Plugin } from "@opencode-ai/plugin"
import { existsSync, readFileSync, readdirSync, statSync, lstatSync } from "fs"
import { join } from "path"
import { cleanupOldState, deleteState, getSessionId, escapeAngleBrackets, sanitizeForPrompt, isPathWithinRoot } from "./utils/state"
import { EVENTS } from "./utils/events"

/**
 * Ring Session Start Plugin (Enhanced)
 * Comprehensive context injection at session start.
 *
 * Equivalent to: default/hooks/session-start.sh
 * Event: session.created
 *
 * Features:
 * 1. Injects critical rules (3-file rule, agent triggers)
 * 2. Injects doubt-triggered questions pattern
 * 3. Loads and injects active continuity ledger
 * 4. Generates skills quick reference
 * 5. Resets context usage state
 * 6. Cleans up old state files
 */

// Critical rules that MUST survive compact
const CRITICAL_RULES = `## ORCHESTRATOR CRITICAL RULES (SURVIVE COMPACT)

**3-FILE RULE: HARD GATE**
DO NOT read/edit >3 files directly. This is a PROHIBITION.
- >3 files -> STOP. Launch specialist agent. DO NOT proceed manually.
- Already touched 3 files? -> At gate. Dispatch agent NOW.

**AUTO-TRIGGER PHRASES -> MANDATORY AGENT:**
- "fix issues/remaining/findings" -> Launch specialist agent
- "apply fixes", "fix the X issues" -> Launch specialist agent
- "find where", "search for", "understand how" -> Launch Explore agent

**If you think "this task is small" or "I can handle 5 files":**
WRONG. Count > 3 = agent. No exceptions. Task size is irrelevant.

**Full rules:** Use skill tool with "ring-default:using-ring" if needed.`

// Doubt-triggered questions pattern
const DOUBT_QUESTIONS = `## DOUBT-TRIGGERED QUESTIONS (WHEN TO ASK)

**Resolution hierarchy (check BEFORE asking):**
1. User dispatch context -> Did they already specify?
2. CLAUDE.md / repo conventions -> Is there a standard?
3. Codebase patterns -> What does existing code do?
4. Best practice -> Is one approach clearly superior?
5. **-> ASK** -> Only if ALL above fail AND affects correctness

**Genuine doubt criteria (ALL must be true):**
- Cannot resolve from hierarchy above
- Multiple approaches genuinely viable
- Choice significantly impacts correctness
- Getting it wrong wastes substantial effort

**Question quality - show your work:**
- GOOD: "Found PostgreSQL in docker-compose but MongoDB in docs.
    This feature needs time-series. Which should I extend?"
- BAD: "Which database should I use?"

**If proceeding without asking:**
State assumption -> Explain why -> Note what would change it`

// Agent usage reminder (compact version)
const AGENT_REMINDER = `## AGENT USAGE REMINDER

CONTEXT CHECK: Before using Glob/Grep/Read chains, consider agents:

| Task | Agent |
|------|-------|
| Explore codebase | Explore |
| Multi-file search | Explore |
| Complex research | general-purpose |
| Code review | code-reviewer + business-logic-reviewer + security-reviewer (PARALLEL) |
| Implementation plan | write-plan |
| Deep architecture | codebase-explorer |

**3-File Rule:** If reading >3 files, use an agent instead. 15x more context-efficient.`

// Duplication prevention
const DUPLICATION_GUARD = `## DUPLICATION PREVENTION

**BEFORE ADDING CONTENT** to any file:
1. SEARCH FIRST: \`grep -r 'keyword' --include='*.md'\`
2. If exists -> REFERENCE it, don't copy
3. Canonical sources: CLAUDE.md (rules), docs/*.md (details)
4. NEVER duplicate - always link to single source of truth`

/**
 * Ensure Python dependencies are available.
 */
type ShellExec = PromiseLike<unknown> & {
  quiet: () => Promise<unknown>
  text: () => Promise<string>
}

type Shell = (strings: TemplateStringsArray, ...values: unknown[]) => ShellExec

const ensurePythonDeps = async ($: Shell): Promise<boolean> => {
  try {
    await $`python3 -c "import yaml"`.quiet()
    return true
  } catch {
    try {
      await $`pip3 install --quiet --user 'PyYAML>=6.0'`
      return true
    } catch (err) {
      return false
    }
  }
}

export const RingSessionStart: Plugin = async ({ directory, $ }) => {
  const projectRoot = directory
  const sessionId = getSessionId()

  /**
   * Find and read the active continuity ledger.
   */
  const findActiveLedger = (): { name: string; content: string; currentPhase: string } | null => {
    const ledgerDir = join(projectRoot, ".ring/ledgers")

    if (!existsSync(ledgerDir)) {
      return null
    }

    try {
      const files = readdirSync(ledgerDir)
        .filter((f) => f.startsWith("CONTINUITY-") && f.endsWith(".md"))
        .map((f) => ({
          name: f,
          path: join(ledgerDir, f),
          mtime: statSync(join(ledgerDir, f)).mtimeMs,
          isSymlink: lstatSync(join(ledgerDir, f)).isSymbolicLink(),
        }))
        .filter((f) => !f.isSymlink)
        .sort((a, b) => b.mtime - a.mtime)

      if (files.length === 0) return null

      const newest = files[0]
      const content = readFileSync(newest.path, "utf-8")

      // Extract current phase (line with [->] marker)
      const phaseMatch = content.match(/\[->/m)
      const currentPhase = phaseMatch ? content.split("\n").find((line) => line.includes("[->"))?.trim() || "" : ""

      return {
        name: newest.name.replace(".md", ""),
        content,
        currentPhase,
      }
    } catch (err) {
      return null
    }
  }

  /**
   * Generate skills quick reference.
   * Tries Python script first, then TypeScript fallback.
   */
  const generateSkillsOverview = async (): Promise<string> => {
    // Ensure Python dependencies first
    await ensurePythonDeps($)

    // Try to find and run the Python generator
    const pythonGenerators = [
      join(projectRoot, "default/hooks/generate-skills-ref.py"),
      join(projectRoot, ".ring/hooks/generate-skills-ref.py"),
    ]

    for (const generator of pythonGenerators) {
      if (!existsSync(generator)) continue
      if (!isPathWithinRoot(generator, projectRoot)) continue

      try {
        const st = lstatSync(generator)
        if (st.isSymbolicLink()) continue
      } catch {
        continue
      }

      try {
        const result = await $`python3 ${[generator]}`.text()
        return result
      } catch (err) {
        continue
      }
    }

    // Fallback: Generate minimal reference from .opencode/skill/
    const skillDir = join(projectRoot, ".opencode/skill")
    if (!existsSync(skillDir)) {
      return `# Ring Skills Quick Reference

**Note:** Skills directory not found.
Skills are available via the skill() tool.

Run: \`skill: "using-ring"\` to see available workflows.`
    }

    try {
      const skills = readdirSync(skillDir)
        .filter((d) => statSync(join(skillDir, d)).isDirectory())
        .sort()

      let overview = "# Ring Skills Quick Reference\n\n"
      overview += `**${skills.length} skills available**\n\n`

      for (const skill of skills.slice(0, 20)) {
        overview += `- **${skill}**\n`
      }

      if (skills.length > 20) {
        overview += `\n...and ${skills.length - 20} more\n`
      }

      overview += "\n## Usage\n"
      overview += "To use a skill: Use the skill() tool with skill name\n"
      overview += 'Example: `skill: "using-ring"`\n'

      return overview
    } catch (err) {
      return "# Ring Skills Quick Reference\n\n**Error loading skills.**\n"
    }
  }

  // Cache for skills overview (generated once per session)
  let cachedSkillsOverview: string | null = null

  return {
    // Handle session lifecycle events
    event: async ({ event }) => {
      if (event.type === EVENTS.SESSION_CREATED) {
        // Reset context usage state for new session
        deleteState(projectRoot, "context-usage", sessionId)
        // Clean up old state files (> 7 days)
        cleanupOldState(projectRoot, 7)
        // Clear skills cache for new session
        cachedSkillsOverview = null
      }
    },

    // Inject Ring context into the SYSTEM PROMPT (runs before each LLM call)
    "experimental.chat.system.transform": async (_input: {}, output: { system: string[] }) => {
      // Generate skills overview (cached after first call)
      if (!cachedSkillsOverview) {
        cachedSkillsOverview = await generateSkillsOverview()
      }

      // Find active ledger
      const ledger = findActiveLedger()

      // Build context to inject into system prompt
      let context = `<ring-critical-rules>
${CRITICAL_RULES}
</ring-critical-rules>

<ring-agent-reminder>
${AGENT_REMINDER}
</ring-agent-reminder>

<ring-duplication-guard>
${DUPLICATION_GUARD}
</ring-duplication-guard>

<ring-doubt-questions>
${DOUBT_QUESTIONS}
</ring-doubt-questions>

<ring-skills-system>
${cachedSkillsOverview}
</ring-skills-system>`

      // Add ledger context if present
      if (ledger) {
        const safeLedgerName = sanitizeForPrompt(ledger.name, 120)
        const safePhase = sanitizeForPrompt(ledger.currentPhase || "No active phase marked", 200)
        const safeLedgerContent = escapeAngleBrackets(ledger.content)

        context += `

<ring-continuity-ledger>
## Active Continuity Ledger: ${safeLedgerName}

**Current Phase:** ${safePhase}

<continuity-ledger-content>
${safeLedgerContent}
</continuity-ledger-content>

**Instructions:**
1. Review the State section - find [->] for current work
2. Check Open Questions for UNCONFIRMED items
3. Continue from where you left off
</ring-continuity-ledger>`
      }

      // Append to system prompt array
      output.system.push(context)
    },
  }
}

export default RingSessionStart
