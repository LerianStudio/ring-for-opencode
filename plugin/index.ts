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

// Hook factory config types
export type {
  ContextInjectionConfig,
  NotificationConfig as HookNotificationConfig,
  SessionStartConfig,
  TaskCompletionConfig,
} from "./hooks/factories/index.js"
// Hook factories
export {
  builtInHookEntries,
  createContextInjectionHook,
  createNotificationHook,
  createSessionStartHook,
  createTaskCompletionHook,
  registerBuiltInHooks,
} from "./hooks/factories/index.js"
// Type definitions
export type {
  Hook,
  HookChatHandler,
  HookCompactionHandler,
  HookConfig,
  HookContext,
  HookEventHandler,
  HookFactory,
  HookLifecycle,
  HookName,
  HookOutput,
  HookRegistryEntry,
  HookResult,
  HookSystemHandler,
} from "./hooks/index.js"
// Registry and utilities
export { HookRegistry, hookRegistry, isHookDisabled } from "./hooks/index.js"

// ============================================================================
// Configuration Exports
// ============================================================================

// Config types
export type {
  AgentName,
  BackgroundTaskConfig,
  CommandName,
  ConfigLayer,
  ExperimentalConfig,
  HookName as ConfigHookName,
  NotificationConfig,
  RingConfig,
  SkillName,
} from "./config/index.js"
// Core config functions
// Disabled item checks
// Config getters
// Config utilities
// Default values
export {
  checkConfigChanged,
  clearConfigCache,
  DEFAULT_RING_CONFIG,
  deepMerge,
  getBackgroundTaskConfig,
  getCachedConfig,
  getConfigLayers,
  getExperimentalConfig,
  getHookConfig,
  getNotificationConfig,
  isAgentDisabledInConfig,
  isCommandDisabledInConfig,
  isHookDisabledInConfig,
  isSkillDisabledInConfig,
  loadConfig,
  parseJsoncContent,
  startConfigWatch,
  stopConfigWatch,
} from "./config/index.js"

// ============================================================================
// Background Task Exports
// ============================================================================

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
} from "./background/index.js"
// Classes
export { BackgroundManager, ConcurrencyManager } from "./background/index.js"

// ============================================================================
// State Utilities
// ============================================================================

export {
  cleanupOldState,
  deleteState,
  escapeAngleBrackets,
  findMostRecentFile,
  getSessionId,
  isPathWithinRoot,
  readFileSafe,
  readState,
  sanitizeForPrompt,
  writeState,
} from "./utils/state.js"

// ============================================================================
// Component Loaders (for unified plugin)
// ============================================================================

export type {
  AgentConfig,
  CommandConfig,
  SkillConfig,
} from "./loaders/index.js"
export {
  countRingAgents,
  countRingCommands,
  countRingSkills,
  loadRingAgents,
  loadRingCommands,
  loadRingSkills,
} from "./loaders/index.js"

// ============================================================================
// Tools (for unified plugin)
// ============================================================================

export { ringDoubtTool, ringTools } from "./tools/index.js"

// ============================================================================
// Lifecycle Router (for unified plugin)
// ============================================================================

export type {
  LifecycleRouterDeps,
  OpenCodeEvent,
} from "./lifecycle/index.js"
export {
  createLifecycleRouter,
  EVENTS,
} from "./lifecycle/index.js"

// ============================================================================
// Config Handler (for unified plugin)
// ============================================================================

export type {
  ConfigHandlerDeps,
  OpenCodeConfig,
} from "./config/index.js"
export { createConfigHandler } from "./config/index.js"
