/**
 * Orchestrator Configuration
 *
 * Three-level configuration merging:
 * 1. Built-in defaults (in code)
 * 2. Global config (~/.config/opencode/ring/orchestrator.json)
 * 3. Project config (.ring/orchestrator.json or .opencode/orchestrator.json)
 *
 * Ported from Orchestra's configuration pattern.
 */

import {
  closeSync,
  existsSync,
  fstatSync,
  lstatSync,
  openSync,
  readSync,
  realpathSync,
} from "node:fs"
import { homedir } from "node:os"
import { join, normalize } from "node:path"
import { z } from "zod"
import type { OrchestratorContext, WorkerProfile, WorkflowRunLimits } from "./types.js"
import { builtInProfiles } from "./profiles.js"
import { BLOCKED_HOSTS } from "./types.js"

// =============================================================================
// Security Constants
// =============================================================================

const PROTECTED_PROFILE_IDS = ["vision", "docs", "coder", "architect", "explorer"] as const
const MAX_CONFIG_SIZE = 1_000_000 // 1MB

// =============================================================================
// Zod Schemas
// =============================================================================

const WorkerProfileSchema = z.object({
  id: z.string(),
  name: z.string(),
  kind: z.enum(["server", "agent", "subagent"]).optional(),
  model: z.string(),
  purpose: z.string(),
  whenToUse: z.string(),
  systemPrompt: z.string().optional(),
  supportsVision: z.boolean().optional(),
  supportsWeb: z.boolean().optional(),
  tools: z.record(z.string(), z.boolean()).optional(),
  temperature: z.number().optional(),
  tags: z.array(z.string()).optional(),
})

const WorkflowLimitsSchema = z.object({
  maxSteps: z.number().default(4),
  maxTaskChars: z.number().default(12000),
  maxCarryChars: z.number().default(24000),
  perStepTimeoutMs: z.number().default(120000),
})

const OrchestratorConfigFileSchema = z
  .object({
    $schema: z.string().optional(),
    basePort: z.number().default(14096),
    autoSpawn: z.boolean().default(true),
    startupTimeout: z.number().default(30000),
    ui: z
      .object({
        toasts: z.boolean().default(true),
        injectSystemContext: z.boolean().default(true),
        systemContextMaxWorkers: z.number().default(12),
        defaultListFormat: z.enum(["markdown", "json"]).default("markdown"),
        debug: z.boolean().default(false),
      })
      .default({
        toasts: true,
        injectSystemContext: true,
        systemContextMaxWorkers: 12,
        defaultListFormat: "markdown",
        debug: false,
      }),
    profiles: z.array(z.union([z.string(), WorkerProfileSchema])).default([]),
    workers: z.array(z.union([z.string(), WorkerProfileSchema])).default([]),
    workflows: z
      .object({
        enabled: z.boolean().default(true),
        limits: WorkflowLimitsSchema.default({
          maxSteps: 4,
          maxTaskChars: 12000,
          maxCarryChars: 24000,
          perStepTimeoutMs: 120000,
        }),
      })
      .default({
        enabled: true,
        limits: {
          maxSteps: 4,
          maxTaskChars: 12000,
          maxCarryChars: 24000,
          perStepTimeoutMs: 120000,
        },
      }),
    security: z
      .object({
        preventRecursiveSpawn: z.boolean().default(true),
        recursiveSpawnEnvVar: z.string().default("RING_ORCHESTRATOR_WORKER"),
      })
      .default({
        preventRecursiveSpawn: true,
        recursiveSpawnEnvVar: "RING_ORCHESTRATOR_WORKER",
      }),
  })
  .passthrough()

export type OrchestratorConfigFile = z.infer<typeof OrchestratorConfigFileSchema>

// =============================================================================
// Default Configuration
// =============================================================================

const DEFAULT_CONFIG: OrchestratorConfigFile = {
  basePort: 14096,
  autoSpawn: true,
  startupTimeout: 30000,
  ui: {
    toasts: true,
    injectSystemContext: true,
    systemContextMaxWorkers: 12,
    defaultListFormat: "markdown",
    debug: false,
  },
  profiles: [],
  workers: [],
  workflows: {
    enabled: true,
    limits: {
      maxSteps: 4,
      maxTaskChars: 12000,
      maxCarryChars: 24000,
      perStepTimeoutMs: 120000,
    },
  },
  security: {
    preventRecursiveSpawn: true,
    recursiveSpawnEnvVar: "RING_ORCHESTRATOR_WORKER",
  },
}

// =============================================================================
// Configuration Loading
// =============================================================================

function getUserConfigDir(): string {
  const home = homedir()
  const xdgConfig = process.env.XDG_CONFIG_HOME
  if (xdgConfig && xdgConfig.startsWith("/")) {
    return join(xdgConfig, "opencode", "ring")
  }
  return join(home, ".config", "opencode", "ring")
}

/**
 * Safely read JSON config file with TOCTOU protection.
 * Uses file descriptor to ensure atomic check and read.
 */
function tryReadJson(filePath: string, projectRoot?: string): unknown | null {
  let fd: number | null = null
  try {
    if (!existsSync(filePath)) return null

    // Pre-check with lstat for symlink outside project (before opening)
    const lstats = lstatSync(filePath)
    if (lstats.isSymbolicLink()) {
      const realPath = realpathSync(filePath)
      const normalizedRoot = projectRoot ? normalize(projectRoot) : null

      if (normalizedRoot && !realPath.startsWith(normalizedRoot)) {
        console.warn(`[Orchestrator] Blocked symlink outside project: ${filePath}`)
        return null
      }
    }

    // SEC-HIGH-2 FIX: Open file and use fd for atomic stat+read
    fd = openSync(filePath, "r")
    const stats = fstatSync(fd)

    // Size limit to prevent DoS
    if (stats.size > MAX_CONFIG_SIZE) {
      console.warn(`[Orchestrator] Config file too large (${stats.size} bytes): ${filePath}`)
      return null
    }

    // Read from same fd to prevent TOCTOU
    const buffer = Buffer.alloc(stats.size)
    const bytesRead = readSync(fd, buffer, 0, stats.size, 0)
    const content = buffer.slice(0, bytesRead).toString("utf-8")

    return JSON.parse(content)
  } catch (err) {
    if (existsSync(filePath)) {
      console.warn(
        `[Orchestrator] Failed to parse ${filePath}:`,
        err instanceof Error ? err.message : err,
      )
    }
    return null
  } finally {
    if (fd !== null) {
      try {
        closeSync(fd)
      } catch {
        // Ignore close errors
      }
    }
  }
}

/**
 * Validate serverUrl is not pointing to internal/metadata services.
 */
export function validateServerUrl(url: string): void {
  let parsed: URL
  try {
    parsed = new URL(url)
  } catch {
    throw new Error(`Invalid server URL: ${url}`)
  }

  // Block known metadata endpoints
  if (BLOCKED_HOSTS.includes(parsed.hostname as (typeof BLOCKED_HOSTS)[number])) {
    throw new Error(`Blocked host: ${parsed.hostname}`)
  }

  // Block private IP ranges (basic check)
  const hostname = parsed.hostname
  if (
    hostname.startsWith("10.") ||
    hostname.startsWith("192.168.") ||
    hostname.match(/^172\.(1[6-9]|2[0-9]|3[0-1])\./)
  ) {
    throw new Error(`Private IP not allowed: ${hostname}`)
  }

  // Require HTTPS in production
  if (process.env.NODE_ENV === "production" && parsed.protocol !== "https:") {
    throw new Error("HTTPS required for worker URLs in production")
  }
}

function deepMerge<T extends Record<string, unknown>>(target: T, source: Partial<T>): T {
  const result = { ...target }

  for (const key of Object.keys(source)) {
    const sourceValue = source[key as keyof T]
    const targetValue = target[key as keyof T]

    if (
      sourceValue !== null &&
      typeof sourceValue === "object" &&
      !Array.isArray(sourceValue) &&
      targetValue !== null &&
      typeof targetValue === "object" &&
      !Array.isArray(targetValue)
    ) {
      result[key as keyof T] = deepMerge(
        targetValue as Record<string, unknown>,
        sourceValue as Record<string, unknown>,
      ) as T[keyof T]
    } else if (sourceValue !== undefined) {
      result[key as keyof T] = sourceValue as T[keyof T]
    }
  }

  return result
}

/**
 * Resolve a worker profile entry (string ID or full object).
 * Protected profiles cannot be overridden by config.
 */
function resolveProfileEntry(
  entry: string | Partial<WorkerProfile>,
  existingProfiles: Record<string, WorkerProfile>,
): WorkerProfile | undefined {
  if (typeof entry === "string") {
    return existingProfiles[entry] ?? builtInProfiles[entry]
  }

  if (!entry.id) return undefined

  // Validate ID format (alphanumeric + hyphen only)
  if (!/^[a-z0-9-]+$/i.test(entry.id)) {
    console.warn(`[Orchestrator] Invalid profile ID format: ${entry.id}`)
    return undefined
  }

  // Protect built-in profiles from override
  if (PROTECTED_PROFILE_IDS.includes(entry.id as (typeof PROTECTED_PROFILE_IDS)[number])) {
    console.warn(`[Orchestrator] Cannot override protected profile: ${entry.id}`)
    return builtInProfiles[entry.id]
  }

  const base = existingProfiles[entry.id] ?? builtInProfiles[entry.id]
  if (base) {
    return { ...base, ...entry, id: entry.id }
  }

  // New custom profile - validate required fields
  if (!entry.name || !entry.model || !entry.purpose || !entry.whenToUse) {
    return undefined
  }

  // Explicit construction instead of type assertion
  return {
    id: entry.id,
    name: entry.name,
    model: entry.model,
    purpose: entry.purpose,
    whenToUse: entry.whenToUse,
    kind: entry.kind,
    systemPrompt: entry.systemPrompt,
    supportsVision: entry.supportsVision,
    supportsWeb: entry.supportsWeb,
    tools: entry.tools,
    temperature: entry.temperature,
    tags: entry.tags,
  }
}

export interface LoadedOrchestratorConfig {
  config: OrchestratorConfigFile
  profiles: Record<string, WorkerProfile>
  spawn: string[]
  sources: { global?: string; project?: string }
}

/**
 * Load orchestrator configuration with three-level merging.
 */
export function loadOrchestratorConfig(directory: string): LoadedOrchestratorConfig {
  const sources: LoadedOrchestratorConfig["sources"] = {}

  // Layer 1: Built-in defaults
  let merged = { ...DEFAULT_CONFIG }

  // Layer 2: Global config
  const globalPath = join(getUserConfigDir(), "orchestrator.json")
  const globalRaw = tryReadJson(globalPath) // No project root check for global
  if (globalRaw) {
    sources.global = globalPath
    const parsed = OrchestratorConfigFileSchema.safeParse(globalRaw)
    if (parsed.success) {
      merged = deepMerge(merged, parsed.data)
    }
  }

  // Layer 3: Project config (with symlink protection)
  const projectPaths = [
    join(directory, ".ring", "orchestrator.json"),
    join(directory, ".opencode", "orchestrator.json"),
  ]

  for (const projectPath of projectPaths) {
    const projectRaw = tryReadJson(projectPath, directory) // Pass directory for symlink check
    if (projectRaw) {
      sources.project = projectPath
      const parsed = OrchestratorConfigFileSchema.safeParse(projectRaw)
      if (parsed.success) {
        merged = deepMerge(merged, parsed.data)
      }
      break
    }
  }

  // Build profiles map
  const profiles: Record<string, WorkerProfile> = { ...builtInProfiles }
  const spawn: string[] = []
  const seenSpawn = new Set<string>()

  // Process profile definitions
  for (const entry of merged.profiles) {
    const resolved = resolveProfileEntry(entry, profiles)
    if (resolved) {
      profiles[resolved.id] = resolved
    }
  }

  // Process workers to spawn
  for (const entry of merged.workers) {
    if (typeof entry === "string") {
      if (profiles[entry] && !seenSpawn.has(entry)) {
        spawn.push(entry)
        seenSpawn.add(entry)
      }
    } else {
      const resolved = resolveProfileEntry(entry, profiles)
      if (resolved) {
        profiles[resolved.id] = resolved
        if (!seenSpawn.has(resolved.id)) {
          spawn.push(resolved.id)
          seenSpawn.add(resolved.id)
        }
      }
    }
  }

  return { config: merged, profiles, spawn, sources }
}

/**
 * Check if we're running inside a worker (to prevent recursive spawn).
 */
export function isInsideWorker(config: OrchestratorConfigFile): boolean {
  if (!config.security.preventRecursiveSpawn) return false
  return process.env[config.security.recursiveSpawnEnvVar] === "true"
}

/**
 * Create orchestrator context from loaded config.
 */
export function createOrchestratorContext(
  directory: string,
  loaded: LoadedOrchestratorConfig,
): OrchestratorContext {
  return {
    directory,
    profiles: loaded.profiles,
    spawnDefaults: {
      basePort: loaded.config.basePort,
      timeout: loaded.config.startupTimeout,
    },
    defaultListFormat: loaded.config.ui.defaultListFormat,
  }
}

/**
 * Get workflow limits from config.
 */
export function getWorkflowLimits(config: OrchestratorConfigFile): WorkflowRunLimits {
  return {
    maxSteps: config.workflows.limits.maxSteps,
    maxTaskChars: config.workflows.limits.maxTaskChars,
    maxCarryChars: config.workflows.limits.maxCarryChars,
    perStepTimeoutMs: config.workflows.limits.perStepTimeoutMs,
  }
}
