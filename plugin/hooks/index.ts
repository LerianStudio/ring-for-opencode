/**
 * Ring Hook System
 *
 * Exports all hook types, registry, and utilities.
 */

// Type definitions
export type {
  Hook,
  HookChatHandler,
  HookCompactionHandler,
  HookContext,
  HookEventHandler,
  HookFactory,
  HookLifecycle,
  HookName,
  HookOutput,
  HookRegistryEntry,
  HookResult,
  HookSystemHandler,
} from "./types.js"

// Registry
export {
  HookRegistry,
  hookRegistry,
  isHookDisabled,
} from "./registry.js"

export type { HookConfig } from "./registry.js"

// Hook factories (will be populated as factories are created)
export * from "./factories/index.js"
