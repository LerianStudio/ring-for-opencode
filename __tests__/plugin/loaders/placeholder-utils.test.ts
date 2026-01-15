/**
 * Tests for placeholder-utils.ts
 *
 * Tests the placeholder expansion functionality including:
 * - Environment variable priority (OPENCODE_CONFIG_DIR > XDG_CONFIG_HOME > default)
 * - Input validation
 * - Placeholder expansion patterns
 */

import { afterEach, beforeEach, describe, expect, test } from "bun:test"
import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"
import { loadRingAgents } from "../../../plugin/loaders/agent-loader.js"
import {
  expandPlaceholders,
  getOpenCodeConfigDir,
  OPENCODE_CONFIG_PLACEHOLDER,
} from "../../../plugin/loaders/placeholder-utils.js"

describe("placeholder-utils", () => {
  // Store original env to restore after each test
  let originalEnv: NodeJS.ProcessEnv

  beforeEach(() => {
    originalEnv = { ...process.env }
  })

  afterEach(() => {
    // Restore original environment
    process.env = { ...originalEnv }
  })

  describe("OPENCODE_CONFIG_PLACEHOLDER constant", () => {
    test("has correct value", () => {
      expect(OPENCODE_CONFIG_PLACEHOLDER).toBe("{OPENCODE_CONFIG}")
    })
  })

  describe("getOpenCodeConfigDir", () => {
    test("uses OPENCODE_CONFIG_DIR when set", () => {
      process.env.OPENCODE_CONFIG_DIR = "/custom/config"
      delete process.env.XDG_CONFIG_HOME

      const result = getOpenCodeConfigDir()
      expect(result).toBe("/custom/config")
    })

    test("falls back to XDG_CONFIG_HOME/opencode when OPENCODE_CONFIG_DIR unset", () => {
      delete process.env.OPENCODE_CONFIG_DIR
      process.env.XDG_CONFIG_HOME = "/xdg/home"

      const result = getOpenCodeConfigDir()
      expect(result).toBe("/xdg/home/opencode")
    })

    test("ignores XDG_CONFIG_HOME if not absolute path", () => {
      delete process.env.OPENCODE_CONFIG_DIR
      process.env.XDG_CONFIG_HOME = "relative/path"

      const result = getOpenCodeConfigDir()
      // Should fall back to default, not use relative XDG path
      expect(result).toContain(".config/opencode")
      expect(result).not.toContain("relative/path")
    })

    test("defaults to ~/.config/opencode when both env vars unset", () => {
      delete process.env.OPENCODE_CONFIG_DIR
      delete process.env.XDG_CONFIG_HOME

      const result = getOpenCodeConfigDir()
      // Should contain home directory and .config/opencode
      expect(result).toContain(".config")
      expect(result).toContain("opencode")
      expect(result).toBe(path.join(os.homedir(), ".config", "opencode"))
    })

    test("OPENCODE_CONFIG_DIR takes priority over XDG_CONFIG_HOME", () => {
      process.env.OPENCODE_CONFIG_DIR = "/explicit/override"
      process.env.XDG_CONFIG_HOME = "/xdg/should/not/use"

      const result = getOpenCodeConfigDir()
      expect(result).toBe("/explicit/override")
    })
  })

  describe("expandPlaceholders", () => {
    test("expands single placeholder", () => {
      process.env.OPENCODE_CONFIG_DIR = "/test"
      const result = expandPlaceholders("Read {OPENCODE_CONFIG}/standards/go.md")
      expect(result).toBe("Read /test/standards/go.md")
    })

    test("expands multiple placeholders", () => {
      process.env.OPENCODE_CONFIG_DIR = "/test"
      const result = expandPlaceholders("{OPENCODE_CONFIG}/a and {OPENCODE_CONFIG}/b")
      expect(result).toBe("/test/a and /test/b")
    })

    test("handles empty string", () => {
      const result = expandPlaceholders("")
      expect(result).toBe("")
    })

    test("passes through string without placeholders", () => {
      const result = expandPlaceholders("No placeholders here")
      expect(result).toBe("No placeholders here")
    })

    test("handles placeholder at string start", () => {
      process.env.OPENCODE_CONFIG_DIR = "/test"
      const result = expandPlaceholders("{OPENCODE_CONFIG}/end")
      expect(result).toBe("/test/end")
    })

    test("handles placeholder at string end", () => {
      process.env.OPENCODE_CONFIG_DIR = "/test"
      const result = expandPlaceholders("start/{OPENCODE_CONFIG}")
      expect(result).toBe("start//test")
    })

    test("handles placeholder as entire string", () => {
      process.env.OPENCODE_CONFIG_DIR = "/test"
      const result = expandPlaceholders("{OPENCODE_CONFIG}")
      expect(result).toBe("/test")
    })

    test("handles null input gracefully", () => {
      // @ts-expect-error - Testing runtime behavior with invalid input
      const result = expandPlaceholders(null)
      expect(result).toBe("")
    })

    test("handles undefined input gracefully", () => {
      // @ts-expect-error - Testing runtime behavior with invalid input
      const result = expandPlaceholders(undefined)
      expect(result).toBe("")
    })

    test("handles non-string input gracefully", () => {
      // @ts-expect-error - Testing runtime behavior with invalid input
      const result = expandPlaceholders(123)
      expect(result).toBe("")
    })

    test("expands with XDG_CONFIG_HOME fallback", () => {
      delete process.env.OPENCODE_CONFIG_DIR
      process.env.XDG_CONFIG_HOME = "/xdg/config"

      const result = expandPlaceholders("Path: {OPENCODE_CONFIG}/file.md")
      expect(result).toBe("Path: /xdg/config/opencode/file.md")
    })

    test("expands with default fallback", () => {
      delete process.env.OPENCODE_CONFIG_DIR
      delete process.env.XDG_CONFIG_HOME

      const result = expandPlaceholders("{OPENCODE_CONFIG}/test")
      const expected = path.join(os.homedir(), ".config", "opencode", "test")
      expect(result).toBe(expected)
    })
  })

  describe("integration: agent-loader placeholder expansion", () => {
    // Helper to create temporary directories
    function mkdtemp(prefix: string): string {
      return fs.mkdtempSync(path.join(os.tmpdir(), prefix))
    }

    function writeFile(filePath: string, content: string): void {
      fs.mkdirSync(path.dirname(filePath), { recursive: true })
      fs.writeFileSync(filePath, content, "utf-8")
    }

    test("agent loader expands placeholders in prompts", () => {
      const tmp = mkdtemp("ring-placeholder-int-")

      try {
        const pluginRoot = path.join(tmp, "plugin")
        const projectRoot = path.join(tmp, "project")

        // Create an agent with placeholder in prompt
        const agentContent = `---
description: Test agent with placeholder
mode: subagent
---
Read the standards from {OPENCODE_CONFIG}/standards/golang.md

Then apply them.
`
        writeFile(path.join(pluginRoot, "assets", "agent", "test-agent.md"), agentContent)

        // Set a known config dir for testing
        process.env.OPENCODE_CONFIG_DIR = "/test/config/dir"

        const agents = loadRingAgents(pluginRoot, projectRoot)

        // Verify agent was loaded and placeholder was expanded
        expect(agents["ring:test-agent"]).toBeDefined()
        expect(agents["ring:test-agent"].prompt).toContain("/test/config/dir/standards/golang.md")
        expect(agents["ring:test-agent"].prompt).not.toContain("{OPENCODE_CONFIG}")
      } finally {
        fs.rmSync(tmp, { recursive: true, force: true })
      }
    })

    test("agent loader expands multiple placeholders in prompt", () => {
      const tmp = mkdtemp("ring-placeholder-multi-")

      try {
        const pluginRoot = path.join(tmp, "plugin")
        const projectRoot = path.join(tmp, "project")

        // Create an agent with multiple placeholders
        const agentContent = `---
description: Test agent with multiple placeholders
---
First: {OPENCODE_CONFIG}/file1.md
Second: {OPENCODE_CONFIG}/file2.md
`
        writeFile(path.join(pluginRoot, "assets", "agent", "multi-placeholder.md"), agentContent)

        process.env.OPENCODE_CONFIG_DIR = "/multi/test"

        const agents = loadRingAgents(pluginRoot, projectRoot)

        expect(agents["ring:multi-placeholder"]).toBeDefined()
        expect(agents["ring:multi-placeholder"].prompt).toBe(
          "First: /multi/test/file1.md\nSecond: /multi/test/file2.md",
        )
      } finally {
        fs.rmSync(tmp, { recursive: true, force: true })
      }
    })

    test("agent loader handles prompts without placeholders", () => {
      const tmp = mkdtemp("ring-placeholder-none-")

      try {
        const pluginRoot = path.join(tmp, "plugin")
        const projectRoot = path.join(tmp, "project")

        // Create an agent without placeholders
        const agentContent = `---
description: Agent without placeholders
---
This prompt has no placeholders at all.
`
        writeFile(path.join(pluginRoot, "assets", "agent", "no-placeholder.md"), agentContent)

        const agents = loadRingAgents(pluginRoot, projectRoot)

        expect(agents["ring:no-placeholder"]).toBeDefined()
        expect(agents["ring:no-placeholder"].prompt).toBe("This prompt has no placeholders at all.")
      } finally {
        fs.rmSync(tmp, { recursive: true, force: true })
      }
    })
  })
})
