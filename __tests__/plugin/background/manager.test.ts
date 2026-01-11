import { describe, expect, test } from "bun:test"

import { BackgroundManager } from "../../../plugin/background/manager.js"
import type { BackgroundClient, BackgroundTask } from "../../../plugin/background/types.js"

describe("BackgroundManager - pollTasks completion detection", () => {
  test("does not complete a task when todoResult.data is undefined", async () => {
    const client: BackgroundClient = {
      session: {
        create: async () => ({ data: { id: "unused" } }),
        prompt: async () => {},
        messages: async () => ({ data: [] }),
        status: async () => ({ data: { s_task: { type: "idle" } } }),
        todo: async () => ({ data: undefined }),
      },
      tui: {
        showToast: async () => {},
      },
    }

    const manager = new BackgroundManager(client, ".", {
      defaultConcurrency: 1,
      taskTimeoutMs: 9999999,
    })

    const task: BackgroundTask = {
      id: "t1",
      sessionId: "s_task",
      parentSessionId: "s_parent",
      description: "desc",
      prompt: "prompt",
      agent: "agent",
      status: "running",
      createdAt: new Date(),
      startedAt: new Date(),
    }

    const internal = manager as unknown as {
      tasks: Map<string, BackgroundTask>
      sessionToTask: Map<string, string>
    }

    internal.tasks.set(task.id, task)
    internal.sessionToTask.set(task.sessionId, task.id)

    await (manager as unknown as { pollTasks: () => Promise<void> }).pollTasks()

    expect(task.status).toBe("running")
  })
})
