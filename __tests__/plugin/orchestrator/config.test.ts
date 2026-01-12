import { describe, expect, it } from "bun:test"
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs"
import { tmpdir } from "node:os"
import { join } from "node:path"

import type { OrchestratorConfigFile } from "../../../plugin/orchestrator/config.js"
import {
  isInsideWorker,
  loadOrchestratorConfig,
  validateServerUrl,
} from "../../../plugin/orchestrator/config.js"

const withTempConfigDir = <T>(fn: (configDir: string) => T): T => {
  const original = process.env.XDG_CONFIG_HOME
  const configDir = mkdtempSync(join(tmpdir(), "ring-opencode-test-"))
  process.env.XDG_CONFIG_HOME = configDir

  try {
    return fn(configDir)
  } finally {
    if (original === undefined) {
      delete process.env.XDG_CONFIG_HOME
    } else {
      process.env.XDG_CONFIG_HOME = original
    }
    rmSync(configDir, { recursive: true, force: true })
  }
}

describe("loadOrchestratorConfig", () => {
  it("uses defaults when no config files exist", () => {
    const loaded = withTempConfigDir(() => loadOrchestratorConfig("/nonexistent/path"))

    expect(loaded.config.basePort).toBe(14096)
    expect(loaded.config.autoSpawn).toBe(true)
    expect(loaded.profiles).toHaveProperty("vision")
    expect(loaded.profiles).toHaveProperty("coder")
    expect(loaded.profiles).toHaveProperty("docs")
  })

  it("includes all built-in profiles", () => {
    const loaded = withTempConfigDir(() => loadOrchestratorConfig("/tmp"))

    expect(Object.keys(loaded.profiles)).toContain("vision")
    expect(Object.keys(loaded.profiles)).toContain("docs")
    expect(Object.keys(loaded.profiles)).toContain("coder")
    expect(Object.keys(loaded.profiles)).toContain("architect")
    expect(Object.keys(loaded.profiles)).toContain("explorer")
  })

  it("prefers project config over global config", () => {
    const projectDir = mkdtempSync(join(tmpdir(), "ring-project-"))

    try {
      const loaded = withTempConfigDir((configDir) => {
        const globalDir = join(configDir, "opencode", "ring")
        mkdirSync(globalDir, { recursive: true })
        writeFileSync(
          join(globalDir, "orchestrator.json"),
          JSON.stringify({
            basePort: 15000,
            profiles: [
              {
                id: "custom",
                name: "Global Custom",
                model: "node",
                purpose: "Global",
                whenToUse: "Global",
              },
            ],
            workers: ["custom"],
          }),
        )

        const projectConfigDir = join(projectDir, ".ring")
        mkdirSync(projectConfigDir, { recursive: true })
        writeFileSync(
          join(projectConfigDir, "orchestrator.json"),
          JSON.stringify({
            basePort: 16000,
            profiles: [
              {
                id: "custom",
                name: "Project Custom",
                model: "node",
                purpose: "Project",
                whenToUse: "Project",
              },
            ],
            workers: ["custom"],
          }),
        )

        return loadOrchestratorConfig(projectDir)
      })

      expect(loaded.config.basePort).toBe(16000)
      expect(loaded.profiles.custom.name).toBe("Project Custom")
      expect(loaded.spawn).toEqual(["custom"])
    } finally {
      rmSync(projectDir, { recursive: true, force: true })
    }
  })
})

describe("isInsideWorker", () => {
  it("returns false when env var not set", () => {
    const config = {
      security: {
        preventRecursiveSpawn: true,
        recursiveSpawnEnvVar: "RING_ORCHESTRATOR_WORKER",
      },
    } as OrchestratorConfigFile

    const original = process.env.RING_ORCHESTRATOR_WORKER
    delete process.env.RING_ORCHESTRATOR_WORKER

    expect(isInsideWorker(config)).toBe(false)

    if (original) {
      process.env.RING_ORCHESTRATOR_WORKER = original
    }
  })
})

describe("validateServerUrl", () => {
  it("allows valid HTTPS URLs", () => {
    expect(() => validateServerUrl("https://api.example.com")).not.toThrow()
  })

  it("blocks AWS metadata endpoint", () => {
    expect(() => validateServerUrl("http://169.254.169.254/latest/meta-data/")).toThrow(
      /Blocked host/,
    )
  })

  it("blocks localhost", () => {
    expect(() => validateServerUrl("http://localhost:8080")).toThrow(/Blocked host/)
  })

  it("blocks private IPs", () => {
    expect(() => validateServerUrl("http://192.168.1.1:8080")).toThrow(/Private IP/)
    expect(() => validateServerUrl("http://10.0.0.1:8080")).toThrow(/Private IP/)
  })

  it("throws for invalid URLs", () => {
    expect(() => validateServerUrl("not-a-url")).toThrow(/Invalid server URL/)
  })
})
