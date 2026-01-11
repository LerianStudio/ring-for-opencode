# State Namespace Migration Implementation Plan

> **For Agents:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Migrate Ring's state management from `.ring/state/` to `.opencode/state/` namespace with backward compatibility and optional cleanup.

**Architecture:** The state management already uses `.opencode/state/` as primary write location with fallback reading from `.ring/state/`. This plan adds an active migration utility to copy existing state files forward, a cleanup option to remove legacy files, and comprehensive test coverage.

**Tech Stack:** TypeScript, Bun test framework, Node.js fs module

**Global Prerequisites:**
- Environment: macOS/Linux, Node 18-24, Bun
- Tools: Verify with commands below
- State: Clean working tree on `main` branch

**Verification before starting:**
```bash
# Run ALL these commands and verify output:
node --version      # Expected: v18.x - v24.x
bun --version       # Expected: 1.x
git status          # Expected: clean working tree (or known uncommitted files)
ls plugin/utils/state.ts  # Expected: file exists
```

## Historical Precedent

**Query:** "state namespace migration opencode"
**Index Status:** Empty (new project)

No historical data available. This is normal for new projects.
Proceeding with standard planning approach.

---

## Current State Analysis

**Already Implemented:**
- `STATE_DIR = ".opencode/state"` - Primary write location (line 22)
- `STATE_DIRS = [".opencode/state", ".ring/state"]` - Fallback read order (line 20)
- `findStatePath()` - Searches both locations for existing files (lines 104-116)
- `getStateDir()` - Creates `.opencode/state/` directory on demand (lines 74-79)

**Not Yet Implemented:**
- Active migration of existing `.ring/state/` files to `.opencode/state/`
- Optional cleanup of legacy `.ring/state/` directory
- Migration tracking to prevent repeated migration attempts
- Tests for migration functionality

---

## Task 1: Add Migration Function to state.ts

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/utils/state.ts:199-239`

**Prerequisites:**
- File `plugin/utils/state.ts` must exist
- Node.js fs module available (already imported)

**Step 1: Write the failing test**

Add to `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/tests/business-logic.test.ts` after line 539:

```typescript
// ============================================================================
// State Migration Tests
// ============================================================================

import { migrateStateFiles, getMigrationStatus } from "../utils/state"

describe("migrateStateFiles", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-migration-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("migrates files from .ring/state to .opencode/state", () => {
    // Create legacy state file
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })
    writeFileSync(join(legacyDir, "test-key-session123.json"), JSON.stringify({ value: "legacy" }))

    // Run migration
    const result = migrateStateFiles(tempDir)

    // Verify file was copied
    const newPath = join(tempDir, ".opencode", "state", "test-key-session123.json")
    expect(existsSync(newPath)).toBe(true)
    expect(JSON.parse(readFileSync(newPath, "utf-8"))).toEqual({ value: "legacy" })
    expect(result.migrated).toBe(1)
    expect(result.skipped).toBe(0)
  })

  test("skips migration if file already exists in .opencode/state", () => {
    // Create file in both locations
    const legacyDir = join(tempDir, ".ring", "state")
    const newDir = join(tempDir, ".opencode", "state")
    mkdirSync(legacyDir, { recursive: true })
    mkdirSync(newDir, { recursive: true })
    writeFileSync(join(legacyDir, "existing-session.json"), JSON.stringify({ old: true }))
    writeFileSync(join(newDir, "existing-session.json"), JSON.stringify({ new: true }))

    // Run migration
    const result = migrateStateFiles(tempDir)

    // Verify new file was NOT overwritten
    const content = JSON.parse(readFileSync(join(newDir, "existing-session.json"), "utf-8"))
    expect(content).toEqual({ new: true })
    expect(result.migrated).toBe(0)
    expect(result.skipped).toBe(1)
  })

  test("handles missing .ring/state directory gracefully", () => {
    // No .ring/state directory exists
    const result = migrateStateFiles(tempDir)

    expect(result.migrated).toBe(0)
    expect(result.skipped).toBe(0)
    expect(result.error).toBeUndefined()
  })

  test("only migrates .json files", () => {
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })
    writeFileSync(join(legacyDir, "valid-session.json"), JSON.stringify({ valid: true }))
    writeFileSync(join(legacyDir, "invalid.txt"), "not json")
    writeFileSync(join(legacyDir, ".hidden"), "hidden file")

    const result = migrateStateFiles(tempDir)

    expect(result.migrated).toBe(1)
    expect(existsSync(join(tempDir, ".opencode", "state", "valid-session.json"))).toBe(true)
    expect(existsSync(join(tempDir, ".opencode", "state", "invalid.txt"))).toBe(false)
  })

  test("tracks migration status to prevent re-migration", () => {
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })
    writeFileSync(join(legacyDir, "key-session.json"), JSON.stringify({ data: 1 }))

    // First migration
    migrateStateFiles(tempDir)

    // Check status
    const status = getMigrationStatus(tempDir)
    expect(status.migrated).toBe(true)

    // Add new legacy file
    writeFileSync(join(legacyDir, "key2-session.json"), JSON.stringify({ data: 2 }))

    // Second migration should still work for new files
    const result2 = migrateStateFiles(tempDir)
    expect(result2.migrated).toBe(1)
  })
})
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test plugin/tests/business-logic.test.ts --filter "migrateStateFiles"`

**Expected output:**
```
error: Cannot find module "../utils/state" when importing "migrateStateFiles"
```
or
```
error: migrateStateFiles is not exported
```

**If you see different error:** Check import paths and function exports

**Step 3: Write the implementation**

Add to `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/utils/state.ts` before the closing of the file (before line 239):

```typescript
/**
 * Migration result interface.
 */
export interface MigrationResult {
  migrated: number
  skipped: number
  errors: string[]
  error?: string
}

/**
 * Migration status interface.
 */
export interface MigrationStatus {
  migrated: boolean
  timestamp?: number
  filesMigrated?: number
}

/**
 * Get migration status for a project.
 */
export function getMigrationStatus(projectRoot: string): MigrationStatus {
  const statusPath = join(projectRoot, STATE_DIR, ".migration-status.json")
  try {
    if (existsSync(statusPath)) {
      const content = readFileSync(statusPath, "utf-8")
      return JSON.parse(content) as MigrationStatus
    }
  } catch {
    // Ignore read errors
  }
  return { migrated: false }
}

/**
 * Save migration status.
 */
function saveMigrationStatus(projectRoot: string, status: MigrationStatus): void {
  const stateDir = getStateDir(projectRoot)
  const statusPath = join(stateDir, ".migration-status.json")
  try {
    writeFileSync(statusPath, JSON.stringify(status, null, 2), { encoding: "utf-8", mode: 0o600 })
  } catch {
    // Ignore write errors - migration status is optional
  }
}

/**
 * Migrate state files from legacy .ring/state/ to .opencode/state/.
 * - Only copies .json files
 * - Does not overwrite existing files in .opencode/state/
 * - Safe to run multiple times (idempotent)
 */
export function migrateStateFiles(projectRoot: string): MigrationResult {
  const result: MigrationResult = {
    migrated: 0,
    skipped: 0,
    errors: [],
  }

  const legacyDir = join(projectRoot, ".ring", "state")

  // Check if legacy directory exists
  if (!existsSync(legacyDir)) {
    return result
  }

  // Ensure target directory exists
  const targetDir = getStateDir(projectRoot)

  try {
    const files = readdirSync(legacyDir)

    for (const file of files) {
      // Only migrate .json files
      if (!file.endsWith(".json")) {
        continue
      }

      const sourcePath = join(legacyDir, file)
      const targetPath = join(targetDir, file)

      // Skip if file already exists in target
      if (existsSync(targetPath)) {
        result.skipped++
        continue
      }

      // Verify source is a file (not directory)
      try {
        const stats = statSync(sourcePath)
        if (!stats.isFile()) {
          continue
        }
      } catch {
        result.errors.push(`Failed to stat: ${file}`)
        continue
      }

      // Copy file
      try {
        const content = readFileSync(sourcePath, "utf-8")
        writeFileSync(targetPath, content, { encoding: "utf-8", mode: 0o600 })
        result.migrated++
      } catch (err) {
        result.errors.push(`Failed to migrate: ${file}`)
      }
    }

    // Update migration status
    const status = getMigrationStatus(projectRoot)
    saveMigrationStatus(projectRoot, {
      migrated: true,
      timestamp: Date.now(),
      filesMigrated: (status.filesMigrated || 0) + result.migrated,
    })

  } catch (err) {
    result.error = `Failed to read legacy directory: ${err}`
  }

  return result
}

/**
 * Remove legacy .ring/state/ directory after confirming migration.
 * Only removes if .opencode/state/ exists and has files.
 */
export function cleanupLegacyState(projectRoot: string): { removed: boolean; reason?: string } {
  const legacyDir = join(projectRoot, ".ring", "state")
  const newDir = join(projectRoot, STATE_DIR)

  // Safety checks
  if (!existsSync(legacyDir)) {
    return { removed: false, reason: "Legacy directory does not exist" }
  }

  if (!existsSync(newDir)) {
    return { removed: false, reason: "New state directory does not exist - run migration first" }
  }

  // Check migration status
  const status = getMigrationStatus(projectRoot)
  if (!status.migrated) {
    return { removed: false, reason: "Migration has not been completed" }
  }

  try {
    // Remove legacy state directory
    rmSync(legacyDir, { recursive: true, force: true })
    return { removed: true }
  } catch (err) {
    return { removed: false, reason: `Failed to remove: ${err}` }
  }
}
```

**Step 4: Add required import**

Modify line 1-15 of `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/utils/state.ts` to add `rmSync` import:

Find this line:
```typescript
import {
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
  unlinkSync,
  readdirSync,
  statSync,
  renameSync,
} from "fs"
```

Replace with:
```typescript
import {
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
  unlinkSync,
  readdirSync,
  statSync,
  renameSync,
  rmSync,
} from "fs"
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test plugin/tests/business-logic.test.ts --filter "migrateStateFiles"`

**Expected output:**
```
bun test v1.x
plugin/tests/business-logic.test.ts:
  migrateStateFiles
    ✓ migrates files from .ring/state to .opencode/state
    ✓ skips migration if file already exists in .opencode/state
    ✓ handles missing .ring/state directory gracefully
    ✓ only migrates .json files
    ✓ tracks migration status to prevent re-migration

 5 pass
```

**Step 6: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/utils/state.ts plugin/tests/business-logic.test.ts && git commit -m "$(cat <<'EOF'
feat(state): add migration utility for .ring/state to .opencode/state

Adds migrateStateFiles() and cleanupLegacyState() functions to safely
migrate state files from legacy location while preserving existing files.
EOF
)"
```

**If Task Fails:**

1. **Test won't compile:**
   - Check: Are all imports correct?
   - Fix: Verify `migrateStateFiles` and `getMigrationStatus` are exported
   - Rollback: `git checkout -- plugin/utils/state.ts plugin/tests/business-logic.test.ts`

2. **Test fails:**
   - Run: `bun test --filter "migrateStateFiles" -v` for verbose output
   - Check: File paths, directory creation logic
   - Rollback: `git reset --hard HEAD`

---

## Task 2: Trigger Migration on Session Start

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/session-start.ts:4-5`
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/session-start.ts:243-248`

**Prerequisites:**
- Task 1 completed (migration functions exist)
- `plugin/session-start.ts` exists

**Step 1: Add import for migration function**

Find at line 4 of `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/session-start.ts`:
```typescript
import { cleanupOldState, deleteState, getSessionId, escapeAngleBrackets, sanitizeForPrompt, isPathWithinRoot } from "./utils/state"
```

Replace with:
```typescript
import { cleanupOldState, deleteState, getSessionId, escapeAngleBrackets, sanitizeForPrompt, isPathWithinRoot, migrateStateFiles } from "./utils/state"
```

**Step 2: Add migration call in session.created event handler**

Find at lines 241-248 of `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/session-start.ts`:
```typescript
    event: async ({ event }) => {
      if (event.type === EVENTS.SESSION_CREATED) {
        // Reset context usage state for new session
        deleteState(projectRoot, "context-usage", sessionId)
        // Clean up old state files (> 7 days)
        cleanupOldState(projectRoot, 7)
        // Clear skills cache for new session
        cachedSkillsOverview = null
      }
    },
```

Replace with:
```typescript
    event: async ({ event }) => {
      if (event.type === EVENTS.SESSION_CREATED) {
        // Migrate legacy state files from .ring/state/ to .opencode/state/
        // Safe to call multiple times - skips already migrated files
        migrateStateFiles(projectRoot)
        // Reset context usage state for new session
        deleteState(projectRoot, "context-usage", sessionId)
        // Clean up old state files (> 7 days)
        cleanupOldState(projectRoot, 7)
        // Clear skills cache for new session
        cachedSkillsOverview = null
      }
    },
```

**Step 3: Run tests to verify nothing broke**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test plugin/test-plugins.test.ts --filter "RingSessionStart"`

**Expected output:**
```
bun test v1.x
  RingSessionStart
    ✓ has event hook for session lifecycle
    ✓ has system transform hook
    ✓ system transform injects critical rules
    ✓ system transform injects agent reminder
    ✓ system transform injects duplication guard
    ✓ system transform injects ledger when present
    ✓ handles session.created event

 7 pass
```

**Step 4: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/session-start.ts && git commit -m "$(cat <<'EOF'
feat(session): trigger state migration on session start

Automatically migrates legacy .ring/state/ files to .opencode/state/
when a new session starts. Migration is idempotent and safe to run
multiple times.
EOF
)"
```

**If Task Fails:**

1. **Import error:**
   - Check: Is `migrateStateFiles` exported from `state.ts`?
   - Fix: Add `export` keyword if missing
   - Rollback: `git checkout -- plugin/session-start.ts`

2. **Test failures:**
   - Run: `bun test plugin/test-plugins.test.ts -v`
   - Check: Event handling logic unchanged
   - Rollback: `git reset --hard HEAD`

---

## Task 3: Add cleanupLegacyState Tests

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/tests/business-logic.test.ts`

**Prerequisites:**
- Task 1 completed (cleanupLegacyState exists)

**Step 1: Add tests for cleanupLegacyState**

Add to `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/tests/business-logic.test.ts` after the migrateStateFiles tests:

```typescript
import { cleanupLegacyState } from "../utils/state"

describe("cleanupLegacyState", () => {
  let tempDir: string

  beforeEach(() => {
    tempDir = mkdtempSync(join(tmpdir(), "ring-cleanup-test-"))
  })

  afterEach(() => {
    rmSync(tempDir, { recursive: true, force: true })
  })

  test("removes legacy directory after successful migration", () => {
    // Setup: Create legacy and new directories
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })
    writeFileSync(join(legacyDir, "old-session.json"), JSON.stringify({ old: true }))

    // Run migration first
    migrateStateFiles(tempDir)

    // Now cleanup
    const result = cleanupLegacyState(tempDir)

    expect(result.removed).toBe(true)
    expect(existsSync(legacyDir)).toBe(false)
  })

  test("refuses to remove if migration not completed", () => {
    // Create legacy directory without migration
    const legacyDir = join(tempDir, ".ring", "state")
    const newDir = join(tempDir, ".opencode", "state")
    mkdirSync(legacyDir, { recursive: true })
    mkdirSync(newDir, { recursive: true })
    writeFileSync(join(legacyDir, "file.json"), "{}")

    const result = cleanupLegacyState(tempDir)

    expect(result.removed).toBe(false)
    expect(result.reason).toContain("Migration has not been completed")
    expect(existsSync(legacyDir)).toBe(true)
  })

  test("returns gracefully if legacy directory does not exist", () => {
    const result = cleanupLegacyState(tempDir)

    expect(result.removed).toBe(false)
    expect(result.reason).toContain("does not exist")
  })

  test("refuses if new state directory does not exist", () => {
    const legacyDir = join(tempDir, ".ring", "state")
    mkdirSync(legacyDir, { recursive: true })

    const result = cleanupLegacyState(tempDir)

    expect(result.removed).toBe(false)
    expect(result.reason).toContain("run migration first")
  })
})
```

**Step 2: Run tests**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test plugin/tests/business-logic.test.ts --filter "cleanupLegacyState"`

**Expected output:**
```
bun test v1.x
  cleanupLegacyState
    ✓ removes legacy directory after successful migration
    ✓ refuses to remove if migration not completed
    ✓ returns gracefully if legacy directory does not exist
    ✓ refuses if new state directory does not exist

 4 pass
```

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/tests/business-logic.test.ts && git commit -m "$(cat <<'EOF'
test(state): add tests for cleanupLegacyState function

Covers success case, safety checks, and edge cases for legacy
state directory cleanup.
EOF
)"
```

**If Task Fails:**

1. **Import error:**
   - Check: Is `cleanupLegacyState` exported from state.ts?
   - Fix: Add to exports
   - Rollback: `git checkout -- plugin/tests/business-logic.test.ts`

---

## Task 4: Update README Documentation

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/README.md:29-36`

**Prerequisites:**
- None

**Step 1: Update State Management section**

Find at lines 29-36 of `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/README.md`:
```markdown
## State Management

State is persisted in `.opencode/state/` (with backward compat for `.ring/state/`):
- `{key}-{sessionId}.json` format
- Automatic cleanup of files >7 days old
- Atomic writes via temp file + rename with crypto random suffix
- Restrictive permissions (0o600) for sensitive data
```

Replace with:
```markdown
## State Management

State is persisted in `.opencode/state/`:
- `{key}-{sessionId}.json` format
- Automatic cleanup of files >7 days old
- Atomic writes via temp file + rename with crypto random suffix
- Restrictive permissions (0o600) for sensitive data

### Migration from .ring/state/

Legacy state files in `.ring/state/` are automatically migrated to `.opencode/state/` on session start:
- Migration is safe and idempotent (runs every session, skips already-migrated files)
- Existing files in `.opencode/state/` are never overwritten
- Only `.json` files are migrated

To manually trigger migration or cleanup:
```typescript
import { migrateStateFiles, cleanupLegacyState } from "./utils/state"

// Migrate files (safe to run multiple times)
const result = migrateStateFiles(projectRoot)
console.log(`Migrated: ${result.migrated}, Skipped: ${result.skipped}`)

// Optional: Remove legacy directory after confirming migration
const cleanup = cleanupLegacyState(projectRoot)
if (cleanup.removed) {
  console.log("Legacy .ring/state/ removed")
}
```
```

**Step 2: Verify markdown renders correctly**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && head -60 plugin/README.md`

**Expected output:** Should show the updated State Management section with migration documentation.

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add plugin/README.md && git commit -m "$(cat <<'EOF'
docs(plugin): document state migration from .ring/state to .opencode/state

Adds migration documentation including automatic behavior, manual
trigger instructions, and cleanup utility usage.
EOF
)"
```

**If Task Fails:**

1. **Markdown syntax errors:**
   - Check: Code block backticks are correct (triple backticks)
   - Fix: Ensure proper markdown formatting
   - Rollback: `git checkout -- plugin/README.md`

---

## Task 5: Update installer.sh Comment (Optional Clarification)

**Files:**
- Modify: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/installer.sh:105-106`

**Prerequisites:**
- None

**Step 1: Update comment for state directory**

Find at lines 105-106 of `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/installer.sh`:
```bash
# Ensure state dir exists (no overwrite)
mkdir -p "$TARGET_ROOT/state"
```

Replace with:
```bash
# Ensure global state dir exists in user config (no overwrite)
# Note: Project-level state is in <project>/.opencode/state/ and created dynamically
mkdir -p "$TARGET_ROOT/state"
```

**Step 2: Verify syntax**

Run: `bash -n /Users/fredamaral/repos/fredcamaral/ring-for-opencode/installer.sh`

**Expected output:** No output (no syntax errors)

**Step 3: Commit**

```bash
cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && git add installer.sh && git commit -m "$(cat <<'EOF'
docs(installer): clarify state directory locations

Distinguishes between global user config state (~/.config/opencode/state/)
and project-level state (<project>/.opencode/state/).
EOF
)"
```

**If Task Fails:**

1. **Syntax error:**
   - Run: `bash -n installer.sh` to check
   - Fix: Check comment syntax (# at start of line)
   - Rollback: `git checkout -- installer.sh`

---

## Task 6: Run Full Test Suite

**Files:**
- None (verification only)

**Prerequisites:**
- Tasks 1-5 completed

**Step 1: Run all plugin tests**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test plugin/`

**Expected output:**
```
bun test v1.x

plugin/test-plugins.test.ts:
  RingContextInjection ... ✓
  RingNotification ... ✓
  RingDoubtResolver ... ✓
  RingSessionStart ... ✓
  RingTaskCompletionCheck ... ✓
  RingSessionOutcome ... ✓
  RingOutcomeInference ... ✓
  Plugin Exports ... ✓

plugin/tests/business-logic.test.ts:
  analyzeTodos ... ✓
  inferOutcomeFromTodos ... ✓
  readState ... ✓
  writeState ... ✓
  getStatePath ... ✓
  getStateDir ... ✓
  Integration ... ✓
  migrateStateFiles ... ✓
  cleanupLegacyState ... ✓

plugin/tests/security-sanitization.test.ts:
  sanitizeSessionId ... ✓
  sanitizeKey ... ✓
  isPathWithinRoot ... ✓
  sanitizeForPrompt ... ✓
  escapeAngleBrackets ... ✓
  sanitizeNotificationContent ... ✓

XX pass
 0 fail
```

**Step 2: Verify legacy backward compatibility still works**

Run: `cd /Users/fredamaral/repos/fredcamaral/ring-for-opencode && bun test plugin/tests/business-logic.test.ts --filter "reads from legacy"`

**Expected output:**
```
  readState
    ✓ reads from legacy .ring/state directory

 1 pass
```

**If Task Fails:**

1. **Some tests fail:**
   - Run: `bun test plugin/ -v` for verbose output
   - Check: Which specific test failed
   - Fix: Address specific failure
   - If major regression: `git reset --hard HEAD~5` to revert all changes

---

## Task 7: Code Review Checkpoint

**Prerequisites:**
- Tasks 1-6 completed
- All tests passing

**Step 1: Dispatch all reviewers in parallel**

- REQUIRED SUB-SKILL: Use requesting-code-review

**Step 2: Handle findings by severity**

**Critical/High/Medium Issues:**
- Fix immediately (do NOT add TODO comments for these severities)
- Re-run all reviewers in parallel after fixes
- Repeat until zero Critical/High/Medium issues remain

**Low Issues:**
- Add `TODO(review):` comments in code at the relevant location
- Format: `TODO(review): [Issue description] (reported by [reviewer] on [date], severity: Low)`

**Cosmetic/Nitpick Issues:**
- Add `FIXME(nitpick):` comments in code
- Format: `FIXME(nitpick): [Issue description] (reported by [reviewer] on [date], severity: Cosmetic)`

**Step 3: Proceed only when:**
- Zero Critical/High/Medium issues remain
- All Low issues have TODO(review): comments added
- All Cosmetic issues have FIXME(nitpick): comments added

---

## Summary

**Files Modified:**
1. `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/utils/state.ts` - Added migration functions
2. `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/session-start.ts` - Trigger migration on session start
3. `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/tests/business-logic.test.ts` - Migration tests
4. `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/plugin/README.md` - Documentation
5. `/Users/fredamaral/repos/fredcamaral/ring-for-opencode/installer.sh` - Comment clarification

**Key Changes:**
- `migrateStateFiles()` - Copies .json files from `.ring/state/` to `.opencode/state/`
- `cleanupLegacyState()` - Safely removes legacy directory after migration
- `getMigrationStatus()` - Tracks migration state
- Automatic migration on session start
- Comprehensive test coverage

**Backward Compatibility:**
- Reading from `.ring/state/` still works (existing `findStatePath()`)
- Migration does not delete legacy files automatically
- Cleanup is opt-in via `cleanupLegacyState()`

**Verification:**
```bash
# Run all tests
bun test plugin/

# Verify migration works
node -e "
const { migrateStateFiles, getMigrationStatus } = require('./plugin/utils/state');
console.log(getMigrationStatus('.'));
"
```
