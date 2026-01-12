import { beforeEach, describe, expect, it } from "bun:test"

import { JobRegistry } from "../../../plugin/orchestrator/jobs.js"

describe("JobRegistry", () => {
  let registry: JobRegistry

  beforeEach(() => {
    registry = new JobRegistry()
  })

  describe("create", () => {
    it("creates a job with running status", () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })

      expect(job.id).toBeDefined()
      expect(job.workerId).toBe("test-worker")
      expect(job.message).toBe("Test task")
      expect(job.status).toBe("running")
      expect(job.startedAt).toBeDefined()
    })

    it("includes optional sessionId and requestedBy", () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
        sessionId: "session-123",
        requestedBy: "agent",
      })

      expect(job.sessionId).toBe("session-123")
      expect(job.requestedBy).toBe("agent")
    })
  })

  describe("get", () => {
    it("retrieves existing job", () => {
      const created = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })

      const retrieved = registry.get(created.id)
      expect(retrieved).toBe(created)
    })

    it("returns undefined for unknown job", () => {
      expect(registry.get("nonexistent")).toBeUndefined()
    })
  })

  describe("setResult", () => {
    it("marks job as succeeded", () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })

      registry.setResult(job.id, { responseText: "Success!" })

      const updated = registry.get(job.id)
      expect(updated?.status).toBe("succeeded")
      expect(updated?.responseText).toBe("Success!")
      expect(updated?.finishedAt).toBeDefined()
      expect(updated?.durationMs).toBeDefined()
    })
  })

  describe("setError", () => {
    it("marks job as failed", () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })

      registry.setError(job.id, { error: "Something went wrong" })

      const updated = registry.get(job.id)
      expect(updated?.status).toBe("failed")
      expect(updated?.error).toBe("Something went wrong")
    })
  })

  describe("cancel", () => {
    it("marks job as canceled", () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })

      registry.cancel(job.id, { reason: "User requested" })

      const updated = registry.get(job.id)
      expect(updated?.status).toBe("canceled")
      expect(updated?.error).toBe("User requested")
    })
  })

  describe("list", () => {
    it("lists jobs sorted by startedAt descending", () => {
      const job1 = registry.create({ workerId: "w1", message: "Task 1" })
      const job2 = registry.create({ workerId: "w2", message: "Task 2" })
      job2.startedAt = job1.startedAt + 10

      const list = registry.list()
      expect(list.length).toBe(2)
      expect(list[0].id).toBe(job2.id)
    })

    it("filters by workerId", () => {
      registry.create({ workerId: "w1", message: "Task 1" })
      registry.create({ workerId: "w2", message: "Task 2" })

      const list = registry.list({ workerId: "w1" })
      expect(list.length).toBe(1)
      expect(list[0].workerId).toBe("w1")
    })

    it("respects limit", () => {
      for (let i = 0; i < 10; i++) {
        registry.create({ workerId: "w", message: `Task ${i}` })
      }

      const list = registry.list({ limit: 5 })
      expect(list.length).toBe(5)
    })
  })

  describe("await", () => {
    it("resolves immediately for completed job", async () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })
      registry.setResult(job.id, { responseText: "Done" })

      const result = await registry.await(job.id)
      expect(result.status).toBe("succeeded")
    })

    it("waits for job completion", async () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })

      queueMicrotask(() => {
        registry.setResult(job.id, { responseText: "Done" })
      })

      const result = await registry.await(job.id, { timeoutMs: 1000 })
      expect(result.status).toBe("succeeded")
    })

    it("throws for unknown job", async () => {
      await expect(registry.await("nonexistent")).rejects.toThrow('Unknown job "nonexistent"')
    })

    it("rejects when timeout expires", async () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })

      await expect(registry.await(job.id, { timeoutMs: 0 })).rejects.toThrow(
        /Timed out waiting for job/,
      )
    })

    it("rejects when timer expires", async () => {
      const job = registry.create({
        workerId: "test-worker",
        message: "Test task",
      })

      await expect(registry.await(job.id, { timeoutMs: 10 })).rejects.toThrow(
        /Timed out waiting for job/,
      )
    })
  })
})
