import { describe, expect, test } from "bun:test"
import * as fs from "node:fs"
import * as os from "node:os"
import * as path from "node:path"

import { DEFAULT_RING_CONFIG } from "../../../plugin/config/index.js"
import { createLifecycleRouter } from "../../../plugin/lifecycle/router.js"
import { readState, writeState } from "../../../plugin/utils/state.js"

describe("createLifecycleRouter - session.created", () => {
  test("resets state using event sessionID (not env/default session)", async () => {
    const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "ring-router-"))

    // Seed two session-specific state files
    writeState(tmpDir, "context-usage", { seeded: true }, "event-session")
    writeState(tmpDir, "context-usage", { seeded: true }, "other-session")

    const router = createLifecycleRouter({
      projectRoot: tmpDir,
      ringConfig: DEFAULT_RING_CONFIG,
    })

    await router({
      event: {
        type: "session.created",
        properties: {
          sessionID: "event-session",
        },
      },
    })

    expect(readState(tmpDir, "context-usage", "event-session")).toBeNull()
    expect(readState(tmpDir, "context-usage", "other-session")).not.toBeNull()
  })
})
