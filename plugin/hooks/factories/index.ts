/**
 * Ring Hook Factories Index
 *
 * Exports all hook factories and their configuration types.
 */

// Session Start Hook
export {
  createSessionStartHook,
  sessionStartEntry,
} from "./session-start.js"
export type { SessionStartConfig } from "./session-start.js"

// Context Injection Hook
export {
  createContextInjectionHook,
  contextInjectionEntry,
} from "./context-injection.js"
export type { ContextInjectionConfig } from "./context-injection.js"

// Notification Hook
export {
  createNotificationHook,
  notificationEntry,
} from "./notification.js"
export type { NotificationConfig } from "./notification.js"

// Task Completion Hook
export {
  createTaskCompletionHook,
  taskCompletionEntry,
} from "./task-completion.js"
export type { TaskCompletionConfig } from "./task-completion.js"

// All registry entries for bulk registration
import { sessionStartEntry } from "./session-start.js"
import { contextInjectionEntry } from "./context-injection.js"
import { notificationEntry } from "./notification.js"
import { taskCompletionEntry } from "./task-completion.js"

/**
 * All built-in hook registry entries.
 */
export const builtInHookEntries = [
  sessionStartEntry,
  contextInjectionEntry,
  notificationEntry,
  taskCompletionEntry,
] as const

/**
 * Register all built-in hooks with a registry.
 */
export function registerBuiltInHooks(
  registry: { registerFactory: (entry: typeof builtInHookEntries[number]) => void }
): void {
  for (const entry of builtInHookEntries) {
    registry.registerFactory(entry)
  }
}
