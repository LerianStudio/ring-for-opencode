import { beforeEach, describe, expect, it } from "bun:test"

import type { WorkerInstance, WorkerProfile } from "../../../plugin/orchestrator/types.js"
import { WorkerPool } from "../../../plugin/orchestrator/worker-pool.js"

describe("WorkerPool", () => {
  let pool: WorkerPool

  const mockProfile: WorkerProfile = {
    id: "test-worker",
    name: "Test Worker",
    model: "test-model",
    purpose: "Testing",
    whenToUse: "During tests",
  }

  const createMockInstance = (profile: WorkerProfile): WorkerInstance => ({
    profile,
    status: "ready",
    port: 14100,
    startedAt: new Date(),
  })

  beforeEach(() => {
    pool = new WorkerPool()
  })

  describe("register/unregister", () => {
    it("registers a worker", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)

      expect(pool.get(mockProfile.id)).toBe(instance)
    })

    it("unregisters a worker", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)
      pool.unregister(mockProfile.id)

      expect(pool.get(mockProfile.id)).toBeUndefined()
    })

    it("returns false when unregistering unknown worker", () => {
      expect(pool.unregister("nonexistent")).toBe(false)
    })
  })

  describe("list", () => {
    it("returns all workers", () => {
      pool.register(createMockInstance(mockProfile))
      pool.resetRateLimit()
      pool.register(createMockInstance({ ...mockProfile, id: "worker-2" }))

      const list = pool.list()
      expect(list.length).toBe(2)
    })
  })

  describe("getWorkersByStatus", () => {
    it("filters by status", () => {
      const ready = createMockInstance(mockProfile)
      ready.status = "ready"
      pool.register(ready)
      pool.resetRateLimit()

      const busy = createMockInstance({ ...mockProfile, id: "busy-worker" })
      busy.status = "busy"
      pool.register(busy)

      const readyWorkers = pool.getWorkersByStatus("ready")
      expect(readyWorkers.length).toBe(1)
      expect(readyWorkers[0].profile.id).toBe(mockProfile.id)
    })
  })

  describe("getWorkersByCapability", () => {
    it("matches by purpose", () => {
      pool.register(
        createMockInstance({
          ...mockProfile,
          purpose: "Documentation research",
        }),
      )

      const matches = pool.getWorkersByCapability("documentation")
      expect(matches.length).toBe(1)
    })

    it("matches by tags", () => {
      pool.register(
        createMockInstance({
          ...mockProfile,
          tags: ["vision", "images"],
        }),
      )

      const matches = pool.getWorkersByCapability("vision")
      expect(matches.length).toBe(1)
    })
  })

  describe("updateStatus", () => {
    it("updates worker status", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)

      pool.updateStatus(mockProfile.id, "busy")

      expect(pool.get(mockProfile.id)?.status).toBe("busy")
    })

    it("sets error on error status", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)

      pool.updateStatus(mockProfile.id, "error", "Connection lost")

      const updated = pool.get(mockProfile.id)
      expect(updated?.status).toBe("error")
      expect(updated?.error).toBe("Connection lost")
    })
  })

  describe("session ownership", () => {
    it("tracks worker ownership by session", () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)
      pool.trackOwnership("session-1", mockProfile.id)

      const owned = pool.getWorkersForSession("session-1")
      expect(owned).toContain(mockProfile.id)
    })

    it("clears session ownership", () => {
      pool.trackOwnership("session-1", mockProfile.id)
      pool.clearSessionOwnership("session-1")

      const owned = pool.getWorkersForSession("session-1")
      expect(owned.length).toBe(0)
    })
  })

  describe("events", () => {
    it("emits spawn event on register", () => {
      const events: string[] = []
      pool.on("spawn", () => events.push("spawn"))

      pool.register(createMockInstance(mockProfile))

      expect(events).toContain("spawn")
    })

    it("emits stop event on unregister", () => {
      const events: string[] = []
      pool.register(createMockInstance(mockProfile))
      pool.on("stop", () => events.push("stop"))

      pool.unregister(mockProfile.id)

      expect(events).toContain("stop")
    })
  })

  describe("getOrSpawn", () => {
    it("returns existing worker", async () => {
      const instance = createMockInstance(mockProfile)
      pool.register(instance)

      const spawnFn = async () => createMockInstance(mockProfile)
      const result = await pool.getOrSpawn(
        mockProfile,
        { basePort: 14100, timeout: 5000, directory: "/tmp" },
        spawnFn,
      )

      expect(result).toBe(instance)
    })

    it("deduplicates concurrent spawns", async () => {
      let spawnCount = 0
      let resolveSpawn: ((instance: WorkerInstance) => void) | undefined

      const spawnFn = async () => {
        spawnCount += 1
        return await new Promise<WorkerInstance>((resolve) => {
          resolveSpawn = resolve
        })
      }

      const options = { basePort: 14100, timeout: 5000, directory: "/tmp" }

      const first = pool.getOrSpawn(mockProfile, options, spawnFn)
      const second = pool.getOrSpawn(mockProfile, options, spawnFn)

      if (!resolveSpawn) {
        throw new Error("Spawn callback not initialized")
      }

      resolveSpawn(createMockInstance(mockProfile))

      const [r1, r2] = await Promise.all([first, second])

      expect(spawnCount).toBe(1)
      expect(r1).toBe(r2)
    })
  })

  describe("register rate limiting", () => {
    it("enforces rate limit between spawns", () => {
      const rateLimitedPool = new WorkerPool()
      rateLimitedPool.register(createMockInstance(mockProfile))

      const secondProfile = { ...mockProfile, id: "worker-2" }
      expect(() => rateLimitedPool.register(createMockInstance(secondProfile))).toThrow(
        /Rate limited/,
      )
    })

    it("allows registration after rate limit window", async () => {
      const rateLimitedPool = new WorkerPool()
      rateLimitedPool.register(createMockInstance(mockProfile))

      rateLimitedPool.resetRateLimit()

      const secondProfile = { ...mockProfile, id: "worker-2" }
      expect(() => rateLimitedPool.register(createMockInstance(secondProfile))).not.toThrow()
    })
  })

  describe("state machine validation", () => {
    it("rejects invalid state transitions", () => {
      const statePool = new WorkerPool()
      const instance = createMockInstance(mockProfile)
      instance.status = "starting"
      statePool.register(instance)

      statePool.updateStatus(mockProfile.id, "busy")

      expect(statePool.get(mockProfile.id)?.status).toBe("starting")
    })

    it("allows valid state transitions", () => {
      const statePool = new WorkerPool()
      const instance = createMockInstance(mockProfile)
      instance.status = "starting"
      statePool.register(instance)

      statePool.updateStatus(mockProfile.id, "ready")
      expect(statePool.get(mockProfile.id)?.status).toBe("ready")

      statePool.updateStatus(mockProfile.id, "busy")
      expect(statePool.get(mockProfile.id)?.status).toBe("busy")
    })

    it("allows stopped from any active state via stop", async () => {
      const statePool = new WorkerPool()
      const instance = createMockInstance(mockProfile)
      instance.status = "starting"
      statePool.register(instance)

      const result = await statePool.stop(mockProfile.id)
      expect(result).toBe(true)
    })
  })

  describe("orphaned worker cleanup", () => {
    it("cleans up spawned worker when registration fails", async () => {
      const cleanupPool = new WorkerPool()

      cleanupPool.register(createMockInstance(mockProfile))

      const secondProfile = { ...mockProfile, id: "worker-2" }
      let shutdownCalled = false

      const spawnFn = async () => {
        const instance = createMockInstance(secondProfile)
        instance.shutdown = async () => {
          shutdownCalled = true
        }
        return instance
      }

      await expect(
        cleanupPool.getOrSpawn(
          secondProfile,
          { basePort: 14100, timeout: 5000, directory: "/tmp" },
          spawnFn,
        ),
      ).rejects.toThrow(/Rate limited/)

      expect(shutdownCalled).toBe(true)
    })
  })
})
