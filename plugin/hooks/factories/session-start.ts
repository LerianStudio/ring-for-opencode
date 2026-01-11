/**
 * Session Start Hook Factory
 *
 * Injects critical context at session initialization and chat params.
 * Handles continuity ledger loading and system prompt injection.
 */

import * as fs from "node:fs"
import * as path from "node:path"
import type { Hook, HookContext, HookFactory, HookOutput, HookResult } from "../types.js"
import { findMostRecentFile, isPathWithinRoot, readFileSafe, sanitizeForPrompt } from "../../utils/state.js"

/**
 * Configuration for session start hook.
 */
export interface SessionStartConfig {
  /** Enable critical rules injection */
  injectCriticalRules?: boolean
  /** Enable agent reminder injection */
  injectAgentReminder?: boolean
  /** Enable duplication guard */
  injectDuplicationGuard?: boolean
  /** Enable doubt questions */
  injectDoubtQuestions?: boolean
  /** Enable continuity ledger */
  loadContinuityLedger?: boolean
  /** Max ledger content length */
  maxLedgerLength?: number
}

/** Default configuration */
const DEFAULT_CONFIG: Required<SessionStartConfig> = {
  injectCriticalRules: true,
  injectAgentReminder: true,
  injectDuplicationGuard: true,
  injectDoubtQuestions: true,
  loadContinuityLedger: true,
  maxLedgerLength: 8000,
}

/**
 * Critical rules that must be followed in every session.
 */
const CRITICAL_RULES = `
<ring-critical-rules>
## Critical Rules (NON-NEGOTIABLE)

1. **NEVER skip verification** - Always verify changes work before claiming completion
2. **NEVER ignore test failures** - Fix failing tests, don't comment them out
3. **NEVER commit secrets** - Check for API keys, passwords, tokens before committing
4. **NEVER make assumptions** - If unclear, ask for clarification
5. **ALWAYS follow project conventions** - Match existing code style and patterns
6. **ALWAYS preserve working code** - Don't break existing functionality
7. **ALWAYS use TodoWrite** - Track multi-step tasks with the todo list
</ring-critical-rules>
`.trim()

/**
 * Agent reminder for maintaining quality.
 */
const AGENT_REMINDER = `
<ring-agent-reminder>
## Quality Checklist

Before completing any task:
- [ ] Code compiles without errors
- [ ] Tests pass (if applicable)
- [ ] No console.log/print debugging left behind
- [ ] Error handling is in place
- [ ] Changes are minimal and focused
</ring-agent-reminder>
`.trim()

/**
 * Duplication guard to prevent redundant work.
 */
const DUPLICATION_GUARD = `
<ring-duplication-guard>
## Avoid Duplication

Before implementing:
1. Check if similar functionality already exists
2. Reuse existing utilities and helpers
3. Don't reinvent patterns that are already established
4. Reference existing code for style consistency
</ring-duplication-guard>
`.trim()

/**
 * Doubt questions to resolve ambiguity.
 */
const DOUBT_QUESTIONS = `
<ring-doubt-resolver>
## When In Doubt

Ask yourself:
1. Is this requirement clear enough to implement?
2. Are there edge cases I should clarify?
3. Does this conflict with existing behavior?
4. Should I confirm the approach before proceeding?

If ANY answer is "no" or "maybe" - ASK THE USER before proceeding.
</ring-doubt-resolver>
`.trim()

/**
 * Find active continuity ledger from .ring/ledgers directory.
 * Includes path validation and symlink checking for security.
 */
function findActiveLedger(directory: string): string | null {
  const ledgersDir = path.join(directory, ".ring", "ledgers")

  // Find most recent markdown file in ledgers directory
  const ledgerPath = findMostRecentFile(ledgersDir, /\.md$/)

  if (!ledgerPath) {
    return null
  }

  // Validate path is within project root (prevent path traversal)
  if (!isPathWithinRoot(ledgerPath, directory)) {
    return null
  }

  // Check for symlinks and validate target
  try {
    const stats = fs.lstatSync(ledgerPath)
    if (stats.isSymbolicLink()) {
      const realPath = fs.realpathSync(ledgerPath)
      if (!isPathWithinRoot(realPath, directory)) {
        return null
      }
    }
  } catch {
    return null
  }

  return readFileSafe(ledgerPath)
}

/**
 * Format ledger content for injection.
 */
function formatLedgerContent(content: string, maxLength: number): string {
  const sanitized = sanitizeForPrompt(content, maxLength)

  return `
<ring-continuity-ledger>
## Active Continuity Ledger

The following ledger contains context from previous sessions:

${sanitized}
</ring-continuity-ledger>
`.trim()
}

/**
 * Create a session start hook.
 */
export const createSessionStartHook: HookFactory<SessionStartConfig> = (
  config?: SessionStartConfig
): Hook => {
  const cfg = { ...DEFAULT_CONFIG, ...config }

  return {
    name: "session-start",
    lifecycles: ["session.created", "chat.params"],
    priority: 10, // Run early
    enabled: true,

    async execute(ctx: HookContext, output: HookOutput): Promise<HookResult> {
      const systemInjections: string[] = []

      try {
        // Inject critical rules
        if (cfg.injectCriticalRules) {
          systemInjections.push(CRITICAL_RULES)
        }

        // Inject agent reminder
        if (cfg.injectAgentReminder) {
          systemInjections.push(AGENT_REMINDER)
        }

        // Inject duplication guard
        if (cfg.injectDuplicationGuard) {
          systemInjections.push(DUPLICATION_GUARD)
        }

        // Inject doubt questions
        if (cfg.injectDoubtQuestions) {
          systemInjections.push(DOUBT_QUESTIONS)
        }

        // Load continuity ledger (cache result to avoid duplicate lookup)
        let ledgerContent: string | null = null
        if (cfg.loadContinuityLedger) {
          ledgerContent = findActiveLedger(ctx.directory)

          if (ledgerContent) {
            const formattedLedger = formatLedgerContent(
              ledgerContent,
              cfg.maxLedgerLength
            )
            systemInjections.push(formattedLedger)
          }
        }

        // Add to output
        if (systemInjections.length > 0) {
          output.system = output.system ?? []
          output.system.push(...systemInjections)
        }

        return {
          success: true,
          data: {
            injectionsCount: systemInjections.length,
            hasLedger: cfg.loadContinuityLedger && ledgerContent !== null,
          },
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error)
        return {
          success: false,
          error: `Session start hook failed: ${errorMessage}`,
        }
      }
    },
  }
}

/**
 * Hook registry entry for session start.
 */
export const sessionStartEntry = {
  name: "session-start" as const,
  factory: createSessionStartHook,
  defaultEnabled: true,
  description: "Injects critical rules, reminders, and continuity ledger at session start",
}
