/**
 * Ring Tools
 *
 * Custom tools registered by Ring plugin.
 * Uses the tool() helper from @opencode-ai/plugin.
 */

import type { BackgroundManager } from "../background/index.js"
import type { BackgroundTask } from "../background/types.js"
import type {
  TaskDispatchInput,
  TaskDispatchResult,
  WorkerInstance,
  WorkerProfile,
  WorkerSpawnOptions,
} from "../orchestrator/index.js"
import {
  createOrchestratorContext,
  createTaskTools,
  loadOrchestratorConfig,
} from "../orchestrator/index.js"

type RingToolOptions = {
  backgroundManager?: BackgroundManager
}

const DEFAULT_POLL_INTERVAL_MS = 1_000
const EMPTY_RESPONSE_WARNING = "Background task completed without captured response."

function formatAttachmentSummary(attachments?: TaskDispatchInput["attachments"]): string[] {
  if (!attachments || attachments.length === 0) return []

  return attachments.map((attachment, index) => {
    const parts: string[] = []
    if (attachment.path) parts.push(`path=${attachment.path}`)
    if (attachment.mimeType) parts.push(`mime=${attachment.mimeType}`)
    if (attachment.base64) parts.push(`base64=${attachment.base64.length} chars`)
    const detail = parts.length ? ` (${parts.join(", ")})` : ""
    return `- Attachment ${index + 1}: ${attachment.type}${detail}`
  })
}

function buildPrompt(task: string, attachments?: TaskDispatchInput["attachments"]): string {
  const summaries = formatAttachmentSummary(attachments)
  if (summaries.length === 0) return task
  return [task, "", "Attachments:", ...summaries].join("\n")
}

function parseModelOverride(model?: string): { providerId: string; modelId: string } | undefined {
  if (!model) return undefined

  const separator = model.includes(":") ? ":" : model.includes("/") ? "/" : null
  if (!separator) return undefined

  const [providerId, modelId] = model.split(separator, 2)
  if (!providerId || !modelId) return undefined

  return { providerId, modelId }
}

function resolveTaskResult(task: BackgroundTask): TaskDispatchResult {
  if (task.status === "completed") {
    const responseText = task.result ?? task.progress?.lastMessage ?? ""
    if (!responseText) {
      return { responseText, warning: EMPTY_RESPONSE_WARNING }
    }
    return { responseText }
  }

  if (task.status === "error" || task.status === "timeout" || task.status === "cancelled") {
    return { error: task.error ?? `Task ${task.status}` }
  }

  return { error: "Task did not complete" }
}

async function waitForBackgroundTask(
  backgroundManager: BackgroundManager,
  taskId: string,
  timeoutMs: number,
): Promise<TaskDispatchResult> {
  const start = Date.now()
  const pollInterval = Math.max(
    500,
    Math.min(DEFAULT_POLL_INTERVAL_MS, Math.floor(timeoutMs / 10) || DEFAULT_POLL_INTERVAL_MS),
  )

  while (Date.now() - start < timeoutMs) {
    const task = backgroundManager.getTask(taskId)
    if (!task) {
      return { error: `Background task "${taskId}" not found` }
    }

    if (task.status !== "running" && task.status !== "pending") {
      return resolveTaskResult(task)
    }

    await new Promise((resolve) => setTimeout(resolve, pollInterval))
  }

  return { error: `Task timed out after ${timeoutMs}ms` }
}

async function dispatchBackgroundTask(
  backgroundManager: BackgroundManager,
  input: TaskDispatchInput,
): Promise<TaskDispatchResult> {
  if (!input.sessionId) {
    return { error: "Session ID required to dispatch background task." }
  }

  const profile = input.worker.profile
  const description = `${profile.name ?? profile.id} task`
  const prompt = buildPrompt(input.task, input.attachments)
  const model = parseModelOverride(profile.model)

  const task = await backgroundManager.launch({
    description,
    prompt,
    agent: profile.id,
    parentSessionId: input.sessionId,
    model,
    taskType: "custom",
    skillContent: profile.systemPrompt,
  })

  return await waitForBackgroundTask(backgroundManager, task.id, input.timeoutMs)
}

async function spawnBackgroundWorker(
  profile: WorkerProfile,
  options: WorkerSpawnOptions,
): Promise<WorkerInstance> {
  return {
    profile,
    kind: profile.kind,
    execution: "background",
    status: "ready",
    port: options.basePort,
    startedAt: new Date(),
    parentSessionId: options.parentSessionId,
    directory: options.directory,
  }
}

/**
 * Create all Ring tools including orchestrator task tools.
 */
export function createRingTools(directory: string, options: RingToolOptions = {}) {
  const loaded = loadOrchestratorConfig(directory)

  const context = createOrchestratorContext(directory, loaded)

  const backgroundManager = options.backgroundManager
  if (backgroundManager) {
    context.dispatchTask = (input) => dispatchBackgroundTask(backgroundManager, input)
    context.spawnWorker = spawnBackgroundWorker
  }

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
