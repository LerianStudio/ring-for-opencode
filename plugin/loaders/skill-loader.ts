/**
 * Ring Skill Loader
 *
 * Loads Ring skills from .opencode/skill/{name}/SKILL.md files.
 * Skills are exposed as commands in OpenCode's system.
 */

import { existsSync, readdirSync, readFileSync } from "fs"
import { join } from "path"

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
 * Load skills from a directory.
 * Expects structure: skill/<skill-name>/SKILL.md
 */
function loadSkillsFromDir(
  skillsDir: string,
  disabledSkills: Set<string>
): Record<string, SkillConfig> {
  if (!existsSync(skillsDir)) {
    return {}
  }

  const result: Record<string, SkillConfig> = {}

  try {
    const entries = readdirSync(skillsDir, { withFileTypes: true })

    for (const entry of entries) {
      if (!entry.isDirectory()) continue

      const skillName = entry.name

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
        continue
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
 * Load Ring skills from .opencode/skill/ directory.
 */
export function loadRingSkills(
  projectRoot: string,
  disabledSkills: string[] = []
): Record<string, SkillConfig> {
  const skillsDir = join(projectRoot, ".opencode", "skill")
  const disabledSet = new Set(disabledSkills)
  return loadSkillsFromDir(skillsDir, disabledSet)
}

/**
 * Get count of available skills.
 */
export function countRingSkills(projectRoot: string): number {
  const skillsDir = join(projectRoot, ".opencode", "skill")
  if (!existsSync(skillsDir)) return 0

  try {
    const entries = readdirSync(skillsDir, { withFileTypes: true })
    let count = 0
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const skillFile = join(skillsDir, entry.name, "SKILL.md")
        if (existsSync(skillFile)) count++
      }
    }
    return count
  } catch {
    return 0
  }
}
