import { describe, expect, test } from "bun:test"
import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"

import { createNotificationHook } from "../../../plugin/hooks/factories/notification.js"
import { createTaskCompletionHook } from "../../../plugin/hooks/factories/task-completion.js"
import { hookRegistry } from "../../../plugin/hooks/registry.js"
import type { HookContext, HookOutput } from "../../../plugin/hooks/types.js"

let dispatchCalls = 0

const testDispatch = async (): Promise<boolean> => {
  dispatchCalls++
  return true
}

describe("task-completion -> notification (end-to-end)", () => {
  test("emits a single notification when todos transition to all complete", async () => {
    dispatchCalls = 0

    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "ring-task-completion-"))

    hookRegistry.clear()
    hookRegistry.register(createTaskCompletionHook({ persistState: true, showToast: true }))
    hookRegistry.register(
      createNotificationHook({
        enabled: true,
        // ensure taskComplete notifications are enabled
        notifyOn: { taskComplete: true },
        dispatch: testDispatch,
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

    expect(dispatchCalls).toBe(0)

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

    expect(dispatchCalls).toBe(1)
  })
})
