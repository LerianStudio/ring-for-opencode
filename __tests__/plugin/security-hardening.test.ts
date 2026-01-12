import { describe, expect, test } from "bun:test"
import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"

import { deepMerge } from "../../plugin/config/loader.js"
import { loadRingAgents } from "../../plugin/loaders/agent-loader.js"
import { loadRingCommands } from "../../plugin/loaders/command-loader.js"
import { loadRingSkills } from "../../plugin/loaders/skill-loader.js"

function mkdtemp(prefix: string): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), prefix))
}

function writeFile(filePath: string, content: string): void {
  fs.mkdirSync(path.dirname(filePath), { recursive: true })
  fs.writeFileSync(filePath, content, "utf-8")
}

describe("security hardening", () => {
  test("deepMerge blocks prototype pollution (__proto__/constructor/prototype)", () => {
    const protoBefore = ({} as Record<string, unknown>).polluted

    // JSON.parse ensures __proto__ is treated as a plain enumerable key.
    const source = JSON.parse(
      '{"__proto__":{"polluted":true},"constructor":{"prototype":{"polluted":true}},"prototype":{"polluted":true},"safe":{"__proto__":{"pollutedNested":true},"b":2}}',
    ) as Record<string, unknown>

    try {
      const merged = deepMerge({ safe: { a: 1 } }, source as unknown as Partial<{ safe: unknown }>)

      expect(({} as Record<string, unknown>).polluted).toBeUndefined()
      expect(({} as Record<string, unknown>).pollutedNested).toBeUndefined()
      expect(protoBefore).toBeUndefined()
      expect(merged).toMatchObject({ safe: { a: 1, b: 2 } })
    } finally {
      // In case of regression, clean global prototype to avoid cascading failures.
      delete (Object.prototype as unknown as Record<string, unknown>).polluted
      delete (Object.prototype as unknown as Record<string, unknown>).pollutedNested
    }
  })

  test("loaders building maps from filenames use null-prototype objects and skip forbidden keys", () => {
    const tmp = mkdtemp("ring-sec-loaders-")

    try {
      const pluginRoot = path.join(tmp, "plugin")
      const projectRoot = path.join(tmp, "project")

      // commands
      writeFile(path.join(pluginRoot, "assets", "command", "ok.md"), "# ok")
      writeFile(path.join(pluginRoot, "assets", "command", "__proto__.md"), "# bad")

      const { commands } = loadRingCommands(pluginRoot, projectRoot)
      expect(Object.getPrototypeOf(commands)).toBeNull()
      expect(Object.keys(commands)).toEqual(["ring:ok"])

      // agents
      writeFile(
        path.join(pluginRoot, "assets", "agent", "good.md"),
        "---\ndescription: good\n---\nhello",
      )
      writeFile(path.join(pluginRoot, "assets", "agent", "constructor.md"), "---\n---\nnope")

      const agents = loadRingAgents(pluginRoot, projectRoot)
      expect(Object.getPrototypeOf(agents)).toBeNull()
      expect(Object.keys(agents)).toEqual(["ring:good"])

      // skills
      writeFile(path.join(pluginRoot, "assets", "skill", "skill-1", "SKILL.md"), "# skill")
      writeFile(
        path.join(pluginRoot, "assets", "skill", "__proto__", "SKILL.md"),
        "# should be skipped",
      )

      const skills = loadRingSkills(pluginRoot, projectRoot)
      expect(Object.getPrototypeOf(skills)).toBeNull()
      expect(Object.keys(skills)).toEqual(["ring:skill-1"])

      // ensure no global pollution from attempting to set __proto__ keys
      expect(({} as Record<string, unknown>).ok).toBeUndefined()
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true })
    }
  })
})
