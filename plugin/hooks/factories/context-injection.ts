/**
 * Context Injection Hook Factory
 *
 * Handles context injection during session compaction.
 * Provides compact versions of rules, skills, commands, and agents references.
 */

import * as fs from "node:fs"
import * as path from "node:path"
import { isPathWithinRoot, readFileSafe, sanitizeForPrompt } from "../../utils/state.js"
import type { Hook, HookContext, HookFactory, HookOutput, HookResult } from "../types.js"

/**
 * Configuration for context injection hook.
 */
export interface ContextInjectionConfig {
  /** Enable compact critical rules */
  injectCompactRules?: boolean
  /** Enable skills reference */
  injectSkillsRef?: boolean
  /** Enable commands reference */
  injectCommandsRef?: boolean
  /** Enable agents reference */
  injectAgentsRef?: boolean
  /** Enable ledger summary */
  injectLedgerSummary?: boolean
  /** Max ledger summary length */
  maxLedgerSummaryLength?: number
}

/** Default configuration */
const DEFAULT_CONFIG: Required<ContextInjectionConfig> = {
  injectCompactRules: true,
  injectSkillsRef: true,
  injectCommandsRef: true,
  injectAgentsRef: true,
  injectLedgerSummary: true,
  maxLedgerSummaryLength: 2000,
}

/**
 * Compact critical rules for compaction context.
 */
const COMPACT_CRITICAL_RULES = `
<ring-compact-rules>
CRITICAL: verify changes | fix failing tests | no secrets | ask if unclear | match conventions | preserve working code | use TodoWrite
</ring-compact-rules>
`.trim()

/**
 * Skills reference for compaction context.
 */
const SKILLS_REFERENCE = `
<ring-skills-ref>
Available Ring skills: /commit, /review-pr, /explore-codebase, /write-plan, /execute-plan, /brainstorm, /interview-me, /create-handoff, /resume-handoff, /lint, /codereview, /dev-cycle
Use "Skill" tool with skill name to invoke.
</ring-skills-ref>
`.trim()

/**
 * Commands reference for compaction context.
 */
const COMMANDS_REFERENCE = `
<ring-commands-ref>
Ring commands: Use skills via Skill tool. Check /help for full list.
Key patterns: TDD (test-first), systematic debugging, defense-in-depth validation.
</ring-commands-ref>
`.trim()

/**
 * Agents reference for compaction context.
 */
const AGENTS_REFERENCE = `
<ring-agents-ref>
Dev agents: backend-go, backend-ts, frontend-ts, devops, sre, qa, designer
PM agents: pre-dev-full (large features), pre-dev-feature (small features)
Dispatch via dev-cycle or pre-dev workflows.
</ring-agents-ref>
`.trim()

/**
 * Find most recent continuity ledger, with path validation.
 *
 * SECURITY: Mirrors the root/symlink checks from session-start.
 */
function findActiveLedgerPath(directory: string): string | null {
  const ledgersDir = path.join(directory, ".ring", "ledgers")

  // Choose the most recent *valid* ledger, rather than failing the whole
  // summary injection when the newest entry is an invalid symlink.
  let candidates: Array<{ filePath: string; mtime: number }> = []

  try {
    if (!fs.existsSync(ledgersDir)) {
      return null
    }

    const entries = fs.readdirSync(ledgersDir)

    for (const entry of entries) {
      if (!/\.md$/.test(entry)) continue

      const filePath = path.join(ledgersDir, entry)

      try {
        const stats = fs.lstatSync(filePath)
        candidates.push({ filePath, mtime: stats.mtimeMs })
      } catch {
        // Ignore unreadable entries
      }
    }
  } catch {
    return null
  }

  candidates.sort((a, b) => b.mtime - a.mtime)

  for (const { filePath } of candidates) {
    // Validate path is within project root (prevent path traversal)
    if (!isPathWithinRoot(filePath, directory)) {
      continue
    }

    // Check for symlinks and validate target
    try {
      const stats = fs.lstatSync(filePath)
      if (stats.isSymbolicLink()) {
        const realPath = fs.realpathSync(filePath)
        if (!isPathWithinRoot(realPath, directory)) {
          continue
        }
      }
    } catch {
      continue
    }

    return filePath
  }

  return null
}

/**
 * Find and summarize continuity ledger for compaction.
 */
function getLedgerSummary(directory: string, maxLength: number): string | null {
  const ledgerPath = findActiveLedgerPath(directory)

  if (!ledgerPath) {
    return null
  }

  const content = readFileSafe(ledgerPath)
  if (!content) {
    return null
  }

  // Extract key sections for summary
  const sections: string[] = []

  // Extract current task if present
  const taskMatch = content.match(/##\s*Current\s*Task[:\s]*\n([\s\S]*?)(?=\n##|\n$|$)/i)
  if (taskMatch) {
    const task = taskMatch[1].trim().split("\n")[0]
    if (task) {
      sections.push(`Task: ${task}`)
    }
  }

  // Extract key decisions if present
  const decisionsMatch = content.match(/##\s*(?:Key\s*)?Decisions[:\s]*\n([\s\S]*?)(?=\n##|\n$|$)/i)
  if (decisionsMatch) {
    const decisions = decisionsMatch[1]
      .trim()
      .split("\n")
      .filter((line) => line.trim().startsWith("-") || line.trim().startsWith("*"))
      .slice(0, 3)
      .map((line) => line.trim())
      .join("; ")
    if (decisions) {
      sections.push(`Decisions: ${decisions}`)
    }
  }

  // Extract blockers if present
  const blockersMatch = content.match(/##\s*Blockers?[:\s]*\n([\s\S]*?)(?=\n##|\n$|$)/i)
  if (blockersMatch) {
    const blockers = blockersMatch[1].trim().split("\n")[0]
    if (blockers && blockers.toLowerCase() !== "none") {
      sections.push(`Blocker: ${blockers}`)
    }
  }

  if (sections.length === 0) {
    // Fallback: first meaningful line
    const firstLine = content.split("\n").find((line) => line.trim() && !line.startsWith("#"))
    if (firstLine) {
      sections.push(firstLine.trim())
    }
  }

  const summary = sections.join(" | ")
  return sanitizeForPrompt(summary, maxLength)
}

/**
 * Format ledger summary for injection.
 */
function formatLedgerSummary(summary: string): string {
  return `
<ring-ledger-summary>
Ledger: ${summary}
</ring-ledger-summary>
`.trim()
}

/**
 * Create a context injection hook.
 */
export const createContextInjectionHook: HookFactory<ContextInjectionConfig> = (
  config?: ContextInjectionConfig,
): Hook => {
  const cfg = { ...DEFAULT_CONFIG, ...config }

  return {
    name: "context-injection",
    lifecycles: ["session.compacting"],
    priority: 20,
    enabled: true,

    async execute(ctx: HookContext, output: HookOutput): Promise<HookResult> {
      const contextInjections: string[] = []

      try {
        // Inject compact critical rules
        if (cfg.injectCompactRules) {
          contextInjections.push(COMPACT_CRITICAL_RULES)
        }

        // Inject skills reference
        if (cfg.injectSkillsRef) {
          contextInjections.push(SKILLS_REFERENCE)
        }

        // Inject commands reference
        if (cfg.injectCommandsRef) {
          contextInjections.push(COMMANDS_REFERENCE)
        }

        // Inject agents reference
        if (cfg.injectAgentsRef) {
          contextInjections.push(AGENTS_REFERENCE)
        }

        // Inject ledger summary
        if (cfg.injectLedgerSummary) {
          const summary = getLedgerSummary(ctx.directory, cfg.maxLedgerSummaryLength)
          if (summary) {
            contextInjections.push(formatLedgerSummary(summary))
          }
        }

        // Add to output context
        if (contextInjections.length > 0) {
          output.context = output.context ?? []
          output.context.push(...contextInjections)
        }

        return {
          success: true,
          data: {
            injectionsCount: contextInjections.length,
          },
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error)
        return {
          success: false,
          error: `Context injection hook failed: ${errorMessage}`,
        }
      }
    },
  }
}

/**
 * Hook registry entry for context injection.
 */
export const contextInjectionEntry = {
  name: "context-injection" as const,
  factory: createContextInjectionHook,
  defaultEnabled: true,
  description: "Injects compact rules, skills, commands, and agents references during compaction",
}
