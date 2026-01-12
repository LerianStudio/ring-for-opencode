/**
 * Ring Help Command
 *
 * Provides comprehensive help for Ring capabilities including
 * skills, commands, and agents with their descriptions.
 */

import { existsSync, readdirSync, readFileSync } from "node:fs"
import { basename, dirname, join } from "node:path"
import { fileURLToPath } from "node:url"

/**
 * Options for the help command.
 */
export interface HelpOptions {
  /** Show only skills */
  skills?: boolean
  /** Show only commands */
  commands?: boolean
  /** Show only agents */
  agents?: boolean
  /** Show details for a specific item */
  item?: string
  /** Output in JSON format */
  json?: boolean
}

/**
 * Item info structure.
 */
interface ItemInfo {
  name: string
  description: string
  category: "skill" | "command" | "agent"
  path: string
}

/**
 * Get the assets path for Ring.
 * Returns null if assets directory doesn't exist.
 */
function getAssetsPath(): string | null {
  // ESM-compatible path resolution
  const currentFile = fileURLToPath(import.meta.url)
  const currentDir = dirname(currentFile)

  // Try relative to CLI (src/cli -> assets)
  const fromCli = join(currentDir, "..", "..", "assets")
  if (existsSync(fromCli)) {
    return fromCli
  }

  // Try relative to dist (dist/cli -> assets)
  const fromDist = join(currentDir, "..", "..", "..", "assets")
  if (existsSync(fromDist)) {
    return fromDist
  }

  return null
}

/**
 * Parse frontmatter from markdown content.
 */
function parseFrontmatter(content: string): { description?: string } {
  const normalizedContent = content.replace(/\r\n/g, "\n")
  const frontmatterRegex = /^---\n([\s\S]*?)\n---/
  const match = normalizedContent.match(frontmatterRegex)

  if (!match) {
    return {}
  }

  const yamlContent = match[1]
  const data: { description?: string } = {}

  for (const line of yamlContent.split("\n")) {
    const colonIndex = line.indexOf(":")
    if (colonIndex === -1) continue

    const key = line.slice(0, colonIndex).trim()
    let value = line.slice(colonIndex + 1).trim()

    // Remove quotes
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1)
    }

    if (key === "description") {
      data.description = value
    }
  }

  return data
}

/**
 * Load skills from assets/skill directory.
 */
function loadSkills(assetsPath: string): ItemInfo[] {
  const skillsDir = join(assetsPath, "skill")
  if (!existsSync(skillsDir)) {
    return []
  }

  const skills: ItemInfo[] = []
  const entries = readdirSync(skillsDir, { withFileTypes: true })

  for (const entry of entries) {
    if (!entry.isDirectory()) continue
    if (entry.name.startsWith(".") || entry.name === "shared-patterns") continue

    const skillFile = join(skillsDir, entry.name, "SKILL.md")
    if (!existsSync(skillFile)) continue

    try {
      const content = readFileSync(skillFile, "utf-8")
      const { description } = parseFrontmatter(content)

      // If no frontmatter description, try to extract from first heading or paragraph
      let desc = description
      if (!desc) {
        const firstPara = content.match(/^#[^\n]+\n+([^\n#]+)/m)
        desc = firstPara?.[1]?.trim().slice(0, 100) ?? "No description available"
      }

      skills.push({
        name: entry.name,
        description: desc,
        category: "skill",
        path: skillFile,
      })
    } catch {
      // Skip unreadable skills
    }
  }

  return skills.sort((a, b) => a.name.localeCompare(b.name))
}

/**
 * Load commands from assets/command directory.
 */
function loadCommands(assetsPath: string): ItemInfo[] {
  const commandsDir = join(assetsPath, "command")
  if (!existsSync(commandsDir)) {
    return []
  }

  const commands: ItemInfo[] = []
  const entries = readdirSync(commandsDir, { withFileTypes: true })

  for (const entry of entries) {
    if (!entry.isFile() || !entry.name.endsWith(".md")) continue

    const commandPath = join(commandsDir, entry.name)
    const commandName = basename(entry.name, ".md")

    try {
      const content = readFileSync(commandPath, "utf-8")
      const { description } = parseFrontmatter(content)

      commands.push({
        name: commandName,
        description: description ?? "No description available",
        category: "command",
        path: commandPath,
      })
    } catch {
      // Skip unreadable commands
    }
  }

  return commands.sort((a, b) => a.name.localeCompare(b.name))
}

/**
 * Load agents from assets/agent directory.
 */
function loadAgents(assetsPath: string): ItemInfo[] {
  const agentsDir = join(assetsPath, "agent")
  if (!existsSync(agentsDir)) {
    return []
  }

  const agents: ItemInfo[] = []
  const entries = readdirSync(agentsDir, { withFileTypes: true })

  for (const entry of entries) {
    if (!entry.isFile() || !entry.name.endsWith(".md")) continue

    const agentPath = join(agentsDir, entry.name)
    const agentName = basename(entry.name, ".md")

    try {
      const content = readFileSync(agentPath, "utf-8")
      const { description } = parseFrontmatter(content)

      agents.push({
        name: agentName,
        description: description ?? "No description available",
        category: "agent",
        path: agentPath,
      })
    } catch {
      // Skip unreadable agents
    }
  }

  return agents.sort((a, b) => a.name.localeCompare(b.name))
}

/**
 * Format items for display.
 */
function formatItems(items: ItemInfo[], category: string): string {
  if (items.length === 0) {
    return `No ${category} found.\n`
  }

  const maxNameLen = Math.max(...items.map((i) => i.name.length))
  const lines = items.map((item) => {
    const paddedName = item.name.padEnd(maxNameLen)
    const desc =
      item.description.length > 60 ? `${item.description.slice(0, 57)}...` : item.description
    return `  ${paddedName}  ${desc}`
  })

  return lines.join("\n")
}

/**
 * Show details for a specific item.
 */
function showItemDetails(
  item: string,
  skills: ItemInfo[],
  commands: ItemInfo[],
  agents: ItemInfo[],
): string {
  const allItems = [...skills, ...commands, ...agents]
  const found = allItems.find((i) => i.name === item || i.name === item.replace("ring:", ""))

  if (!found) {
    return `Item '${item}' not found. Use '/ring:help' to see available items.`
  }

  try {
    const content = readFileSync(found.path, "utf-8")
    // Remove frontmatter for display
    const withoutFrontmatter = content.replace(/^---\n[\s\S]*?\n---\n/, "")
    return `## ${found.category}: ${found.name}\n\n${withoutFrontmatter}`
  } catch {
    return `Unable to read details for '${item}'.`
  }
}

/**
 * Execute the help command.
 */
export async function help(options: HelpOptions = {}): Promise<number> {
  const assetsPath = getAssetsPath()

  if (!assetsPath) {
    console.error("Error: Ring assets directory not found.")
    console.error("Make sure Ring is properly installed.")
    return 1
  }

  const skills = loadSkills(assetsPath)
  const commands = loadCommands(assetsPath)
  const agents = loadAgents(assetsPath)

  // Show details for specific item
  if (options.item) {
    const details = showItemDetails(options.item, skills, commands, agents)
    console.log(details)
    return 0
  }

  // JSON output
  if (options.json) {
    const output: Record<string, ItemInfo[]> = {}
    if (!options.commands && !options.agents) output.skills = skills
    if (!options.skills && !options.agents) output.commands = commands
    if (!options.skills && !options.commands) output.agents = agents
    console.log(JSON.stringify(output, null, 2))
    return 0
  }

  // Filter by category
  const showSkills = options.skills || (!options.commands && !options.agents)
  const showCommands = options.commands || (!options.skills && !options.agents)
  const showAgents = options.agents || (!options.skills && !options.commands)

  const lines: string[] = []
  lines.push("Ring for OpenCode - Available Capabilities\n")
  lines.push(`${"=".repeat(50)}\n`)

  if (showSkills) {
    lines.push(`\nSkills (${skills.length})`)
    lines.push("-".repeat(30))
    lines.push(formatItems(skills, "skills"))
    lines.push("\nUsage: Load with 'skill: <name>' tool")
  }

  if (showCommands) {
    lines.push(`\n\nCommands (${commands.length})`)
    lines.push("-".repeat(30))
    lines.push(formatItems(commands, "commands"))
    lines.push("\nUsage: Invoke with '/ring:<name>'")
  }

  if (showAgents) {
    lines.push(`\n\nAgents (${agents.length})`)
    lines.push("-".repeat(30))
    lines.push(formatItems(agents, "agents"))
    lines.push("\nUsage: Dispatch with '@<name>' or Task tool")
  }

  lines.push(`\n\n${"=".repeat(50)}`)
  lines.push("For details on a specific item: /ring:help <name>")
  lines.push("Documentation: https://github.com/fredcamaral/ring-for-opencode")

  console.log(lines.join("\n"))
  return 0
}
