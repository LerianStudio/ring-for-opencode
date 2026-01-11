import type { Plugin } from "@opencode-ai/plugin"
import { readState, writeState, getSessionId } from "./utils/state"
import { EVENTS } from "./utils/events"

/**
 * Ring Outcome Inference Plugin
 * Infers session outcome from todo state.
 *
 * Equivalent to: default/lib/outcome-inference/outcome_inference.py
 * Event: session.idle
 *
 * This plugin runs when a session goes idle and infers the outcome
 * based on todo completion status. The outcome is saved for
 * learning extraction to pick up.
 *
 * NOTE: outcome-inference and learning-extract both use session.idle.
 * learning-extract should run after outcome-inference completes.
 */

// Outcome constants
const OUTCOME = {
  SUCCEEDED: "SUCCEEDED",
  PARTIAL_PLUS: "PARTIAL_PLUS",
  PARTIAL_MINUS: "PARTIAL_MINUS",
  FAILED: "FAILED",
  UNKNOWN: "UNKNOWN",
} as const

type OutcomeType = (typeof OUTCOME)[keyof typeof OUTCOME]

interface Todo {
  content: string
  status: "pending" | "in_progress" | "completed"
}

interface OutcomeResult {
  outcome: OutcomeType
  reason: string
  confidence: "low" | "medium" | "high"
  sources: string[]
}

interface SessionOutcome {
  status: OutcomeType
  summary: string
  confidence: string
  inferred_at: string
  key_decisions: string[]
}

/**
 * Analyze todos and return counts.
 */
function analyzeTodos(todos: Todo[]): { total: number; completed: number; inProgress: number; pending: number } {
  if (!todos || todos.length === 0) {
    return { total: 0, completed: 0, inProgress: 0, pending: 0 }
  }

  return {
    total: todos.length,
    completed: todos.filter((t) => t.status === "completed").length,
    inProgress: todos.filter((t) => t.status === "in_progress").length,
    pending: todos.filter((t) => t.status === "pending").length,
  }
}

/**
 * Infer outcome from todo counts.
 */
function inferOutcomeFromTodos(
  total: number,
  completed: number,
  inProgress: number,
  pending: number
): { outcome: OutcomeType; reason: string } {
  if (total === 0) {
    return { outcome: OUTCOME.FAILED, reason: "No todos tracked - unable to measure completion" }
  }

  const completionRatio = completed / total

  if (completed === total) {
    return { outcome: OUTCOME.SUCCEEDED, reason: `All ${total} tasks completed` }
  }

  if (completionRatio >= 0.8) {
    const remaining = total - completed
    return { outcome: OUTCOME.PARTIAL_PLUS, reason: `${completed}/${total} tasks done, ${remaining} minor items remain` }
  }

  if (completionRatio >= 0.5) {
    return { outcome: OUTCOME.PARTIAL_MINUS, reason: `${completed}/${total} tasks done, significant work remains` }
  }

  // <50% completion is FAILED
  return {
    outcome: OUTCOME.FAILED,
    reason: `Insufficient progress: only ${completed}/${total} tasks completed (${Math.floor(completionRatio * 100)}%)`,
  }
}

/**
 * Infer outcome from available state.
 */
function inferOutcome(todos: Todo[] | null): OutcomeResult {
  const result: OutcomeResult = {
    outcome: OUTCOME.UNKNOWN,
    reason: "Unable to determine outcome",
    confidence: "low",
    sources: [],
  }

  if (todos !== null) {
    const { total, completed, inProgress, pending } = analyzeTodos(todos)
    const inferred = inferOutcomeFromTodos(total, completed, inProgress, pending)

    result.outcome = inferred.outcome
    result.reason = inferred.reason
    result.sources.push("todos")

    // Confidence based on completion clarity
    if (total === 0) {
      result.confidence = "high" // Empty todos = high confidence it's a failure
    } else if (completed === total || completed === 0) {
      result.confidence = "high"
    } else if (completed / total >= 0.8 || completed / total <= 0.2) {
      result.confidence = "medium"
    } else {
      result.confidence = "low"
    }
  }

  return result
}

// Export for testing
export { analyzeTodos, inferOutcomeFromTodos, OUTCOME }
export type { Todo, OutcomeType }

export const RingOutcomeInference: Plugin = async ({ directory }) => {
  const projectRoot = directory
  const sessionId = getSessionId()

  const IDEMPOTENCY_WINDOW_MS = 10_000

  return {
    event: async ({ event }) => {
      if (event.type !== EVENTS.SESSION_IDLE) {
        return
      }

      const alreadyRan = readState<{ timestamp: number }>(projectRoot, "outcome-inference-ran", sessionId)
      if (alreadyRan && Date.now() - alreadyRan.timestamp < IDEMPOTENCY_WINDOW_MS) {
        return
      }


      // Read todos state
      const todos = readState<Todo[]>(projectRoot, "todos-state", sessionId)

      // Infer outcome
      const result = inferOutcome(todos)

      // Save outcome for learning extraction
      const outcome: SessionOutcome = {
        status: result.outcome,
        summary: result.reason,
        confidence: result.confidence,
        inferred_at: new Date().toISOString(),
        key_decisions: [],
      }

      writeState(projectRoot, "session-outcome", outcome, sessionId)
      writeState(projectRoot, "outcome-inference-ran", { timestamp: Date.now() }, sessionId)
    },
  }
}

export default RingOutcomeInference
