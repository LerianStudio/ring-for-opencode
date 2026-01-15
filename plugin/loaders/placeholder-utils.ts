/**
 * Placeholder Expansion Utilities
 *
 * Expands template placeholders in markdown content before passing to OpenCode.
 *
 * Supported placeholders:
 * - {OPENCODE_CONFIG} - Expands to the OpenCode config directory path
 */

import { homedir } from "node:os"
import { isAbsolute, join } from "node:path"

/**
 * Named constant for the placeholder string.
 * Using a constant improves maintainability and ensures consistency.
 */
export const OPENCODE_CONFIG_PLACEHOLDER = "{OPENCODE_CONFIG}"

/**
 * Get the OpenCode config directory path.
 *
 * Respects OPENCODE_CONFIG_DIR environment variable if set,
 * falls back to XDG_CONFIG_HOME/opencode if set and absolute,
 * otherwise defaults to ~/.config/opencode (following XDG standards).
 *
 * @returns The absolute path to OpenCode's config directory
 * @throws Error if home directory cannot be determined
 */
export function getOpenCodeConfigDir(): string {
  // Priority 1: Explicit OPENCODE_CONFIG_DIR
  if (process.env.OPENCODE_CONFIG_DIR) {
    return process.env.OPENCODE_CONFIG_DIR
  }

  // Priority 2: XDG_CONFIG_HOME (must be absolute path per XDG spec)
  // Matches validation in config/loader.ts:196
  const xdgConfigHome = process.env.XDG_CONFIG_HOME
  if (xdgConfigHome && isAbsolute(xdgConfigHome)) {
    return join(xdgConfigHome, "opencode")
  }

  // Priority 3: Default to ~/.config/opencode
  const home = homedir()
  if (!home) {
    throw new Error(
      "Cannot determine home directory. Set OPENCODE_CONFIG_DIR or HOME environment variable.",
    )
  }
  return join(home, ".config", "opencode")
}

/**
 * Expand placeholders in markdown content.
 *
 * Currently supports:
 * - {OPENCODE_CONFIG} -> actual config directory path
 *
 * @param content - The markdown content with placeholders
 * @returns Content with all placeholders expanded to their actual values
 *
 * @example
 * ```typescript
 * const content = "Read from {OPENCODE_CONFIG}/standards/golang.md"
 * const expanded = expandPlaceholders(content)
 * // Result: "Read from /Users/john/.config/opencode/standards/golang.md"
 * ```
 */
export function expandPlaceholders(content: string): string {
  // Input validation: handle null/undefined/non-string gracefully
  if (typeof content !== "string") {
    return ""
  }
  if (!content) {
    return ""
  }

  const configDir = getOpenCodeConfigDir()

  // Replace all occurrences of {OPENCODE_CONFIG} with the actual path
  return content.replace(new RegExp(OPENCODE_CONFIG_PLACEHOLDER, "g"), configDir)
}
