import { describe, expect, test } from "bun:test"
import { RingOpenCodeConfigSchema } from "../../../src/config/schema"

describe("RingOpenCodeConfigSchema - backwards compatible aliases", () => {
  test("accepts canonical keys (agent/permission)", () => {
    const result = RingOpenCodeConfigSchema.safeParse({
      agent: {
        "code-reviewer": {
          model: "anthropic/claude-sonnet-4-20250514",
        },
      },
      permission: {
        edit: "allow",
        bash: "ask",
      },
    })

    expect(result.success).toBe(true)
    if (!result.success) return

    expect(result.data.agent?.["code-reviewer"]?.model).toBe("anthropic/claude-sonnet-4-20250514")
    expect(result.data.permission?.edit).toBe("allow")
  })

  test("accepts alias keys (agents/permissions) and normalizes to canonical", () => {
    const result = RingOpenCodeConfigSchema.safeParse({
      agents: {
        "write-plan": {
          model: "anthropic/claude-opus-4-5-20251101",
        },
      },
      permissions: {
        edit: "deny",
      },
    })

    expect(result.success).toBe(true)
    if (!result.success) return

    // canonical keys present
    expect(result.data.agent?.["write-plan"]?.model).toBe("anthropic/claude-opus-4-5-20251101")
    expect(result.data.permission?.edit).toBe("deny")

    // aliases removed from parsed output
    expect("agents" in (result.data as Record<string, unknown>)).toBe(false)
    expect("permissions" in (result.data as Record<string, unknown>)).toBe(false)
  })

  test("fails when both canonical and alias keys are present (agent + agents)", () => {
    const result = RingOpenCodeConfigSchema.safeParse({
      agent: { "code-reviewer": { model: "x" } },
      agents: { "code-reviewer": { model: "y" } },
    })

    expect(result.success).toBe(false)
  })

  test("fails when both canonical and alias keys are present (permission + permissions)", () => {
    const result = RingOpenCodeConfigSchema.safeParse({
      permission: { edit: "allow" },
      permissions: { edit: "deny" },
    })

    expect(result.success).toBe(false)
  })
})
