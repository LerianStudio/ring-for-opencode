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

// Config handler for OpenCode injection
export {
  type ConfigHandlerDeps,
  createConfigHandler,
  type OpenCodeConfig,
} from "./config-handler"

// Loader exports - configuration loading and management
export {
  // Types
  type ConfigLayer,
  checkConfigChanged,
  clearConfigCache,
  deepMerge,
  getCachedConfig,
  getConfigLayers,
  getExperimentalConfig,
  // Config getters
  getHookConfig,
  isAgentDisabledInConfig,
  isCommandDisabledInConfig,
  // Disabled checks
  isHookDisabledInConfig,
  isSkillDisabledInConfig,
  // Core functions
  loadConfig,
  // Utilities
  parseJsoncContent,
  // File watching
  startConfigWatch,
  stopConfigWatch,
} from "./loader"
// Schema exports - types and validation
export {
  type AgentName,
  AgentNameSchema,
  type CommandName,
  CommandNameSchema,
  // Default values
  DEFAULT_RING_CONFIG,
  type ExperimentalConfig,
  ExperimentalConfigSchema,
  // TypeScript types
  type HookName,
  // Zod schemas
  HookNameSchema,
  type RingConfig,
  RingConfigSchema,
  type SkillName,
  SkillNameSchema,
} from "./schema"
