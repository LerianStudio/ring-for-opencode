// Re-export shared constants from parent module
export { EXIT_CODES, STATUS_COLORS, SYMBOLS } from "../constants"

export const CHECK_IDS = {
  OPENCODE_INSTALLATION: "opencode-installation",
  CONFIG_EXISTS: "config-exists",
  CONFIG_VALIDATION: "config-validation",
  SCHEMA_PRESENT: "schema-present",
  PLUGIN_DIRECTORY: "plugin-directory",
  SKILL_DIRECTORY: "skill-directory",
  STATE_DIRECTORY: "state-directory",
  BUN_INSTALLED: "bun-installed",
  GIT_INSTALLED: "git-installed",
} as const

export const CHECK_NAMES: Record<string, string> = {
  [CHECK_IDS.OPENCODE_INSTALLATION]: "OpenCode Installation",
  [CHECK_IDS.CONFIG_EXISTS]: "Configuration File",
  [CHECK_IDS.CONFIG_VALIDATION]: "Configuration Validity",
  [CHECK_IDS.SCHEMA_PRESENT]: "Schema Reference",
  [CHECK_IDS.PLUGIN_DIRECTORY]: "Plugin Directory",
  [CHECK_IDS.SKILL_DIRECTORY]: "Skill Directory",
  [CHECK_IDS.STATE_DIRECTORY]: "State Directory",
  [CHECK_IDS.BUN_INSTALLED]: "Bun Runtime",
  [CHECK_IDS.GIT_INSTALLED]: "Git",
} as const

export const CATEGORY_NAMES: Record<string, string> = {
  installation: "Installation",
  configuration: "Configuration",
  plugins: "Plugins & Skills",
  dependencies: "Dependencies",
} as const
