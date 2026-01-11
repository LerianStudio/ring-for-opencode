/**
 * Ring Configuration Loader
 *
 * Implements a 4-layer configuration system:
 * 1. Built-in defaults
 * 2. User config: ~/.config/opencode/ring/config.jsonc (or .json)
 * 3. Project config: .opencode/ring.jsonc or .ring/config.jsonc
 * 4. Directory overrides: .ring/local.jsonc
 */

import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"
import { parse as parseJsonc } from "jsonc-parser"
import {
  AgentNameSchema,
  CommandNameSchema,
  DEFAULT_RING_CONFIG,
  HookNameSchema,
  type RingConfig,
  RingConfigSchema,
  SkillNameSchema,
} from "./schema"

/**
 * Configuration layer metadata.
 */
export interface ConfigLayer {
  /** Layer name for debugging */
  name: string
  /** File path that was loaded (if any) */
  path: string | null
  /** Whether the file exists */
  exists: boolean
  /** Last modified timestamp */
  mtime: number | null
  /** The partial config from this layer */
  config: Partial<RingConfig> | null
}

/**
 * Module-level state for configuration caching and watching.
 */
let cachedConfig: RingConfig | null = null
let configLayers: ConfigLayer[] = []
let fileWatchers: fs.FSWatcher[] = []
let lastLoadedRoot: string | null = null

/**
 * Parse JSONC content (JSON with comments and trailing commas).
 * Uses jsonc-parser for robust parsing.
 *
 * @param content - The JSONC string content
 * @returns Parsed object
 * @throws Error if parsing fails
 */
export function parseJsoncContent<T>(content: string): T {
  const errors: Array<{ error: number; offset: number; length: number }> = []
  const result = parseJsonc(content, errors, {
    allowTrailingComma: true,
    disallowComments: false,
  })

  if (errors.length > 0) {
    const firstError = errors[0]
    throw new Error(
      `JSONC parse error at offset ${firstError.offset}: error code ${firstError.error}`,
    )
  }

  return result as T
}

/**
 * Keys that must never be merged/assigned from untrusted input.
 *
 * These are the common gadget keys used for prototype pollution.
 */
const FORBIDDEN_OBJECT_KEYS = new Set(["__proto__", "constructor", "prototype"])

function isForbiddenObjectKey(key: string): boolean {
  return FORBIDDEN_OBJECT_KEYS.has(key)
}

function isMergeableObject(value: unknown): value is Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return false
  }

  // Only merge plain objects (or null-prototype maps)
  const proto = Object.getPrototypeOf(value)
  return proto === Object.prototype || proto === null
}

/**
 * Deep merge two objects, with source overwriting target.
 * Arrays are replaced, not merged.
 *
 * SECURITY: Defends against prototype pollution by:
 * - using null-prototype result objects (Object.create(null))
 * - skipping forbidden gadget keys (__proto__/constructor/prototype)
 *
 * @param target - The base object
 * @param source - The object to merge in
 * @returns Merged object
 */
export function deepMerge<T extends Record<string, unknown>>(target: T, source: Partial<T>): T {
  const result = Object.create(null) as T

  // Copy existing keys from target (excluding forbidden keys)
  for (const [key, value] of Object.entries(target)) {
    if (isForbiddenObjectKey(key)) continue
    ;(result as Record<string, unknown>)[key] = value
  }

  if (!isMergeableObject(source)) {
    return result
  }

  for (const key of Object.keys(source)) {
    if (isForbiddenObjectKey(key)) {
      continue
    }

    const sourceValue = (source as Record<string, unknown>)[key]
    const targetValue = (target as Record<string, unknown>)[key]

    if (sourceValue === undefined) {
      continue
    }

    if (isMergeableObject(sourceValue) && isMergeableObject(targetValue)) {
      ;(result as Record<string, unknown>)[key] = deepMerge(
        targetValue as Record<string, unknown>,
        sourceValue as Record<string, unknown>,
      )
    } else {
      ;(result as Record<string, unknown>)[key] = sourceValue
    }
  }

  return result
}

/**
 * Try to load a config file from the given path.
 * Supports both .jsonc and .json extensions.
 *
 * @param filePath - Path to config file (without extension for auto-detection)
 * @param withExtension - If true, use path as-is; if false, try .jsonc then .json
 * @returns Config layer with metadata
 */
function tryLoadConfigFile(filePath: string, withExtension = true): ConfigLayer {
  const layer: ConfigLayer = {
    name: path.basename(filePath),
    path: null,
    exists: false,
    mtime: null,
    config: null,
  }

  const pathsToTry = withExtension ? [filePath] : [`${filePath}.jsonc`, `${filePath}.json`]

  for (const tryPath of pathsToTry) {
    try {
      const stats = fs.statSync(tryPath)
      if (stats.isFile()) {
        layer.path = tryPath
        layer.exists = true
        layer.mtime = stats.mtimeMs

        const content = fs.readFileSync(tryPath, "utf-8")
        layer.config = parseJsoncContent<Partial<RingConfig>>(content)
        break
      }
    } catch {
      // File doesn't exist or can't be read, continue
    }
  }

  return layer
}

/**
 * Get the user config directory path.
 * Uses XDG_CONFIG_HOME if set and absolute, otherwise ~/.config.
 * M4: Removed redundant darwin branch that returned the same path.
 */
function getUserConfigDir(): string {
  const home = os.homedir()
  const xdgConfig = process.env.XDG_CONFIG_HOME
  if (xdgConfig && path.isAbsolute(xdgConfig)) {
    return path.join(xdgConfig, "opencode", "ring")
  }
  return path.join(home, ".config", "opencode", "ring")
}

/**
 * Load Ring configuration with 4-layer merging.
 *
 * @param root - Project root directory
 * @param forceReload - Force reload even if cached
 * @returns Merged configuration
 */
export function loadConfig(root: string, forceReload = false): RingConfig {
  // Return cached config if available and not forcing reload
  if (cachedConfig && lastLoadedRoot === root && !forceReload) {
    return cachedConfig
  }

  const layers: ConfigLayer[] = []

  // Layer 1: Built-in defaults
  layers.push({
    name: "defaults",
    path: null,
    exists: true,
    mtime: null,
    config: DEFAULT_RING_CONFIG,
  })

  // Layer 2: User config (~/.config/opencode/ring/config.jsonc)
  const userConfigDir = getUserConfigDir()
  const userConfigPath = path.join(userConfigDir, "config")
  layers.push(tryLoadConfigFile(userConfigPath, false))
  layers[1].name = "user"

  // Layer 3: Project config (.opencode/ring.jsonc or .ring/config.jsonc)
  const projectConfig1 = path.join(root, ".opencode", "ring.jsonc")
  const projectConfig2 = path.join(root, ".ring", "config.jsonc")

  let projectLayer = tryLoadConfigFile(projectConfig1)
  if (!projectLayer.exists) {
    projectLayer = tryLoadConfigFile(projectConfig2)
  }
  projectLayer.name = "project"
  layers.push(projectLayer)

  // Layer 4: Directory overrides (.ring/local.jsonc)
  const localConfigPath = path.join(root, ".ring", "local.jsonc")
  const localLayer = tryLoadConfigFile(localConfigPath)
  localLayer.name = "local"
  layers.push(localLayer)

  // Merge all layers
  let merged: RingConfig = { ...DEFAULT_RING_CONFIG }
  for (const layer of layers) {
    if (layer.config) {
      merged = deepMerge(merged, layer.config)
    }
  }

  // Validate merged config
  const parseResult = RingConfigSchema.safeParse(merged)
  if (!parseResult.success) {
    console.error("[Ring] Configuration validation failed:", parseResult.error.issues)
    // Return defaults on validation failure
    merged = { ...DEFAULT_RING_CONFIG }
  } else {
    merged = parseResult.data
  }

  // Cache the result
  cachedConfig = merged
  configLayers = layers
  lastLoadedRoot = root

  return merged
}

/**
 * Get metadata about loaded configuration layers.
 *
 * @returns Array of config layer metadata
 */
export function getConfigLayers(): ConfigLayer[] {
  return [...configLayers]
}

/**
 * Check if any configuration file has changed since last load.
 *
 * @returns True if files have changed
 */
export function checkConfigChanged(): boolean {
  for (const layer of configLayers) {
    if (!layer.path) continue

    try {
      const stats = fs.statSync(layer.path)
      if (stats.mtimeMs !== layer.mtime) {
        return true
      }
    } catch {
      // File was deleted
      if (layer.exists) {
        return true
      }
    }
  }

  return false
}

/**
 * Start watching configuration files for changes.
 *
 * @param root - Project root directory
 * @param onChange - Callback when config changes
 */
export function startConfigWatch(root: string, onChange: (config: RingConfig) => void): void {
  // Stop any existing watchers
  stopConfigWatch()

  // Paths to watch
  const watchPaths = [
    path.join(getUserConfigDir(), "config.jsonc"),
    path.join(getUserConfigDir(), "config.json"),
    path.join(root, ".opencode", "ring.jsonc"),
    path.join(root, ".ring", "config.jsonc"),
    path.join(root, ".ring", "local.jsonc"),
  ]

  // Debounce reload to avoid multiple rapid reloads
  let reloadTimeout: NodeJS.Timeout | null = null
  const debouncedReload = () => {
    if (reloadTimeout) {
      clearTimeout(reloadTimeout)
    }
    reloadTimeout = setTimeout(() => {
      const newConfig = loadConfig(root, true)
      onChange(newConfig)
    }, 100)
  }

  for (const watchPath of watchPaths) {
    try {
      // Ensure parent directory exists before watching
      const dir = path.dirname(watchPath)
      if (fs.existsSync(dir)) {
        const watcher = fs.watch(dir, (_eventType, filename) => {
          if (filename && path.join(dir, filename) === watchPath) {
            debouncedReload()
          }
        })
        fileWatchers.push(watcher)
      }
    } catch {
      // Directory doesn't exist, skip watching
    }
  }
}

/**
 * Stop watching configuration files.
 */
export function stopConfigWatch(): void {
  for (const watcher of fileWatchers) {
    watcher.close()
  }
  fileWatchers = []
}

/**
 * Clear the configuration cache.
 * Next loadConfig call will reload from disk.
 */
export function clearConfigCache(): void {
  cachedConfig = null
  configLayers = []
  lastLoadedRoot = null
}

/**
 * Get the current cached configuration.
 * Returns null if not loaded.
 */
export function getCachedConfig(): RingConfig | null {
  return cachedConfig
}

/**
 * Check if a hook is disabled in the configuration.
 * H5: Uses Zod validation instead of unsafe type assertion.
 *
 * @param hookName - The hook name to check
 * @returns True if the hook is disabled
 */
export function isHookDisabledInConfig(hookName: string): boolean {
  if (!cachedConfig) {
    return false
  }
  const parsed = HookNameSchema.safeParse(hookName)
  if (!parsed.success) {
    return false
  }
  return cachedConfig.disabled_hooks.includes(parsed.data)
}

/**
 * Check if an agent is disabled in the configuration.
 * H5: Uses Zod validation instead of unsafe type assertion.
 *
 * @param agentName - The agent name to check
 * @returns True if the agent is disabled
 */
export function isAgentDisabledInConfig(agentName: string): boolean {
  if (!cachedConfig) {
    return false
  }
  const parsed = AgentNameSchema.safeParse(agentName)
  if (!parsed.success) {
    return false
  }
  return cachedConfig.disabled_agents.includes(parsed.data)
}

/**
 * Check if a skill is disabled in the configuration.
 * H5: Uses Zod validation instead of unsafe type assertion.
 *
 * @param skillName - The skill name to check
 * @returns True if the skill is disabled
 */
export function isSkillDisabledInConfig(skillName: string): boolean {
  if (!cachedConfig) {
    return false
  }
  const parsed = SkillNameSchema.safeParse(skillName)
  if (!parsed.success) {
    return false
  }
  return cachedConfig.disabled_skills.includes(parsed.data)
}

/**
 * Check if a command is disabled in the configuration.
 * H5: Uses Zod validation instead of unsafe type assertion.
 *
 * @param commandName - The command name to check
 * @returns True if the command is disabled
 */
export function isCommandDisabledInConfig(commandName: string): boolean {
  if (!cachedConfig) {
    return false
  }
  const parsed = CommandNameSchema.safeParse(commandName)
  if (!parsed.success) {
    return false
  }
  return cachedConfig.disabled_commands.includes(parsed.data)
}

/**
 * Get a specific hook's custom configuration.
 *
 * @param hookName - The hook name
 * @returns Hook config or undefined
 */
export function getHookConfig(hookName: string): Record<string, unknown> | undefined {
  if (!cachedConfig?.hooks) {
    return undefined
  }
  return cachedConfig.hooks[hookName]
}

/**
 * Get the background tasks configuration.
 *
 * @returns Background task config
 */
export function getBackgroundTaskConfig(): RingConfig["background_tasks"] {
  if (!cachedConfig) {
    return DEFAULT_RING_CONFIG.background_tasks
  }
  return cachedConfig.background_tasks
}

/**
 * Get the notification configuration.
 *
 * @returns Notification config
 */
export function getNotificationConfig(): RingConfig["notifications"] {
  if (!cachedConfig) {
    return DEFAULT_RING_CONFIG.notifications
  }
  return cachedConfig.notifications
}

/**
 * Get the experimental features configuration.
 *
 * @returns Experimental config
 */
export function getExperimentalConfig(): RingConfig["experimental"] {
  if (!cachedConfig) {
    return DEFAULT_RING_CONFIG.experimental
  }
  return cachedConfig.experimental
}
