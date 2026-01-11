/**
 * Ring Unified Plugin
 *
 * Single entry point that matches oh-my-opencode's registration pattern.
 * Combines all Ring functionality into one Plugin export.
 *
 * Features:
 * - Config handler: Injects agents, skills, commands
 * - Tool registration: ring_doubt tool
 * - Event routing: Lifecycle events to hooks
 * - System transform: Context injection
 * - Compaction: Context preservation
 */

import type { Plugin, PluginInput } from "@opencode-ai/plugin"
import type { Config as OpenCodeSdkConfig } from "@opencode-ai/sdk"
// Background
import { BackgroundManager } from "./background/index.js"
import type { BackgroundClient } from "./background/types.js"
import type { RingConfig } from "./config/index.js"
// Config
import { createConfigHandler, loadConfig } from "./config/index.js"
import {
  createContextInjectionHook,
  createNotificationHook,
  createSessionStartHook,
  createTaskCompletionHook,
} from "./hooks/factories/index.js"
// Hooks
import { hookRegistry } from "./hooks/index.js"
import type { HookContext, HookOutput } from "./hooks/types.js"
// Lifecycle
import { createLifecycleRouter } from "./lifecycle/index.js"
// Tools
import { ringTools } from "./tools/index.js"

// Utils
import { getSessionId } from "./utils/state.js"

/**
 * Initialize all hooks based on configuration.
 */
function initializeHooks(config: RingConfig): void {
  hookRegistry.clear()
  hookRegistry.setDisabledHooks(config.disabled_hooks)

  const isDisabled = (name: string) => config.disabled_hooks?.includes(name as never) ?? false

  if (!isDisabled("session-start")) {
    hookRegistry.register(createSessionStartHook(config.hooks?.["session-start"]))
  }

  if (!isDisabled("context-injection")) {
    hookRegistry.register(createContextInjectionHook(config.hooks?.["context-injection"]))
  }

  if (!isDisabled("notification")) {
    hookRegistry.register(
      createNotificationHook({
        enabled: config.notifications.enabled,
        notifyOn: {
          sessionIdle: config.notifications.onIdle,
          error: config.notifications.onError,
        },
      }),
    )
  }

  if (!isDisabled("task-completion")) {
    hookRegistry.register(createTaskCompletionHook(config.hooks?.["task-completion"]))
  }
}

/**
 * Build a HookContext from available context data.
 */
function buildHookContext(
  sessionId: string,
  directory: string,
  lifecycle: HookContext["lifecycle"],
  event?: { type: string; properties?: Record<string, unknown> },
): HookContext {
  return {
    sessionId,
    directory,
    lifecycle,
    event,
  }
}

/**
 * Ring Unified Plugin
 *
 * Matches oh-my-opencode's Plugin signature:
 * Plugin = (input: PluginInput) => Promise<Hooks>
 *
 * Note: Return type is inferred to allow custom properties (getBackgroundManager, dispose)
 * that extend the base Hooks interface for Ring-specific functionality.
 */
export const RingUnifiedPlugin: Plugin = async (ctx: PluginInput) => {
  const { directory, client } = ctx
  const projectRoot = directory

  // Load Ring configuration
  const config = loadConfig(projectRoot)
  const debug = process.env.RING_DEBUG === "true"

  if (debug) {
    console.debug("[ring] Initializing unified plugin")
  }

  // Initialize hooks
  initializeHooks(config)

  // Initialize background manager
  const backgroundManager = new BackgroundManager(
    client as unknown as BackgroundClient,
    directory,
    config.background_tasks,
  )

  // Create config handler for OpenCode config injection
  const configHandler = createConfigHandler({
    projectRoot,
    ringConfig: config,
  })

  // Create lifecycle router (state side-effects only; notifications go through hooks)
  const lifecycleRouter = createLifecycleRouter({
    projectRoot,
    ringConfig: config,
  })

  const sessionId = getSessionId()

  return {
    // Register Ring tools
    tool: ringTools,

    // Config handler - inject agents, skills, commands
    // Type assertion needed: our handler modifies a subset of config properties
    // that are compatible with the SDK's Config type at runtime
    config: configHandler as unknown as (input: OpenCodeSdkConfig) => Promise<void>,

    // Event handler - lifecycle routing + hook execution
    event: async ({ event }) => {
      // Route to background manager for task tracking
      backgroundManager.handleEvent(event)

      // Route to lifecycle router
      await lifecycleRouter({ event })

      // Build output object for hooks to modify
      const output: HookOutput = {}
      // Extract sessionID from event properties (may be present in various event types)
      const props = event.properties as Record<string, unknown> | undefined
      const eventSessionId = (props?.sessionID as string) ?? sessionId

      // Build normalized event for hook context
      const normalizedEvent = {
        type: event.type,
        properties: props,
      }

      // Execute hooks based on event type
      if (event.type === "session.created") {
        const hookCtx = buildHookContext(
          eventSessionId,
          directory,
          "session.created",
          normalizedEvent,
        )
        await hookRegistry.executeLifecycle("session.created", hookCtx, output)
      }

      if (event.type === "session.idle") {
        const hookCtx = buildHookContext(eventSessionId, directory, "session.idle", normalizedEvent)
        await hookRegistry.executeLifecycle("session.idle", hookCtx, output)
      }

      if (event.type === "session.error") {
        const hookCtx = buildHookContext(
          eventSessionId,
          directory,
          "session.error",
          normalizedEvent,
        )
        await hookRegistry.executeLifecycle("session.error", hookCtx, output)
      }

      if (event.type === "todo.updated") {
        const hookCtx = buildHookContext(eventSessionId, directory, "todo.updated", normalizedEvent)
        await hookRegistry.executeLifecycle("todo.updated", hookCtx, output)
      }
    },

    // System prompt transformation
    "experimental.chat.system.transform": async (
      _input: Record<string, unknown>,
      output: { system: string[] },
    ) => {
      if (!output?.system || !Array.isArray(output.system)) return

      const hookCtx = buildHookContext(sessionId, directory, "chat.params")
      const hookOutput: HookOutput = { system: output.system }
      await hookRegistry.executeLifecycle("chat.params", hookCtx, hookOutput)
    },

    // Compaction context injection
    "experimental.session.compacting": async (
      input: { sessionID: string },
      output: { context: string[] },
    ) => {
      if (!output?.context || !Array.isArray(output.context)) return

      const hookCtx = buildHookContext(input.sessionID, directory, "session.compacting")
      const hookOutput: HookOutput = { context: output.context }
      await hookRegistry.executeLifecycle("session.compacting", hookCtx, hookOutput)
    },

    // Expose background manager
    getBackgroundManager: () => backgroundManager,

    // Cleanup
    dispose: () => {
      backgroundManager.cleanup()
    },
  }
}

export default RingUnifiedPlugin
