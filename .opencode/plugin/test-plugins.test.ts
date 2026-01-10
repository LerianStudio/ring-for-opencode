/**
 * Ring OpenCode Plugin Behavioral Tests
 *
 * Tests actual plugin behavior, not just exports.
 *
 * Run with: bun test .opencode/plugin/test-plugins.test.ts
 *
 * For simple smoke tests (backward compatibility):
 *   bun .opencode/plugin/test-plugins.ts
 */

import { describe, test, expect, mock, beforeEach, afterEach } from "bun:test"
import { mkdirSync, rmSync, writeFileSync, existsSync, readFileSync } from "fs"
import { join } from "path"

// Import plugins
import {
  RingContextInjection,
  RingNotification,
  RingDoubtResolver,
  RingSessionStart,
  RingTaskCompletionCheck,
  RingSessionOutcome,
  RingOutcomeInference,
} from "./index"

// Test directory for file system operations
const TEST_DIR = "/tmp/ring-plugin-tests"
const TEST_STATE_DIR = join(TEST_DIR, ".opencode/state")
const TEST_LEDGER_DIR = join(TEST_DIR, ".ring/ledgers")
const TEST_SKILL_DIR = join(TEST_DIR, ".opencode/skill")

// Helper to create mock context
function createMockContext(overrides: Record<string, unknown> = {}) {
  const mockSessionList = mock(() =>
    Promise.resolve({
      data: [{ id: "test-session-1", updatedAt: Date.now(), name: "Test Session" }],
      error: null,
    })
  )

  const mockSessionPrompt = mock(() => Promise.resolve({ success: true }))
  const mockShowToast = mock(() => Promise.resolve({}))
  const mockShellExec = mock(() =>
    Promise.resolve({
      text: () => Promise.resolve(""),
      quiet: () => Promise.resolve({ text: () => Promise.resolve("") }),
      exitCode: 0,
    })
  )

  const mockShell = Object.assign(
    (strings: TemplateStringsArray, ...values: unknown[]) => mockShellExec(),
    { quiet: () => mockShellExec() }
  )

  return {
    project: { name: "test-project", path: TEST_DIR },
    client: {
      session: {
        list: mockSessionList,
        prompt: mockSessionPrompt,
      },
      tui: {
        showToast: mockShowToast,
      },
    },
    $: mockShell,
    directory: TEST_DIR,
    worktree: TEST_DIR,
    // Expose mocks for verification
    __mocks: {
      sessionList: mockSessionList,
      sessionPrompt: mockSessionPrompt,
      showToast: mockShowToast,
      shellExec: mockShellExec,
    },
    ...overrides,
  }
}

// Fixed session ID for testing
const TEST_SESSION_ID = "test-session-12345"

// Helper to setup test environment
function setupTestEnv() {
  process.env.OPENCODE_SESSION_ID = TEST_SESSION_ID
  ;(globalThis as any)._ringSecureSessionId = undefined

  if (existsSync(TEST_DIR)) {
    rmSync(TEST_DIR, { recursive: true, force: true })
  }
  mkdirSync(TEST_STATE_DIR, { recursive: true })
}

// Helper to teardown test environment
function teardownTestEnv() {
  if (existsSync(TEST_DIR)) {
    rmSync(TEST_DIR, { recursive: true, force: true })
  }
  delete process.env.OPENCODE_SESSION_ID
  ;(globalThis as any)._ringSecureSessionId = undefined
}

// Setup and teardown for each test
beforeEach(() => setupTestEnv())
afterEach(() => teardownTestEnv())

// ============================================================================
// RingContextInjection Tests
// ============================================================================

describe("RingContextInjection", () => {
  test("injects critical rules on compaction", async () => {
    const ctx = createMockContext()
    const plugin = await RingContextInjection(ctx as any)
    const output = { context: [] as string[] }

    await plugin["experimental.session.compacting"]!({}, output)

    expect(output.context.length).toBeGreaterThan(0)
    expect(output.context.some((c) => c.includes("Ring Critical Rules"))).toBe(true)
  })

  test("injects skills system reference on compaction", async () => {
    const ctx = createMockContext()
    const plugin = await RingContextInjection(ctx as any)
    const output = { context: [] as string[] }

    await plugin["experimental.session.compacting"]!({}, output)

    expect(output.context.some((c) => c.includes("Ring Skills System"))).toBe(true)
    expect(output.context.some((c) => c.includes("test-driven-development"))).toBe(true)
  })

  test("handles missing output.context gracefully", async () => {
    const ctx = createMockContext()
    const plugin = await RingContextInjection(ctx as any)

    // Should not throw with undefined context
    await expect(
      plugin["experimental.session.compacting"]!({}, { context: undefined as any })
    ).resolves.toBeUndefined()

    // Should not throw with null context
    await expect(
      plugin["experimental.session.compacting"]!({}, { context: null as any })
    ).resolves.toBeUndefined()
  })

  test("injects ledger summary when ledger exists", async () => {
    // Create a test ledger
    mkdirSync(TEST_LEDGER_DIR, { recursive: true })
    writeFileSync(
      join(TEST_LEDGER_DIR, "CONTINUITY-test-feature.md"),
      `## BRIEF_SUMMARY
Testing feature implementation

## State
[->] Phase 2: Implementation

## Open Questions (UNCONFIRMED)
- [ ] Should we use feature flags?
- [ ] What about backward compatibility?
`
    )

    const ctx = createMockContext()
    const plugin = await RingContextInjection(ctx as any)
    const output = { context: [] as string[] }

    await plugin["experimental.session.compacting"]!({}, output)

    expect(output.context.some((c) => c.includes("Active Ledger"))).toBe(true)
    expect(output.context.some((c) => c.includes("CONTINUITY-test-feature"))).toBe(true)
  })
})

// ============================================================================
// RingNotification Tests
// ============================================================================

describe("RingNotification", () => {
  test("has event hook", async () => {
    const ctx = createMockContext()
    const plugin = await RingNotification(ctx as any)

    expect(plugin.event).toBeDefined()
    expect(typeof plugin.event).toBe("function")
  })

  test("handles session.idle event", async () => {
    const ctx = createMockContext()
    const plugin = await RingNotification(ctx as any)

    // Should not throw
    await expect(plugin.event!({ event: { type: "session.idle" } })).resolves.toBeUndefined()
  })

  test("handles session.error event", async () => {
    const ctx = createMockContext()
    const plugin = await RingNotification(ctx as any)

    // Should not throw
    await expect(plugin.event!({ event: { type: "session.error" } })).resolves.toBeUndefined()
  })

  test("ignores unrelated events", async () => {
    const ctx = createMockContext()
    const plugin = await RingNotification(ctx as any)

    // Should not throw for unrelated event
    await expect(plugin.event!({ event: { type: "some.other.event" } })).resolves.toBeUndefined()
  })
})

// ============================================================================
// RingDoubtResolver Tests
// ============================================================================

describe("RingDoubtResolver", () => {
  test("provides ring_doubt tool", async () => {
    const ctx = createMockContext()
    const plugin = await RingDoubtResolver(ctx as any)

    expect(plugin.tool).toBeDefined()
    expect(plugin.tool!.ring_doubt).toBeDefined()
  })

  test("ring_doubt returns formatted prompt with options", async () => {
    const ctx = createMockContext()
    const plugin = await RingDoubtResolver(ctx as any)

    const result = await plugin.tool!.ring_doubt.execute({
      question: "Which database should I use?",
      options: ["PostgreSQL", "MongoDB", "SQLite"],
    })

    expect(result).toContain("Which database should I use?")
    expect(result).toContain("1) PostgreSQL")
    expect(result).toContain("2) MongoDB")
    expect(result).toContain("3) SQLite")
    expect(result).toContain("DOUBT CHECKPOINT")
  })

  test("ring_doubt supports multi-select mode", async () => {
    const ctx = createMockContext()
    const plugin = await RingDoubtResolver(ctx as any)

    const result = await plugin.tool!.ring_doubt.execute({
      question: "Which features to enable?",
      options: ["Logging", "Metrics", "Tracing"],
      multi: true,
    })

    expect(result).toContain('one or more numbers')
    expect(result).toContain('"1,3"')
  })

  test("ring_doubt supports context parameter", async () => {
    const ctx = createMockContext()
    const plugin = await RingDoubtResolver(ctx as any)

    const result = await plugin.tool!.ring_doubt.execute({
      question: "Which approach?",
      options: ["Option A", "Option B"],
      context: "This affects the entire system architecture",
    })

    expect(result).toContain("Context:")
    expect(result).toContain("This affects the entire system architecture")
  })

  test("ring_doubt includes custom option instructions when allowCustom is true", async () => {
    const ctx = createMockContext()
    const plugin = await RingDoubtResolver(ctx as any)

    const result = await plugin.tool!.ring_doubt.execute({
      question: "Pick one",
      options: ["A", "B"],
      allowCustom: true,
    })

    expect(result).toContain('Or type your own answer')
  })
})

// ============================================================================
// RingSessionStart Tests
// ============================================================================

describe("RingSessionStart", () => {
  test("has event hook for session lifecycle", async () => {
    const ctx = createMockContext()
    const plugin = await RingSessionStart(ctx as any)

    expect(plugin.event).toBeDefined()
    expect(typeof plugin.event).toBe("function")
  })

  test("has system transform hook", async () => {
    const ctx = createMockContext()
    const plugin = await RingSessionStart(ctx as any)

    expect(plugin["experimental.chat.system.transform"]).toBeDefined()
  })

  test("system transform injects critical rules", async () => {
    mkdirSync(TEST_SKILL_DIR, { recursive: true })

    const ctx = createMockContext()
    const plugin = await RingSessionStart(ctx as any)
    const output = { system: [] as string[] }

    await plugin["experimental.chat.system.transform"]!({}, output)

    expect(output.system.length).toBeGreaterThan(0)
    expect(output.system.some((s) => s.includes("3-FILE RULE"))).toBe(true)
  })

  test("system transform injects agent reminder", async () => {
    mkdirSync(TEST_SKILL_DIR, { recursive: true })

    const ctx = createMockContext()
    const plugin = await RingSessionStart(ctx as any)
    const output = { system: [] as string[] }

    await plugin["experimental.chat.system.transform"]!({}, output)

    expect(output.system.some((s) => s.includes("AGENT USAGE REMINDER"))).toBe(true)
  })

  test("system transform injects duplication guard", async () => {
    mkdirSync(TEST_SKILL_DIR, { recursive: true })

    const ctx = createMockContext()
    const plugin = await RingSessionStart(ctx as any)
    const output = { system: [] as string[] }

    await plugin["experimental.chat.system.transform"]!({}, output)

    expect(output.system.some((s) => s.includes("DUPLICATION PREVENTION"))).toBe(true)
  })

  test("system transform injects ledger when present", async () => {
    mkdirSync(TEST_SKILL_DIR, { recursive: true })
    mkdirSync(TEST_LEDGER_DIR, { recursive: true })
    writeFileSync(
      join(TEST_LEDGER_DIR, "CONTINUITY-my-task.md"),
      `## State
[->] Phase 1: Planning
`
    )

    const ctx = createMockContext()
    const plugin = await RingSessionStart(ctx as any)
    const output = { system: [] as string[] }

    await plugin["experimental.chat.system.transform"]!({}, output)

    expect(output.system.some((s) => s.includes("ring-continuity-ledger"))).toBe(true)
    expect(output.system.some((s) => s.includes("CONTINUITY-my-task"))).toBe(true)
  })

  test("handles session.created event", async () => {
    const ctx = createMockContext()
    const plugin = await RingSessionStart(ctx as any)

    // Should not throw
    await expect(plugin.event!({ event: { type: "session.created" } })).resolves.toBeUndefined()
  })
})

// ============================================================================
// RingTaskCompletionCheck Tests
// ============================================================================

describe("RingTaskCompletionCheck", () => {
  test("has event hook", async () => {
    const ctx = createMockContext()
    const plugin = await RingTaskCompletionCheck(ctx as any)

    expect(plugin.event).toBeDefined()
  })

  test("shows toast when all todos are complete", async () => {
    const ctx = createMockContext()
    const plugin = await RingTaskCompletionCheck(ctx as any)

    const todos = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "completed" },
      { content: "Task 3", status: "completed" },
    ]

    await plugin.event!({
      event: {
        type: "todo.updated",
        properties: { todos },
      },
    })

    expect(ctx.__mocks.showToast).toHaveBeenCalled()
  })

  test("does not show toast when todos are incomplete", async () => {
    const ctx = createMockContext()
    const plugin = await RingTaskCompletionCheck(ctx as any)

    const todos = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "in_progress" },
      { content: "Task 3", status: "pending" },
    ]

    await plugin.event!({
      event: {
        type: "todo.updated",
        properties: { todos },
      },
    })

    expect(ctx.__mocks.showToast).not.toHaveBeenCalled()
  })

  test("ignores non-todo events", async () => {
    const ctx = createMockContext()
    const plugin = await RingTaskCompletionCheck(ctx as any)

    await plugin.event!({
      event: {
        type: "some.other.event",
        properties: {},
      },
    })

    expect(ctx.__mocks.showToast).not.toHaveBeenCalled()
  })

  test("handles empty todos array", async () => {
    const ctx = createMockContext()
    const plugin = await RingTaskCompletionCheck(ctx as any)

    await plugin.event!({
      event: {
        type: "todo.updated",
        properties: { todos: [] },
      },
    })

    expect(ctx.__mocks.showToast).not.toHaveBeenCalled()
  })

  test("persists todos state to file", async () => {
    const ctx = createMockContext()
    const plugin = await RingTaskCompletionCheck(ctx as any)

    const todos = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "pending" },
    ]

    await plugin.event!({
      event: {
        type: "todo.updated",
        properties: { todos },
      },
    })

    // Check that state was written
    const stateFiles = existsSync(TEST_STATE_DIR)
      ? require("fs").readdirSync(TEST_STATE_DIR)
      : []
    const todosStateFile = stateFiles.find((f: string) => f.startsWith("todos-state"))

    expect(todosStateFile).toBeDefined()
  })
})

// ============================================================================
// RingSessionOutcome Tests
// ============================================================================

describe("RingSessionOutcome", () => {
  test("has session.compacted hook", async () => {
    const ctx = createMockContext()
    const plugin = await RingSessionOutcome(ctx as any)

    expect(plugin["session.compacted"]).toBeDefined()
  })

  test("has event hook", async () => {
    const ctx = createMockContext()
    const plugin = await RingSessionOutcome(ctx as any)

    expect(plugin.event).toBeDefined()
  })

  test("writes pending-outcome-prompt on session.compacted", async () => {
    const ctx = createMockContext()
    const plugin = await RingSessionOutcome(ctx as any)

    await plugin["session.compacted"]!()

    // Check state file was created
    const stateFiles = existsSync(TEST_STATE_DIR)
      ? require("fs").readdirSync(TEST_STATE_DIR)
      : []
    const pendingFile = stateFiles.find((f: string) => f.startsWith("pending-outcome-prompt"))

    expect(pendingFile).toBeDefined()
  })

  test("does not inject prompt when no session work detected", async () => {
    const ctx = createMockContext()
    const plugin = await RingSessionOutcome(ctx as any)

    await plugin.event!({ event: { type: "session.created" } })

    // No prompt should be injected without work indicators
    expect(ctx.__mocks.sessionPrompt).not.toHaveBeenCalled()
  })
})

// ============================================================================
// RingOutcomeInference Tests
// ============================================================================

describe("RingOutcomeInference", () => {
  test("has event hook", async () => {
    const ctx = createMockContext()
    const plugin = await RingOutcomeInference(ctx as any)

    expect(plugin.event).toBeDefined()
  })

  test("ignores non-idle events", async () => {
    const ctx = createMockContext()
    const plugin = await RingOutcomeInference(ctx as any)

    await plugin.event!({ event: { type: "session.created" } })

    // No outcome file should be created for non-idle events
    const stateFiles = existsSync(TEST_STATE_DIR)
      ? require("fs").readdirSync(TEST_STATE_DIR)
      : []
    const outcomeFile = stateFiles.find((f: string) => f.startsWith("session-outcome"))

    expect(outcomeFile).toBeUndefined()
  })

  test("infers SUCCEEDED when all todos complete", async () => {
    const ctx = createMockContext()

    // Write todos state first (using test session ID)
    const todos = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "completed" },
    ]
    writeFileSync(
      join(TEST_STATE_DIR, `todos-state-${TEST_SESSION_ID}.json`),
      JSON.stringify(todos)
    )

    const plugin = await RingOutcomeInference(ctx as any)
    await plugin.event!({ event: { type: "session.idle" } })

    // Check outcome was written
    const stateFiles = require("fs").readdirSync(TEST_STATE_DIR)
    const outcomeFile = stateFiles.find((f: string) => f.startsWith("session-outcome"))

    expect(outcomeFile).toBeDefined()

    const outcome = JSON.parse(
      readFileSync(join(TEST_STATE_DIR, outcomeFile), "utf-8")
    )
    expect(outcome.status).toBe("SUCCEEDED")
  })

  test("infers UNKNOWN when no todos state exists", async () => {
    const ctx = createMockContext()
    const plugin = await RingOutcomeInference(ctx as any)

    await plugin.event!({ event: { type: "session.idle" } })

    // Check outcome was written
    const stateFiles = require("fs").readdirSync(TEST_STATE_DIR)
    const outcomeFile = stateFiles.find((f: string) => f.startsWith("session-outcome"))

    expect(outcomeFile).toBeDefined()

    const outcome = JSON.parse(
      readFileSync(join(TEST_STATE_DIR, outcomeFile), "utf-8")
    )
    // When no todos file exists (null), returns UNKNOWN
    expect(outcome.status).toBe("UNKNOWN")
    expect(outcome.summary).toContain("Unable to determine")
  })

  test("infers FAILED when todos exist but all empty", async () => {
    const ctx = createMockContext()

    // Write empty todos array (not null)
    writeFileSync(
      join(TEST_STATE_DIR, `todos-state-${TEST_SESSION_ID}.json`),
      JSON.stringify([])
    )

    const plugin = await RingOutcomeInference(ctx as any)
    await plugin.event!({ event: { type: "session.idle" } })

    const stateFiles = require("fs").readdirSync(TEST_STATE_DIR)
    const outcomeFile = stateFiles.find((f: string) => f.startsWith("session-outcome"))
    const outcome = JSON.parse(
      readFileSync(join(TEST_STATE_DIR, outcomeFile), "utf-8")
    )

    // Empty array = 0 todos = FAILED
    expect(outcome.status).toBe("FAILED")
    expect(outcome.summary).toContain("No todos tracked")
  })

  test("infers PARTIAL_PLUS when >= 80% complete", async () => {
    const ctx = createMockContext()

    const todos = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "completed" },
      { content: "Task 3", status: "completed" },
      { content: "Task 4", status: "completed" },
      { content: "Task 5", status: "pending" },
    ]
    writeFileSync(
      join(TEST_STATE_DIR, `todos-state-${TEST_SESSION_ID}.json`),
      JSON.stringify(todos)
    )

    const plugin = await RingOutcomeInference(ctx as any)
    await plugin.event!({ event: { type: "session.idle" } })

    const stateFiles = require("fs").readdirSync(TEST_STATE_DIR)
    const outcomeFile = stateFiles.find((f: string) => f.startsWith("session-outcome"))
    const outcome = JSON.parse(
      readFileSync(join(TEST_STATE_DIR, outcomeFile), "utf-8")
    )

    expect(outcome.status).toBe("PARTIAL_PLUS")
  })

  test("infers PARTIAL_MINUS when 50-79% complete", async () => {
    const ctx = createMockContext()

    const todos = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "completed" },
      { content: "Task 3", status: "pending" },
      { content: "Task 4", status: "pending" },
    ]
    writeFileSync(
      join(TEST_STATE_DIR, `todos-state-${TEST_SESSION_ID}.json`),
      JSON.stringify(todos)
    )

    const plugin = await RingOutcomeInference(ctx as any)
    await plugin.event!({ event: { type: "session.idle" } })

    const stateFiles = require("fs").readdirSync(TEST_STATE_DIR)
    const outcomeFile = stateFiles.find((f: string) => f.startsWith("session-outcome"))
    const outcome = JSON.parse(
      readFileSync(join(TEST_STATE_DIR, outcomeFile), "utf-8")
    )

    expect(outcome.status).toBe("PARTIAL_MINUS")
  })

  test("infers FAILED when < 50% complete", async () => {
    const ctx = createMockContext()

    const todos = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "pending" },
      { content: "Task 3", status: "pending" },
      { content: "Task 4", status: "pending" },
      { content: "Task 5", status: "pending" },
    ]
    writeFileSync(
      join(TEST_STATE_DIR, `todos-state-${TEST_SESSION_ID}.json`),
      JSON.stringify(todos)
    )

    const plugin = await RingOutcomeInference(ctx as any)
    await plugin.event!({ event: { type: "session.idle" } })

    const stateFiles = require("fs").readdirSync(TEST_STATE_DIR)
    const outcomeFile = stateFiles.find((f: string) => f.startsWith("session-outcome"))
    const outcome = JSON.parse(
      readFileSync(join(TEST_STATE_DIR, outcomeFile), "utf-8")
    )

    expect(outcome.status).toBe("FAILED")
  })
})

// ============================================================================
// Integration: All plugins export correctly
// ============================================================================

describe("Plugin Exports", () => {
  test("all plugins are functions", () => {
    expect(typeof RingContextInjection).toBe("function")
    expect(typeof RingNotification).toBe("function")
    expect(typeof RingDoubtResolver).toBe("function")
    expect(typeof RingSessionStart).toBe("function")
    expect(typeof RingTaskCompletionCheck).toBe("function")
    expect(typeof RingSessionOutcome).toBe("function")
    expect(typeof RingOutcomeInference).toBe("function")
  })

  test("all plugins return hooks objects", async () => {
    const ctx = createMockContext()

    const plugins = [
      RingContextInjection,
      RingNotification,
      RingDoubtResolver,
      RingSessionStart,
      RingTaskCompletionCheck,
      RingSessionOutcome,
      RingOutcomeInference,
    ]

    for (const plugin of plugins) {
      const hooks = await plugin(ctx as any)
      expect(typeof hooks).toBe("object")
    }
  })
})

// For direct execution (smoke tests), use: bun .opencode/plugin/test-plugins.ts
