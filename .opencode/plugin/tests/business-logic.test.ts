/**
 * Business Logic and State Management Tests
 *
 * Tests for outcome inference logic and state persistence.
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test"
import { mkdtempSync, rmSync, existsSync, readFileSync, mkdirSync, writeFileSync, readdirSync } from "fs"
import { tmpdir } from "os"
import { join } from "path"

// Import functions under test from source
import { analyzeTodos, inferOutcomeFromTodos, OUTCOME } from "../outcome-inference"
import type { Todo, OutcomeType } from "../outcome-inference"

// ============================================================================
// analyzeTodos Tests
// ============================================================================

describe("analyzeTodos", () => {
  test("counts completed, in_progress, pending correctly", () => {
    const todos: Todo[] = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "completed" },
      { content: "Task 3", status: "in_progress" },
      { content: "Task 4", status: "pending" },
      { content: "Task 5", status: "pending" },
    ]

    const result = analyzeTodos(todos)

    expect(result.total).toBe(5)
    expect(result.completed).toBe(2)
    expect(result.inProgress).toBe(1)
    expect(result.pending).toBe(2)
  })

  test("handles empty array", () => {
    const result = analyzeTodos([])

    expect(result.total).toBe(0)
    expect(result.completed).toBe(0)
    expect(result.inProgress).toBe(0)
    expect(result.pending).toBe(0)
  })

  test("handles null todos", () => {
    const result = analyzeTodos(null as unknown as Todo[])

    expect(result.total).toBe(0)
    expect(result.completed).toBe(0)
    expect(result.inProgress).toBe(0)
    expect(result.pending).toBe(0)
  })

  test("handles undefined todos", () => {
    const result = analyzeTodos(undefined as unknown as Todo[])

    expect(result.total).toBe(0)
    expect(result.completed).toBe(0)
    expect(result.inProgress).toBe(0)
    expect(result.pending).toBe(0)
  })

  test("handles all completed todos", () => {
    const todos: Todo[] = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "completed" },
      { content: "Task 3", status: "completed" },
    ]

    const result = analyzeTodos(todos)

    expect(result.total).toBe(3)
    expect(result.completed).toBe(3)
    expect(result.inProgress).toBe(0)
    expect(result.pending).toBe(0)
  })

  test("handles all in_progress todos", () => {
    const todos: Todo[] = [
      { content: "Task 1", status: "in_progress" },
      { content: "Task 2", status: "in_progress" },
    ]

    const result = analyzeTodos(todos)

    expect(result.total).toBe(2)
    expect(result.completed).toBe(0)
    expect(result.inProgress).toBe(2)
    expect(result.pending).toBe(0)
  })

  test("handles single todo", () => {
    const todos: Todo[] = [{ content: "Task 1", status: "pending" }]

    const result = analyzeTodos(todos)

    expect(result.total).toBe(1)
    expect(result.completed).toBe(0)
    expect(result.inProgress).toBe(0)
    expect(result.pending).toBe(1)
  })
})

// ============================================================================
// inferOutcomeFromTodos Tests
// ============================================================================

describe("inferOutcomeFromTodos", () => {
  test("0 todos returns FAILED", () => {
    const result = inferOutcomeFromTodos(0, 0, 0, 0)

    expect(result.outcome).toBe("FAILED")
    expect(result.reason).toContain("No todos tracked")
  })

  test("100% complete returns SUCCEEDED", () => {
    const result = inferOutcomeFromTodos(5, 5, 0, 0)

    expect(result.outcome).toBe("SUCCEEDED")
    expect(result.reason).toBe("All 5 tasks completed")
  })

  test("single task completed returns SUCCEEDED", () => {
    const result = inferOutcomeFromTodos(1, 1, 0, 0)

    expect(result.outcome).toBe("SUCCEEDED")
    expect(result.reason).toBe("All 1 tasks completed")
  })

  test("80% complete returns PARTIAL_PLUS", () => {
    // 4 out of 5 = 80%
    const result = inferOutcomeFromTodos(5, 4, 1, 0)

    expect(result.outcome).toBe("PARTIAL_PLUS")
    expect(result.reason).toContain("4/5 tasks done")
    expect(result.reason).toContain("1 minor items remain")
  })

  test("90% complete returns PARTIAL_PLUS", () => {
    // 9 out of 10 = 90%
    const result = inferOutcomeFromTodos(10, 9, 0, 1)

    expect(result.outcome).toBe("PARTIAL_PLUS")
    expect(result.reason).toContain("9/10 tasks done")
  })

  test("exact 80% boundary returns PARTIAL_PLUS", () => {
    // 8 out of 10 = exactly 80%
    const result = inferOutcomeFromTodos(10, 8, 1, 1)

    expect(result.outcome).toBe("PARTIAL_PLUS")
    expect(result.reason).toContain("8/10 tasks done")
  })

  test("79% complete returns PARTIAL_MINUS", () => {
    // 79 out of 100 = 79%
    const result = inferOutcomeFromTodos(100, 79, 10, 11)

    expect(result.outcome).toBe("PARTIAL_MINUS")
    expect(result.reason).toContain("79/100 tasks done")
    expect(result.reason).toContain("significant work remains")
  })

  test("50% complete returns PARTIAL_MINUS", () => {
    // 5 out of 10 = 50%
    const result = inferOutcomeFromTodos(10, 5, 2, 3)

    expect(result.outcome).toBe("PARTIAL_MINUS")
    expect(result.reason).toContain("5/10 tasks done")
  })

  test("exact 50% boundary returns PARTIAL_MINUS", () => {
    // 1 out of 2 = exactly 50%
    const result = inferOutcomeFromTodos(2, 1, 0, 1)

    expect(result.outcome).toBe("PARTIAL_MINUS")
    expect(result.reason).toContain("1/2 tasks done")
  })

  test("49% complete returns FAILED", () => {
    // 49 out of 100 = 49%
    const result = inferOutcomeFromTodos(100, 49, 25, 26)

    expect(result.outcome).toBe("FAILED")
    expect(result.reason).toContain("Insufficient progress")
    expect(result.reason).toContain("49/100")
    expect(result.reason).toContain("49%")
  })

  test("0% complete (all pending) returns FAILED", () => {
    const result = inferOutcomeFromTodos(5, 0, 0, 5)

    expect(result.outcome).toBe("FAILED")
    expect(result.reason).toContain("Insufficient progress")
    expect(result.reason).toContain("0/5")
    expect(result.reason).toContain("0%")
  })

  test("0% complete (all in_progress) returns FAILED", () => {
    const result = inferOutcomeFromTodos(3, 0, 3, 0)

    expect(result.outcome).toBe("FAILED")
    expect(result.reason).toContain("Insufficient progress")
  })

  test("25% complete returns FAILED", () => {
    // 1 out of 4 = 25%
    const result = inferOutcomeFromTodos(4, 1, 1, 2)

    expect(result.outcome).toBe("FAILED")
    expect(result.reason).toContain("Insufficient progress")
    expect(result.reason).toContain("25%")
  })
})

// ============================================================================
// State Management Tests (readState, writeState)
// ============================================================================

// TODO(review): Consolidate multiple import statements from "../utils/state" into single import.
// (code-reviewer, 2025-01-11, severity: Low)
// Import state functions directly for testing
import {
  readState,
  writeState,
  getStatePath,
  sanitizeKey,
  sanitizeSessionId,
  getStateDir,
} from "../utils/state"

describe("readState", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-state-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("returns null for non-existent file", () => {
    const result = readState(tempDir, "nonexistent", "session123")

    expect(result).toBeNull()
  })

  test("returns parsed JSON for valid file", () => {
    const testData = { foo: "bar", count: 42, nested: { value: true } }
    const stateDir = join(tempDir, ".opencode", "state")
    mkdirSync(stateDir, { recursive: true })
    writeFileSync(join(stateDir, "mykey-session123.json"), JSON.stringify(testData))

    const result = readState<typeof testData>(tempDir, "mykey", "session123")

    expect(result).toEqual(testData)
  })

  test("handles corrupted JSON gracefully", () => {
    const stateDir = join(tempDir, ".opencode", "state")
    mkdirSync(stateDir, { recursive: true })
    writeFileSync(join(stateDir, "corrupt-session456.json"), "{ invalid json here")

    const result = readState(tempDir, "corrupt", "session456")

    expect(result).toBeNull()
  })

  test("handles empty file gracefully", () => {
    const stateDir = join(tempDir, ".opencode", "state")
    mkdirSync(stateDir, { recursive: true })
    writeFileSync(join(stateDir, "empty-session789.json"), "")

    const result = readState(tempDir, "empty", "session789")

    expect(result).toBeNull()
  })

  test("reads global state (no sessionId)", () => {
    const testData = { global: true }
    const stateDir = join(tempDir, ".opencode", "state")
    mkdirSync(stateDir, { recursive: true })
    writeFileSync(join(stateDir, "globalkey-global.json"), JSON.stringify(testData))

    const result = readState<typeof testData>(tempDir, "globalkey")

    expect(result).toEqual(testData)
  })

  test("reads from legacy .ring/state directory", () => {
    const testData = { legacy: true }
    const legacyStateDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyStateDir, { recursive: true })
    writeFileSync(join(legacyStateDir, "legacykey-session111.json"), JSON.stringify(testData))

    const result = readState<typeof testData>(tempDir, "legacykey", "session111")

    expect(result).toEqual(testData)
  })
})

describe("writeState", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-state-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("creates file with correct content", () => {
    const testData = { message: "hello", number: 123 }

    writeState(tempDir, "testkey", testData, "sessionABC")

    const filePath = join(tempDir, ".opencode", "state", "testkey-sessionABC.json")
    expect(existsSync(filePath)).toBe(true)

    const content = JSON.parse(readFileSync(filePath, "utf-8"))
    expect(content).toEqual(testData)
  })

  test("creates directory if needed", () => {
    const stateDir = join(tempDir, ".opencode", "state")
    expect(existsSync(stateDir)).toBe(false)

    writeState(tempDir, "newkey", { data: "value" }, "newSession")

    expect(existsSync(stateDir)).toBe(true)
    const filePath = join(stateDir, "newkey-newSession.json")
    expect(existsSync(filePath)).toBe(true)
  })

  test("atomic write uses temp file then rename", () => {
    const testData = { atomic: true }

    writeState(tempDir, "atomickey", testData, "atomicSession")

    const finalPath = join(tempDir, ".opencode", "state", "atomickey-atomicSession.json")
    expect(existsSync(finalPath)).toBe(true)

    const content = JSON.parse(readFileSync(finalPath, "utf-8"))
    expect(content).toEqual(testData)

    // Verify no temp files left behind
    const stateDir = join(tempDir, ".opencode", "state")
    const files = readdirSync(stateDir)
    const tempFiles = files.filter((f: string) => f.includes(".tmp."))
    expect(tempFiles.length).toBe(0)
  })

  test("writes global state (no sessionId)", () => {
    const testData = { global: "state" }

    writeState(tempDir, "globalwrite", testData)

    const filePath = join(tempDir, ".opencode", "state", "globalwrite-global.json")
    expect(existsSync(filePath)).toBe(true)

    const content = JSON.parse(readFileSync(filePath, "utf-8"))
    expect(content).toEqual(testData)
  })

  test("overwrites existing file", () => {
    writeState(tempDir, "overwrite", { version: 1 }, "sess")
    writeState(tempDir, "overwrite", { version: 2 }, "sess")

    const filePath = join(tempDir, ".opencode", "state", "overwrite-sess.json")
    const content = JSON.parse(readFileSync(filePath, "utf-8"))
    expect(content).toEqual({ version: 2 })
  })

  test("writes complex nested data", () => {
    const testData = {
      array: [1, 2, 3],
      nested: {
        deep: {
          value: "test",
        },
      },
      date: "2024-01-01T00:00:00Z",
      unicode: "Hello Unicode Test",
    }

    writeState(tempDir, "complex", testData, "sess")

    const result = readState<typeof testData>(tempDir, "complex", "sess")
    expect(result).toEqual(testData)
  })
})

// ============================================================================
// Utility Function Tests
// ============================================================================

// Note: sanitizeSessionId and sanitizeKey are comprehensively tested in
// security-sanitization.test.ts. These tests focus on state path utilities.

describe("getStatePath", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-state-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("constructs correct path with sessionId", () => {
    const path = getStatePath(tempDir, "mykey", "session123")
    expect(path).toBe(join(tempDir, ".opencode/state/mykey-session123.json"))
  })

  test("constructs correct path without sessionId (global)", () => {
    const path = getStatePath(tempDir, "mykey")
    expect(path).toBe(join(tempDir, ".opencode/state/mykey-global.json"))
  })

  test("sanitizes key and sessionId in path", () => {
    const path = getStatePath(tempDir, "key/../bad", "session/../attack")
    expect(path).toBe(join(tempDir, ".opencode/state/keybad-sessionattack.json"))
  })
})

describe("getStateDir", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-state-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("creates state directory if not exists", () => {
    const stateDir = getStateDir(tempDir)

    expect(existsSync(stateDir)).toBe(true)
    expect(stateDir).toBe(join(tempDir, ".opencode/state"))
  })

  test("returns existing state directory", () => {
    const expectedDir = join(tempDir, ".opencode", "state")
    mkdirSync(expectedDir, { recursive: true })

    const stateDir = getStateDir(tempDir)

    expect(stateDir).toBe(expectedDir)
  })
})

// ============================================================================
// Integration Tests
// ============================================================================

describe("Integration: analyzeTodos + inferOutcomeFromTodos", () => {
  test("full workflow: analyze then infer", () => {
    const todos: Todo[] = [
      { content: "Task 1", status: "completed" },
      { content: "Task 2", status: "completed" },
      { content: "Task 3", status: "completed" },
      { content: "Task 4", status: "completed" },
      { content: "Task 5", status: "in_progress" },
    ]

    const analysis = analyzeTodos(todos)
    const outcome = inferOutcomeFromTodos(analysis.total, analysis.completed, analysis.inProgress, analysis.pending)

    expect(outcome.outcome).toBe("PARTIAL_PLUS") // 80%
    expect(outcome.reason).toContain("4/5")
  })

  test("full workflow: empty todos", () => {
    const todos: Todo[] = []

    const analysis = analyzeTodos(todos)
    const outcome = inferOutcomeFromTodos(analysis.total, analysis.completed, analysis.inProgress, analysis.pending)

    expect(outcome.outcome).toBe("FAILED")
    expect(outcome.reason).toContain("No todos tracked")
  })
})

describe("Integration: writeState + readState roundtrip", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-state-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("write then read returns same data", () => {
    const testData = {
      todos: [
        { content: "Task 1", status: "completed" },
        { content: "Task 2", status: "pending" },
      ],
      metadata: {
        createdAt: "2024-01-01T00:00:00Z",
        version: 1,
      },
    }

    writeState(tempDir, "todos-state", testData, "integration-session")
    const result = readState<typeof testData>(tempDir, "todos-state", "integration-session")

    expect(result).toEqual(testData)
  })

  test("multiple keys with same session are isolated", () => {
    writeState(tempDir, "key1", { value: "one" }, "session")
    writeState(tempDir, "key2", { value: "two" }, "session")

    const result1 = readState<{ value: string }>(tempDir, "key1", "session")
    const result2 = readState<{ value: string }>(tempDir, "key2", "session")

    expect(result1?.value).toBe("one")
    expect(result2?.value).toBe("two")
  })

  test("same key with different sessions are isolated", () => {
    writeState(tempDir, "shared-key", { value: "session1" }, "session1")
    writeState(tempDir, "shared-key", { value: "session2" }, "session2")

    const result1 = readState<{ value: string }>(tempDir, "shared-key", "session1")
    const result2 = readState<{ value: string }>(tempDir, "shared-key", "session2")

    expect(result1?.value).toBe("session1")
    expect(result2?.value).toBe("session2")
  })
})

// ============================================================================
// State Migration Tests
// ============================================================================

import { migrateStateFiles, getMigrationStatus, cleanupLegacyState } from "../utils/state"

describe("migrateStateFiles", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-migration-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("migrates files from .ring/state to .opencode/state", () => {
    // Create legacy state file
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })
    writeFileSync(join(legacyDir, "test-key-session123.json"), JSON.stringify({ value: "legacy" }))

    // Run migration
    const result = migrateStateFiles(tempDir)

    // Verify file was copied
    const newPath = join(tempDir, ".opencode", "state", "test-key-session123.json")
    expect(existsSync(newPath)).toBe(true)
    expect(JSON.parse(readFileSync(newPath, "utf-8"))).toEqual({ value: "legacy" })
    expect(result.migrated).toBe(1)
    expect(result.skipped).toBe(0)
  })

  test("skips migration if file already exists in .opencode/state", () => {
    // Create file in both locations
    const legacyDir = join(tempDir, ".ring", "state")
    const newDir = join(tempDir, ".opencode", "state")
    mkdirSync(legacyDir, { recursive: true })
    mkdirSync(newDir, { recursive: true })
    writeFileSync(join(legacyDir, "existing-session.json"), JSON.stringify({ old: true }))
    writeFileSync(join(newDir, "existing-session.json"), JSON.stringify({ new: true }))

    // Run migration
    const result = migrateStateFiles(tempDir)

    // Verify new file was NOT overwritten
    const content = JSON.parse(readFileSync(join(newDir, "existing-session.json"), "utf-8"))
    expect(content).toEqual({ new: true })
    expect(result.migrated).toBe(0)
    expect(result.skipped).toBe(1)
  })

  test("handles missing .ring/state directory gracefully", () => {
    // No .ring/state directory exists
    const result = migrateStateFiles(tempDir)

    expect(result.migrated).toBe(0)
    expect(result.skipped).toBe(0)
    expect(result.error).toBeUndefined()
  })

  test("only migrates .json files", () => {
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })
    writeFileSync(join(legacyDir, "valid-session.json"), JSON.stringify({ valid: true }))
    writeFileSync(join(legacyDir, "invalid.txt"), "not json")
    writeFileSync(join(legacyDir, ".hidden"), "hidden file")

    const result = migrateStateFiles(tempDir)

    expect(result.migrated).toBe(1)
    expect(existsSync(join(tempDir, ".opencode", "state", "valid-session.json"))).toBe(true)
    expect(existsSync(join(tempDir, ".opencode", "state", "invalid.txt"))).toBe(false)
  })

  test("tracks migration status to prevent re-migration", () => {
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })
    writeFileSync(join(legacyDir, "key-session.json"), JSON.stringify({ data: 1 }))

    // First migration
    migrateStateFiles(tempDir)

    // Check status
    const status = getMigrationStatus(tempDir)
    expect(status.migrated).toBe(true)

    // Add new legacy file
    writeFileSync(join(legacyDir, "key2-session.json"), JSON.stringify({ data: 2 }))

    // Second migration should still work for new files
    const result2 = migrateStateFiles(tempDir)
    expect(result2.migrated).toBe(1)
  })

  // TODO(review): Add test verifying subdirectories in legacy state directory are skipped.
  // (code-reviewer, 2025-01-11, severity: Low)

  // TODO(review): Add test for file copy failure scenario (e.g., unreadable source file).
  // (code-reviewer, 2025-01-11, severity: Low)
})

// ============================================================================
// cleanupLegacyState Tests
// ============================================================================

describe("cleanupLegacyState", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-cleanup-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("removes legacy directory after successful migration", () => {
    // Setup: Create legacy and new directories
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })
    writeFileSync(join(legacyDir, "old-session.json"), JSON.stringify({ old: true }))

    // Run migration first
    migrateStateFiles(tempDir)

    // Now cleanup
    const result = cleanupLegacyState(tempDir)

    expect(result.removed).toBe(true)
    expect(existsSync(legacyDir)).toBe(false)
  })

  test("refuses to remove if migration not completed", () => {
    // Create legacy directory without migration
    const legacyDir = join(tempDir, ".ring", "state")
    const newDir = join(tempDir, ".opencode", "state")
    mkdirSync(legacyDir, { recursive: true })
    mkdirSync(newDir, { recursive: true })
    writeFileSync(join(legacyDir, "file.json"), "{}")

    const result = cleanupLegacyState(tempDir)

    expect(result.removed).toBe(false)
    expect(result.reason).toContain("Migration has not been completed")
    expect(existsSync(legacyDir)).toBe(true)
  })

  test("returns gracefully if legacy directory does not exist", () => {
    const result = cleanupLegacyState(tempDir)

    expect(result.removed).toBe(false)
    expect(result.reason).toContain("does not exist")
  })

  test("refuses if new state directory does not exist", () => {
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })

    const result = cleanupLegacyState(tempDir)

    expect(result.removed).toBe(false)
    expect(result.reason).toContain("run migration first")
  })

  // TODO(review): Add test for cleanup refusing when files added after migration.
  // (business-logic-reviewer, 2025-01-11, severity: Low)
})
