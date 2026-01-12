/**
 * Context Injection Hook Factory
 *
 * Handles context injection during session compaction.
 * Provides compact versions of rules, skills, commands, and agents references.
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
}

/** Default configuration */
const DEFAULT_CONFIG: Required<ContextInjectionConfig> = {
  injectCompactRules: true,
  injectSkillsRef: true,
  injectCommandsRef: true,
  injectAgentsRef: true,
}

// Load prompt content from external files
const COMPACT_CRITICAL_RULES_CONTENT = loadPrompt("compact-rules.txt")
const SKILLS_REFERENCE_CONTENT = loadPrompt("skills-reference.txt")
const COMMANDS_REFERENCE_CONTENT = loadPrompt("commands-reference.txt")
const AGENTS_REFERENCE_CONTENT = loadPrompt("agents-reference.txt")

/**
 * Compact critical rules for compaction context.
 */
const COMPACT_CRITICAL_RULES = COMPACT_CRITICAL_RULES_CONTENT
  ? `<ring-compact-rules>
${COMPACT_CRITICAL_RULES_CONTENT}
</ring-compact-rules>`
  : ""

/**
 * Skills reference for compaction context.
 */
const SKILLS_REFERENCE = SKILLS_REFERENCE_CONTENT
  ? `<ring-skills-ref>
${SKILLS_REFERENCE_CONTENT}
</ring-skills-ref>`
  : ""

/**
 * Commands reference for compaction context.
 */
const COMMANDS_REFERENCE = COMMANDS_REFERENCE_CONTENT
  ? `<ring-commands-ref>
${COMMANDS_REFERENCE_CONTENT}
</ring-commands-ref>`
  : ""

/**
 * Agents reference for compaction context.
 */
const AGENTS_REFERENCE = AGENTS_REFERENCE_CONTENT
  ? `<ring-agents-ref>
${AGENTS_REFERENCE_CONTENT}
</ring-agents-ref>`
  : ""

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

    async execute(_ctx: HookContext, output: HookOutput): Promise<HookResult> {
      const contextInjections: string[] = []

      try {
        // Inject compact critical rules
        if (cfg.injectCompactRules && COMPACT_CRITICAL_RULES) {
          contextInjections.push(COMPACT_CRITICAL_RULES)
        }

        // Inject skills reference
        if (cfg.injectSkillsRef && SKILLS_REFERENCE) {
          contextInjections.push(SKILLS_REFERENCE)
        }

        // Inject commands reference
        if (cfg.injectCommandsRef && COMMANDS_REFERENCE) {
          contextInjections.push(COMMANDS_REFERENCE)
        }

        // Inject agents reference
        if (cfg.injectAgentsRef && AGENTS_REFERENCE) {
          contextInjections.push(AGENTS_REFERENCE)
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
