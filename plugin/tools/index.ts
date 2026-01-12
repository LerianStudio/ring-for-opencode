/**
 * Ring Tools
 *
 * Custom tools registered by Ring plugin.
 * Currently a placeholder - orchestrator tools have been removed.
 */

/**
 * Create Ring tools (currently returns empty object).
 * Orchestrator and background task tools have been removed for simplification.
 */
export function createRingTools(_directory: string, _options: Record<string, unknown> = {}) {
  return {}
}

/**
 * Legacy export for backwards compatibility.
 * @deprecated Use createRingTools(directory) instead.
 */
export const ringTools = {}
