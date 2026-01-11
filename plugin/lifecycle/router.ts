/**
 * Ring Lifecycle Router
 *
 * Routes OpenCode lifecycle events to Ring's plugin handlers.
 * Maps OpenCode events to Ring hook registry.
 */

import type { RingConfig } from "../config/index.js"
import { cleanupOldState, deleteState, getSessionId } from "../utils/state.js"

/**
 * OpenCode event structure.
 */
export interface OpenCodeEvent {
  type: string
  properties?: Record<string, unknown>
}

/**
 * Dependencies for lifecycle router.
 */
export interface LifecycleRouterDeps {
  projectRoot: string
  ringConfig: RingConfig
}

/**
 * Create the event handler that routes lifecycle events.
 *
 * Event mappings:
 * - session.created -> Reset context state, cleanup old files
 * - session.idle -> Hook execution (via RingUnifiedPlugin)
 * - session.error -> Hook execution (via RingUnifiedPlugin)
 * - todo.updated -> Task completion hook
 * - experimental.session.compacting -> Context injection
 */
export function createLifecycleRouter(deps: LifecycleRouterDeps) {
  const { projectRoot, ringConfig } = deps
  const debug = process.env.RING_DEBUG === "true"

  if (debug) {
    console.debug("[ring] Lifecycle router initialized", {
      disabledAgents: ringConfig.disabled_agents?.length ?? 0,
      disabledSkills: ringConfig.disabled_skills?.length ?? 0,
    })
  }

  return async (input: { event: OpenCodeEvent }): Promise<void> => {
    const { event } = input
    const eventType = event.type

    if (debug) {
      console.debug(`[ring] Event: ${eventType}`)
    }

    // session.created - Initialize session
    if (eventType === "session.created") {
      const props = event.properties as Record<string, unknown> | undefined
      const eventSessionId = (props?.sessionID as string | undefined) ?? getSessionId()

      // Reset context usage state for new session
      deleteState(projectRoot, "context-usage", eventSessionId)
      // Clean up old state files (> 7 days)
      cleanupOldState(projectRoot, 7)

      if (debug) {
        console.debug("[ring] Session initialized, state reset")
      }
      return
    }

    // session.idle / session.error / todo.updated are handled via hook execution
    // in RingUnifiedPlugin (this router only manages state side effects).
    if (
      eventType === "session.idle" ||
      eventType === "session.error" ||
      eventType === "todo.updated"
    ) {
      return
    }

    // Other events - no action needed
  }
}

/**
 * Event type constants for type safety.
 */
export const EVENTS = {
  SESSION_CREATED: "session.created",
  SESSION_IDLE: "session.idle",
  SESSION_ERROR: "session.error",
  SESSION_DELETED: "session.deleted",
  TODO_UPDATED: "todo.updated",
  MESSAGE_PART_UPDATED: "message.part.updated",
} as const
