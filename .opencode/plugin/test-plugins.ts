#!/usr/bin/env bun
/**
 * Ring OpenCode Plugin Smoke Tests
 *
 * Simple validation that all plugins export correctly and return hooks.
 * For full behavioral tests, run: bun test .opencode/plugin/test-plugins.test.ts
 *
 * Run with: bun .opencode/plugin/test-plugins.ts
 */

import { mkdirSync, rmSync, existsSync } from "fs"

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
const CLI_TEST_DIR = "/tmp/ring-plugin-cli-tests"

// Create simple mock context
function createMockContext() {
  const noop = () => Promise.resolve({})

  return {
    project: { name: "test-project", path: CLI_TEST_DIR },
    client: {
      session: {
        list: () =>
          Promise.resolve({
            data: [{ id: "test-session-1", updatedAt: Date.now(), name: "Test Session" }],
            error: null,
          }),
        prompt: noop,
      },
      tui: {
        showToast: noop,
      },
    },
    $: Object.assign(
      async () => ({
        text: async () => "",
        quiet: () => Promise.resolve({ text: async () => "" }),
        exitCode: 0,
      }),
      { quiet: () => Promise.resolve({ text: async () => "" }) }
    ),
    directory: CLI_TEST_DIR,
    worktree: CLI_TEST_DIR,
  }
}

interface TestResult {
  name: string
  passed: boolean
  error?: string
}

async function runTests() {
  console.log("Running Ring OpenCode Plugin Smoke Tests\n")
  console.log("For full behavioral tests, run: bun test .opencode/plugin/test-plugins.test.ts\n")
  console.log("=".repeat(60))

  // Setup test directory
  if (existsSync(CLI_TEST_DIR)) {
    rmSync(CLI_TEST_DIR, { recursive: true, force: true })
  }
  mkdirSync(CLI_TEST_DIR, { recursive: true })

  const ctx = createMockContext()
  const results: TestResult[] = []

  const plugins = [
    { name: "RingContextInjection", plugin: RingContextInjection },
    { name: "RingNotification", plugin: RingNotification },
    { name: "RingDoubtResolver", plugin: RingDoubtResolver },
    { name: "RingSessionStart", plugin: RingSessionStart },
    { name: "RingTaskCompletionCheck", plugin: RingTaskCompletionCheck },
    { name: "RingSessionOutcome", plugin: RingSessionOutcome },
    { name: "RingOutcomeInference", plugin: RingOutcomeInference },
  ]

  for (const { name, plugin } of plugins) {
    try {
      // Check it's a function
      if (typeof plugin !== "function") {
        results.push({ name, passed: false, error: "Not a function" })
        continue
      }

      // Initialize plugin
      const hooks = await plugin(ctx as any)

      // Check it returns an object
      if (typeof hooks !== "object") {
        results.push({ name, passed: false, error: "Does not return hooks object" })
        continue
      }

      // Check for valid hooks
      const validHooks = [
        "event",
        "tool",
        "tool.execute.before",
        "tool.execute.after",
        "experimental.session.compacting",
        "experimental.chat.system.transform",
        "experimental.chat.context.transform",
        "session.compacted",
      ]

      const hookKeys = Object.keys(hooks)
      let hasInvalidHook = false
      for (const key of hookKeys) {
        if (!validHooks.includes(key)) {
          results.push({ name, passed: false, error: `Invalid hook: ${key}` })
          hasInvalidHook = true
          break
        }
      }

      if (!hasInvalidHook) {
        results.push({ name, passed: true })
      }
    } catch (err) {
      results.push({ name, passed: false, error: String(err) })
    }
  }

  // Print results
  for (const result of results) {
    const status = result.passed ? "PASS" : "FAIL"
    console.log(`${status}: ${result.name}`)
    if (result.error) {
      console.log(`       Error: ${result.error}`)
    }
  }

  // Cleanup
  rmSync(CLI_TEST_DIR, { recursive: true, force: true })

  console.log("\n" + "=".repeat(60))

  const passed = results.filter((r) => r.passed).length
  const failed = results.filter((r) => !r.passed).length

  console.log(`\nResults: ${passed} passed, ${failed} failed`)

  if (failed > 0) {
    process.exit(1)
  }
}

// Run when executed directly
if (import.meta.main) {
  runTests().catch(console.error)
}
