/**
 * Ring Plugin - Integrated Entry Point
 *
 * Main plugin that initializes the hook system, configuration,
 * and background manager. Replaces direct plugin exports with
 * hook-based architecture.
 */

import type { Plugin, PluginInput } from "@opencode-ai/plugin"
import { hookRegistry } from "./hooks/index.js"
import type { HookName, HookContext, HookOutput } from "./hooks/types.js"
import {
  loadConfig,
  startConfigWatch,
  stopConfigWatch,
  isHookDisabledInConfig,
  type RingConfig,
} from "./config/index.js"
import { BackgroundManager } from "./background/index.js"
import type { BackgroundClient } from "./background/types.js"
import {
  createSessionStartHook,
  createContextInjectionHook,
  createNotificationHook,
  createTaskCompletionHook,
} from "./hooks/factories/index.js"
import { getSessionId } from "./utils/state.js"

/**
 * Initialize all hooks based on configuration.
 *
 * @param config - The loaded Ring configuration
 */
function initializeHooks(config: RingConfig): void {
  // Clear existing hooks
  hookRegistry.clear()

  // Set disabled hooks from config
  hookRegistry.setDisabledHooks(config.disabled_hooks)

  // Register enabled hooks
  if (!isHookDisabledInConfig("session-start")) {
    hookRegistry.register(createSessionStartHook(config.hooks?.["session-start"]))
  }

  if (!isHookDisabledInConfig("context-injection")) {
    hookRegistry.register(createContextInjectionHook(config.hooks?.["context-injection"]))
  }

  if (!isHookDisabledInConfig("notification")) {
    hookRegistry.register(createNotificationHook({
      enabled: config.notifications.enabled,
      notifyOn: {
        sessionIdle: config.notifications.onIdle,
        error: config.notifications.onError,
      },
    }))
  }

  if (!isHookDisabledInConfig("task-completion")) {
    hookRegistry.register(createTaskCompletionHook(config.hooks?.["task-completion"]))
  }
}

/**
 * Build a HookContext from available context data.
 *
 * @param sessionId - The session ID
 * @param directory - The project directory
 * @param lifecycle - The hook lifecycle event
 * @param event - Optional event data
 * @returns A complete HookContext
 */
function buildHookContext(
  sessionId: string,
  directory: string,
  lifecycle: HookContext["lifecycle"],
  event?: { type: string; properties?: Record<string, unknown> }
): HookContext {
  return {
    sessionId,
    directory,
    lifecycle,
    event,
  }
}

/**
 * Ring Plugin with Hook Architecture
 *
 * This plugin implements the middleware pattern using hooks for:
 * - Session lifecycle events (created, idle)
 * - System prompt transformation
 * - Compaction context injection
 * - Todo list monitoring
 * - Background task management
 */
export const RingPlugin: Plugin = async (ctx: PluginInput) => {
  const { directory, client } = ctx

  // Load configuration
  const config = loadConfig(directory)

  // Initialize hooks
  initializeHooks(config)

  // Initialize background manager
  // OpenCode client satisfies BackgroundClient interface for session and tui methods
  const backgroundManager = new BackgroundManager(
    client as unknown as BackgroundClient,
    directory,
    config.background_tasks
  )

  // Start config watch for hot-reload
  startConfigWatch(directory, (newConfig) => {
    initializeHooks(newConfig)
  })

  const sessionId = getSessionId()

  return {
    /**
     * Handle session lifecycle events.
     * Routes events to both background manager and hook system.
     */
    event: async ({ event }) => {
      // Route to background manager for task tracking
      backgroundManager.handleEvent(event)

      // Build output object for hooks to modify
      const output: HookOutput = {}

      // Extract session ID from event, fall back to plugin-init sessionId
      const eventSessionId = (event.properties?.sessionID as string) ?? sessionId

      // Execute hooks based on event type
      if (event.type === "session.created") {
        const hookCtx = buildHookContext(eventSessionId, directory, "session.created", event)
        await hookRegistry.executeLifecycle("session.created", hookCtx, output)
      }

      if (event.type === "session.idle") {
        const hookCtx = buildHookContext(eventSessionId, directory, "session.idle", event)
        await hookRegistry.executeLifecycle("session.idle", hookCtx, output)
      }

      if (event.type === "todo.updated") {
        const hookCtx = buildHookContext(eventSessionId, directory, "todo.updated", event)
        await hookRegistry.executeLifecycle("todo.updated", hookCtx, output)
      }
    },

    /**
     * Transform system prompt by injecting Ring context.
     * Hooks can add context to the system prompt array.
     */
    "experimental.chat.system.transform": async (
      _input: Record<string, unknown>,
      output: { system: string[] }
    ) => {
      if (!output?.system || !Array.isArray(output.system)) return

      const hookCtx = buildHookContext(sessionId, directory, "chat.params")
      const hookOutput: HookOutput = { system: output.system }
      await hookRegistry.executeLifecycle("chat.params", hookCtx, hookOutput)
    },

    /**
     * Inject context during session compaction.
     * Hooks can add important context to preserve during compaction.
     */
    "experimental.session.compacting": async (
      input: { sessionID: string },
      output: { context: string[] }
    ) => {
      if (!output?.context || !Array.isArray(output.context)) return

      const hookCtx = buildHookContext(input.sessionID, directory, "session.compacting")
      const hookOutput: HookOutput = { context: output.context }
      await hookRegistry.executeLifecycle("session.compacting", hookCtx, hookOutput)
    },

    /**
     * Expose background manager for tools and external access.
     *
     * @returns The BackgroundManager instance
     */
    getBackgroundManager: () => backgroundManager,

    /**
     * Clean up plugin resources.
     * Stops config watch and cleans up background manager.
     */
    dispose: () => {
      stopConfigWatch()
      backgroundManager.cleanup()
    },
  }
}

export default RingPlugin
