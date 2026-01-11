/**
 * Ring OpenCode
 *
 * Configuration schema validation and CLI tools for OpenCode.
 */

// Config exports
export {
  RingOpenCodeConfigSchema,
  RingHookNameSchema,
} from "./config"

export type {
  RingOpenCodeConfig,
  AgentConfig,
  AgentPermission,
  Permission,
  SkillsConfig,
  SkillDefinition,
  StateConfig,
  NotificationConfig,
  RingHookName,
} from "./config"

// Shared utilities
export {
  parseJsonc,
  parseJsoncSafe,
  readJsoncFile,
  detectConfigFile,
} from "./shared"

export type {
  JsoncParseResult,
} from "./shared"
