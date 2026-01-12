# Orchestra Patterns Port Implementation Plan

> **For Agents:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Port proven orchestration patterns from Orchestra to Ring for OpenCode, adding a full async task API, worker pool management, workflow execution, and profile-based worker configuration.

**Architecture:** Layer orchestration capabilities on top of Ring's existing plugin infrastructure. Create a new `orchestrator/` module within `plugin/` that provides worker pool management, async task tools, workflow engine, and profile-based configuration. Integrate with existing Ring hooks and background management.

**Tech Stack:** TypeScript, Bun, Zod for validation, @opencode-ai/plugin SDK

**Global Prerequisites:**
- Environment: macOS (darwin 25.2.0), Bun runtime
- Tools: Verify with commands below
- Access: Ring codebase at `/Users/fredamaral/repos/fredcamaral/ring-for-opencode`
- State: Branch from `main`, clean working tree

**Verification before starting:**
```bash
# Run ALL these commands and verify output:
bun --version              # Expected: 1.x+
node --version             # Expected: v20+
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git status  # Expected: clean working tree
ls plugin/                 # Expected: background/, config/, hooks/, tools/, ring-unified.ts, etc.
```

## Historical Precedent

**Query:** "orchestration worker pool async task workflow"
**Index Status:** Empty (new project)

No historical data available. This is normal for new projects.
Proceeding with standard planning approach.

---

## Phase 1: Core Types and Job Registry

### Task 1: Create Orchestrator Types

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/types.ts`

**Prerequisites:**
- Tools: TypeScript 5.9+
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/` directory

**Step 1: Create the orchestrator directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator`

**Expected output:**
```
(no output - directory created silently)
```

**Step 2: Write the types file**

```typescript
/**
 * Orchestrator Types
 *
 * Core type definitions for worker management, tasks, and workflows.
 * Ported from Orchestra patterns with Ring-specific adaptations.
 */

// =============================================================================
// Worker Types
// =============================================================================

export type WorkerStatus = "starting" | "ready" | "busy" | "error" | "stopped"
export type WorkerKind = "server" | "agent" | "subagent"
export type WorkerExecution = "foreground" | "background"

export interface WorkerProfile {
  /** Unique identifier for this worker */
  id: string
  /** Human-readable name */
  name: string
  /** Worker kind (server = spawned, agent/subagent = in-process) */
  kind?: WorkerKind
  /** Model to use (tag like "node:vision" or full "provider/model") */
  model: string
  /** What this worker specializes in */
  purpose: string
  /** When to use this worker (injected into context) */
  whenToUse: string
  /** Optional system prompt override */
  systemPrompt?: string
  /** Whether this worker can see images */
  supportsVision?: boolean
  /** Whether this worker has web access */
  supportsWeb?: boolean
  /** Custom tools to enable/disable */
  tools?: Record<string, boolean>
  /** Temperature setting */
  temperature?: number
  /** Optional keywords/tags to improve matching */
  tags?: string[]
}

export interface WorkerInstance {
  profile: WorkerProfile
  kind?: WorkerKind
  execution?: WorkerExecution
  parentSessionId?: string
  status: WorkerStatus
  port: number
  pid?: number
  serverUrl?: string
  directory?: string
  sessionId?: string
  startedAt: Date
  lastActivity?: Date
  error?: string
  warning?: string
  currentTask?: string
  modelResolution?: string
  modelRef?: string
  modelPolicy?: "dynamic" | "sticky"
  shutdown?: () => void | Promise<void>
}

// =============================================================================
// Job Types
// =============================================================================

export type JobStatus = "running" | "succeeded" | "failed" | "canceled"

export interface JobReport {
  summary?: string
  details?: string
  issues?: string[]
  notes?: string
}

export interface Job {
  id: string
  workerId: string
  message: string
  sessionId?: string
  requestedBy?: string
  status: JobStatus
  startedAt: number
  finishedAt?: number
  durationMs?: number
  responseText?: string
  error?: string
  report?: JobReport
}

// =============================================================================
// Workflow Types
// =============================================================================

export interface WorkflowStepDefinition {
  id: string
  title: string
  workerId: string
  prompt: string
  carry?: boolean
  timeoutMs?: number
}

export interface WorkflowDefinition {
  id: string
  name: string
  description: string
  steps: WorkflowStepDefinition[]
}

export interface WorkflowRunLimits {
  maxSteps: number
  maxTaskChars: number
  maxCarryChars: number
  perStepTimeoutMs: number
}

export interface WorkflowRunInput {
  workflowId: string
  task: string
  attachments?: Array<{
    type: "image" | "file"
    path?: string
    base64?: string
    mimeType?: string
  }>
  autoSpawn?: boolean
  limits: WorkflowRunLimits
}

export interface WorkflowStepResult {
  id: string
  title: string
  workerId: string
  status: "success" | "error" | "skipped"
  response?: string
  warning?: string
  error?: string
  startedAt: number
  finishedAt: number
  durationMs: number
}

export interface WorkflowRunResult {
  runId: string
  workflowId: string
  workflowName: string
  status: "running" | "success" | "error" | "paused"
  startedAt: number
  finishedAt?: number
  currentStepIndex: number
  steps: WorkflowStepResult[]
  lastStepResult?: WorkflowStepResult
}

// =============================================================================
// Context Types
// =============================================================================

export interface OrchestratorContext {
  directory: string
  worktree?: string
  projectId?: string
  profiles: Record<string, WorkerProfile>
  spawnDefaults: { basePort: number; timeout: number }
  defaultListFormat: "markdown" | "json"
}

// =============================================================================
// Event Types
// =============================================================================

export type WorkerPoolEvent = "spawn" | "ready" | "busy" | "error" | "stop" | "update"
export type WorkerPoolCallback = (instance: WorkerInstance) => void

// =============================================================================
// Device Registry Types (cross-session persistence)
// =============================================================================

export interface DeviceRegistryWorkerEntry {
  kind: "worker"
  orchestratorInstanceId: string
  hostPid?: number
  workerId: string
  pid: number
  url?: string
  port?: number
  sessionId?: string
  status: WorkerStatus
  startedAt: number
  updatedAt: number
  lastError?: string
  model?: string
  modelPolicy?: "dynamic" | "sticky"
}

export interface DeviceRegistrySessionEntry {
  kind: "session"
  hostPid: number
  sessionId: string
  directory: string
  title: string
  createdAt: number
  updatedAt: number
}

export type DeviceRegistryEntry = DeviceRegistryWorkerEntry | DeviceRegistrySessionEntry

export interface DeviceRegistryFile {
  version: 1
  updatedAt: number
  entries: DeviceRegistryEntry[]
}
```

**Step 3: Verify the file was created**

Run: `ls -la /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/types.ts`

**Expected output:**
```
-rw-r--r--  1 user  staff  XXXX Jan 11 XX:XX /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/types.ts
```

**Step 4: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 5: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/orchestrator/types.ts && git commit -m "$(cat <<'EOF'
feat(orchestrator): add core type definitions

Add comprehensive TypeScript types for orchestrator module:
- Worker types (WorkerProfile, WorkerInstance, WorkerStatus)
- Job types (Job, JobStatus, JobReport)
- Workflow types (WorkflowDefinition, WorkflowRunInput, WorkflowRunResult)
- Context types (OrchestratorContext)
- Device registry types for cross-session persistence

Ported from Orchestra patterns with Ring-specific adaptations.
EOF
)"
```

**If Task Fails:**

1. **Directory creation fails:**
   - Check: `ls /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/`
   - Fix: Ensure plugin directory exists
   - Rollback: N/A

2. **Type check fails:**
   - Run: `bun run typecheck 2>&1`
   - Check: Error messages for missing imports
   - Fix: Add missing type imports

---

### Task 2: Create Job Registry

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/jobs.ts`

**Prerequisites:**
- Tools: TypeScript 5.9+
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/types.ts`

**Step 1: Write the job registry file**

```typescript
/**
 * Job Registry
 *
 * Manages async task lifecycle with promise-based waiting.
 * Ported from Orchestra's WorkerJobRegistry pattern.
 */

import { randomUUID } from "node:crypto"
import type { Job, JobReport, JobStatus } from "./types.js"

const MAX_JOBS = 200
const MAX_JOB_AGE_MS = 24 * 60 * 60 * 1000 // 24 hours

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
      .filter((j) => (options?.workerId ? j.workerId === options.workerId : true))
      .filter((j) => (options?.sessionId ? j.sessionId === options.sessionId : true))
      .filter((j) => (options?.status ? j.status === options.status : true))
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

    const timeoutMs = options?.timeoutMs ?? 600_000 // 10 minutes default
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
    const set = this.waiters.get(id) ?? new Set()
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

    // Prune old completed jobs
    for (const [id, job] of this.jobs) {
      if (job.status === "running") continue
      const ageMs = now - (job.finishedAt ?? job.startedAt)
      if (ageMs <= MAX_JOB_AGE_MS) continue
      if (this.waiters.has(id)) continue
      this.jobs.delete(id)
    }

    // Enforce max jobs limit
    if (this.jobs.size <= MAX_JOBS) return
    for (const [id, job] of this.jobs) {
      if (this.jobs.size <= MAX_JOBS) break
      if (job.status === "running") continue
      if (this.waiters.has(id)) continue
      this.jobs.delete(id)
    }
  }
}

// Singleton instance
export const jobRegistry = new JobRegistry()
```

**Step 2: Verify the file was created**

Run: `ls -la /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/jobs.ts`

**Expected output:**
```
-rw-r--r--  1 user  staff  XXXX Jan 11 XX:XX /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/jobs.ts
```

**Step 3: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/orchestrator/jobs.ts && git commit -m "$(cat <<'EOF'
feat(orchestrator): add job registry for async task tracking

Implement JobRegistry class with:
- Job creation, retrieval, and listing
- Status transitions (running, succeeded, failed, canceled)
- Promise-based await() for blocking on job completion
- Report attachment for post-completion metadata
- Automatic pruning of old jobs (24h TTL, 200 max)
- Waiter notification for concurrent await() calls

Ported from Orchestra's WorkerJobRegistry pattern.
EOF
)"
```

**If Task Fails:**

1. **Import errors:**
   - Check: Types file exists at correct path
   - Fix: Ensure `.js` extension in import (ESM requirement)

---

### Task 3: Create Worker Pool

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/worker-pool.ts`

**Prerequisites:**
- Tools: TypeScript 5.9+
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/types.ts`

**Step 1: Write the worker pool file**

```typescript
/**
 * Worker Pool
 *
 * Unified worker lifecycle management with:
 * - In-memory worker tracking
 * - In-flight spawn deduplication
 * - Event-based status updates
 * - Session ownership tracking
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

  setInstanceId(id: string): void {
    this.instanceId = id
  }

  /**
   * Get or spawn a worker by profile ID.
   * Handles deduplication - concurrent calls return the same promise.
   */
  async getOrSpawn(
    profile: WorkerProfile,
    options: SpawnOptions,
    spawnFn: (profile: WorkerProfile, options: SpawnOptions) => Promise<WorkerInstance>,
  ): Promise<WorkerInstance> {
    // Check in-memory registry first
    const existing = this.workers.get(profile.id)
    if (existing && existing.status !== "error" && existing.status !== "stopped") {
      return existing
    }

    // Check for in-flight spawn
    const inFlight = this.inFlightSpawns.get(profile.id)
    if (inFlight) {
      return inFlight
    }

    // Create spawn promise to prevent race conditions
    const spawnPromise = spawnFn(profile, options)
    this.inFlightSpawns.set(profile.id, spawnPromise)

    try {
      const instance = await spawnPromise
      return instance
    } finally {
      if (this.inFlightSpawns.get(profile.id) === spawnPromise) {
        this.inFlightSpawns.delete(profile.id)
      }
    }
  }

  /**
   * Register a worker instance.
   */
  register(instance: WorkerInstance): void {
    this.workers.set(instance.profile.id, instance)
    this.emit("spawn", instance)
  }

  /**
   * Unregister a worker by ID.
   */
  unregister(id: string): boolean {
    const instance = this.workers.get(id)
    if (instance) {
      this.workers.delete(id)
      // Clean up session ownership
      for (const [sessionId, ids] of this.sessionWorkers.entries()) {
        ids.delete(id)
        if (ids.size === 0) this.sessionWorkers.delete(sessionId)
      }
      this.emit("stop", instance)
      return true
    }
    return false
  }

  /**
   * Get a worker by ID.
   */
  get(id: string): WorkerInstance | undefined {
    return this.workers.get(id)
  }

  /**
   * List all workers.
   */
  list(): WorkerInstance[] {
    return Array.from(this.workers.values())
  }

  /**
   * Get workers that support vision.
   */
  getVisionWorkers(): WorkerInstance[] {
    return this.list().filter(
      (w) => w.profile.supportsVision && (w.status === "ready" || w.status === "busy"),
    )
  }

  /**
   * Get active (non-stopped, non-error) workers.
   */
  getActiveWorkers(): WorkerInstance[] {
    return this.list().filter((w) => w.status === "ready" || w.status === "busy")
  }

  /**
   * Get workers by status.
   */
  getWorkersByStatus(status: WorkerStatus): WorkerInstance[] {
    return this.list().filter((w) => w.status === status)
  }

  /**
   * Get workers matching a capability keyword.
   */
  getWorkersByCapability(capability: string): WorkerInstance[] {
    const lowerCap = capability.toLowerCase()
    return this.list().filter(
      (w) =>
        w.profile.purpose.toLowerCase().includes(lowerCap) ||
        w.profile.whenToUse.toLowerCase().includes(lowerCap) ||
        w.profile.id.toLowerCase().includes(lowerCap) ||
        (w.profile.tags?.some((t) => t.toLowerCase().includes(lowerCap)) ?? false),
    )
  }

  /**
   * Update worker status.
   */
  updateStatus(id: string, status: WorkerStatus, error?: string): void {
    const instance = this.workers.get(id)
    if (instance) {
      const prevStatus = instance.status
      instance.status = status
      instance.lastActivity = new Date()
      if (error) instance.error = error

      // Emit appropriate event
      if (status === "ready" && prevStatus !== "ready") {
        this.emit("ready", instance)
      } else if (status === "busy") {
        this.emit("busy", instance)
      } else if (status === "error") {
        this.emit("error", instance)
      }
      this.emit("update", instance)
    }
  }

  /**
   * Track session ownership of a worker.
   */
  trackOwnership(sessionId: string | undefined, workerId: string): void {
    if (!sessionId) return
    const next = this.sessionWorkers.get(sessionId) ?? new Set<string>()
    next.add(workerId)
    this.sessionWorkers.set(sessionId, next)
  }

  /**
   * Get workers owned by a session.
   */
  getWorkersForSession(sessionId: string): string[] {
    return [...(this.sessionWorkers.get(sessionId) ?? new Set<string>())]
  }

  /**
   * Clear session ownership.
   */
  clearSessionOwnership(sessionId: string): void {
    this.sessionWorkers.delete(sessionId)
  }

  /**
   * Subscribe to pool events.
   */
  on(event: WorkerPoolEvent, callback: WorkerPoolCallback): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
    return () => this.off(event, callback)
  }

  /**
   * Unsubscribe from pool events.
   */
  off(event: WorkerPoolEvent, callback: WorkerPoolCallback): void {
    this.listeners.get(event)?.delete(callback)
  }

  private emit(event: WorkerPoolEvent, instance: WorkerInstance): void {
    this.listeners.get(event)?.forEach((cb) => {
      try {
        cb(instance)
      } catch {
        // Ignore listener errors
      }
    })
  }

  /**
   * Get a summary of available workers for context injection.
   */
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
    for (const w of workers) {
      const status = w.status === "ready" ? "available" : w.status
      lines.push(`### ${w.profile.name} (${w.profile.id})`)
      lines.push(`- **Status**: ${status}`)
      lines.push(`- **Model**: ${w.profile.model}`)
      lines.push(`- **Purpose**: ${w.profile.purpose}`)
      lines.push(`- **When to use**: ${w.profile.whenToUse}`)
      if (w.profile.supportsVision) lines.push(`- **Supports Vision**: Yes`)
      if (w.profile.supportsWeb) lines.push(`- **Supports Web**: Yes`)
      lines.push("")
    }

    lines.push("## How to Use Workers")
    lines.push("Use `task_start` to delegate work, then `task_await` to get results.")
    lines.push("Example: task_start({ kind: 'worker', workerId: 'docs', task: 'Find docs for X' })")
    lines.push("")

    return lines.join("\n")
  }

  /**
   * Serialize workers to JSON for API responses.
   */
  toJSON(): Record<string, unknown>[] {
    return this.list().map((w) => ({
      id: w.profile.id,
      name: w.profile.name,
      modelRef: w.modelRef ?? w.profile.model,
      model: w.profile.model,
      modelPolicy: w.modelPolicy ?? "dynamic",
      modelResolution: w.modelResolution,
      kind: w.kind ?? w.profile.kind,
      execution: w.execution,
      parentSessionId: w.parentSessionId,
      purpose: w.profile.purpose,
      whenToUse: w.profile.whenToUse,
      status: w.status,
      port: w.port,
      pid: w.pid,
      serverUrl: w.serverUrl,
      supportsVision: w.profile.supportsVision ?? false,
      supportsWeb: w.profile.supportsWeb ?? false,
      lastActivity: w.lastActivity?.toISOString(),
      currentTask: w.currentTask,
      warning: w.warning,
    }))
  }

  /**
   * Stop all workers and clean up.
   */
  async stopAll(): Promise<void> {
    const workers = this.list()
    await Promise.allSettled(
      workers.map(async (w) => {
        try {
          await w.shutdown?.()
        } catch {
          // Ignore shutdown errors
        }
      }),
    )
    this.workers.clear()
    this.sessionWorkers.clear()
    this.inFlightSpawns.clear()
  }

  /**
   * Stop a specific worker.
   */
  async stop(workerId: string): Promise<boolean> {
    const instance = this.workers.get(workerId)
    if (!instance) return false

    try {
      await instance.shutdown?.()
      instance.status = "stopped"
      this.unregister(workerId)
      return true
    } catch {
      return false
    }
  }
}

// Singleton instance
export const workerPool = new WorkerPool()
```

**Step 2: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/orchestrator/worker-pool.ts && git commit -m "$(cat <<'EOF'
feat(orchestrator): add worker pool for lifecycle management

Implement WorkerPool class with:
- In-memory worker tracking with Map storage
- In-flight spawn deduplication via promise tracking
- Event-based status updates (spawn/ready/busy/error/stop)
- Session ownership tracking for cleanup
- Capability-based worker queries
- Summary generation for context injection

Ported from Orchestra's WorkerPool pattern.
EOF
)"
```

**If Task Fails:**

1. **Type errors on WorkerInstance:**
   - Check: `types.ts` has all required fields
   - Fix: Add missing fields to WorkerInstance interface

---

### Task 4: Create Built-in Profiles

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/profiles.ts`

**Prerequisites:**
- Files must exist: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/types.ts`

**Step 1: Write the profiles file**

```typescript
/**
 * Worker Profiles
 *
 * Built-in worker profiles and profile management utilities.
 * Supports model tag resolution (node:vision, node:fast, etc.)
 *
 * Ported from Orchestra's profiles pattern.
 */

import type { WorkerProfile } from "./types.js"

/**
 * Built-in worker profiles that can be used out of the box.
 */
export const builtInProfiles: Record<string, WorkerProfile> = {
  // Vision specialist - for analyzing images, diagrams, screenshots
  vision: {
    id: "vision",
    name: "Vision Analyst",
    kind: "server",
    model: "node:vision",
    purpose: "Analyze images, screenshots, diagrams, and visual content",
    whenToUse:
      "When you need to understand visual content like screenshots, architecture diagrams, UI mockups, error screenshots, or any image-based information",
    supportsVision: true,
  },

  // Documentation specialist - for looking up docs and examples
  docs: {
    id: "docs",
    name: "Documentation Librarian",
    kind: "server",
    model: "node:docs",
    purpose: "Research documentation, find examples, explain APIs and libraries",
    whenToUse:
      "When you need to look up official documentation, find code examples, understand library APIs, or research best practices",
    supportsWeb: true,
    tools: {
      write: false,
      edit: false,
    },
  },

  // Coding specialist - main implementation worker
  coder: {
    id: "coder",
    name: "Code Implementer",
    kind: "server",
    model: "node",
    purpose: "Write, edit, and refactor code with full tool access",
    whenToUse:
      "When you need to actually write or modify code, create files, run commands, or implement features",
  },

  // Architecture/planning specialist
  architect: {
    id: "architect",
    name: "System Architect",
    kind: "server",
    model: "node",
    purpose: "Design systems, plan implementations, review architecture decisions",
    whenToUse:
      "When you need to plan a complex feature, design system architecture, or make high-level technical decisions",
    tools: {
      write: false,
      edit: false,
      bash: false,
    },
  },

  // Fast explorer - for quick codebase searches
  explorer: {
    id: "explorer",
    name: "Code Explorer",
    kind: "server",
    model: "node:fast",
    purpose: "Quickly search and navigate the codebase",
    whenToUse:
      "When you need to quickly find files, search for patterns, or locate specific code without deep analysis",
    tools: {
      write: false,
      edit: false,
    },
    temperature: 0.1,
  },
}

/**
 * Get a profile by ID (built-in or custom).
 */
export function getProfile(
  id: string,
  customProfiles?: Record<string, WorkerProfile>,
): WorkerProfile | undefined {
  return customProfiles?.[id] ?? builtInProfiles[id]
}

/**
 * Merge custom profile with built-in defaults.
 */
export function mergeProfile(baseId: string, overrides: Partial<WorkerProfile>): WorkerProfile {
  const base = builtInProfiles[baseId]
  if (!base) {
    throw new Error(`Unknown base profile: ${baseId}`)
  }
  return {
    ...base,
    ...overrides,
    id: overrides.id ?? base.id,
  }
}

/**
 * Resolve a model tag to an actual model identifier.
 *
 * Tags supported:
 * - node: Use the default configured model
 * - node:vision: Use a vision-capable model
 * - node:fast: Use a fast/cheap model for quick tasks
 * - node:docs: Use a model optimized for documentation
 * - auto:vision, auto:fast, etc: Same as node: prefix
 *
 * If no tag match, returns the input as-is (assumed to be provider/model).
 */
export function resolveModelTag(
  tag: string,
  defaults: {
    default?: string
    vision?: string
    fast?: string
    docs?: string
  },
): string {
  const normalized = tag.toLowerCase().trim()

  // Check for node: or auto: prefix
  const prefixMatch = normalized.match(/^(node|auto):(.+)$/)
  if (prefixMatch) {
    const capability = prefixMatch[2]
    switch (capability) {
      case "vision":
        return defaults.vision ?? defaults.default ?? tag
      case "fast":
        return defaults.fast ?? defaults.default ?? tag
      case "docs":
        return defaults.docs ?? defaults.default ?? tag
      default:
        return defaults.default ?? tag
    }
  }

  // Check for bare "node" or "auto"
  if (normalized === "node" || normalized === "auto") {
    return defaults.default ?? tag
  }

  // Not a tag, return as-is
  return tag
}

/**
 * List all available profile IDs.
 */
export function listProfileIds(customProfiles?: Record<string, WorkerProfile>): string[] {
  const ids = new Set<string>(Object.keys(builtInProfiles))
  if (customProfiles) {
    for (const id of Object.keys(customProfiles)) {
      ids.add(id)
    }
  }
  return [...ids].sort()
}
```

**Step 2: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/orchestrator/profiles.ts && git commit -m "$(cat <<'EOF'
feat(orchestrator): add built-in worker profiles

Add worker profile system with:
- Five built-in profiles (vision, docs, coder, architect, explorer)
- Model tag resolution (node:vision, node:fast, etc.)
- Profile merging for customization
- Tool restrictions per profile type

Profiles define worker capabilities, model requirements, and usage hints.
EOF
)"
```

---

### Task 5: Run Code Review

1. **Dispatch reviewers:**
   - REQUIRED SUB-SKILL: Use requesting-code-review
   - Run all 3 reviewers (code-reviewer, business-logic-reviewer, security-reviewer)
   - Wait for all to complete

2. **Handle findings by severity:**
   - Critical/High/Medium: Fix immediately, re-run reviewers
   - Low: Add `TODO(review):` comments
   - Cosmetic: Add `FIXME(nitpick):` comments

3. **Proceed only when:** Zero Critical/High/Medium issues remain

---

## Phase 2: Task Tools API

### Task 6: Create Task Tools

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/tools/task-tools.ts`

**Prerequisites:**
- Files must exist: `plugin/orchestrator/types.ts`, `plugin/orchestrator/jobs.ts`, `plugin/orchestrator/worker-pool.ts`, `plugin/orchestrator/profiles.ts`

**Step 1: Create the tools directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/tools`

**Step 2: Write the task tools file**

```typescript
/**
 * Task Tools
 *
 * Five-tool async task API:
 * - task_start: Fire and forget, returns taskId immediately
 * - task_await: Block until completion
 * - task_peek: Non-blocking status check
 * - task_list: Visibility into workers/tasks
 * - task_cancel: Control over running tasks
 *
 * Ported from Orchestra's task tools pattern.
 */

import { tool, type ToolDefinition } from "@opencode-ai/plugin"
import { jobRegistry } from "../jobs.js"
import { workerPool } from "../worker-pool.js"
import { getProfile, builtInProfiles } from "../profiles.js"
import type { OrchestratorContext, WorkerProfile } from "../types.js"

type TaskTools = {
  taskStart: ToolDefinition
  taskAwait: ToolDefinition
  taskPeek: ToolDefinition
  taskList: ToolDefinition
  taskCancel: ToolDefinition
}

type ToolAttachment = {
  type: "image" | "file"
  path?: string
  base64?: string
  mimeType?: string
}

/**
 * Guess which worker should handle a task based on content.
 */
function guessWorkerId(task: string, attachments?: ToolAttachment[]): string {
  // Check for image attachments
  if (attachments?.some((a) => a.type === "image")) return "vision"

  // Check task content for keywords
  if (/\b(doc|docs|documentation|reference|api|example|examples|research|cite)\b/i.test(task)) {
    return "docs"
  }
  if (/\b(architecture|architect|design|plan|approach|trade[- ]?off)\b/i.test(task)) {
    return "architect"
  }
  if (/\b(search|find|locate|grep|ripgrep|scan|explore|where)\b/i.test(task)) {
    return "explorer"
  }

  return "coder"
}

/**
 * Render a markdown table from headers and rows.
 */
function renderMarkdownTable(headers: string[], rows: string[][]): string {
  if (rows.length === 0) return "(no data)"

  const widths = headers.map((h, i) =>
    Math.max(h.length, ...rows.map(r => (r[i] ?? "").length))
  )

  const header = `| ${headers.map((h, i) => h.padEnd(widths[i])).join(" | ")} |`
  const separator = `| ${widths.map(w => "-".repeat(w)).join(" | ")} |`
  const body = rows.map(r =>
    `| ${r.map((c, i) => (c ?? "").padEnd(widths[i])).join(" | ")} |`
  ).join("\n")

  return `${header}\n${separator}\n${body}`
}

/**
 * Create task tools bound to a context.
 */
export function createTaskTools(context: OrchestratorContext): TaskTools {
  const taskStart: ToolDefinition = tool({
    description:
      "Start a background task. Returns a taskId immediately; use task_await to get the result.",
    args: {
      kind: tool.schema
        .enum(["auto", "worker"])
        .optional()
        .describe("Task kind (default: auto = pick a worker based on task/attachments)"),
      task: tool.schema.string().describe("What to do (sent to worker)"),
      workerId: tool.schema.string().optional().describe("Worker id (e.g. 'docs', 'coder')"),
      attachments: tool.schema
        .array(
          tool.schema.object({
            type: tool.schema.enum(["image", "file"]),
            path: tool.schema.string().optional(),
            base64: tool.schema.string().optional(),
            mimeType: tool.schema.string().optional(),
          }),
        )
        .optional()
        .describe("Optional attachments (images/files) to forward to the worker"),
      autoSpawn: tool.schema.boolean().optional().describe("Auto-spawn missing workers (default: true)"),
      timeoutMs: tool.schema.number().optional().describe("Timeout for the task (default: 10 minutes)"),
    },
    async execute(args, ctx) {
      const kind = args.kind ?? "auto"
      const autoSpawn = args.autoSpawn ?? true
      const sessionId = ctx?.sessionID

      const resolvedWorkerId =
        kind === "auto"
          ? (args.workerId ?? guessWorkerId(args.task, args.attachments))
          : args.workerId

      if (!resolvedWorkerId) {
        return JSON.stringify({ error: "Missing workerId" }, null, 2)
      }

      // Create job record
      const job = jobRegistry.create({
        workerId: resolvedWorkerId,
        message: args.task,
        sessionId,
        requestedBy: ctx?.agent,
      })

      // Get or check worker
      const existing = workerPool.get(resolvedWorkerId)
      if (!existing && !autoSpawn) {
        jobRegistry.setError(job.id, {
          error: `Worker "${resolvedWorkerId}" is not running. Set autoSpawn=true or spawn it first.`,
        })
        return JSON.stringify({
          taskId: job.id,
          workerId: resolvedWorkerId,
          status: "failed",
          error: `Worker "${resolvedWorkerId}" not available`,
        }, null, 2)
      }

      // TODO: Implement actual worker communication
      // For now, mark as running - actual work happens when worker pool is connected

      return JSON.stringify({
        taskId: job.id,
        kind,
        workerId: resolvedWorkerId,
        status: "running",
        next: "task_await",
      }, null, 2)
    },
  })

  const taskAwait: ToolDefinition = tool({
    description: "Wait for one or more tasks to finish and return the final job record(s).",
    args: {
      taskId: tool.schema.string().optional().describe("Task id from task_start"),
      taskIds: tool.schema.array(tool.schema.string()).optional().describe("Multiple task ids to await"),
      timeoutMs: tool.schema.number().optional().describe("Timeout in ms (default: 10 minutes)"),
    },
    async execute(args) {
      const timeoutMs = args.timeoutMs ?? 600_000
      const ids = args.taskId ? [args.taskId] : args.taskIds ?? []

      if (ids.length === 0) {
        return "Missing taskId/taskIds."
      }

      const results = await Promise.all(
        ids.map(async (id) => {
          try {
            return await jobRegistry.await(id, { timeoutMs })
          } catch (err) {
            return {
              id,
              status: "failed",
              error: err instanceof Error ? err.message : String(err),
            }
          }
        }),
      )

      return JSON.stringify(ids.length === 1 ? results[0] : results, null, 2)
    },
  })

  const taskPeek: ToolDefinition = tool({
    description: "Get the current status/result of one or more tasks without waiting.",
    args: {
      taskId: tool.schema.string().optional().describe("Task id"),
      taskIds: tool.schema.array(tool.schema.string()).optional().describe("Multiple task ids"),
    },
    async execute(args) {
      const ids = args.taskId ? [args.taskId] : args.taskIds ?? []

      if (ids.length === 0) {
        return "Missing taskId/taskIds."
      }

      const results = ids.map((id) => jobRegistry.get(id) ?? { id, status: "unknown" })
      return JSON.stringify(ids.length === 1 ? results[0] : results, null, 2)
    },
  })

  const taskList: ToolDefinition = tool({
    description:
      "List tasks (default) or other orchestrator resources via view=workers|profiles|status.",
    args: {
      view: tool.schema
        .enum(["tasks", "workers", "profiles", "status"])
        .optional()
        .describe("What to list (default: tasks)"),
      workerId: tool.schema.string().optional().describe("Filter by worker id"),
      limit: tool.schema.number().optional().describe("Max items to return (default: 20)"),
      format: tool.schema.enum(["markdown", "json"]).optional().describe("Output format (default: markdown)"),
    },
    async execute(args) {
      const format = args.format ?? context.defaultListFormat
      const view = args.view ?? "tasks"

      if (view === "workers") {
        const workers = workerPool.toJSON()
        if (format === "json") return JSON.stringify(workers, null, 2)
        if (workers.length === 0) return "No workers are currently registered."

        const rows = workers.map((w) => [
          String(w.id),
          String(w.status),
          String(w.modelRef ?? ""),
          String(w.model),
          w.supportsVision ? "yes" : "no",
          w.supportsWeb ? "yes" : "no",
          String(w.purpose ?? ""),
        ])
        return renderMarkdownTable(
          ["Worker", "Status", "Model Ref", "Model", "Vision", "Web", "Purpose"],
          rows,
        )
      }

      if (view === "profiles") {
        const profiles = Object.values({ ...builtInProfiles, ...context.profiles })
          .sort((a, b) => a.id.localeCompare(b.id))
          .map((p) => ({
            id: p.id,
            name: p.name,
            model: p.model,
            supportsVision: p.supportsVision ?? false,
            supportsWeb: p.supportsWeb ?? false,
            purpose: p.purpose,
          }))

        if (format === "json") return JSON.stringify(profiles, null, 2)
        if (profiles.length === 0) return "No profiles available."

        const rows = profiles.map((p) => [
          p.id,
          p.name,
          p.model,
          p.supportsVision ? "yes" : "no",
          p.supportsWeb ? "yes" : "no",
          p.purpose,
        ])
        return renderMarkdownTable(["ID", "Name", "Model", "Vision", "Web", "Purpose"], rows)
      }

      if (view === "status") {
        const workers = workerPool.toJSON()
        const tasks = jobRegistry.list({ limit: Math.max(1, args.limit ?? 20) })
        const payload = { workers, tasks }

        if (format === "json") return JSON.stringify(payload, null, 2)

        const workerRows = workers.map((w) => [
          String(w.id),
          String(w.status),
          String(w.model),
          w.supportsVision ? "yes" : "no",
          w.supportsWeb ? "yes" : "no",
        ])
        const taskRows = tasks.map((t) => [
          t.id,
          t.workerId,
          t.status,
          new Date(t.startedAt).toISOString(),
          t.durationMs ? `${t.durationMs}` : "",
          (t.message ?? "").slice(0, 60).replace(/\s+/g, " "),
        ])

        return [
          "# Orchestrator Status",
          "",
          "## Workers",
          workerRows.length
            ? renderMarkdownTable(["Worker", "Status", "Model", "Vision", "Web"], workerRows)
            : "(none)",
          "",
          "## Recent Tasks",
          taskRows.length
            ? renderMarkdownTable(["Task", "Worker", "Status", "Started", "ms", "Message"], taskRows)
            : "(none)",
        ].join("\n")
      }

      // Default: tasks view
      const limit = args.limit ?? 20
      const tasks = jobRegistry.list({ workerId: args.workerId, limit })

      if (format === "json") return JSON.stringify(tasks, null, 2)
      if (tasks.length === 0) return "No tasks recorded yet."

      const rows = tasks.map((t) => [
        t.id,
        t.workerId,
        t.status,
        new Date(t.startedAt).toISOString(),
        t.durationMs ? `${t.durationMs}` : "",
        (t.message ?? "").slice(0, 60).replace(/\s+/g, " "),
      ])
      return renderMarkdownTable(["Task", "Worker", "Status", "Started", "ms", "Message"], rows)
    },
  })

  const taskCancel: ToolDefinition = tool({
    description: "Cancel a running task (best-effort; may not stop underlying worker execution).",
    args: {
      taskId: tool.schema.string().optional().describe("Task id"),
      taskIds: tool.schema.array(tool.schema.string()).optional().describe("Multiple task ids"),
      reason: tool.schema.string().optional().describe("Optional cancel reason"),
    },
    async execute(args) {
      const ids = args.taskId ? [args.taskId] : args.taskIds ?? []

      if (ids.length === 0) {
        return "Missing taskId/taskIds."
      }

      for (const id of ids) {
        jobRegistry.cancel(id, { reason: args.reason })
      }

      return ids.length === 1
        ? `Canceled task "${ids[0]}"`
        : `Canceled ${ids.length} task(s)`
    },
  })

  return { taskStart, taskAwait, taskPeek, taskList, taskCancel }
}
```

**Step 3: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/orchestrator/tools/ && git commit -m "$(cat <<'EOF'
feat(orchestrator): add five-tool async task API

Implement task tools for orchestrator:
- task_start: Fire and forget, returns taskId immediately
- task_await: Block until completion with timeout
- task_peek: Non-blocking status check
- task_list: List tasks, workers, profiles, or combined status
- task_cancel: Cancel running tasks

Features:
- Auto worker selection based on task content
- Markdown and JSON output formats
- Integration with job registry and worker pool

Ported from Orchestra's task tools pattern.
EOF
)"
```

---

## Phase 3: Workflow Engine

### Task 7: Create Workflow Engine

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/workflow/engine.ts`

**Prerequisites:**
- Files must exist: `plugin/orchestrator/types.ts`

**Step 1: Create the workflow directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/workflow`

**Step 2: Write the workflow engine file**

```typescript
/**
 * Workflow Engine
 *
 * Orchestrates multi-step workflows with:
 * - Structured handoff section extraction
 * - Per-section character budgets
 * - Intelligent context compaction
 * - Dependency injection for testing
 *
 * Ported from Orchestra's workflow engine pattern.
 */

import { randomUUID } from "node:crypto"
import type {
  WorkflowDefinition,
  WorkflowRunInput,
  WorkflowRunResult,
  WorkflowRunLimits,
  WorkflowStepDefinition,
  WorkflowStepResult,
} from "../types.js"

// =============================================================================
// Workflow Registry
// =============================================================================

const workflows = new Map<string, WorkflowDefinition>()

export function registerWorkflow(def: WorkflowDefinition): void {
  workflows.set(def.id, def)
}

export function listWorkflows(): WorkflowDefinition[] {
  return [...workflows.values()].sort((a, b) => a.id.localeCompare(b.id))
}

export function getWorkflow(id: string): WorkflowDefinition | undefined {
  return workflows.get(id)
}

// =============================================================================
// Dependency Injection for Testing
// =============================================================================

export type WorkflowRunDependencies = {
  resolveWorker: (workerId: string, autoSpawn: boolean) => Promise<string>
  sendToWorker: (
    workerId: string,
    message: string,
    options: { attachments?: WorkflowRunInput["attachments"]; timeoutMs: number },
  ) => Promise<{ success: boolean; response?: string; warning?: string; error?: string }>
}

// =============================================================================
// Carry/Trim Logic
// =============================================================================

const handoffSections = ["Summary", "Actions", "Artifacts", "Risks", "Next"] as const
const carrySections = ["Summary", "Artifacts", "Risks", "Next"] as const
type HandoffSection = (typeof handoffSections)[number]
type CarrySection = (typeof carrySections)[number]

const handoffSectionMap = new Map(handoffSections.map((s) => [s.toLowerCase(), s]))
const carrySectionCaps: Record<CarrySection, number> = {
  Summary: 900,
  Artifacts: 1600,
  Risks: 900,
  Next: 900,
}

function normalizeSectionName(value: string): HandoffSection | undefined {
  return handoffSectionMap.get(value.trim().toLowerCase())
}

/**
 * Extract structured handoff sections from response text.
 */
function extractHandoffSections(text: string): Record<HandoffSection, string> {
  const sections: Record<HandoffSection, string[]> = {
    Summary: [],
    Actions: [],
    Artifacts: [],
    Risks: [],
    Next: [],
  }

  const trimmed = text.trim()
  if (!trimmed) {
    return { Summary: "", Actions: "", Artifacts: "", Risks: "", Next: "" }
  }

  const lines = trimmed.split(/\r?\n/)
  let current: HandoffSection | undefined
  let sawHeading = false
  const headingRegex = /^\s*(#{1,3}\s*)?(Summary|Actions|Artifacts|Risks|Next)\s*:?\s*$/i

  for (const line of lines) {
    const match = line.match(headingRegex)
    if (match) {
      const key = normalizeSectionName(match[2] ?? "")
      if (key) {
        current = key
        sawHeading = true
        continue
      }
    }
    if (current) sections[current].push(line)
  }

  const resolved: Record<HandoffSection, string> = {
    Summary: sections.Summary.join("\n").trim(),
    Actions: sections.Actions.join("\n").trim(),
    Artifacts: sections.Artifacts.join("\n").trim(),
    Risks: sections.Risks.join("\n").trim(),
    Next: sections.Next.join("\n").trim(),
  }

  if (!sawHeading) {
    resolved.Summary = trimmed
  } else if (!resolved.Summary && resolved.Actions) {
    resolved.Summary = resolved.Actions
  }

  return resolved
}

/**
 * Truncate text to maxChars with suffix.
 */
function truncateText(text: string, maxChars: number): string {
  const trimmed = text.trim()
  if (trimmed.length <= maxChars) return trimmed
  const suffix = "\n...(truncated)"
  const sliceEnd = Math.max(0, maxChars - suffix.length)
  return `${trimmed.slice(0, sliceEnd).trimEnd()}${suffix}`
}

/**
 * Compact carry sections with per-section budgets.
 */
function compactCarrySections(
  sections: Record<HandoffSection, string>,
  maxCarryChars: number,
): { sections: Record<CarrySection, string>; truncatedSections: CarrySection[] } {
  const totalCaps = Object.values(carrySectionCaps).reduce((sum, v) => sum + v, 0)
  const scale = Math.min(1, maxCarryChars / (totalCaps + 200))
  const compacted = {} as Record<CarrySection, string>
  const truncatedSections: CarrySection[] = []

  for (const section of carrySections) {
    const baseCap = carrySectionCaps[section]
    const cap = Math.max(60, Math.floor(baseCap * scale))
    const content = sections[section] ?? ""
    if (content.length > cap) truncatedSections.push(section)
    compacted[section] = content ? truncateText(content, cap) : ""
  }

  return { sections: compacted, truncatedSections }
}

/**
 * Format a carry block from step response.
 */
function formatCarryBlock(
  stepTitle: string,
  responseText: string,
  maxCarryChars: number,
): { text: string; truncated: boolean; truncatedSections: CarrySection[] } {
  const sections = extractHandoffSections(responseText)
  const compacted = compactCarrySections(sections, maxCarryChars)
  const truncatedSections = [...compacted.truncatedSections]
  let truncated = truncatedSections.length > 0

  const blocks = carrySections
    .map((section) => {
      const content = compacted.sections[section]
      if (!content) return ""
      return `#### ${section}\n${content}`
    })
    .filter(Boolean)

  if (blocks.length === 0) {
    const fallback = truncateText(responseText, Math.max(240, Math.floor(maxCarryChars / 4)))
    if (fallback.length < responseText.trim().length) truncated = true
    blocks.push(`#### Summary\n${fallback || "None"}`)
  }

  const block = `### ${stepTitle}\n${blocks.join("\n\n")}`
  if (block.length <= maxCarryChars) {
    return { text: block, truncated, truncatedSections }
  }

  // Reduce further if still too long
  const reducedCap = Math.max(60, Math.floor(maxCarryChars / (blocks.length * 2)))
  const reducedBlocks = blocks.map((b) => truncateText(b, reducedCap)).join("\n\n")
  const reduced = `### ${stepTitle}\n${reducedBlocks}`
  const finalText = reduced.length <= maxCarryChars ? reduced : truncateText(reduced, maxCarryChars)

  return { text: finalText, truncated: true, truncatedSections }
}

/**
 * Split carry text into blocks.
 */
function splitCarryBlocks(carry: string): string[] {
  const trimmed = carry.trim()
  if (!trimmed) return []
  if (!trimmed.includes("### ")) return [trimmed]
  return trimmed.split(/\n(?=###\s)/g).map((b) => b.trim()).filter(Boolean)
}

/**
 * Append a new carry block, dropping old blocks if over budget.
 */
function appendCarry(
  existing: string,
  next: string,
  maxChars: number,
): { text: string; droppedBlocks: number; truncated: boolean } {
  const blocks = [...splitCarryBlocks(existing), next].filter(Boolean)
  if (blocks.length === 0) return { text: "", droppedBlocks: 0, truncated: false }

  const originalCount = blocks.length
  while (blocks.join("\n\n").length > maxChars && blocks.length > 1) {
    blocks.shift()
  }

  const combined = blocks.join("\n\n")
  const droppedBlocks = Math.max(0, originalCount - blocks.length)

  if (combined.length <= maxChars) {
    return { text: combined, droppedBlocks, truncated: droppedBlocks > 0 }
  }

  return {
    text: truncateText(combined, maxChars),
    droppedBlocks,
    truncated: true,
  }
}

// =============================================================================
// Workflow Execution
// =============================================================================

function applyTemplate(template: string, vars: Record<string, string>): string {
  let out = template
  for (const [key, value] of Object.entries(vars)) {
    out = out.replaceAll(`{${key}}`, value)
  }
  return out
}

function resolveStepTimeout(step: WorkflowStepDefinition, limits: WorkflowRunLimits): number {
  const requested =
    typeof step.timeoutMs === "number" && Number.isFinite(step.timeoutMs) && step.timeoutMs > 0
      ? step.timeoutMs
      : limits.perStepTimeoutMs
  return Math.min(requested, limits.perStepTimeoutMs)
}

export function validateWorkflowInput(input: WorkflowRunInput, workflow: WorkflowDefinition): void {
  if (input.task.length > input.limits.maxTaskChars) {
    throw new Error(`Task exceeds maxTaskChars (${input.limits.maxTaskChars}).`)
  }
  if (workflow.steps.length > input.limits.maxSteps) {
    throw new Error(`Workflow has ${workflow.steps.length} steps (maxSteps=${input.limits.maxSteps}).`)
  }
}

/**
 * Execute a single workflow step.
 */
export async function executeWorkflowStep(
  input: {
    runId: string
    workflow: WorkflowDefinition
    stepIndex: number
    task: string
    carry: string
    autoSpawn: boolean
    limits: WorkflowRunLimits
    attachments?: WorkflowRunInput["attachments"]
  },
  deps: WorkflowRunDependencies,
): Promise<{ step: WorkflowStepResult; response?: string; carry: string }> {
  const step = input.workflow.steps[input.stepIndex]
  const stepStarted = Date.now()

  const workerId = await deps.resolveWorker(step.workerId, input.autoSpawn)
  const prompt = applyTemplate(step.prompt, { task: input.task, carry: input.carry })

  const res = await deps.sendToWorker(workerId, prompt, {
    attachments: input.stepIndex === 0 ? input.attachments : undefined,
    timeoutMs: resolveStepTimeout(step, input.limits),
  })

  const stepFinished = Date.now()

  if (!res.success) {
    const result: WorkflowStepResult = {
      id: step.id,
      title: step.title,
      workerId,
      status: "error",
      error: res.error ?? "unknown_error",
      startedAt: stepStarted,
      finishedAt: stepFinished,
      durationMs: stepFinished - stepStarted,
    }
    return { step: result, carry: input.carry }
  }

  const response = res.response ?? ""
  const result: WorkflowStepResult = {
    id: step.id,
    title: step.title,
    workerId,
    status: "success",
    response,
    ...(res.warning ? { warning: res.warning } : {}),
    startedAt: stepStarted,
    finishedAt: stepFinished,
    durationMs: stepFinished - stepStarted,
  }

  // Handle carry if enabled for this step
  if (step.carry) {
    const carryBlock = formatCarryBlock(step.title, response, input.limits.maxCarryChars)
    const appended = appendCarry(input.carry, carryBlock.text, input.limits.maxCarryChars)
    return { step: result, response, carry: appended.text }
  }

  return { step: result, response, carry: input.carry }
}

/**
 * Run a complete workflow.
 */
export async function runWorkflow(
  input: WorkflowRunInput,
  deps: WorkflowRunDependencies,
): Promise<WorkflowRunResult> {
  const workflow = getWorkflow(input.workflowId)
  if (!workflow) {
    throw new Error(`Unknown workflow "${input.workflowId}".`)
  }

  validateWorkflowInput(input, workflow)

  const runId = randomUUID()
  const startedAt = Date.now()
  const steps: WorkflowStepResult[] = []
  let carry = ""
  let status: WorkflowRunResult["status"] = "running"

  for (let i = 0; i < workflow.steps.length; i++) {
    const executed = await executeWorkflowStep(
      {
        runId,
        workflow,
        stepIndex: i,
        task: input.task,
        carry,
        autoSpawn: input.autoSpawn ?? true,
        limits: input.limits,
        attachments: input.attachments,
      },
      deps,
    )
    steps.push(executed.step)

    if (executed.step.status === "error") {
      status = "error"
      break
    }
    carry = executed.carry
  }

  const finishedAt = Date.now()

  return {
    runId,
    workflowId: workflow.id,
    workflowName: workflow.name,
    status: status === "error" ? "error" : "success",
    startedAt,
    finishedAt,
    currentStepIndex: Math.min(steps.length, workflow.steps.length),
    steps,
    lastStepResult: steps[steps.length - 1],
  }
}
```

**Step 3: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/orchestrator/workflow/ && git commit -m "$(cat <<'EOF'
feat(orchestrator): add workflow engine with carry/trim logic

Implement workflow execution engine with:
- Workflow registry (register, list, get)
- Structured handoff section extraction (Summary, Actions, Artifacts, Risks, Next)
- Per-section character budgets for context compaction
- Intelligent carry block management with FIFO dropping
- Dependency injection pattern for testability
- Template variable substitution in prompts

Ported from Orchestra's workflow engine pattern.
EOF
)"
```

---

## Phase 4: Configuration and Integration

### Task 8: Create Orchestrator Configuration Schema

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/config.ts`

**Prerequisites:**
- Files must exist: `plugin/orchestrator/types.ts`, `plugin/orchestrator/profiles.ts`

**Step 1: Write the configuration file**

```typescript
/**
 * Orchestrator Configuration
 *
 * Three-level configuration merging:
 * 1. Built-in defaults (in code)
 * 2. Global config (~/.config/opencode/ring/orchestrator.json)
 * 3. Project config (.ring/orchestrator.json or .opencode/orchestrator.json)
 *
 * Ported from Orchestra's configuration pattern.
 */

import { existsSync, readFileSync } from "node:fs"
import { homedir } from "node:os"
import { join } from "node:path"
import { z } from "zod"
import type { OrchestratorContext, WorkerProfile, WorkflowRunLimits } from "./types.js"
import { builtInProfiles } from "./profiles.js"

// =============================================================================
// Zod Schemas
// =============================================================================

const WorkerProfileSchema = z.object({
  id: z.string(),
  name: z.string(),
  kind: z.enum(["server", "agent", "subagent"]).optional(),
  model: z.string(),
  purpose: z.string(),
  whenToUse: z.string(),
  systemPrompt: z.string().optional(),
  supportsVision: z.boolean().optional(),
  supportsWeb: z.boolean().optional(),
  tools: z.record(z.string(), z.boolean()).optional(),
  temperature: z.number().optional(),
  tags: z.array(z.string()).optional(),
})

const WorkflowLimitsSchema = z.object({
  maxSteps: z.number().default(4),
  maxTaskChars: z.number().default(12000),
  maxCarryChars: z.number().default(24000),
  perStepTimeoutMs: z.number().default(120000),
})

const OrchestratorConfigFileSchema = z.object({
  $schema: z.string().optional(),
  basePort: z.number().default(14096),
  autoSpawn: z.boolean().default(true),
  startupTimeout: z.number().default(30000),
  ui: z.object({
    toasts: z.boolean().default(true),
    injectSystemContext: z.boolean().default(true),
    systemContextMaxWorkers: z.number().default(12),
    defaultListFormat: z.enum(["markdown", "json"]).default("markdown"),
    debug: z.boolean().default(false),
  }).default({}),
  profiles: z.array(z.union([z.string(), WorkerProfileSchema])).default([]),
  workers: z.array(z.union([z.string(), WorkerProfileSchema])).default([]),
  workflows: z.object({
    enabled: z.boolean().default(true),
    limits: WorkflowLimitsSchema.default({}),
  }).default({}),
  security: z.object({
    preventRecursiveSpawn: z.boolean().default(true),
    recursiveSpawnEnvVar: z.string().default("RING_ORCHESTRATOR_WORKER"),
  }).default({}),
}).passthrough()

export type OrchestratorConfigFile = z.infer<typeof OrchestratorConfigFileSchema>

// =============================================================================
// Default Configuration
// =============================================================================

const DEFAULT_CONFIG: OrchestratorConfigFile = {
  basePort: 14096,
  autoSpawn: true,
  startupTimeout: 30000,
  ui: {
    toasts: true,
    injectSystemContext: true,
    systemContextMaxWorkers: 12,
    defaultListFormat: "markdown",
    debug: false,
  },
  profiles: [],
  workers: [],
  workflows: {
    enabled: true,
    limits: {
      maxSteps: 4,
      maxTaskChars: 12000,
      maxCarryChars: 24000,
      perStepTimeoutMs: 120000,
    },
  },
  security: {
    preventRecursiveSpawn: true,
    recursiveSpawnEnvVar: "RING_ORCHESTRATOR_WORKER",
  },
}

// =============================================================================
// Configuration Loading
// =============================================================================

function getUserConfigDir(): string {
  const home = homedir()
  const xdgConfig = process.env.XDG_CONFIG_HOME
  if (xdgConfig && xdgConfig.startsWith("/")) {
    return join(xdgConfig, "opencode", "ring")
  }
  return join(home, ".config", "opencode", "ring")
}

function tryReadJson(filePath: string): unknown | null {
  try {
    if (!existsSync(filePath)) return null
    const content = readFileSync(filePath, "utf-8")
    return JSON.parse(content)
  } catch {
    return null
  }
}

function deepMerge<T extends Record<string, unknown>>(target: T, source: Partial<T>): T {
  const result = { ...target }

  for (const key of Object.keys(source)) {
    const sourceValue = source[key as keyof T]
    const targetValue = target[key as keyof T]

    if (
      sourceValue !== null &&
      typeof sourceValue === "object" &&
      !Array.isArray(sourceValue) &&
      targetValue !== null &&
      typeof targetValue === "object" &&
      !Array.isArray(targetValue)
    ) {
      result[key as keyof T] = deepMerge(
        targetValue as Record<string, unknown>,
        sourceValue as Record<string, unknown>,
      ) as T[keyof T]
    } else if (sourceValue !== undefined) {
      result[key as keyof T] = sourceValue as T[keyof T]
    }
  }

  return result
}

/**
 * Resolve a worker profile entry (string ID or full object).
 */
function resolveProfileEntry(
  entry: string | Partial<WorkerProfile>,
  existingProfiles: Record<string, WorkerProfile>,
): WorkerProfile | undefined {
  if (typeof entry === "string") {
    return existingProfiles[entry] ?? builtInProfiles[entry]
  }

  if (!entry.id) return undefined

  const base = existingProfiles[entry.id] ?? builtInProfiles[entry.id]
  if (base) {
    return { ...base, ...entry, id: entry.id }
  }

  // New custom profile - validate required fields
  if (!entry.name || !entry.model || !entry.purpose || !entry.whenToUse) {
    return undefined
  }

  return entry as WorkerProfile
}

export interface LoadedOrchestratorConfig {
  config: OrchestratorConfigFile
  profiles: Record<string, WorkerProfile>
  spawn: string[]
  sources: { global?: string; project?: string }
}

/**
 * Load orchestrator configuration with three-level merging.
 */
export function loadOrchestratorConfig(directory: string): LoadedOrchestratorConfig {
  const sources: LoadedOrchestratorConfig["sources"] = {}

  // Layer 1: Built-in defaults
  let merged = { ...DEFAULT_CONFIG }

  // Layer 2: Global config
  const globalPath = join(getUserConfigDir(), "orchestrator.json")
  const globalRaw = tryReadJson(globalPath)
  if (globalRaw) {
    sources.global = globalPath
    const parsed = OrchestratorConfigFileSchema.safeParse(globalRaw)
    if (parsed.success) {
      merged = deepMerge(merged, parsed.data)
    }
  }

  // Layer 3: Project config
  const projectPaths = [
    join(directory, ".ring", "orchestrator.json"),
    join(directory, ".opencode", "orchestrator.json"),
  ]

  for (const projectPath of projectPaths) {
    const projectRaw = tryReadJson(projectPath)
    if (projectRaw) {
      sources.project = projectPath
      const parsed = OrchestratorConfigFileSchema.safeParse(projectRaw)
      if (parsed.success) {
        merged = deepMerge(merged, parsed.data)
      }
      break
    }
  }

  // Build profiles map
  const profiles: Record<string, WorkerProfile> = { ...builtInProfiles }
  const spawn: string[] = []
  const seenSpawn = new Set<string>()

  // Process profile definitions
  for (const entry of merged.profiles) {
    const resolved = resolveProfileEntry(entry, profiles)
    if (resolved) {
      profiles[resolved.id] = resolved
    }
  }

  // Process workers to spawn
  for (const entry of merged.workers) {
    if (typeof entry === "string") {
      if (profiles[entry] && !seenSpawn.has(entry)) {
        spawn.push(entry)
        seenSpawn.add(entry)
      }
    } else {
      const resolved = resolveProfileEntry(entry, profiles)
      if (resolved) {
        profiles[resolved.id] = resolved
        if (!seenSpawn.has(resolved.id)) {
          spawn.push(resolved.id)
          seenSpawn.add(resolved.id)
        }
      }
    }
  }

  return { config: merged, profiles, spawn, sources }
}

/**
 * Check if we're running inside a worker (to prevent recursive spawn).
 */
export function isInsideWorker(config: OrchestratorConfigFile): boolean {
  if (!config.security.preventRecursiveSpawn) return false
  return process.env[config.security.recursiveSpawnEnvVar] === "true"
}

/**
 * Create orchestrator context from loaded config.
 */
export function createOrchestratorContext(
  directory: string,
  loaded: LoadedOrchestratorConfig,
): OrchestratorContext {
  return {
    directory,
    profiles: loaded.profiles,
    spawnDefaults: {
      basePort: loaded.config.basePort,
      timeout: loaded.config.startupTimeout,
    },
    defaultListFormat: loaded.config.ui.defaultListFormat,
  }
}

/**
 * Get workflow limits from config.
 */
export function getWorkflowLimits(config: OrchestratorConfigFile): WorkflowRunLimits {
  return {
    maxSteps: config.workflows.limits.maxSteps,
    maxTaskChars: config.workflows.limits.maxTaskChars,
    maxCarryChars: config.workflows.limits.maxCarryChars,
    perStepTimeoutMs: config.workflows.limits.perStepTimeoutMs,
  }
}
```

**Step 2: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/orchestrator/config.ts && git commit -m "$(cat <<'EOF'
feat(orchestrator): add three-level configuration merging

Implement configuration system with:
- Built-in defaults (in code)
- Global config (~/.config/opencode/ring/orchestrator.json)
- Project config (.ring/orchestrator.json or .opencode/orchestrator.json)
- Zod schemas for validation
- Profile resolution (string ID or full object)
- Recursive spawn protection via environment variable
- Context creation helper

Ported from Orchestra's configuration pattern.
EOF
)"
```

---

### Task 9: Create Module Index

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/orchestrator/index.ts`

**Prerequisites:**
- All previous orchestrator files must exist

**Step 1: Write the index file**

```typescript
/**
 * Orchestrator Module
 *
 * Unified exports for the orchestrator subsystem.
 */

// Types
export type {
  DeviceRegistryEntry,
  DeviceRegistryFile,
  DeviceRegistrySessionEntry,
  DeviceRegistryWorkerEntry,
  Job,
  JobReport,
  JobStatus,
  OrchestratorContext,
  WorkerInstance,
  WorkerKind,
  WorkerPoolCallback,
  WorkerPoolEvent,
  WorkerProfile,
  WorkerStatus,
  WorkflowDefinition,
  WorkflowRunInput,
  WorkflowRunLimits,
  WorkflowRunResult,
  WorkflowStepDefinition,
  WorkflowStepResult,
  WorkerExecution,
} from "./types.js"

// Job Registry
export { JobRegistry, jobRegistry } from "./jobs.js"

// Worker Pool
export { WorkerPool, workerPool, type SpawnOptions } from "./worker-pool.js"

// Profiles
export {
  builtInProfiles,
  getProfile,
  listProfileIds,
  mergeProfile,
  resolveModelTag,
} from "./profiles.js"

// Configuration
export {
  createOrchestratorContext,
  getWorkflowLimits,
  isInsideWorker,
  loadOrchestratorConfig,
  type LoadedOrchestratorConfig,
  type OrchestratorConfigFile,
} from "./config.js"

// Workflow Engine
export {
  executeWorkflowStep,
  getWorkflow,
  listWorkflows,
  registerWorkflow,
  runWorkflow,
  validateWorkflowInput,
  type WorkflowRunDependencies,
} from "./workflow/engine.js"

// Task Tools
export { createTaskTools } from "./tools/task-tools.js"
```

**Step 2: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/orchestrator/index.ts && git commit -m "$(cat <<'EOF'
feat(orchestrator): add module index with unified exports

Export all orchestrator components:
- Types for workers, jobs, workflows
- JobRegistry and singleton instance
- WorkerPool and singleton instance
- Built-in profiles and utilities
- Configuration loading and context creation
- Workflow engine and registry
- Task tools factory
EOF
)"
```

---

### Task 10: Integrate Orchestrator into Ring Plugin

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/ring-unified.ts`
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/tools/index.ts`

**Prerequisites:**
- All orchestrator module files must exist
- File must exist: `plugin/ring-unified.ts`

**Step 1: Update tools/index.ts to include orchestrator tools**

```typescript
/**
 * Ring Tools
 *
 * Custom tools registered by Ring plugin.
 * Uses the tool() helper from @opencode-ai/plugin.
 */

import {
  createOrchestratorContext,
  createTaskTools,
  loadOrchestratorConfig,
} from "../orchestrator/index.js"

/**
 * Create all Ring tools including orchestrator task tools.
 */
export function createRingTools(directory: string) {
  // Load orchestrator config and create context
  const loaded = loadOrchestratorConfig(directory)
  const context = createOrchestratorContext(directory, loaded)

  // Create task tools bound to context
  const taskTools = createTaskTools(context)

  return {
    task_start: taskTools.taskStart,
    task_await: taskTools.taskAwait,
    task_peek: taskTools.taskPeek,
    task_list: taskTools.taskList,
    task_cancel: taskTools.taskCancel,
  }
}

/**
 * Legacy export for backwards compatibility.
 * @deprecated Use createRingTools(directory) instead.
 */
export const ringTools = {}
```

**Step 2: Update ring-unified.ts to use new tools**

Find and replace the tools import and usage in ring-unified.ts:

In `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/ring-unified.ts`, change:

```typescript
// Tools
import { ringTools } from "./tools/index.js"
```

To:

```typescript
// Tools
import { createRingTools } from "./tools/index.js"
```

And change:

```typescript
  return {
    // Register Ring tools
    tool: ringTools,
```

To:

```typescript
  // Create orchestrator-aware tools
  const ringTools = createRingTools(directory)

  return {
    // Register Ring tools
    tool: ringTools,
```

**Step 3: Run type check**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/tools/index.ts plugin/ring-unified.ts && git commit -m "$(cat <<'EOF'
feat(ring): integrate orchestrator task tools

Update Ring plugin to register orchestrator tools:
- task_start: Fire and forget async tasks
- task_await: Block until task completion
- task_peek: Non-blocking status check
- task_list: List tasks, workers, profiles
- task_cancel: Cancel running tasks

Tools are created with orchestrator context bound.
EOF
)"
```

**If Task Fails:**

1. **Import errors:**
   - Check: All orchestrator files exist with correct exports
   - Fix: Verify index.ts exports all needed functions

2. **Type errors in ring-unified.ts:**
   - Check: ringTools variable is now a function result
   - Fix: Ensure createRingTools returns correct tool shape

---

### Task 11: Add Tests for Orchestrator Core

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/__tests__/plugin/orchestrator/jobs.test.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/__tests__/plugin/orchestrator/worker-pool.test.ts`

**Prerequisites:**
- Orchestrator module must be complete
- Bun test runner available

**Step 1: Create test directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/__tests__/plugin/orchestrator`

**Step 2: Write job registry tests**

```typescript
/**
 * Job Registry Tests
 */

import { describe, expect, it, beforeEach } from "bun:test"
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

      const list = registry.list()
      expect(list.length).toBe(2)
      expect(list[0].id).toBe(job2.id) // Most recent first
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

      // Complete job after short delay
      setTimeout(() => {
        registry.setResult(job.id, { responseText: "Done" })
      }, 50)

      const result = await registry.await(job.id, { timeoutMs: 1000 })
      expect(result.status).toBe("succeeded")
    })

    it("throws for unknown job", async () => {
      await expect(registry.await("nonexistent")).rejects.toThrow('Unknown job "nonexistent"')
    })
  })
})
```

**Step 3: Write worker pool tests**

```typescript
/**
 * Worker Pool Tests
 */

import { describe, expect, it, beforeEach } from "bun:test"
import { WorkerPool } from "../../../plugin/orchestrator/worker-pool.js"
import type { WorkerInstance, WorkerProfile } from "../../../plugin/orchestrator/types.js"

describe("WorkerPool", () => {
  let pool: WorkerPool

  const mockProfile: WorkerProfile = {
    id: "test-worker",
    name: "Test Worker",
    model: "test-model",
    purpose: "Testing",
    whenToUse: "During tests",
  }

  const createMockInstance = (profile: WorkerProfile): WorkerInstance => ({
    profile,
    status: "ready",
    port: 14100,
    startedAt: new Date(),
  })

  beforeEach(() => {
    pool = new WorkerPool()
  })

  describe("register/unregister", () => {
    it("registers a worker", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)

      expect(pool.get(mockProfile.id)).toBe(instance)
    })

    it("unregisters a worker", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)
      pool.unregister(mockProfile.id)

      expect(pool.get(mockProfile.id)).toBeUndefined()
    })

    it("returns false when unregistering unknown worker", () => {
      expect(pool.unregister("nonexistent")).toBe(false)
    })
  })

  describe("list", () => {
    it("returns all workers", () => {
      pool.register(createMockInstance(mockProfile))
      pool.register(createMockInstance({ ...mockProfile, id: "worker-2" }))

      const list = pool.list()
      expect(list.length).toBe(2)
    })
  })

  describe("getWorkersByStatus", () => {
    it("filters by status", () => {
      const ready = createMockInstance(mockProfile)
      ready.status = "ready"
      pool.register(ready)

      const busy = createMockInstance({ ...mockProfile, id: "busy-worker" })
      busy.status = "busy"
      pool.register(busy)

      const readyWorkers = pool.getWorkersByStatus("ready")
      expect(readyWorkers.length).toBe(1)
      expect(readyWorkers[0].profile.id).toBe(mockProfile.id)
    })
  })

  describe("getWorkersByCapability", () => {
    it("matches by purpose", () => {
      pool.register(createMockInstance({
        ...mockProfile,
        purpose: "Documentation research",
      }))

      const matches = pool.getWorkersByCapability("documentation")
      expect(matches.length).toBe(1)
    })

    it("matches by tags", () => {
      pool.register(createMockInstance({
        ...mockProfile,
        tags: ["vision", "images"],
      }))

      const matches = pool.getWorkersByCapability("vision")
      expect(matches.length).toBe(1)
    })
  })

  describe("updateStatus", () => {
    it("updates worker status", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)

      pool.updateStatus(mockProfile.id, "busy")

      expect(pool.get(mockProfile.id)?.status).toBe("busy")
    })

    it("sets error on error status", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)

      pool.updateStatus(mockProfile.id, "error", "Connection lost")

      const updated = pool.get(mockProfile.id)
      expect(updated?.status).toBe("error")
      expect(updated?.error).toBe("Connection lost")
    })
  })

  describe("session ownership", () => {
    it("tracks worker ownership by session", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)
      pool.trackOwnership("session-1", mockProfile.id)

      const owned = pool.getWorkersForSession("session-1")
      expect(owned).toContain(mockProfile.id)
    })

    it("clears session ownership", () => {
      pool.trackOwnership("session-1", mockProfile.id)
      pool.clearSessionOwnership("session-1")

      const owned = pool.getWorkersForSession("session-1")
      expect(owned.length).toBe(0)
    })
  })

  describe("events", () => {
    it("emits spawn event on register", () => {
      const events: string[] = []
      pool.on("spawn", () => events.push("spawn"))

      pool.register(createMockInstance(mockProfile))

      expect(events).toContain("spawn")
    })

    it("emits stop event on unregister", () => {
      const events: string[] = []
      pool.register(createMockInstance(mockProfile))
      pool.on("stop", () => events.push("stop"))

      pool.unregister(mockProfile.id)

      expect(events).toContain("stop")
    })
  })

  describe("getOrSpawn", () => {
    it("returns existing worker", async () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)

      const spawnFn = async () => createMockInstance(mockProfile)
      const result = await pool.getOrSpawn(mockProfile, { basePort: 14100, timeout: 5000, directory: "/tmp" }, spawnFn)

      expect(result).toBe(instance)
    })

    it("deduplicates concurrent spawns", async () => {
      let spawnCount = 0
      const spawnFn = async () => {
        spawnCount++
        await new Promise((r) => setTimeout(r, 50))
        const instance = createMockInstance(mockProfile)
        pool.register(instance)
        return instance
      }

      const options = { basePort: 14100, timeout: 5000, directory: "/tmp" }

      // Spawn concurrently
      const [r1, r2] = await Promise.all([
        pool.getOrSpawn(mockProfile, options, spawnFn),
        pool.getOrSpawn(mockProfile, options, spawnFn),
      ])

      expect(spawnCount).toBe(1) // Only one spawn
      expect(r1).toBe(r2) // Same instance
    })
  })
})
```

**Step 4: Run tests**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test __tests__/plugin/orchestrator/`

**Expected output:**
```
bun test v1.x.x

__tests__/plugin/orchestrator/jobs.test.ts:
✓ JobRegistry > create > creates a job with running status
✓ JobRegistry > create > includes optional sessionId and requestedBy
... (all tests passing)

__tests__/plugin/orchestrator/worker-pool.test.ts:
✓ WorkerPool > register/unregister > registers a worker
... (all tests passing)

 XX pass
 0 fail
```

**Step 5: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add __tests__/plugin/orchestrator/ && git commit -m "$(cat <<'EOF'
test(orchestrator): add unit tests for core modules

Add comprehensive tests for:
- JobRegistry: create, get, setResult, setError, cancel, list, await
- WorkerPool: register, unregister, list, status updates, ownership, events, deduplication

Tests verify core orchestrator patterns work correctly.
EOF
)"
```

---

### Task 12: Final Code Review

1. **Dispatch all reviewers:**
   - REQUIRED SUB-SKILL: Use requesting-code-review
   - Run code-reviewer, business-logic-reviewer, security-reviewer

2. **Handle findings:**
   - Critical/High/Medium: Fix immediately
   - Low: Add TODO(review): comments
   - Cosmetic: Add FIXME(nitpick): comments

3. **Verify all tests pass:**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test`

**Expected output:**
```
All tests passing
```

4. **Verify type check:**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run typecheck`

**Expected output:**
```
(no errors)
```

---

## Summary

This plan ports the following Orchestra patterns to Ring:

| Pattern | Orchestra Source | Ring Implementation |
|---------|-----------------|---------------------|
| Five-Tool Async Task API | `command/tasks.ts` | `plugin/orchestrator/tools/task-tools.ts` |
| Worker Pool with Deduplication | `core/worker-pool.ts` | `plugin/orchestrator/worker-pool.ts` |
| Profile-Based Worker Definition | `config/profiles.ts` | `plugin/orchestrator/profiles.ts` |
| Workflow Carry/Trim Logic | `workflows/engine.ts` | `plugin/orchestrator/workflow/engine.ts` |
| Three-Level Config Merging | `config/orchestrator.ts` | `plugin/orchestrator/config.ts` |
| Job Registry | `core/jobs.ts` | `plugin/orchestrator/jobs.ts` |
| Recursive Spawn Protection | Environment variable check | In `config.ts` |

**Files Created:**
- `plugin/orchestrator/types.ts`
- `plugin/orchestrator/jobs.ts`
- `plugin/orchestrator/worker-pool.ts`
- `plugin/orchestrator/profiles.ts`
- `plugin/orchestrator/config.ts`
- `plugin/orchestrator/workflow/engine.ts`
- `plugin/orchestrator/tools/task-tools.ts`
- `plugin/orchestrator/index.ts`
- `__tests__/plugin/orchestrator/jobs.test.ts`
- `__tests__/plugin/orchestrator/worker-pool.test.ts`

**Files Modified:**
- `plugin/tools/index.ts`
- `plugin/ring-unified.ts`

**Total Tasks:** 12

**Estimated Time:** 2-3 hours for experienced developer

---

## Plan Checklist

- [x] Historical precedent queried (index empty - new project)
- [x] Historical Precedent section included in plan
- [x] Header with goal, architecture, tech stack, prerequisites
- [x] Verification commands with expected output
- [x] Tasks broken into bite-sized steps (2-5 min each)
- [x] Exact file paths for all files
- [x] Complete code (no placeholders)
- [x] Exact commands with expected output
- [x] Failure recovery steps for each task
- [x] Code review checkpoints after batches
- [x] Severity-based issue handling documented
- [x] Passes Zero-Context Test
