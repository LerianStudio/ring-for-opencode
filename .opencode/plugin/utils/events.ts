/**
 * OpenCode event type constants used by Ring plugins.
 * Keeps event strings consistent across plugins.
 */

export const EVENTS = {
  MESSAGE_UPDATED: "message.updated",
  SESSION_COMPACTED: "session.compacted",
  SESSION_CREATED: "session.created",
  SESSION_IDLE: "session.idle",
  TODO_UPDATED: "todo.updated",
} as const
