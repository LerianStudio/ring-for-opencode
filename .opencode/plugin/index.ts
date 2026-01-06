/**
 * Ring OpenCode Plugins
 *
 * This module exports all Ring plugins for OpenCode.
 * Plugins provide session management, context injection,
 * notifications, security features, and workflow automation.
 */

// Core plugins
export { RingSessionStart } from "./session-start"
export { RingContextInjection } from "./context-injection"
export { RingNotification } from "./notification"

// Workflow plugins
export { RingTaskCompletionCheck } from "./task-completion-check"
export { RingSessionOutcome } from "./session-outcome"

// Session analytics plugins
export { RingOutcomeInference } from "./outcome-inference"

// Interactive tools
export { RingDoubtResolver } from "./doubt-resolver"

// Default export for convenience
export { RingSessionStart as default } from "./session-start"
