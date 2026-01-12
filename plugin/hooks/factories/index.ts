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
export type { SessionStartConfig } from "./session-start.js"
// Session Start Hook
export {
  createSessionStartHook,
  sessionStartEntry,
} from "./session-start.js"

import { contextInjectionEntry } from "./context-injection.js"
// All registry entries for bulk registration
import { sessionStartEntry } from "./session-start.js"

/**
 * All built-in hook registry entries.
 */
export const builtInHookEntries = [sessionStartEntry, contextInjectionEntry] as const

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
