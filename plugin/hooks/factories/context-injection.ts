/**
 * Context Injection Hook Factory
 *
 * Handles context injection during session compaction.
 * Provides compact versions of rules, skills, commands, and agents references.
 */

import * as fs from "node:fs"
import * as path from "node:path"
import { fileURLToPath } from "node:url"

import { type AgentConfig, loadRingAgents } from "../../loaders/agent-loader.js"
import { loadRingCommands } from "../../loaders/command-loader.js"
import { loadRingSkills, type SkillConfig } from "../../loaders/skill-loader.js"
import type { Hook, HookContext, HookFactory, HookOutput, HookResult } from "../types.js"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

/** Plugin root directory (contains assets/) */
const PLUGIN_ROOT = path.resolve(__dirname, "../../..")

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
 * Generate skills reference from loaded skills.
 * Format: "Available Ring skills: /skill1, /skill2, ... Use 'Skill' tool with skill name to invoke."
 */
function generateSkillsReference(skills: Record<string, SkillConfig>): string {
  const skillNames = Object.keys(skills)
    .map((key) => key.replace("ring:", ""))
    .sort()
    .map((name) => `/${name}`)
    .join(", ")

  if (!skillNames) return ""
  return `Available Ring skills: ${skillNames}\nUse "Skill" tool with skill name to invoke.`
}

/**
 * Generate commands reference from loaded commands.
 * Format: "Ring commands: /cmd1, /cmd2, ... Use skills via Skill tool."
 */
function generateCommandsReference(commands: Record<string, unknown>): string {
  const cmdNames = Object.keys(commands)
    .map((key) => key.replace("ring:", ""))
    .sort()

  if (cmdNames.length === 0) return ""
  return `Ring commands: Use skills via Skill tool. Check /help for full list.\nKey patterns: TDD (test-first), systematic debugging, defense-in-depth validation.`
}

/**
 * Generate agents reference from loaded agents.
 * Format categorizes agents by role prefix for compact representation.
 */
function generateAgentsReference(agents: Record<string, AgentConfig>): string {
  const agentNames = Object.keys(agents)
    .map((key) => key.replace("ring:", ""))
    .sort()

  if (agentNames.length === 0) return ""

  // Categorize agents by common prefixes
  const devAgents: string[] = []
  const reviewers: string[] = []
  const pmAgents: string[] = []
  const otherAgents: string[] = []

  for (const name of agentNames) {
    if (name.includes("reviewer")) {
      reviewers.push(name.replace("-reviewer", ""))
    } else if (name.startsWith("pre-dev")) {
      pmAgents.push(name)
    } else if (
      name.includes("engineer") ||
      name === "devops" ||
      name === "sre" ||
      name === "qa-analyst" ||
      name === "frontend-designer"
    ) {
      devAgents.push(name)
    } else {
      otherAgents.push(name)
    }
  }

  const parts: string[] = []
  if (devAgents.length > 0) parts.push(`Dev agents: ${devAgents.join(", ")}`)
  if (reviewers.length > 0) parts.push(`Reviewers: ${reviewers.join(", ")}`)
  if (pmAgents.length > 0) parts.push(`PM agents: ${pmAgents.join(", ")}`)
  if (otherAgents.length > 0) parts.push(`Other: ${otherAgents.join(", ")}`)

  return `${parts.join("\n")}\nDispatch via dev-cycle or pre-dev workflows.`
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

// Load static prompt content (compact-rules.txt is still static)
const COMPACT_CRITICAL_RULES_CONTENT = loadPrompt("context-injection/compact-rules.txt")

/**
 * Compact critical rules for compaction context.
 */
const COMPACT_CRITICAL_RULES = COMPACT_CRITICAL_RULES_CONTENT
  ? `<ring-compact-rules>
${COMPACT_CRITICAL_RULES_CONTENT}
</ring-compact-rules>`
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

    async execute(ctx: HookContext, output: HookOutput): Promise<HookResult> {
      const contextInjections: string[] = []
      const projectRoot = ctx.directory

      try {
        // Inject compact critical rules
        if (cfg.injectCompactRules && COMPACT_CRITICAL_RULES) {
          contextInjections.push(COMPACT_CRITICAL_RULES)
        }

        // Generate and inject skills reference dynamically
        if (cfg.injectSkillsRef) {
          const skills = loadRingSkills(PLUGIN_ROOT, projectRoot)
          const skillsContent = generateSkillsReference(skills)
          if (skillsContent) {
            contextInjections.push(`<ring-skills-ref>\n${skillsContent}\n</ring-skills-ref>`)
          }
        }

        // Generate and inject commands reference dynamically
        if (cfg.injectCommandsRef) {
          const { commands } = loadRingCommands(PLUGIN_ROOT, projectRoot)
          const commandsContent = generateCommandsReference(commands)
          if (commandsContent) {
            contextInjections.push(`<ring-commands-ref>\n${commandsContent}\n</ring-commands-ref>`)
          }
        }

        // Generate and inject agents reference dynamically
        if (cfg.injectAgentsRef) {
          const agents = loadRingAgents(PLUGIN_ROOT, projectRoot)
          const agentsContent = generateAgentsReference(agents)
          if (agentsContent) {
            contextInjections.push(`<ring-agents-ref>\n${agentsContent}\n</ring-agents-ref>`)
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
