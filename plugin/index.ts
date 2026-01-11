/**
 * Ring OpenCode Plugins
 *
 * This module exports Ring plugins for OpenCode.
 *
 * Architecture:
 * - Hook system: Middleware pattern with lifecycle events
 * - Layered config: 4-layer priority with deep merge
 * - Background tasks: Parallel agent execution manager
 */

// Main plugin - new hook-based architecture
export { RingPlugin, RingPlugin as default } from "./ring-plugin.js"

// ============================================================================
// Hook System Exports
// ============================================================================

// Registry and utilities
export { HookRegistry, hookRegistry, isHookDisabled } from "./hooks/index.js"
export type { HookConfig } from "./hooks/index.js"

// Type definitions
export type {
  Hook,
  HookFactory,
  HookName,
  HookLifecycle,
  HookContext,
  HookOutput,
  HookResult,
  HookRegistryEntry,
  HookEventHandler,
  HookChatHandler,
  HookCompactionHandler,
  HookSystemHandler,
} from "./hooks/index.js"

// Hook factories
export {
  createSessionStartHook,
  createContextInjectionHook,
  createNotificationHook,
  createTaskCompletionHook,
  builtInHookEntries,
  registerBuiltInHooks,
} from "./hooks/factories/index.js"

// Hook factory config types
export type {
  SessionStartConfig,
  ContextInjectionConfig,
  NotificationConfig as HookNotificationConfig,
  TaskCompletionConfig,
} from "./hooks/factories/index.js"

// ============================================================================
// Configuration Exports
// ============================================================================

// Core config functions
export {
  loadConfig,
  getConfigLayers,
  startConfigWatch,
  stopConfigWatch,
  clearConfigCache,
  getCachedConfig,
  checkConfigChanged,
} from "./config/index.js"

// Disabled item checks
export {
  isHookDisabledInConfig,
  isAgentDisabledInConfig,
  isSkillDisabledInConfig,
  isCommandDisabledInConfig,
} from "./config/index.js"

// Config getters
export {
  getHookConfig,
  getBackgroundTaskConfig,
  getNotificationConfig,
  getExperimentalConfig,
} from "./config/index.js"

// Config utilities
export {
  parseJsoncContent,
  deepMerge,
} from "./config/index.js"

// Default values
export { DEFAULT_RING_CONFIG } from "./config/index.js"

// Config types
export type {
  RingConfig,
  HookName as ConfigHookName,
  AgentName,
  SkillName,
  CommandName,
  BackgroundTaskConfig,
  NotificationConfig,
  ExperimentalConfig,
  ConfigLayer,
} from "./config/index.js"

// ============================================================================
// Background Task Exports
// ============================================================================

// Classes
export { BackgroundManager, ConcurrencyManager } from "./background/index.js"

// Types
export type {
  BackgroundTask,
  BackgroundTaskStatus,
  BackgroundTaskType,
  LaunchTaskInput,
  ResumeTaskInput,
  TaskProgress,
  TaskNotification,
  BackgroundManagerEvents,
  BackgroundClient,
} from "./background/index.js"

// ============================================================================
// State Utilities
// ============================================================================

export {
  getSessionId,
  writeState,
  readState,
  deleteState,
  cleanupOldState,
  sanitizeForPrompt,
  escapeAngleBrackets,
  isPathWithinRoot,
  findMostRecentFile,
  readFileSafe,
} from "./utils/state.js"
