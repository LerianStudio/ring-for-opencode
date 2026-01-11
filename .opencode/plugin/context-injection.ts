import type { Plugin } from "@opencode-ai/plugin"
import { existsSync, readFileSync, readdirSync, statSync } from "fs"
import { join } from "path"

/**
 * Ring Context Injection Plugin (Enhanced)
 * Injects comprehensive Ring context during session compaction.
 *
 * Equivalent to: default/hooks/session-start.sh (compaction portion)
 * Event: experimental.session.compacting
 *
 * When OpenCode compacts session history for context management,
 * this plugin ensures critical Ring rules, skill information,
 * and continuity ledger state are preserved and re-injected.
 */

// Critical rules - abbreviated for compaction context
const CRITICAL_RULES_COMPACT = `**3-FILE RULE: HARD GATE** - >3 files = dispatch agent, no exceptions.
**AUTO-TRIGGER:** "fix issues" -> agent, "find where" -> Explore agent.
**TDD:** Test must fail (RED) before implementation.
**REVIEW:** All 5 reviewers must pass before merge.
**COMMIT:** Use /commit command for intelligent grouping.`

export const RingContextInjection: Plugin = async ({ directory }) => {
  const projectRoot = directory

  /**
   * Find and summarize active continuity ledger.
   */
  const getLedgerSummary = (): string | null => {
    const ledgerDir = join(projectRoot, ".ring/ledgers")

    if (!existsSync(ledgerDir)) {
      return null
    }

    try {
      const files = readdirSync(ledgerDir)
        .filter((f) => f.startsWith("CONTINUITY-") && f.endsWith(".md"))
        .map((f) => ({
          name: f.replace(".md", ""),
          path: join(ledgerDir, f),
          mtime: statSync(join(ledgerDir, f)).mtimeMs,
        }))
        .sort((a, b) => b.mtime - a.mtime)

      if (files.length === 0) return null

      const ledger = files[0]
      const content = readFileSync(ledger.path, "utf-8")

      // Extract current phase
      const phaseLine = content.split("\n").find((line) => line.includes("[->"))
      const currentPhase = phaseLine?.trim() || "No active phase"

      // Extract key sections for summary
      const sections: string[] = []

      // Get BRIEF_SUMMARY if present
      const briefMatch = content.match(/## BRIEF_SUMMARY\n([^#]+)/s)
      if (briefMatch) {
        sections.push(briefMatch[1].trim().slice(0, 200))
      }

      // Get UNCONFIRMED items from open questions
      const openQMatch = content.match(/## Open Questions \(UNCONFIRMED\)\n([^#]+)/s)
      if (openQMatch && openQMatch[1].includes("[ ]")) {
        const unconfirmed = openQMatch[1]
          .split("\n")
          .filter((line) => line.includes("[ ]"))
          .slice(0, 3)
          .join("\n")
        if (unconfirmed) {
          sections.push(`**Open Questions:**\n${unconfirmed}`)
        }
      }

      return `## Active Ledger: ${ledger.name}
**Current Phase:** ${currentPhase}
${sections.join("\n\n")}

**Full ledger at:** .ring/ledgers/${ledger.name}.md`
    } catch (err) {
      return null
    }
  }

  return {
    "experimental.session.compacting": async (input, output) => {
      // Defensive check for output.context
      if (!output?.context || !Array.isArray(output.context)) {
        return
      }

      // Inject critical rules (compact version)
      output.context.push(`## Ring Critical Rules (Compact)
${CRITICAL_RULES_COMPACT}`)

      // Inject skills reference
      output.context.push(`## Ring Skills System

Key skills for common tasks:
- **test-driven-development**: TDD methodology (RED -> GREEN -> REFACTOR)
- **requesting-code-review**: Parallel review with 5 specialized reviewers
- **systematic-debugging**: 4-phase debugging methodology
- **writing-plans**: Create detailed implementation plans
- **executing-plans**: Execute plans in batches with review checkpoints
- **dispatching-parallel-agents**: Run multiple agents concurrently
- **verification-before-completion**: Ensure work complete before done

## Ring Commands

- /commit - Intelligent commit grouping
- /codereview - Dispatch 5 parallel reviewers
- /brainstorm - Socratic design refinement
- /execute-plan - Batch execution with checkpoints
- /worktree - Create isolated development branch
- /lint - Run and fix lint issues
- /create-handoff - Create session handoff document

## Ring Agents

- code-reviewer - Technical code quality review
- security-reviewer - Security vulnerability analysis
- business-logic-reviewer - Business requirements alignment
- codebase-explorer - Autonomous codebase exploration
- write-plan - Implementation planning agent`)

      // Inject ledger summary if present
      const ledgerSummary = getLedgerSummary()
      if (ledgerSummary) {
        output.context.push(ledgerSummary)
      }

    },
  }
}

export default RingContextInjection
