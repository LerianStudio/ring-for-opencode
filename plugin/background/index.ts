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

// Classes
export { ConcurrencyManager } from "./concurrency"
export { BackgroundManager } from "./manager"
// Types
export type {
  BackgroundClient,
  BackgroundManagerEvents,
  BackgroundTask,
  BackgroundTaskStatus,
  BackgroundTaskType,
  LaunchTaskInput,
  ResumeTaskInput,
  TaskNotification,
  TaskProgress,
} from "./types"
