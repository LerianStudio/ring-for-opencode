/**
 * Ring Background Task Types
 *
 * Type definitions for background agent execution system.
 */

/**
 * Background task execution status.
 */
export type BackgroundTaskStatus =
  | "pending"
  | "running"
  | "completed"
  | "error"
  | "cancelled"
  | "timeout"

/**
 * Task progress information.
 */
export interface TaskProgress {
  /** Number of tool calls made */
  toolCalls: number
  /** Last tool that was called */
  lastTool?: string
  /** Last progress update timestamp */
  lastUpdate: Date
  /** Last message from the task */
  lastMessage?: string
  /** Timestamp of last message */
  lastMessageAt?: Date
}

/**
 * Background task definition.
 */
export interface BackgroundTask {
  /** Unique task identifier */
  id: string
  /** OpenCode session ID for this task */
  sessionId: string
  /** Parent session that launched this task */
  parentSessionId: string
  /** Human-readable description */
  description: string
  /** The prompt sent to the agent */
  prompt: string
  /** Agent name to execute the task */
  agent: string
  /** Current task status */
  status: BackgroundTaskStatus
  /** When the task was created */
  createdAt: Date
  /** When the task started running */
  startedAt?: Date
  /** When the task completed */
  completedAt?: Date
  /** Task result if completed successfully */
  result?: string
  /** Error message if failed */
  error?: string
  /** Progress tracking */
  progress?: TaskProgress
  /** Model configuration */
  model?: { providerId: string; modelId: string }
  /** Concurrency tracking key */
  concurrencyKey?: string
  /** Task type for categorization */
  taskType?: BackgroundTaskType
}

/**
 * Background task types.
 */
export type BackgroundTaskType =
  | "exploration"    // Codebase exploration
  | "review"         // Code review
  | "validation"     // Test/lint validation
  | "generation"     // Code generation
  | "analysis"       // Analysis task
  | "custom"         // Custom task

/**
 * Input for launching a new background task.
 */
export interface LaunchTaskInput {
  /** Human-readable description */
  description: string
  /** Prompt to send to the agent */
  prompt: string
  /** Agent name to execute */
  agent: string
  /** Parent session ID */
  parentSessionId: string
  /** Optional model override */
  model?: { providerId: string; modelId: string }
  /** Task type for categorization */
  taskType?: BackgroundTaskType
  /** Additional skills to inject */
  skills?: string[]
  /** Skill content to inject into system prompt */
  skillContent?: string
}

/**
 * Input for resuming a background task.
 */
export interface ResumeTaskInput {
  /** Task ID to resume */
  taskId: string
  /** New prompt for continuation */
  prompt: string
  /** New parent session */
  parentSessionId: string
}

/**
 * Background task notification payload.
 */
export interface TaskNotification {
  /** Task that triggered notification */
  task: BackgroundTask
  /** Notification type */
  type: "started" | "progress" | "completed" | "error"
  /** Human-readable message */
  message: string
  /** Timestamp */
  timestamp: Date
}

/**
 * Background manager events.
 */
export interface BackgroundManagerEvents {
  "task:started": (task: BackgroundTask) => void
  "task:progress": (task: BackgroundTask) => void
  "task:completed": (task: BackgroundTask) => void
  "task:error": (task: BackgroundTask, error: Error) => void
  "task:cancelled": (task: BackgroundTask) => void
}

/**
 * OpenCode client interface (subset used by background manager).
 */
export interface BackgroundClient {
  session: {
    create: (opts: { body: { parentID?: string; title?: string } }) => Promise<{ data: { id: string }; error?: unknown }>
    prompt: (opts: {
      path: { id: string }
      body: {
        agent?: string
        model?: { providerId: string; modelId: string }
        system?: string
        tools?: Record<string, boolean>
        parts: Array<{ type: string; text: string }>
        noReply?: boolean
      }
    }) => Promise<void>
    messages: (opts: { path: { id: string } }) => Promise<{ data?: Array<unknown>; error?: unknown }>
    status: () => Promise<{ data?: Record<string, { type: string }>; error?: unknown }>
    todo: (opts: { path: { id: string } }) => Promise<{ data?: Array<{ status: string }> }>
  }
  tui: {
    showToast: (opts: {
      body: { title: string; message: string; variant: string; duration: number }
    }) => Promise<void>
  }
}
