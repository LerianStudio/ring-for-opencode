# Zod Configuration Schema and CLI Tooling Implementation Plan

> **For Agents:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Add Zod-based configuration schema validation and CLI tooling (install, doctor, version commands) to Ring, following oh-my-opencode patterns.

**Architecture:** Create a new `src/` directory structure for TypeScript CLI and config modules. The CLI uses Commander.js for command parsing and @clack/prompts for interactive installation. The doctor command uses a modular check system with typed results. Zod schemas are exported to JSON Schema for IDE autocomplete support.

**Tech Stack:**
- TypeScript 5.x with Bun runtime
- Zod ^4.1.8 (already in project dependencies)
- Commander.js ^14.0.x for CLI
- @clack/prompts ^0.11.x for interactive TUI
- picocolors ^1.1.x for terminal colors
- jsonc-parser ^3.3.x for JSONC support

**Global Prerequisites:**
- Environment: macOS/Linux, Bun 1.0+, Node 18+
- Tools: Verify with commands below
- Access: No external API keys required
- State: Work from `main` branch, clean working tree

**Verification before starting:**
```bash
# Run ALL these commands and verify output:
bun --version        # Expected: 1.0+
node --version       # Expected: v18+
git status           # Expected: clean working tree
ls opencode.json     # Expected: file exists
ls plugin/           # Expected: directory exists with TypeScript files
```

## Historical Precedent

**Query:** "zod schema cli config validation typescript"
**Index Status:** Empty (new project)

No historical data available. This is normal for new projects.
Proceeding with standard planning approach.

---

## Task 1: Install Required Dependencies

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/package.json`

**Prerequisites:**
- Tools: bun 1.0+
- Files must exist: `package.json`

**Step 1: Update package.json with new dependencies**

Open `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/package.json` and replace its entire content with:

```json
{
  "name": "ring-opencode",
  "version": "1.0.0",
  "description": "Ring skills library for OpenCode - enforces proven software engineering practices",
  "type": "module",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "bin": {
    "ring": "./dist/cli/index.js"
  },
  "files": [
    "dist",
    "assets"
  ],
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js"
    },
    "./schema.json": "./assets/ring-opencode.schema.json"
  },
  "scripts": {
    "build": "bun build src/index.ts --outdir dist --target bun --format esm && tsc --emitDeclarationOnly && bun build src/cli/index.ts --outdir dist/cli --target bun --format esm && bun run build:schema",
    "build:schema": "bun run script/build-schema.ts",
    "clean": "rm -rf dist",
    "prepublishOnly": "bun run clean && bun run build",
    "typecheck": "tsc --noEmit",
    "test": "bun test",
    "test:plugin": "bun test plugin/"
  },
  "dependencies": {
    "@clack/prompts": "^0.11.0",
    "@opencode-ai/plugin": "1.1.3",
    "better-sqlite3": "12.6.0",
    "commander": "^14.0.2",
    "jsonc-parser": "^3.3.1",
    "picocolors": "^1.1.1",
    "zod": "^4.1.8"
  },
  "devDependencies": {
    "@types/better-sqlite3": "7.6.13",
    "@types/node": "22.19.5",
    "bun-types": "latest",
    "typescript": "5.9.3"
  }
}
```

**Step 2: Install dependencies**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun install`

**Expected output:**
```
bun install v1.x.x

+ @clack/prompts@0.11.x
+ commander@14.x.x
+ jsonc-parser@3.3.x
+ picocolors@1.1.x
+ zod@4.1.x
...
done
```

**Step 3: Verify installation**

Run: `ls /Users/fredamaral/repos/fredcamaral/ring-for-opencode/node_modules/commander/package.json`

**Expected output:** File path displayed (exists)

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add package.json bun.lock && git commit -m "$(cat <<'EOF'
chore(deps): add CLI and schema validation dependencies

Add commander, @clack/prompts, picocolors, jsonc-parser for CLI tooling
and Zod schema validation support.
EOF
)"
```

**If Task Fails:**

1. **Dependencies fail to install:**
   - Check: `bun --version` (must be 1.0+)
   - Fix: Update bun with `curl -fsSL https://bun.sh/install | bash`
   - Rollback: `git checkout -- package.json`

2. **Version conflicts:**
   - Run: `rm -rf node_modules bun.lock && bun install`
   - Rollback: `git checkout -- .`

---

## Task 2: Create TypeScript Configuration

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/tsconfig.json`

**Prerequisites:**
- Tools: TypeScript 5.x
- Dependencies from Task 1 installed

**Step 1: Create tsconfig.json**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/tsconfig.json` with:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "lib": ["ES2022"],
    "types": ["bun-types", "node"],
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "plugin", "script"]
}
```

**Step 2: Verify TypeScript configuration**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && npx tsc --noEmit --project tsconfig.json 2>&1 | head -5`

**Expected output:**
```
error TS18003: No inputs were found...
```
(This is expected - we haven't created src files yet)

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add tsconfig.json && git commit -m "$(cat <<'EOF'
build: add TypeScript configuration for src directory

Configure TypeScript for ES2022 target with bundler module resolution.
Includes bun-types and node types for CLI development.
EOF
)"
```

**If Task Fails:**

1. **TypeScript not found:**
   - Check: `npx tsc --version`
   - Fix: `bun add -d typescript`
   - Rollback: `git checkout -- tsconfig.json`

---

## Task 3: Create Shared Utilities - JSONC Parser

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/shared/jsonc-parser.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/shared/index.ts`

**Prerequisites:**
- jsonc-parser dependency installed (Task 1)

**Step 1: Create src/shared directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/shared`

**Expected output:** (no output, directory created)

**Step 2: Create JSONC parser utility**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/shared/jsonc-parser.ts` with:

```typescript
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
  }) as T

  if (errors.length > 0) {
    const errorMessages = errors
      .map((e) => `${printParseErrorCode(e.error)} at offset ${e.offset}`)
      .join(", ")
    throw new SyntaxError(`JSONC parse error: ${errorMessages}`)
  }

  return result
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
```

**Step 3: Create shared index**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/shared/index.ts` with:

```typescript
export * from "./jsonc-parser"
```

**Step 4: Verify syntax**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/shared/index.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/shared/index.ts  X.XX KB
```

**Step 5: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/shared/ && git commit -m "$(cat <<'EOF'
feat(shared): add JSONC parser utilities

Add parseJsonc, parseJsoncSafe, readJsoncFile, and detectConfigFile
utilities for handling JSON with comments configuration files.
EOF
)"
```

**If Task Fails:**

1. **Import error:**
   - Check: `ls node_modules/jsonc-parser/`
   - Fix: `bun install`
   - Rollback: `rm -rf src/shared/`

---

## Task 4: Create Zod Configuration Schema

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/config/schema.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/config/index.ts`

**Prerequisites:**
- Zod dependency installed (Task 1)
- Shared utilities created (Task 3)

**Step 1: Create src/config directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/config`

**Expected output:** (no output, directory created)

**Step 2: Create Zod schema**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/config/schema.ts` with:

```typescript
import { z } from "zod"

// Permission value types
const PermissionValue = z.enum(["ask", "allow", "deny"])

const BashPermission = z.union([
  PermissionValue,
  z.record(z.string(), PermissionValue),
])

// Agent permission schema
const AgentPermissionSchema = z.object({
  edit: PermissionValue.optional(),
  bash: BashPermission.optional(),
  webfetch: PermissionValue.optional(),
  doom_loop: PermissionValue.optional(),
  external_directory: PermissionValue.optional(),
})

// Skill permission schema
const SkillPermissionSchema = z.union([
  PermissionValue,
  z.record(z.string(), PermissionValue),
])

// Global permission schema
const PermissionSchema = z.object({
  skill: SkillPermissionSchema.optional(),
  edit: PermissionValue.optional(),
  bash: BashPermission.optional(),
  webfetch: PermissionValue.optional(),
  doom_loop: PermissionValue.optional(),
  external_directory: PermissionValue.optional(),
})

// Agent mode
const AgentMode = z.enum(["primary", "subagent", "all"])

// Agent configuration schema
const AgentConfigSchema = z.object({
  mode: AgentMode.optional(),
  model: z.string().optional(),
  temperature: z.number().min(0).max(2).optional(),
  top_p: z.number().min(0).max(1).optional(),
  prompt: z.string().optional(),
  prompt_append: z.string().optional(),
  tools: z.record(z.string(), z.boolean()).optional(),
  disable: z.boolean().optional(),
  description: z.string().optional(),
  color: z
    .string()
    .regex(/^#[0-9A-Fa-f]{6}$/)
    .optional(),
  permission: AgentPermissionSchema.optional(),
})

// Agents configuration (keyed by agent name)
const AgentsConfigSchema = z.record(z.string(), AgentConfigSchema)

// Hook names specific to Ring
export const RingHookNameSchema = z.enum([
  "session-start",
  "context-injection",
  "notification",
  "task-completion-check",
  "session-outcome",
  "outcome-inference",
  "doubt-resolver",
])

// Skill source schema
const SkillSourceSchema = z.union([
  z.string(),
  z.object({
    path: z.string(),
    recursive: z.boolean().optional(),
    glob: z.string().optional(),
  }),
])

// Skill definition schema
const SkillDefinitionSchema = z.object({
  description: z.string().optional(),
  template: z.string().optional(),
  from: z.string().optional(),
  model: z.string().optional(),
  agent: z.string().optional(),
  subtask: z.boolean().optional(),
  "argument-hint": z.string().optional(),
  license: z.string().optional(),
  compatibility: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  "allowed-tools": z.array(z.string()).optional(),
  disable: z.boolean().optional(),
})

const SkillEntrySchema = z.union([z.boolean(), SkillDefinitionSchema])

// Skills configuration
const SkillsConfigSchema = z.union([
  z.array(z.string()),
  z
    .record(z.string(), SkillEntrySchema)
    .and(
      z
        .object({
          sources: z.array(SkillSourceSchema).optional(),
          enable: z.array(z.string()).optional(),
          disable: z.array(z.string()).optional(),
        })
        .partial()
    ),
])

// State directory configuration
const StateConfigSchema = z.object({
  directory: z.string().optional(),
  session_tracking: z.boolean().optional(),
  context_warnings: z.boolean().optional(),
})

// Notification configuration
const NotificationConfigSchema = z.object({
  enabled: z.boolean().optional(),
  sound: z.boolean().optional(),
  on_completion: z.boolean().optional(),
  on_error: z.boolean().optional(),
})

// Main Ring OpenCode configuration schema
export const RingOpenCodeConfigSchema = z.object({
  $schema: z.string().optional(),
  version: z.string().optional(),
  name: z.string().optional(),
  description: z.string().optional(),
  permission: PermissionSchema.optional(),
  agent: AgentsConfigSchema.optional(),
  disabled_hooks: z.array(RingHookNameSchema).optional(),
  skills: SkillsConfigSchema.optional(),
  state: StateConfigSchema.optional(),
  notification: NotificationConfigSchema.optional(),
})

// Type exports
export type RingOpenCodeConfig = z.infer<typeof RingOpenCodeConfigSchema>
export type AgentConfig = z.infer<typeof AgentConfigSchema>
export type AgentPermission = z.infer<typeof AgentPermissionSchema>
export type Permission = z.infer<typeof PermissionSchema>
export type SkillsConfig = z.infer<typeof SkillsConfigSchema>
export type SkillDefinition = z.infer<typeof SkillDefinitionSchema>
export type StateConfig = z.infer<typeof StateConfigSchema>
export type NotificationConfig = z.infer<typeof NotificationConfigSchema>
export type RingHookName = z.infer<typeof RingHookNameSchema>
```

**Step 3: Create config index**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/config/index.ts` with:

```typescript
export {
  RingOpenCodeConfigSchema,
  RingHookNameSchema,
} from "./schema"

export type {
  RingOpenCodeConfig,
  AgentConfig,
  AgentPermission,
  Permission,
  SkillsConfig,
  SkillDefinition,
  StateConfig,
  NotificationConfig,
  RingHookName,
} from "./schema"
```

**Step 4: Verify schema compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/config/index.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/config/index.ts  X.XX KB
```

**Step 5: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/config/ && git commit -m "$(cat <<'EOF'
feat(config): add Zod schema for Ring configuration

Define RingOpenCodeConfigSchema with validation for permissions, agents,
hooks, skills, state, and notification settings.
EOF
)"
```

**If Task Fails:**

1. **Zod import error:**
   - Check: `ls node_modules/zod/`
   - Fix: `bun install zod@^4.1.8`
   - Rollback: `rm -rf src/config/`

---

## Task 5: Create JSON Schema Build Script

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/script/build-schema.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/` (directory)

**Prerequisites:**
- Zod schema created (Task 4)

**Step 1: Create script and assets directories**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/script /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets`

**Expected output:** (no output, directories created)

**Step 2: Create build-schema script**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/script/build-schema.ts` with:

```typescript
#!/usr/bin/env bun
import * as z from "zod"
import { RingOpenCodeConfigSchema } from "../src/config/schema"

const SCHEMA_OUTPUT_PATH = "assets/ring-opencode.schema.json"

async function main() {
  console.log("Generating JSON Schema...")

  const jsonSchema = z.toJSONSchema(RingOpenCodeConfigSchema, {
    io: "input",
    target: "draft-7",
  })

  const finalSchema = {
    $schema: "http://json-schema.org/draft-07/schema#",
    $id: "https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-opencode.schema.json",
    title: "Ring OpenCode Configuration",
    description: "Configuration schema for ring-opencode plugin",
    ...jsonSchema,
  }

  await Bun.write(SCHEMA_OUTPUT_PATH, JSON.stringify(finalSchema, null, 2))

  console.log(`Done: ${SCHEMA_OUTPUT_PATH}`)
}

main()
```

**Step 3: Run build-schema script**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run script/build-schema.ts`

**Expected output:**
```
Generating JSON Schema...
Done: assets/ring-opencode.schema.json
```

**Step 4: Verify schema was generated**

Run: `head -20 /Users/fredamaral/repos/fredcamaral/ring-for-opencode/assets/ring-opencode.schema.json`

**Expected output:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-opencode.schema.json",
  "title": "Ring OpenCode Configuration",
  ...
}
```

**Step 5: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add script/ assets/ && git commit -m "$(cat <<'EOF'
feat(schema): add JSON schema build script and generated schema

Add script/build-schema.ts to generate JSON Schema from Zod definitions.
Export assets/ring-opencode.schema.json for IDE autocomplete support.
EOF
)"
```

**If Task Fails:**

1. **Zod toJSONSchema not found:**
   - Check: Zod version must be 4.x+ (v3 doesn't have toJSONSchema)
   - Fix: `bun install zod@^4.1.8`
   - Rollback: `rm -rf script/ assets/`

---

## Task 6: Create CLI Types and Constants

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/types.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/constants.ts`

**Prerequisites:**
- src directory structure exists (Task 3+)

**Step 1: Create src/cli directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli`

**Expected output:** (no output, directory created)

**Step 2: Create CLI types**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/types.ts` with:

```typescript
export interface InstallArgs {
  tui: boolean
  skipValidation?: boolean
}

export interface InstallConfig {
  configPath: string
  isNewInstall: boolean
}

export interface ConfigMergeResult {
  success: boolean
  configPath: string
  error?: string
}

export interface DetectedConfig {
  isInstalled: boolean
  configPath: string | null
  hasSchema: boolean
  version: string | null
}
```

**Step 3: Create CLI constants**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/constants.ts` with:

```typescript
import color from "picocolors"

export const PACKAGE_NAME = "ring-opencode"
export const CONFIG_FILE_NAME = "opencode.json"
export const SCHEMA_URL = "https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-opencode.schema.json"

export const SYMBOLS = {
  check: color.green("\u2713"),
  cross: color.red("\u2717"),
  warn: color.yellow("\u26A0"),
  info: color.blue("\u2139"),
  arrow: color.cyan("\u2192"),
  bullet: color.dim("\u2022"),
  skip: color.dim("\u25CB"),
} as const

export const STATUS_COLORS = {
  pass: color.green,
  fail: color.red,
  warn: color.yellow,
  skip: color.dim,
} as const

export const EXIT_CODES = {
  SUCCESS: 0,
  FAILURE: 1,
} as const
```

**Step 4: Verify files compile**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/cli/types.ts src/cli/constants.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/cli/types.ts      X.XX KB
  src/cli/constants.ts  X.XX KB
```

**Step 5: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/cli/types.ts src/cli/constants.ts && git commit -m "$(cat <<'EOF'
feat(cli): add CLI types and constants

Define InstallArgs, InstallConfig, ConfigMergeResult, DetectedConfig types.
Add SYMBOLS, STATUS_COLORS, EXIT_CODES constants for CLI output.
EOF
)"
```

**If Task Fails:**

1. **picocolors import error:**
   - Fix: `bun install picocolors`
   - Rollback: `rm -rf src/cli/`

---

## Task 7: Create Config Manager Utility

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/config-manager.ts`

**Prerequisites:**
- CLI types created (Task 6)
- Shared utilities created (Task 3)
- Config schema created (Task 4)

**Step 1: Create config-manager.ts**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/config-manager.ts` with:

```typescript
import { existsSync, mkdirSync, readFileSync, writeFileSync, statSync } from "node:fs"
import { join } from "node:path"
import { parseJsonc, detectConfigFile } from "../shared"
import { RingOpenCodeConfigSchema } from "../config"
import { SCHEMA_URL, CONFIG_FILE_NAME } from "./constants"
import type { ConfigMergeResult, DetectedConfig } from "./types"

interface NodeError extends Error {
  code?: string
}

function isPermissionError(err: unknown): boolean {
  const nodeErr = err as NodeError
  return nodeErr?.code === "EACCES" || nodeErr?.code === "EPERM"
}

function isFileNotFoundError(err: unknown): boolean {
  const nodeErr = err as NodeError
  return nodeErr?.code === "ENOENT"
}

function formatErrorWithSuggestion(err: unknown, context: string): string {
  if (isPermissionError(err)) {
    return `Permission denied: Cannot ${context}. Try running with elevated permissions.`
  }

  if (isFileNotFoundError(err)) {
    return `File not found while trying to ${context}.`
  }

  if (err instanceof SyntaxError) {
    return `JSON syntax error while trying to ${context}: ${err.message}`
  }

  const message = err instanceof Error ? err.message : String(err)
  return `Failed to ${context}: ${message}`
}

/**
 * Get the project root config path (opencode.json in cwd)
 */
export function getConfigPath(): string {
  return join(process.cwd(), CONFIG_FILE_NAME)
}

/**
 * Detect current Ring configuration
 */
export function detectCurrentConfig(): DetectedConfig {
  const configPath = getConfigPath()
  const result: DetectedConfig = {
    isInstalled: false,
    configPath: null,
    hasSchema: false,
    version: null,
  }

  if (!existsSync(configPath)) {
    return result
  }

  result.configPath = configPath

  try {
    const content = readFileSync(configPath, "utf-8")
    const config = parseJsonc<Record<string, unknown>>(content)

    if (config) {
      result.isInstalled = true
      result.hasSchema = typeof config.$schema === "string"
      result.version = typeof config.version === "string" ? config.version : null
    }
  } catch {
    // Config exists but is invalid
    result.isInstalled = false
  }

  return result
}

/**
 * Validate configuration against Zod schema
 */
export function validateConfig(configPath: string): { valid: boolean; errors: string[] } {
  try {
    const content = readFileSync(configPath, "utf-8")
    const rawConfig = parseJsonc<Record<string, unknown>>(content)
    const result = RingOpenCodeConfigSchema.safeParse(rawConfig)

    if (!result.success) {
      const errors = result.error.issues.map(
        (i) => `${i.path.join(".")}: ${i.message}`
      )
      return { valid: false, errors }
    }

    return { valid: true, errors: [] }
  } catch (err) {
    return {
      valid: false,
      errors: [err instanceof Error ? err.message : "Failed to parse config"],
    }
  }
}

/**
 * Add $schema to existing config for IDE autocomplete
 */
export function addSchemaToConfig(): ConfigMergeResult {
  const configPath = getConfigPath()

  try {
    if (!existsSync(configPath)) {
      // Create minimal config with schema
      const config = {
        $schema: SCHEMA_URL,
        version: "1.0.0",
        name: "ring-opencode",
        description: "Ring configuration",
      }
      writeFileSync(configPath, JSON.stringify(config, null, 2) + "\n")
      return { success: true, configPath }
    }

    const content = readFileSync(configPath, "utf-8")
    const config = parseJsonc<Record<string, unknown>>(content)

    if (!config) {
      return { success: false, configPath, error: "Failed to parse existing config" }
    }

    // Already has schema
    if (config.$schema === SCHEMA_URL) {
      return { success: true, configPath }
    }

    // Add/update $schema
    const updatedConfig = { $schema: SCHEMA_URL, ...config }
    delete (updatedConfig as Record<string, unknown>)["$schema"]
    const finalConfig = { $schema: SCHEMA_URL, ...updatedConfig }

    writeFileSync(configPath, JSON.stringify(finalConfig, null, 2) + "\n")
    return { success: true, configPath }
  } catch (err) {
    return { success: false, configPath, error: formatErrorWithSuggestion(err, "update config") }
  }
}

/**
 * Check if opencode CLI is installed
 */
export async function isOpenCodeInstalled(): Promise<boolean> {
  try {
    const proc = Bun.spawn(["opencode", "--version"], {
      stdout: "pipe",
      stderr: "pipe",
    })
    await proc.exited
    return proc.exitCode === 0
  } catch {
    return false
  }
}

/**
 * Get opencode version
 */
export async function getOpenCodeVersion(): Promise<string | null> {
  try {
    const proc = Bun.spawn(["opencode", "--version"], {
      stdout: "pipe",
      stderr: "pipe",
    })
    const output = await new Response(proc.stdout).text()
    await proc.exited
    if (proc.exitCode === 0) {
      return output.trim()
    }
    return null
  } catch {
    return null
  }
}
```

**Step 2: Verify config-manager compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/cli/config-manager.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/cli/config-manager.ts  X.XX KB
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/cli/config-manager.ts && git commit -m "$(cat <<'EOF'
feat(cli): add config manager utility

Add detectCurrentConfig, validateConfig, addSchemaToConfig,
isOpenCodeInstalled, getOpenCodeVersion utilities for CLI operations.
EOF
)"
```

**If Task Fails:**

1. **Import errors:**
   - Check: All previous tasks completed
   - Fix: Verify src/shared/index.ts and src/config/index.ts exist
   - Rollback: `git checkout -- src/cli/config-manager.ts`

---

## Task 8: Create Install Command

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/install.ts`

**Prerequisites:**
- Config manager created (Task 7)
- CLI types/constants created (Task 6)

**Step 1: Create install.ts**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/install.ts` with:

```typescript
import * as p from "@clack/prompts"
import color from "picocolors"
import type { InstallArgs, InstallConfig } from "./types"
import {
  addSchemaToConfig,
  isOpenCodeInstalled,
  getOpenCodeVersion,
  detectCurrentConfig,
  validateConfig,
} from "./config-manager"
import { SYMBOLS, SCHEMA_URL } from "./constants"

function printHeader(isUpdate: boolean): void {
  const mode = isUpdate ? "Update" : "Install"
  console.log()
  console.log(color.bgCyan(color.white(` Ring ${mode} `)))
  console.log()
}

function printStep(step: number, total: number, message: string): void {
  const progress = color.dim(`[${step}/${total}]`)
  console.log(`${progress} ${message}`)
}

function printSuccess(message: string): void {
  console.log(`${SYMBOLS.check} ${message}`)
}

function printError(message: string): void {
  console.log(`${SYMBOLS.cross} ${color.red(message)}`)
}

function printInfo(message: string): void {
  console.log(`${SYMBOLS.info} ${message}`)
}

function printWarning(message: string): void {
  console.log(`${SYMBOLS.warn} ${color.yellow(message)}`)
}

async function runTuiInstall(): Promise<number> {
  const detected = detectCurrentConfig()
  const isUpdate = detected.isInstalled

  p.intro(color.bgCyan(color.white(isUpdate ? " Ring Update " : " Ring Install ")))

  if (isUpdate && detected.configPath) {
    p.log.info(`Existing configuration found: ${detected.configPath}`)
  }

  const s = p.spinner()
  s.start("Checking OpenCode installation")

  const installed = await isOpenCodeInstalled()
  if (!installed) {
    s.stop("OpenCode is not installed")
    p.log.error("OpenCode is not installed on this system.")
    p.note("Visit https://opencode.ai/docs for installation instructions", "Installation Guide")
    p.outro(color.red("Please install OpenCode first."))
    return 1
  }

  const version = await getOpenCodeVersion()
  s.stop(`OpenCode ${version ?? "installed"} ${color.green("\u2713")}`)

  // Confirm installation
  const shouldContinue = await p.confirm({
    message: isUpdate
      ? "Update Ring configuration with schema validation?"
      : "Install Ring configuration with schema validation?",
    initialValue: true,
  })

  if (p.isCancel(shouldContinue) || !shouldContinue) {
    p.cancel("Installation cancelled.")
    return 1
  }

  s.start("Adding schema to configuration")
  const schemaResult = addSchemaToConfig()
  if (!schemaResult.success) {
    s.stop(`Failed: ${schemaResult.error}`)
    p.outro(color.red("Installation failed."))
    return 1
  }
  s.stop(`Schema added to ${color.cyan(schemaResult.configPath)}`)

  // Validate the updated config
  s.start("Validating configuration")
  const validation = validateConfig(schemaResult.configPath)
  if (!validation.valid) {
    s.stop("Configuration has validation errors")
    p.log.warn("Validation errors found:")
    for (const err of validation.errors) {
      p.log.message(`  ${SYMBOLS.bullet} ${err}`)
    }
  } else {
    s.stop("Configuration is valid")
  }

  p.note(
    `Your opencode.json now has schema validation.\n` +
      `IDE autocomplete is available via the $schema field.\n\n` +
      `Schema URL: ${color.cyan(SCHEMA_URL)}`,
    isUpdate ? "Configuration Updated" : "Installation Complete"
  )

  p.log.success(color.bold(isUpdate ? "Ring configuration updated!" : "Ring installed!"))
  p.log.message(`Run ${color.cyan("ring doctor")} to check your setup.`)

  p.outro(color.green("Happy coding with Ring!"))

  return 0
}

async function runNonTuiInstall(args: InstallArgs): Promise<number> {
  const detected = detectCurrentConfig()
  const isUpdate = detected.isInstalled

  printHeader(isUpdate)

  const totalSteps = 3
  let step = 1

  printStep(step++, totalSteps, "Checking OpenCode installation...")
  const installed = await isOpenCodeInstalled()
  if (!installed) {
    printError("OpenCode is not installed on this system.")
    printInfo("Visit https://opencode.ai/docs for installation instructions")
    return 1
  }

  const version = await getOpenCodeVersion()
  printSuccess(`OpenCode ${version ?? ""} detected`)

  printStep(step++, totalSteps, "Adding schema to configuration...")
  const schemaResult = addSchemaToConfig()
  if (!schemaResult.success) {
    printError(`Failed: ${schemaResult.error}`)
    return 1
  }
  printSuccess(`Schema added ${SYMBOLS.arrow} ${color.dim(schemaResult.configPath)}`)

  if (!args.skipValidation) {
    printStep(step++, totalSteps, "Validating configuration...")
    const validation = validateConfig(schemaResult.configPath)
    if (!validation.valid) {
      printWarning("Configuration has validation errors:")
      for (const err of validation.errors) {
        console.log(`  ${SYMBOLS.bullet} ${err}`)
      }
    } else {
      printSuccess("Configuration is valid")
    }
  }

  console.log()
  printSuccess(color.bold(isUpdate ? "Ring configuration updated!" : "Ring installed!"))
  console.log(`  Run ${color.cyan("ring doctor")} to check your setup.`)
  console.log()

  return 0
}

export async function install(args: InstallArgs): Promise<number> {
  if (!args.tui) {
    return runNonTuiInstall(args)
  }

  return runTuiInstall()
}
```

**Step 2: Verify install.ts compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/cli/install.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/cli/install.ts  X.XX KB
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/cli/install.ts && git commit -m "$(cat <<'EOF'
feat(cli): add interactive install command

Implement ring install with TUI mode using @clack/prompts.
Adds schema validation to opencode.json configuration.
EOF
)"
```

**If Task Fails:**

1. **@clack/prompts import error:**
   - Fix: `bun install @clack/prompts`
   - Rollback: `rm src/cli/install.ts`

---

## Task 9: Create Doctor Types and Constants

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/types.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/constants.ts`

**Prerequisites:**
- CLI constants created (Task 6)

**Step 1: Create src/cli/doctor directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor`

**Expected output:** (no output, directory created)

**Step 2: Create doctor types**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/types.ts` with:

```typescript
export type CheckStatus = "pass" | "fail" | "warn" | "skip"

export interface CheckResult {
  name: string
  status: CheckStatus
  message: string
  details?: string[]
  duration?: number
}

export type CheckFunction = () => Promise<CheckResult>

export type CheckCategory =
  | "installation"
  | "configuration"
  | "plugins"
  | "dependencies"

export interface CheckDefinition {
  id: string
  name: string
  category: CheckCategory
  check: CheckFunction
  critical?: boolean
}

export interface DoctorOptions {
  verbose?: boolean
  json?: boolean
  category?: CheckCategory
}

export interface DoctorSummary {
  total: number
  passed: number
  failed: number
  warnings: number
  skipped: number
  duration: number
}

export interface DoctorResult {
  results: CheckResult[]
  summary: DoctorSummary
  exitCode: number
}

export interface OpenCodeInfo {
  installed: boolean
  version: string | null
  path: string | null
}

export interface ConfigInfo {
  exists: boolean
  path: string | null
  format: "json" | "jsonc" | null
  valid: boolean
  errors: string[]
  hasSchema: boolean
}

export interface PluginInfo {
  name: string
  loaded: boolean
  path: string | null
  error?: string
}
```

**Step 3: Create doctor constants**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/constants.ts` with:

```typescript
import color from "picocolors"

export const SYMBOLS = {
  check: color.green("\u2713"),
  cross: color.red("\u2717"),
  warn: color.yellow("\u26A0"),
  info: color.blue("\u2139"),
  arrow: color.cyan("\u2192"),
  bullet: color.dim("\u2022"),
  skip: color.dim("\u25CB"),
} as const

export const STATUS_COLORS = {
  pass: color.green,
  fail: color.red,
  warn: color.yellow,
  skip: color.dim,
} as const

export const CHECK_IDS = {
  OPENCODE_INSTALLATION: "opencode-installation",
  CONFIG_EXISTS: "config-exists",
  CONFIG_VALIDATION: "config-validation",
  SCHEMA_PRESENT: "schema-present",
  PLUGIN_DIRECTORY: "plugin-directory",
  SKILL_DIRECTORY: "skill-directory",
  STATE_DIRECTORY: "state-directory",
  BUN_INSTALLED: "bun-installed",
  GIT_INSTALLED: "git-installed",
} as const

export const CHECK_NAMES: Record<string, string> = {
  [CHECK_IDS.OPENCODE_INSTALLATION]: "OpenCode Installation",
  [CHECK_IDS.CONFIG_EXISTS]: "Configuration File",
  [CHECK_IDS.CONFIG_VALIDATION]: "Configuration Validity",
  [CHECK_IDS.SCHEMA_PRESENT]: "Schema Reference",
  [CHECK_IDS.PLUGIN_DIRECTORY]: "Plugin Directory",
  [CHECK_IDS.SKILL_DIRECTORY]: "Skill Directory",
  [CHECK_IDS.STATE_DIRECTORY]: "State Directory",
  [CHECK_IDS.BUN_INSTALLED]: "Bun Runtime",
  [CHECK_IDS.GIT_INSTALLED]: "Git",
} as const

export const CATEGORY_NAMES: Record<string, string> = {
  installation: "Installation",
  configuration: "Configuration",
  plugins: "Plugins & Skills",
  dependencies: "Dependencies",
} as const

export const EXIT_CODES = {
  SUCCESS: 0,
  FAILURE: 1,
} as const
```

**Step 4: Verify files compile**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/cli/doctor/types.ts src/cli/doctor/constants.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/cli/doctor/types.ts      X.XX KB
  src/cli/doctor/constants.ts  X.XX KB
```

**Step 5: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/cli/doctor/ && git commit -m "$(cat <<'EOF'
feat(cli/doctor): add doctor types and constants

Define CheckResult, CheckDefinition, DoctorOptions, DoctorSummary types.
Add CHECK_IDS, CHECK_NAMES, CATEGORY_NAMES for health check system.
EOF
)"
```

**If Task Fails:**

1. **Directory creation fails:**
   - Check: Write permissions
   - Fix: `chmod 755 src/cli/`
   - Rollback: `rm -rf src/cli/doctor/`

---

## Task 10: Create Doctor Checks

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks/config.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks/installation.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks/dependencies.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks/index.ts`

**Prerequisites:**
- Doctor types/constants created (Task 9)
- Config manager created (Task 7)

**Step 1: Create checks directory**

Run: `mkdir -p /Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks`

**Expected output:** (no output, directory created)

**Step 2: Create installation checks**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks/installation.ts` with:

```typescript
import { existsSync } from "node:fs"
import { join } from "node:path"
import type { CheckResult, CheckDefinition } from "../types"
import { CHECK_IDS, CHECK_NAMES } from "../constants"
import { isOpenCodeInstalled, getOpenCodeVersion } from "../../config-manager"

export async function checkOpenCodeInstallation(): Promise<CheckResult> {
  const installed = await isOpenCodeInstalled()

  if (!installed) {
    return {
      name: CHECK_NAMES[CHECK_IDS.OPENCODE_INSTALLATION],
      status: "fail",
      message: "OpenCode is not installed",
      details: [
        "Install OpenCode: https://opencode.ai/docs",
        "Run: curl -fsSL https://opencode.ai/install | bash",
      ],
    }
  }

  const version = await getOpenCodeVersion()
  return {
    name: CHECK_NAMES[CHECK_IDS.OPENCODE_INSTALLATION],
    status: "pass",
    message: version ? `Version ${version}` : "Installed",
  }
}

export async function checkPluginDirectory(): Promise<CheckResult> {
  const pluginDir = join(process.cwd(), "plugin")
  const exists = existsSync(pluginDir)

  if (!exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.PLUGIN_DIRECTORY],
      status: "warn",
      message: "Plugin directory not found",
      details: ["Expected: ./plugin/"],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.PLUGIN_DIRECTORY],
    status: "pass",
    message: "Found",
    details: [`Path: ${pluginDir}`],
  }
}

export async function checkSkillDirectory(): Promise<CheckResult> {
  const skillDir = join(process.cwd(), "skill")
  const exists = existsSync(skillDir)

  if (!exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.SKILL_DIRECTORY],
      status: "warn",
      message: "Skill directory not found",
      details: ["Expected: ./skill/"],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.SKILL_DIRECTORY],
    status: "pass",
    message: "Found",
    details: [`Path: ${skillDir}`],
  }
}

export async function checkStateDirectory(): Promise<CheckResult> {
  const stateDir = join(process.cwd(), ".ring", "state")
  const exists = existsSync(stateDir)

  if (!exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.STATE_DIRECTORY],
      status: "skip",
      message: "State directory will be created on first run",
      details: ["Expected: ./.ring/state/"],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.STATE_DIRECTORY],
    status: "pass",
    message: "Found",
    details: [`Path: ${stateDir}`],
  }
}

export function getInstallationCheckDefinitions(): CheckDefinition[] {
  return [
    {
      id: CHECK_IDS.OPENCODE_INSTALLATION,
      name: CHECK_NAMES[CHECK_IDS.OPENCODE_INSTALLATION],
      category: "installation",
      check: checkOpenCodeInstallation,
      critical: true,
    },
  ]
}

export function getPluginCheckDefinitions(): CheckDefinition[] {
  return [
    {
      id: CHECK_IDS.PLUGIN_DIRECTORY,
      name: CHECK_NAMES[CHECK_IDS.PLUGIN_DIRECTORY],
      category: "plugins",
      check: checkPluginDirectory,
    },
    {
      id: CHECK_IDS.SKILL_DIRECTORY,
      name: CHECK_NAMES[CHECK_IDS.SKILL_DIRECTORY],
      category: "plugins",
      check: checkSkillDirectory,
    },
    {
      id: CHECK_IDS.STATE_DIRECTORY,
      name: CHECK_NAMES[CHECK_IDS.STATE_DIRECTORY],
      category: "plugins",
      check: checkStateDirectory,
    },
  ]
}
```

**Step 3: Create config checks**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks/config.ts` with:

```typescript
import { existsSync } from "node:fs"
import type { CheckResult, CheckDefinition } from "../types"
import { CHECK_IDS, CHECK_NAMES } from "../constants"
import { getConfigPath, validateConfig, detectCurrentConfig } from "../../config-manager"
import { SCHEMA_URL } from "../../constants"

export async function checkConfigExists(): Promise<CheckResult> {
  const configPath = getConfigPath()
  const exists = existsSync(configPath)

  if (!exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.CONFIG_EXISTS],
      status: "fail",
      message: "opencode.json not found",
      details: [
        "Create opencode.json in project root",
        "Run: ring install",
      ],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.CONFIG_EXISTS],
    status: "pass",
    message: "Found",
    details: [`Path: ${configPath}`],
  }
}

export async function checkConfigValidity(): Promise<CheckResult> {
  const configPath = getConfigPath()

  if (!existsSync(configPath)) {
    return {
      name: CHECK_NAMES[CHECK_IDS.CONFIG_VALIDATION],
      status: "skip",
      message: "No config file to validate",
    }
  }

  const validation = validateConfig(configPath)

  if (!validation.valid) {
    return {
      name: CHECK_NAMES[CHECK_IDS.CONFIG_VALIDATION],
      status: "fail",
      message: "Configuration has validation errors",
      details: [
        `Path: ${configPath}`,
        ...validation.errors.map((e) => `Error: ${e}`),
      ],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.CONFIG_VALIDATION],
    status: "pass",
    message: "Valid configuration",
    details: [`Path: ${configPath}`],
  }
}

export async function checkSchemaPresent(): Promise<CheckResult> {
  const detected = detectCurrentConfig()

  if (!detected.isInstalled) {
    return {
      name: CHECK_NAMES[CHECK_IDS.SCHEMA_PRESENT],
      status: "skip",
      message: "No config file",
    }
  }

  if (!detected.hasSchema) {
    return {
      name: CHECK_NAMES[CHECK_IDS.SCHEMA_PRESENT],
      status: "warn",
      message: "No $schema reference for IDE autocomplete",
      details: [
        `Add to opencode.json: "$schema": "${SCHEMA_URL}"`,
        "Or run: ring install",
      ],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.SCHEMA_PRESENT],
    status: "pass",
    message: "Schema reference present",
  }
}

export function getConfigCheckDefinitions(): CheckDefinition[] {
  return [
    {
      id: CHECK_IDS.CONFIG_EXISTS,
      name: CHECK_NAMES[CHECK_IDS.CONFIG_EXISTS],
      category: "configuration",
      check: checkConfigExists,
      critical: true,
    },
    {
      id: CHECK_IDS.CONFIG_VALIDATION,
      name: CHECK_NAMES[CHECK_IDS.CONFIG_VALIDATION],
      category: "configuration",
      check: checkConfigValidity,
      critical: false,
    },
    {
      id: CHECK_IDS.SCHEMA_PRESENT,
      name: CHECK_NAMES[CHECK_IDS.SCHEMA_PRESENT],
      category: "configuration",
      check: checkSchemaPresent,
      critical: false,
    },
  ]
}
```

**Step 4: Create dependency checks**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks/dependencies.ts` with:

```typescript
import type { CheckResult, CheckDefinition } from "../types"
import { CHECK_IDS, CHECK_NAMES } from "../constants"

async function checkCommandExists(command: string, args: string[]): Promise<{ exists: boolean; version: string | null }> {
  try {
    const proc = Bun.spawn([command, ...args], {
      stdout: "pipe",
      stderr: "pipe",
    })
    const output = await new Response(proc.stdout).text()
    await proc.exited

    if (proc.exitCode === 0) {
      return { exists: true, version: output.trim().split("\n")[0] }
    }
    return { exists: false, version: null }
  } catch {
    return { exists: false, version: null }
  }
}

export async function checkBunInstalled(): Promise<CheckResult> {
  const result = await checkCommandExists("bun", ["--version"])

  if (!result.exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.BUN_INSTALLED],
      status: "fail",
      message: "Bun is not installed",
      details: [
        "Install Bun: curl -fsSL https://bun.sh/install | bash",
      ],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.BUN_INSTALLED],
    status: "pass",
    message: `Version ${result.version}`,
  }
}

export async function checkGitInstalled(): Promise<CheckResult> {
  const result = await checkCommandExists("git", ["--version"])

  if (!result.exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.GIT_INSTALLED],
      status: "warn",
      message: "Git is not installed",
      details: [
        "Git is recommended for version control",
        "Install: https://git-scm.com/downloads",
      ],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.GIT_INSTALLED],
    status: "pass",
    message: result.version ?? "Installed",
  }
}

export function getDependencyCheckDefinitions(): CheckDefinition[] {
  return [
    {
      id: CHECK_IDS.BUN_INSTALLED,
      name: CHECK_NAMES[CHECK_IDS.BUN_INSTALLED],
      category: "dependencies",
      check: checkBunInstalled,
      critical: true,
    },
    {
      id: CHECK_IDS.GIT_INSTALLED,
      name: CHECK_NAMES[CHECK_IDS.GIT_INSTALLED],
      category: "dependencies",
      check: checkGitInstalled,
      critical: false,
    },
  ]
}
```

**Step 5: Create checks index**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/checks/index.ts` with:

```typescript
import type { CheckDefinition } from "../types"
import { getInstallationCheckDefinitions, getPluginCheckDefinitions } from "./installation"
import { getConfigCheckDefinitions } from "./config"
import { getDependencyCheckDefinitions } from "./dependencies"

export * from "./installation"
export * from "./config"
export * from "./dependencies"

export function getAllCheckDefinitions(): CheckDefinition[] {
  return [
    ...getInstallationCheckDefinitions(),
    ...getConfigCheckDefinitions(),
    ...getPluginCheckDefinitions(),
    ...getDependencyCheckDefinitions(),
  ]
}
```

**Step 6: Verify checks compile**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/cli/doctor/checks/index.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/cli/doctor/checks/index.ts  X.XX KB
```

**Step 7: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/cli/doctor/checks/ && git commit -m "$(cat <<'EOF'
feat(cli/doctor): add health check implementations

Add checks for OpenCode installation, config validation, schema presence,
plugin/skill directories, and dependencies (bun, git).
EOF
)"
```

**If Task Fails:**

1. **Import errors:**
   - Check: All previous task files exist
   - Fix: Verify file paths match imports
   - Rollback: `rm -rf src/cli/doctor/checks/`

---

## Task 11: Create Doctor Formatter and Runner

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/formatter.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/runner.ts`
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/index.ts`

**Prerequisites:**
- Doctor checks created (Task 10)
- Doctor types/constants created (Task 9)

**Step 1: Create formatter.ts**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/formatter.ts` with:

```typescript
import color from "picocolors"
import type { CheckResult, DoctorSummary, CheckCategory, DoctorResult } from "./types"
import { SYMBOLS, STATUS_COLORS, CATEGORY_NAMES } from "./constants"

export function formatStatusSymbol(status: CheckResult["status"]): string {
  switch (status) {
    case "pass":
      return SYMBOLS.check
    case "fail":
      return SYMBOLS.cross
    case "warn":
      return SYMBOLS.warn
    case "skip":
      return SYMBOLS.skip
  }
}

export function formatCheckResult(result: CheckResult, verbose: boolean): string {
  const symbol = formatStatusSymbol(result.status)
  const colorFn = STATUS_COLORS[result.status]
  const name = colorFn(result.name)
  const message = color.dim(result.message)

  let line = `  ${symbol} ${name}`
  if (result.message) {
    line += ` ${SYMBOLS.arrow} ${message}`
  }

  if (verbose && result.details && result.details.length > 0) {
    const detailLines = result.details.map((d) => `      ${SYMBOLS.bullet} ${color.dim(d)}`).join("\n")
    line += "\n" + detailLines
  }

  return line
}

export function formatCategoryHeader(category: CheckCategory): string {
  const name = CATEGORY_NAMES[category] || category
  return `\n${color.bold(color.white(name))}\n${color.dim("\u2500".repeat(40))}`
}

export function formatSummary(summary: DoctorSummary): string {
  const lines: string[] = []

  lines.push(color.bold(color.white("Summary")))
  lines.push(color.dim("\u2500".repeat(40)))
  lines.push("")

  const passText = summary.passed > 0 ? color.green(`${summary.passed} passed`) : color.dim("0 passed")
  const failText = summary.failed > 0 ? color.red(`${summary.failed} failed`) : color.dim("0 failed")
  const warnText = summary.warnings > 0 ? color.yellow(`${summary.warnings} warnings`) : color.dim("0 warnings")
  const skipText = summary.skipped > 0 ? color.dim(`${summary.skipped} skipped`) : ""

  const parts = [passText, failText, warnText]
  if (skipText) parts.push(skipText)

  lines.push(`  ${parts.join(", ")}`)
  lines.push(`  ${color.dim(`Total: ${summary.total} checks in ${summary.duration}ms`)}`)

  return lines.join("\n")
}

export function formatHeader(): string {
  return `\n${color.bgCyan(color.white(" Ring Doctor "))}\n`
}

export function formatFooter(summary: DoctorSummary): string {
  if (summary.failed > 0) {
    return `\n${SYMBOLS.cross} ${color.red("Issues detected. Please review the errors above.")}\n`
  }
  if (summary.warnings > 0) {
    return `\n${SYMBOLS.warn} ${color.yellow("All systems operational with warnings.")}\n`
  }
  return `\n${SYMBOLS.check} ${color.green("All systems operational!")}\n`
}

export function formatJsonOutput(result: DoctorResult): string {
  return JSON.stringify(result, null, 2)
}
```

**Step 2: Create runner.ts**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/runner.ts` with:

```typescript
import type {
  DoctorOptions,
  DoctorResult,
  CheckDefinition,
  CheckResult,
  DoctorSummary,
  CheckCategory,
} from "./types"
import { getAllCheckDefinitions } from "./checks"
import { EXIT_CODES } from "./constants"
import {
  formatHeader,
  formatCategoryHeader,
  formatCheckResult,
  formatSummary,
  formatFooter,
  formatJsonOutput,
} from "./formatter"

export async function runCheck(check: CheckDefinition): Promise<CheckResult> {
  const start = performance.now()
  try {
    const result = await check.check()
    result.duration = Math.round(performance.now() - start)
    return result
  } catch (err) {
    return {
      name: check.name,
      status: "fail",
      message: err instanceof Error ? err.message : "Unknown error",
      duration: Math.round(performance.now() - start),
    }
  }
}

export function calculateSummary(results: CheckResult[], duration: number): DoctorSummary {
  return {
    total: results.length,
    passed: results.filter((r) => r.status === "pass").length,
    failed: results.filter((r) => r.status === "fail").length,
    warnings: results.filter((r) => r.status === "warn").length,
    skipped: results.filter((r) => r.status === "skip").length,
    duration: Math.round(duration),
  }
}

export function determineExitCode(results: CheckResult[]): number {
  const hasFailures = results.some((r) => r.status === "fail")
  return hasFailures ? EXIT_CODES.FAILURE : EXIT_CODES.SUCCESS
}

export function filterChecksByCategory(
  checks: CheckDefinition[],
  category?: CheckCategory
): CheckDefinition[] {
  if (!category) return checks
  return checks.filter((c) => c.category === category)
}

export function groupChecksByCategory(
  checks: CheckDefinition[]
): Map<CheckCategory, CheckDefinition[]> {
  const groups = new Map<CheckCategory, CheckDefinition[]>()

  for (const check of checks) {
    const existing = groups.get(check.category) ?? []
    existing.push(check)
    groups.set(check.category, existing)
  }

  return groups
}

const CATEGORY_ORDER: CheckCategory[] = [
  "installation",
  "configuration",
  "plugins",
  "dependencies",
]

export async function runDoctor(options: DoctorOptions): Promise<DoctorResult> {
  const start = performance.now()
  const allChecks = getAllCheckDefinitions()
  const filteredChecks = filterChecksByCategory(allChecks, options.category)
  const groupedChecks = groupChecksByCategory(filteredChecks)

  const results: CheckResult[] = []

  if (!options.json) {
    console.log(formatHeader())
  }

  for (const category of CATEGORY_ORDER) {
    const checks = groupedChecks.get(category)
    if (!checks || checks.length === 0) continue

    if (!options.json) {
      console.log(formatCategoryHeader(category))
    }

    for (const check of checks) {
      const result = await runCheck(check)
      results.push(result)

      if (!options.json) {
        console.log(formatCheckResult(result, options.verbose ?? false))
      }
    }
  }

  const duration = performance.now() - start
  const summary = calculateSummary(results, duration)
  const exitCode = determineExitCode(results)

  const doctorResult: DoctorResult = {
    results,
    summary,
    exitCode,
  }

  if (options.json) {
    console.log(formatJsonOutput(doctorResult))
  } else {
    console.log("")
    console.log(formatSummary(summary))
    console.log(formatFooter(summary))
  }

  return doctorResult
}
```

**Step 3: Create doctor index**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/doctor/index.ts` with:

```typescript
import type { DoctorOptions } from "./types"
import { runDoctor } from "./runner"

export async function doctor(options: DoctorOptions = {}): Promise<number> {
  const result = await runDoctor(options)
  return result.exitCode
}

export * from "./types"
export { runDoctor } from "./runner"
export { formatJsonOutput } from "./formatter"
```

**Step 4: Verify doctor module compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/cli/doctor/index.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/cli/doctor/index.ts  X.XX KB
```

**Step 5: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/cli/doctor/formatter.ts src/cli/doctor/runner.ts src/cli/doctor/index.ts && git commit -m "$(cat <<'EOF'
feat(cli/doctor): add formatter, runner, and main doctor module

Implement formatCheckResult, formatSummary, formatHeader for output.
Add runDoctor with category grouping and summary calculation.
EOF
)"
```

**If Task Fails:**

1. **Import errors:**
   - Check: Previous doctor files exist
   - Fix: Verify all src/cli/doctor/ files from Tasks 9-10
   - Rollback: `rm src/cli/doctor/{formatter,runner,index}.ts`

---

## Task 12: Create Version Command

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/version.ts`

**Prerequisites:**
- CLI constants created (Task 6)

**Step 1: Create version.ts**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/version.ts` with:

```typescript
import color from "picocolors"
import { PACKAGE_NAME } from "./constants"

interface VersionOptions {
  json?: boolean
}

interface VersionInfo {
  name: string
  version: string
  nodeVersion: string
  bunVersion: string | null
  platform: string
  arch: string
}

async function getBunVersion(): Promise<string | null> {
  try {
    const proc = Bun.spawn(["bun", "--version"], {
      stdout: "pipe",
      stderr: "pipe",
    })
    const output = await new Response(proc.stdout).text()
    await proc.exited
    if (proc.exitCode === 0) {
      return output.trim()
    }
    return null
  } catch {
    return null
  }
}

export async function version(options: VersionOptions = {}): Promise<number> {
  // Read version from package.json
  const packageJson = await import("../../package.json")
  const ver = packageJson.version ?? "unknown"

  const bunVersion = await getBunVersion()

  const info: VersionInfo = {
    name: PACKAGE_NAME,
    version: ver,
    nodeVersion: process.version,
    bunVersion,
    platform: process.platform,
    arch: process.arch,
  }

  if (options.json) {
    console.log(JSON.stringify(info, null, 2))
    return 0
  }

  console.log()
  console.log(`${color.bold(color.cyan(PACKAGE_NAME))} ${color.green(`v${ver}`)}`)
  console.log()
  console.log(`  ${color.dim("Node:")}    ${info.nodeVersion}`)
  if (info.bunVersion) {
    console.log(`  ${color.dim("Bun:")}     ${info.bunVersion}`)
  }
  console.log(`  ${color.dim("Platform:")} ${info.platform} (${info.arch})`)
  console.log()

  return 0
}
```

**Step 2: Verify version.ts compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/cli/version.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/cli/version.ts  X.XX KB
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/cli/version.ts && git commit -m "$(cat <<'EOF'
feat(cli): add version command

Show ring-opencode version, Node version, Bun version, and platform info.
Support --json flag for machine-readable output.
EOF
)"
```

**If Task Fails:**

1. **Import error:**
   - Check: package.json exists
   - Fix: Verify package.json path
   - Rollback: `rm src/cli/version.ts`

---

## Task 13: Create CLI Entry Point

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/index.ts`

**Prerequisites:**
- All CLI commands created (Tasks 8, 11, 12)

**Step 1: Create CLI entry point**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/cli/index.ts` with:

```typescript
#!/usr/bin/env bun
import { Command } from "commander"
import { install } from "./install"
import { doctor } from "./doctor"
import { version as versionCmd } from "./version"
import type { InstallArgs } from "./types"
import type { DoctorOptions } from "./doctor"

const packageJson = await import("../../package.json")
const VERSION = packageJson.version ?? "0.0.0"

const program = new Command()

program
  .name("ring")
  .description("Ring - CLI tools for OpenCode configuration and health checks")
  .version(VERSION, "-v, --version", "Show version number")

program
  .command("install")
  .description("Install and configure Ring with schema validation")
  .option("--no-tui", "Run in non-interactive mode")
  .option("--skip-validation", "Skip config validation after install")
  .addHelpText("after", `
Examples:
  $ ring install
  $ ring install --no-tui
  $ ring install --no-tui --skip-validation

This command:
  - Adds $schema to opencode.json for IDE autocomplete
  - Validates configuration against Ring schema
  - Creates opencode.json if it doesn't exist
`)
  .action(async (options) => {
    const args: InstallArgs = {
      tui: options.tui !== false,
      skipValidation: options.skipValidation ?? false,
    }
    const exitCode = await install(args)
    process.exit(exitCode)
  })

program
  .command("doctor")
  .description("Check Ring installation health and diagnose issues")
  .option("--verbose", "Show detailed diagnostic information")
  .option("--json", "Output results in JSON format")
  .option("--category <category>", "Run only specific category (installation, configuration, plugins, dependencies)")
  .addHelpText("after", `
Examples:
  $ ring doctor
  $ ring doctor --verbose
  $ ring doctor --json
  $ ring doctor --category configuration

Categories:
  installation     Check OpenCode installation
  configuration    Validate configuration files
  plugins          Check plugin and skill directories
  dependencies     Check runtime dependencies (bun, git)
`)
  .action(async (options) => {
    const doctorOptions: DoctorOptions = {
      verbose: options.verbose ?? false,
      json: options.json ?? false,
      category: options.category,
    }
    const exitCode = await doctor(doctorOptions)
    process.exit(exitCode)
  })

program
  .command("version")
  .description("Show detailed version information")
  .option("--json", "Output in JSON format")
  .action(async (options) => {
    const exitCode = await versionCmd({ json: options.json ?? false })
    process.exit(exitCode)
  })

program.parse()
```

**Step 2: Verify CLI entry point compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/cli/index.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/cli/index.ts  X.XX KB
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/cli/index.ts && git commit -m "$(cat <<'EOF'
feat(cli): add main CLI entry point with install, doctor, version commands

Implement ring CLI using Commander.js with:
- ring install: Add schema validation to config
- ring doctor: Health check system
- ring version: Show version info
EOF
)"
```

**If Task Fails:**

1. **Commander import error:**
   - Fix: `bun install commander`
   - Rollback: `rm src/cli/index.ts`

---

## Task 14: Create Main Module Export

**Files:**
- Create: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/index.ts`

**Prerequisites:**
- Config module created (Task 4)
- Shared utilities created (Task 3)

**Step 1: Create main index.ts**

Create file `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/src/index.ts` with:

```typescript
/**
 * Ring OpenCode
 *
 * Configuration schema validation and CLI tools for OpenCode.
 */

// Config exports
export {
  RingOpenCodeConfigSchema,
  RingHookNameSchema,
} from "./config"

export type {
  RingOpenCodeConfig,
  AgentConfig,
  AgentPermission,
  Permission,
  SkillsConfig,
  SkillDefinition,
  StateConfig,
  NotificationConfig,
  RingHookName,
} from "./config"

// Shared utilities
export {
  parseJsonc,
  parseJsoncSafe,
  readJsoncFile,
  detectConfigFile,
} from "./shared"

export type {
  JsoncParseResult,
} from "./shared"
```

**Step 2: Verify main module compiles**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun build src/index.ts --outdir /tmp/ring-test --target bun 2>&1 | head -5`

**Expected output:**
```
  src/index.ts  X.XX KB
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add src/index.ts && git commit -m "$(cat <<'EOF'
feat: add main module exports for config and shared utilities

Export RingOpenCodeConfigSchema, type definitions, and JSONC utilities
for external consumption.
EOF
)"
```

**If Task Fails:**

1. **Export errors:**
   - Check: All exported modules exist
   - Fix: Verify src/config/index.ts and src/shared/index.ts
   - Rollback: `rm src/index.ts`

---

## Task 15: Update opencode.json with Schema Reference

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/opencode.json`

**Prerequisites:**
- JSON schema generated (Task 5)

**Step 1: Read current opencode.json**

Run: `cat /Users/fredamaral/repos/fredcamaral/ring-for-opencode/opencode.json`

**Step 2: Update opencode.json with schema reference**

Replace the content of `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/opencode.json` with:

```json
{
  "$schema": "https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-opencode.schema.json",
  "version": "1.0.0",
  "name": "ring-opencode",
  "description": "Ring skills library for OpenCode - enforces proven software engineering practices",
  "permission": {
    "skill": {
      "*": "allow"
    },
    "edit": "ask",
    "bash": {
      "git *": "allow",
      "ls *": "allow",
      "pwd": "allow",
      "tree *": "allow",
      "head *": "allow",
      "tail *": "allow",
      "wc *": "allow",
      "find * -name *": "allow",
      "mkdir *": "ask",
      "rm *": "ask",
      "mv *": "ask",
      "cp *": "ask",
      "curl *": "deny",
      "wget *": "deny",
      "*": "ask"
    },
    "webfetch": "ask",
    "external_directory": "ask",
    "doom_loop": "ask"
  },
  "agent": {
    "build": {
      "mode": "primary",
      "model": "anthropic/claude-opus-4-20250514"
    },
    "plan": {
      "mode": "primary",
      "model": "anthropic/claude-opus-4-5-20251101",
      "permission": {
        "edit": "ask",
        "bash": "ask"
      }
    }
  }
}
```

**Step 3: Verify JSON is valid**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun -e "console.log(JSON.parse(require('fs').readFileSync('opencode.json', 'utf-8')).$schema)"`

**Expected output:**
```
https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-opencode.schema.json
```

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add opencode.json && git commit -m "$(cat <<'EOF'
feat(config): add $schema reference for IDE autocomplete

Add JSON Schema reference to opencode.json for IDE validation
and autocomplete support.
EOF
)"
```

**If Task Fails:**

1. **JSON parse error:**
   - Check: Valid JSON syntax
   - Fix: Validate with `jq . opencode.json`
   - Rollback: `git checkout -- opencode.json`

---

## Task 16: Build and Test CLI

**Files:**
- Build output in `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/dist/`

**Prerequisites:**
- All source files created (Tasks 1-15)

**Step 1: Run full build**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun run build`

**Expected output:**
```
Generating JSON Schema...
Done: assets/ring-opencode.schema.json
  src/index.ts  X.XX KB
  src/cli/index.ts  X.XX KB
```

**Step 2: Verify dist directory structure**

Run: `ls -la /Users/fredamaral/repos/fredcamaral/ring-for-opencode/dist/`

**Expected output:**
```
drwxr-xr-x  cli
-rw-r--r--  index.d.ts
-rw-r--r--  index.js
...
```

**Step 3: Test CLI help**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun dist/cli/index.js --help`

**Expected output:**
```
Usage: ring [options] [command]

Ring - CLI tools for OpenCode configuration and health checks

Options:
  -v, --version   Show version number
  -h, --help      display help for command

Commands:
  install         Install and configure Ring with schema validation
  doctor          Check Ring installation health and diagnose issues
  version         Show detailed version information
  help [command]  display help for command
```

**Step 4: Test ring doctor**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun dist/cli/index.js doctor`

**Expected output:**
```
 Ring Doctor

Installation
----------------------------------------
  [check] OpenCode Installation -> ...

Configuration
----------------------------------------
  [check] Configuration File -> Found
  [check] Configuration Validity -> Valid configuration
  [check] Schema Reference -> Schema reference present

...

Summary
----------------------------------------
  X passed, 0 failed, X warnings
  Total: X checks in Xms

[check] All systems operational!
```

**Step 5: Test ring version**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun dist/cli/index.js version`

**Expected output:**
```
ring-opencode v1.0.0

  Node:    vXX.X.X
  Bun:     X.X.X
  Platform: darwin (arm64)
```

**Step 6: Commit build artifacts (optional, or add to .gitignore)**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && echo "dist/" >> .gitignore && git add .gitignore && git commit -m "$(cat <<'EOF'
chore: add dist to gitignore

Build artifacts should not be committed to version control.
EOF
)"
```

**If Task Fails:**

1. **Build fails:**
   - Check: All source files exist with correct content
   - Run: `bun run typecheck` to see TypeScript errors
   - Fix: Address any compilation errors
   - Rollback: `rm -rf dist/`

2. **CLI doesn't work:**
   - Check: `bun dist/cli/index.js` runs
   - Fix: Verify all imports resolve correctly
   - Rollback: Rebuild after fixing source files

---

## Task 17: Run Code Review

**Prerequisites:**
- All tasks 1-16 completed
- Build successful

**Step 1: Dispatch all reviewers**

REQUIRED SUB-SKILL: Use requesting-code-review

Run code review on all new files in `src/` directory.

**Step 2: Handle findings by severity**

**Critical/High/Medium Issues:**
- Fix immediately (do NOT add TODO comments for these severities)
- Re-run all reviewers in parallel after fixes
- Repeat until zero Critical/High/Medium issues remain

**Low Issues:**
- Add `TODO(review):` comments in code at the relevant location
- Format: `TODO(review): [Issue description] (reported by [reviewer] on [date], severity: Low)`

**Cosmetic/Nitpick Issues:**
- Add `FIXME(nitpick):` comments in code at the relevant location
- Format: `FIXME(nitpick): [Issue description] (reported by [reviewer] on [date], severity: Cosmetic)`

**Step 3: Commit review fixes**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add -A && git commit -m "$(cat <<'EOF'
fix: address code review findings

Apply fixes for Critical/High/Medium issues from code review.
Add TODO(review) and FIXME(nitpick) comments for Low/Cosmetic items.
EOF
)"
```

**Step 4: Proceed only when:**
- Zero Critical/High/Medium issues remain
- All Low issues have TODO(review): comments added
- All Cosmetic issues have FIXME(nitpick): comments added

---

## Plan Checklist

- [x] Historical precedent queried (artifact-query --mode planning)
- [x] Historical Precedent section included in plan
- [x] Header with goal, architecture, tech stack, prerequisites
- [x] Verification commands with expected output
- [x] Tasks broken into bite-sized steps (2-5 min each)
- [x] Exact file paths for all files
- [x] Complete code (no placeholders)
- [x] Exact commands with expected output
- [x] Failure recovery steps for each task
- [x] Code review checkpoints after batches
- [x] Severity-based issue handling documented
- [x] Passes Zero-Context Test

---

## Summary

This plan creates a complete Zod-based configuration schema and CLI tooling for Ring:

**New Files Created:**
1. `src/shared/jsonc-parser.ts` - JSONC parsing utilities
2. `src/shared/index.ts` - Shared exports
3. `src/config/schema.ts` - Zod schema definitions
4. `src/config/index.ts` - Config exports
5. `src/cli/types.ts` - CLI type definitions
6. `src/cli/constants.ts` - CLI constants
7. `src/cli/config-manager.ts` - Config management utilities
8. `src/cli/install.ts` - Install command
9. `src/cli/doctor/types.ts` - Doctor types
10. `src/cli/doctor/constants.ts` - Doctor constants
11. `src/cli/doctor/checks/installation.ts` - Installation checks
12. `src/cli/doctor/checks/config.ts` - Config checks
13. `src/cli/doctor/checks/dependencies.ts` - Dependency checks
14. `src/cli/doctor/checks/index.ts` - Checks index
15. `src/cli/doctor/formatter.ts` - Output formatter
16. `src/cli/doctor/runner.ts` - Check runner
17. `src/cli/doctor/index.ts` - Doctor module
18. `src/cli/version.ts` - Version command
19. `src/cli/index.ts` - CLI entry point
20. `src/index.ts` - Main module exports
21. `script/build-schema.ts` - JSON schema generator
22. `assets/ring-opencode.schema.json` - Generated JSON schema
23. `tsconfig.json` - TypeScript configuration

**Modified Files:**
1. `package.json` - Dependencies and scripts
2. `opencode.json` - Schema reference
3. `.gitignore` - Exclude dist/

**CLI Commands Available:**
- `ring install` - Interactive configuration setup
- `ring doctor` - Health check system
- `ring version` - Version information
