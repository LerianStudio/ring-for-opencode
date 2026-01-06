#!/usr/bin/env bun
/**
 * Ring OpenCode Plugin Test Script
 *
 * Validates that all plugins:
 * 1. Export correctly
 * 2. Have proper type signatures
 * 3. Return valid hook objects
 *
 * Run with: bun .opencode/plugin/test-plugins.ts
 */

import * as plugins from "./index"

interface TestResult {
  name: string
  passed: boolean
  error?: string
}

const results: TestResult[] = []

// Mock context for plugin initialization
// SDK returns array directly, not { data: [] }
const mockContext = {
  project: { name: "test-project", path: "/tmp/test-project" },
  client: {
    session: {
      // Returns array with realistic session structure
      list: async () => [{ id: "test-session-1", updatedAt: Date.now(), name: "Test Session" }],
      prompt: async () => ({ success: true }),
    },
  },
  // Mock Bun shell ($) with realistic structure
  $: Object.assign(
    async (strings: TemplateStringsArray, ...values: unknown[]) => ({
      text: async () => "",
      quiet: () => Promise.resolve({ text: async () => "" }),
      exitCode: 0,
    }),
    { quiet: () => Promise.resolve({ text: async () => "" }) }
  ),
  directory: "/tmp/test-project",
  worktree: "/tmp/test-project",
}

async function testPlugin(name: string, plugin: any): Promise<TestResult> {
  try {
    // Check it's a function
    if (typeof plugin !== "function") {
      return { name, passed: false, error: "Not a function" }
    }

    // Initialize plugin
    const hooks = await plugin(mockContext)

    // Check it returns an object
    if (typeof hooks !== "object") {
      return { name, passed: false, error: "Does not return hooks object" }
    }

    // Check for valid hooks
    const validHooks = [
      "event",
      "tool.execute.before",
      "tool.execute.after",
      "experimental.session.compacting",
      "session.compacted",
    ]

    const hookKeys = Object.keys(hooks)
    for (const key of hookKeys) {
      if (!validHooks.includes(key) && key !== "tool") {
        return { name, passed: false, error: `Invalid hook: ${key}` }
      }
    }

    return { name, passed: true }
  } catch (err) {
    return { name, passed: false, error: String(err) }
  }
}

async function runTests() {
  console.log("Testing Ring OpenCode Plugins\n")
  console.log("=".repeat(50))

  const pluginEntries = Object.entries(plugins).filter(([name]) => name !== "default" && name.startsWith("Ring"))

  for (const [name, plugin] of pluginEntries) {
    const result = await testPlugin(name, plugin)
    results.push(result)

    const status = result.passed ? "PASS" : "FAIL"
    console.log(`${status}: ${name}`)
    if (result.error) {
      console.log(`       Error: ${result.error}`)
    }
  }

  console.log("\n" + "=".repeat(50))

  const passed = results.filter((r) => r.passed).length
  const failed = results.filter((r) => !r.passed).length

  console.log(`\nResults: ${passed} passed, ${failed} failed`)

  if (failed > 0) {
    process.exit(1)
  }
}

// Only run when executed directly (not when imported by OpenCode)
if (import.meta.main) {
  runTests().catch(console.error)
}
