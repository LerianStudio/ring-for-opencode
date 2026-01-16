# Ring for OpenCode - Architecture

This document provides a comprehensive technical overview of Ring's architecture, component interactions, and extension patterns.

## Table of Contents

- [Overview](#overview)
- [Architectural Patterns](#architectural-patterns)
- [Plugin Entry Point](#plugin-entry-point)
- [Hook System](#hook-system)
- [Component Loaders](#component-loaders)
- [Configuration System](#configuration-system)
- [Component Formats](#component-formats)
- [Data Flow](#data-flow)
- [Extension Guide](#extension-guide)

---

## Overview

Ring is a **unified plugin** for OpenCode that extends the AI assistant with specialized agents, reusable skills/workflows, and slash commands. The architecture follows three key principles:

1. **Declarative Components** - Agents, skills, and commands are defined in markdown files with YAML frontmatter
2. **Event-Driven Hooks** - Lifecycle events trigger middleware chains with priority-based execution

```
┌─────────────────────────────────────────────────────────────────┐
│                     OpenCode Framework                           │
└────────────────────────────┬────────────────────────────────────┘
                             │ Plugin Interface
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│              RingUnifiedPlugin (Single Entry Point)              │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────┐  ┌──────────────────────────┐    │
│  │         Loaders          │  │          Hooks           │    │
│  │  (agents, skills,        │  │  (lifecycle events)      │    │
│  │   commands)              │  │                          │    │
│  └──────────────────────────┘  └──────────────────────────┘    │
├─────────────────────────────────────────────────────────────────┤
│                    Component Layer (.opencode/)                  │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────┐  │
│  │  17 Agents │  │ 34 Skills  │  │17 Commands │  │ Patterns │  │
│  └────────────┘  └────────────┘  └────────────┘  └──────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Architectural Patterns

### Pattern 1: Unified Plugin Architecture

A single entry point (`RingUnifiedPlugin`) implements OpenCode's `Plugin` interface:

```typescript
type Plugin = (input: PluginInput) => Promise<Hooks>
```

This pattern provides:
- Single point of initialization
- Centralized configuration loading
- Unified component injection
- Consistent lifecycle management

### Pattern 2: Event-Driven Middleware (Hooks)

Hooks follow the **Chain of Responsibility** pattern:

```
Event → Hook Registry → Sort by Priority → Execute Sequentially
                                              │
                                    ┌─────────┴─────────┐
                                    │ Hook 1 (p:10)     │
                                    │ → returns data    │
                                    └─────────┬─────────┘
                                              │ chainData accumulated
                                    ┌─────────┴─────────┐
                                    │ Hook 2 (p:20)     │
                                    │ → receives data   │
                                    └─────────┬─────────┘
                                              │
                                    ┌─────────┴─────────┐
                                    │ Hook 3 (p:100)    │
                                    │ → can stop chain  │
                                    └───────────────────┘
```

### Pattern 3: Declarative Configuration

Components are defined in markdown files, not TypeScript code:

```
.opencode/agent/code-reviewer.md  →  AgentConfig
.opencode/skill/tdd/SKILL.md      →  SkillConfig
.opencode/command/commit.md       →  CommandConfig
```

This enables extension without code changes.

---

## Plugin Entry Point

**Location:** `plugin/ring-unified.ts`

### Initialization Sequence

```typescript
RingUnifiedPlugin(ctx: PluginInput) {
  // Phase 1: Load configuration (4-layer system)
  const config = loadConfig(ctx.directory)

  // Phase 2: Initialize hook registry
  initializeHooks(config)

  // Phase 3: Create config handler (loads components)
  const configHandler = createConfigHandler(projectRoot, config)

  // Phase 4: Create lifecycle router
  const lifecycleRouter = createLifecycleRouter(projectRoot, config)

  // Phase 5: Return hooks object
  return {
    tool: ringTools,
    config: configHandler,
    event: eventHandler,
    experimental: { ... }
  }
}
```

### Returned Hooks

| Hook | Type | Purpose |
|------|------|---------|
| `tool` | Object | Registers custom Ring tools |
| `config` | Function | Injects agents/skills/commands into OpenCode |
| `event` | Function | Routes lifecycle events to hook registry |
| `experimental.chat.system.transform` | Function | Injects context into system prompts |
| `experimental.session.compacting` | Function | Preserves context during compaction |

---

## Hook System

**Location:** `plugin/hooks/`

### Core Types

```typescript
interface Hook {
  name: HookName
  lifecycles: HookLifecycle[]
  priority?: number  // Lower = runs earlier (default: 100)
  enabled: boolean
  execute: (ctx: HookContext, output: HookOutput) => Promise<HookResult>
}

interface HookContext {
  sessionId: string
  directory: string
  lifecycle: HookLifecycle
  event?: { type: string; properties?: Record<string, unknown> }
  chainData?: Record<string, unknown>  // Data from previous hooks
}

interface HookResult {
  success: boolean
  error?: string
  data?: Record<string, unknown>  // Passed to next hook
  stopChain?: boolean             // Stop executing remaining hooks
}
```

### Lifecycle Events

| Event | Trigger | Typical Hooks |
|-------|---------|---------------|
| `session.created` | Session starts | session-start, context-injection |
| `session.idle` | Session becomes idle | notification |
| `session.compacting` | Context window recovery | context-injection |
| `chat.params` | Before model call | session-start |
| `todo.updated` | Todo list changes | task-completion |

### Registered Hooks

| Hook | Priority | Events | Purpose |
|------|----------|--------|---------|
| `session-start` | 10 | session.created, chat.params | Inject critical rules |
| `context-injection` | 20 | session.compacting | Inject compact context |
| `task-completion` | 100 | todo.updated | Monitor completion |
| `notification` | 200 | session.idle, event | Desktop notifications |

### Hook Registry

```typescript
// Registration
hookRegistry.register(hook)
hookRegistry.registerFactory(hookFactory)

// Execution
const results = await hookRegistry.executeLifecycle(
  "session.created",
  context,
  output
)
```

The registry:
1. Filters hooks by lifecycle event
2. Sorts by priority (ascending)
3. Executes sequentially
4. Accumulates chainData between hooks
5. Stops if `stopChain: true` returned

---

## Component Loaders

**Location:** `plugin/loaders/`

### Agent Loader

**File:** `agent-loader.ts`

```
.opencode/agent/*.md → Parse frontmatter → Validate → AgentConfig
```

**Frontmatter Schema:**
```yaml
---
description: "Agent description"
mode: subagent | primary
model: optional-model-override
tools: "bash,webfetch,!task"  # !prefix disables tool
color: optional-ui-color
---
```

**Output:** `Record<string, AgentConfig>`

### Skill Loader

**File:** `skill-loader.ts`

```
.opencode/skill/*/SKILL.md → Parse frontmatter → SkillConfig
```

**Frontmatter Schema:**
```yaml
---
name: skill-name
description: Multi-line description
license: MIT
compatibility: opencode
metadata:
  trigger: "When to use"
  skip_when: "When not to use"
---
```

**Output:** `Record<string, SkillConfig>` (prefixed with `ring-default:`)

### Command Loader

**File:** `command-loader.ts`

```
.opencode/command/*.md → Parse frontmatter → CommandConfig
```

**Frontmatter Schema:**
```yaml
---
description: "Command description"
subtask: false
---
```

**Output:** `Record<string, CommandConfig>`

### Config Injection

```typescript
// In config-handler.ts
config.agent = {
  ...ringAgents,           // Ring agents as base
  ...(config.agent ?? {}), // Project overrides take precedence
}

config.command = {
  ...ringSkills,           // Skills with "ring-default:" namespace
  ...ringCommands,         // Commands directly
  ...(config.command ?? {}),
}
```

---

## Configuration System

**Location:** `plugin/config/`

Ring reads config from `~/.config/opencode/ring/config.jsonc` (user scope) and project files at `.opencode/ring.jsonc` or `.ring/config.jsonc`.

### 4-Layer Priority System

```
Layer 1: Built-in defaults (DEFAULT_RING_CONFIG)
    ↓
Layer 2: User config (~/.config/opencode/ring/config.jsonc)
    ↓
Layer 3: Project config (.opencode/ring.jsonc or .ring/config.jsonc)
    ↓
Layer 4: Local overrides (.ring/local.jsonc)
```

Later layers override earlier ones.

### Configuration Schema

```typescript
interface RingConfig {
  disabled_hooks: string[]
  disabled_agents: string[]
  disabled_skills: string[]
  disabled_commands: string[]

  notifications: {
    enabled: boolean
    onIdle: boolean
    onError: boolean
  }

  hooks: {
    "session-start": SessionStartConfig
    "context-injection": ContextInjectionConfig
    // ...
  }
}
```

### Schema Validation

JSON Schema at `.opencode/ring-config.schema.json` enables IDE autocomplete.

---

## Component Formats

### Agent Format

```markdown
---
description: "What this agent does"
mode: subagent
---

# Agent Name

## Shared Patterns (MANDATORY)

| Pattern | Description |
|---------|-------------|
| [exit-criteria](../skill/shared-patterns/exit-criteria.md) | Completion gates |

## Your Role

[Agent prompt content...]

## Output Format

[Expected output structure...]
```

### Skill Format

```markdown
---
name: skill-name
description: |
  Multi-line description of what
  this skill teaches.
license: MIT
compatibility: opencode
metadata:
  trigger: "When to invoke"
  skip_when: "When to skip"
---

# Skill Name

## Overview

[Core principle]

## When to Use

- Always: [conditions]
- Never: [conditions]

## Workflow

1. Step one
2. Step two
3. Step three

## Verification Checklist

- [ ] Criterion 1
- [ ] Criterion 2
```

### Command Format

```markdown
---
description: What the command does
subtask: false
---

# Command Name

## Process

1. Gather context
2. Execute action
3. Verify result

## $ARGUMENTS

[Argument documentation]
```

### Shared Pattern Format

```markdown
# Pattern Name

## Purpose

[What this pattern ensures]

## Rules

| Situation | Required Action |
|-----------|-----------------|
| X happens | Do Y |

## Template

```markdown
[Copy-paste template]
```
```

---

## Data Flow

### Session Initialization

```
OpenCode starts session
         │
         ▼
RingUnifiedPlugin.config() called
         │
         ├─→ loadRingAgents() → 17 agents injected
         ├─→ loadRingSkills() → 34 skills injected
         └─→ loadRingCommands() → 17 commands injected
         │
         ▼
OpenCode fires session.created event
         │
         ▼
RingUnifiedPlugin.event() called
         │
         ├─→ lifecycleRouter() → cleanup old state
         └─→ hookRegistry.executeLifecycle("session.created")
                  │
                  ├─→ session-start hook (p:10)
                  │   └─→ Inject critical rules into system prompt
                  │
                  └─→ [other hooks by priority]
```

### Command Execution

```
User types /ring:commit
         │
         ▼
OpenCode loads .opencode/command/commit.md
         │
         ▼
Frontmatter specifies: agent: "build"
         │
         ▼
Agent "build" orchestrates workflow
         │
         ├─→ May invoke skills
         ├─→ May dispatch subagents
         └─→ Uses shared patterns for quality gates
```

### Parallel Code Review

```
User types /ring:codereview
         │
         ▼
Command dispatches 5 reviewers in parallel
         │
         ├─→ @ring:code-reviewer
         ├─→ @ring:business-logic-reviewer
         ├─→ @ring:security-reviewer
         ├─→ @ring:test-reviewer
         └─→ @ring:nil-safety-reviewer
         │
         ▼
Results aggregated into unified report
```

---

## Extension Guide

### Adding a New Agent

1. Create `.opencode/agent/my-agent.md`:

```markdown
---
description: "What this agent does"
mode: subagent
---

# My Agent

## Shared Patterns (MANDATORY)

| Pattern | Description |
|---------|-------------|
| [exit-criteria](../skill/shared-patterns/exit-criteria.md) | Completion gates |

## Your Role

You are a [role description]...

## Output Format

[Define expected output]
```

2. Agent auto-loads on next session
3. Invoke via `@my-agent` mention

### Adding a New Skill

1. Create `.opencode/skill/my-skill/SKILL.md`:

```markdown
---
name: my-skill
description: What this skill teaches
license: MIT
compatibility: opencode
metadata:
  trigger: "When to use"
  skip_when: "When not to use"
---

# My Skill

## Workflow

1. Step one
2. Step two

## Verification

- [ ] Done criterion
```

2. Skill auto-loads as `/my-skill` command

### Adding a New Command

1. Create `.opencode/command/my-command.md`:

```markdown
---
description: What the command does
subtask: false
---

# My Command

## Process

1. First step
2. Second step

## $ARGUMENTS

[Argument docs]
```

2. Command auto-loads as `/my-command`

### Adding a New Hook

1. Create `plugin/hooks/factories/my-hook.ts`:

```typescript
import type { Hook, HookFactory } from "../types"

export const createMyHook: HookFactory = (config?) => ({
  name: "my-hook",
  lifecycles: ["session.created"],
  priority: 50,
  enabled: true,
  execute: async (ctx, output) => {
    // Hook logic
    output.system?.push("Injected content")
    return { success: true }
  }
})
```

2. Register in `plugin/ring-unified.ts`:

```typescript
if (!isDisabled("my-hook")) {
  hookRegistry.register(createMyHook(config.hooks?.["my-hook"]))
}
```

3. Add to schema for enable/disable control

### Adding a New Shared Pattern

1. Create `.opencode/skill/shared-patterns/my-pattern.md`:

```markdown
# My Pattern

## Purpose

[What this ensures]

## Rules

| Situation | Action |
|-----------|--------|
| X | Do Y |
```

2. Reference from agents/skills:

```markdown
| [my-pattern](../skill/shared-patterns/my-pattern.md) | Description |
```

---

## Key Files Reference

| Purpose | Location |
|---------|----------|
| Plugin entry | `plugin/ring-unified.ts` |
| Hook registry | `plugin/hooks/registry.ts` |
| Hook types | `plugin/hooks/types.ts` |
| Agent loader | `plugin/loaders/agent-loader.ts` |
| Skill loader | `plugin/loaders/skill-loader.ts` |
| Command loader | `plugin/loaders/command-loader.ts` |
| Config loader | `plugin/config/loader.ts` |
| Config handler | `plugin/config/config-handler.ts` |
| Config schema | `assets/ring-config.schema.json` |
| Runtime config | `.ring/config.jsonc` |

---

## Debugging

### Enable Debug Mode

```bash
RING_DEBUG=true opencode
```

### Health Check

```bash
ring doctor
```

### Key Debug Points

| Issue | Check |
|-------|-------|
| Components not loading | `plugin/loaders/*.ts` - frontmatter parsing |
| Hooks not firing | `plugin/hooks/registry.ts:174` - executeLifecycle |
| Config not applied | `plugin/config/loader.ts` - 4-layer merge |

---

## Design Decisions

### Why Markdown Components?

1. **Self-documenting** - The file IS both config and documentation
2. **No compilation** - Changes take effect immediately
3. **Version control friendly** - Easy to diff and review
4. **Low barrier** - Add components without TypeScript knowledge

### Why Priority-Based Hooks?

1. **Predictable ordering** - Clear execution sequence
2. **Extensible** - Insert new hooks at any priority
3. **Data passing** - Hooks can communicate via chainData
4. **Error resilience** - Failed hooks don't break the chain

### Why Shared Patterns?

1. **Consistency** - All agents follow same rules
2. **Maintainability** - Update once, applies everywhere
3. **Quality gates** - Enforce standards systematically
4. **Documentation** - Patterns explain the "why"
