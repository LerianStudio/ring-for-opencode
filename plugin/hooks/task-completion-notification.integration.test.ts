import { describe, expect, mock, test } from "bun:test"
import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"

import { hookRegistry } from "./registry.js"
import type { HookContext, HookOutput } from "./types.js"
import { createTaskCompletionHook } from "./factories/task-completion.js"

// Avoid OS notifications in tests
let execCalls: Array<{ cmd: string; args: string[] }> = []
mock.module("node:child_process", () => {
  return {
    execFile: (cmd: string, args: string[], cb: (err: Error | null) => void) => {
      execCalls.push({ cmd, args })
      cb(null)
    },
  }
})

describe("task-completion -> notification (end-to-end)", () => {
  test("emits a single notification when todos transition to all complete", async () => {
    execCalls = []

    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "ring-task-completion-"))

    // Dynamic import so the child_process mock applies
    const { createNotificationHook } = await import("./factories/notification.js")

    hookRegistry.clear()
    hookRegistry.register(createTaskCompletionHook({ persistState: true, showToast: true }))
    hookRegistry.register(
      createNotificationHook({
        enabled: true,
        // ensure taskComplete notifications are enabled
        notifyOn: { taskComplete: true },
      }),
    )

    const baseCtx: Omit<HookContext, "event" | "lifecycle"> = {
      sessionId: "s1",
      directory: tmpDir,
    }

    const output: HookOutput = {}

    // 1) pending todos -> no notification
    await hookRegistry.executeLifecycle(
      "todo.updated",
      {
        ...baseCtx,
        lifecycle: "todo.updated",
        event: {
          type: "todo.updated",
          properties: {
            todos: [{ content: "t1", status: "pending" }],
          },
        },
      },
      output,
    )

    expect(execCalls.length).toBe(0)

    // 2) all complete -> exactly one notification
    await hookRegistry.executeLifecycle(
      "todo.updated",
      {
        ...baseCtx,
        lifecycle: "todo.updated",
        event: {
          type: "todo.updated",
          properties: {
            todos: [{ content: "t1", status: "completed" }],
          },
        },
      },
      output,
    )

    expect(execCalls.length).toBe(1)
  })
})
