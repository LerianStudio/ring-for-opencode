import { describe, expect, test } from "bun:test"
import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"

import { deepMerge } from "../../plugin/config/loader.js"
import { createContextInjectionHook } from "../../plugin/hooks/factories/context-injection.js"
import {
  buildNotifySendArgs,
  buildOsaScriptArgs,
} from "../../plugin/hooks/factories/notification.js"
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

      const commands = loadRingCommands(pluginRoot, projectRoot)
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

  test("context-injection ledger discovery rejects symlinks that resolve outside project root", async () => {
    const root = mkdtemp("ring-sec-ledger-")

    try {
      const ledgersDir = path.join(root, ".ring", "ledgers")
      fs.mkdirSync(ledgersDir, { recursive: true })

      // valid ledger
      const ledger1 = path.join(ledgersDir, "001.md")
      writeFile(ledger1, "## Current Task\nDo the thing\n")

      // symlink ledger to outside root
      const outsideDir = mkdtemp("ring-sec-ledger-outside-")
      const outsideLedger = path.join(outsideDir, "outside.md")
      writeFile(outsideLedger, "## Current Task\nMALICIOUS\n")

      const symlinkPath = path.join(ledgersDir, "999.md")
      fs.symlinkSync(outsideLedger, symlinkPath)

      // make symlink look most-recent
      const now = new Date()
      fs.utimesSync(symlinkPath, now, now)

      const hook = createContextInjectionHook({
        injectCompactRules: false,
        injectSkillsRef: false,
        injectCommandsRef: false,
        injectAgentsRef: false,
        injectLedgerSummary: true,
        maxLedgerSummaryLength: 2000,
      })

      const output: { context?: string[] } = {}
      const res = await hook.execute(
        {
          sessionId: "s",
          directory: root,
          lifecycle: "session.compacting",
        },
        output,
      )

      expect(res.success).toBe(true)

      const joined = (output.context ?? []).join("\n")
      expect(joined).toContain("Do the thing")
      expect(joined).not.toContain("MALICIOUS")

      fs.rmSync(outsideDir, { recursive: true, force: true })
    } finally {
      fs.rmSync(root, { recursive: true, force: true })
    }
  })

  test("notifications are injection-safe (osascript argv, notify-send --)", () => {
    const title = "--replace=all"
    const message = '"; do shell script "rm -rf /"; "'

    const notifySendArgs = buildNotifySendArgs(title, message)
    expect(notifySendArgs).toEqual(["--", title, message])

    const osaArgs = buildOsaScriptArgs(title, message, true)
    expect(osaArgs[0]).toBe("-e")
    expect(osaArgs).toContain("--")

    // user-provided content must NOT be interpolated into the script itself
    const script = osaArgs[1]
    expect(script).not.toContain(title)
    expect(script).not.toContain(message)
  })
})
