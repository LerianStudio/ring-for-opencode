/**
 * Ring Background Manager
 *
 * Manages background task execution, tracking, and notifications.
 * Coordinates with OpenCode sessions for async agent execution.
 */

import { randomBytes } from "node:crypto"
import { AgentNameSchema, type BackgroundTaskConfig } from "../config"
import { ConcurrencyManager } from "./concurrency"
import type { BackgroundClient, BackgroundTask, LaunchTaskInput } from "./types"

/**
 * Default polling interval for task status checks (10 seconds).
 */
const DEFAULT_POLL_INTERVAL_MS = 10_000

/**
 * Manages background task lifecycle.
 *
 * Features:
 * - Task launching with concurrency control
 * - Status polling and progress tracking
 * - Notification queuing for parent sessions
 * - Timeout handling
 */
export class BackgroundManager {
  private tasks: Map<string, BackgroundTask> = new Map()
  private pendingByParent: Map<string, Set<string>> = new Map()
  private pendingNotifications: Map<string, Set<string>> = new Map()
  private sessionToTask: Map<string, string> = new Map()
  private concurrency: ConcurrencyManager
  private config: BackgroundTaskConfig
  private client: BackgroundClient
  // M2: Removed unused 'directory' field
  private pollInterval: ReturnType<typeof setInterval> | null = null
  private isPolling = false

  /**
   * Creates a new BackgroundManager.
   *
   * @param client - OpenCode client for session management
   * @param _directory - Working directory for tasks (reserved for future use)
   * @param config - Background task configuration
   */
  constructor(client: BackgroundClient, _directory: string, config?: BackgroundTaskConfig) {
    this.client = client
    // M2: directory parameter kept for API compatibility but not stored
    this.config = config ?? {
      defaultConcurrency: 3,
      taskTimeoutMs: 1800000,
    }
    this.concurrency = new ConcurrencyManager(this.config)
  }

  /**
   * Launches a new background task.
   *
   * @param input - Task launch parameters
   * @returns The created background task
   * @throws Error if agent name is invalid
   */
  async launch(input: LaunchTaskInput): Promise<BackgroundTask> {
    // M1: Validate agent name before proceeding
    const agentResult = AgentNameSchema.safeParse(input.agent)
    if (!agentResult.success) {
      throw new Error(`Invalid agent name: ${input.agent}`)
    }

    const taskId = `bg_${randomBytes(4).toString("hex")}`
    const concurrencyKey = input.agent

    // Acquire concurrency slot (may wait)
    await this.concurrency.acquire(concurrencyKey)

    // Create OpenCode session
    const sessionResult = await this.client.session.create({
      body: {
        parentID: input.parentSessionId,
        title: `[Ring] ${input.description}`,
      },
    })

    if (sessionResult.error || !sessionResult.data?.id) {
      this.concurrency.release(concurrencyKey)
      throw new Error(`Failed to create session: ${JSON.stringify(sessionResult.error)}`)
    }

    const sessionId = sessionResult.data.id

    // Build system prompt with skill content if provided
    let systemPrompt: string | undefined
    if (input.skillContent) {
      systemPrompt = input.skillContent
    }

    // Create task record
    const task: BackgroundTask = {
      id: taskId,
      sessionId,
      parentSessionId: input.parentSessionId,
      description: input.description,
      prompt: input.prompt,
      agent: input.agent,
      status: "pending",
      createdAt: new Date(),
      model: input.model,
      concurrencyKey,
      taskType: input.taskType ?? "custom",
      progress: {
        toolCalls: 0,
        lastUpdate: new Date(),
      },
    }

    // Store task
    this.tasks.set(taskId, task)
    this.sessionToTask.set(sessionId, taskId)
    this.trackPending(input.parentSessionId, taskId)

    // Send prompt to session
    try {
      await this.client.session.prompt({
        path: { id: sessionId },
        body: {
          agent: input.agent,
          model: input.model,
          system: systemPrompt,
          parts: [{ type: "text", text: input.prompt }],
        },
      })

      // Update status to running
      task.status = "running"
      task.startedAt = new Date()

      // Start polling if not already
      this.startPolling()

      // Show toast notification
      await this.client.tui.showToast({
        body: {
          title: "Background Task Started",
          message: `${input.description} (${taskId})`,
          variant: "info",
          duration: 3000,
        },
      })
    } catch (error) {
      await this.handleTaskError(task, error instanceof Error ? error : new Error(String(error)))
      throw error
    }

    return task
  }

  /**
   * Gets a task by ID.
   *
   * @param id - Task ID
   * @returns The task or undefined
   */
  getTask(id: string): BackgroundTask | undefined {
    return this.tasks.get(id)
  }

  /**
   * Gets all tasks for a parent session.
   *
   * @param parentSessionId - Parent session ID
   * @returns Array of tasks
   */
  getTasksByParent(parentSessionId: string): BackgroundTask[] {
    const taskIds = this.pendingByParent.get(parentSessionId)
    if (!taskIds) {
      return []
    }
    const result: BackgroundTask[] = []
    for (const id of taskIds) {
      const task = this.tasks.get(id)
      if (task) {
        result.push(task)
      }
    }
    return result
  }

  /**
   * Gets all running tasks.
   *
   * @returns Array of running tasks
   */
  getRunningTasks(): BackgroundTask[] {
    const result: BackgroundTask[] = []
    for (const task of this.tasks.values()) {
      if (task.status === "running" || task.status === "pending") {
        result.push(task)
      }
    }
    return result
  }

  /**
   * Gets all completed tasks.
   *
   * @returns Array of completed tasks
   */
  getCompletedTasks(): BackgroundTask[] {
    const result: BackgroundTask[] = []
    for (const task of this.tasks.values()) {
      if (
        task.status === "completed" ||
        task.status === "error" ||
        task.status === "cancelled" ||
        task.status === "timeout"
      ) {
        result.push(task)
      }
    }
    return result
  }

  /**
   * Finds a task by its session ID.
   *
   * @param sessionId - OpenCode session ID
   * @returns The task or undefined
   */
  findBySession(sessionId: string): BackgroundTask | undefined {
    const taskId = this.sessionToTask.get(sessionId)
    if (!taskId) {
      return undefined
    }
    return this.tasks.get(taskId)
  }

  /**
   * Handles OpenCode events for task progress tracking.
   *
   * @param event - Event from OpenCode
   */
  handleEvent(event: { type: string; properties?: Record<string, unknown> }): void {
    const sessionId = event.properties?.sessionID as string | undefined
    if (!sessionId) {
      return
    }

    const task = this.findBySession(sessionId)
    if (!task) {
      return
    }

    // Update progress based on event type
    if (event.type === "assistant.tool.call") {
      if (task.progress) {
        task.progress.toolCalls++
        task.progress.lastTool = event.properties?.toolName as string | undefined
        task.progress.lastUpdate = new Date()
      }
    } else if (event.type === "assistant.message") {
      if (task.progress) {
        task.progress.lastMessage = event.properties?.text as string | undefined
        task.progress.lastMessageAt = new Date()
        task.progress.lastUpdate = new Date()
      }
    }
  }

  /**
   * Gets pending notifications for a session.
   *
   * @param sessionId - Parent session ID
   * @returns Array of tasks needing notification
   */
  getPendingNotifications(sessionId: string): BackgroundTask[] {
    const taskIds = this.pendingNotifications.get(sessionId)
    if (!taskIds) {
      return []
    }
    const result: BackgroundTask[] = []
    for (const id of taskIds) {
      const task = this.tasks.get(id)
      if (task) {
        result.push(task)
      }
    }
    return result
  }

  /**
   * Clears notifications for a session.
   *
   * @param sessionId - Parent session ID
   */
  clearNotifications(sessionId: string): void {
    this.pendingNotifications.delete(sessionId)
  }

  /**
   * Cleans up manager state.
   */
  cleanup(): void {
    this.stopPolling()
    this.concurrency.clear()
    this.tasks.clear()
    this.pendingByParent.clear()
    this.pendingNotifications.clear()
    this.sessionToTask.clear()
  }

  /**
   * Cancels a running task.
   *
   * @param taskId - Task ID to cancel
   * @returns True if task was cancelled, false if task not found or not running
   */
  async cancelTask(taskId: string): Promise<boolean> {
    const task = this.tasks.get(taskId)
    if (!task || task.status !== "running") {
      return false
    }

    task.status = "cancelled"
    task.completedAt = new Date()
    task.error = "Task cancelled by user"
    this.markForNotification(task)
    await this.notifyParent(task)
    this.releaseTask(task)

    await this.client.tui.showToast({
      body: {
        title: "Background Task Cancelled",
        message: task.description,
        variant: "warning",
        duration: 3000,
      },
    })

    return true
  }

  // ========== Private Methods ==========

  /**
   * Tracks a task as pending for a parent session.
   */
  private trackPending(parentSessionId: string, taskId: string): void {
    const pending = this.pendingByParent.get(parentSessionId) ?? new Set()
    pending.add(taskId)
    this.pendingByParent.set(parentSessionId, pending)
  }

  /**
   * Marks a task as completed and queues notification.
   */
  private async completeTask(task: BackgroundTask): Promise<void> {
    task.status = "completed"
    task.completedAt = new Date()
    if (!task.result && task.progress?.lastMessage) {
      task.result = task.progress.lastMessage
    }
    this.releaseTask(task)
    this.markForNotification(task)
    await this.notifyParent(task)
  }

  /**
   * Handles task errors.
   */
  private async handleTaskError(task: BackgroundTask, error: Error): Promise<void> {
    task.status = "error"
    task.error = error.message
    task.completedAt = new Date()
    this.releaseTask(task)
    this.markForNotification(task)

    await this.client.tui.showToast({
      body: {
        title: "Background Task Failed",
        message: `${task.description}: ${error.message}`,
        variant: "error",
        duration: 5000,
      },
    })
  }

  /**
   * Releases concurrency slot and cleans up tracking.
   */
  private releaseTask(task: BackgroundTask): void {
    // H1: Clean up sessionToTask mapping to prevent memory leak
    this.sessionToTask.delete(task.sessionId)

    if (task.concurrencyKey) {
      this.concurrency.release(task.concurrencyKey)
    }

    // Remove from pending
    const pending = this.pendingByParent.get(task.parentSessionId)
    if (pending) {
      pending.delete(task.id)
      if (pending.size === 0) {
        this.pendingByParent.delete(task.parentSessionId)
      }
    }
  }

  /**
   * Marks a task for notification.
   */
  private markForNotification(task: BackgroundTask): void {
    const notifications = this.pendingNotifications.get(task.parentSessionId) ?? new Set()
    notifications.add(task.id)
    this.pendingNotifications.set(task.parentSessionId, notifications)
  }

  /**
   * Sends completion notification to parent session.
   */
  private async notifyParent(task: BackgroundTask): Promise<void> {
    const duration = this.formatDuration(task.startedAt, task.completedAt)
    const statusText = task.status === "completed" ? "completed" : "failed"
    const variant = task.status === "completed" ? "success" : "error"

    await this.client.tui.showToast({
      body: {
        title: `Background Task ${statusText.charAt(0).toUpperCase() + statusText.slice(1)}`,
        message: `${task.description} ${statusText} in ${duration}`,
        variant,
        duration: 5000,
      },
    })
  }

  /**
   * Formats duration between two dates.
   */
  private formatDuration(start?: Date, end?: Date): string {
    if (!start || !end) {
      return "unknown"
    }
    const ms = end.getTime() - start.getTime()
    const seconds = Math.floor(ms / 1000)
    const minutes = Math.floor(seconds / 60)

    if (minutes > 0) {
      const remainingSeconds = seconds % 60
      return `${minutes}m ${remainingSeconds}s`
    }
    return `${seconds}s`
  }

  /**
   * Starts polling for task status updates.
   */
  private startPolling(): void {
    if (this.pollInterval) {
      return
    }
    this.pollInterval = setInterval(() => {
      void this.pollTasks()
    }, DEFAULT_POLL_INTERVAL_MS)
  }

  /**
   * Stops polling for task status updates.
   */
  private stopPolling(): void {
    if (this.pollInterval) {
      clearInterval(this.pollInterval)
      this.pollInterval = null
    }
  }

  /**
   * Polls running tasks for status updates.
   */
  private async pollTasks(): Promise<void> {
    if (this.isPolling) {
      return
    }
    this.isPolling = true

    try {
      const runningTasks = this.getRunningTasks()
      if (runningTasks.length === 0) {
        this.stopPolling()
        return
      }

      // Get session statuses
      const statusResult = await this.client.session.status()
      if (statusResult.error || !statusResult.data) {
        return
      }

      const sessionStatuses = statusResult.data

      for (const task of runningTasks) {
        // H3: CHECK TIMEOUT FIRST (before completion check)
        const now = Date.now()
        const elapsed = now - (task.startedAt?.getTime() ?? task.createdAt.getTime())
        if (elapsed >= this.config.taskTimeoutMs) {
          task.status = "timeout"
          task.completedAt = new Date()
          task.error = `Task timed out after ${Math.round(this.config.taskTimeoutMs / 60000)} minutes`
          this.markForNotification(task)
          // H4: Add notification on timeout
          await this.notifyParent(task)
          this.releaseTask(task)

          await this.client.tui.showToast({
            body: {
              title: "Background Task Timeout",
              message: `${task.description} exceeded ${Math.floor(this.config.taskTimeoutMs / 60000)}min limit`,
              variant: "warning",
              duration: 5000,
            },
          })
          continue // Skip completion check for timed-out tasks
        }

        const sessionStatus = sessionStatuses[task.sessionId]
        if (!sessionStatus) {
          continue
        }

        // Check for completion (only if not timed out)
        if (sessionStatus.type === "idle") {
          // Session is idle - check if task completed
          const todoResult = await this.client.session.todo({
            path: { id: task.sessionId },
          })

          // If no pending todos or all completed, task is done
          // NOTE: Some OpenCode responses may omit `data` temporarily; treat that as "unknown"
          // instead of "no todos" to avoid premature completion.
          const todos = todoResult.data
          if (!todos) {
            continue
          }

          const allComplete = todos.every((t) => t.status === "completed")

          if (allComplete || todos.length === 0) {
            await this.completeTask(task)
          }
        }
      }
    } finally {
      this.isPolling = false
    }
  }
}
