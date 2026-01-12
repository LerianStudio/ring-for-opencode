import { beforeEach, describe, expect, it } from "bun:test"

import { jobRegistry } from "../../../plugin/orchestrator/jobs.js"
import { createTaskTools } from "../../../plugin/orchestrator/tools/task-tools.js"
import type {
  OrchestratorContext,
  TaskDispatchInput,
  TaskDispatchResult,
  WorkerInstance,
  WorkerProfile,
  WorkerSpawnOptions,
} from "../../../plugin/orchestrator/types.js"
import { workerPool } from "../../../plugin/orchestrator/worker-pool.js"

const baseContext: OrchestratorContext = {
  directory: process.cwd(),
  profiles: {},
  spawnDefaults: { basePort: 14096, timeout: 1000 },
  defaultListFormat: "json",
}

const resetRegistry = () => {
  const state = jobRegistry as unknown as {
    jobs: Map<string, unknown>
    waiters: Map<string, Set<unknown>>
  }
  state.jobs.clear()
  state.waiters.clear()
}

const makeToolContext = (sessionID: string) => ({
  sessionID,
  agent: "agent",
  messageID: "test-message",
  abort: new AbortController().signal,
})

const spawnWorker = async (
  profile: WorkerProfile,
  options: WorkerSpawnOptions,
): Promise<WorkerInstance> => ({
  profile,
  status: "ready",
  port: options.basePort,
  startedAt: new Date(),
})

const dispatchTask = async (input: TaskDispatchInput): Promise<TaskDispatchResult> => ({
  responseText: `done:${input.task}`,
})

describe("task tools", () => {
  beforeEach(async () => {
    resetRegistry()
    await workerPool.stopAll()
  })

  it("returns error when workerId missing for kind=worker", async () => {
    const tools = createTaskTools(baseContext)
    const result = await tools.taskStart.execute(
      { kind: "worker", task: "Test" },
      makeToolContext("session-a"),
    )

    const payload = JSON.parse(result as string)
    expect(payload.error).toContain("Missing workerId")
  })

  it("rejects invalid attachment paths", async () => {
    const tools = createTaskTools(baseContext)
    const result = await tools.taskStart.execute(
      {
        task: "Test",
        attachments: [{ type: "file", path: "../secret.txt" }],
      },
      makeToolContext("session-a"),
    )

    const payload = JSON.parse(result as string)
    expect(payload.error).toContain("Invalid attachment path")
  })

  it("fails for unknown worker profiles", async () => {
    const tools = createTaskTools({
      ...baseContext,
      dispatchTask,
      spawnWorker,
    })

    const result = await tools.taskStart.execute(
      { kind: "worker", workerId: "unknown", task: "Test" },
      makeToolContext("session-a"),
    )

    const payload = JSON.parse(result as string)
    expect(payload.status).toBe("failed")
    expect(payload.error).toContain("Unknown worker")
  })

  it("fails when dispatchTask is not configured", async () => {
    const tools = createTaskTools({
      ...baseContext,
      spawnWorker,
    })

    const result = await tools.taskStart.execute(
      { kind: "worker", workerId: "coder", task: "Test" },
      makeToolContext("session-a"),
    )

    const payload = JSON.parse(result as string)
    expect(payload.status).toBe("failed")
    expect(payload.error).toContain("Task dispatch is not configured")
  })

  it("fails when worker is missing and autoSpawn is false", async () => {
    const tools = createTaskTools({
      ...baseContext,
      dispatchTask,
    })

    const result = await tools.taskStart.execute(
      { kind: "worker", workerId: "coder", task: "Test", autoSpawn: false },
      makeToolContext("session-a"),
    )

    const payload = JSON.parse(result as string)
    expect(payload.status).toBe("failed")
    expect(payload.error).toContain("not running")
  })

  it("fails when autoSpawn is true but spawnWorker is missing", async () => {
    const tools = createTaskTools({
      ...baseContext,
      dispatchTask,
    })

    const result = await tools.taskStart.execute(
      { kind: "worker", workerId: "coder", task: "Test", autoSpawn: true },
      makeToolContext("session-a"),
    )

    const payload = JSON.parse(result as string)
    expect(payload.status).toBe("failed")
    expect(payload.error).toContain("Auto-spawn not configured")
  })

  it("dispatches a task and returns result on await", async () => {
    const tools = createTaskTools({
      ...baseContext,
      dispatchTask,
      spawnWorker,
    })

    const start = await tools.taskStart.execute(
      {
        kind: "worker",
        workerId: "coder",
        task: "Do work",
      },
      makeToolContext("session-a"),
    )

    const startPayload = JSON.parse(start as string)
    const awaited = await tools.taskAwait.execute(
      { taskId: startPayload.taskId },
      makeToolContext("session-a"),
    )

    const awaitedPayload = JSON.parse(awaited as string)
    expect(awaitedPayload.status).toBe("succeeded")
    expect(awaitedPayload.responseText).toBe("done:Do work")
  })

  it("scopes task peek to session", async () => {
    const tools = createTaskTools(baseContext)
    const job = jobRegistry.create({
      workerId: "coder",
      message: "Test",
      sessionId: "session-a",
    })

    const result = await tools.taskPeek.execute({ taskId: job.id }, makeToolContext("session-b"))

    const payload = JSON.parse(result as string)
    expect(payload.status).toBe("forbidden")
  })

  it("scopes task await to session", async () => {
    const tools = createTaskTools(baseContext)
    const job = jobRegistry.create({
      workerId: "coder",
      message: "Test",
      sessionId: "session-a",
    })

    const result = await tools.taskAwait.execute({ taskId: job.id }, makeToolContext("session-b"))

    const payload = JSON.parse(result as string)
    expect(payload.status).toBe("forbidden")
  })

  it("returns failed status when await times out", async () => {
    const tools = createTaskTools(baseContext)
    const job = jobRegistry.create({
      workerId: "coder",
      message: "Test",
      sessionId: "session-a",
    })

    const result = await tools.taskAwait.execute(
      { taskId: job.id, timeoutMs: 0 },
      makeToolContext("session-a"),
    )

    const payload = JSON.parse(result as string)
    expect(payload.status).toBe("failed")
    expect(payload.error).toContain("Timed out waiting")
  })

  it("scopes task list to session", async () => {
    const tools = createTaskTools(baseContext)
    jobRegistry.create({ workerId: "coder", message: "A", sessionId: "session-a" })
    jobRegistry.create({ workerId: "coder", message: "B", sessionId: "session-b" })

    const result = await tools.taskList.execute(
      { view: "tasks", format: "json" },
      makeToolContext("session-a"),
    )

    const payload = JSON.parse(result as string)
    expect(payload.length).toBe(1)
    expect(payload[0].sessionId).toBe("session-a")
  })

  it("rejects task cancel from another session", async () => {
    const tools = createTaskTools(baseContext)
    const job = jobRegistry.create({
      workerId: "coder",
      message: "Test",
      sessionId: "session-a",
    })

    const result = await tools.taskCancel.execute({ taskId: job.id }, makeToolContext("session-b"))

    const payload = JSON.parse(result as string)
    expect(payload.status).toBe("forbidden")
  })
})
