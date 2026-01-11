/**
 * Ring Concurrency Manager
 *
 * Manages concurrency limits for background task execution.
 * Prevents overloading the system with too many parallel tasks.
 */

import type { BackgroundTaskConfig } from "../config"

/**
 * Queued waiter for slot acquisition.
 */
interface QueuedWaiter {
  resolve: () => void
  reject: (error: Error) => void
}

/**
 * Manages concurrency limits for background tasks.
 *
 * Features:
 * - Global default concurrency limit
 * - Per-key (agent) concurrency limits
 * - Waiting queue for slot acquisition
 * - Config hot-reload support
 */
export class ConcurrencyManager {
  private activeCounts: Map<string, number> = new Map()
  private waitingQueues: Map<string, QueuedWaiter[]> = new Map()
  private config: BackgroundTaskConfig

  /**
   * Creates a new ConcurrencyManager.
   *
   * @param config - Background task configuration with concurrency settings
   */
  constructor(config?: BackgroundTaskConfig) {
    this.config = config ?? {
      defaultConcurrency: 3,
      taskTimeoutMs: 1800000,
    }
  }

  /**
   * Gets the concurrency limit for a given key.
   *
   * @param key - The concurrency key (usually agent name)
   * @returns The concurrency limit for this key
   */
  getLimit(key: string): number {
    const agentLimit = this.config.agentConcurrency?.[key]
    if (agentLimit !== undefined) {
      return agentLimit
    }
    return this.config.defaultConcurrency
  }

  /**
   * Gets the current active count for a key.
   *
   * @param key - The concurrency key
   * @returns Number of active slots for this key
   */
  getActiveCount(key: string): number {
    return this.activeCounts.get(key) ?? 0
  }

  /**
   * Checks if a slot can be acquired for the given key.
   *
   * @param key - The concurrency key
   * @returns True if a slot is available
   */
  canAcquire(key: string): boolean {
    const active = this.getActiveCount(key)
    const limit = this.getLimit(key)
    return active < limit
  }

  /**
   * Acquires a concurrency slot for the given key.
   * If no slot is available, waits until one becomes free or timeout expires.
   *
   * @param key - The concurrency key
   * @param timeoutMs - Timeout in milliseconds (default: 5 minutes)
   * @returns Promise that resolves when slot is acquired
   * @throws Error if timeout expires before slot is acquired
   */
  async acquire(key: string, timeoutMs = 300000): Promise<void> {
    if (this.canAcquire(key)) {
      const current = this.activeCounts.get(key) ?? 0
      this.activeCounts.set(key, current + 1)
      return
    }

    // H2: Wait for a slot with timeout to prevent unbounded waits
    return new Promise<void>((resolve, reject) => {
      let resolved = false

      const timeoutId = setTimeout(() => {
        if (resolved) return
        resolved = true

        // Remove from queue
        const queue = this.waitingQueues.get(key)
        if (queue) {
          const index = queue.findIndex((w) => w.resolve === resolveWrapper)
          if (index !== -1) {
            queue.splice(index, 1)
          }
          if (queue.length === 0) {
            this.waitingQueues.delete(key)
          }
        }
        reject(
          new Error(
            `Concurrency slot acquisition timed out after ${timeoutMs}ms for key: ${key}`
          )
        )
      }, timeoutMs)

      const resolveWrapper = () => {
        if (resolved) return
        resolved = true
        clearTimeout(timeoutId)
        resolve()
      }

      const rejectWrapper = (error: Error) => {
        if (resolved) return
        resolved = true
        clearTimeout(timeoutId)
        reject(error)
      }

      const waiter = { resolve: resolveWrapper, reject: rejectWrapper }
      const queue = this.waitingQueues.get(key) ?? []
      queue.push(waiter)
      this.waitingQueues.set(key, queue)
    })
  }

  /**
   * Releases a concurrency slot for the given key.
   * If waiters are queued, wakes up the next one.
   *
   * @param key - The concurrency key
   */
  release(key: string): void {
    const current = this.activeCounts.get(key) ?? 0
    if (current <= 0) {
      return
    }

    // Check if there are waiters
    const queue = this.waitingQueues.get(key)
    if (queue && queue.length > 0) {
      // Don't decrement - pass slot to next waiter
      const waiter = queue.shift()
      if (waiter) {
        waiter.resolve()
      }
      if (queue.length === 0) {
        this.waitingQueues.delete(key)
      }
    } else {
      // Decrement active count
      this.activeCounts.set(key, current - 1)
      if (current - 1 === 0) {
        this.activeCounts.delete(key)
      }
    }
  }

  /**
   * Gets total active slots across all keys.
   *
   * @returns Total number of active slots
   */
  getTotalActive(): number {
    let total = 0
    for (const count of this.activeCounts.values()) {
      total += count
    }
    return total
  }

  /**
   * Gets total waiting requests across all keys.
   *
   * @returns Total number of waiting requests
   */
  getTotalWaiting(): number {
    let total = 0
    for (const queue of this.waitingQueues.values()) {
      total += queue.length
    }
    return total
  }

  /**
   * Clears all state and rejects waiting requests.
   */
  clear(): void {
    // Reject all waiting requests
    for (const queue of this.waitingQueues.values()) {
      for (const waiter of queue) {
        waiter.reject(new Error("ConcurrencyManager cleared"))
      }
    }
    this.activeCounts.clear()
    this.waitingQueues.clear()
  }

  /**
   * Updates configuration with new values.
   *
   * @param config - New background task configuration
   */
  updateConfig(config: BackgroundTaskConfig): void {
    this.config = config
  }
}
