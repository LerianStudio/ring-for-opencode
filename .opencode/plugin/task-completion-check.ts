import type { Plugin } from "@opencode-ai/plugin"
import { writeState, getSessionId } from "./utils/state"
import { EVENTS } from "./utils/events"

/**
 * Ring Task Completion Check Plugin
 * Detects when all todos are complete and suggests handoff creation.
 *
 * Equivalent to: default/hooks/task-completion-check.sh
 * Event: todo.updated
 *
 * This plugin monitors todo list updates and:
 * 1. Persists todo state for outcome inference
 * 2. Injects handoff recommendation when all tasks complete
 */

interface Todo {
  content: string
  status: "pending" | "in_progress" | "completed"
  activeForm?: string
}

interface TodoEvent {
  type: "todo.updated"
  properties?: {
    todos?: Todo[]
  }
  data?: {
    todos?: Todo[]
  }
}

export const RingTaskCompletionCheck: Plugin = async ({ client, directory }) => {
  const projectRoot = directory

  return {
    event: async ({ event }) => {
      if (event.type !== EVENTS.TODO_UPDATED) {
        return
      }

      // Type guard for TodoEvent with defensive payload handling
      const todoEvent = event as unknown as TodoEvent
      const todos = todoEvent?.properties?.todos || todoEvent?.data?.todos || []

      // Runtime validation
      if (!Array.isArray(todos)) {
        return
      }

      if (todos.length === 0) {
        return
      }

      // Persist todos state for outcome inference
      const sessionId = getSessionId()
      writeState(projectRoot, "todos-state", todos, sessionId)

      // Count todos by status
      const total = todos.length
      const completed = todos.filter((t) => t.status === "completed").length
      const inProgress = todos.filter((t) => t.status === "in_progress").length
      const pending = todos.filter((t) => t.status === "pending").length

      // Check if all todos are complete
      if (total > 0 && completed === total) {
        // Show toast notification instead of injecting a fake user message
        await client.tui.showToast({
          body: {
            title: "All Tasks Complete!",
            message: `${total} tasks done. Consider /create-handoff if context is high.`,
            variant: "success",
            duration: 8000,
          },
        })

      } else if (completed > 0) {
        // Progress update - no action needed (silent)
      }
    },
  }
}

export default RingTaskCompletionCheck
