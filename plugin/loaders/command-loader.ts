/**
 * Ring Command Loader
 *
 * Loads Ring commands from .opencode/command/*.md files.
 * Commands become slash commands available via OpenCode.
 */

import { existsSync, readdirSync, readFileSync } from "fs"
import { join, basename } from "path"

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

    if ((value.startsWith('"') && value.endsWith('"')) ||
        (value.startsWith("'") && value.endsWith("'"))) {
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
  disabledCommands: Set<string>
): Record<string, CommandConfig> {
  if (!existsSync(commandsDir)) {
    return {}
  }

  const result: Record<string, CommandConfig> = {}

  try {
    const entries = readdirSync(commandsDir, { withFileTypes: true })

    for (const entry of entries) {
      if (!entry.isFile() || !entry.name.endsWith(".md")) continue

      const commandPath = join(commandsDir, entry.name)
      const commandName = basename(entry.name, ".md")

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

        result[commandName] = config
      } catch (error) {
        if (process.env.RING_DEBUG === "true") {
          console.debug(`[ring] Failed to parse ${commandPath}:`, error)
        }
        continue
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
 * Load Ring commands from .opencode/command/ directory.
 */
export function loadRingCommands(
  projectRoot: string,
  disabledCommands: string[] = []
): Record<string, CommandConfig> {
  const commandsDir = join(projectRoot, ".opencode", "command")
  const disabledSet = new Set(disabledCommands)
  return loadCommandsFromDir(commandsDir, disabledSet)
}

/**
 * Get count of available commands.
 */
export function countRingCommands(projectRoot: string): number {
  const commandsDir = join(projectRoot, ".opencode", "command")
  if (!existsSync(commandsDir)) return 0

  try {
    const entries = readdirSync(commandsDir)
    return entries.filter((f) => f.endsWith(".md")).length
  } catch {
    return 0
  }
}
