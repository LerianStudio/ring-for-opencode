import { describe, expect, test } from "bun:test"

import { deepMerge } from "../config/loader.js"

describe("security hardening", () => {
  test("deepMerge blocks prototype pollution (__proto__/constructor/prototype)", () => {
    const protoBefore = ({} as Record<string, unknown>)["polluted"]

    // JSON.parse ensures __proto__ is treated as a plain enumerable key.
    const source = JSON.parse(
      '{"__proto__":{"polluted":true},"constructor":{"prototype":{"polluted":true}},"prototype":{"polluted":true},"safe":{"b":2}}',
    ) as Record<string, unknown>

    try {
      const merged = deepMerge({ safe: { a: 1 } }, source as unknown as Partial<{ safe: unknown }>)

      expect(({} as Record<string, unknown>)["polluted"]).toBeUndefined()
      expect(protoBefore).toBeUndefined()
      expect(merged).toMatchObject({ safe: { a: 1, b: 2 } })
    } finally {
      delete (Object.prototype as unknown as Record<string, unknown>)["polluted"]
    }
  })
})
