/**
 * Ring State Management Utilities
 *
 * Provides state persistence and utility functions for hooks.
 */

import * as fs from "node:fs"
import * as path from "node:path"
import { randomBytes } from "node:crypto"

/** State directory name within .ring */
const STATE_DIR = ".ring/state"

/**
 * Get the state directory path for a project.
 */
function getStateDir(directory: string): string {
  return path.join(directory, STATE_DIR)
}

/**
 * Get the state file path for a key and session.
 */
function getStateFilePath(directory: string, key: string, sessionId: string): string {
  const stateDir = getStateDir(directory)
  const sanitizedKey = key.replace(/[^a-zA-Z0-9-_]/g, "_")
  const sanitizedSession = sessionId.replace(/[^a-zA-Z0-9-_]/g, "_")
  return path.join(stateDir, `${sanitizedKey}-${sanitizedSession}.json`)
}

/**
 * Clean up old state files beyond max age.
 */
export function cleanupOldState(directory: string, maxAgeDays: number): void {
  const stateDir = getStateDir(directory)

  if (!fs.existsSync(stateDir)) {
    return
  }

  const maxAgeMs = maxAgeDays * 24 * 60 * 60 * 1000
  const now = Date.now()

  try {
    const files = fs.readdirSync(stateDir)

    for (const file of files) {
      if (!file.endsWith(".json")) {
        continue
      }

      const filePath = path.join(stateDir, file)
      const stats = fs.statSync(filePath)
      const age = now - stats.mtimeMs

      if (age > maxAgeMs) {
        fs.unlinkSync(filePath)
      }
    }
  } catch (error) {
    // Log cleanup errors in debug mode
    if (process.env.RING_DEBUG) {
      console.debug(`[ring] State cleanup failed:`, error)
    }
  }
}

/**
 * Delete a specific state file.
 */
export function deleteState(directory: string, key: string, sessionId: string): void {
  const filePath = getStateFilePath(directory, key, sessionId)

  try {
    if (fs.existsSync(filePath)) {
      fs.unlinkSync(filePath)
    }
  } catch (error) {
    // Log delete errors in debug mode
    if (process.env.RING_DEBUG) {
      console.debug(`[ring] State delete failed:`, error)
    }
  }
}

/**
 * Get or generate a session ID.
 * Uses environment variable if available, otherwise generates one.
 */
export function getSessionId(): string {
  // Check for OpenCode session ID in environment
  const envSessionId = process.env.OPENCODE_SESSION_ID

  if (envSessionId) {
    return envSessionId
  }

  // Generate a new session ID based on timestamp and cryptographically secure random suffix
  const timestamp = Date.now().toString(36)
  const random = randomBytes(6).toString("hex")
  return `session-${timestamp}-${random}`
}

/**
 * Escape angle brackets to prevent XML/HTML injection in prompts.
 */
export function escapeAngleBrackets(str: string): string {
  return str
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
}

/**
 * Sanitize a string for safe inclusion in prompts.
 * Escapes special characters and truncates to max length.
 */
export function sanitizeForPrompt(str: string, maxLength: number): string {
  let sanitized = str
    // Remove null bytes
    .replace(/\0/g, "")
    // Escape angle brackets
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    // Normalize whitespace
    .replace(/\r\n/g, "\n")
    .replace(/\r/g, "\n")

  // Truncate if needed
  if (sanitized.length > maxLength) {
    sanitized = sanitized.substring(0, maxLength - 3) + "..."
  }

  return sanitized
}

/**
 * Check if a path is within a root directory.
 * Prevents path traversal attacks.
 */
export function isPathWithinRoot(targetPath: string, rootPath: string): boolean {
  const resolvedTarget = path.resolve(targetPath)
  const resolvedRoot = path.resolve(rootPath)

  // Ensure root ends with separator for proper prefix matching
  const rootWithSep = resolvedRoot.endsWith(path.sep)
    ? resolvedRoot
    : resolvedRoot + path.sep

  return resolvedTarget === resolvedRoot || resolvedTarget.startsWith(rootWithSep)
}

/**
 * Write state data to a file.
 */
export function writeState(
  directory: string,
  key: string,
  data: unknown,
  sessionId: string
): void {
  const stateDir = getStateDir(directory)
  const filePath = getStateFilePath(directory, key, sessionId)

  try {
    // Ensure state directory exists
    if (!fs.existsSync(stateDir)) {
      fs.mkdirSync(stateDir, { recursive: true })
    }

    // Write state with timestamp
    const stateData = {
      key,
      sessionId,
      timestamp: new Date().toISOString(),
      data,
    }

    fs.writeFileSync(filePath, JSON.stringify(stateData, null, 2), {
      encoding: "utf-8",
      mode: 0o600,  // Owner read/write only
    })
  } catch (error) {
    // Log write errors in debug mode
    if (process.env.RING_DEBUG) {
      console.debug(`[ring] State write failed:`, error)
    }
  }
}

/**
 * Read state data from a file.
 */
export function readState<T = unknown>(
  directory: string,
  key: string,
  sessionId: string
): T | null {
  const filePath = getStateFilePath(directory, key, sessionId)

  try {
    if (!fs.existsSync(filePath)) {
      return null
    }

    const content = fs.readFileSync(filePath, "utf-8")
    const stateData = JSON.parse(content)
    return stateData.data as T
  } catch (error) {
    // Log read errors in debug mode
    if (process.env.RING_DEBUG) {
      console.debug(`[ring] State read failed:`, error)
    }
    return null
  }
}

/**
 * Find the most recent file matching a pattern in a directory.
 */
export function findMostRecentFile(
  directory: string,
  pattern: RegExp
): string | null {
  try {
    if (!fs.existsSync(directory)) {
      return null
    }

    const files = fs.readdirSync(directory)
    let mostRecent: { path: string; mtime: number } | null = null

    for (const file of files) {
      if (!pattern.test(file)) {
        continue
      }

      const filePath = path.join(directory, file)
      const stats = fs.statSync(filePath)

      if (!mostRecent || stats.mtimeMs > mostRecent.mtime) {
        mostRecent = { path: filePath, mtime: stats.mtimeMs }
      }
    }

    return mostRecent?.path ?? null
  } catch {
    return null
  }
}

/**
 * Read file contents safely.
 */
export function readFileSafe(filePath: string): string | null {
  try {
    if (!fs.existsSync(filePath)) {
      return null
    }
    return fs.readFileSync(filePath, "utf-8")
  } catch {
    return null
  }
}
