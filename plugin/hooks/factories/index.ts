/**
 * Ring Hook Factories Index
 *
 * Exports all hook factories and their configuration types.
 */

export type { ContextInjectionConfig } from "./context-injection.js"
// Context Injection Hook
export {
  contextInjectionEntry,
  createContextInjectionHook,
} from "./context-injection.js"
export type { NotificationConfig } from "./notification.js"
// Notification Hook
export {
  createNotificationHook,
  notificationEntry,
} from "./notification.js"
export type { SessionStartConfig } from "./session-start.js"
// Session Start Hook
export {
  createSessionStartHook,
  sessionStartEntry,
} from "./session-start.js"
export type { TaskCompletionConfig } from "./task-completion.js"
// Task Completion Hook
export {
  createTaskCompletionHook,
  taskCompletionEntry,
} from "./task-completion.js"

import { contextInjectionEntry } from "./context-injection.js"
import { notificationEntry } from "./notification.js"
// All registry entries for bulk registration
import { sessionStartEntry } from "./session-start.js"
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
export function registerBuiltInHooks(registry: {
  registerFactory: (entry: (typeof builtInHookEntries)[number]) => void
}): void {
  for (const entry of builtInHookEntries) {
    registry.registerFactory(entry)
  }
}
