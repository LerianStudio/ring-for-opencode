/**
 * State management utilities for Ring OpenCode plugins.
 * Provides session-isolated state persistence in .opencode/state/
 */

import {
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
  unlinkSync,
  readdirSync,
  statSync,
  renameSync,
} from "fs"
import { join, dirname, resolve, relative, isAbsolute } from "path"
import { randomBytes } from "crypto"

// State directories - support both new and legacy locations
const STATE_DIRS = [".opencode/state", ".ring/state"]
// Primary state directory for writes
const STATE_DIR = ".opencode/state"

/**
 * Generate a cryptographically secure session ID.
 * Used as fallback when OPENCODE_SESSION_ID is not available.
 */
export function generateSecureSessionId(): string {
  return randomBytes(16).toString("hex")
}

// Global session ID cache to maintain consistency within a process
declare global {
  // eslint-disable-next-line no-var
  var _ringSecureSessionId: string | undefined
}

/**
 * Get the current session ID from environment or generate a secure one.
 * Supports both OPENCODE_SESSION_ID and CLAUDE_SESSION_ID for compatibility.
 */
export function getSessionId(): string {
  return (
    process.env.OPENCODE_SESSION_ID ||
    process.env.CLAUDE_SESSION_ID ||
    (globalThis._ringSecureSessionId ??= generateSecureSessionId())
  )
}

/**
 * Validate that a path is within the allowed root directory.
 * Prevents path traversal attacks.
 */
export function isPathWithinRoot(filePath: string, rootPath: string): boolean {
  const resolvedPath = resolve(filePath)
  const resolvedRoot = resolve(rootPath)
  const rel = relative(resolvedRoot, resolvedPath)
  return !rel.startsWith("..") && !isAbsolute(rel)
}

/**
 * Sanitize session ID for safe use in filenames.
 * Prevents path traversal attacks.
 */
export function sanitizeSessionId(sessionId: string): string {
  const sanitized = sessionId.replace(/[^a-zA-Z0-9_-]/g, "")
  if (!sanitized) return "unknown"
  return sanitized.slice(0, 64)
}

/**
 * Get the state directory path, creating it if needed.
 */
export function getStateDir(projectRoot: string): string {
  const stateDir = join(projectRoot, STATE_DIR)
  if (!existsSync(stateDir)) {
    mkdirSync(stateDir, { recursive: true })
  }
  return stateDir
}

/**
 * Sanitize state key for safe use in filenames.
 * Prevents path traversal attacks via key parameter.
 */
export function sanitizeKey(key: string): string {
  return key.replace(/[^a-zA-Z0-9_-]/g, "").slice(0, 64) || "default"
}

/**
 * Get a state file path for a specific key and session.
 */
export function getStatePath(projectRoot: string, key: string, sessionId?: string): string {
  const stateDir = getStateDir(projectRoot)
  const safeKey = sanitizeKey(key)
  const safeSession = sessionId ? sanitizeSessionId(sessionId) : "global"
  return join(stateDir, `${safeKey}-${safeSession}.json`)
}

/**
 * Find a state file across all supported state directories.
 * Reads from either location for backward compatibility.
 */
function findStatePath(projectRoot: string, key: string, sessionId?: string): string | null {
  const safeKey = sanitizeKey(key)
  const safeSession = sessionId ? sanitizeSessionId(sessionId) : "global"
  const filename = `${safeKey}-${safeSession}.json`

  for (const dir of STATE_DIRS) {
    const statePath = join(projectRoot, dir, filename)
    if (existsSync(statePath)) {
      return statePath
    }
  }
  return null
}

/**
 * Read state from file, returning null if not found.
 */
export function readState<T>(projectRoot: string, key: string, sessionId?: string): T | null {
  // First try to find existing file in any supported location
  const existingPath = findStatePath(projectRoot, key, sessionId)
  const statePath = existingPath || getStatePath(projectRoot, key, sessionId)

  try {
    if (!existsSync(statePath)) return null
    const content = readFileSync(statePath, "utf-8")
    return JSON.parse(content) as T
  } catch (err) {
    return null
  }
}

/**
 * Write state to file atomically with secure temp file.
 * Uses restrictive permissions (0o600) for sensitive data.
 */
export function writeState<T>(projectRoot: string, key: string, data: T, sessionId?: string): void {
  const statePath = getStatePath(projectRoot, key, sessionId)
  const randomSuffix = randomBytes(8).toString("hex")
  const tempPath = `${statePath}.tmp.${randomSuffix}`

  try {
    mkdirSync(dirname(statePath), { recursive: true })
    writeFileSync(tempPath, JSON.stringify(data, null, 2), { encoding: "utf-8", mode: 0o600 })
    // Atomic rename
    renameSync(tempPath, statePath)
  } catch (err) {
    // Clean up temp file on error
    try {
      unlinkSync(tempPath)
    } catch {
      // ignore
    }
    throw err
  }
}

/**
 * Delete state file.
 */
export function deleteState(projectRoot: string, key: string, sessionId?: string): void {
  const statePath = getStatePath(projectRoot, key, sessionId)
  try {
    if (existsSync(statePath)) {
      unlinkSync(statePath)
    }
  } catch (err) {
    if (process.env.DEBUG) {
      console.debug('[ring] State operation failed:', err)
    }
  }
}

/**
 * Clean up old state files (older than specified days).
 */
export function cleanupOldState(projectRoot: string, maxAgeDays: number = 7): void {
  const stateDir = getStateDir(projectRoot)
  const maxAgeMs = maxAgeDays * 24 * 60 * 60 * 1000
  const now = Date.now()

  try {
    const files = readdirSync(stateDir)
    for (const file of files) {
      if (!file.endsWith(".json")) continue
      const filePath = join(stateDir, file)
      const stats = statSync(filePath)
      if (now - stats.mtimeMs > maxAgeMs) {
        unlinkSync(filePath)
      }
    }
  } catch (err) {
    if (process.env.DEBUG) {
      console.debug('[ring] State operation failed:', err)
    }
  }
}

/**
 * JSON escape utility for safe string embedding.
 */
export function jsonEscape(str: string): string {
  return str
    .replace(/\\/g, "\\\\")
    .replace(/"/g, "\\\"")
    .replace(/\n/g, "\\n")
    .replace(/\r/g, "\\r")
    .replace(/\t/g, "\\t")
}

/**
 * Sanitize content for safe inclusion in prompts.
 * Prevents prompt injection attacks.
 */
export function sanitizeForPrompt(content: string, maxLength: number = 500): string {
  return content.replace(/</g, "&lt;").replace(/>/g, "&gt;").slice(0, maxLength)
}

/**
 * Escapes angle brackets to prevent prompt injection.
 *
 * This intentionally escapes ALL `<` and `>` characters rather than
 * attempting tag-aware parsing. It's safer and avoids bypasses.
 */
export function escapeAngleBrackets(content: string): string {
  return content.replace(/</g, "&lt;").replace(/>/g, "&gt;")
}

/**
 * Backwards-compatible alias.
 *
 * @deprecated Prefer `escapeAngleBrackets()`.
 */
export function escapeXmlTags(content: string): string {
  return escapeAngleBrackets(content)
}
