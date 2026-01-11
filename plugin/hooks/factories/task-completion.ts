/**
 * Task Completion Hook Factory
 *
 * Monitors todo list updates and triggers notifications on completion.
 * Persists todo state for tracking across sessions.
 */

import type { Hook, HookContext, HookFactory, HookOutput, HookResult } from "../types.js"
import { readState, writeState } from "../../utils/state.js"

/**
 * Configuration for task completion hook.
 */
export interface TaskCompletionConfig {
  /** Enable task completion notifications */
  notifyOnComplete?: boolean
  /** Enable toast display on completion */
  showToast?: boolean
  /** Persist todos state */
  persistState?: boolean
}

/** Default configuration */
const DEFAULT_CONFIG: Required<TaskCompletionConfig> = {
  notifyOnComplete: true,
  showToast: true,
  persistState: true,
}

/** State key for todos persistence */
const TODOS_STATE_KEY = "todos"

/**
 * Todo item structure from OpenCode.
 */
interface TodoItem {
  id?: string
  content: string
  status: "pending" | "in_progress" | "completed"
  activeForm?: string
}

/**
 * Persisted todos state.
 */
interface TodosState {
  todos: TodoItem[]
  lastUpdated: string
  completedCount: number
  totalCount: number
}

/**
 * Check if all todos are completed.
 */
function areAllTodosComplete(todos: TodoItem[]): boolean {
  if (todos.length === 0) {
    return false
  }
  return todos.every(todo => todo.status === "completed")
}

/**
 * Count completed todos.
 */
function countCompleted(todos: TodoItem[]): number {
  return todos.filter(todo => todo.status === "completed").length
}

/**
 * Extract todos from event properties.
 */
function extractTodos(properties?: Record<string, unknown>): TodoItem[] {
  if (!properties) {
    return []
  }

  // Handle different possible property structures
  const todosData = properties.todos ?? properties.items ?? properties.list

  if (!Array.isArray(todosData)) {
    return []
  }

  return todosData.map((item: unknown) => {
    if (typeof item === "object" && item !== null) {
      const todoItem = item as Record<string, unknown>
      return {
        id: typeof todoItem.id === "string" ? todoItem.id : undefined,
        content: typeof todoItem.content === "string" ? todoItem.content : String(todoItem.content ?? ""),
        status: isValidStatus(todoItem.status) ? todoItem.status : "pending",
        activeForm: typeof todoItem.activeForm === "string" ? todoItem.activeForm : undefined,
      }
    }
    return {
      content: String(item),
      status: "pending" as const,
    }
  })
}

/**
 * Type guard for valid todo status.
 */
function isValidStatus(status: unknown): status is "pending" | "in_progress" | "completed" {
  return status === "pending" || status === "in_progress" || status === "completed"
}

/**
 * Compare two todo lists to detect completion transition.
 */
function detectCompletionTransition(
  previousTodos: TodoItem[],
  currentTodos: TodoItem[]
): { wasJustCompleted: boolean; newlyCompleted: number } {
  const prevComplete = areAllTodosComplete(previousTodos)
  const currComplete = areAllTodosComplete(currentTodos)

  const prevCompletedCount = countCompleted(previousTodos)
  const currCompletedCount = countCompleted(currentTodos)

  return {
    wasJustCompleted: !prevComplete && currComplete && currentTodos.length > 0,
    newlyCompleted: Math.max(0, currCompletedCount - prevCompletedCount),
  }
}

/**
 * Create a task completion hook.
 */
export const createTaskCompletionHook: HookFactory<TaskCompletionConfig> = (
  config?: TaskCompletionConfig
): Hook => {
  const cfg = { ...DEFAULT_CONFIG, ...config }

  return {
    name: "task-completion",
    lifecycles: ["todo.updated"],
    priority: 100,
    enabled: true,

    async execute(ctx: HookContext, output: HookOutput): Promise<HookResult> {
      try {
        // Extract current todos from event
        const currentTodos = extractTodos(ctx.event?.properties)

        if (currentTodos.length === 0) {
          // Persist empty state to clear stale data instead of early return
          if (cfg.persistState) {
            const emptyState: TodosState = {
              todos: [],
              lastUpdated: new Date().toISOString(),
              completedCount: 0,
              totalCount: 0,
            }
            writeState(ctx.directory, TODOS_STATE_KEY, emptyState, ctx.sessionId)
          }
          return {
            success: true,
            data: { skipped: false, todosCleared: true },
          }
        }

        // Load previous state
        let previousState: TodosState | null = null
        if (cfg.persistState) {
          previousState = readState<TodosState>(ctx.directory, TODOS_STATE_KEY, ctx.sessionId)
        }

        const previousTodos = previousState?.todos ?? []

        // Detect completion transition
        const { wasJustCompleted, newlyCompleted } = detectCompletionTransition(
          previousTodos,
          currentTodos
        )

        // Persist current state
        if (cfg.persistState) {
          const newState: TodosState = {
            todos: currentTodos,
            lastUpdated: new Date().toISOString(),
            completedCount: countCompleted(currentTodos),
            totalCount: currentTodos.length,
          }
          writeState(ctx.directory, TODOS_STATE_KEY, newState, ctx.sessionId)
        }

        // Check if all tasks just completed
        if (wasJustCompleted && cfg.notifyOnComplete) {
          // Add completion notification to chain data for notification hook
          return {
            success: true,
            data: {
              allTasksComplete: true,
              totalTasks: currentTodos.length,
              newlyCompleted,
              triggerNotification: cfg.showToast,
            },
          }
        }

        return {
          success: true,
          data: {
            allTasksComplete: areAllTodosComplete(currentTodos),
            completedCount: countCompleted(currentTodos),
            totalCount: currentTodos.length,
            newlyCompleted,
          },
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error)
        return {
          success: false,
          error: `Task completion hook failed: ${errorMessage}`,
        }
      }
    },
  }
}

/**
 * Hook registry entry for task completion.
 */
export const taskCompletionEntry = {
  name: "task-completion" as const,
  factory: createTaskCompletionHook,
  defaultEnabled: true,
  description: "Monitors todo updates and notifies on all tasks completion",
}
