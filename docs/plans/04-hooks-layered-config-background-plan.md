# Hook Architecture, Layered Config, and Background Tasks Implementation Plan

> **For Agents:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Transform Ring's plugin system into a hook-based middleware architecture with layered configuration and background task management.

**Architecture:** Create a hook abstraction layer with lifecycle types (chat.params, chat.message, event, config, compaction), implement 4-layer configuration loading with deep merge, and add a BackgroundManager for parallel agent execution. Existing plugins will be converted to hook factories for consistent registration and conditional loading.

**Tech Stack:** TypeScript, Bun runtime, @opencode-ai/plugin, Zod for schema validation

**Global Prerequisites:**
- Environment: macOS/Linux, Bun 1.0+, Node.js 18+
- Tools: Verify with commands below
- Access: No external API keys required for this implementation
- State: Clean working tree on main branch

**Verification before starting:**
```bash
# Run ALL these commands and verify output:
bun --version        # Expected: 1.0.0+
node --version       # Expected: v18.0.0+
git status           # Expected: clean working tree (or only docs/plans changes)
ls plugin/           # Expected: index.ts, session-start.ts, context-injection.ts, etc.
```

## Historical Precedent

**Query:** "hook architecture config loading background tasks"
**Index Status:** Empty (new project)

No historical data available. This is normal for new projects.
Proceeding with standard planning approach based on oh-my-opencode reference patterns.

---

## Part 1: Hook Architecture Foundation

### Task 1: Create Hook Type Definitions

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/types.ts`

**Prerequisites:**
- Tools: None
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/` directory

**Step 1: Create the hooks directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks
```

**Expected output:**
```
(no output - silent success)
```

**Step 2: Write the type definitions file**

```typescript
/**
 * Ring Hook System Type Definitions
 *
 * Defines hook lifecycle types and interfaces for middleware pattern.
 * Based on oh-my-opencode patterns adapted for Ring.
 */

/**
 * Hook lifecycle events supported by Ring.
 * Each event corresponds to a specific point in the OpenCode pipeline.
 */
export type HookLifecycle =
  | "session.created"        // Session initialized
  | "session.idle"           // Session became idle
  | "session.compacting"     // Context being compacted
  | "chat.message"           // User message received
  | "chat.params"            // Chat parameters being set
  | "tool.before"            // Before tool execution
  | "tool.after"             // After tool execution
  | "todo.updated"           // Todo list changed
  | "event"                  // Generic event handler

/**
 * Hook execution result indicating success/failure.
 */
export interface HookResult {
  success: boolean
  error?: string
  /** Optional data to pass to next hook in chain */
  data?: Record<string, unknown>
  /** If true, stop hook chain execution */
  stopChain?: boolean
}

/**
 * Context provided to all hooks.
 */
export interface HookContext {
  /** Session identifier */
  sessionId: string
  /** Project root directory */
  directory: string
  /** Hook lifecycle event that triggered this hook */
  lifecycle: HookLifecycle
  /** Original event data from OpenCode */
  event?: {
    type: string
    properties?: Record<string, unknown>
  }
  /** Data passed from previous hook in chain */
  chainData?: Record<string, unknown>
}

/**
 * Base hook interface that all hooks must implement.
 */
export interface Hook {
  /** Unique identifier for this hook */
  name: HookName
  /** Hook lifecycle events this hook responds to */
  lifecycles: HookLifecycle[]
  /** Priority for execution order (lower = earlier, default 100) */
  priority?: number
  /** Whether this hook is enabled */
  enabled: boolean
  /** Execute the hook */
  execute: (ctx: HookContext, output: HookOutput) => Promise<HookResult>
}

/**
 * Hook output object for modifying OpenCode behavior.
 */
export interface HookOutput {
  /** System prompt context to inject */
  system?: string[]
  /** Compaction context to inject */
  context?: string[]
  /** Message parts to modify */
  parts?: Array<{ type: string; text?: string; [key: string]: unknown }>
  /** Block the operation */
  block?: boolean
  /** Reason for blocking */
  blockReason?: string
}

/**
 * Hook factory function signature.
 * All hooks are created via factory functions for consistent initialization.
 */
export type HookFactory<TConfig = Record<string, unknown>> = (
  config?: TConfig
) => Hook

/**
 * Built-in hook names supported by Ring.
 */
export type HookName =
  | "session-start"
  | "context-injection"
  | "notification"
  | "task-completion"
  | "session-outcome"
  | "outcome-inference"
  | "doubt-resolver"
  | "background-notification"
  | "compaction-context"
  | "rules-injector"
  | "agent-reminder"

/**
 * Hook registry entry with metadata.
 */
export interface HookRegistryEntry {
  name: HookName
  factory: HookFactory
  defaultEnabled: boolean
  description: string
}

/**
 * Event handler signature for hooks.
 */
export type HookEventHandler = (input: { event: { type: string; properties?: unknown } }) => Promise<void>

/**
 * Chat message handler signature.
 */
export type HookChatHandler = (
  input: { sessionID: string; agent?: string },
  output: { parts: Array<{ type: string; text?: string }> }
) => Promise<void>

/**
 * Compaction handler signature.
 */
export type HookCompactionHandler = (
  input: { sessionID: string },
  output: { context: string[] }
) => Promise<void>

/**
 * System transform handler signature.
 */
export type HookSystemHandler = (
  input: Record<string, unknown>,
  output: { system: string[] }
) => Promise<void>
```

**Step 3: Verify the file was created**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/types.ts | head -20`

**Expected output:**
```
/**
 * Ring Hook System Type Definitions
 *
 * Defines hook lifecycle types and interfaces for middleware pattern.
 * Based on oh-my-opencode patterns adapted for Ring.
 */

/**
 * Hook lifecycle events supported by Ring.
 * Each event corresponds to a specific point in the OpenCode pipeline.
 */
export type HookLifecycle =
  | "session.created"        // Session initialized
  | "session.idle"           // Session became idle
  | "session.compacting"     // Context being compacted
  | "chat.message"           // User message received
  | "chat.params"            // Chat parameters being set
  | "tool.before"            // Before tool execution
  | "tool.after"             // After tool execution
```

**If Task Fails:**
1. **Directory doesn't exist:** Run `mkdir -p plugin/hooks` first
2. **Permission denied:** Check file permissions with `ls -la plugin/`
3. **Can't recover:** Document error and return to human partner

---

### Task 2: Create Hook Registry

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/registry.ts`

**Prerequisites:**
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/types.ts`

**Step 1: Write the hook registry implementation**

```typescript
/**
 * Ring Hook Registry
 *
 * Central registry for all hooks with registration, lookup, and execution.
 * Supports conditional loading based on configuration.
 */

import type {
  Hook,
  HookName,
  HookLifecycle,
  HookContext,
  HookOutput,
  HookResult,
  HookRegistryEntry,
  HookFactory,
} from "./types"

/**
 * Hook execution order comparator.
 * Lower priority executes first.
 */
function compareHookPriority(a: Hook, b: Hook): number {
  return (a.priority ?? 100) - (b.priority ?? 100)
}

/**
 * Central hook registry singleton.
 */
class HookRegistry {
  private hooks: Map<HookName, Hook> = new Map()
  private factories: Map<HookName, HookRegistryEntry> = new Map()
  private disabledHooks: Set<HookName> = new Set()

  /**
   * Register a hook factory for later instantiation.
   */
  registerFactory(entry: HookRegistryEntry): void {
    this.factories.set(entry.name, entry)
  }

  /**
   * Get all registered factory entries.
   */
  getFactories(): HookRegistryEntry[] {
    return Array.from(this.factories.values())
  }

  /**
   * Instantiate and register a hook from its factory.
   */
  instantiate<TConfig>(
    name: HookName,
    config?: TConfig
  ): Hook | null {
    const entry = this.factories.get(name)
    if (!entry) {
      console.warn(`[ring-hooks] Unknown hook: ${name}`)
      return null
    }

    if (this.disabledHooks.has(name)) {
      return null
    }

    const hook = entry.factory(config)
    this.hooks.set(name, hook)
    return hook
  }

  /**
   * Register an already-instantiated hook.
   */
  register(hook: Hook): void {
    if (this.disabledHooks.has(hook.name)) {
      return
    }
    this.hooks.set(hook.name, hook)
  }

  /**
   * Unregister a hook by name.
   */
  unregister(name: HookName): void {
    this.hooks.delete(name)
  }

  /**
   * Get a hook by name.
   */
  get(name: HookName): Hook | undefined {
    return this.hooks.get(name)
  }

  /**
   * Check if a hook is registered and enabled.
   */
  has(name: HookName): boolean {
    return this.hooks.has(name) && !this.disabledHooks.has(name)
  }

  /**
   * Disable a hook by name.
   */
  disable(name: HookName): void {
    this.disabledHooks.add(name)
    this.hooks.delete(name)
  }

  /**
   * Enable a hook by name.
   */
  enable(name: HookName): void {
    this.disabledHooks.delete(name)
  }

  /**
   * Check if a hook is disabled.
   */
  isDisabled(name: HookName): boolean {
    return this.disabledHooks.has(name)
  }

  /**
   * Set multiple hooks as disabled.
   */
  setDisabledHooks(names: HookName[]): void {
    this.disabledHooks = new Set(names)
    // Remove any already-registered hooks that are now disabled
    for (const name of names) {
      this.hooks.delete(name)
    }
  }

  /**
   * Get all hooks that respond to a specific lifecycle event.
   */
  getHooksForLifecycle(lifecycle: HookLifecycle): Hook[] {
    const matching: Hook[] = []
    for (const hook of this.hooks.values()) {
      if (hook.enabled && hook.lifecycles.includes(lifecycle)) {
        matching.push(hook)
      }
    }
    return matching.sort(compareHookPriority)
  }

  /**
   * Execute all hooks for a lifecycle event in order.
   */
  async executeLifecycle(
    lifecycle: HookLifecycle,
    ctx: Omit<HookContext, "lifecycle">,
    output: HookOutput
  ): Promise<HookResult[]> {
    const hooks = this.getHooksForLifecycle(lifecycle)
    const results: HookResult[] = []
    let chainData: Record<string, unknown> = {}

    for (const hook of hooks) {
      const fullCtx: HookContext = {
        ...ctx,
        lifecycle,
        chainData,
      }

      try {
        const result = await hook.execute(fullCtx, output)
        results.push(result)

        if (result.data) {
          chainData = { ...chainData, ...result.data }
        }

        if (result.stopChain) {
          break
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error)
        results.push({
          success: false,
          error: `Hook ${hook.name} failed: ${errorMessage}`,
        })
        // Continue with other hooks unless explicitly stopped
      }
    }

    return results
  }

  /**
   * Get all registered hook names.
   */
  getRegisteredNames(): HookName[] {
    return Array.from(this.hooks.keys())
  }

  /**
   * Get count of registered hooks.
   */
  count(): number {
    return this.hooks.size
  }

  /**
   * Clear all registered hooks (useful for testing).
   */
  clear(): void {
    this.hooks.clear()
    this.disabledHooks.clear()
  }
}

/**
 * Global hook registry instance.
 */
export const hookRegistry = new HookRegistry()

/**
 * Helper to check if a hook is disabled in config.
 */
export function isHookDisabled(
  config: { disabledHooks?: HookName[] } | undefined,
  hookName: HookName
): boolean {
  if (!config?.disabledHooks) return false
  return config.disabledHooks.includes(hookName)
}
```

**Step 2: Verify the file was created**

Run: `wc -l /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/registry.ts`

**Expected output:**
```
     180 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/registry.ts
```

**If Task Fails:**
1. **Import error:** Verify types.ts exists first
2. **Syntax error:** Check TypeScript syntax with `bun build plugin/hooks/registry.ts --no-bundle`

---

### Task 3: Create Hook Index Export

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/index.ts`

**Prerequisites:**
- Files must exist: `types.ts`, `registry.ts`

**Step 1: Write the index file**

```typescript
/**
 * Ring Hook System
 *
 * Central export for all hook-related types, registry, and utilities.
 */

// Type exports
export type {
  HookLifecycle,
  HookResult,
  HookContext,
  Hook,
  HookOutput,
  HookFactory,
  HookName,
  HookRegistryEntry,
  HookEventHandler,
  HookChatHandler,
  HookCompactionHandler,
  HookSystemHandler,
} from "./types"

// Registry exports
export { hookRegistry, isHookDisabled } from "./registry"
```

**Step 2: Verify TypeScript compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build plugin/hooks/index.ts --no-bundle --outdir=/tmp/ring-test 2>&1 | head -5`

**Expected output:**
```
(empty or build success message)
```

**If you see errors:** Check import paths match actual file names

---

### Task 4: Convert Session Start Plugin to Hook Factory

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/session-start.ts`

**Prerequisites:**
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/types.ts`
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/session-start.ts` (reference)

**Step 1: Create factories directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories
```

**Step 2: Write the session-start hook factory**

```typescript
/**
 * Session Start Hook Factory
 *
 * Injects Ring context at session creation and into system prompts.
 * Converted from RingSessionStart plugin to hook pattern.
 */

import type { Hook, HookFactory, HookContext, HookOutput, HookResult } from "../types"
import { existsSync, readFileSync, readdirSync, statSync, lstatSync } from "fs"
import { join } from "path"
import { cleanupOldState, deleteState, getSessionId, escapeAngleBrackets, sanitizeForPrompt, isPathWithinRoot } from "../../utils/state"

/**
 * Configuration for session start hook.
 */
export interface SessionStartConfig {
  /** Include critical rules in injection */
  includeCriticalRules?: boolean
  /** Include agent reminder */
  includeAgentReminder?: boolean
  /** Include doubt questions pattern */
  includeDoubtQuestions?: boolean
  /** Include duplication guard */
  includeDuplicationGuard?: boolean
  /** Maximum age of state files to keep (days) */
  stateMaxAgeDays?: number
}

const DEFAULT_CONFIG: Required<SessionStartConfig> = {
  includeCriticalRules: true,
  includeAgentReminder: true,
  includeDoubtQuestions: true,
  includeDuplicationGuard: true,
  stateMaxAgeDays: 7,
}

// Critical rules content
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
WRONG. Count > 3 = agent. No exceptions. Task size is irrelevant.`

const AGENT_REMINDER = `## AGENT USAGE REMINDER

CONTEXT CHECK: Before using Glob/Grep/Read chains, consider agents:

| Task | Agent |
|------|-------|
| Explore codebase | Explore |
| Multi-file search | Explore |
| Code review | code-reviewer + business-logic-reviewer + security-reviewer (PARALLEL) |
| Implementation plan | write-plan |

**3-File Rule:** If reading >3 files, use an agent instead.`

const DOUBT_QUESTIONS = `## DOUBT-TRIGGERED QUESTIONS

**Resolution hierarchy (check BEFORE asking):**
1. User dispatch context -> Did they already specify?
2. CLAUDE.md / repo conventions -> Is there a standard?
3. Codebase patterns -> What does existing code do?
4. Best practice -> Is one approach clearly superior?
5. **-> ASK** -> Only if ALL above fail AND affects correctness`

const DUPLICATION_GUARD = `## DUPLICATION PREVENTION

**BEFORE ADDING CONTENT** to any file:
1. SEARCH FIRST
2. If exists -> REFERENCE it, don't copy
3. NEVER duplicate - always link to single source of truth`

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
    const currentPhase = phaseMatch ? content.split("\n").find((line) => line.includes("[->"))?.trim() || "" : ""

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
 * Create a session start hook instance.
 */
export const createSessionStartHook: HookFactory<SessionStartConfig> = (
  userConfig = {}
) => {
  const config = { ...DEFAULT_CONFIG, ...userConfig }

  const hook: Hook = {
    name: "session-start",
    lifecycles: ["session.created", "chat.params"],
    priority: 10, // Run early
    enabled: true,

    async execute(ctx: HookContext, output: HookOutput): Promise<HookResult> {
      const { directory, lifecycle, sessionId } = ctx

      // Handle session.created - cleanup and initialization
      if (lifecycle === "session.created") {
        deleteState(directory, "context-usage", sessionId)
        cleanupOldState(directory, config.stateMaxAgeDays)
        return { success: true, data: { initialized: true } }
      }

      // Handle chat.params - inject system context
      if (lifecycle === "chat.params") {
        if (!output.system) {
          output.system = []
        }

        const sections: string[] = []

        if (config.includeCriticalRules) {
          sections.push(`<ring-critical-rules>\n${CRITICAL_RULES}\n</ring-critical-rules>`)
        }

        if (config.includeAgentReminder) {
          sections.push(`<ring-agent-reminder>\n${AGENT_REMINDER}\n</ring-agent-reminder>`)
        }

        if (config.includeDuplicationGuard) {
          sections.push(`<ring-duplication-guard>\n${DUPLICATION_GUARD}\n</ring-duplication-guard>`)
        }

        if (config.includeDoubtQuestions) {
          sections.push(`<ring-doubt-questions>\n${DOUBT_QUESTIONS}\n</ring-doubt-questions>`)
        }

        // Add ledger context if present
        const ledger = findActiveLedger(directory)
        if (ledger) {
          const safeName = sanitizeForPrompt(ledger.name, 120)
          const safePhase = sanitizeForPrompt(ledger.currentPhase || "No active phase", 200)
          const safeContent = escapeAngleBrackets(ledger.content)

          sections.push(`<ring-continuity-ledger>
## Active Continuity Ledger: ${safeName}
**Current Phase:** ${safePhase}

<continuity-ledger-content>
${safeContent}
</continuity-ledger-content>
</ring-continuity-ledger>`)
        }

        output.system.push(sections.join("\n\n"))
        return { success: true }
      }

      return { success: true }
    },
  }

  return hook
}

export default createSessionStartHook
```

**Step 3: Verify file created and compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/session-start.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

**Expected output:**
```
(empty or success message, no errors)
```

---

### Task 5: Convert Context Injection to Hook Factory

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/context-injection.ts`

**Prerequisites:**
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/types.ts`

**Step 1: Write the context injection hook factory**

```typescript
/**
 * Context Injection Hook Factory
 *
 * Injects Ring context during session compaction.
 * Converted from RingContextInjection plugin to hook pattern.
 */

import type { Hook, HookFactory, HookContext, HookOutput, HookResult } from "../types"
import { existsSync, readFileSync, readdirSync, statSync } from "fs"
import { join } from "path"

/**
 * Configuration for context injection hook.
 */
export interface ContextInjectionConfig {
  /** Include skills reference in compaction */
  includeSkillsReference?: boolean
  /** Include commands reference */
  includeCommandsReference?: boolean
  /** Include agents reference */
  includeAgentsReference?: boolean
}

const DEFAULT_CONFIG: Required<ContextInjectionConfig> = {
  includeSkillsReference: true,
  includeCommandsReference: true,
  includeAgentsReference: true,
}

// Compact critical rules for compaction context
const CRITICAL_RULES_COMPACT = `**3-FILE RULE: HARD GATE** - >3 files = dispatch agent, no exceptions.
**AUTO-TRIGGER:** "fix issues" -> agent, "find where" -> Explore agent.
**TDD:** Test must fail (RED) before implementation.
**REVIEW:** All 5 reviewers must pass before merge.
**COMMIT:** Use /commit command for intelligent grouping.`

const SKILLS_REFERENCE = `## Ring Skills System

Key skills for common tasks:
- **test-driven-development**: TDD methodology (RED -> GREEN -> REFACTOR)
- **requesting-code-review**: Parallel review with 5 specialized reviewers
- **systematic-debugging**: 4-phase debugging methodology
- **writing-plans**: Create detailed implementation plans
- **executing-plans**: Execute plans in batches with review checkpoints
- **dispatching-parallel-agents**: Run multiple agents concurrently
- **verification-before-completion**: Ensure work complete before done`

const COMMANDS_REFERENCE = `## Ring Commands

- /commit - Intelligent commit grouping
- /codereview - Dispatch 5 parallel reviewers
- /brainstorm - Socratic design refinement
- /execute-plan - Batch execution with checkpoints
- /worktree - Create isolated development branch
- /lint - Run and fix lint issues
- /create-handoff - Create session handoff document`

const AGENTS_REFERENCE = `## Ring Agents

- code-reviewer - Technical code quality review
- security-reviewer - Security vulnerability analysis
- business-logic-reviewer - Business requirements alignment
- codebase-explorer - Autonomous codebase exploration
- write-plan - Implementation planning agent`

/**
 * Get ledger summary for compaction context.
 */
function getLedgerSummary(projectRoot: string): string | null {
  const ledgerDir = join(projectRoot, ".ring/ledgers")
  if (!existsSync(ledgerDir)) return null

  try {
    const files = readdirSync(ledgerDir)
      .filter((f) => f.startsWith("CONTINUITY-") && f.endsWith(".md"))
      .map((f) => ({
        name: f.replace(".md", ""),
        path: join(ledgerDir, f),
        mtime: statSync(join(ledgerDir, f)).mtimeMs,
      }))
      .sort((a, b) => b.mtime - a.mtime)

    if (files.length === 0) return null

    const ledger = files[0]
    const content = readFileSync(ledger.path, "utf-8")
    const phaseLine = content.split("\n").find((line) => line.includes("[->"))
    const currentPhase = phaseLine?.trim() || "No active phase"

    const sections: string[] = []
    const briefMatch = content.match(/## BRIEF_SUMMARY\n([^#]+)/s)
    if (briefMatch) {
      sections.push(briefMatch[1].trim().slice(0, 200))
    }

    return `## Active Ledger: ${ledger.name}
**Current Phase:** ${currentPhase}
${sections.join("\n\n")}

**Full ledger at:** .ring/ledgers/${ledger.name}.md`
  } catch {
    return null
  }
}

/**
 * Create a context injection hook instance.
 */
export const createContextInjectionHook: HookFactory<ContextInjectionConfig> = (
  userConfig = {}
) => {
  const config = { ...DEFAULT_CONFIG, ...userConfig }

  const hook: Hook = {
    name: "context-injection",
    lifecycles: ["session.compacting"],
    priority: 20,
    enabled: true,

    async execute(ctx: HookContext, output: HookOutput): Promise<HookResult> {
      if (ctx.lifecycle !== "session.compacting") {
        return { success: true }
      }

      if (!output.context) {
        output.context = []
      }

      // Inject critical rules (compact version)
      output.context.push(`## Ring Critical Rules (Compact)\n${CRITICAL_RULES_COMPACT}`)

      // Inject skills reference
      if (config.includeSkillsReference) {
        output.context.push(SKILLS_REFERENCE)
      }

      // Inject commands reference
      if (config.includeCommandsReference) {
        output.context.push(COMMANDS_REFERENCE)
      }

      // Inject agents reference
      if (config.includeAgentsReference) {
        output.context.push(AGENTS_REFERENCE)
      }

      // Inject ledger summary if present
      const ledgerSummary = getLedgerSummary(ctx.directory)
      if (ledgerSummary) {
        output.context.push(ledgerSummary)
      }

      return { success: true }
    },
  }

  return hook
}

export default createContextInjectionHook
```

**Step 2: Verify file compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/context-injection.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

**Expected output:** (no errors)

---

### Task 6: Convert Notification to Hook Factory

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/notification.ts`

**Step 1: Write the notification hook factory**

```typescript
/**
 * Notification Hook Factory
 *
 * Sends desktop notifications on session events.
 * Converted from RingNotification plugin to hook pattern.
 */

import type { Hook, HookFactory, HookContext, HookOutput, HookResult } from "../types"

/**
 * Configuration for notification hook.
 */
export interface NotificationConfig {
  /** Enable notifications on session idle */
  notifyOnIdle?: boolean
  /** Enable notifications on session error */
  notifyOnError?: boolean
  /** Custom notification title */
  title?: string
}

const DEFAULT_CONFIG: Required<NotificationConfig> = {
  notifyOnIdle: true,
  notifyOnError: true,
  title: "Ring",
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
  $: (strings: TemplateStringsArray, ...values: unknown[]) => Promise<unknown>
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
    } else if (platform === "win32") {
      const script = `
        [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
        $template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
        $textNodes = $template.GetElementsByTagName("text")
        $textNodes.Item(0).AppendChild($template.CreateTextNode("${safeTitle}")) | Out-Null
        $textNodes.Item(1).AppendChild($template.CreateTextNode("${safeMessage}")) | Out-Null
        $toast = [Windows.UI.Notifications.ToastNotification]::new($template)
        [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Ring").Show($toast)
      `
      await $`powershell -Command ${script}`
    }
  } catch {
    // Notification not critical - silent failure
  }
}

// Store shell reference for notification sending
let shellRef: ((strings: TemplateStringsArray, ...values: unknown[]) => Promise<unknown>) | null = null

/**
 * Set the shell reference for notifications.
 * Must be called during plugin initialization.
 */
export function setNotificationShell(
  $: (strings: TemplateStringsArray, ...values: unknown[]) => Promise<unknown>
): void {
  shellRef = $
}

/**
 * Create a notification hook instance.
 */
export const createNotificationHook: HookFactory<NotificationConfig> = (
  userConfig = {}
) => {
  const config = { ...DEFAULT_CONFIG, ...userConfig }

  const hook: Hook = {
    name: "notification",
    lifecycles: ["session.idle", "event"],
    priority: 90, // Run late
    enabled: true,

    async execute(ctx: HookContext, _output: HookOutput): Promise<HookResult> {
      if (!shellRef) {
        return { success: false, error: "Shell not initialized" }
      }

      const eventType = ctx.event?.type

      if (eventType === "session.idle" && config.notifyOnIdle) {
        await sendNotification(config.title, "Session completed!", shellRef)
        return { success: true }
      }

      if (eventType === "session.error" && config.notifyOnError) {
        await sendNotification(config.title, "Session encountered an error", shellRef)
        return { success: true }
      }

      return { success: true }
    },
  }

  return hook
}

export default createNotificationHook
```

**Step 2: Verify file compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/notification.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

**Expected output:** (no errors)

---

### Task 7: Convert Task Completion to Hook Factory

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/task-completion.ts`

**Step 1: Write the task completion hook factory**

```typescript
/**
 * Task Completion Hook Factory
 *
 * Monitors todo updates and suggests handoff on completion.
 * Converted from RingTaskCompletionCheck plugin to hook pattern.
 */

import type { Hook, HookFactory, HookContext, HookOutput, HookResult } from "../types"
import { writeState, getSessionId } from "../../utils/state"

/**
 * Configuration for task completion hook.
 */
export interface TaskCompletionConfig {
  /** Show toast on all tasks complete */
  showCompletionToast?: boolean
  /** Toast duration in milliseconds */
  toastDuration?: number
}

const DEFAULT_CONFIG: Required<TaskCompletionConfig> = {
  showCompletionToast: true,
  toastDuration: 8000,
}

interface Todo {
  content: string
  status: "pending" | "in_progress" | "completed"
  activeForm?: string
}

// Store client reference for toast notifications
let clientRef: {
  tui: {
    showToast: (opts: { body: { title: string; message: string; variant: string; duration: number } }) => Promise<void>
  }
} | null = null

/**
 * Set the client reference for toast notifications.
 */
export function setTaskCompletionClient(client: typeof clientRef): void {
  clientRef = client
}

/**
 * Create a task completion hook instance.
 */
export const createTaskCompletionHook: HookFactory<TaskCompletionConfig> = (
  userConfig = {}
) => {
  const config = { ...DEFAULT_CONFIG, ...userConfig }

  const hook: Hook = {
    name: "task-completion",
    lifecycles: ["todo.updated"],
    priority: 50,
    enabled: true,

    async execute(ctx: HookContext, _output: HookOutput): Promise<HookResult> {
      if (ctx.lifecycle !== "todo.updated") {
        return { success: true }
      }

      // Extract todos from event
      const eventProps = ctx.event?.properties as { todos?: Todo[] } | undefined
      const todos = eventProps?.todos || []

      if (!Array.isArray(todos) || todos.length === 0) {
        return { success: true }
      }

      // Persist todos state
      const sessionId = getSessionId()
      writeState(ctx.directory, "todos-state", todos, sessionId)

      // Count by status
      const total = todos.length
      const completed = todos.filter((t) => t.status === "completed").length

      // Check if all complete
      if (total > 0 && completed === total && config.showCompletionToast && clientRef) {
        await clientRef.tui.showToast({
          body: {
            title: "All Tasks Complete!",
            message: `${total} tasks done. Consider /create-handoff if context is high.`,
            variant: "success",
            duration: config.toastDuration,
          },
        })
      }

      return {
        success: true,
        data: {
          total,
          completed,
          allComplete: total > 0 && completed === total,
        },
      }
    },
  }

  return hook
}

export default createTaskCompletionHook
```

**Step 2: Verify file compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/task-completion.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 8: Create Factories Index

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/index.ts`

**Step 1: Write the factories index**

```typescript
/**
 * Ring Hook Factories Index
 *
 * Central export for all hook factories.
 */

export { createSessionStartHook, type SessionStartConfig } from "./session-start"
export { createContextInjectionHook, type ContextInjectionConfig } from "./context-injection"
export { createNotificationHook, setNotificationShell, type NotificationConfig } from "./notification"
export { createTaskCompletionHook, setTaskCompletionClient, type TaskCompletionConfig } from "./task-completion"

// Re-export for convenience
export type { Hook, HookFactory, HookName } from "../types"
```

**Step 2: Verify exports compile**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/hooks/factories/index.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 9: Run Code Review Checkpoint (Part 1)

**Step 1: Dispatch code reviewers in parallel**

REQUIRED SUB-SKILL: Use requesting-code-review

Review the hook architecture files created in Tasks 1-8:
- `plugin/hooks/types.ts`
- `plugin/hooks/registry.ts`
- `plugin/hooks/index.ts`
- `plugin/hooks/factories/*.ts`

**Step 2: Handle findings by severity**

**Critical/High/Medium Issues:**
- Fix immediately
- Re-run reviewers after fixes
- Repeat until zero Critical/High/Medium issues

**Low Issues:**
- Add `TODO(review):` comments

**Step 3: Proceed only when zero Critical/High/Medium issues remain**

---

## Part 2: Layered Configuration System

### Task 10: Create Configuration Schema with Zod

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/config/schema.ts`

**Prerequisites:**
- Add Zod dependency if not present

**Step 1: Check if Zod is installed**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && grep -q "zod" package.json && echo "zod installed" || echo "zod not installed"`

**Step 2: Install Zod if needed**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun add zod`

**Step 3: Create config directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/config
```

**Step 4: Write the schema file**

```typescript
/**
 * Ring Configuration Schema
 *
 * Defines the structure for Ring's layered configuration system.
 * Uses Zod for runtime validation.
 */

import { z } from "zod"

/**
 * Hook names that can be disabled.
 */
export const HookNameSchema = z.enum([
  "session-start",
  "context-injection",
  "notification",
  "task-completion",
  "session-outcome",
  "outcome-inference",
  "doubt-resolver",
  "background-notification",
  "compaction-context",
  "rules-injector",
  "agent-reminder",
])

/**
 * Agent names that can be disabled.
 */
export const AgentNameSchema = z.enum([
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
  "devops-engineer",
  "sre",
  "qa-analyst",
])

/**
 * Skill names that can be disabled.
 */
export const SkillNameSchema = z.enum([
  "using-ring",
  "test-driven-development",
  "systematic-debugging",
  "requesting-code-review",
  "writing-plans",
  "executing-plans",
  "dispatching-parallel-agents",
  "verification-before-completion",
  "commit",
  "brainstorm",
  "codereview",
  "lint",
  "worktree",
])

/**
 * Command names that can be disabled.
 */
export const CommandNameSchema = z.enum([
  "commit",
  "codereview",
  "brainstorm",
  "execute-plan",
  "worktree",
  "lint",
  "create-handoff",
  "resume-handoff",
  "explore-codebase",
  "interview-me",
  "write-plan",
])

/**
 * Background task configuration.
 */
export const BackgroundTaskConfigSchema = z.object({
  /** Default concurrency for background tasks */
  defaultConcurrency: z.number().min(1).max(10).default(3),
  /** Per-agent concurrency limits */
  agentConcurrency: z.record(z.string(), z.number().min(1).max(10)).optional(),
  /** Task timeout in milliseconds */
  taskTimeoutMs: z.number().min(60000).max(3600000).default(1800000), // 30 min default
})

/**
 * Notification configuration.
 */
export const NotificationConfigSchema = z.object({
  /** Enable desktop notifications */
  enabled: z.boolean().default(true),
  /** Notify on session idle */
  onIdle: z.boolean().default(true),
  /** Notify on session error */
  onError: z.boolean().default(true),
  /** Notify on background task completion */
  onBackgroundComplete: z.boolean().default(true),
})

/**
 * Experimental features configuration.
 */
export const ExperimentalConfigSchema = z.object({
  /** Enable preemptive compaction */
  preemptiveCompaction: z.boolean().default(false),
  /** Compaction threshold (0.5-0.95) */
  compactionThreshold: z.number().min(0.5).max(0.95).default(0.80),
  /** Enable aggressive tool output truncation */
  aggressiveTruncation: z.boolean().default(false),
})

/**
 * Main Ring configuration schema.
 */
export const RingConfigSchema = z.object({
  /** Schema URL for IDE support */
  $schema: z.string().optional(),

  /** Disabled hooks (won't be loaded) */
  disabled_hooks: z.array(HookNameSchema).default([]),

  /** Disabled agents (won't be available) */
  disabled_agents: z.array(AgentNameSchema).default([]),

  /** Disabled skills (won't be loaded) */
  disabled_skills: z.array(SkillNameSchema).default([]),

  /** Disabled commands (won't be registered) */
  disabled_commands: z.array(CommandNameSchema).default([]),

  /** Background task configuration */
  background_tasks: BackgroundTaskConfigSchema.default({}),

  /** Notification configuration */
  notifications: NotificationConfigSchema.default({}),

  /** Experimental features */
  experimental: ExperimentalConfigSchema.default({}),

  /** Custom hook configurations */
  hooks: z.record(z.string(), z.record(z.string(), z.unknown())).optional(),
})

/**
 * Inferred TypeScript types from schemas.
 */
export type HookName = z.infer<typeof HookNameSchema>
export type AgentName = z.infer<typeof AgentNameSchema>
export type SkillName = z.infer<typeof SkillNameSchema>
export type CommandName = z.infer<typeof CommandNameSchema>
export type BackgroundTaskConfig = z.infer<typeof BackgroundTaskConfigSchema>
export type NotificationConfig = z.infer<typeof NotificationConfigSchema>
export type ExperimentalConfig = z.infer<typeof ExperimentalConfigSchema>
export type RingConfig = z.infer<typeof RingConfigSchema>

/**
 * Default configuration values.
 */
export const DEFAULT_RING_CONFIG: RingConfig = {
  disabled_hooks: [],
  disabled_agents: [],
  disabled_skills: [],
  disabled_commands: [],
  background_tasks: {
    defaultConcurrency: 3,
    taskTimeoutMs: 1800000,
  },
  notifications: {
    enabled: true,
    onIdle: true,
    onError: true,
    onBackgroundComplete: true,
  },
  experimental: {
    preemptiveCompaction: false,
    compactionThreshold: 0.80,
    aggressiveTruncation: false,
  },
}
```

**Step 5: Verify schema compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/config/schema.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 11: Create Configuration Loader

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/config/loader.ts`

**Step 1: Write the configuration loader**

```typescript
/**
 * Ring Configuration Loader
 *
 * Implements 4-layer configuration loading with deep merge:
 * 1. Built-in defaults
 * 2. User config: ~/.config/opencode/ring/config.jsonc
 * 3. Project config: .opencode/ring.jsonc (or .ring/config.jsonc)
 * 4. Directory overrides: .ring/local.jsonc
 */

import { existsSync, readFileSync, statSync, watch, type FSWatcher } from "fs"
import { join } from "path"
import { homedir } from "os"
import { RingConfigSchema, DEFAULT_RING_CONFIG, type RingConfig } from "./schema"

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
 * Deep merge two objects, with source overriding target.
 */
function deepMerge<T extends Record<string, unknown>>(
  target: T,
  source: Partial<T>
): T {
  const result = { ...target }

  for (const key of Object.keys(source) as Array<keyof T>) {
    const sourceValue = source[key]
    const targetValue = result[key]

    if (
      sourceValue !== null &&
      typeof sourceValue === "object" &&
      !Array.isArray(sourceValue) &&
      targetValue !== null &&
      typeof targetValue === "object" &&
      !Array.isArray(targetValue)
    ) {
      result[key] = deepMerge(
        targetValue as Record<string, unknown>,
        sourceValue as Record<string, unknown>
      ) as T[keyof T]
    } else if (sourceValue !== undefined) {
      result[key] = sourceValue as T[keyof T]
    }
  }

  return result
}

/**
 * Configuration layer with metadata.
 */
interface ConfigLayer {
  name: string
  path: string
  config: Partial<RingConfig>
  mtime: number
}

/**
 * Get configuration file paths in priority order.
 */
function getConfigPaths(projectRoot: string): { name: string; path: string }[] {
  const userConfigDir = join(homedir(), ".config", "opencode", "ring")

  return [
    // Layer 2: User config
    { name: "user", path: join(userConfigDir, "config.jsonc") },
    { name: "user", path: join(userConfigDir, "config.json") },
    // Layer 3: Project config (multiple locations)
    { name: "project", path: join(projectRoot, ".opencode", "ring.jsonc") },
    { name: "project", path: join(projectRoot, ".opencode", "ring.json") },
    { name: "project", path: join(projectRoot, ".ring", "config.jsonc") },
    { name: "project", path: join(projectRoot, ".ring", "config.json") },
    // Layer 4: Directory overrides
    { name: "local", path: join(projectRoot, ".ring", "local.jsonc") },
    { name: "local", path: join(projectRoot, ".ring", "local.json") },
  ]
}

/**
 * Load a single configuration file.
 */
function loadConfigFile(path: string): Partial<RingConfig> | null {
  try {
    if (!existsSync(path)) return null
    const content = readFileSync(path, "utf-8")
    if (!content.trim()) return null

    const parsed = parseJsonc<Record<string, unknown>>(content)
    return parsed as Partial<RingConfig>
  } catch (error) {
    console.warn(`[ring-config] Failed to load ${path}:`, error)
    return null
  }
}

/**
 * Configuration cache.
 */
let cachedConfig: RingConfig | null = null
let cachedLayers: ConfigLayer[] = []
let projectRoot: string = ""
let watchers: FSWatcher[] = []

/**
 * Load all configuration layers and merge.
 */
export function loadConfig(root: string, forceReload = false): RingConfig {
  if (cachedConfig && !forceReload && root === projectRoot) {
    return cachedConfig
  }

  projectRoot = root
  const layers: ConfigLayer[] = []

  // Layer 1: Built-in defaults
  let merged: RingConfig = { ...DEFAULT_RING_CONFIG }

  // Load layers 2-4
  const paths = getConfigPaths(root)
  const seenLayers = new Set<string>()

  for (const { name, path } of paths) {
    // Only use first file found for each layer
    if (seenLayers.has(name)) continue

    const config = loadConfigFile(path)
    if (config) {
      seenLayers.add(name)
      const mtime = statSync(path).mtimeMs
      layers.push({ name, path, config, mtime })
      merged = deepMerge(merged, config)
    }
  }

  // Validate merged config
  const result = RingConfigSchema.safeParse(merged)
  if (!result.success) {
    console.warn("[ring-config] Invalid configuration:", result.error.issues)
    cachedConfig = DEFAULT_RING_CONFIG
  } else {
    cachedConfig = result.data
  }

  cachedLayers = layers
  return cachedConfig
}

/**
 * Get currently loaded configuration layers.
 */
export function getConfigLayers(): ConfigLayer[] {
  return cachedLayers
}

/**
 * Check if config files have changed.
 */
export function checkConfigChanged(): boolean {
  for (const layer of cachedLayers) {
    try {
      if (!existsSync(layer.path)) return true
      const currentMtime = statSync(layer.path).mtimeMs
      if (currentMtime !== layer.mtime) return true
    } catch {
      return true
    }
  }
  return false
}

/**
 * Start watching configuration files for changes.
 */
export function startConfigWatch(
  root: string,
  onChange: (config: RingConfig) => void
): void {
  stopConfigWatch()

  const paths = getConfigPaths(root)

  for (const { path } of paths) {
    if (!existsSync(path)) continue

    try {
      const watcher = watch(path, () => {
        const newConfig = loadConfig(root, true)
        onChange(newConfig)
      })
      watchers.push(watcher)
    } catch {
      // Watching not supported or file inaccessible
    }
  }
}

/**
 * Stop watching configuration files.
 */
export function stopConfigWatch(): void {
  for (const watcher of watchers) {
    watcher.close()
  }
  watchers = []
}

/**
 * Clear configuration cache.
 */
export function clearConfigCache(): void {
  cachedConfig = null
  cachedLayers = []
}

/**
 * Check if a hook is disabled in config.
 */
export function isHookDisabledInConfig(hookName: string): boolean {
  if (!cachedConfig) return false
  return cachedConfig.disabled_hooks.includes(hookName as never)
}

/**
 * Check if an agent is disabled in config.
 */
export function isAgentDisabledInConfig(agentName: string): boolean {
  if (!cachedConfig) return false
  return cachedConfig.disabled_agents.includes(agentName as never)
}

/**
 * Check if a skill is disabled in config.
 */
export function isSkillDisabledInConfig(skillName: string): boolean {
  if (!cachedConfig) return false
  return cachedConfig.disabled_skills.includes(skillName as never)
}

/**
 * Check if a command is disabled in config.
 */
export function isCommandDisabledInConfig(commandName: string): boolean {
  if (!cachedConfig) return false
  return cachedConfig.disabled_commands.includes(commandName as never)
}
```

**Step 2: Verify loader compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/config/loader.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 12: Create Configuration Index

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/config/index.ts`

**Step 1: Write the config index**

```typescript
/**
 * Ring Configuration System
 *
 * Central export for configuration schema, loader, and utilities.
 */

// Schema exports
export {
  RingConfigSchema,
  HookNameSchema,
  AgentNameSchema,
  SkillNameSchema,
  CommandNameSchema,
  BackgroundTaskConfigSchema,
  NotificationConfigSchema,
  ExperimentalConfigSchema,
  DEFAULT_RING_CONFIG,
} from "./schema"

export type {
  RingConfig,
  HookName,
  AgentName,
  SkillName,
  CommandName,
  BackgroundTaskConfig,
  NotificationConfig,
  ExperimentalConfig,
} from "./schema"

// Loader exports
export {
  loadConfig,
  getConfigLayers,
  checkConfigChanged,
  startConfigWatch,
  stopConfigWatch,
  clearConfigCache,
  isHookDisabledInConfig,
  isAgentDisabledInConfig,
  isSkillDisabledInConfig,
  isCommandDisabledInConfig,
} from "./loader"
```

**Step 2: Verify config module compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/config/index.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

## Part 3: Background Tasks System

### Task 13: Create Background Task Types

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background/types.ts`

**Step 1: Create background directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background
```

**Step 2: Write the background types**

```typescript
/**
 * Ring Background Task Types
 *
 * Type definitions for background agent execution system.
 */

/**
 * Background task execution status.
 */
export type BackgroundTaskStatus =
  | "pending"
  | "running"
  | "completed"
  | "error"
  | "cancelled"
  | "timeout"

/**
 * Task progress information.
 */
export interface TaskProgress {
  /** Number of tool calls made */
  toolCalls: number
  /** Last tool that was called */
  lastTool?: string
  /** Last progress update timestamp */
  lastUpdate: Date
  /** Last message from the task */
  lastMessage?: string
  /** Timestamp of last message */
  lastMessageAt?: Date
}

/**
 * Background task definition.
 */
export interface BackgroundTask {
  /** Unique task identifier */
  id: string
  /** OpenCode session ID for this task */
  sessionId: string
  /** Parent session that launched this task */
  parentSessionId: string
  /** Human-readable description */
  description: string
  /** The prompt sent to the agent */
  prompt: string
  /** Agent name to execute the task */
  agent: string
  /** Current task status */
  status: BackgroundTaskStatus
  /** When the task was created */
  createdAt: Date
  /** When the task started running */
  startedAt?: Date
  /** When the task completed */
  completedAt?: Date
  /** Task result if completed successfully */
  result?: string
  /** Error message if failed */
  error?: string
  /** Progress tracking */
  progress?: TaskProgress
  /** Model configuration */
  model?: { providerId: string; modelId: string }
  /** Concurrency tracking key */
  concurrencyKey?: string
  /** Task type for categorization */
  taskType?: BackgroundTaskType
}

/**
 * Background task types.
 */
export type BackgroundTaskType =
  | "exploration"    // Codebase exploration
  | "review"         // Code review
  | "validation"     // Test/lint validation
  | "generation"     // Code generation
  | "analysis"       // Analysis task
  | "custom"         // Custom task

/**
 * Input for launching a new background task.
 */
export interface LaunchTaskInput {
  /** Human-readable description */
  description: string
  /** Prompt to send to the agent */
  prompt: string
  /** Agent name to execute */
  agent: string
  /** Parent session ID */
  parentSessionId: string
  /** Optional model override */
  model?: { providerId: string; modelId: string }
  /** Task type for categorization */
  taskType?: BackgroundTaskType
  /** Additional skills to inject */
  skills?: string[]
  /** Skill content to inject into system prompt */
  skillContent?: string
}

/**
 * Input for resuming a background task.
 */
export interface ResumeTaskInput {
  /** Task ID to resume */
  taskId: string
  /** New prompt for continuation */
  prompt: string
  /** New parent session */
  parentSessionId: string
}

/**
 * Background task notification payload.
 */
export interface TaskNotification {
  /** Task that triggered notification */
  task: BackgroundTask
  /** Notification type */
  type: "started" | "progress" | "completed" | "error"
  /** Human-readable message */
  message: string
  /** Timestamp */
  timestamp: Date
}

/**
 * Background manager events.
 */
export interface BackgroundManagerEvents {
  "task:started": (task: BackgroundTask) => void
  "task:progress": (task: BackgroundTask) => void
  "task:completed": (task: BackgroundTask) => void
  "task:error": (task: BackgroundTask, error: Error) => void
  "task:cancelled": (task: BackgroundTask) => void
}

/**
 * OpenCode client interface (subset used by background manager).
 */
export interface BackgroundClient {
  session: {
    create: (opts: { body: { parentID?: string; title?: string } }) => Promise<{ data: { id: string }; error?: unknown }>
    prompt: (opts: {
      path: { id: string }
      body: {
        agent?: string
        model?: { providerId: string; modelId: string }
        system?: string
        tools?: Record<string, boolean>
        parts: Array<{ type: string; text: string }>
        noReply?: boolean
      }
    }) => Promise<void>
    messages: (opts: { path: { id: string } }) => Promise<{ data?: Array<unknown>; error?: unknown }>
    status: () => Promise<{ data?: Record<string, { type: string }>; error?: unknown }>
    todo: (opts: { path: { id: string } }) => Promise<{ data?: Array<{ status: string }> }>
  }
  tui: {
    showToast: (opts: {
      body: { title: string; message: string; variant: string; duration: number }
    }) => Promise<void>
  }
}
```

**Step 3: Verify types compile**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background/types.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 14: Create Concurrency Manager

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background/concurrency.ts`

**Step 1: Write the concurrency manager**

```typescript
/**
 * Ring Background Concurrency Manager
 *
 * Manages concurrent background task execution with per-agent limits.
 */

import type { BackgroundTaskConfig } from "../config/schema"

/**
 * Manages concurrency limits for background tasks.
 */
export class ConcurrencyManager {
  private activeCount: Map<string, number> = new Map()
  private waitQueue: Map<string, Array<() => void>> = new Map()
  private config: BackgroundTaskConfig

  constructor(config?: BackgroundTaskConfig) {
    this.config = config ?? {
      defaultConcurrency: 3,
      taskTimeoutMs: 1800000,
    }
  }

  /**
   * Get the concurrency limit for a specific key.
   */
  private getLimit(key: string): number {
    return this.config.agentConcurrency?.[key] ?? this.config.defaultConcurrency
  }

  /**
   * Get current active count for a key.
   */
  getActiveCount(key: string): number {
    return this.activeCount.get(key) ?? 0
  }

  /**
   * Check if a new task can be started for the given key.
   */
  canAcquire(key: string): boolean {
    const current = this.getActiveCount(key)
    const limit = this.getLimit(key)
    return current < limit
  }

  /**
   * Acquire a concurrency slot, waiting if necessary.
   */
  async acquire(key: string): Promise<void> {
    const current = this.getActiveCount(key)
    const limit = this.getLimit(key)

    if (current < limit) {
      this.activeCount.set(key, current + 1)
      return
    }

    // Wait in queue
    return new Promise<void>((resolve) => {
      const queue = this.waitQueue.get(key) ?? []
      queue.push(() => {
        this.activeCount.set(key, (this.activeCount.get(key) ?? 0) + 1)
        resolve()
      })
      this.waitQueue.set(key, queue)
    })
  }

  /**
   * Release a concurrency slot.
   */
  release(key: string): void {
    const current = this.getActiveCount(key)
    if (current <= 0) return

    this.activeCount.set(key, current - 1)

    // Process waiting queue
    const queue = this.waitQueue.get(key)
    if (queue && queue.length > 0) {
      const next = queue.shift()
      if (next) {
        next()
      }
      if (queue.length === 0) {
        this.waitQueue.delete(key)
      }
    }
  }

  /**
   * Get total active count across all keys.
   */
  getTotalActive(): number {
    let total = 0
    for (const count of this.activeCount.values()) {
      total += count
    }
    return total
  }

  /**
   * Get total waiting count across all keys.
   */
  getTotalWaiting(): number {
    let total = 0
    for (const queue of this.waitQueue.values()) {
      total += queue.length
    }
    return total
  }

  /**
   * Clear all state (for testing).
   */
  clear(): void {
    this.activeCount.clear()
    this.waitQueue.clear()
  }

  /**
   * Update configuration.
   */
  updateConfig(config: BackgroundTaskConfig): void {
    this.config = config
  }
}
```

**Step 2: Verify concurrency manager compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background/concurrency.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 15: Create Background Manager

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background/manager.ts`

**Step 1: Write the background manager**

```typescript
/**
 * Ring Background Manager
 *
 * Manages background task execution with parallel agent dispatch.
 */

import type {
  BackgroundTask,
  BackgroundTaskStatus,
  LaunchTaskInput,
  ResumeTaskInput,
  TaskNotification,
  BackgroundClient,
} from "./types"
import { ConcurrencyManager } from "./concurrency"
import type { BackgroundTaskConfig } from "../config/schema"
import { randomBytes } from "crypto"

/**
 * Generate a unique task ID.
 */
function generateTaskId(): string {
  return `bg_${randomBytes(4).toString("hex")}`
}

/**
 * Background task manager.
 */
export class BackgroundManager {
  private tasks: Map<string, BackgroundTask> = new Map()
  private notifications: Map<string, BackgroundTask[]> = new Map()
  private pendingByParent: Map<string, Set<string>> = new Map()
  private client: BackgroundClient
  private directory: string
  private concurrency: ConcurrencyManager
  private pollingInterval?: ReturnType<typeof setInterval>
  private config: BackgroundTaskConfig

  constructor(
    client: BackgroundClient,
    directory: string,
    config?: BackgroundTaskConfig
  ) {
    this.client = client
    this.directory = directory
    this.config = config ?? {
      defaultConcurrency: 3,
      taskTimeoutMs: 1800000,
    }
    this.concurrency = new ConcurrencyManager(this.config)
  }

  /**
   * Launch a new background task.
   */
  async launch(input: LaunchTaskInput): Promise<BackgroundTask> {
    const {
      description,
      prompt,
      agent,
      parentSessionId,
      model,
      taskType = "custom",
      skillContent,
    } = input

    if (!agent?.trim()) {
      throw new Error("Agent parameter is required")
    }

    // Acquire concurrency slot
    await this.concurrency.acquire(agent)

    // Create session
    let sessionId: string
    try {
      const result = await this.client.session.create({
        body: {
          parentID: parentSessionId,
          title: `Background: ${description}`,
        },
      })

      if (result.error) {
        this.concurrency.release(agent)
        throw new Error(`Failed to create session: ${result.error}`)
      }

      sessionId = result.data.id
    } catch (error) {
      this.concurrency.release(agent)
      throw error
    }

    // Create task
    const task: BackgroundTask = {
      id: generateTaskId(),
      sessionId,
      parentSessionId,
      description,
      prompt,
      agent,
      status: "running",
      createdAt: new Date(),
      startedAt: new Date(),
      model,
      concurrencyKey: agent,
      taskType,
      progress: {
        toolCalls: 0,
        lastUpdate: new Date(),
      },
    }

    this.tasks.set(task.id, task)
    this.trackPending(parentSessionId, task.id)
    this.startPolling()

    // Fire and forget - send prompt to session
    this.client.session
      .prompt({
        path: { id: sessionId },
        body: {
          agent,
          ...(model ? { model } : {}),
          system: skillContent,
          tools: {
            task: false,
            call_omo_agent: false,
          },
          parts: [{ type: "text", text: prompt }],
        },
      })
      .catch((error) => {
        this.handleTaskError(task, error)
      })

    return task
  }

  /**
   * Get a task by ID.
   */
  getTask(id: string): BackgroundTask | undefined {
    return this.tasks.get(id)
  }

  /**
   * Get all tasks for a parent session.
   */
  getTasksByParent(parentSessionId: string): BackgroundTask[] {
    return Array.from(this.tasks.values()).filter(
      (t) => t.parentSessionId === parentSessionId
    )
  }

  /**
   * Get all running tasks.
   */
  getRunningTasks(): BackgroundTask[] {
    return Array.from(this.tasks.values()).filter((t) => t.status === "running")
  }

  /**
   * Get all completed tasks.
   */
  getCompletedTasks(): BackgroundTask[] {
    return Array.from(this.tasks.values()).filter((t) => t.status !== "running")
  }

  /**
   * Find task by session ID.
   */
  findBySession(sessionId: string): BackgroundTask | undefined {
    for (const task of this.tasks.values()) {
      if (task.sessionId === sessionId) {
        return task
      }
    }
    return undefined
  }

  /**
   * Handle OpenCode events.
   */
  handleEvent(event: { type: string; properties?: Record<string, unknown> }): void {
    const { type, properties } = event

    if (type === "message.part.updated") {
      const sessionId = properties?.sessionID as string
      if (!sessionId) return

      const task = this.findBySession(sessionId)
      if (!task) return

      // Update progress
      if (!task.progress) {
        task.progress = { toolCalls: 0, lastUpdate: new Date() }
      }

      const partType = properties?.type as string
      if (partType === "tool" || properties?.tool) {
        task.progress.toolCalls += 1
        task.progress.lastTool = properties?.tool as string
        task.progress.lastUpdate = new Date()
      }
    }

    if (type === "session.idle") {
      const sessionId = properties?.sessionID as string
      if (!sessionId) return

      const task = this.findBySession(sessionId)
      if (!task || task.status !== "running") return

      // Check minimum run time
      const elapsedMs = Date.now() - (task.startedAt?.getTime() ?? 0)
      if (elapsedMs < 5000) return

      this.completeTask(task)
    }

    if (type === "session.deleted") {
      const info = properties?.info as { id?: string }
      if (!info?.id) return

      const task = this.findBySession(info.id)
      if (!task) return

      if (task.status === "running") {
        task.status = "cancelled"
        task.completedAt = new Date()
        task.error = "Session deleted"
      }

      this.releaseTask(task)
    }
  }

  /**
   * Get pending notifications for a session.
   */
  getPendingNotifications(sessionId: string): BackgroundTask[] {
    return this.notifications.get(sessionId) ?? []
  }

  /**
   * Clear notifications for a session.
   */
  clearNotifications(sessionId: string): void {
    this.notifications.delete(sessionId)
  }

  /**
   * Cleanup all resources.
   */
  cleanup(): void {
    this.stopPolling()
    this.tasks.clear()
    this.notifications.clear()
    this.pendingByParent.clear()
    this.concurrency.clear()
  }

  // Private methods

  private trackPending(parentSessionId: string, taskId: string): void {
    const pending = this.pendingByParent.get(parentSessionId) ?? new Set()
    pending.add(taskId)
    this.pendingByParent.set(parentSessionId, pending)
  }

  private async completeTask(task: BackgroundTask): Promise<void> {
    task.status = "completed"
    task.completedAt = new Date()
    this.markForNotification(task)
    await this.notifyParent(task)
  }

  private handleTaskError(task: BackgroundTask, error: unknown): void {
    task.status = "error"
    task.completedAt = new Date()
    task.error = error instanceof Error ? error.message : String(error)

    if (task.concurrencyKey) {
      this.concurrency.release(task.concurrencyKey)
    }

    this.markForNotification(task)
    this.notifyParent(task).catch(() => {})
  }

  private releaseTask(task: BackgroundTask): void {
    if (task.concurrencyKey) {
      this.concurrency.release(task.concurrencyKey)
    }
    this.tasks.delete(task.id)
    this.clearNotificationForTask(task.id)
  }

  private markForNotification(task: BackgroundTask): void {
    const queue = this.notifications.get(task.parentSessionId) ?? []
    queue.push(task)
    this.notifications.set(task.parentSessionId, queue)
  }

  private clearNotificationForTask(taskId: string): void {
    for (const [sessionId, tasks] of this.notifications.entries()) {
      const filtered = tasks.filter((t) => t.id !== taskId)
      if (filtered.length === 0) {
        this.notifications.delete(sessionId)
      } else {
        this.notifications.set(sessionId, filtered)
      }
    }
  }

  private async notifyParent(task: BackgroundTask): Promise<void> {
    const duration = this.formatDuration(task.startedAt ?? task.createdAt, task.completedAt)

    // Update pending tracking
    const pending = this.pendingByParent.get(task.parentSessionId)
    if (pending) {
      pending.delete(task.id)
      if (pending.size === 0) {
        this.pendingByParent.delete(task.parentSessionId)
      }
    }

    const allComplete = !pending || pending.size === 0
    const statusText = task.status === "error" ? "FAILED" : "COMPLETED"

    let notification: string
    if (allComplete) {
      const completedTasks = Array.from(this.tasks.values())
        .filter((t) => t.parentSessionId === task.parentSessionId && t.status !== "running")
        .map((t) => `- \`${t.id}\`: ${t.description}`)
        .join("\n")

      notification = `<system-reminder>
[ALL BACKGROUND TASKS COMPLETE]

**Completed:**
${completedTasks || `- \`${task.id}\`: ${task.description}`}

Use \`background_output(task_id="<id>")\` to retrieve each result.
</system-reminder>`
    } else {
      notification = `<system-reminder>
[BACKGROUND TASK ${statusText}]
**ID:** \`${task.id}\`
**Description:** ${task.description}
**Duration:** ${duration}

Use \`background_output(task_id="${task.id}")\` to retrieve result.
</system-reminder>`
    }

    try {
      await this.client.session.prompt({
        path: { id: task.parentSessionId },
        body: {
          noReply: !allComplete,
          parts: [{ type: "text", text: notification }],
        },
      })
    } catch {
      // Notification failed - non-critical
    }

    // Schedule cleanup
    setTimeout(() => {
      this.releaseTask(task)
    }, 5 * 60 * 1000)
  }

  private formatDuration(start: Date, end?: Date): string {
    const duration = (end ?? new Date()).getTime() - start.getTime()
    const seconds = Math.floor(duration / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)

    if (hours > 0) return `${hours}h ${minutes % 60}m`
    if (minutes > 0) return `${minutes}m ${seconds % 60}s`
    return `${seconds}s`
  }

  private startPolling(): void {
    if (this.pollingInterval) return

    this.pollingInterval = setInterval(() => {
      this.pollTasks()
    }, 2000)
    this.pollingInterval.unref()
  }

  private stopPolling(): void {
    if (this.pollingInterval) {
      clearInterval(this.pollingInterval)
      this.pollingInterval = undefined
    }
  }

  private hasRunningTasks(): boolean {
    for (const task of this.tasks.values()) {
      if (task.status === "running") return true
    }
    return false
  }

  private async pollTasks(): Promise<void> {
    // Prune stale tasks
    const now = Date.now()
    const timeoutMs = this.config.taskTimeoutMs

    for (const [taskId, task] of this.tasks.entries()) {
      const age = now - (task.startedAt?.getTime() ?? task.createdAt.getTime())
      if (age > timeoutMs && task.status === "running") {
        task.status = "timeout"
        task.completedAt = new Date()
        task.error = `Task timed out after ${Math.round(timeoutMs / 60000)} minutes`
        this.releaseTask(task)
      }
    }

    // Check session status
    try {
      const statusResult = await this.client.session.status()
      const statuses = statusResult.data ?? {}

      for (const task of this.tasks.values()) {
        if (task.status !== "running") continue

        const sessionStatus = statuses[task.sessionId]
        if (sessionStatus?.type === "idle") {
          await this.completeTask(task)
        }
      }
    } catch {
      // Polling failed - will retry
    }

    if (!this.hasRunningTasks()) {
      this.stopPolling()
    }
  }
}
```

**Step 2: Verify manager compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background/manager.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 16: Create Background Index

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background/index.ts`

**Step 1: Write the background index**

```typescript
/**
 * Ring Background Task System
 *
 * Central export for background task types, manager, and utilities.
 */

// Type exports
export type {
  BackgroundTask,
  BackgroundTaskStatus,
  BackgroundTaskType,
  TaskProgress,
  LaunchTaskInput,
  ResumeTaskInput,
  TaskNotification,
  BackgroundManagerEvents,
  BackgroundClient,
} from "./types"

// Manager exports
export { BackgroundManager } from "./manager"

// Concurrency exports
export { ConcurrencyManager } from "./concurrency"
```

**Step 2: Verify background module compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/background/index.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 17: Create Background Tasks JSON Schema

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/background-tasks.schema.json`

**Step 1: Create .opencode directory if needed**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode
```

**Step 2: Write the JSON schema**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Ring Background Tasks",
  "description": "Schema for Ring background task state persistence",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["id", "sessionId", "parentSessionId", "description", "agent", "status", "createdAt"],
    "properties": {
      "id": {
        "type": "string",
        "description": "Unique task identifier",
        "pattern": "^bg_[a-f0-9]{8}$"
      },
      "sessionId": {
        "type": "string",
        "description": "OpenCode session ID for this task"
      },
      "parentSessionId": {
        "type": "string",
        "description": "Parent session that launched this task"
      },
      "description": {
        "type": "string",
        "description": "Human-readable task description"
      },
      "prompt": {
        "type": "string",
        "description": "The prompt sent to the agent"
      },
      "agent": {
        "type": "string",
        "description": "Agent name executing the task"
      },
      "status": {
        "type": "string",
        "enum": ["pending", "running", "completed", "error", "cancelled", "timeout"],
        "description": "Current task status"
      },
      "taskType": {
        "type": "string",
        "enum": ["exploration", "review", "validation", "generation", "analysis", "custom"],
        "description": "Task categorization"
      },
      "createdAt": {
        "type": "string",
        "format": "date-time",
        "description": "Task creation timestamp"
      },
      "startedAt": {
        "type": "string",
        "format": "date-time",
        "description": "Task start timestamp"
      },
      "completedAt": {
        "type": "string",
        "format": "date-time",
        "description": "Task completion timestamp"
      },
      "error": {
        "type": "string",
        "description": "Error message if task failed"
      },
      "progress": {
        "type": "object",
        "properties": {
          "toolCalls": {
            "type": "integer",
            "minimum": 0
          },
          "lastTool": {
            "type": "string"
          },
          "lastUpdate": {
            "type": "string",
            "format": "date-time"
          }
        }
      }
    }
  }
}
```

**Step 3: Verify JSON is valid**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/background-tasks.schema.json | python3 -m json.tool > /dev/null && echo "Valid JSON"`

**Expected output:**
```
Valid JSON
```

---

### Task 18: Run Code Review Checkpoint (Part 2)

**Step 1: Dispatch code reviewers in parallel**

REQUIRED SUB-SKILL: Use requesting-code-review

Review the configuration and background task files created in Tasks 10-17:
- `plugin/config/schema.ts`
- `plugin/config/loader.ts`
- `plugin/config/index.ts`
- `plugin/background/types.ts`
- `plugin/background/concurrency.ts`
- `plugin/background/manager.ts`
- `plugin/background/index.ts`

**Step 2: Handle findings and proceed when zero Critical/High/Medium issues**

---

## Part 4: Integration and Main Plugin

### Task 19: Create Integrated Plugin Entry Point

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/ring-plugin.ts`

**Step 1: Write the integrated plugin**

```typescript
/**
 * Ring Plugin - Integrated Entry Point
 *
 * Main plugin that initializes the hook system, configuration,
 * and background manager. Replaces direct plugin exports with
 * hook-based architecture.
 */

import type { Plugin, PluginInput } from "@opencode-ai/plugin"
import { hookRegistry } from "./hooks"
import { loadConfig, startConfigWatch, isHookDisabledInConfig } from "./config"
import { BackgroundManager } from "./background"
import {
  createSessionStartHook,
  createContextInjectionHook,
  createNotificationHook,
  createTaskCompletionHook,
  setNotificationShell,
  setTaskCompletionClient,
} from "./hooks/factories"
import { getSessionId } from "./utils/state"

/**
 * Initialize all hooks based on configuration.
 */
function initializeHooks(config: ReturnType<typeof loadConfig>): void {
  // Clear existing hooks
  hookRegistry.clear()

  // Set disabled hooks from config
  hookRegistry.setDisabledHooks(config.disabled_hooks as never[])

  // Register enabled hooks
  if (!isHookDisabledInConfig("session-start")) {
    hookRegistry.register(createSessionStartHook(config.hooks?.["session-start"]))
  }

  if (!isHookDisabledInConfig("context-injection")) {
    hookRegistry.register(createContextInjectionHook(config.hooks?.["context-injection"]))
  }

  if (!isHookDisabledInConfig("notification")) {
    hookRegistry.register(createNotificationHook({
      notifyOnIdle: config.notifications.onIdle,
      notifyOnError: config.notifications.onError,
    }))
  }

  if (!isHookDisabledInConfig("task-completion")) {
    hookRegistry.register(createTaskCompletionHook())
  }
}

/**
 * Ring Plugin with Hook Architecture
 */
export const RingPlugin: Plugin = async (ctx: PluginInput) => {
  const { directory, $, client } = ctx

  // Load configuration
  const config = loadConfig(directory)

  // Initialize hooks
  initializeHooks(config)

  // Set up hook dependencies
  setNotificationShell($)
  setTaskCompletionClient(client)

  // Initialize background manager
  const backgroundManager = new BackgroundManager(
    client as never,
    directory,
    config.background_tasks
  )

  // Start config watch for hot-reload
  startConfigWatch(directory, (newConfig) => {
    initializeHooks(newConfig)
    backgroundManager.cleanup()
  })

  const sessionId = getSessionId()

  return {
    // Session lifecycle events
    event: async ({ event }) => {
      const ctx = {
        sessionId,
        directory,
        event,
      }

      // Route to background manager
      backgroundManager.handleEvent(event)

      // Execute hooks for this event type
      const output = {}

      if (event.type === "session.created") {
        await hookRegistry.executeLifecycle("session.created", ctx, output)
      }

      if (event.type === "session.idle") {
        await hookRegistry.executeLifecycle("session.idle", ctx, output)
      }

      if (event.type === "todo.updated") {
        await hookRegistry.executeLifecycle("todo.updated", ctx, output)
      }
    },

    // System prompt transformation
    "experimental.chat.system.transform": async (
      _input: Record<string, unknown>,
      output: { system: string[] }
    ) => {
      if (!output?.system || !Array.isArray(output.system)) return

      const ctx = { sessionId, directory }
      await hookRegistry.executeLifecycle("chat.params", ctx, output)
    },

    // Compaction context injection
    "experimental.session.compacting": async (
      input: { sessionID: string },
      output: { context: string[] }
    ) => {
      if (!output?.context || !Array.isArray(output.context)) return

      const ctx = { sessionId: input.sessionID, directory }
      await hookRegistry.executeLifecycle("session.compacting", ctx, output)
    },

    // Expose background manager for tools
    getBackgroundManager: () => backgroundManager,
  }
}

export default RingPlugin
```

**Step 2: Verify plugin compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/ring-plugin.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 20: Update Main Plugin Index

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/index.ts`

**Step 1: Update the index to export the new architecture**

```typescript
/**
 * Ring OpenCode Plugins
 *
 * This module exports Ring plugins for OpenCode.
 *
 * Architecture:
 * - Hook system: Middleware pattern with lifecycle events
 * - Layered config: 4-layer priority with deep merge
 * - Background tasks: Parallel agent execution manager
 *
 * Legacy plugins are still exported for backward compatibility.
 */

// New hook-based architecture
export { RingPlugin, RingPlugin as default } from "./ring-plugin"

// Hook system exports
export { hookRegistry, isHookDisabled } from "./hooks"
export type {
  Hook,
  HookFactory,
  HookName,
  HookLifecycle,
  HookContext,
  HookOutput,
  HookResult,
} from "./hooks"

// Hook factories
export {
  createSessionStartHook,
  createContextInjectionHook,
  createNotificationHook,
  createTaskCompletionHook,
} from "./hooks/factories"

// Configuration exports
export {
  loadConfig,
  getConfigLayers,
  startConfigWatch,
  stopConfigWatch,
  isHookDisabledInConfig,
  isAgentDisabledInConfig,
  isSkillDisabledInConfig,
  isCommandDisabledInConfig,
  DEFAULT_RING_CONFIG,
} from "./config"

export type {
  RingConfig,
  HookName as ConfigHookName,
  AgentName,
  SkillName,
  CommandName,
  BackgroundTaskConfig,
} from "./config"

// Background task exports
export { BackgroundManager, ConcurrencyManager } from "./background"
export type {
  BackgroundTask,
  BackgroundTaskStatus,
  BackgroundTaskType,
  LaunchTaskInput,
} from "./background"

// Legacy plugin exports (for backward compatibility)
export { RingSessionStart } from "./session-start"
export { RingContextInjection } from "./context-injection"
export { RingNotification } from "./notification"
export { RingTaskCompletionCheck } from "./task-completion-check"
export { RingSessionOutcome } from "./session-outcome"
export { RingOutcomeInference } from "./outcome-inference"
export { RingDoubtResolver } from "./doubt-resolver"
```

**Step 2: Verify index compiles**

Run: `bun build /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/index.ts --no-bundle --outdir=/tmp/ring-test 2>&1`

---

### Task 21: Create Sample Configuration File

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.ring/config.jsonc`

**Step 1: Create .ring directory**

```bash
mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.ring
```

**Step 2: Write the sample configuration**

```jsonc
{
  // Ring Configuration
  // Layers: defaults < ~/.config/opencode/ring/config.jsonc < .opencode/ring.jsonc < .ring/local.jsonc
  "$schema": "../.opencode/ring-config.schema.json",

  // Hooks to disable (comment out to enable)
  "disabled_hooks": [
    // "session-start",
    // "context-injection",
    // "notification",
    // "task-completion"
  ],

  // Agents to disable
  "disabled_agents": [
    // "code-reviewer",
    // "security-reviewer"
  ],

  // Skills to disable
  "disabled_skills": [
    // "test-driven-development"
  ],

  // Commands to disable
  "disabled_commands": [
    // "commit"
  ],

  // Background task configuration
  "background_tasks": {
    "defaultConcurrency": 3,
    "taskTimeoutMs": 1800000,
    "agentConcurrency": {
      "codebase-explorer": 1,
      "write-plan": 1
    }
  },

  // Notification configuration
  "notifications": {
    "enabled": true,
    "onIdle": true,
    "onError": true,
    "onBackgroundComplete": true
  },

  // Experimental features
  "experimental": {
    "preemptiveCompaction": false,
    "compactionThreshold": 0.80,
    "aggressiveTruncation": false
  }
}
```

**Step 3: Verify JSONC is valid**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.ring/config.jsonc | sed 's/\/\/.*$//' | python3 -m json.tool > /dev/null && echo "Valid JSONC"`

**Expected output:**
```
Valid JSONC
```

---

### Task 22: Create Ring Config JSON Schema

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/ring-config.schema.json`

**Step 1: Write the JSON schema for IDE support**

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Ring Configuration",
  "description": "Configuration schema for Ring OpenCode plugin",
  "type": "object",
  "properties": {
    "$schema": {
      "type": "string",
      "description": "JSON Schema reference"
    },
    "disabled_hooks": {
      "type": "array",
      "description": "Hooks to disable",
      "items": {
        "type": "string",
        "enum": [
          "session-start",
          "context-injection",
          "notification",
          "task-completion",
          "session-outcome",
          "outcome-inference",
          "doubt-resolver",
          "background-notification",
          "compaction-context",
          "rules-injector",
          "agent-reminder"
        ]
      }
    },
    "disabled_agents": {
      "type": "array",
      "description": "Agents to disable",
      "items": {
        "type": "string",
        "enum": [
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
          "devops-engineer",
          "sre",
          "qa-analyst"
        ]
      }
    },
    "disabled_skills": {
      "type": "array",
      "description": "Skills to disable",
      "items": {
        "type": "string"
      }
    },
    "disabled_commands": {
      "type": "array",
      "description": "Commands to disable",
      "items": {
        "type": "string"
      }
    },
    "background_tasks": {
      "type": "object",
      "description": "Background task configuration",
      "properties": {
        "defaultConcurrency": {
          "type": "integer",
          "minimum": 1,
          "maximum": 10,
          "default": 3
        },
        "taskTimeoutMs": {
          "type": "integer",
          "minimum": 60000,
          "maximum": 3600000,
          "default": 1800000
        },
        "agentConcurrency": {
          "type": "object",
          "additionalProperties": {
            "type": "integer",
            "minimum": 1,
            "maximum": 10
          }
        }
      }
    },
    "notifications": {
      "type": "object",
      "description": "Notification configuration",
      "properties": {
        "enabled": { "type": "boolean", "default": true },
        "onIdle": { "type": "boolean", "default": true },
        "onError": { "type": "boolean", "default": true },
        "onBackgroundComplete": { "type": "boolean", "default": true }
      }
    },
    "experimental": {
      "type": "object",
      "description": "Experimental features",
      "properties": {
        "preemptiveCompaction": { "type": "boolean", "default": false },
        "compactionThreshold": {
          "type": "number",
          "minimum": 0.5,
          "maximum": 0.95,
          "default": 0.80
        },
        "aggressiveTruncation": { "type": "boolean", "default": false }
      }
    },
    "hooks": {
      "type": "object",
      "description": "Per-hook configuration overrides",
      "additionalProperties": {
        "type": "object"
      }
    }
  }
}
```

**Step 2: Verify schema is valid JSON**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/.opencode/ring-config.schema.json | python3 -m json.tool > /dev/null && echo "Valid JSON"`

---

### Task 23: Final Code Review Checkpoint

**Step 1: Dispatch all 5 reviewers in parallel**

REQUIRED SUB-SKILL: Use requesting-code-review

Review all files created in this plan:
- Hook system: `plugin/hooks/*.ts`, `plugin/hooks/factories/*.ts`
- Config system: `plugin/config/*.ts`
- Background system: `plugin/background/*.ts`
- Integration: `plugin/ring-plugin.ts`, `plugin/index.ts`
- Schemas: `.opencode/*.json`, `.ring/config.jsonc`

**Step 2: Handle all findings by severity**

**Step 3: Proceed only when ALL issues are addressed**

---

### Task 24: Run Full Build and Tests

**Step 1: Run TypeScript build**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build plugin/index.ts --no-bundle --outdir=/tmp/ring-build
```

**Expected output:**
```
(no errors)
```

**Step 2: Run existing tests**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test plugin/
```

**Expected output:**
```
(all tests pass)
```

**Step 3: Verify hook registry**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun -e "
import { hookRegistry } from './plugin/hooks';
import { createSessionStartHook } from './plugin/hooks/factories';

const hook = createSessionStartHook();
hookRegistry.register(hook);
console.log('Registered hooks:', hookRegistry.getRegisteredNames());
console.log('Session-start enabled:', hookRegistry.has('session-start'));
"
```

**Expected output:**
```
Registered hooks: [ 'session-start' ]
Session-start enabled: true
```

---

### Task 25: Commit Changes

**Step 1: Stage all new files**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/hooks/ plugin/config/ plugin/background/ plugin/ring-plugin.ts plugin/index.ts .ring/config.jsonc .opencode/*.json
```

**Step 2: Create commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git commit -m "$(cat <<'EOF'
feat(plugin): implement hook architecture, layered config, and background tasks

- Add hook system with lifecycle types and factory pattern
- Create 4-layer configuration loader with deep merge
- Implement BackgroundManager for parallel agent execution
- Convert existing plugins to hook factories
- Add JSON schemas for config and background tasks
- Maintain backward compatibility with legacy plugin exports
EOF
)"
```

**Step 3: Verify commit**

Run: `git log -1 --oneline`

**Expected output:**
```
<hash> feat(plugin): implement hook architecture, layered config, and background tasks
```

---

## Summary

This plan implements three major features:

1. **Hook Architecture** (Tasks 1-9)
   - Type definitions for hook lifecycle events
   - Central hook registry with enable/disable support
   - Factory pattern for hook creation
   - Converted 4 existing plugins to hook factories

2. **Layered Configuration** (Tasks 10-12)
   - Zod schema validation
   - 4-layer config loading with deep merge
   - Hot-reload support via file watching
   - disabled_hooks, disabled_agents, disabled_skills, disabled_commands arrays

3. **Background Tasks** (Tasks 13-17)
   - Type definitions for background tasks
   - Concurrency manager for parallel execution
   - Background manager with notification system
   - JSON schema for task persistence

4. **Integration** (Tasks 18-25)
   - Unified plugin entry point
   - Sample configuration file
   - JSON schemas for IDE support
   - Code review checkpoints
   - Final verification and commit

**Total estimated time:** 2-3 hours for complete implementation

**Key files created:**
- `plugin/hooks/types.ts` - Hook type definitions
- `plugin/hooks/registry.ts` - Hook registration and execution
- `plugin/hooks/factories/*.ts` - Hook factory implementations
- `plugin/config/schema.ts` - Zod configuration schema
- `plugin/config/loader.ts` - Layered config loader
- `plugin/background/manager.ts` - Background task manager
- `plugin/ring-plugin.ts` - Integrated plugin entry
- `.ring/config.jsonc` - Sample configuration
- `.opencode/ring-config.schema.json` - Config JSON schema
