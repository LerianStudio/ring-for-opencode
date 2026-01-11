/**
 * Ring Hook Registry
 *
 * Manages hook registration, instantiation, and lifecycle execution.
 * Implements middleware pattern for extensible hook chains.
 */

import type {
  Hook,
  HookContext,
  HookFactory,
  HookLifecycle,
  HookName,
  HookOutput,
  HookRegistryEntry,
  HookResult,
} from "./types.js"

/**
 * Configuration interface for hook system.
 */
export interface HookConfig {
  disabledHooks?: HookName[]
}

/**
 * Registry for managing hooks and their lifecycle execution.
 */
export class HookRegistry {
  /** Registered hook factories */
  private factories: Map<HookName, HookRegistryEntry> = new Map()

  /** Instantiated hooks */
  private hooks: Map<HookName, Hook> = new Map()

  /** Disabled hooks set */
  private disabledHooks: Set<HookName> = new Set()

  /**
   * Register a hook factory with metadata.
   */
  registerFactory(entry: HookRegistryEntry): void {
    this.factories.set(entry.name, entry)
  }

  /**
   * Get all registered factories.
   */
  getFactories(): HookRegistryEntry[] {
    return Array.from(this.factories.values())
  }

  /**
   * Instantiate a hook from its factory.
   */
  instantiate<TConfig extends Record<string, unknown> = Record<string, unknown>>(
    name: HookName,
    config?: TConfig
  ): Hook | null {
    const entry = this.factories.get(name)
    if (!entry) {
      return null
    }

    const hook = entry.factory(config as Record<string, unknown> | undefined)

    // Respect disabled state
    if (this.disabledHooks.has(name)) {
      hook.enabled = false
    }

    this.hooks.set(name, hook)
    return hook
  }

  /**
   * Register an instantiated hook directly.
   */
  register(hook: Hook): void {
    // Respect disabled state
    if (this.disabledHooks.has(hook.name)) {
      hook.enabled = false
    }
    this.hooks.set(hook.name, hook)
  }

  /**
   * Unregister a hook by name.
   */
  unregister(name: HookName): void {
    this.hooks.delete(name)
  }

  /**
   * Get a hook by name.
   */
  get(name: HookName): Hook | undefined {
    return this.hooks.get(name)
  }

  /**
   * Check if a hook is registered.
   */
  has(name: HookName): boolean {
    return this.hooks.has(name)
  }

  /**
   * Disable a hook by name.
   */
  disable(name: HookName): void {
    this.disabledHooks.add(name)
    const hook = this.hooks.get(name)
    if (hook) {
      hook.enabled = false
    }
  }

  /**
   * Enable a hook by name.
   */
  enable(name: HookName): void {
    this.disabledHooks.delete(name)
    const hook = this.hooks.get(name)
    if (hook) {
      hook.enabled = true
    }
  }

  /**
   * Check if a hook is disabled.
   */
  isDisabled(name: HookName): boolean {
    return this.disabledHooks.has(name)
  }

  /**
   * Set the disabled hooks from configuration.
   */
  setDisabledHooks(names: HookName[]): void {
    this.disabledHooks = new Set(names)

    // Update existing hooks
    for (const [hookName, hook] of this.hooks) {
      hook.enabled = !this.disabledHooks.has(hookName)
    }
  }

  /**
   * Get all hooks that respond to a specific lifecycle event.
   * Returns hooks sorted by priority (lower = earlier).
   */
  getHooksForLifecycle(lifecycle: HookLifecycle): Hook[] {
    const matchingHooks: Hook[] = []

    for (const hook of this.hooks.values()) {
      if (hook.enabled && hook.lifecycles.includes(lifecycle)) {
        matchingHooks.push(hook)
      }
    }

    // Sort by priority (lower = earlier, default 100)
    // Use name as secondary key for deterministic ordering
    return matchingHooks.sort((a, b) => {
      const priorityDiff = (a.priority ?? 100) - (b.priority ?? 100)
      if (priorityDiff !== 0) return priorityDiff
      return a.name.localeCompare(b.name)
    })
  }

  /**
   * Execute all hooks for a lifecycle event in priority order.
   * Implements chain-of-responsibility pattern.
   */
  async executeLifecycle(
    lifecycle: HookLifecycle,
    ctx: HookContext,
    output: HookOutput
  ): Promise<HookResult[]> {
    const hooks = this.getHooksForLifecycle(lifecycle)
    const results: HookResult[] = []
    let chainData: Record<string, unknown> = {}

    for (const hook of hooks) {
      try {
        // Pass chain data from previous hooks
        const contextWithChainData: HookContext = {
          ...ctx,
          chainData,
        }

        const result = await hook.execute(contextWithChainData, output)
        results.push(result)

        // Accumulate chain data
        if (result.data) {
          chainData = { ...chainData, ...result.data }
        }

        // Stop chain if requested
        if (result.stopChain) {
          break
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error)
        results.push({
          success: false,
          error: `Hook ${hook.name} failed: ${errorMessage}`,
        })
      }
    }

    return results
  }

  /**
   * Get names of all registered hooks.
   */
  getRegisteredNames(): HookName[] {
    return Array.from(this.hooks.keys())
  }

  /**
   * Get count of registered hooks.
   */
  count(): number {
    return this.hooks.size
  }

  /**
   * Clear all registered hooks.
   */
  clear(): void {
    this.hooks.clear()
    this.disabledHooks.clear()
  }
}

/**
 * Singleton hook registry instance.
 */
export const hookRegistry = new HookRegistry()

/**
 * Check if a hook is disabled in the given configuration.
 */
export function isHookDisabled(config: HookConfig | undefined, hookName: HookName): boolean {
  if (!config?.disabledHooks) {
    return false
  }
  return config.disabledHooks.includes(hookName)
}
