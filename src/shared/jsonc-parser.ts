import { existsSync, readFileSync } from "node:fs"
import { parse, ParseError, printParseErrorCode } from "jsonc-parser"

export interface JsoncParseResult<T> {
  data: T | null
  errors: Array<{ message: string; offset: number; length: number }>
}

/**
 * Parse JSONC content (JSON with comments) into typed object.
 * Throws SyntaxError if parsing fails.
 */
export function parseJsonc<T = unknown>(content: string): T {
  const errors: ParseError[] = []
  const result = parse(content, errors, {
    allowTrailingComma: true,
    disallowComments: false,
  })

  if (errors.length > 0) {
    const errorMessages = errors
      .map((e) => `${printParseErrorCode(e.error)} at offset ${e.offset}`)
      .join(", ")
    throw new SyntaxError(`JSONC parse error: ${errorMessages}`)
  }

  // Handle empty or whitespace-only input
  if (result === undefined) {
    throw new SyntaxError("JSONC parse error: empty or invalid input")
  }

  return result as T
}

/**
 * Parse JSONC content safely, returning errors instead of throwing.
 */
export function parseJsoncSafe<T = unknown>(content: string): JsoncParseResult<T> {
  const errors: ParseError[] = []
  const data = parse(content, errors, {
    allowTrailingComma: true,
    disallowComments: false,
  }) as T | null

  return {
    data: errors.length > 0 ? null : data,
    errors: errors.map((e) => ({
      message: printParseErrorCode(e.error),
      offset: e.offset,
      length: e.length,
    })),
  }
}

/**
 * Read and parse a JSONC file. Returns null if file doesn't exist or parse fails.
 */
export function readJsoncFile<T = unknown>(filePath: string): T | null {
  try {
    const content = readFileSync(filePath, "utf-8")
    return parseJsonc<T>(content)
  } catch {
    return null
  }
}

/**
 * Detect whether a config file exists as .json or .jsonc
 */
export function detectConfigFile(basePath: string): {
  format: "json" | "jsonc" | "none"
  path: string
} {
  const jsoncPath = `${basePath}.jsonc`
  const jsonPath = `${basePath}.json`

  if (existsSync(jsoncPath)) {
    return { format: "jsonc", path: jsoncPath }
  }
  if (existsSync(jsonPath)) {
    return { format: "json", path: jsonPath }
  }
  return { format: "none", path: jsonPath }
}
