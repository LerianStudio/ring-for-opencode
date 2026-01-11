import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs"
import { homedir } from "node:os"
import { dirname, join } from "node:path"
import { RingOpenCodeConfigSchema } from "../config"
import { parseJsonc } from "../shared"
import { PROJECT_CONFIG_PATHS, SCHEMA_URL, USER_CONFIG_PATHS } from "./constants"
import type { ConfigMergeResult, DetectedConfig } from "./types"

interface NodeError extends Error {
  code?: string
}

function isPermissionError(err: unknown): boolean {
  const nodeErr = err as NodeError
  return nodeErr?.code === "EACCES" || nodeErr?.code === "EPERM"
}

function isFileNotFoundError(err: unknown): boolean {
  const nodeErr = err as NodeError
  return nodeErr?.code === "ENOENT"
}

function formatErrorWithSuggestion(err: unknown, context: string): string {
  if (isPermissionError(err)) {
    return `Permission denied: Cannot ${context}. Try running with elevated permissions.`
  }

  if (isFileNotFoundError(err)) {
    return `File not found while trying to ${context}.`
  }

  if (err instanceof SyntaxError) {
    return `JSON syntax error while trying to ${context}: ${err.message}`
  }

  const message = err instanceof Error ? err.message : String(err)
  return `Failed to ${context}: ${message}`
}

/**
 * Get candidate config paths (project then user)
 */
export function getConfigPaths(): string[] {
  const projectPaths = PROJECT_CONFIG_PATHS.map((configPath) => join(process.cwd(), configPath))
  const userPaths = USER_CONFIG_PATHS.map((configPath) => join(homedir(), configPath))
  return [...projectPaths, ...userPaths]
}

/**
 * Get the active config path
 */
export function getConfigPath(): string {
  const configPaths = getConfigPaths()
  for (const configPath of configPaths) {
    if (existsSync(configPath)) {
      return configPath
    }
  }
  return configPaths[0]
}

/**
 * Detect current Ring configuration
 */
export function detectCurrentConfig(): DetectedConfig {
  const configPath = getConfigPath()
  const result: DetectedConfig = {
    isInstalled: false,
    configPath: null,
    hasSchema: false,
    version: null,
  }

  if (!existsSync(configPath)) {
    return result
  }

  result.configPath = configPath

  try {
    const content = readFileSync(configPath, "utf-8")
    const config = parseJsonc<Record<string, unknown>>(content)

    if (config) {
      result.isInstalled = true
      result.hasSchema = typeof config.$schema === "string"
      result.version = typeof config.version === "string" ? config.version : null
    }
  } catch {
    // Config exists but is invalid
    result.isInstalled = false
  }

  return result
}

/**
 * Validate configuration against Zod schema
 */
export function validateConfig(configPath: string): { valid: boolean; errors: string[] } {
  try {
    const content = readFileSync(configPath, "utf-8")
    const rawConfig = parseJsonc<Record<string, unknown>>(content)
    const result = RingOpenCodeConfigSchema.safeParse(rawConfig)

    if (!result.success) {
      const errors = result.error.issues.map((i) => `${i.path.join(".")}: ${i.message}`)
      return { valid: false, errors }
    }

    return { valid: true, errors: [] }
  } catch (err) {
    return {
      valid: false,
      errors: [err instanceof Error ? err.message : "Failed to parse config"],
    }
  }
}

/**
 * Add $schema to existing config for IDE autocomplete
 */
export function addSchemaToConfig(): ConfigMergeResult {
  const configPath = getConfigPath()

  try {
    if (!existsSync(configPath)) {
      // Create minimal config with schema
      const config = {
        $schema: SCHEMA_URL,
        version: "1.0.0",
        name: "ring-opencode",
        description: "Ring configuration",
      }
      mkdirSync(dirname(configPath), { recursive: true })
      writeFileSync(configPath, `${JSON.stringify(config, null, 2)}\n`)
      return { success: true, configPath }
    }

    const content = readFileSync(configPath, "utf-8")
    const config = parseJsonc<Record<string, unknown>>(content)

    if (!config) {
      return { success: false, configPath, error: "Failed to parse existing config" }
    }

    // Already has schema
    if (config.$schema === SCHEMA_URL) {
      return { success: true, configPath }
    }

    // Add/update $schema - destructure to avoid duplicate key
    const { $schema: _existingSchema, ...restConfig } = config as { $schema?: string }
    const finalConfig = { $schema: SCHEMA_URL, ...restConfig }

    writeFileSync(configPath, `${JSON.stringify(finalConfig, null, 2)}\n`)
    return { success: true, configPath }
  } catch (err) {
    return { success: false, configPath, error: formatErrorWithSuggestion(err, "update config") }
  }
}

/**
 * Check if opencode CLI is installed
 */
export async function isOpenCodeInstalled(): Promise<boolean> {
  try {
    const proc = Bun.spawn(["opencode", "--version"], {
      stdout: "pipe",
      stderr: "pipe",
    })
    await proc.exited
    return proc.exitCode === 0
  } catch {
    return false
  }
}

/**
 * Get opencode version
 */
export async function getOpenCodeVersion(): Promise<string | null> {
  try {
    const proc = Bun.spawn(["opencode", "--version"], {
      stdout: "pipe",
      stderr: "pipe",
    })
    const output = await new Response(proc.stdout).text()
    await proc.exited
    if (proc.exitCode === 0) {
      return output.trim()
    }
    return null
  } catch {
    return null
  }
}
