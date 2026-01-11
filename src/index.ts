/**
 * Ring OpenCode
 *
 * Configuration schema validation and CLI tools for OpenCode.
 */

export type {
  AgentConfig,
  AgentPermission,
  NotificationConfig,
  Permission,
  RingHookName,
  RingOpenCodeConfig,
  SkillDefinition,
  SkillsConfig,
  StateConfig,
} from "./config"
// Config exports
export {
  RingHookNameSchema,
  RingOpenCodeConfigSchema,
} from "./config"
export type { JsoncParseResult } from "./shared"
// Shared utilities
export {
  detectConfigFile,
  parseJsonc,
  parseJsoncSafe,
  readJsoncFile,
} from "./shared"
