/**
 * Ring Agent Loader
 *
 * Loads Ring agents from:
 * 1. Plugin's assets/agent/*.md files (Ring's built-in agents)
 * 2. User's .opencode/agent/*.md files (user customizations)
 *
 * User's agents take priority over Ring's built-in agents.
 */

import { existsSync, readdirSync, readFileSync } from "node:fs"
import { basename, join } from "node:path"

/**
 * Agent configuration compatible with OpenCode SDK.
 */
export interface AgentConfig {
  description?: string
  mode?: "primary" | "subagent"
  prompt?: string
  model?: string
  temperature?: number
  tools?: Record<string, boolean>
  permission?: Record<string, string>
  color?: string
}

/**
 * Frontmatter data from agent markdown files.
 */
interface AgentFrontmatter {
  description?: string
  mode?: string
  model?: string
  temperature?: number
  tools?: string
  color?: string
}

/**
 * Keys that must never be used as object properties when building maps from filenames.
 */
const FORBIDDEN_OBJECT_KEYS = new Set(["__proto__", "constructor", "prototype"])

function isForbiddenObjectKey(key: string): boolean {
  return FORBIDDEN_OBJECT_KEYS.has(key)
}

/**
 * Parse YAML frontmatter from markdown content.
 */
function parseFrontmatter(content: string): { data: AgentFrontmatter; body: string } {
  // Normalize line endings for cross-platform support
  const normalizedContent = content.replace(/\r\n/g, "\n")
  const frontmatterRegex = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/
  const match = normalizedContent.match(frontmatterRegex)

  if (!match) {
    return { data: {}, body: normalizedContent }
  }

  const yamlContent = match[1]
  const body = match[2]

  // Simple YAML parsing for our use case
  const data: AgentFrontmatter = {}
  const lines = yamlContent.split("\n")

  for (const line of lines) {
    const colonIndex = line.indexOf(":")
    if (colonIndex === -1) continue

    const key = line.slice(0, colonIndex).trim()
    let value = line.slice(colonIndex + 1).trim()

    // Remove quotes if present
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1)
    }

    if (key === "description") data.description = value
    if (key === "mode") data.mode = value
    if (key === "model") data.model = value
    if (key === "temperature") {
      const parsed = parseFloat(value)
      if (!isNaN(parsed)) data.temperature = parsed
    }
    if (key === "tools") data.tools = value
    if (key === "color") data.color = value
  }

  return { data, body }
}

/**
 * Parse tools string into tools config object.
 */
function parseToolsConfig(toolsStr?: string): Record<string, boolean> | undefined {
  if (!toolsStr) return undefined

  const tools = toolsStr
    .split(",")
    .map((t) => t.trim())
    .filter(Boolean)
  if (tools.length === 0) return undefined

  const result: Record<string, boolean> = {}
  for (const tool of tools) {
    // Handle negation (e.g., "!task" means task: false)
    if (tool.startsWith("!")) {
      result[tool.slice(1).toLowerCase()] = false
    } else {
      result[tool.toLowerCase()] = true
    }
  }
  return result
}

/**
 * Get default temperature based on agent name pattern.
 * Role-based defaults ensure consistent behavior across agent types.
 */
function getDefaultTemperature(agentName: string): number {
  const name = agentName.toLowerCase()

  // Reviewers: precise, consistent analysis (0.1)
  if (name.includes("reviewer")) {
    return 0.1
  }

  // Ops roles: precise, consistent operations (0.1)
  if (name === "devops-engineer" || name === "sre" || name === "qa-analyst") {
    return 0.1
  }

  // Planners/explorers: balanced creativity with structure (0.2)
  if (name === "write-plan" || name === "codebase-explorer") {
    return 0.2
  }

  // Engineers: balanced creativity with precision (0.2)
  if (
    name === "backend-engineer-golang" ||
    name === "backend-engineer-typescript" ||
    name === "frontend-engineer" ||
    name === "frontend-bff-engineer-typescript"
  ) {
    return 0.2
  }

  // Creative roles: higher creativity (0.4)
  if (name === "frontend-designer") {
    return 0.4
  }

  // Default fallback
  return 0.2
}

/**
 * Load agents from a directory.
 */
function loadAgentsFromDir(
  agentsDir: string,
  disabledAgents: Set<string>,
): Record<string, AgentConfig> {
  if (!existsSync(agentsDir)) {
    return {}
  }

  const result: Record<string, AgentConfig> = Object.create(null)

  try {
    const entries = readdirSync(agentsDir, { withFileTypes: true })

    for (const entry of entries) {
      if (!entry.isFile() || !entry.name.endsWith(".md")) continue

      const agentPath = join(agentsDir, entry.name)
      const agentName = basename(entry.name, ".md")

      // SECURITY: Skip forbidden gadget keys
      if (isForbiddenObjectKey(agentName)) continue

      // Skip disabled agents
      if (disabledAgents.has(agentName)) continue

      try {
        const content = readFileSync(agentPath, "utf-8")
        const { data, body } = parseFrontmatter(content)

        // Validate mode before casting
        const validModes = ["primary", "subagent"] as const
        const mode = validModes.includes(data.mode as (typeof validModes)[number])
          ? (data.mode as "primary" | "subagent")
          : "subagent"

        const config: AgentConfig = {
          description: data.description
            ? `(ring) ${data.description}`
            : `(ring) ${agentName} agent`,
          mode,
          prompt: body.trim(),
        }

        if (data.model) {
          config.model = data.model
        }

        if (data.color) {
          config.color = data.color
        }

        // Apply temperature: explicit frontmatter value or role-based default
        config.temperature = data.temperature ?? getDefaultTemperature(agentName)

        const toolsConfig = parseToolsConfig(data.tools)
        if (toolsConfig) {
          config.tools = toolsConfig
        }

        // Use ring namespace for agents
        result[`ring:${agentName}`] = config
      } catch (error) {
        if (process.env.RING_DEBUG === "true") {
          console.debug(`[ring] Failed to parse ${agentPath}:`, error)
        }
      }
    }
  } catch (error) {
    if (process.env.RING_DEBUG === "true") {
      console.debug(`[ring] Failed to read agents directory:`, error)
    }
    return {}
  }

  return result
}

/**
 * Load Ring agents from both plugin's assets/ and user's .opencode/ directories.
 *
 * @param pluginRoot - Path to the plugin directory (contains assets/)
 * @param projectRoot - Path to the user's project directory (contains .opencode/)
 * @param disabledAgents - List of agent names to skip
 * @returns Merged agent configs with user's taking priority
 */
export function loadRingAgents(
  pluginRoot: string,
  projectRoot: string,
  disabledAgents: string[] = [],
): Record<string, AgentConfig> {
  const disabledSet = new Set(disabledAgents)

  // Load Ring's built-in agents from assets/agent/
  const builtInDir = join(pluginRoot, "assets", "agent")
  const builtInAgents = loadAgentsFromDir(builtInDir, disabledSet)

  // Load user's custom agents from .opencode/agent/
  const userDir = join(projectRoot, ".opencode", "agent")
  const userAgents = loadAgentsFromDir(userDir, disabledSet)

  // Merge with user's taking priority, using a null-prototype map
  const merged: Record<string, AgentConfig> = Object.create(null)
  Object.assign(merged, builtInAgents)
  Object.assign(merged, userAgents)
  return merged
}

/**
 * Get count of available agents from both plugin's assets/ and user's .opencode/.
 */
export function countRingAgents(pluginRoot: string, projectRoot: string): number {
  const uniqueAgents = new Set<string>()

  // Count built-in agents
  const builtInDir = join(pluginRoot, "assets", "agent")
  if (existsSync(builtInDir)) {
    try {
      const entries = readdirSync(builtInDir)
      for (const f of entries) {
        if (f.endsWith(".md")) uniqueAgents.add(f)
      }
    } catch {
      // Ignore errors
    }
  }

  // Count user agents
  const userDir = join(projectRoot, ".opencode", "agent")
  if (existsSync(userDir)) {
    try {
      const entries = readdirSync(userDir)
      for (const f of entries) {
        if (f.endsWith(".md")) uniqueAgents.add(f)
      }
    } catch {
      // Ignore errors
    }
  }

  return uniqueAgents.size
}
