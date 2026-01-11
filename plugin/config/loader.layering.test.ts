import { describe, expect, test } from "bun:test"
import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"

import { clearConfigCache, getConfigLayers, loadConfig } from "./loader.js"

describe("config/loader", () => {
  test("ignores config layers that parse to non-objects", () => {
    const tmpRoot = fs.mkdtempSync(path.join(os.tmpdir(), "ring-config-nonobject-"))

    const prevXdg = process.env.XDG_CONFIG_HOME
    process.env.XDG_CONFIG_HOME = path.join(tmpRoot, "xdg")

    try {
      // Valid project layer
      fs.mkdirSync(path.join(tmpRoot, ".opencode"), { recursive: true })
      fs.writeFileSync(
        path.join(tmpRoot, ".opencode", "ring.jsonc"),
        JSON.stringify({ disabled_hooks: ["session-start"] }),
      )

      // Invalid local layer (non-object) should be ignored.
      fs.mkdirSync(path.join(tmpRoot, ".ring"), { recursive: true })
      fs.writeFileSync(path.join(tmpRoot, ".ring", "local.jsonc"), '"oops"')

      clearConfigCache()
      const config = loadConfig(tmpRoot, true)

      expect(config.disabled_hooks).toContain("session-start")
    } finally {
      process.env.XDG_CONFIG_HOME = prevXdg
      fs.rmSync(tmpRoot, { recursive: true, force: true })
    }
  })

  test("resolves user config from XDG_CONFIG_HOME as config.json when config.jsonc is absent", () => {
    const tmpRoot = fs.mkdtempSync(path.join(os.tmpdir(), "ring-config-xdg-"))

    const prevXdg = process.env.XDG_CONFIG_HOME
    process.env.XDG_CONFIG_HOME = path.join(tmpRoot, "xdg")

    try {
      const userDir = path.join(tmpRoot, "xdg", "opencode", "ring")
      fs.mkdirSync(userDir, { recursive: true })
      fs.writeFileSync(
        path.join(userDir, "config.json"),
        JSON.stringify({ disabled_agents: ["code-reviewer"] }),
      )

      clearConfigCache()
      const config = loadConfig(tmpRoot, true)

      expect(config.disabled_agents).toContain("code-reviewer")

      const userLayer = getConfigLayers().find((l) => l.name === "user")
      expect(userLayer?.exists).toBe(true)
      expect(userLayer?.path?.endsWith(path.join("ring", "config.json"))).toBe(true)
    } finally {
      process.env.XDG_CONFIG_HOME = prevXdg
      fs.rmSync(tmpRoot, { recursive: true, force: true })
    }
  })
})
