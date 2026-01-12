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
  /** Cryptographically random auth token for worker verification */
  authToken?: string
  /** Token expiration timestamp */
  tokenExpiry?: number
  /** Allowed origins for serverUrl (SSRF protection) */
  allowedOrigins?: string[]
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
// Task Dispatch Types
// =============================================================================

export interface WorkerSpawnOptions {
  basePort: number
  timeout: number
  directory: string
  parentSessionId?: string
  forceNew?: boolean
}

export interface TaskDispatchInput {
  worker: WorkerInstance
  workerId: string
  task: string
  attachments?: WorkflowRunInput["attachments"]
  timeoutMs: number
  sessionId?: string
  requestedBy?: string
}

export interface TaskDispatchResult {
  responseText?: string
  warning?: string
  error?: string
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
  dispatchTask?: (input: TaskDispatchInput) => Promise<TaskDispatchResult>
  spawnWorker?: (profile: WorkerProfile, options: WorkerSpawnOptions) => Promise<WorkerInstance>
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

// =============================================================================
// Security Constants
// =============================================================================

export const BLOCKED_HOSTS = [
  "169.254.169.254", // AWS metadata
  "metadata.google.internal", // GCP metadata
  "localhost",
  "127.0.0.1",
  "0.0.0.0",
  "::1",
] as const

export const MAX_WORKERS = 50
export const MAX_WORKERS_PER_SESSION = 10
export const SPAWN_RATE_LIMIT_MS = 1000
export const MAX_TASK_LENGTH = 100_000
export const MAX_BASE64_LENGTH = 10_000_000
export const MAX_ATTACHMENTS = 10
