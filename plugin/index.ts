/**
 * Ring OpenCode Plugins
 *
 * This module exports Ring plugins for OpenCode.
 *
 * Architecture:
 * - Unified plugin: Single entry matching oh-my-opencode pattern (recommended)
 * - Hook system: Middleware pattern with lifecycle events
 * - Layered config: 4-layer priority with deep merge
 * - Background tasks: Parallel agent execution manager
 * - Component loaders: Load agents, skills, commands from .opencode/
 *
 * The unified plugin combines:
 * - Config handler: Injects 16 agents, 30 skills, 16 commands
 * - Tool registration: ring_doubt tool
 * - Event routing: Lifecycle events to hooks
 * - System transform: Context injection
 * - Compaction: Context preservation
 */

// =============================================================================
// UNIFIED PLUGIN (Recommended - matches oh-my-opencode pattern)
// =============================================================================

export { RingUnifiedPlugin } from "./ring-unified.js"

// =============================================================================
// LEGACY PLUGIN (Hook-based architecture)
// =============================================================================

// Main plugin - hook-based architecture
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

// ============================================================================
// Component Loaders (for unified plugin)
// ============================================================================

export {
  loadRingAgents,
  loadRingSkills,
  loadRingCommands,
  countRingAgents,
  countRingSkills,
  countRingCommands,
} from "./loaders/index.js"

export type {
  AgentConfig,
  SkillConfig,
  CommandConfig,
} from "./loaders/index.js"

// ============================================================================
// Tools (for unified plugin)
// ============================================================================

export { ringTools, ringDoubtTool } from "./tools/index.js"

// ============================================================================
// Lifecycle Router (for unified plugin)
// ============================================================================

export {
  createLifecycleRouter,
  EVENTS,
} from "./lifecycle/index.js"

export type {
  OpenCodeEvent,
  LifecycleRouterDeps,
} from "./lifecycle/index.js"

// ============================================================================
// Config Handler (for unified plugin)
// ============================================================================

export { createConfigHandler } from "./config/index.js"

export type {
  OpenCodeConfig,
  ConfigHandlerDeps,
} from "./config/index.js"
