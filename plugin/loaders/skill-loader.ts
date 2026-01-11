/**
 * Ring Skill Loader
 *
 * Loads Ring skills from:
 * 1. Plugin's assets/skill/{name}/SKILL.md files (Ring's built-in skills)
 * 2. User's .opencode/skill/{name}/SKILL.md files (user customizations)
 *
 * User's skills take priority over Ring's built-in skills.
 */

import { existsSync, readdirSync, readFileSync } from "node:fs"
import { join } from "node:path"

/**
 * Skill configuration compatible with OpenCode SDK.
 */
export interface SkillConfig {
  description?: string
  agent?: string
  subtask?: boolean
}

/**
 * Frontmatter data from skill markdown files.
 */
interface SkillFrontmatter {
  description?: string
  agent?: string
  subtask?: string | boolean
}

/**
 * Keys that must never be used as object properties when building maps from filenames.
 */
const FORBIDDEN_OBJECT_KEYS = new Set(["__proto__", "constructor", "prototype"])

function isForbiddenObjectKey(key: string): boolean {
  return FORBIDDEN_OBJECT_KEYS.has(key)
}

// TODO(review): Consider using js-yaml for multiline YAML support

/**
 * Parse YAML frontmatter from markdown content.
 */
function parseFrontmatter(content: string): { data: SkillFrontmatter; body: string } {
  // Normalize line endings for cross-platform support
  const normalizedContent = content.replace(/\r\n/g, "\n")
  const frontmatterRegex = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/
  const match = normalizedContent.match(frontmatterRegex)

  if (!match) {
    return { data: {}, body: normalizedContent }
  }

  const yamlContent = match[1]
  const body = match[2]

  const data: SkillFrontmatter = {}
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
 * Load skills from a directory.
 * Expects structure: skill/<skill-name>/SKILL.md
 */
function loadSkillsFromDir(
  skillsDir: string,
  disabledSkills: Set<string>,
): Record<string, SkillConfig> {
  if (!existsSync(skillsDir)) {
    return {}
  }

  const result: Record<string, SkillConfig> = Object.create(null)

  try {
    const entries = readdirSync(skillsDir, { withFileTypes: true })

    for (const entry of entries) {
      if (!entry.isDirectory()) continue

      const skillName = entry.name

      // SECURITY: Skip forbidden gadget keys
      if (isForbiddenObjectKey(skillName)) continue

      // Skip disabled skills
      if (disabledSkills.has(skillName)) continue

      // Look for SKILL.md in the directory
      const skillFile = join(skillsDir, skillName, "SKILL.md")
      if (!existsSync(skillFile)) continue

      try {
        const content = readFileSync(skillFile, "utf-8")
        const { data } = parseFrontmatter(content)

        const config: SkillConfig = {
          description: data.description || `Ring skill: ${skillName}`,
        }

        if (data.agent) {
          config.agent = data.agent
        }

        if (typeof data.subtask === "boolean") {
          config.subtask = data.subtask
        }

        // Use ring-default namespace for skills
        result[`ring-default:${skillName}`] = config
      } catch (error) {
        if (process.env.RING_DEBUG === "true") {
          console.debug(`[ring] Failed to parse skill ${skillFile}:`, error)
        }
      }
    }
  } catch (error) {
    if (process.env.RING_DEBUG === "true") {
      console.debug(`[ring] Failed to read skills directory:`, error)
    }
    return {}
  }

  return result
}

/**
 * Load Ring skills from both plugin's assets/ and user's .opencode/ directories.
 *
 * @param pluginRoot - Path to the plugin directory (contains assets/)
 * @param projectRoot - Path to the user's project directory (contains .opencode/)
 * @param disabledSkills - List of skill names to skip
 * @returns Merged skill configs with user's taking priority
 */
export function loadRingSkills(
  pluginRoot: string,
  projectRoot: string,
  disabledSkills: string[] = [],
): Record<string, SkillConfig> {
  const disabledSet = new Set(disabledSkills)

  // Load Ring's built-in skills from assets/skill/
  const builtInDir = join(pluginRoot, "assets", "skill")
  const builtInSkills = loadSkillsFromDir(builtInDir, disabledSet)

  // Load user's custom skills from .opencode/skill/
  const userDir = join(projectRoot, ".opencode", "skill")
  const userSkills = loadSkillsFromDir(userDir, disabledSet)

  // Merge with user's taking priority, using a null-prototype map
  const merged: Record<string, SkillConfig> = Object.create(null)
  Object.assign(merged, builtInSkills)
  Object.assign(merged, userSkills)
  return merged
}

/**
 * Get count of available skills from both plugin's assets/ and user's .opencode/.
 */
export function countRingSkills(pluginRoot: string, projectRoot: string): number {
  const uniqueSkills = new Set<string>()

  // Count built-in skills
  const builtInDir = join(pluginRoot, "assets", "skill")
  if (existsSync(builtInDir)) {
    try {
      const entries = readdirSync(builtInDir, { withFileTypes: true })
      for (const entry of entries) {
        if (entry.isDirectory()) {
          const skillFile = join(builtInDir, entry.name, "SKILL.md")
          if (existsSync(skillFile)) uniqueSkills.add(entry.name)
        }
      }
    } catch {
      // Ignore errors
    }
  }

  // Count user skills
  const userDir = join(projectRoot, ".opencode", "skill")
  if (existsSync(userDir)) {
    try {
      const entries = readdirSync(userDir, { withFileTypes: true })
      for (const entry of entries) {
        if (entry.isDirectory()) {
          const skillFile = join(userDir, entry.name, "SKILL.md")
          if (existsSync(skillFile)) uniqueSkills.add(entry.name)
        }
      }
    } catch {
      // Ignore errors
    }
  }

  return uniqueSkills.size
}
