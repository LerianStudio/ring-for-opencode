import { describe, expect, test } from "bun:test"

import { ConcurrencyManager } from "../../../plugin/background/concurrency.js"

describe("background/ConcurrencyManager", () => {
  test("acquire rejects after timeout when no slots are available", async () => {
    const manager = new ConcurrencyManager({
      defaultConcurrency: 1,
      taskTimeoutMs: 60_000,
    })

    await manager.acquire("key")

    await expect(manager.acquire("key", 25)).rejects.toThrow(/timed out/i)

    manager.release("key")
  })
})
