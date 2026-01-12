/**
 * Worker Profiles
 *
 * Built-in worker profiles and profile management utilities.
 * Supports model tag resolution (node:vision, node:fast, etc.)
 *
 * Ported from Orchestra's profiles pattern.
 */

import type { WorkerProfile } from "./types.js"

/**
 * Built-in worker profiles that can be used out of the box.
 */
export const builtInProfiles: Record<string, WorkerProfile> = {
  // Vision specialist - for analyzing images, diagrams, screenshots
  vision: {
    id: "vision",
    name: "Vision Analyst",
    kind: "server",
    model: "node:vision",
    purpose: "Analyze images, screenshots, diagrams, and visual content",
    whenToUse:
      "When you need to understand visual content like screenshots, architecture diagrams, UI mockups, error screenshots, or any image-based information",
    supportsVision: true,
  },

  // Documentation specialist - for looking up docs and examples
  docs: {
    id: "docs",
    name: "Documentation Librarian",
    kind: "server",
    model: "node:docs",
    purpose: "Research documentation, find examples, explain APIs and libraries",
    whenToUse:
      "When you need to look up official documentation, find code examples, understand library APIs, or research best practices",
    supportsWeb: true,
    tools: {
      write: false,
      edit: false,
    },
  },

  // Coding specialist - main implementation worker
  coder: {
    id: "coder",
    name: "Code Implementer",
    kind: "server",
    model: "node",
    purpose: "Write, edit, and refactor code with full tool access",
    whenToUse:
      "When you need to actually write or modify code, create files, run commands, or implement features",
  },

  // Architecture/planning specialist
  architect: {
    id: "architect",
    name: "System Architect",
    kind: "server",
    model: "node",
    purpose: "Design systems, plan implementations, review architecture decisions",
    whenToUse:
      "When you need to plan a complex feature, design system architecture, or make high-level technical decisions",
    tools: {
      write: false,
      edit: false,
      bash: false,
    },
  },

  // Fast explorer - for quick codebase searches
  explorer: {
    id: "explorer",
    name: "Code Explorer",
    kind: "server",
    model: "node:fast",
    purpose: "Quickly search and navigate the codebase",
    whenToUse:
      "When you need to quickly find files, search for patterns, or locate specific code without deep analysis",
    tools: {
      write: false,
      edit: false,
    },
    temperature: 0.1,
  },
}

/**
 * Get a profile by ID (built-in or custom).
 */
export function getProfile(
  id: string,
  customProfiles?: Record<string, WorkerProfile>,
): WorkerProfile | undefined {
  return customProfiles?.[id] ?? builtInProfiles[id]
}

/**
 * Merge custom profile with built-in defaults.
 */
export function mergeProfile(baseId: string, overrides: Partial<WorkerProfile>): WorkerProfile {
  const base = builtInProfiles[baseId]
  if (!base) {
    throw new Error(`Unknown base profile: ${baseId}`)
  }
  return {
    ...base,
    ...overrides,
    id: overrides.id ?? base.id,
  }
}

/**
 * Resolve a model tag to an actual model identifier.
 *
 * Tags supported:
 * - node: Use the default configured model
 * - node:vision: Use a vision-capable model
 * - node:fast: Use a fast/cheap model for quick tasks
 * - node:docs: Use a model optimized for documentation
 * - auto:vision, auto:fast, etc: Same as node: prefix
 *
 * If no tag match, returns the input as-is (assumed to be provider/model).
 */
export function resolveModelTag(
  tag: string,
  defaults: {
    default?: string
    vision?: string
    fast?: string
    docs?: string
  },
): string {
  const normalized = tag.toLowerCase().trim()

  // Check for node: or auto: prefix
  const prefixMatch = normalized.match(/^(node|auto):(.+)$/)
  if (prefixMatch) {
    const capability = prefixMatch[2]
    switch (capability) {
      case "vision":
        return defaults.vision ?? defaults.default ?? tag
      case "fast":
        return defaults.fast ?? defaults.default ?? tag
      case "docs":
        return defaults.docs ?? defaults.default ?? tag
      default:
        return defaults.default ?? tag
    }
  }

  // Check for bare "node" or "auto"
  if (normalized === "node" || normalized === "auto") {
    return defaults.default ?? tag
  }

  // Not a tag, return as-is
  return tag
}

/**
 * List all available profile IDs.
 */
export function listProfileIds(customProfiles?: Record<string, WorkerProfile>): string[] {
  const ids = new Set<string>(Object.keys(builtInProfiles))
  if (customProfiles) {
    for (const id of Object.keys(customProfiles)) {
      ids.add(id)
    }
  }
  return [...ids].sort()
}
