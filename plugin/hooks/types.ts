/**
 * Ring Hook System Type Definitions
 *
 * Defines hook lifecycle types and interfaces for middleware pattern.
 * Based on oh-my-opencode patterns adapted for Ring.
 */

/**
 * Hook lifecycle events supported by Ring.
 * Each event corresponds to a specific point in the OpenCode pipeline.
 */
export type HookLifecycle =
  | "session.created" // Session initialized
  | "session.idle" // Session became idle
  | "session.compacting" // Context being compacted
  | "chat.message" // User message received
  | "chat.params" // Chat parameters being set
  | "tool.before" // Before tool execution
  | "tool.after" // After tool execution
  | "todo.updated" // Todo list changed
  | "event" // Generic event handler

/**
 * Hook execution result indicating success/failure.
 */
export interface HookResult {
  success: boolean
  error?: string
  /** Optional data to pass to next hook in chain */
  data?: Record<string, unknown>
  /** If true, stop hook chain execution */
  stopChain?: boolean
}

/**
 * Context provided to all hooks.
 */
export interface HookContext {
  /** Session identifier */
  sessionId: string
  /** Project root directory */
  directory: string
  /** Hook lifecycle event that triggered this hook */
  lifecycle: HookLifecycle
  /** Original event data from OpenCode */
  event?: {
    type: string
    properties?: Record<string, unknown>
  }
  /** Data passed from previous hook in chain */
  chainData?: Record<string, unknown>
}

/**
 * Base hook interface that all hooks must implement.
 */
export interface Hook {
  /** Unique identifier for this hook */
  name: HookName
  /** Hook lifecycle events this hook responds to */
  lifecycles: HookLifecycle[]
  /** Priority for execution order (lower = earlier, default 100) */
  priority?: number
  /** Whether this hook is enabled */
  enabled: boolean
  /** Execute the hook */
  execute: (ctx: HookContext, output: HookOutput) => Promise<HookResult>
}

/**
 * Hook output object for modifying OpenCode behavior.
 */
export interface HookOutput {
  /** System prompt context to inject */
  system?: string[]
  /** Compaction context to inject */
  context?: string[]
  /** Message parts to modify */
  parts?: Array<{ type: string; text?: string; [key: string]: unknown }>
  /** Block the operation */
  block?: boolean
  /** Reason for blocking */
  blockReason?: string
}

/**
 * Hook factory function signature.
 * All hooks are created via factory functions for consistent initialization.
 */
export type HookFactory<TConfig = Record<string, unknown>> = (config?: TConfig) => Hook

/**
 * Built-in hook names supported by Ring.
 */
export type HookName =
  | "session-start"
  | "context-injection"
  | "notification"
  | "task-completion"
  | "session-outcome"
  | "outcome-inference"
  | "doubt-resolver"
  | "background-notification"
  | "compaction-context"
  | "rules-injector"
  | "agent-reminder"

/**
 * Hook registry entry with metadata.
 */
export interface HookRegistryEntry {
  name: HookName
  factory: HookFactory
  defaultEnabled: boolean
  description: string
}

/**
 * Event handler signature for hooks.
 */
export type HookEventHandler = (input: {
  event: { type: string; properties?: unknown }
}) => Promise<void>

/**
 * Chat message handler signature.
 */
export type HookChatHandler = (
  input: { sessionID: string; agent?: string },
  output: { parts: Array<{ type: string; text?: string }> },
) => Promise<void>

/**
 * Compaction handler signature.
 */
export type HookCompactionHandler = (
  input: { sessionID: string },
  output: { context: string[] },
) => Promise<void>

/**
 * System transform handler signature.
 */
export type HookSystemHandler = (
  input: Record<string, unknown>,
  output: { system: string[] },
) => Promise<void>
