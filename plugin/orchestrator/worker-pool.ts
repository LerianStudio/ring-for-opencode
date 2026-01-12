/**
 * Worker Pool
 *
 * Unified worker lifecycle management with:
 * - In-memory worker tracking
 * - In-flight spawn deduplication
 * - Event-based status updates
 * - Session ownership tracking
 * - Worker limits and rate limiting (SEC-H1)
 * - State machine validation for status transitions (BL-H3)
 *
 * Ported from Orchestra's WorkerPool pattern.
 */

import type {
  DeviceRegistryEntry,
  DeviceRegistryFile,
  WorkerInstance,
  WorkerPoolCallback,
  WorkerPoolEvent,
  WorkerProfile,
  WorkerStatus,
} from "./types.js"
import { MAX_WORKERS, MAX_WORKERS_PER_SESSION, SPAWN_RATE_LIMIT_MS } from "./types.js"
import { validateServerUrl } from "./config.js"

const VALID_TRANSITIONS: Record<WorkerStatus, WorkerStatus[]> = {
  starting: ["ready", "error", "stopped"],
  ready: ["busy", "error", "stopped"],
  busy: ["ready", "error", "stopped"],
  error: ["stopped", "starting"],
  stopped: [],
}

function isValidTransition(from: WorkerStatus, to: WorkerStatus): boolean {
  return VALID_TRANSITIONS[from]?.includes(to) ?? false
}

export interface SpawnOptions {
  basePort: number
  timeout: number
  directory: string
  parentSessionId?: string
  forceNew?: boolean
}

export class WorkerPool {
  readonly workers: Map<string, WorkerInstance> = new Map()
  private listeners: Map<WorkerPoolEvent, Set<WorkerPoolCallback>> = new Map()
  private sessionWorkers: Map<string, Set<string>> = new Map()
  private inFlightSpawns: Map<string, Promise<WorkerInstance>> = new Map()
  private instanceId = ""
  private lastSpawnTime = 0

  setInstanceId(id: string): void {
    this.instanceId = id
  }

  async getOrSpawn(
    profile: WorkerProfile,
    options: SpawnOptions,
    spawnFn: (profile: WorkerProfile, options: SpawnOptions) => Promise<WorkerInstance>,
  ): Promise<WorkerInstance> {
    const existing = this.workers.get(profile.id)
    if (existing && existing.status !== "error" && existing.status !== "stopped") {
      return existing
    }

    const inFlight = this.inFlightSpawns.get(profile.id)
    if (inFlight) {
      return inFlight
    }

    const spawnPromise = spawnFn(profile, options)
    this.inFlightSpawns.set(profile.id, spawnPromise)

    try {
      const instance = await spawnPromise

      if (!this.workers.has(profile.id)) {
        try {
          this.register(instance)
        } catch (regError) {
          try {
            await instance.shutdown?.()
          } catch {
            // Ignore shutdown errors during cleanup
          }
          throw regError
        }
      }
      return instance
    } finally {
      if (this.inFlightSpawns.get(profile.id) === spawnPromise) {
        this.inFlightSpawns.delete(profile.id)
      }
    }
  }

  /**
   * Register a worker instance.
   * @throws Error if worker limit reached or rate limited (SEC-H1)
   */
  register(instance: WorkerInstance): void {
    if (this.workers.size >= MAX_WORKERS) {
      throw new Error(`Worker limit reached (max: ${MAX_WORKERS})`)
    }

    if (instance.parentSessionId) {
      const sessionCount = this.getWorkersForSession(instance.parentSessionId).length
      if (sessionCount >= MAX_WORKERS_PER_SESSION) {
        throw new Error(`Session worker limit reached (max: ${MAX_WORKERS_PER_SESSION})`)
      }
    }

    if (instance.serverUrl) {
      validateServerUrl(instance.serverUrl)
    }

    const now = Date.now()
    if (now - this.lastSpawnTime < SPAWN_RATE_LIMIT_MS) {
      throw new Error(`Rate limited: wait ${SPAWN_RATE_LIMIT_MS}ms between spawns`)
    }
    this.lastSpawnTime = now

    this.workers.set(instance.profile.id, instance)
    this.emit("spawn", instance)
  }

  unregister(id: string): boolean {
    const instance = this.workers.get(id)
    if (instance) {
      this.workers.delete(id)
      for (const [sessionId, ids] of this.sessionWorkers.entries()) {
        ids.delete(id)
        if (ids.size === 0) this.sessionWorkers.delete(sessionId)
      }
      this.emit("stop", instance)
      return true
    }
    return false
  }

  get(id: string): WorkerInstance | undefined {
    return this.workers.get(id)
  }

  list(): WorkerInstance[] {
    return Array.from(this.workers.values())
  }

  getVisionWorkers(): WorkerInstance[] {
    return this.list().filter(
      (worker) =>
        worker.profile.supportsVision && (worker.status === "ready" || worker.status === "busy"),
    )
  }

  getActiveWorkers(): WorkerInstance[] {
    return this.list().filter((worker) => worker.status === "ready" || worker.status === "busy")
  }

  getWorkersByStatus(status: WorkerStatus): WorkerInstance[] {
    return this.list().filter((worker) => worker.status === status)
  }

  getWorkersByCapability(capability: string): WorkerInstance[] {
    const lowerCap = capability.toLowerCase()
    return this.list().filter(
      (worker) =>
        worker.profile.purpose.toLowerCase().includes(lowerCap) ||
        worker.profile.whenToUse.toLowerCase().includes(lowerCap) ||
        worker.profile.id.toLowerCase().includes(lowerCap) ||
        (worker.profile.tags?.some((tag) => tag.toLowerCase().includes(lowerCap)) ?? false),
    )
  }

  updateStatus(id: string, status: WorkerStatus, error?: string): void {
    const instance = this.workers.get(id)
    if (!instance) return

    const prevStatus = instance.status

    if (!isValidTransition(prevStatus, status)) {
      console.warn(`[WorkerPool] Invalid state transition for ${id}: ${prevStatus} -> ${status}`)
      return
    }

    instance.status = status
    instance.lastActivity = new Date()
    if (error) instance.error = error

    if (status === "ready" && prevStatus !== "ready") {
      this.emit("ready", instance)
    } else if (status === "busy") {
      this.emit("busy", instance)
    } else if (status === "error") {
      this.emit("error", instance)
    }
    this.emit("update", instance)
  }

  trackOwnership(sessionId: string | undefined, workerId: string): void {
    if (!sessionId) return
    const next = this.sessionWorkers.get(sessionId) ?? new Set<string>()
    next.add(workerId)
    this.sessionWorkers.set(sessionId, next)
  }

  getWorkersForSession(sessionId: string): string[] {
    return [...(this.sessionWorkers.get(sessionId) ?? new Set<string>())]
  }

  clearSessionOwnership(sessionId: string): void {
    this.sessionWorkers.delete(sessionId)
  }

  on(event: WorkerPoolEvent, callback: WorkerPoolCallback): () => void {
    let set = this.listeners.get(event)
    if (!set) {
      set = new Set()
      this.listeners.set(event, set)
    }
    set.add(callback)
    return () => this.off(event, callback)
  }

  off(event: WorkerPoolEvent, callback: WorkerPoolCallback): void {
    this.listeners.get(event)?.delete(callback)
  }

  private emit(event: WorkerPoolEvent, instance: WorkerInstance): void {
    this.listeners.get(event)?.forEach((callback) => {
      try {
        callback(instance)
      } catch (err) {
        if (process.env.RING_DEBUG === "true") {
          console.warn(`[WorkerPool] Listener error on '${event}':`, err)
        }
      }
    })
  }

  getSummary(options: { maxWorkers?: number } = {}): string {
    const maxWorkers = options.maxWorkers ?? 12
    const workers = this.list()
      .sort((a, b) => a.profile.id.localeCompare(b.profile.id))
      .slice(0, Math.max(0, maxWorkers))

    if (workers.length === 0) {
      return "No workers currently registered."
    }

    const total = this.workers.size
    const lines = ["## Available Workers", ""]
    if (total > workers.length) {
      lines.push(`(showing ${workers.length} of ${total})`, "")
    }
    for (const worker of workers) {
      const status = worker.status === "ready" ? "available" : worker.status
      lines.push(`### ${worker.profile.name} (${worker.profile.id})`)
      lines.push(`- **Status**: ${status}`)
      lines.push(`- **Model**: ${worker.profile.model}`)
      lines.push(`- **Purpose**: ${worker.profile.purpose}`)
      lines.push(`- **When to use**: ${worker.profile.whenToUse}`)
      if (worker.profile.supportsVision) lines.push(`- **Supports Vision**: Yes`)
      if (worker.profile.supportsWeb) lines.push(`- **Supports Web**: Yes`)
      lines.push("")
    }

    lines.push("## How to Use Workers")
    lines.push("Use `task_start` to delegate work, then `task_await` to get results.")
    lines.push("Example: task_start({ kind: 'worker', workerId: 'docs', task: 'Find docs for X' })")
    lines.push("")

    return lines.join("\n")
  }

  toJSON(): Record<string, unknown>[] {
    return this.list().map((worker) => ({
      id: worker.profile.id,
      name: worker.profile.name,
      modelRef: worker.modelRef ?? worker.profile.model,
      model: worker.profile.model,
      modelPolicy: worker.modelPolicy ?? "dynamic",
      modelResolution: worker.modelResolution,
      kind: worker.kind ?? worker.profile.kind,
      execution: worker.execution,
      parentSessionId: worker.parentSessionId,
      purpose: worker.profile.purpose,
      whenToUse: worker.profile.whenToUse,
      status: worker.status,
      port: worker.port,
      pid: worker.pid,
      serverUrl: worker.serverUrl,
      supportsVision: worker.profile.supportsVision ?? false,
      supportsWeb: worker.profile.supportsWeb ?? false,
      lastActivity: worker.lastActivity?.toISOString(),
      currentTask: worker.currentTask,
      warning: worker.warning,
    }))
  }

  async stopAll(): Promise<void> {
    const workers = this.list()
    await Promise.allSettled(
      workers.map(async (worker) => {
        try {
          await worker.shutdown?.()
        } catch {
          // Ignore shutdown errors
        }
      }),
    )
    this.workers.clear()
    this.sessionWorkers.clear()
    this.inFlightSpawns.clear()
  }

  async stop(workerId: string): Promise<boolean> {
    const instance = this.workers.get(workerId)
    if (!instance) return false

    try {
      await instance.shutdown?.()
      this.updateStatus(workerId, "stopped")
      this.unregister(workerId)
      return true
    } catch (err) {
      if (process.env.RING_DEBUG === "true") {
        console.warn(`[WorkerPool] Stop failed for ${workerId}:`, err)
      }
      return false
    }
  }
}

export const workerPool = new WorkerPool()
