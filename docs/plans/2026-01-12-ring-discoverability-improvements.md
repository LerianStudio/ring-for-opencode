# Implementation Plan: Ring Discoverability & Usability Improvements

**Created:** 2026-01-12
**Goal:** Implement 7 recommendations for improving Ring discoverability, developer experience, and quality assurance.
**Estimated Total Time:** 4-6 hours

---

## Architecture Overview

```
ring-for-opencode/
├── src/cli/                    # CLI commands (ring install, ring doctor)
│   ├── index.ts               # CLI entry point (add help, init commands)
│   ├── install.ts             # Installation logic
│   └── doctor/                # Health check system
├── plugin/                     # OpenCode plugin system
│   ├── loaders/               # Component loaders (validation needed)
│   │   ├── command-loader.ts  # Command loading + validation
│   │   └── skill-loader.ts    # Skill loading
│   └── hooks/                 # Hook system (extension docs needed)
├── assets/                     # Declarative components
│   ├── command/               # Command definitions
│   ├── skill/                 # Skill definitions
│   └── agent/                 # Agent definitions
├── __tests__/                  # Test suite
│   └── plugin/                # Plugin tests
└── docs/                       # Documentation
    └── GETTING_STARTED.md     # New tutorial
```

**Tech Stack:** TypeScript, Bun, Zod (validation), Commander (CLI), @clack/prompts (TUI)

**Prerequisites:**
- Bun installed (`bun --version`)
- Project cloned and dependencies installed (`bun install`)

---

## Historical Precedent

*No previous failures related to these features in artifact index.*

---

## Task Summary

| # | Task | Priority | Est. Time | Agent |
|---|------|----------|-----------|-------|
| 1 | Add `/ring:help` command | High | 30 min | backend-engineer-typescript |
| 2 | Document hook extension pattern | Medium | 25 min | backend-engineer-typescript |
| 3 | Add command reference validation | Medium | 35 min | backend-engineer-typescript |
| 4 | Create getting started tutorial | Medium | 30 min | backend-engineer-typescript |
| 5 | Add integration tests for skill workflows | Medium | 45 min | qa-analyst |
| 6 | Auto-generate config template on first run | Low | 25 min | backend-engineer-typescript |
| 7 | Add opt-in telemetry for skill usage | Low | 40 min | backend-engineer-typescript |

---

## Batch 1: High Priority (Discoverability)

### Task 1.1: Add `/ring:help` Command

**File:** `assets/command/help.md`

**Goal:** Create a discoverable help command that lists all available Ring skills, commands, and agents.

**Complete Code:**

```markdown
---
description: List all available Ring skills, commands, and agents with usage examples
agent: write-plan
subtask: false
---

# Ring Help

Display comprehensive help for Ring's capabilities.

## Process

1. **List Skills** - Show all available skills with descriptions
2. **List Commands** - Show all slash commands with usage
3. **List Agents** - Show all available agents with their roles
4. **Show Examples** - Provide quick-start examples

## Output Format

Present the information in organized sections:

### Available Skills

Query the skill tool to list all skills matching `ring:*` pattern:
- `ring:test-driven-development` - RED-GREEN-REFACTOR methodology
- `ring:brainstorming` - Socratic design refinement  
- `ring:requesting-code-review` - Parallel 5-reviewer code review
- `ring:systematic-debugging` - 4-phase debugging methodology
- `ring:executing-plans` - Batch task execution with checkpoints
- `ring:writing-plans` - Comprehensive implementation planning
- `ring:exploring-codebase` - Two-phase codebase exploration
- `ring:dispatching-parallel-agents` - Concurrent agent investigation
- `ring:verification-before-completion` - Evidence-first completion
- `ring:handoff-tracking` - Session transition documents
- `ring:using-git-worktrees` - Isolated workspace creation
- `ring:defense-in-depth` - Multi-layer validation pattern
- `ring:root-cause-tracing` - Backward call-chain tracing
- `ring:linting-codebase` - Parallel lint fixing
- `ring:interviewing-user` - Requirements gathering
- `ring:finishing-a-development-branch` - Branch completion workflow

### Available Commands

List all `/ring:*` commands:
- `/ring:brainstorm` - Start a design refinement session
- `/ring:codereview` - Dispatch 5 parallel reviewers
- `/ring:commit` - Create atomic commits with intelligent grouping
- `/ring:lint` - Fix lint issues in parallel
- `/ring:worktree` - Create isolated git worktree
- `/ring:explore-codebase` - Autonomously explore codebase
- `/ring:interview-me` - Gather requirements through questions
- `/ring:write-plan` - Create implementation plan
- `/ring:execute-plan` - Execute a saved plan
- `/ring:create-handoff` - Create session handoff document
- `/ring:resume-handoff` - Resume from handoff document
- `/ring:dev-cycle` - Start 6-gate development workflow
- `/ring:dev-status` - Check dev-cycle status
- `/ring:dev-report` - Generate dev-cycle report
- `/ring:dev-cancel` - Cancel active dev-cycle
- `/ring:dev-refactor` - Generate refactoring tasks

### Available Agents

List specialized agents:
- `@code-reviewer` - Code quality, architecture, design patterns
- `@business-logic-reviewer` - Business rules, domain correctness
- `@security-reviewer` - Security vulnerabilities, data protection
- `@test-reviewer` - Test quality, coverage, anti-patterns
- `@nil-safety-reviewer` - Nil/null pointer safety
- `@codebase-explorer` - Codebase exploration and analysis
- `@write-plan` - Implementation planning
- `@backend-engineer-golang` - Go backend development
- `@backend-engineer-typescript` - TypeScript backend development
- `@frontend-engineer` - Frontend development
- `@frontend-designer` - UI/UX design
- `@frontend-bff-engineer-typescript` - BFF/API layer
- `@devops-engineer` - Infrastructure and CI/CD
- `@sre` - Reliability engineering
- `@qa-analyst` - Quality assurance

### Quick Start Examples

```
# Start a design session
/ring:brainstorm I want to add user authentication

# Get code reviewed
/ring:codereview

# Create a plan for a feature
/ring:write-plan Add payment processing

# Learn TDD workflow
Load the test-driven-development skill
```

## $ARGUMENTS

Optional filter to show specific category:
- `skills` - Show only skills
- `commands` - Show only commands  
- `agents` - Show only agents
- (no argument) - Show everything
```

**Verification:**

```bash
# Check file exists and has valid frontmatter
cat assets/command/help.md | head -10

# Expected output:
# ---
# description: List all available Ring skills, commands, and agents with usage examples
# agent: write-plan
# subtask: false
# ---
```

**Failure Recovery:**
- If frontmatter parsing fails: Check for proper YAML syntax (no tabs, proper quoting)
- If command doesn't appear: Run `ring doctor` to check command loading

---

### Task 1.2: Add CLI `ring help` Subcommand

**File:** `src/cli/index.ts`

**Goal:** Add a `ring help` command that outputs skill/command/agent lists to terminal.

**Changes to make:**

```typescript
// Add after line 89 (after version command, before program.parse())

program
  .command("help")
  .alias("list")
  .description("List available Ring skills, commands, and agents")
  .option("--skills", "Show only skills")
  .option("--commands", "Show only commands")
  .option("--agents", "Show only agents")
  .option("--json", "Output in JSON format")
  .action(async (options) => {
    const exitCode = await help({
      skills: options.skills ?? false,
      commands: options.commands ?? false,
      agents: options.agents ?? false,
      json: options.json ?? false,
    })
    process.exit(exitCode)
  })
```

**New File:** `src/cli/help.ts`

```typescript
import color from "picocolors"
import { existsSync, readdirSync, readFileSync } from "node:fs"
import { join, basename } from "node:path"

export interface HelpOptions {
  skills: boolean
  commands: boolean
  agents: boolean
  json: boolean
}

interface ComponentInfo {
  name: string
  description: string
  type: "skill" | "command" | "agent"
}

function getAssetsPath(): string {
  // Try to find assets relative to this file's location
  const possiblePaths = [
    join(__dirname, "..", "..", "assets"),
    join(__dirname, "..", "..", "..", "assets"),
    join(process.cwd(), "assets"),
    join(process.cwd(), "node_modules", "ring-opencode", "assets"),
  ]

  for (const p of possiblePaths) {
    if (existsSync(p)) return p
  }

  throw new Error("Could not find Ring assets directory")
}

function parseFrontmatter(content: string): { description?: string } {
  const match = content.match(/^---\n([\s\S]*?)\n---/)
  if (!match) return {}

  const yaml = match[1]
  const descMatch = yaml.match(/description:\s*(.+)/)
  return { description: descMatch?.[1]?.replace(/^["']|["']$/g, "") }
}

function loadSkills(assetsPath: string): ComponentInfo[] {
  const skillsDir = join(assetsPath, "skill")
  if (!existsSync(skillsDir)) return []

  const skills: ComponentInfo[] = []
  const entries = readdirSync(skillsDir, { withFileTypes: true })

  for (const entry of entries) {
    if (!entry.isDirectory()) continue
    if (entry.name === "shared-patterns") continue

    const skillFile = join(skillsDir, entry.name, "SKILL.md")
    if (!existsSync(skillFile)) continue

    try {
      const content = readFileSync(skillFile, "utf-8")
      const { description } = parseFrontmatter(content)
      skills.push({
        name: entry.name,
        description: description || `Ring skill: ${entry.name}`,
        type: "skill",
      })
    } catch {
      // Skip unparseable files
    }
  }

  return skills.sort((a, b) => a.name.localeCompare(b.name))
}

function loadCommands(assetsPath: string): ComponentInfo[] {
  const commandsDir = join(assetsPath, "command")
  if (!existsSync(commandsDir)) return []

  const commands: ComponentInfo[] = []
  const entries = readdirSync(commandsDir)

  for (const file of entries) {
    if (!file.endsWith(".md")) continue

    const commandPath = join(commandsDir, file)
    const commandName = basename(file, ".md")

    try {
      const content = readFileSync(commandPath, "utf-8")
      const { description } = parseFrontmatter(content)
      commands.push({
        name: commandName,
        description: description || `Ring command: ${commandName}`,
        type: "command",
      })
    } catch {
      // Skip unparseable files
    }
  }

  return commands.sort((a, b) => a.name.localeCompare(b.name))
}

function loadAgents(assetsPath: string): ComponentInfo[] {
  const agentsDir = join(assetsPath, "agent")
  if (!existsSync(agentsDir)) return []

  const agents: ComponentInfo[] = []
  const entries = readdirSync(agentsDir)

  for (const file of entries) {
    if (!file.endsWith(".md")) continue

    const agentPath = join(agentsDir, file)
    const agentName = basename(file, ".md")

    try {
      const content = readFileSync(agentPath, "utf-8")
      const { description } = parseFrontmatter(content)
      agents.push({
        name: agentName,
        description: description || `Ring agent: ${agentName}`,
        type: "agent",
      })
    } catch {
      // Skip unparseable files
    }
  }

  return agents.sort((a, b) => a.name.localeCompare(b.name))
}

function printSection(title: string, items: ComponentInfo[], prefix: string): void {
  console.log()
  console.log(color.bold(color.cyan(title)))
  console.log(color.dim("─".repeat(50)))

  for (const item of items) {
    const name = color.green(`${prefix}${item.name}`)
    const desc = color.dim(item.description)
    console.log(`  ${name}`)
    console.log(`    ${desc}`)
  }
}

export async function help(options: HelpOptions): Promise<number> {
  try {
    const assetsPath = getAssetsPath()
    const showAll = !options.skills && !options.commands && !options.agents

    const skills = (showAll || options.skills) ? loadSkills(assetsPath) : []
    const commands = (showAll || options.commands) ? loadCommands(assetsPath) : []
    const agents = (showAll || options.agents) ? loadAgents(assetsPath) : []

    if (options.json) {
      const output: Record<string, ComponentInfo[]> = {}
      if (skills.length > 0) output.skills = skills
      if (commands.length > 0) output.commands = commands
      if (agents.length > 0) output.agents = agents
      console.log(JSON.stringify(output, null, 2))
      return 0
    }

    console.log()
    console.log(color.bold(color.bgCyan(" Ring for OpenCode ")))
    console.log(color.dim("Proven software engineering practices"))

    if (skills.length > 0) {
      printSection(`Skills (${skills.length})`, skills, "ring:")
    }

    if (commands.length > 0) {
      printSection(`Commands (${commands.length})`, commands, "/ring:")
    }

    if (agents.length > 0) {
      printSection(`Agents (${agents.length})`, agents, "@")
    }

    console.log()
    console.log(color.dim("Use skills: ") + color.cyan("Load the <skill-name> skill"))
    console.log(color.dim("Use commands: ") + color.cyan("/<command-name>"))
    console.log(color.dim("Use agents: ") + color.cyan("@<agent-name> in your prompt"))
    console.log()

    return 0
  } catch (error) {
    console.error(color.red("Error:"), error instanceof Error ? error.message : error)
    return 1
  }
}
```

**Update imports in `src/cli/index.ts`:**

```typescript
// Add at top with other imports (after line 6)
import { help } from "./help"
import type { HelpOptions } from "./help"
```

**Verification:**

```bash
# Build the CLI
bun run build

# Test the help command
./dist/cli/index.js help

# Expected output: Lists of skills, commands, agents with descriptions

# Test with filters
./dist/cli/index.js help --skills
./dist/cli/index.js help --json
```

**Failure Recovery:**
- If assets not found: Check `getAssetsPath()` paths match your install location
- If build fails: Run `bun run typecheck` for type errors

---

## Batch 1 Review Checkpoint

After completing Tasks 1.1 and 1.2:

```bash
# Verify all files exist
ls -la assets/command/help.md
ls -la src/cli/help.ts

# Run tests
bun test

# Build and verify
bun run build
./dist/cli/index.js help --skills | head -20
```

**Review Criteria:**
- [ ] `/ring:help` command file has valid frontmatter
- [ ] CLI `ring help` command runs without errors
- [ ] JSON output is valid JSON
- [ ] All skills/commands/agents are listed

---

## Batch 2: Medium Priority (Developer Experience)

### Task 2.1: Document Hook Extension Pattern

**File:** `docs/EXTENDING_HOOKS.md`

**Goal:** Provide clear documentation for extending the hook system with custom hooks.

**Complete Code:**

```markdown
# Extending Ring Hooks

This guide explains how to create custom hooks for Ring.

## Overview

Ring's hook system uses a middleware pattern where hooks execute in priority order during OpenCode lifecycle events.

## Hook Lifecycle Events

| Event | When Triggered | Common Use Cases |
|-------|----------------|------------------|
| `session.created` | Session starts | Initialize state, inject context |
| `session.idle` | Session becomes idle | Send notifications |
| `session.error` | Error occurs | Log errors, notify user |
| `session.compacting` | Context recovery | Inject compact context |
| `chat.message` | User sends message | Preprocess messages |
| `chat.params` | Before model call | Modify parameters |
| `tool.before` | Before tool runs | Validate, log |
| `tool.after` | After tool runs | Process results |
| `todo.updated` | Todo list changes | Track progress |
| `event` | Generic events | Custom handling |

## Creating a Custom Hook

### Step 1: Define the Hook Factory

Create a new file in `plugin/hooks/factories/`:

```typescript
// plugin/hooks/factories/my-custom-hook.ts
import type { Hook, HookContext, HookFactory, HookOutput, HookResult } from "../types.js"

/**
 * Configuration for your custom hook.
 */
export interface MyCustomHookConfig {
  enabled?: boolean
  customOption?: string
}

const DEFAULT_CONFIG: Required<MyCustomHookConfig> = {
  enabled: true,
  customOption: "default-value",
}

/**
 * Factory function that creates the hook.
 */
export const createMyCustomHook: HookFactory<MyCustomHookConfig> = (
  config?: MyCustomHookConfig,
): Hook => {
  const cfg = { ...DEFAULT_CONFIG, ...config }

  return {
    // Unique identifier (must be added to HookName type)
    name: "my-custom-hook" as const,

    // Which lifecycle events this hook responds to
    lifecycles: ["session.created", "chat.params"],

    // Lower = runs earlier (default: 100)
    priority: 50,

    // Whether hook is active
    enabled: cfg.enabled,

    // The hook implementation
    async execute(ctx: HookContext, output: HookOutput): Promise<HookResult> {
      try {
        // Access context information
        const { sessionId, directory, lifecycle, chainData } = ctx

        // Modify output to inject content
        output.system = output.system ?? []
        output.system.push(`Custom content from my-hook: ${cfg.customOption}`)

        // Return success with optional data for next hook
        return {
          success: true,
          data: {
            myHookRan: true,
            timestamp: Date.now(),
          },
        }
      } catch (error) {
        return {
          success: false,
          error: error instanceof Error ? error.message : String(error),
        }
      }
    },
  }
}

/**
 * Registry entry for automatic discovery.
 */
export const myCustomHookEntry = {
  name: "my-custom-hook" as const,
  factory: createMyCustomHook,
  defaultEnabled: true,
  description: "Description of what this hook does",
}
```

### Step 2: Register the Hook

Add your hook type to `plugin/hooks/types.ts`:

```typescript
// Add to HookName type
export type HookName =
  | "session-start"
  | "context-injection"
  // ... existing hooks ...
  | "my-custom-hook"  // Add your hook name
```

Register in `plugin/ring-unified.ts`:

```typescript
import { createMyCustomHook } from "./hooks/factories/my-custom-hook.js"

// In the initialization section, after other hooks:
if (!isDisabled("my-custom-hook")) {
  hookRegistry.register(createMyCustomHook(config.hooks?.["my-custom-hook"]))
}
```

### Step 3: Add Configuration Schema (Optional)

If your hook needs configurable options, add to `plugin/config/schema.ts`:

```typescript
// Add to HookNameSchema
export const HookNameSchema = z.enum([
  // ... existing hooks ...
  "my-custom-hook",
])
```

And update `config.template.jsonc`:

```jsonc
{
  "hooks": {
    "my-custom-hook": {
      "enabled": true,
      "customOption": "custom-value"
    }
  }
}
```

## Hook Execution Flow

```
Lifecycle Event Triggered
         │
         ▼
Hook Registry filters by lifecycle
         │
         ▼
Sort by priority (ascending)
         │
         ├─→ Hook A (priority: 10)
         │   └─→ Returns { success: true, data: {...} }
         │
         ├─→ Hook B (priority: 50) ← receives chainData from A
         │   └─→ Returns { success: true, stopChain: true }
         │
         └─→ Hook C (priority: 100) ← SKIPPED (chain stopped)
```

## Best Practices

1. **Keep hooks focused** - One responsibility per hook
2. **Use appropriate priority** - Lower for critical, higher for optional
3. **Handle errors gracefully** - Return `{ success: false, error }` instead of throwing
4. **Document configuration** - Explain all options in JSDoc
5. **Test thoroughly** - Write unit tests for all code paths

## Testing Hooks

```typescript
// __tests__/plugin/hooks/my-custom-hook.test.ts
import { describe, expect, test } from "bun:test"
import { createMyCustomHook } from "../../../plugin/hooks/factories/my-custom-hook.js"

describe("my-custom-hook", () => {
  test("injects custom content", async () => {
    const hook = createMyCustomHook({ customOption: "test-value" })
    const output = { system: [] }

    const result = await hook.execute(
      {
        sessionId: "test",
        directory: "/tmp",
        lifecycle: "session.created",
      },
      output,
    )

    expect(result.success).toBe(true)
    expect(output.system).toContain("Custom content from my-hook: test-value")
  })
})
```

## Disabling Built-in Hooks

In `config.jsonc`:

```jsonc
{
  "disabled_hooks": ["notification", "my-custom-hook"]
}
```

Or programmatically:

```typescript
hookRegistry.disable("my-custom-hook")
```
```

**Verification:**

```bash
# File exists
cat docs/EXTENDING_HOOKS.md | head -20

# Check for code examples
grep -c "```typescript" docs/EXTENDING_HOOKS.md
# Expected: 4+ code blocks
```

---

### Task 2.2: Add Command Reference Validation

**File:** `plugin/loaders/command-loader.ts`

**Goal:** Warn when commands reference non-existent agents.

**Changes to `loadCommandsFromDir` function (around line 88):**

```typescript
/**
 * Validation results for command loading.
 */
export interface CommandValidationResult {
  warnings: string[]
  errors: string[]
}

/**
 * Known valid agent names for validation.
 */
const KNOWN_AGENTS = new Set([
  "code-reviewer",
  "security-reviewer",
  "business-logic-reviewer",
  "test-reviewer",
  "nil-safety-reviewer",
  "codebase-explorer",
  "write-plan",
  "backend-engineer-golang",
  "backend-engineer-typescript",
  "frontend-engineer",
  "frontend-designer",
  "frontend-bff-engineer-typescript",
  "devops-engineer",
  "sre",
  "qa-analyst",
  "prompt-quality-reviewer",
])

/**
 * Validate command references.
 */
function validateCommandAgent(
  commandName: string,
  agentName: string | undefined,
  customAgents: Set<string>,
): string | null {
  if (!agentName) return null

  // Check built-in agents
  if (KNOWN_AGENTS.has(agentName)) return null

  // Check custom agents (loaded from .opencode/agent/)
  if (customAgents.has(agentName)) return null

  return `Command "${commandName}" references unknown agent "${agentName}"`
}

/**
 * Load commands from a directory with validation.
 */
function loadCommandsFromDir(
  commandsDir: string,
  disabledCommands: Set<string>,
  customAgents: Set<string> = new Set(),
  validation?: CommandValidationResult,
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

        // Validate agent reference
        if (validation && data.agent) {
          const warning = validateCommandAgent(commandName, data.agent, customAgents)
          if (warning) {
            validation.warnings.push(warning)
            if (process.env.RING_DEBUG === "true") {
              console.warn(`[ring] Warning: ${warning}`)
            }
          }
        }

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
        if (validation) {
          validation.errors.push(`Failed to parse command "${commandName}": ${error}`)
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
```

**Update `loadRingCommands` signature:**

```typescript
/**
 * Load Ring commands with validation.
 *
 * @param pluginRoot - Path to the plugin directory
 * @param projectRoot - Path to the user's project directory
 * @param disabledCommands - List of command names to skip
 * @param customAgents - Optional set of custom agent names for validation
 * @returns Object with commands and validation results
 */
export function loadRingCommands(
  pluginRoot: string,
  projectRoot: string,
  disabledCommands: string[] = [],
  customAgents: Set<string> = new Set(),
): { commands: Record<string, CommandConfig>; validation: CommandValidationResult } {
  const disabledSet = new Set(disabledCommands)
  const validation: CommandValidationResult = { warnings: [], errors: [] }

  // Load Ring's built-in commands from assets/command/
  const builtInDir = join(pluginRoot, "assets", "command")
  const builtInCommands = loadCommandsFromDir(builtInDir, disabledSet, customAgents, validation)

  // Load user's custom commands from .opencode/command/
  const userDir = join(projectRoot, ".opencode", "command")
  const userCommands = loadCommandsFromDir(userDir, disabledSet, customAgents, validation)

  // Merge with user's taking priority
  const merged: Record<string, CommandConfig> = Object.create(null)
  Object.assign(merged, builtInCommands)
  Object.assign(merged, userCommands)

  return { commands: merged, validation }
}
```

**Verification:**

```bash
# Run type check
bun run typecheck

# Test with debug enabled
RING_DEBUG=true bun test __tests__/plugin/

# Create a test command with invalid agent
cat > /tmp/test-cmd.md << 'EOF'
---
description: Test command
agent: non-existent-agent
---
# Test
EOF

# Verify warning is generated (in tests or debug output)
```

---

### Task 2.3: Create Getting Started Tutorial

**File:** `docs/GETTING_STARTED.md`

**Complete Code:**

```markdown
# Getting Started with Ring for OpenCode

Ring is a skills library that enforces proven software engineering practices through mandatory workflows, parallel code review, and systematic development.

## Installation

### Prerequisites

- [OpenCode](https://opencode.ai) installed
- [Bun](https://bun.sh) runtime

### Quick Install

```bash
# Install Ring globally
bun add -g ring-opencode

# Or add to your project
bun add ring-opencode
```

### Configure Ring

```bash
# Run the installer
ring install

# Verify installation
ring doctor
```

## Your First Commands

### 1. Get Help

```bash
# List all available skills, commands, and agents
ring help

# Or in OpenCode chat
/ring:help
```

### 2. Start a Design Session

```
/ring:brainstorm I want to add user authentication to my app
```

Ring will guide you through:
- Clarifying requirements
- Exploring alternatives
- Validating design decisions

### 3. Get Code Reviewed

After making changes:

```
/ring:codereview
```

This dispatches 5 parallel reviewers:
- Code quality reviewer
- Business logic reviewer
- Security reviewer
- Test reviewer
- Nil-safety reviewer

### 4. Create Atomic Commits

```
/ring:commit
```

Ring will:
- Group related changes
- Generate conventional commit messages
- Verify before committing

## Core Workflows

### Test-Driven Development

Load the TDD skill:

```
Use the test-driven-development skill
```

Ring enforces:
1. **RED** - Write a failing test first
2. **GREEN** - Minimal code to pass
3. **REFACTOR** - Clean up while tests pass

### Systematic Debugging

When facing a bug:

```
Use the systematic-debugging skill
```

Ring guides you through:
1. Root cause investigation
2. Pattern analysis
3. Hypothesis testing
4. Implementation

### Implementation Planning

For complex features:

```
/ring:write-plan Add payment processing with Stripe
```

Ring creates a comprehensive plan with:
- Exact file paths
- Complete code examples
- Verification commands
- Failure recovery steps

## Available Agents

Invoke agents with `@` mention:

| Agent | Role |
|-------|------|
| `@code-reviewer` | Code quality and architecture |
| `@security-reviewer` | Security vulnerabilities |
| `@test-reviewer` | Test coverage and quality |
| `@backend-engineer-golang` | Go development |
| `@backend-engineer-typescript` | TypeScript development |
| `@devops-engineer` | CI/CD and infrastructure |
| `@qa-analyst` | Quality assurance |

Example:

```
@backend-engineer-golang help me implement a REST API for user management
```

## Configuration

Ring configuration lives in:
- Project: `.opencode/ring.jsonc` or `.ring/config.jsonc`
- User: `~/.config/opencode/ring/config.jsonc`

### Example Configuration

```jsonc
{
  "$schema": "https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-config.schema.json",

  // Disable specific hooks
  "disabled_hooks": [],

  // Background task settings
  "background_tasks": {
    "defaultConcurrency": 3,
    "taskTimeoutMs": 1800000
  },

  // Desktop notifications
  "notifications": {
    "enabled": true,
    "onIdle": true,
    "onBackgroundComplete": true
  }
}
```

## Troubleshooting

### Ring commands not appearing

```bash
ring doctor --verbose
```

### Configuration issues

```bash
ring doctor --category configuration
```

### Debug mode

```bash
RING_DEBUG=true opencode
```

## Next Steps

1. **Explore skills**: `ring help --skills`
2. **Read the architecture**: [ARCHITECTURE.md](../ARCHITECTURE.md)
3. **Extend Ring**: [EXTENDING_HOOKS.md](./EXTENDING_HOOKS.md)

## Community

- Report issues: https://github.com/fredcamaral/ring-for-opencode/issues
- Discussions: https://github.com/fredcamaral/ring-for-opencode/discussions
```

**Verification:**

```bash
# Check file exists
cat docs/GETTING_STARTED.md | head -30

# Verify all links are valid
grep -E '\[.*\]\(.*\.md\)' docs/GETTING_STARTED.md
```

---

## Batch 2 Review Checkpoint

After completing Tasks 2.1-2.3:

```bash
# Verify documentation files
ls -la docs/EXTENDING_HOOKS.md docs/GETTING_STARTED.md

# Type check after command-loader changes
bun run typecheck

# Run existing tests
bun test
```

**Review Criteria:**
- [ ] Hook documentation has complete code examples
- [ ] Command validation logs warnings for invalid agents
- [ ] Getting started tutorial covers all major features
- [ ] No type errors introduced

---

## Batch 3: Testing (Medium Priority)

### Task 3.1: Add Integration Tests for Skill Workflows

**File:** `__tests__/plugin/skills/workflow.integration.test.ts`

**Goal:** Test full skill loading and execution workflows.

**Complete Code:**

```typescript
import { describe, expect, test, beforeAll, afterAll } from "bun:test"
import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"

import { loadRingSkills, countRingSkills } from "../../../plugin/loaders/skill-loader.js"
import { loadRingCommands, countRingCommands } from "../../../plugin/loaders/command-loader.js"
import { loadRingAgents, countRingAgents } from "../../../plugin/loaders/agent-loader.js"

describe("Skill Workflow Integration", () => {
  let tmpDir: string
  let pluginRoot: string

  beforeAll(() => {
    // Create temp project directory
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "ring-skill-test-"))

    // Plugin root is the actual assets location
    pluginRoot = path.join(__dirname, "..", "..", "..", "assets")

    // Verify assets exist
    if (!fs.existsSync(pluginRoot)) {
      pluginRoot = path.join(__dirname, "..", "..", "..", "..", "assets")
    }
  })

  afterAll(() => {
    // Cleanup
    fs.rmSync(tmpDir, { recursive: true, force: true })
  })

  describe("Skill Loading", () => {
    test("loads all built-in skills", () => {
      const skills = loadRingSkills(
        path.dirname(pluginRoot),
        tmpDir,
        [],
      )

      // Should have multiple skills
      expect(Object.keys(skills).length).toBeGreaterThan(10)

      // Verify key skills exist
      expect(skills["ring:test-driven-development"]).toBeDefined()
      expect(skills["ring:brainstorming"]).toBeDefined()
      expect(skills["ring:systematic-debugging"]).toBeDefined()
      expect(skills["ring:requesting-code-review"]).toBeDefined()
    })

    test("respects disabled_skills configuration", () => {
      const allSkills = loadRingSkills(
        path.dirname(pluginRoot),
        tmpDir,
        [],
      )

      const filteredSkills = loadRingSkills(
        path.dirname(pluginRoot),
        tmpDir,
        ["test-driven-development", "brainstorming"],
      )

      expect(Object.keys(filteredSkills).length).toBe(
        Object.keys(allSkills).length - 2
      )
      expect(filteredSkills["ring:test-driven-development"]).toBeUndefined()
      expect(filteredSkills["ring:brainstorming"]).toBeUndefined()
    })

    test("user skills override built-in skills", () => {
      // Create user skill directory
      const userSkillDir = path.join(tmpDir, ".opencode", "skill", "test-driven-development")
      fs.mkdirSync(userSkillDir, { recursive: true })

      // Create custom skill file
      fs.writeFileSync(
        path.join(userSkillDir, "SKILL.md"),
        `---
description: Custom TDD skill
---
# Custom TDD
User override.
`
      )

      const skills = loadRingSkills(
        path.dirname(pluginRoot),
        tmpDir,
        [],
      )

      // User skill should override
      expect(skills["ring:test-driven-development"]?.description).toBe("Custom TDD skill")
    })
  })

  describe("Command Loading", () => {
    test("loads all built-in commands", () => {
      const { commands } = loadRingCommands(
        path.dirname(pluginRoot),
        tmpDir,
        [],
      )

      // Should have multiple commands
      expect(Object.keys(commands).length).toBeGreaterThan(10)

      // Verify key commands exist
      expect(commands["ring:commit"]).toBeDefined()
      expect(commands["ring:codereview"]).toBeDefined()
      expect(commands["ring:brainstorm"]).toBeDefined()
    })

    test("validates agent references", () => {
      // Create command with invalid agent
      const userCmdDir = path.join(tmpDir, ".opencode", "command")
      fs.mkdirSync(userCmdDir, { recursive: true })

      fs.writeFileSync(
        path.join(userCmdDir, "test-invalid.md"),
        `---
description: Test command
agent: non-existent-agent
---
# Test
`
      )

      const { commands, validation } = loadRingCommands(
        path.dirname(pluginRoot),
        tmpDir,
        [],
      )

      expect(validation.warnings.length).toBeGreaterThan(0)
      expect(validation.warnings[0]).toContain("non-existent-agent")
    })
  })

  describe("Agent Loading", () => {
    test("loads all built-in agents", () => {
      const agents = loadRingAgents(
        path.dirname(pluginRoot),
        tmpDir,
        [],
      )

      // Should have multiple agents
      expect(Object.keys(agents).length).toBeGreaterThan(10)

      // Verify key agents exist
      expect(agents["code-reviewer"]).toBeDefined()
      expect(agents["backend-engineer-golang"]).toBeDefined()
      expect(agents["devops-engineer"]).toBeDefined()
    })
  })

  describe("Full Workflow", () => {
    test("complete skill discovery to loading workflow", () => {
      // Count components
      const skillCount = countRingSkills(path.dirname(pluginRoot), tmpDir)
      const commandCount = countRingCommands(path.dirname(pluginRoot), tmpDir)
      const agentCount = countRingAgents(path.dirname(pluginRoot), tmpDir)

      // Verify counts match actual loading
      const skills = loadRingSkills(path.dirname(pluginRoot), tmpDir, [])
      const { commands } = loadRingCommands(path.dirname(pluginRoot), tmpDir, [])
      const agents = loadRingAgents(path.dirname(pluginRoot), tmpDir, [])

      // Note: counts may differ from loaded due to namespacing
      expect(skillCount).toBeGreaterThan(0)
      expect(commandCount).toBeGreaterThan(0)
      expect(agentCount).toBeGreaterThan(0)

      // Verify loaded components are usable
      for (const [name, config] of Object.entries(skills)) {
        expect(name).toMatch(/^ring:/)
        expect(config.description).toBeDefined()
      }
    })

    test("skill dependencies are loadable", () => {
      const skills = loadRingSkills(path.dirname(pluginRoot), tmpDir, [])

      // Dev-cycle depends on other skills
      expect(skills["ring:dev-cycle"]).toBeDefined()
      expect(skills["ring:dev-implementation"]).toBeDefined()
      expect(skills["ring:dev-testing"]).toBeDefined()
      expect(skills["ring:dev-validation"]).toBeDefined()
    })
  })
})

describe("Skill Content Validation", () => {
  test("all skills have required frontmatter", () => {
    const skillsDir = path.join(__dirname, "..", "..", "..", "assets", "skill")

    if (!fs.existsSync(skillsDir)) {
      // Try alternate path
      const altPath = path.join(__dirname, "..", "..", "..", "..", "assets", "skill")
      if (!fs.existsSync(altPath)) {
        console.warn("Skills directory not found, skipping content validation")
        return
      }
    }

    const entries = fs.readdirSync(skillsDir, { withFileTypes: true })

    for (const entry of entries) {
      if (!entry.isDirectory()) continue
      if (entry.name === "shared-patterns") continue

      const skillFile = path.join(skillsDir, entry.name, "SKILL.md")
      if (!fs.existsSync(skillFile)) {
        console.warn(`Missing SKILL.md in ${entry.name}`)
        continue
      }

      const content = fs.readFileSync(skillFile, "utf-8")

      // Check frontmatter exists
      expect(content.startsWith("---")).toBe(true)
      expect(content.includes("description:")).toBe(true)
    }
  })
})
```

**Verification:**

```bash
# Run the integration tests
bun test __tests__/plugin/skills/workflow.integration.test.ts

# Expected: All tests pass
```

---

## Batch 3 Review Checkpoint

After completing Task 3.1:

```bash
# Run all tests
bun test

# Check test coverage
bun test --coverage
```

**Review Criteria:**
- [ ] All integration tests pass
- [ ] Tests cover skill, command, and agent loading
- [ ] Validation warnings are tested
- [ ] No test flakiness

---

## Batch 4: Low Priority (Polish)

### Task 4.1: Auto-Generate Config Template on First Run

**File:** `src/cli/install.ts`

**Goal:** Automatically create a config file with helpful defaults on first install.

**Changes to `addSchemaToConfig` in `src/cli/config-manager.ts`:**

```typescript
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs"
import { dirname, join } from "node:path"
import { homedir } from "node:os"
import { parse, modify, applyEdits } from "jsonc-parser"

// Template for new config files
const CONFIG_TEMPLATE = `{
  // Ring Configuration
  // Documentation: https://github.com/fredcamaral/ring-for-opencode
  "$schema": "${SCHEMA_URL}",

  // Hooks to disable (uncomment to disable)
  // "disabled_hooks": ["notification"],

  // Agents to disable (uncomment to disable)
  // "disabled_agents": [],

  // Skills to disable (uncomment to disable)
  // "disabled_skills": [],

  // Commands to disable (uncomment to disable)
  // "disabled_commands": [],

  // Background task settings
  "background_tasks": {
    // Number of parallel background tasks (1-10)
    "defaultConcurrency": 3,
    // Task timeout in milliseconds (default: 30 minutes)
    "taskTimeoutMs": 1800000
  },

  // Desktop notifications
  "notifications": {
    "enabled": true,
    "onIdle": true,
    "onError": true,
    "onBackgroundComplete": true
  },

  // Experimental features (use with caution)
  "experimental": {
    "preemptiveCompaction": false,
    "compactionThreshold": 0.8,
    "aggressiveTruncation": false
  }
}
`

export interface SchemaResult {
  success: boolean
  error?: string
  configPath: string
  created: boolean  // New field to indicate if file was created
}

/**
 * Add schema to existing config or create new config with template.
 */
export function addSchemaToConfig(): SchemaResult {
  const detected = detectCurrentConfig()

  // Determine config path
  let configPath: string
  if (detected.configPath) {
    configPath = detected.configPath
  } else {
    // Create in user config directory by default
    configPath = join(homedir(), ".config", "opencode", "ring", "config.jsonc")
  }

  try {
    // Check if file exists
    if (!existsSync(configPath)) {
      // Create directory if needed
      const dir = dirname(configPath)
      if (!existsSync(dir)) {
        mkdirSync(dir, { recursive: true })
      }

      // Write template
      writeFileSync(configPath, CONFIG_TEMPLATE, "utf-8")

      return {
        success: true,
        configPath,
        created: true,
      }
    }

    // File exists - add/update schema
    const content = readFileSync(configPath, "utf-8")
    const parsed = parse(content)

    // Check if schema already exists and matches
    if (parsed.$schema === SCHEMA_URL) {
      return {
        success: true,
        configPath,
        created: false,
      }
    }

    // Add or update schema
    const edits = modify(content, ["$schema"], SCHEMA_URL, {
      formattingOptions: {
        tabSize: 2,
        insertSpaces: true,
      },
    })

    const newContent = applyEdits(content, edits)
    writeFileSync(configPath, newContent, "utf-8")

    return {
      success: true,
      configPath,
      created: false,
    }
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : String(error),
      configPath,
      created: false,
    }
  }
}
```

**Update install.ts to show "created" message:**

```typescript
// In runTuiInstall, after s.stop for schema:
s.stop(
  schemaResult.created
    ? `Config created at ${color.cyan(schemaResult.configPath)}`
    : `Schema added to ${color.cyan(schemaResult.configPath)}`
)
```

**Verification:**

```bash
# Remove existing config
rm -f ~/.config/opencode/ring/config.jsonc

# Run install
./dist/cli/index.js install --no-tui

# Verify config was created with template
cat ~/.config/opencode/ring/config.jsonc

# Expected: Full config template with comments
```

---

### Task 4.2: Add Opt-In Telemetry for Skill Usage

**File:** `plugin/telemetry/index.ts`

**Goal:** Track skill usage patterns to improve Ring (opt-in only).

**Complete Code:**

```typescript
/**
 * Ring Telemetry Module
 *
 * Opt-in telemetry for tracking skill usage patterns.
 * NO personally identifiable information is collected.
 * All data is anonymized and aggregated.
 */

import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs"
import { join } from "node:path"
import { homedir } from "node:os"
import { createHash } from "node:crypto"

/**
 * Telemetry configuration.
 */
export interface TelemetryConfig {
  /** Whether telemetry is enabled (default: false) */
  enabled: boolean
  /** Anonymous installation ID */
  installId?: string
  /** Last sync timestamp */
  lastSync?: number
}

/**
 * Telemetry event types.
 */
export type TelemetryEventType =
  | "skill.loaded"
  | "command.executed"
  | "agent.invoked"
  | "hook.triggered"
  | "error.occurred"

/**
 * Telemetry event data.
 */
export interface TelemetryEvent {
  type: TelemetryEventType
  name: string
  timestamp: number
  properties?: Record<string, string | number | boolean>
}

/**
 * Local telemetry storage.
 */
interface TelemetryStore {
  config: TelemetryConfig
  events: TelemetryEvent[]
}

const TELEMETRY_DIR = join(homedir(), ".config", "opencode", "ring", "telemetry")
const TELEMETRY_FILE = join(TELEMETRY_DIR, "data.json")
const MAX_EVENTS = 1000
const SYNC_INTERVAL = 24 * 60 * 60 * 1000 // 24 hours

/**
 * Generate anonymous installation ID.
 */
function generateInstallId(): string {
  const data = `${homedir()}-${Date.now()}-${Math.random()}`
  return createHash("sha256").update(data).digest("hex").slice(0, 16)
}

/**
 * Load telemetry store from disk.
 */
function loadStore(): TelemetryStore {
  try {
    if (existsSync(TELEMETRY_FILE)) {
      const content = readFileSync(TELEMETRY_FILE, "utf-8")
      return JSON.parse(content)
    }
  } catch {
    // Ignore parse errors
  }

  return {
    config: { enabled: false },
    events: [],
  }
}

/**
 * Save telemetry store to disk.
 */
function saveStore(store: TelemetryStore): void {
  try {
    if (!existsSync(TELEMETRY_DIR)) {
      mkdirSync(TELEMETRY_DIR, { recursive: true })
    }
    writeFileSync(TELEMETRY_FILE, JSON.stringify(store, null, 2), "utf-8")
  } catch {
    // Ignore write errors
  }
}

/**
 * Telemetry client for Ring.
 */
export class Telemetry {
  private store: TelemetryStore
  private enabled: boolean

  constructor() {
    this.store = loadStore()
    this.enabled = this.store.config.enabled

    // Generate install ID if enabled and missing
    if (this.enabled && !this.store.config.installId) {
      this.store.config.installId = generateInstallId()
      saveStore(this.store)
    }
  }

  /**
   * Check if telemetry is enabled.
   */
  isEnabled(): boolean {
    return this.enabled
  }

  /**
   * Enable telemetry (opt-in).
   */
  enable(): void {
    this.enabled = true
    this.store.config.enabled = true

    if (!this.store.config.installId) {
      this.store.config.installId = generateInstallId()
    }

    saveStore(this.store)
  }

  /**
   * Disable telemetry.
   */
  disable(): void {
    this.enabled = false
    this.store.config.enabled = false
    this.store.events = []
    saveStore(this.store)
  }

  /**
   * Track an event.
   */
  track(type: TelemetryEventType, name: string, properties?: Record<string, string | number | boolean>): void {
    if (!this.enabled) return

    const event: TelemetryEvent = {
      type,
      name,
      timestamp: Date.now(),
      properties,
    }

    this.store.events.push(event)

    // Trim old events
    if (this.store.events.length > MAX_EVENTS) {
      this.store.events = this.store.events.slice(-MAX_EVENTS)
    }

    saveStore(this.store)
  }

  /**
   * Track skill usage.
   */
  trackSkill(skillName: string): void {
    this.track("skill.loaded", skillName)
  }

  /**
   * Track command execution.
   */
  trackCommand(commandName: string): void {
    this.track("command.executed", commandName)
  }

  /**
   * Track agent invocation.
   */
  trackAgent(agentName: string): void {
    this.track("agent.invoked", agentName)
  }

  /**
   * Get aggregated statistics (for local display).
   */
  getStats(): Record<string, number> {
    const stats: Record<string, number> = {}

    for (const event of this.store.events) {
      const key = `${event.type}:${event.name}`
      stats[key] = (stats[key] || 0) + 1
    }

    return stats
  }

  /**
   * Get raw event count.
   */
  getEventCount(): number {
    return this.store.events.length
  }

  /**
   * Clear all telemetry data.
   */
  clear(): void {
    this.store.events = []
    saveStore(this.store)
  }
}

/**
 * Singleton telemetry instance.
 */
export const telemetry = new Telemetry()
```

**Add CLI command for telemetry management:**

Add to `src/cli/index.ts`:

```typescript
program
  .command("telemetry")
  .description("Manage anonymous usage telemetry")
  .option("--enable", "Enable telemetry (opt-in)")
  .option("--disable", "Disable telemetry")
  .option("--status", "Show telemetry status")
  .option("--stats", "Show usage statistics")
  .option("--clear", "Clear telemetry data")
  .action(async (options) => {
    const { telemetry } = await import("../../plugin/telemetry/index.js")

    if (options.enable) {
      telemetry.enable()
      console.log(color.green("Telemetry enabled"))
      console.log(color.dim("Thank you for helping improve Ring!"))
    } else if (options.disable) {
      telemetry.disable()
      console.log(color.yellow("Telemetry disabled"))
    } else if (options.clear) {
      telemetry.clear()
      console.log(color.green("Telemetry data cleared"))
    } else if (options.stats) {
      const stats = telemetry.getStats()
      console.log(color.bold("Usage Statistics:"))
      for (const [key, count] of Object.entries(stats).sort((a, b) => b[1] - a[1])) {
        console.log(`  ${key}: ${count}`)
      }
    } else {
      // Default: show status
      console.log(color.bold("Telemetry Status:"))
      console.log(`  Enabled: ${telemetry.isEnabled() ? color.green("yes") : color.yellow("no")}`)
      console.log(`  Events: ${telemetry.getEventCount()}`)
      console.log()
      console.log(color.dim("Ring collects anonymous usage data to improve the tool."))
      console.log(color.dim("No personally identifiable information is collected."))
      console.log()
      console.log(`  Enable:  ${color.cyan("ring telemetry --enable")}`)
      console.log(`  Disable: ${color.cyan("ring telemetry --disable")}`)
    }
  })
```

**Add to config schema (`plugin/config/schema.ts`):**

```typescript
// Add telemetry config
export const TelemetryConfigSchema = z.object({
  /** Enable anonymous telemetry */
  enabled: z.boolean().default(false),
})

// Add to RingConfigSchema
export const RingConfigSchema = z.object({
  // ... existing fields ...

  /** Telemetry configuration */
  telemetry: TelemetryConfigSchema.optional().default({
    enabled: false,
  }),
})
```

**Verification:**

```bash
# Build
bun run build

# Test telemetry commands
./dist/cli/index.js telemetry --status
./dist/cli/index.js telemetry --enable
./dist/cli/index.js telemetry --stats
./dist/cli/index.js telemetry --disable
```

---

## Batch 4 Review Checkpoint

After completing Tasks 4.1-4.2:

```bash
# Run all tests
bun test

# Type check
bun run typecheck

# Build
bun run build

# Verify new commands work
./dist/cli/index.js help
./dist/cli/index.js telemetry --status
```

**Review Criteria:**
- [ ] Config template is created on first run
- [ ] Telemetry is disabled by default
- [ ] Telemetry can be enabled/disabled via CLI
- [ ] No data is sent without explicit opt-in

---

## Final Verification

```bash
# Complete test suite
bun test

# Type check
bun run typecheck

# Lint
bun run lint

# Build
bun run build

# Manual verification
./dist/cli/index.js doctor
./dist/cli/index.js help
./dist/cli/index.js help --json | jq .

# Test in OpenCode
# 1. Start OpenCode
# 2. Type /ring:help
# 3. Verify output lists all components
```

---

## Summary

| Task | Status | Files Modified |
|------|--------|----------------|
| 1. `/ring:help` command | Pending | `assets/command/help.md`, `src/cli/help.ts`, `src/cli/index.ts` |
| 2. Hook extension docs | Pending | `docs/EXTENDING_HOOKS.md` |
| 3. Command validation | Pending | `plugin/loaders/command-loader.ts` |
| 4. Getting started tutorial | Pending | `docs/GETTING_STARTED.md` |
| 5. Integration tests | Pending | `__tests__/plugin/skills/workflow.integration.test.ts` |
| 6. Config auto-generation | Pending | `src/cli/config-manager.ts`, `src/cli/install.ts` |
| 7. Opt-in telemetry | Pending | `plugin/telemetry/index.ts`, `src/cli/index.ts`, `plugin/config/schema.ts` |

---

## Failure Recovery

| Issue | Recovery Steps |
|-------|---------------|
| Build fails | 1. Run `bun run typecheck` 2. Fix type errors 3. Rebuild |
| Tests fail | 1. Run failing test in isolation 2. Check test assumptions 3. Fix and re-run |
| Command not loading | 1. Check frontmatter syntax 2. Run `ring doctor` 3. Enable debug mode |
| Config validation errors | 1. Check JSON syntax 2. Compare against schema 3. Use template |
