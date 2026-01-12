/**
 * Job Registry
 *
 * Manages async task lifecycle with promise-based waiting.
 * Ported from Orchestra's WorkerJobRegistry pattern.
 */

import { randomUUID } from "node:crypto"
import type { Job, JobReport, JobStatus } from "./types.js"

const MAX_JOBS = 200
const MAX_JOB_AGE_MS = 24 * 60 * 60 * 1000

export class JobRegistry {
  private jobs = new Map<string, Job>()
  private waiters = new Map<string, Set<(job: Job) => void>>()

  /**
   * Create a new job and start tracking it.
   */
  create(input: {
    workerId: string
    message: string
    sessionId?: string
    requestedBy?: string
  }): Job {
    const id = randomUUID()
    const job: Job = {
      id,
      workerId: input.workerId,
      message: input.message,
      ...(input.sessionId ? { sessionId: input.sessionId } : {}),
      ...(input.requestedBy ? { requestedBy: input.requestedBy } : {}),
      status: "running",
      startedAt: Date.now(),
    }
    this.jobs.set(id, job)
    this.prune()
    return job
  }

  /**
   * Get a job by ID.
   */
  get(id: string): Job | undefined {
    return this.jobs.get(id)
  }

  /**
   * List jobs with optional filters.
   */
  list(options?: {
    workerId?: string
    sessionId?: string
    status?: JobStatus
    limit?: number
  }): Job[] {
    const limit = Math.max(1, options?.limit ?? 50)
    return [...this.jobs.values()]
      .filter((job) => (options?.workerId ? job.workerId === options.workerId : true))
      .filter((job) => (options?.sessionId ? job.sessionId === options.sessionId : true))
      .filter((job) => (options?.status ? job.status === options.status : true))
      .sort((a, b) => b.startedAt - a.startedAt)
      .slice(0, limit)
  }

  /**
   * Mark a job as succeeded with response text.
   */
  setResult(id: string, input: { responseText: string }): void {
    const job = this.jobs.get(id)
    if (!job || job.status !== "running") return
    job.status = "succeeded"
    job.responseText = input.responseText
    job.finishedAt = Date.now()
    job.durationMs = job.finishedAt - job.startedAt
    this.notify(id, job)
    this.prune()
  }

  /**
   * Mark a job as failed with an error message.
   */
  setError(id: string, input: { error: string }): void {
    const job = this.jobs.get(id)
    if (!job || job.status !== "running") return
    job.status = "failed"
    job.error = input.error
    job.finishedAt = Date.now()
    job.durationMs = job.finishedAt - job.startedAt
    this.notify(id, job)
    this.prune()
  }

  /**
   * Cancel a running job.
   */
  cancel(id: string, input?: { reason?: string }): void {
    const job = this.jobs.get(id)
    if (!job || job.status !== "running") return
    job.status = "canceled"
    if (input?.reason) job.error = input.reason
    job.finishedAt = Date.now()
    job.durationMs = job.finishedAt - job.startedAt
    this.notify(id, job)
    this.prune()
  }

  /**
   * Attach a report to a job (can be called even after completion).
   */
  attachReport(id: string, report: JobReport): void {
    const job = this.jobs.get(id)
    if (!job) return
    job.report = { ...(job.report ?? {}), ...report }
  }

  /**
   * Wait for a job to complete (promise-based).
   */
  async await(id: string, options?: { timeoutMs?: number }): Promise<Job> {
    const existing = this.jobs.get(id)
    if (!existing) throw new Error(`Unknown job "${id}"`)
    if (existing.status !== "running") return existing

    const timeoutMs = options?.timeoutMs ?? 600_000
    return await new Promise<Job>((resolve, reject) => {
      const timer = setTimeout(() => {
        this.offWaiter(id, onDone)
        reject(new Error(`Timed out waiting for job "${id}" after ${timeoutMs}ms`))
      }, timeoutMs)

      const onDone = (job: Job) => {
        clearTimeout(timer)
        resolve(job)
      }
      this.onWaiter(id, onDone)
    })
  }

  private onWaiter(id: string, cb: (job: Job) => void): void {
    const set = this.waiters.get(id) ?? new Set<(job: Job) => void>()
    set.add(cb)
    this.waiters.set(id, set)
  }

  private offWaiter(id: string, cb: (job: Job) => void): void {
    const set = this.waiters.get(id)
    if (!set) return
    set.delete(cb)
    if (set.size === 0) this.waiters.delete(id)
  }

  private notify(id: string, job: Job): void {
    const set = this.waiters.get(id)
    if (!set) return
    this.waiters.delete(id)
    for (const cb of set) {
      cb(job)
    }
  }

  private prune(): void {
    const now = Date.now()

    for (const [id, job] of this.jobs) {
      if (job.status === "running") continue
      const ageMs = now - (job.finishedAt ?? job.startedAt)
      if (ageMs <= MAX_JOB_AGE_MS) continue
      if (this.waiters.has(id)) continue
      this.jobs.delete(id)
    }

    if (this.jobs.size <= MAX_JOBS) return
    for (const [id, job] of this.jobs) {
      if (this.jobs.size <= MAX_JOBS) break
      if (job.status === "running") continue
      if (this.waiters.has(id)) continue
      this.jobs.delete(id)
    }
  }
}

export const jobRegistry = new JobRegistry()
