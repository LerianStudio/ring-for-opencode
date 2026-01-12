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

import { existsSync, lstatSync, realpathSync } from "node:fs"
import { isAbsolute, resolve, sep } from "node:path"
import type { ToolDefinition } from "@opencode-ai/plugin"
import { tool } from "@opencode-ai/plugin/tool"
import { jobRegistry } from "../jobs.js"
import { builtInProfiles } from "../profiles.js"
import type {
  Job,
  OrchestratorContext,
  TaskDispatchResult,
  WorkerInstance,
  WorkerProfile,
  WorkerSpawnOptions,
} from "../types.js"
import { MAX_ATTACHMENTS, MAX_BASE64_LENGTH, MAX_TASK_LENGTH } from "../types.js"
import { workerPool } from "../worker-pool.js"

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

const DEFAULT_TIMEOUT_MS = 600_000
const MAX_TIMEOUT_MS = 600_000
const MIN_TIMEOUT_MS = 1000

function normalizeTimeout(timeoutMs?: number): number {
  if (typeof timeoutMs !== "number" || !Number.isFinite(timeoutMs)) {
    return DEFAULT_TIMEOUT_MS
  }
  if (timeoutMs <= 0) {
    return 0
  }
  return Math.max(MIN_TIMEOUT_MS, Math.min(timeoutMs, MAX_TIMEOUT_MS))
}

function formatError(message: string): string {
  return JSON.stringify({ error: message }, null, 2)
}

function isJobVisible(job: Job, sessionId?: string): boolean {
  return Boolean(sessionId && job.sessionId === sessionId)
}

async function withTimeout<T>(promise: Promise<T>, timeoutMs: number, label: string): Promise<T> {
  let timer: ReturnType<typeof setTimeout> | undefined
  const timeoutPromise = new Promise<never>((_, reject) => {
    timer = setTimeout(() => reject(new Error(label)), timeoutMs)
  })

  try {
    return await Promise.race([promise, timeoutPromise])
  } finally {
    if (timer) clearTimeout(timer)
  }
}

function validateAttachmentPath(path: string, projectRoot: string): string {
  if (isAbsolute(path)) {
    throw new Error("Attachment path must be relative")
  }

  if (path.includes("..")) {
    throw new Error("Path traversal not allowed in attachment paths")
  }

  const resolved = resolve(projectRoot, path)

  if (!existsSync(resolved)) {
    throw new Error(`Attachment file not found: ${path}`)
  }

  const stat = lstatSync(resolved)
  const realRoot = `${realpathSync(projectRoot)}${sep}`
  const realResolved = realpathSync(resolved)

  if (!realResolved.startsWith(realRoot)) {
    throw new Error("Attachment path must be within project directory")
  }

  if (stat.isSymbolicLink()) {
    throw new Error("Symlink attachments are not allowed")
  }

  return realResolved
}

function guessWorkerId(task: string, attachments?: ToolAttachment[]): string {
  if (attachments?.some((a) => a.type === "image")) return "vision"

  const docsRegex = /\b(doc|docs|documentation|reference|api|example|examples|research|cite)\b/i
  const architectRegex = /\b(architecture|architect|design|plan|approach|trade[- ]?off)\b/i
  const explorerRegex = /\b(search|find|locate|grep|ripgrep|scan|explore|where)\b/i

  if (docsRegex.test(task)) {
    return "docs"
  }
  if (architectRegex.test(task)) {
    return "architect"
  }
  if (explorerRegex.test(task)) {
    return "explorer"
  }

  return "coder"
}

function renderMarkdownTable(headers: string[], rows: string[][]): string {
  if (rows.length === 0) return "(no data)"

  const widths = headers.map((h, i) => Math.max(h.length, ...rows.map((r) => (r[i] ?? "").length)))

  const header = `| ${headers.map((h, i) => h.padEnd(widths[i])).join(" | ")} |`
  const separator = `| ${widths.map((w) => "-".repeat(w)).join(" | ")} |`
  const body = rows
    .map((r) => `| ${r.map((c, i) => (c ?? "").padEnd(widths[i])).join(" | ")} |`)
    .join("\n")

  return `${header}\n${separator}\n${body}`
}

function resolveProfile(workerId: string, context: OrchestratorContext): WorkerProfile | undefined {
  return context.profiles[workerId] ?? builtInProfiles[workerId]
}

async function resolveWorkerInstance(input: {
  profile: WorkerProfile
  context: OrchestratorContext
  sessionId?: string
  autoSpawn: boolean
}): Promise<WorkerInstance> {
  const existing = workerPool.get(input.profile.id)
  if (existing && existing.status !== "error" && existing.status !== "stopped") {
    workerPool.trackOwnership(input.sessionId, input.profile.id)
    return existing
  }

  if (!input.autoSpawn) {
    const statusLabel = existing ? existing.status : "not running"
    throw new Error(
      `Worker "${input.profile.id}" is ${statusLabel}. Set autoSpawn=true or spawn it first.`,
    )
  }

  if (!input.context.spawnWorker) {
    throw new Error(`Auto-spawn not configured for worker "${input.profile.id}".`)
  }

  const spawnOptions: WorkerSpawnOptions = {
    basePort: input.context.spawnDefaults.basePort,
    timeout: input.context.spawnDefaults.timeout,
    directory: input.context.directory,
    parentSessionId: input.sessionId,
  }

  const instance = await workerPool.getOrSpawn(
    input.profile,
    spawnOptions,
    input.context.spawnWorker,
  )
  workerPool.trackOwnership(input.sessionId, input.profile.id)
  return instance
}

async function dispatchTask(input: {
  context: OrchestratorContext
  worker: WorkerInstance
  task: string
  attachments?: ToolAttachment[]
  timeoutMs: number
  sessionId?: string
  requestedBy?: string
}): Promise<TaskDispatchResult> {
  if (!input.context.dispatchTask) {
    return { error: "Task dispatch is not configured" }
  }

  const result = await withTimeout(
    input.context.dispatchTask({
      worker: input.worker,
      workerId: input.worker.profile.id,
      task: input.task,
      attachments: input.attachments,
      timeoutMs: input.timeoutMs,
      sessionId: input.sessionId,
      requestedBy: input.requestedBy,
    }),
    input.timeoutMs,
    `Task timed out after ${input.timeoutMs}ms`,
  )

  return result
}

export function createTaskTools(context: OrchestratorContext): TaskTools {
  const taskStart: ToolDefinition = tool({
    description:
      "Start a background task. Returns a taskId immediately; use task_await to get the result.",
    args: {
      kind: tool.schema
        .enum(["auto", "worker"])
        .optional()
        .describe("Task kind (default: auto = pick a worker based on task/attachments)"),
      task: tool.schema
        .string()
        .max(MAX_TASK_LENGTH, `Task must be <= ${MAX_TASK_LENGTH} chars`)
        .describe("What to do (sent to worker)"),
      workerId: tool.schema.string().optional().describe("Worker id (e.g. 'docs', 'coder')"),
      attachments: tool.schema
        .array(
          tool.schema.object({
            type: tool.schema.enum(["image", "file"]),
            path: tool.schema
              .string()
              .refine((p) => !p.includes(".."), "Path must not contain ..")
              .optional(),
            base64: tool.schema.string().max(MAX_BASE64_LENGTH, "Base64 too large").optional(),
            mimeType: tool.schema.string().optional(),
          }),
        )
        .max(MAX_ATTACHMENTS, `Max ${MAX_ATTACHMENTS} attachments`)
        .optional()
        .describe("Optional attachments (images/files) to forward to the worker"),
      autoSpawn: tool.schema
        .boolean()
        .optional()
        .describe("Auto-spawn missing workers (default: true)"),
      timeoutMs: tool.schema
        .number()
        .min(1000, "Timeout must be >= 1000ms")
        .optional()
        .describe("Timeout for the task (default: 10 minutes)"),
    },
    async execute(args, ctx) {
      const kind = args.kind ?? "auto"
      const autoSpawn = args.autoSpawn ?? true
      const sessionId = ctx?.sessionID
      if (!sessionId) {
        return formatError("Session ID required for task tools.")
      }
      const requestedBy = ctx?.agent
      const timeoutMs = normalizeTimeout(args.timeoutMs)

      const resolvedWorkerId =
        kind === "auto"
          ? (args.workerId ?? guessWorkerId(args.task, args.attachments))
          : args.workerId

      if (!resolvedWorkerId) {
        return formatError("Missing workerId")
      }

      let sanitizedAttachments = args.attachments
      if (args.attachments) {
        const next: ToolAttachment[] = []
        for (const attachment of args.attachments) {
          if (attachment.path) {
            try {
              const resolvedPath = validateAttachmentPath(attachment.path, context.directory)
              next.push({ ...attachment, path: resolvedPath })
            } catch (err) {
              return formatError(
                `Invalid attachment path: ${err instanceof Error ? err.message : String(err)}`,
              )
            }
          } else {
            next.push(attachment)
          }
        }
        sanitizedAttachments = next
      }

      const job = jobRegistry.create({
        workerId: resolvedWorkerId,
        message: args.task,
        sessionId,
        requestedBy,
      })

      const profile = resolveProfile(resolvedWorkerId, context)
      if (!profile) {
        const error = `Unknown worker "${resolvedWorkerId}"`
        jobRegistry.setError(job.id, { error })
        return JSON.stringify(
          {
            taskId: job.id,
            workerId: resolvedWorkerId,
            status: "failed",
            error,
          },
          null,
          2,
        )
      }

      if (!context.dispatchTask) {
        const error = "Task dispatch is not configured"
        jobRegistry.setError(job.id, { error })
        return JSON.stringify(
          {
            taskId: job.id,
            workerId: resolvedWorkerId,
            status: "failed",
            error,
          },
          null,
          2,
        )
      }

      const existing = workerPool.get(resolvedWorkerId)
      const usable = existing && existing.status !== "error" && existing.status !== "stopped"

      if (!usable && !autoSpawn) {
        const statusLabel = existing ? existing.status : "not running"
        const error = `Worker "${resolvedWorkerId}" is ${statusLabel}. Set autoSpawn=true or spawn it first.`
        jobRegistry.setError(job.id, { error })
        return JSON.stringify(
          {
            taskId: job.id,
            workerId: resolvedWorkerId,
            status: "failed",
            error,
          },
          null,
          2,
        )
      }

      if (!usable && autoSpawn && !context.spawnWorker) {
        const error = `Auto-spawn not configured for worker "${resolvedWorkerId}".`
        jobRegistry.setError(job.id, { error })
        return JSON.stringify(
          {
            taskId: job.id,
            workerId: resolvedWorkerId,
            status: "failed",
            error,
          },
          null,
          2,
        )
      }

      void (async () => {
        try {
          const worker = await resolveWorkerInstance({
            profile,
            context,
            sessionId,
            autoSpawn,
          })
          const result = await dispatchTask({
            context,
            worker,
            task: args.task,
            attachments: sanitizedAttachments,
            timeoutMs,
            sessionId,
            requestedBy,
          })

          if (result.error) {
            jobRegistry.setError(job.id, { error: result.error })
            return
          }

          jobRegistry.setResult(job.id, { responseText: result.responseText ?? "" })
          if (result.warning) {
            jobRegistry.attachReport(job.id, { notes: result.warning })
          }
        } catch (err) {
          jobRegistry.setError(job.id, { error: err instanceof Error ? err.message : String(err) })
        }
      })()

      return JSON.stringify(
        {
          taskId: job.id,
          kind,
          workerId: resolvedWorkerId,
          status: "running",
          next: "task_await",
        },
        null,
        2,
      )
    },
  })

  const taskAwait: ToolDefinition = tool({
    description: "Wait for one or more tasks to finish and return the final job record(s).",
    args: {
      taskId: tool.schema.string().optional().describe("Task id from task_start"),
      taskIds: tool.schema
        .array(tool.schema.string())
        .optional()
        .describe("Multiple task ids to await"),
      timeoutMs: tool.schema.number().optional().describe("Timeout in ms (default: 10 minutes)"),
    },
    async execute(args, ctx) {
      const timeoutMs = normalizeTimeout(args.timeoutMs)
      const ids = args.taskId ? [args.taskId] : (args.taskIds ?? [])
      const sessionId = ctx?.sessionID

      if (!sessionId) {
        return formatError("Session ID required for task tools.")
      }

      if (ids.length === 0) {
        return formatError("Missing taskId/taskIds.")
      }

      const results = await Promise.all(
        ids.map(async (id) => {
          const job = jobRegistry.get(id)
          if (!job) {
            return { id, status: "unknown" }
          }
          if (!isJobVisible(job, sessionId)) {
            return { id, status: "forbidden" }
          }
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
    async execute(args, ctx) {
      const ids = args.taskId ? [args.taskId] : (args.taskIds ?? [])
      const sessionId = ctx?.sessionID

      if (!sessionId) {
        return formatError("Session ID required for task tools.")
      }

      if (ids.length === 0) {
        return formatError("Missing taskId/taskIds.")
      }

      const results = ids.map((id) => {
        const job = jobRegistry.get(id)
        if (!job) return { id, status: "unknown" }
        if (!isJobVisible(job, sessionId)) return { id, status: "forbidden" }
        return job
      })
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
      format: tool.schema
        .enum(["markdown", "json"])
        .optional()
        .describe("Output format (default: markdown)"),
    },
    async execute(args, ctx) {
      const format = args.format ?? context.defaultListFormat
      const view = args.view ?? "tasks"
      const sessionId = ctx?.sessionID

      if (!sessionId) {
        return formatError("Session ID required for task tools.")
      }

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
        const limit = Math.max(1, args.limit ?? 20)
        const tasks = jobRegistry
          .list({ limit: Number.MAX_SAFE_INTEGER })
          .filter((job) => isJobVisible(job, sessionId))
          .slice(0, limit)
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
            ? renderMarkdownTable(
                ["Task", "Worker", "Status", "Started", "ms", "Message"],
                taskRows,
              )
            : "(none)",
        ].join("\n")
      }

      const limit = args.limit ?? 20
      const tasks = jobRegistry
        .list({ workerId: args.workerId, limit: Number.MAX_SAFE_INTEGER })
        .filter((job) => isJobVisible(job, sessionId))
        .slice(0, limit)

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
    async execute(args, ctx) {
      const ids = args.taskId ? [args.taskId] : (args.taskIds ?? [])
      const sessionId = ctx?.sessionID

      if (!sessionId) {
        return formatError("Session ID required for task tools.")
      }

      if (ids.length === 0) {
        return formatError("Missing taskId/taskIds.")
      }

      const results = ids.map((id) => {
        const job = jobRegistry.get(id)
        if (!job) return { id, status: "unknown" }
        if (!isJobVisible(job, sessionId)) return { id, status: "forbidden" }

        jobRegistry.cancel(id, { reason: args.reason })
        return jobRegistry.get(id) ?? { id, status: "unknown" }
      })

      return JSON.stringify(ids.length === 1 ? results[0] : results, null, 2)
    },
  })

  return { taskStart, taskAwait, taskPeek, taskList, taskCancel }
}
