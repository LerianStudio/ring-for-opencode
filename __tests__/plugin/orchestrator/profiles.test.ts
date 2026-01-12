import { describe, expect, it } from "bun:test"

import {
  builtInProfiles,
  getProfile,
  listProfileIds,
  mergeProfile,
  resolveModelTag,
} from "../../../plugin/orchestrator/profiles.js"

describe("builtInProfiles", () => {
  it("includes all expected profiles", () => {
    expect(builtInProfiles).toHaveProperty("vision")
    expect(builtInProfiles).toHaveProperty("docs")
    expect(builtInProfiles).toHaveProperty("coder")
    expect(builtInProfiles).toHaveProperty("architect")
    expect(builtInProfiles).toHaveProperty("explorer")
  })

  it("vision profile supports vision", () => {
    expect(builtInProfiles.vision.supportsVision).toBe(true)
  })

  it("docs profile supports web", () => {
    expect(builtInProfiles.docs.supportsWeb).toBe(true)
  })
})

describe("getProfile", () => {
  it("returns built-in profile", () => {
    const profile = getProfile("vision")
    expect(profile?.id).toBe("vision")
  })

  it("returns undefined for unknown profile", () => {
    expect(getProfile("nonexistent")).toBeUndefined()
  })
})

describe("mergeProfile", () => {
  it("merges overrides with base profile", () => {
    const merged = mergeProfile("vision", { name: "My Vision Worker" })

    expect(merged.id).toBe("vision")
    expect(merged.name).toBe("My Vision Worker")
    expect(merged.model).toBe(builtInProfiles.vision.model)
  })

  it("throws for unknown base profile", () => {
    expect(() => mergeProfile("nonexistent", {})).toThrow(/Unknown base profile/)
  })
})

describe("resolveModelTag", () => {
  const defaults = {
    default: "anthropic/claude-3-5-sonnet",
    vision: "anthropic/claude-3-5-sonnet",
    fast: "anthropic/claude-3-haiku",
    docs: "anthropic/claude-3-5-sonnet",
  }

  it("resolves node:vision tag", () => {
    expect(resolveModelTag("node:vision", defaults)).toBe(defaults.vision)
  })

  it("resolves node:fast tag", () => {
    expect(resolveModelTag("node:fast", defaults)).toBe(defaults.fast)
  })

  it("returns non-tag strings as-is", () => {
    expect(resolveModelTag("openai/gpt-4", defaults)).toBe("openai/gpt-4")
  })
})

describe("listProfileIds", () => {
  it("includes custom profiles", () => {
    const ids = listProfileIds({ custom: { ...builtInProfiles.coder, id: "custom" } })
    expect(ids).toContain("custom")
    expect(ids).toContain("coder")
  })
})
