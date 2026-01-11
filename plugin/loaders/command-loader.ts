/**
 * Ring Command Loader
 *
 * Loads Ring commands from:
 * 1. Plugin's assets/command/*.md files (Ring's built-in commands)
 * 2. User's .opencode/command/*.md files (user customizations)
 *
 * User's commands take priority over Ring's built-in commands.
 */

import { existsSync, readdirSync, readFileSync } from "node:fs"
import { basename, join } from "node:path"

/**
 * Command configuration compatible with OpenCode SDK.
 */
export interface CommandConfig {
  description?: string
  agent?: string
  subtask?: boolean
}

/**
 * Frontmatter data from command markdown files.
 */
interface CommandFrontmatter {
  description?: string
  agent?: string
  subtask?: string | boolean
}

/**
 * Keys that must never be used as object properties when building maps from filenames.
 * Prevents prototype pollution via "__proto__.md", etc.
 */
const FORBIDDEN_OBJECT_KEYS = new Set(["__proto__", "constructor", "prototype"])

function isForbiddenObjectKey(key: string): boolean {
  return FORBIDDEN_OBJECT_KEYS.has(key)
}

/**
 * Parse YAML frontmatter from markdown content.
 */
function parseFrontmatter(content: string): { data: CommandFrontmatter; body: string } {
  // Normalize line endings for cross-platform support
  const normalizedContent = content.replace(/\r\n/g, "\n")
  const frontmatterRegex = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/
  const match = normalizedContent.match(frontmatterRegex)

  if (!match) {
    return { data: {}, body: normalizedContent }
  }

  const yamlContent = match[1]
  const body = match[2]

  const data: CommandFrontmatter = {}
  const lines = yamlContent.split("\n")

  for (const line of lines) {
    const colonIndex = line.indexOf(":")
    if (colonIndex === -1) continue

    const key = line.slice(0, colonIndex).trim()
    let value = line.slice(colonIndex + 1).trim()

    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1)
    }

    if (key === "description") data.description = value
    if (key === "agent") data.agent = value
    if (key === "subtask") {
      data.subtask = value === "true" || value === "false" ? value === "true" : value
    }
  }

  return { data, body }
}

/**
 * Load commands from a directory.
 */
function loadCommandsFromDir(
  commandsDir: string,
  disabledCommands: Set<string>,
): Record<string, CommandConfig> {
  if (!existsSync(commandsDir)) {
    return {}
  }

  const result: Record<string, CommandConfig> = Object.create(null)

  try {
    const entries = readdirSync(commandsDir, { withFileTypes: true })

    for (const entry of entries) {
      if (!entry.isFile() || !entry.name.endsWith(".md")) continue

      const commandPath = join(commandsDir, entry.name)
      const commandName = basename(entry.name, ".md")

      // SECURITY: Skip forbidden gadget keys
      if (isForbiddenObjectKey(commandName)) continue

      // Skip disabled commands
      if (disabledCommands.has(commandName)) continue

      try {
        const content = readFileSync(commandPath, "utf-8")
        const { data } = parseFrontmatter(content)

        const config: CommandConfig = {
          description: data.description || `Ring command: ${commandName}`,
        }

        if (data.agent) {
          config.agent = data.agent
        }

        if (typeof data.subtask === "boolean") {
          config.subtask = data.subtask
        }

        // Use ring namespace for commands
        result[`ring:${commandName}`] = config
      } catch (error) {
        if (process.env.RING_DEBUG === "true") {
          console.debug(`[ring] Failed to parse ${commandPath}:`, error)
        }
      }
    }
  } catch (error) {
    if (process.env.RING_DEBUG === "true") {
      console.debug(`[ring] Failed to read commands directory:`, error)
    }
    return {}
  }

  return result
}

/**
 * Load Ring commands from both plugin's assets/ and user's .opencode/ directories.
 *
 * @param pluginRoot - Path to the plugin directory (contains assets/)
 * @param projectRoot - Path to the user's project directory (contains .opencode/)
 * @param disabledCommands - List of command names to skip
 * @returns Merged command configs with user's taking priority
 */
export function loadRingCommands(
  pluginRoot: string,
  projectRoot: string,
  disabledCommands: string[] = [],
): Record<string, CommandConfig> {
  const disabledSet = new Set(disabledCommands)

  // Load Ring's built-in commands from assets/command/
  const builtInDir = join(pluginRoot, "assets", "command")
  const builtInCommands = loadCommandsFromDir(builtInDir, disabledSet)

  // Load user's custom commands from .opencode/command/
  const userDir = join(projectRoot, ".opencode", "command")
  const userCommands = loadCommandsFromDir(userDir, disabledSet)

  // Merge with user's taking priority, using a null-prototype map
  const merged: Record<string, CommandConfig> = Object.create(null)
  Object.assign(merged, builtInCommands)
  Object.assign(merged, userCommands)
  return merged
}

/**
 * Get count of available commands from both plugin's assets/ and user's .opencode/.
 */
export function countRingCommands(pluginRoot: string, projectRoot: string): number {
  const uniqueCommands = new Set<string>()

  // Count built-in commands
  const builtInDir = join(pluginRoot, "assets", "command")
  if (existsSync(builtInDir)) {
    try {
      const entries = readdirSync(builtInDir)
      for (const f of entries) {
        if (f.endsWith(".md")) uniqueCommands.add(f)
      }
    } catch {
      // Ignore errors
    }
  }

  // Count user commands
  const userDir = join(projectRoot, ".opencode", "command")
  if (existsSync(userDir)) {
    try {
      const entries = readdirSync(userDir)
      for (const f of entries) {
        if (f.endsWith(".md")) uniqueCommands.add(f)
      }
    } catch {
      // Ignore errors
    }
  }

  return uniqueCommands.size
}
