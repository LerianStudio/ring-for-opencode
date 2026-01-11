/**
 * Ring Background Task System
 *
 * Provides background task execution for long-running agent operations.
 *
 * @example
 * ```typescript
 * import {
 *   BackgroundManager,
 *   ConcurrencyManager,
 *   type BackgroundTask,
 *   type LaunchTaskInput,
 * } from "./background"
 *
 * const manager = new BackgroundManager(client, "/workspace", config)
 * const task = await manager.launch({
 *   description: "Explore codebase",
 *   prompt: "Analyze the project structure",
 *   agent: "codebase-explorer",
 *   parentSessionId: "session-123",
 * })
 * ```
 */

// Types
export type {
  BackgroundTaskStatus,
  TaskProgress,
  BackgroundTask,
  BackgroundTaskType,
  LaunchTaskInput,
  ResumeTaskInput,
  TaskNotification,
  BackgroundManagerEvents,
  BackgroundClient,
} from "./types"

// Classes
export { ConcurrencyManager } from "./concurrency"
export { BackgroundManager } from "./manager"
