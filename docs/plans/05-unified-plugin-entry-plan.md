# Unified Plugin Entry Point and Config Injection Implementation Plan

> **For Agents:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Create a single unified plugin entry point that matches oh-my-opencode's registration pattern, enabling Ring to fully integrate with OpenCode's plugin system.

**Architecture:** Create a unified default export in `.opencode/plugin/index.ts` that matches the `Plugin` type, registers Ring's custom tools via the `tool` property, creates a config handler to inject 16 agents, 29 skills, and 16 commands into OpenCode's configuration, and routes lifecycle events to Ring's hook system. Maintain backward compatibility with existing plugin exports.

**Tech Stack:** TypeScript, Bun runtime, @opencode-ai/plugin, zod for schema validation

**Global Prerequisites:**
- Environment: macOS/Linux, Bun 1.0+, Node.js 18+
- Tools: Verify with commands below
- Access: No external API keys required
- State: Clean working tree or only pending plan changes

**Verification before starting:**
```bash
# Run ALL these commands and verify output:
bun --version              # Expected: 1.0.0+
node --version             # Expected: v18.0.0+
git status                 # Expected: clean or only docs/plans changes
ls .opencode/plugin/       # Expected: index.ts, doubt-resolver.ts, session-start.ts, etc.
ls .opencode/agent/*.md    # Expected: 16 agent files
ls .opencode/command/*.md  # Expected: 16 command files
```

## Historical Precedent

**Query:** "unified plugin config injection hooks"
**Index Status:** Empty (new project)

No historical data available. This is normal for new projects.
Proceeding with standard planning approach based on oh-my-opencode reference patterns.

---

## Part 1: Component Loaders

### Task 1: Create Agent Loader Module

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/agent-loader.ts`

**Prerequisites:**
- Directory `.opencode/plugin/` must exist
- Agent files exist in `.opencode/agent/*.md`

**Step 1: Create loaders directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders
```

**Expected output:**
```
(no output - silent success)
```

**Step 2: Write the agent loader**

Create file with this content:

```typescript
/**
 * Ring Agent Loader
 *
 * Loads Ring agents from .opencode/agent/*.md files.
 * Parses YAML frontmatter and markdown body into agent configs.
 */

import { existsSync, readdirSync, readFileSync } from "fs"
import { join, basename } from "path"

/**
 * Agent configuration compatible with OpenCode SDK.
 */
export interface AgentConfig {
  description?: string
  mode?: "primary" | "subagent"
  prompt?: string
  model?: string
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
  tools?: string
  color?: string
}

/**
 * Parse YAML frontmatter from markdown content.
 */
function parseFrontmatter(content: string): { data: AgentFrontmatter; body: string } {
  const frontmatterRegex = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/
  const match = content.match(frontmatterRegex)

  if (!match) {
    return { data: {}, body: content }
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
    if ((value.startsWith('"') && value.endsWith('"')) ||
        (value.startsWith("'") && value.endsWith("'"))) {
      value = value.slice(1, -1)
    }

    if (key === "description") data.description = value
    if (key === "mode") data.mode = value
    if (key === "model") data.model = value
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

  const tools = toolsStr.split(",").map((t) => t.trim()).filter(Boolean)
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
 * Load agents from a directory.
 */
function loadAgentsFromDir(
  agentsDir: string,
  disabledAgents: Set<string>
): Record<string, AgentConfig> {
  if (!existsSync(agentsDir)) {
    return {}
  }

  const result: Record<string, AgentConfig> = {}

  try {
    const entries = readdirSync(agentsDir, { withFileTypes: true })

    for (const entry of entries) {
      if (!entry.isFile() || !entry.name.endsWith(".md")) continue

      const agentPath = join(agentsDir, entry.name)
      const agentName = basename(entry.name, ".md")

      // Skip disabled agents
      if (disabledAgents.has(agentName)) continue

      try {
        const content = readFileSync(agentPath, "utf-8")
        const { data, body } = parseFrontmatter(content)

        const config: AgentConfig = {
          description: data.description ? `(ring) ${data.description}` : `(ring) ${agentName} agent`,
          mode: (data.mode as "primary" | "subagent") || "subagent",
          prompt: body.trim(),
        }

        if (data.model) {
          config.model = data.model
        }

        if (data.color) {
          config.color = data.color
        }

        const toolsConfig = parseToolsConfig(data.tools)
        if (toolsConfig) {
          config.tools = toolsConfig
        }

        result[agentName] = config
      } catch {
        // Skip files that fail to parse
        continue
      }
    }
  } catch {
    // Directory read failed
    return {}
  }

  return result
}

/**
 * Load Ring agents from .opencode/agent/ directory.
 */
export function loadRingAgents(
  projectRoot: string,
  disabledAgents: string[] = []
): Record<string, AgentConfig> {
  const agentsDir = join(projectRoot, ".opencode", "agent")
  const disabledSet = new Set(disabledAgents)
  return loadAgentsFromDir(agentsDir, disabledSet)
}

/**
 * Get count of available agents.
 */
export function countRingAgents(projectRoot: string): number {
  const agentsDir = join(projectRoot, ".opencode", "agent")
  if (!existsSync(agentsDir)) return 0

  try {
    const entries = readdirSync(agentsDir)
    return entries.filter((f) => f.endsWith(".md")).length
  } catch {
    return 0
  }
}
```

**Step 3: Verify file was created**

Run: `head -20 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/agent-loader.ts`

**Expected output:**
```
/**
 * Ring Agent Loader
 *
 * Loads Ring agents from .opencode/agent/*.md files.
 * Parses YAML frontmatter and markdown body into agent configs.
 */

import { existsSync, readdirSync, readFileSync } from "fs"
import { join, basename } from "path"

/**
 * Agent configuration compatible with OpenCode SDK.
 */
export interface AgentConfig {
  description?: string
  mode?: "primary" | "subagent"
  prompt?: string
  model?: string
```

**If Task Fails:**
1. **Directory doesn't exist:** Run `mkdir -p .opencode/plugin/loaders` first
2. **Permission denied:** Check file permissions with `ls -la .opencode/plugin/`
3. **Can't recover:** Document error and return to human partner

---

### Task 2: Create Skill Loader Module

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/skill-loader.ts`

**Prerequisites:**
- Directory `.opencode/plugin/loaders/` must exist
- Skill directories exist in `.opencode/skill/*/`

**Step 1: Write the skill loader**

Create file with this content:

```typescript
/**
 * Ring Skill Loader
 *
 * Loads Ring skills from .opencode/skill/*/SKILL.md files.
 * Skills are exposed as commands in OpenCode's system.
 */

import { existsSync, readdirSync, readFileSync, statSync } from "fs"
import { join, basename } from "path"

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
 * Parse YAML frontmatter from markdown content.
 */
function parseFrontmatter(content: string): { data: SkillFrontmatter; body: string } {
  const frontmatterRegex = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/
  const match = content.match(frontmatterRegex)

  if (!match) {
    return { data: {}, body: content }
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
      } catch {
        continue
      }
    }
  } catch {
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
```

**Step 2: Verify file was created**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/skill-loader.ts`

**Expected output:**
```
     140 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/skill-loader.ts
```

**If Task Fails:**
1. **Import error:** Verify node_modules available
2. **Syntax error:** Run `bun build .opencode/plugin/loaders/skill-loader.ts --no-bundle 2>&1`

---

### Task 3: Create Command Loader Module

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/command-loader.ts`

**Prerequisites:**
- Directory `.opencode/plugin/loaders/` must exist
- Command files exist in `.opencode/command/*.md`

**Step 1: Write the command loader**

Create file with this content:

```typescript
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
  const frontmatterRegex = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/
  const match = content.match(frontmatterRegex)

  if (!match) {
    return { data: {}, body: content }
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
      } catch {
        continue
      }
    }
  } catch {
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
```

**Step 2: Verify file was created**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/command-loader.ts`

**Expected output:**
```
     126 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/command-loader.ts
```

---

### Task 4: Create Loaders Index

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/loaders/index.ts`

**Prerequisites:**
- All loader files exist: `agent-loader.ts`, `skill-loader.ts`, `command-loader.ts`

**Step 1: Write the loaders index**

Create file with this content:

```typescript
/**
 * Ring Component Loaders
 *
 * Central export for all component loaders.
 */

// Agent loader
export {
  loadRingAgents,
  countRingAgents,
  type AgentConfig,
} from "./agent-loader"

// Skill loader
export {
  loadRingSkills,
  countRingSkills,
  type SkillConfig,
} from "./skill-loader"

// Command loader
export {
  loadRingCommands,
  countRingCommands,
  type CommandConfig,
} from "./command-loader"
```

**Step 2: Verify TypeScript compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build .opencode/plugin/loaders/index.ts --no-bundle --outdir=/tmp/ring-test 2>&1 | head -5`

**Expected output:**
```
(empty or success message - no errors)
```

---

## Part 2: Configuration Handler

### Task 5: Create Ring Configuration Types

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/types.ts`

**Prerequisites:**
- Create config directory

**Step 1: Create config directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config
```

**Step 2: Write the configuration types**

Create file with this content:

```typescript
/**
 * Ring Configuration Types
 *
 * Type definitions for Ring's plugin configuration.
 */

/**
 * Ring plugin configuration.
 */
export interface RingPluginConfig {
  /** Disabled agents (won't be injected) */
  disabled_agents?: string[]

  /** Disabled skills (won't be injected) */
  disabled_skills?: string[]

  /** Disabled commands (won't be injected) */
  disabled_commands?: string[]

  /** Disabled hooks (from Plan 4) */
  disabled_hooks?: string[]

  /** Enable debug logging */
  debug?: boolean
}

/**
 * Default Ring configuration.
 */
export const DEFAULT_RING_CONFIG: RingPluginConfig = {
  disabled_agents: [],
  disabled_skills: [],
  disabled_commands: [],
  disabled_hooks: [],
  debug: false,
}

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
```

**Step 3: Verify file was created**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/types.ts | head -15`

**Expected output:**
```
/**
 * Ring Configuration Types
 *
 * Type definitions for Ring's plugin configuration.
 */

/**
 * Ring plugin configuration.
 */
export interface RingPluginConfig {
  /** Disabled agents (won't be injected) */
  disabled_agents?: string[]

  /** Disabled skills (won't be injected) */
  disabled_skills?: string[]
```

---

### Task 6: Create Configuration Loader

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/loader.ts`

**Prerequisites:**
- Types file exists: `config/types.ts`

**Step 1: Write the configuration loader**

Create file with this content:

```typescript
/**
 * Ring Configuration Loader
 *
 * Loads Ring configuration from various locations.
 * Supports: .ring/config.json, .ring/config.jsonc, .opencode/ring.json
 */

import { existsSync, readFileSync } from "fs"
import { join } from "path"
import { DEFAULT_RING_CONFIG, type RingPluginConfig } from "./types"

/**
 * Strip JSONC comments and trailing commas.
 */
function parseJsonc<T>(content: string): T {
  // Remove single-line comments
  let cleaned = content.replace(/\/\/.*$/gm, "")
  // Remove multi-line comments
  cleaned = cleaned.replace(/\/\*[\s\S]*?\*\//g, "")
  // Remove trailing commas before } or ]
  cleaned = cleaned.replace(/,(\s*[}\]])/g, "$1")
  return JSON.parse(cleaned) as T
}

/**
 * Get configuration file paths in priority order.
 */
function getConfigPaths(projectRoot: string): string[] {
  return [
    join(projectRoot, ".ring", "config.jsonc"),
    join(projectRoot, ".ring", "config.json"),
    join(projectRoot, ".opencode", "ring.jsonc"),
    join(projectRoot, ".opencode", "ring.json"),
  ]
}

/**
 * Load Ring configuration from project.
 */
export function loadRingConfig(projectRoot: string): RingPluginConfig {
  const paths = getConfigPaths(projectRoot)

  for (const path of paths) {
    if (!existsSync(path)) continue

    try {
      const content = readFileSync(path, "utf-8")
      if (!content.trim()) continue

      const parsed = parseJsonc<Partial<RingPluginConfig>>(content)

      // Merge with defaults
      return {
        ...DEFAULT_RING_CONFIG,
        ...parsed,
      }
    } catch (error) {
      if (process.env.DEBUG) {
        console.debug(`[ring] Failed to load config from ${path}:`, error)
      }
      continue
    }
  }

  // Return defaults if no config found
  return { ...DEFAULT_RING_CONFIG }
}

/**
 * Check if a component is disabled.
 */
export function isDisabled(
  config: RingPluginConfig,
  type: "agent" | "skill" | "command" | "hook",
  name: string
): boolean {
  const key = `disabled_${type}s` as keyof RingPluginConfig
  const disabled = config[key] as string[] | undefined
  return disabled?.includes(name) ?? false
}
```

**Step 2: Verify file was created**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/loader.ts`

**Expected output:**
```
      77 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/loader.ts
```

---

### Task 7: Create Config Handler

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/config-handler.ts`

**Prerequisites:**
- Loaders exist: `loaders/index.ts`
- Config types exist: `config/types.ts`

**Step 1: Write the config handler**

Create file with this content:

```typescript
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

import { loadRingAgents } from "../loaders/agent-loader"
import { loadRingSkills } from "../loaders/skill-loader"
import { loadRingCommands } from "../loaders/command-loader"
import type { RingPluginConfig, OpenCodeConfig } from "./types"

/**
 * Dependencies for creating the config handler.
 */
export interface ConfigHandlerDeps {
  /** Project root directory */
  projectRoot: string
  /** Ring plugin configuration */
  ringConfig: RingPluginConfig
}

/**
 * Create the config handler that injects Ring components.
 *
 * This handler is called by OpenCode to modify the configuration
 * before the session starts. We use this to inject:
 * - Ring agents (16 agents from .opencode/agent/)
 * - Ring skills (29 skills from .opencode/skill/)
 * - Ring commands (16 commands from .opencode/command/)
 */
export function createConfigHandler(deps: ConfigHandlerDeps) {
  const { projectRoot, ringConfig } = deps

  return async (config: OpenCodeConfig): Promise<void> => {
    const debug = ringConfig.debug || process.env.DEBUG

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
```

**Step 2: Verify file was created**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/config-handler.ts`

**Expected output:**
```
     124 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/config-handler.ts
```

---

### Task 8: Create Config Index

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/config/index.ts`

**Prerequisites:**
- All config files exist

**Step 1: Write the config index**

Create file with this content:

```typescript
/**
 * Ring Configuration Module
 *
 * Central export for configuration types, loader, and handler.
 */

// Types
export type {
  RingPluginConfig,
  OpenCodeConfig,
} from "./types"

export { DEFAULT_RING_CONFIG } from "./types"

// Loader
export { loadRingConfig, isDisabled } from "./loader"

// Handler
export { createConfigHandler, type ConfigHandlerDeps } from "./config-handler"
```

**Step 2: Verify config module compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build .opencode/plugin/config/index.ts --no-bundle --outdir=/tmp/ring-test 2>&1 | head -5`

**Expected output:**
```
(empty or success message - no errors)
```

---

## Part 3: Unified Plugin Entry Point

### Task 9: Create Ring Tools Module

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/tools/index.ts`

**Prerequisites:**
- Create tools directory

**Step 1: Create tools directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/tools
```

**Step 2: Write the tools module**

Create file with this content:

```typescript
/**
 * Ring Tools
 *
 * Custom tools registered by Ring plugin.
 * Uses the tool() helper from @opencode-ai/plugin.
 */

import { tool } from "@opencode-ai/plugin"

/**
 * Ring Doubt Resolver Tool
 *
 * Provides a structured "pick from options" prompt when the agent has doubts.
 * This is NOT a native OpenCode UI picker - it prints a numbered list
 * and asks the user to reply normally.
 */
export const ringDoubtTool = tool({
  description:
    "Ask the user to choose from options (single/multi + free-text alternative) when the agent has doubts.",
  args: {
    question: tool.schema.string().min(1).describe("What decision is needed"),
    options: tool.schema
      .array(tool.schema.string().min(1))
      .min(1)
      .describe("List of selectable options"),
    multi: tool.schema.boolean().optional().describe("Allow multiple selections"),
    allowCustom: tool.schema
      .boolean()
      .optional()
      .describe("Allow user to type something else instead of picking"),
    context: tool.schema.string().optional().describe("Optional extra context for the decision"),
  },
  execute: async (args) => {
    const multi = Boolean(args.multi)
    const allowCustom = args.allowCustom !== false

    const header =
      "<MANDATORY-USER-MESSAGE>\nDOUBT CHECKPOINT - USER CHOICE REQUIRED\n</MANDATORY-USER-MESSAGE>"

    const lines: string[] = []
    lines.push(header)
    lines.push("")
    lines.push(args.question.trim())

    if (args.context) {
      lines.push("")
      lines.push("Context:")
      lines.push(args.context.trim())
    }

    lines.push("")
    lines.push("Options:")
    args.options.forEach((opt, i) => {
      lines.push(`${i + 1}) ${opt}`)
    })

    lines.push("")

    if (multi) {
      lines.push(`Reply with one or more numbers (example: "1" or "1,3").`)
    } else {
      lines.push(`Reply with one number (example: "2").`)
    }

    if (allowCustom) {
      lines.push('Or type your own answer (example: "Other: ...").')
    }

    return lines.join("\n")
  },
})

/**
 * All Ring tools to register with OpenCode.
 */
export const ringTools = {
  ring_doubt: ringDoubtTool,
}
```

**Step 3: Verify file was created**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/tools/index.ts`

**Expected output:**
```
      76 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/tools/index.ts
```

---

### Task 10: Create Lifecycle Router

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/lifecycle/router.ts`

**Prerequisites:**
- Create lifecycle directory

**Step 1: Create lifecycle directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/lifecycle
```

**Step 2: Write the lifecycle router**

Create file with this content:

```typescript
/**
 * Ring Lifecycle Router
 *
 * Routes OpenCode lifecycle events to Ring's plugin handlers.
 * Maps OpenCode events to Ring hook registry (when Plan 4 is implemented).
 */

import type { RingPluginConfig } from "../config/types"
import { getSessionId, cleanupOldState, deleteState } from "../utils/state"

/**
 * OpenCode event structure.
 */
export interface OpenCodeEvent {
  type: string
  properties?: Record<string, unknown>
}

/**
 * Dependencies for lifecycle router.
 */
export interface LifecycleRouterDeps {
  projectRoot: string
  ringConfig: RingPluginConfig
  notifySessionIdle?: () => Promise<void>
  notifySessionError?: () => Promise<void>
}

/**
 * Create the event handler that routes lifecycle events.
 *
 * Event mappings:
 * - session.created -> Reset context state, cleanup old files
 * - session.idle -> Notification hook
 * - session.error -> Notification hook
 * - todo.updated -> Task completion hook
 * - experimental.session.compacting -> Context injection
 */
export function createLifecycleRouter(deps: LifecycleRouterDeps) {
  const { projectRoot, ringConfig, notifySessionIdle, notifySessionError } = deps
  const sessionId = getSessionId()
  const debug = ringConfig.debug || process.env.DEBUG

  return async (input: { event: OpenCodeEvent }): Promise<void> => {
    const { event } = input
    const eventType = event.type

    if (debug) {
      console.debug(`[ring] Event: ${eventType}`)
    }

    // session.created - Initialize session
    if (eventType === "session.created") {
      // Reset context usage state for new session
      deleteState(projectRoot, "context-usage", sessionId)
      // Clean up old state files (> 7 days)
      cleanupOldState(projectRoot, 7)

      if (debug) {
        console.debug("[ring] Session initialized, state reset")
      }
      return
    }

    // session.idle - Session completed
    if (eventType === "session.idle") {
      if (notifySessionIdle) {
        await notifySessionIdle()
      }
      return
    }

    // session.error - Session failed
    if (eventType === "session.error") {
      if (notifySessionError) {
        await notifySessionError()
      }
      return
    }

    // todo.updated - Task completion tracking
    if (eventType === "todo.updated") {
      // TODO: Route to task-completion hook when Plan 4 is implemented
      // For now, task-completion-check.ts handles this directly
      return
    }

    // Other events - no action needed
  }
}

/**
 * Event type constants for type safety.
 */
export const EVENTS = {
  SESSION_CREATED: "session.created",
  SESSION_IDLE: "session.idle",
  SESSION_ERROR: "session.error",
  SESSION_DELETED: "session.deleted",
  TODO_UPDATED: "todo.updated",
  MESSAGE_PART_UPDATED: "message.part.updated",
} as const
```

**Step 3: Verify file was created**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/lifecycle/router.ts`

**Expected output:**
```
      99 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/lifecycle/router.ts
```

---

### Task 11: Create Lifecycle Index

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/lifecycle/index.ts`

**Step 1: Write the lifecycle index**

Create file with this content:

```typescript
/**
 * Ring Lifecycle Module
 *
 * Central export for lifecycle routing.
 */

export {
  createLifecycleRouter,
  EVENTS,
  type OpenCodeEvent,
  type LifecycleRouterDeps,
} from "./router"
```

**Step 2: Verify file was created**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/lifecycle/index.ts`

**Expected output:**
```
/**
 * Ring Lifecycle Module
 *
 * Central export for lifecycle routing.
 */

export {
  createLifecycleRouter,
  EVENTS,
  type OpenCodeEvent,
  type LifecycleRouterDeps,
} from "./router"
```

---

### Task 12: Create Unified Plugin Entry Point

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/ring-unified.ts`

**Prerequisites:**
- All modules exist: loaders, config, tools, lifecycle

**Step 1: Write the unified plugin**

Create file with this content:

```typescript
/**
 * Ring Unified Plugin
 *
 * Single entry point that matches oh-my-opencode's registration pattern.
 * Combines all Ring functionality into one Plugin export.
 *
 * Features:
 * - Config handler: Injects agents, skills, commands
 * - Tool registration: ring_doubt tool
 * - Event routing: Lifecycle events to hooks
 * - System transform: Context injection
 * - Compaction: Context preservation
 */

import type { Plugin, PluginInput, Hooks } from "@opencode-ai/plugin"

// Config
import { loadRingConfig, createConfigHandler } from "./config"

// Tools
import { ringTools } from "./tools"

// Lifecycle
import { createLifecycleRouter } from "./lifecycle"

// Existing plugins (for backward compatibility)
import { RingSessionStart } from "./session-start"
import { RingContextInjection } from "./context-injection"
import { RingNotification } from "./notification"
import { RingTaskCompletionCheck } from "./task-completion-check"

// Utils
import { getSessionId, sanitizeForPrompt, escapeAngleBrackets } from "./utils/state"
import { existsSync, readFileSync, readdirSync, statSync, lstatSync } from "fs"
import { join } from "path"

/**
 * Critical rules for system prompt injection.
 */
const CRITICAL_RULES = `## ORCHESTRATOR CRITICAL RULES (SURVIVE COMPACT)

**3-FILE RULE: HARD GATE**
DO NOT read/edit >3 files directly. This is a PROHIBITION.
- >3 files -> STOP. Launch specialist agent. DO NOT proceed manually.
- Already touched 3 files? -> At gate. Dispatch agent NOW.

**AUTO-TRIGGER PHRASES -> MANDATORY AGENT:**
- "fix issues/remaining/findings" -> Launch specialist agent
- "apply fixes", "fix the X issues" -> Launch specialist agent
- "find where", "search for", "understand how" -> Launch Explore agent

**If you think "this task is small" or "I can handle 5 files":**
WRONG. Count > 3 = agent. No exceptions. Task size is irrelevant.

**Full rules:** Use skill tool with "ring-default:using-ring" if needed.`

/**
 * Find and read the active continuity ledger.
 */
function findActiveLedger(projectRoot: string): { name: string; content: string; currentPhase: string } | null {
  const ledgerDir = join(projectRoot, ".ring/ledgers")
  if (!existsSync(ledgerDir)) return null

  try {
    const files = readdirSync(ledgerDir)
      .filter((f) => f.startsWith("CONTINUITY-") && f.endsWith(".md"))
      .map((f) => ({
        name: f,
        path: join(ledgerDir, f),
        mtime: statSync(join(ledgerDir, f)).mtimeMs,
        isSymlink: lstatSync(join(ledgerDir, f)).isSymbolicLink(),
      }))
      .filter((f) => !f.isSymlink)
      .sort((a, b) => b.mtime - a.mtime)

    if (files.length === 0) return null

    const newest = files[0]
    const content = readFileSync(newest.path, "utf-8")
    const phaseMatch = content.match(/\[->/m)
    const currentPhase = phaseMatch
      ? content.split("\n").find((line) => line.includes("[->"))?.trim() || ""
      : ""

    return {
      name: newest.name.replace(".md", ""),
      content,
      currentPhase,
    }
  } catch {
    return null
  }
}

/**
 * Sanitize notification content for shell safety.
 */
function sanitizeContent(content: string, maxLength: number = 100): string {
  return content
    .replace(/[^a-zA-Z0-9 .,!?:;()\-]/g, "")
    .slice(0, Math.max(0, maxLength))
}

/**
 * Platform-specific notification sender.
 */
async function sendNotification(
  title: string,
  message: string,
  $: PluginInput["$"]
): Promise<void> {
  const platform = process.platform
  const safeTitle = sanitizeContent(title, 50)
  const safeMessage = sanitizeContent(message, 100)

  try {
    if (platform === "darwin") {
      const appleScript = `display notification "${safeMessage}" with title "${safeTitle}"`
      await $`osascript -e ${appleScript}`
    } else if (platform === "linux") {
      await $`notify-send ${[safeTitle]} ${[safeMessage]}`
    }
  } catch {
    // Notification not critical - silent failure
  }
}

/**
 * Ring Unified Plugin
 *
 * Matches oh-my-opencode's Plugin signature:
 * Plugin = (input: PluginInput) => Promise<Hooks>
 */
const RingUnifiedPlugin: Plugin = async (ctx: PluginInput): Promise<Hooks> => {
  const { directory, $, client } = ctx
  const projectRoot = directory

  // Load Ring configuration
  const ringConfig = loadRingConfig(projectRoot)
  const debug = ringConfig.debug || process.env.DEBUG

  if (debug) {
    console.debug("[ring] Initializing unified plugin")
  }

  // Create config handler
  const configHandler = createConfigHandler({
    projectRoot,
    ringConfig,
  })

  // Create lifecycle router
  const lifecycleRouter = createLifecycleRouter({
    projectRoot,
    ringConfig,
    notifySessionIdle: async () => {
      await sendNotification("Ring", "Session completed!", $)
    },
    notifySessionError: async () => {
      await sendNotification("Ring", "Session encountered an error", $)
    },
  })

  // Initialize individual plugins for their specific functionality
  const sessionStartHooks = await RingSessionStart({ ...ctx })
  const contextInjectionHooks = await RingContextInjection({ ...ctx })
  const taskCompletionHooks = await RingTaskCompletionCheck({ ...ctx })

  // Skills reference for injection
  const skillsReference = `## Ring Skills System

Key skills for common tasks:
- **test-driven-development**: TDD methodology (RED -> GREEN -> REFACTOR)
- **requesting-code-review**: Parallel review with 5 specialized reviewers
- **systematic-debugging**: 4-phase debugging methodology
- **writing-plans**: Create detailed implementation plans
- **executing-plans**: Execute plans in batches with review checkpoints
- **dispatching-parallel-agents**: Run multiple agents concurrently
- **verification-before-completion**: Ensure work complete before done`

  return {
    // Register Ring tools
    tool: ringTools,

    // Config handler - inject agents, skills, commands
    config: configHandler,

    // Event handler - lifecycle routing
    event: async (input) => {
      // Route to lifecycle router
      await lifecycleRouter(input)

      // Delegate to individual plugin handlers
      if (sessionStartHooks.event) {
        await sessionStartHooks.event(input)
      }
      if (taskCompletionHooks.event) {
        await taskCompletionHooks.event(input)
      }
    },

    // System prompt transformation
    "experimental.chat.system.transform": async (
      input: Record<string, unknown>,
      output: { system: string[] }
    ) => {
      if (!output?.system || !Array.isArray(output.system)) return

      // Delegate to session-start plugin
      if (sessionStartHooks["experimental.chat.system.transform"]) {
        await sessionStartHooks["experimental.chat.system.transform"](input, output)
      }
    },

    // Compaction context injection
    "experimental.session.compacting": async (
      input: { sessionID: string },
      output: { context: string[]; prompt?: string }
    ) => {
      if (!output?.context || !Array.isArray(output.context)) return

      // Delegate to context-injection plugin
      if (contextInjectionHooks["experimental.session.compacting"]) {
        await contextInjectionHooks["experimental.session.compacting"](input, output)
      }
    },
  }
}

export default RingUnifiedPlugin
export { RingUnifiedPlugin }
```

**Step 2: Verify file was created**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/ring-unified.ts`

**Expected output:**
```
     227 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/ring-unified.ts
```

---

### Task 13: Update Main Plugin Index

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/index.ts`

**Prerequisites:**
- Backup existing index.ts
- All new modules exist

**Step 1: Read current index.ts**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/plugin/index.ts`

**Step 2: Update the index to add unified export**

Replace the content with:

```typescript
/**
 * Ring OpenCode Plugins
 *
 * This module exports all Ring plugins for OpenCode.
 *
 * Architecture:
 * - Unified plugin (default export): Single entry matching oh-my-opencode pattern
 * - Individual plugins: Backward compatible exports for direct import
 *
 * The unified plugin combines:
 * - Config handler: Injects 16 agents, 29 skills, 16 commands
 * - Tool registration: ring_doubt tool
 * - Event routing: Lifecycle events to hooks
 * - System transform: Context injection
 * - Compaction: Context preservation
 */

// =============================================================================
// UNIFIED PLUGIN (Primary Export)
// =============================================================================

// Default export: Unified plugin matching oh-my-opencode pattern
export { default, RingUnifiedPlugin } from "./ring-unified"

// =============================================================================
// COMPONENT EXPORTS
// =============================================================================

// Config module
export {
  loadRingConfig,
  createConfigHandler,
  isDisabled,
  DEFAULT_RING_CONFIG,
  type RingPluginConfig,
  type OpenCodeConfig,
} from "./config"

// Loaders
export {
  loadRingAgents,
  loadRingSkills,
  loadRingCommands,
  countRingAgents,
  countRingSkills,
  countRingCommands,
  type AgentConfig,
  type SkillConfig,
  type CommandConfig,
} from "./loaders"

// Tools
export { ringTools, ringDoubtTool } from "./tools"

// Lifecycle
export {
  createLifecycleRouter,
  EVENTS,
  type OpenCodeEvent,
  type LifecycleRouterDeps,
} from "./lifecycle"

// =============================================================================
// LEGACY PLUGIN EXPORTS (Backward Compatibility)
// =============================================================================

// Core plugins
export { RingSessionStart } from "./session-start"
export { RingContextInjection } from "./context-injection"
export { RingNotification } from "./notification"

// Workflow plugins
export { RingTaskCompletionCheck } from "./task-completion-check"
export { RingSessionOutcome } from "./session-outcome"

// Session analytics plugins
export { RingOutcomeInference } from "./outcome-inference"

// Interactive tools
export { RingDoubtResolver } from "./doubt-resolver"
```

**Step 3: Verify updated index compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build .opencode/plugin/index.ts --no-bundle --outdir=/tmp/ring-test 2>&1 | head -10`

**Expected output:**
```
(empty or success message - no errors)
```

---

## Part 4: Code Review and Verification

### Task 14: Run Code Review Checkpoint

**Step 1: Dispatch code reviewers in parallel**

REQUIRED SUB-SKILL: Use requesting-code-review

Review all files created in this plan:
- `.opencode/plugin/loaders/*.ts` (3 files)
- `.opencode/plugin/config/*.ts` (4 files)
- `.opencode/plugin/tools/index.ts`
- `.opencode/plugin/lifecycle/*.ts` (2 files)
- `.opencode/plugin/ring-unified.ts`
- `.opencode/plugin/index.ts` (modified)

**Step 2: Handle findings by severity**

**Critical/High/Medium Issues:**
- Fix immediately
- Re-run reviewers after fixes
- Repeat until zero Critical/High/Medium issues

**Low Issues:**
- Add `TODO(review):` comments at relevant locations

**Step 3: Proceed only when zero Critical/High/Medium issues remain**

---

### Task 15: Run TypeScript Build Verification

**Step 1: Build all plugin files**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build .opencode/plugin/index.ts --no-bundle --outdir=/tmp/ring-build 2>&1`

**Expected output:**
```
(no errors)
```

**Step 2: Verify module exports**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun -e "
import plugin, { loadRingAgents, loadRingCommands, loadRingSkills, ringTools, createConfigHandler } from './.opencode/plugin/index.ts';

console.log('Default export is function:', typeof plugin === 'function');
console.log('loadRingAgents exists:', typeof loadRingAgents === 'function');
console.log('loadRingCommands exists:', typeof loadRingCommands === 'function');
console.log('loadRingSkills exists:', typeof loadRingSkills === 'function');
console.log('ringTools has ring_doubt:', 'ring_doubt' in ringTools);
console.log('createConfigHandler exists:', typeof createConfigHandler === 'function');
"`

**Expected output:**
```
Default export is function: true
loadRingAgents exists: true
loadRingCommands exists: true
loadRingSkills exists: true
ringTools has ring_doubt: true
createConfigHandler exists: true
```

**If Task Fails:**
1. **Import error:** Check file paths and exports
2. **Type error:** Run `bun typecheck` if available
3. **Can't recover:** Document error and return to human partner

---

### Task 16: Test Component Loaders

**Step 1: Test agent loader**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun -e "
import { loadRingAgents, countRingAgents } from './.opencode/plugin/loaders/index.ts';

const count = countRingAgents(process.cwd());
const agents = loadRingAgents(process.cwd());
const names = Object.keys(agents);

console.log('Agent count:', count);
console.log('Loaded agents:', names.length);
console.log('First 5:', names.slice(0, 5).join(', '));
console.log('Has code-reviewer:', 'code-reviewer' in agents);
console.log('Agent mode:', agents['code-reviewer']?.mode);
"`

**Expected output:**
```
Agent count: 16
Loaded agents: 16
First 5: backend-engineer-golang, backend-engineer-typescript, business-logic-reviewer, code-reviewer, codebase-explorer
Has code-reviewer: true
Agent mode: subagent
```

**Step 2: Test skill loader**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun -e "
import { loadRingSkills, countRingSkills } from './.opencode/plugin/loaders/index.ts';

const count = countRingSkills(process.cwd());
const skills = loadRingSkills(process.cwd());
const names = Object.keys(skills);

console.log('Skill count:', count);
console.log('Loaded skills:', names.length);
console.log('First 3:', names.slice(0, 3).join(', '));
"`

**Expected output:**
```
Skill count: 29
Loaded skills: 29
First 3: ring-default:brainstorming, ring-default:defense-in-depth, ring-default:dev-cycle
```

**Step 3: Test command loader**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun -e "
import { loadRingCommands, countRingCommands } from './.opencode/plugin/loaders/index.ts';

const count = countRingCommands(process.cwd());
const commands = loadRingCommands(process.cwd());
const names = Object.keys(commands);

console.log('Command count:', count);
console.log('Loaded commands:', names.length);
console.log('First 3:', names.slice(0, 3).join(', '));
console.log('Has commit:', 'commit' in commands);
"`

**Expected output:**
```
Command count: 16
Loaded commands: 16
First 3: brainstorm, codereview, commit
Has commit: true
```

---

### Task 17: Test Unified Plugin

**Step 1: Test plugin initialization**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun -e "
import RingPlugin from './.opencode/plugin/ring-unified.ts';

// Mock context
const mockCtx = {
  directory: process.cwd(),
  worktree: process.cwd(),
  project: { name: 'test' },
  serverUrl: new URL('http://localhost:3000'),
  client: {
    tui: { showToast: async () => {} },
    session: { prompt: async () => {} },
  },
  \$: async () => {},
};

(async () => {
  try {
    const hooks = await RingPlugin(mockCtx);
    console.log('Plugin initialized: true');
    console.log('Has tool property:', 'tool' in hooks);
    console.log('Has config handler:', 'config' in hooks);
    console.log('Has event handler:', 'event' in hooks);
    console.log('Has ring_doubt tool:', 'ring_doubt' in (hooks.tool || {}));
  } catch (e) {
    console.error('Plugin init failed:', e.message);
  }
})();
"`

**Expected output:**
```
Plugin initialized: true
Has tool property: true
Has config handler: true
Has event handler: true
Has ring_doubt tool: true
```

---

### Task 18: Test Config Handler

**Step 1: Test config injection**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun -e "
import { createConfigHandler, loadRingConfig } from './.opencode/plugin/config/index.ts';

const projectRoot = process.cwd();
const ringConfig = loadRingConfig(projectRoot);
const handler = createConfigHandler({ projectRoot, ringConfig });

// Mock OpenCode config
const mockConfig = {
  agent: {},
  command: {},
  permission: {},
};

(async () => {
  await handler(mockConfig);

  const agentCount = Object.keys(mockConfig.agent).length;
  const commandCount = Object.keys(mockConfig.command).length;

  console.log('Injected agents:', agentCount);
  console.log('Injected commands/skills:', commandCount);
  console.log('Has code-reviewer:', 'code-reviewer' in mockConfig.agent);
  console.log('Has commit command:', 'commit' in mockConfig.command);
  console.log('Webfetch permission:', mockConfig.permission.webfetch);
})();
"`

**Expected output:**
```
Injected agents: 16
Injected commands/skills: 45
Has code-reviewer: true
Has commit command: true
Webfetch permission: allow
```

---

### Task 19: Run Existing Tests

**Step 1: Run plugin tests**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test .opencode/plugin/ 2>&1 | tail -20`

**Expected output:**
```
(tests pass or no test files found)
```

**Step 2: If tests exist, verify all pass**

If tests fail, fix the issues before proceeding.

---

### Task 20: Commit Changes

**Step 1: Stage all new and modified files**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add .opencode/plugin/loaders/ .opencode/plugin/config/ .opencode/plugin/tools/ .opencode/plugin/lifecycle/ .opencode/plugin/ring-unified.ts .opencode/plugin/index.ts`

**Step 2: Check staged files**

Run: `git diff --cached --stat`

**Expected output:**
```
 .opencode/plugin/config/config-handler.ts | 124 +++++++++++++
 .opencode/plugin/config/index.ts          |  17 ++
 .opencode/plugin/config/loader.ts         |  77 ++++++++
 .opencode/plugin/config/types.ts          |  50 ++++++
 .opencode/plugin/index.ts                 |  72 +++++---
 .opencode/plugin/lifecycle/index.ts       |  13 ++
 .opencode/plugin/lifecycle/router.ts      |  99 +++++++++++
 .opencode/plugin/loaders/agent-loader.ts  | 165 ++++++++++++++++++
 .opencode/plugin/loaders/command-loader.ts| 126 ++++++++++++++
 .opencode/plugin/loaders/index.ts         |  24 +++
 .opencode/plugin/loaders/skill-loader.ts  | 140 +++++++++++++++
 .opencode/plugin/ring-unified.ts          | 227 +++++++++++++++++++++++++
 .opencode/plugin/tools/index.ts           |  76 +++++++++
 13 files changed, 1180 insertions(+), 30 deletions(-)
```

**Step 3: Create commit**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git commit -m "$(cat <<'EOF'
feat(plugin): create unified plugin entry point with config injection

- Add component loaders for agents (16), skills (29), commands (16)
- Create config handler to inject Ring components into OpenCode config
- Register ring_doubt tool via tool property
- Add lifecycle router for event handling
- Create unified plugin matching oh-my-opencode pattern
- Update index.ts with unified default export
- Maintain backward compatibility with legacy plugin exports

This enables Ring to fully integrate with OpenCode's plugin system
using the same registration pattern as oh-my-opencode.
EOF
)"`

**Step 4: Verify commit**

Run: `git log -1 --oneline`

**Expected output:**
```
<hash> feat(plugin): create unified plugin entry point with config injection
```

---

## Summary

This plan implements the Unified Plugin Entry Point with these components:

### Part 1: Component Loaders (Tasks 1-4)
- **agent-loader.ts**: Loads 16 agents from `.opencode/agent/*.md`
- **skill-loader.ts**: Loads 29 skills from `.opencode/skill/*/SKILL.md`
- **command-loader.ts**: Loads 16 commands from `.opencode/command/*.md`
- **loaders/index.ts**: Central export for all loaders

### Part 2: Configuration Handler (Tasks 5-8)
- **config/types.ts**: Ring configuration types
- **config/loader.ts**: JSONC config file loader
- **config/config-handler.ts**: Injects components into OpenCode config
- **config/index.ts**: Central export for config module

### Part 3: Unified Plugin (Tasks 9-13)
- **tools/index.ts**: Ring tools (ring_doubt) registration
- **lifecycle/router.ts**: Event routing to hooks
- **ring-unified.ts**: Main unified plugin entry point
- **index.ts**: Updated exports with default unified plugin

### Part 4: Verification (Tasks 14-20)
- Code review checkpoint
- TypeScript build verification
- Component loader tests
- Unified plugin tests
- Config handler tests
- Commit changes

**Key Files Created:**
```
.opencode/plugin/
├── loaders/
│   ├── agent-loader.ts      # Load agents from .opencode/agent/
│   ├── skill-loader.ts      # Load skills from .opencode/skill/
│   ├── command-loader.ts    # Load commands from .opencode/command/
│   └── index.ts
├── config/
│   ├── types.ts             # Ring config types
│   ├── loader.ts            # Config file loader
│   ├── config-handler.ts    # OpenCode config injection
│   └── index.ts
├── tools/
│   └── index.ts             # ring_doubt tool
├── lifecycle/
│   ├── router.ts            # Event routing
│   └── index.ts
├── ring-unified.ts          # Unified plugin entry point
└── index.ts                 # Updated main export
```

**Total estimated time:** 1.5-2 hours for complete implementation

**Integration Points:**
1. OpenCode calls `default` export from plugin/index.ts
2. Plugin returns `Hooks` object with tool, config, event handlers
3. Config handler injects 16 agents, 29 skills, 16 commands
4. Event handler routes lifecycle events to Ring hooks
5. ring_doubt tool available to AI for user interaction
