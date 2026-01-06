/**
 * Message tracking utilities for Ring OpenCode plugins.
 *
 * OpenCode doesn't have a UserPromptSubmit equivalent, so we track
 * message counts via message.updated events to throttle periodic
 * context injection (like claude-md-reminder and context-usage-check).
 *
 * IMPORTANT: message.updated fires for both user and assistant messages.
 * Plugins should filter to only count user messages where appropriate.
 */

import { readState, writeState } from "./state"

interface MessageTrackingState {
  messageCount: number
  lastInjectionAt: number
  acknowledgedTier: string
}

const DEFAULT_STATE: MessageTrackingState = {
  messageCount: 0,
  lastInjectionAt: 0,
  acknowledgedTier: "none",
}

/**
 * Get the current message tracking state.
 */
export function getMessageState(projectRoot: string, key: string, sessionId?: string): MessageTrackingState {
  const state = readState<MessageTrackingState>(projectRoot, `msg-${key}`, sessionId)
  return state || { ...DEFAULT_STATE }
}

/**
 * Increment message count and optionally update last injection time.
 */
export function incrementMessageCount(
  projectRoot: string,
  key: string,
  sessionId?: string,
  markInjection: boolean = false
): MessageTrackingState {
  const state = getMessageState(projectRoot, key, sessionId)
  state.messageCount++

  if (markInjection) {
    state.lastInjectionAt = state.messageCount
  }

  writeState(projectRoot, `msg-${key}`, state, sessionId)
  return state
}

/**
 * Check if we should inject based on throttle interval.
 */
export function shouldInject(projectRoot: string, key: string, throttleInterval: number, sessionId?: string): boolean {
  const state = getMessageState(projectRoot, key, sessionId)
  const messagesSinceInjection = state.messageCount - state.lastInjectionAt
  return messagesSinceInjection >= throttleInterval
}

/**
 * Update acknowledged tier for tiered warnings.
 */
export function updateAcknowledgedTier(projectRoot: string, key: string, tier: string, sessionId?: string): void {
  const state = getMessageState(projectRoot, key, sessionId)
  state.acknowledgedTier = tier
  writeState(projectRoot, `msg-${key}`, state, sessionId)
}

/**
 * Reset message tracking state.
 */
export function resetMessageState(projectRoot: string, key: string, sessionId?: string): void {
  writeState(projectRoot, `msg-${key}`, { ...DEFAULT_STATE }, sessionId)
}
