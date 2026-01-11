/**
 * Ring Configuration System
 *
 * Layered configuration with Zod validation.
 *
 * @example
 * ```typescript
 * import {
 *   loadConfig,
 *   isHookDisabledInConfig,
 *   RingConfig,
 * } from "./config"
 *
 * const config = loadConfig("/path/to/project")
 * if (!isHookDisabledInConfig("session-start")) {
 *   // Hook is enabled
 * }
 * ```
 */

// Schema exports - types and validation
export {
  // Zod schemas
  HookNameSchema,
  AgentNameSchema,
  SkillNameSchema,
  CommandNameSchema,
  BackgroundTaskConfigSchema,
  NotificationConfigSchema,
  ExperimentalConfigSchema,
  RingConfigSchema,
  // TypeScript types
  type HookName,
  type AgentName,
  type SkillName,
  type CommandName,
  type BackgroundTaskConfig,
  type NotificationConfig,
  type ExperimentalConfig,
  type RingConfig,
  // Default values
  DEFAULT_RING_CONFIG,
} from "./schema"

// Loader exports - configuration loading and management
export {
  // Types
  type ConfigLayer,
  // Core functions
  loadConfig,
  getConfigLayers,
  checkConfigChanged,
  clearConfigCache,
  getCachedConfig,
  // File watching
  startConfigWatch,
  stopConfigWatch,
  // Disabled checks
  isHookDisabledInConfig,
  isAgentDisabledInConfig,
  isSkillDisabledInConfig,
  isCommandDisabledInConfig,
  // Config getters
  getHookConfig,
  getBackgroundTaskConfig,
  getNotificationConfig,
  getExperimentalConfig,
  // Utilities
  parseJsoncContent,
  deepMerge,
} from "./loader"

// Config handler for OpenCode injection
export {
  createConfigHandler,
  type OpenCodeConfig,
  type ConfigHandlerDeps,
} from "./config-handler"
