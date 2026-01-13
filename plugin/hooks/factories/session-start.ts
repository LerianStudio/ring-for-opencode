/**
 * Session Start Hook Factory
 *
 * Injects critical context at session initialization and chat params.
 */

import * as fs from "node:fs"
import * as path from "node:path"
import { fileURLToPath } from "node:url"

import type { Hook, HookContext, HookFactory, HookOutput, HookResult } from "../types.js"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

/**
 * Load a prompt file from the assets/prompts directory.
 */
function loadPrompt(filename: string): string {
  const promptPath = path.resolve(__dirname, "../../../assets/prompts", filename)
  try {
    return fs.readFileSync(promptPath, "utf-8").trim()
  } catch {
    return ""
  }
}

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
  /** Enable path context injection */
  injectPathContext?: boolean
}

/** Default configuration */
const DEFAULT_CONFIG: Required<SessionStartConfig> = {
  injectCriticalRules: true,
  injectAgentReminder: true,
  injectDuplicationGuard: true,
  injectDoubtQuestions: true,
  injectPathContext: true,
}

// Load prompt content from external files
const CRITICAL_RULES_CONTENT = loadPrompt("session-start/critical-rules.txt")
const AGENT_REMINDER_CONTENT = loadPrompt("session-start/agent-reminder.txt")
const DUPLICATION_GUARD_CONTENT = loadPrompt("session-start/duplication-guard.txt")
const DOUBT_QUESTIONS_CONTENT = loadPrompt("session-start/doubt-questions.txt")
const PATH_CONTEXT_CONTENT = loadPrompt("session-start/path-context.txt")

/**
 * Critical rules that must be followed in every session.
 */
const CRITICAL_RULES = CRITICAL_RULES_CONTENT
  ? `<ring-critical-rules>
${CRITICAL_RULES_CONTENT}
</ring-critical-rules>`
  : ""

/**
 * Agent reminder for maintaining quality.
 */
const AGENT_REMINDER = AGENT_REMINDER_CONTENT
  ? `<ring-agent-reminder>
${AGENT_REMINDER_CONTENT}
</ring-agent-reminder>`
  : ""

/**
 * Duplication guard to prevent redundant work.
 */
const DUPLICATION_GUARD = DUPLICATION_GUARD_CONTENT
  ? `<ring-duplication-guard>
${DUPLICATION_GUARD_CONTENT}
</ring-duplication-guard>`
  : ""

/**
 * Doubt questions to resolve ambiguity.
 */
const DOUBT_QUESTIONS = DOUBT_QUESTIONS_CONTENT
  ? `<ring-doubt-resolver>
${DOUBT_QUESTIONS_CONTENT}
</ring-doubt-resolver>`
  : ""

/**
 * Path context for OpenCode directory structure.
 */
const PATH_CONTEXT = PATH_CONTEXT_CONTENT
  ? `<ring-paths>
${PATH_CONTEXT_CONTENT}
</ring-paths>`
  : ""

/**
 * Create a session start hook.
 */
export const createSessionStartHook: HookFactory<SessionStartConfig> = (
  config?: SessionStartConfig,
): Hook => {
  const cfg = { ...DEFAULT_CONFIG, ...config }

  return {
    name: "session-start",
    lifecycles: ["session.created", "chat.params"],
    priority: 10, // Run early
    enabled: true,

    async execute(_ctx: HookContext, output: HookOutput): Promise<HookResult> {
      const systemInjections: string[] = []

      try {
        // Inject critical rules
        if (cfg.injectCriticalRules && CRITICAL_RULES) {
          systemInjections.push(CRITICAL_RULES)
        }

        // Inject agent reminder
        if (cfg.injectAgentReminder && AGENT_REMINDER) {
          systemInjections.push(AGENT_REMINDER)
        }

        // Inject duplication guard
        if (cfg.injectDuplicationGuard && DUPLICATION_GUARD) {
          systemInjections.push(DUPLICATION_GUARD)
        }

        // Inject doubt questions
        if (cfg.injectDoubtQuestions && DOUBT_QUESTIONS) {
          systemInjections.push(DOUBT_QUESTIONS)
        }

        // Inject path context
        if (cfg.injectPathContext && PATH_CONTEXT) {
          systemInjections.push(PATH_CONTEXT)
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
  description: "Injects critical rules and reminders at session start",
}
