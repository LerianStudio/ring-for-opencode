/**
 * Ring Config Handler
 *
 * Creates a config hook that injects Ring agents, skills, and commands
 * into OpenCode's configuration at runtime.
 *
 * Pattern from oh-my-opencode:
 * config: async (opencodeConfig) => {
 *   // Modify opencodeConfig.agent, .skill, .command
 *   return opencodeConfig
 * }
 */

import { loadRingAgents } from "../loaders/agent-loader.js"
import { loadRingSkills } from "../loaders/skill-loader.js"
import { loadRingCommands } from "../loaders/command-loader.js"
import type { RingConfig } from "./schema.js"

/**
 * OpenCode config structure (subset used by Ring).
 */
export interface OpenCodeConfig {
  /** Agent configurations */
  agent?: Record<string, unknown>

  /** Command configurations */
  command?: Record<string, unknown>

  /** Permission settings */
  permission?: Record<string, string>

  /** Tools configuration */
  tools?: Record<string, boolean>

  /** Model configuration */
  model?: string
}

/**
 * Dependencies for creating the config handler.
 */
export interface ConfigHandlerDeps {
  /** Project root directory */
  projectRoot: string
  /** Ring plugin configuration */
  ringConfig: RingConfig
}

/**
 * Create the config handler that injects Ring components.
 *
 * This handler is called by OpenCode to modify the configuration
 * before the session starts. We use this to inject:
 * - Ring agents (16 agents from .opencode/agent/)
 * - Ring skills (30 skills from .opencode/skill/)
 * - Ring commands (16 commands from .opencode/command/)
 */
export function createConfigHandler(deps: ConfigHandlerDeps) {
  const { projectRoot, ringConfig } = deps

  return async (config: OpenCodeConfig): Promise<void> => {
    const debug = process.env.DEBUG === "true" || process.env.RING_DEBUG === "true"

    // Load Ring agents
    const ringAgents = loadRingAgents(
      projectRoot,
      ringConfig.disabled_agents
    )

    if (debug) {
      const agentNames = Object.keys(ringAgents)
      console.debug(`[ring] Loaded ${agentNames.length} agents:`, agentNames.slice(0, 5).join(", "), agentNames.length > 5 ? "..." : "")
    }

    // Load Ring skills (as commands)
    const ringSkills = loadRingSkills(
      projectRoot,
      ringConfig.disabled_skills
    )

    if (debug) {
      const skillNames = Object.keys(ringSkills)
      console.debug(`[ring] Loaded ${skillNames.length} skills:`, skillNames.slice(0, 5).join(", "), skillNames.length > 5 ? "..." : "")
    }

    // Load Ring commands
    const ringCommands = loadRingCommands(
      projectRoot,
      ringConfig.disabled_commands
    )

    if (debug) {
      const commandNames = Object.keys(ringCommands)
      console.debug(`[ring] Loaded ${commandNames.length} commands:`, commandNames.slice(0, 5).join(", "), commandNames.length > 5 ? "..." : "")
    }

    // Inject agents into config
    // Ring agents are added with lower priority (spread first, then existing)
    // so project-specific overrides can take precedence
    // TODO(review): Consider deep merge for nested agent configs
    config.agent = {
      ...ringAgents,
      ...(config.agent ?? {}),
    }

    // Inject skills and commands
    // Commands and skills both go into config.command
    config.command = {
      ...ringSkills,
      ...ringCommands,
      ...(config.command ?? {}),
    }

    // Set permissions for Ring tools
    config.permission = {
      ...(config.permission ?? {}),
      webfetch: "allow",
      external_directory: "allow",
    }

    // Disable recursive agent calls in certain agents
    const agentConfig = config.agent as Record<string, { tools?: Record<string, boolean> }>

    // Prevent explore agents from using task recursively
    if (agentConfig["codebase-explorer"]) {
      agentConfig["codebase-explorer"].tools = {
        ...agentConfig["codebase-explorer"].tools,
        task: false,
      }
    }

    // Prevent reviewers from spawning more reviewers
    const reviewerAgents = [
      "code-reviewer",
      "security-reviewer",
      "business-logic-reviewer",
      "test-reviewer",
      "nil-safety-reviewer",
    ]

    for (const reviewerName of reviewerAgents) {
      if (agentConfig[reviewerName]) {
        agentConfig[reviewerName].tools = {
          ...agentConfig[reviewerName].tools,
          task: false,
        }
      }
    }

    if (debug) {
      console.debug("[ring] Config injection complete")
    }
  }
}
