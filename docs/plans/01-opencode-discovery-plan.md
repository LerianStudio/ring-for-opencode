# OpenCode Discovery Structure Migration Plan

> **For Agents:** REQUIRED SUB-SKILL: Use executing-plans to implement this plan task-by-task.

**Goal:** Restructure ring-for-opencode to use `.opencode/` directory for native OpenCode plugin discovery.

**Config Locations:** Ring reads user config from `~/.config/opencode/ring/config.jsonc` and project config from `.opencode/ring.jsonc` or `.ring/config.jsonc`.

**Architecture:** Move all Ring assets (commands, agents, skills, plugins) from repository root directories into a unified `.opencode/` directory structure. This follows the oh-my-opencode pattern where OpenCode automatically discovers plugins from `.opencode/` in the project root. The installer.sh already expects this structure but the source files don't exist yet.

**Tech Stack:** Bash (file operations), Git (version control), OpenCode (verification)

**Global Prerequisites:**
- Environment: macOS/Linux with bash
- Tools: git, ls, mv, mkdir
- Access: Write access to repository
- State: Clean working tree recommended (commit any pending changes first)

**Verification before starting:**
```bash
# Run ALL these commands and verify output:
git --version          # Expected: git version 2.x+
ls -la command/        # Expected: 16 .md files listed
ls -la agent/          # Expected: 16 .md files listed
ls -d skill/*/         # Expected: 30 directories listed
ls -la plugin/*.ts     # Expected: TypeScript plugin files listed
```

## Historical Precedent

**Query:** "opencode discovery migration structure directory"
**Index Status:** Empty (new project)

No historical data available. This is normal for new projects.
Proceeding with standard planning approach.

---

## Task Overview

| Task | Description | Duration |
|------|-------------|----------|
| 1 | Create .opencode directory structure | 2 min |
| 2 | Move command files | 3 min |
| 3 | Move agent files | 3 min |
| 4 | Move skill directories | 3 min |
| 5 | Move plugin files | 3 min |
| 6 | Create background-tasks.json | 2 min |
| 7 | Update .opencode/ring.jsonc location reference | 2 min |
| 8 | Verify installer.sh compatibility | 3 min |
| 9 | Update README.md | 5 min |
| 10 | Update AGENTS.md references | 3 min |
| 11 | Run verification tests | 5 min |
| 12 | Code Review Checkpoint | 5 min |
| 13 | Clean up empty directories | 2 min |
| 14 | Final verification and commit | 3 min |

**Total estimated time:** ~45 minutes

---

### Task 1: Create .opencode Directory Structure

**Files:**
- Create: `.opencode/`
- Create: `.opencode/command/`
- Create: `.opencode/agent/`
- Create: `.opencode/skill/`
- Create: `.opencode/plugin/`
- Create: `.opencode/state/`

**Prerequisites:**
- Tools: mkdir
- Working directory: `/Users/fredamaral/repos/fredcamaral/ring-for-opencode`

**Step 1: Create the .opencode directory structure**

Run:
```bash
mkdir -p .opencode/command .opencode/agent .opencode/skill .opencode/plugin .opencode/state
```

**Expected output:**
```
(no output - silent success)
```

**Step 2: Verify the structure was created**

Run:
```bash
ls -la .opencode/
```

**Expected output:**
```
drwxr-xr-x  - user  date  command
drwxr-xr-x  - user  date  agent
drwxr-xr-x  - user  date  skill
drwxr-xr-x  - user  date  plugin
drwxr-xr-x  - user  date  state
```

**If Task Fails:**

1. **Permission denied:**
   - Check: `ls -la .` (verify write permissions)
   - Fix: Ensure you have write access to the repository
   - Rollback: Not needed (no changes made)

2. **Directory already exists:**
   - This is fine, mkdir -p handles this
   - Proceed to next task

---

### Task 2: Move Command Files

**Files:**
- Move: `command/*.md` (16 files) to `.opencode/command/`

**Prerequisites:**
- Task 1 completed
- `.opencode/command/` directory exists

**Step 1: List current command files to verify count**

Run:
```bash
ls command/*.md | wc -l
```

**Expected output:**
```
16
```

**Step 2: Move all command files**

Run:
```bash
mv command/*.md .opencode/command/
```

**Expected output:**
```
(no output - silent success)
```

**Step 3: Verify files were moved**

Run:
```bash
ls .opencode/command/*.md | wc -l && ls .opencode/command/
```

**Expected output:**
```
16
brainstorm.md
codereview.md
commit.md
create-handoff.md
dev-cancel.md
dev-cycle.md
dev-refactor.md
dev-report.md
dev-status.md
execute-plan.md
explore-codebase.md
interview-me.md
lint.md
resume-handoff.md
worktree.md
write-plan.md
```

**If Task Fails:**

1. **No such file or directory:**
   - Check: `ls -la command/` (verify source exists)
   - Fix: Ensure command directory contains .md files
   - Rollback: `mv .opencode/command/*.md command/`

2. **Directory not empty (destination):**
   - Check: `ls .opencode/command/`
   - Fix: Remove conflicting files or merge manually
   - Rollback: `mv .opencode/command/*.md command/`

---

### Task 3: Move Agent Files

**Files:**
- Move: `agent/*.md` (16 files) to `.opencode/agent/`

**Prerequisites:**
- Task 1 completed
- `.opencode/agent/` directory exists

**Step 1: List current agent files to verify count**

Run:
```bash
ls agent/*.md | wc -l
```

**Expected output:**
```
16
```

**Step 2: Move all agent files**

Run:
```bash
mv agent/*.md .opencode/agent/
```

**Expected output:**
```
(no output - silent success)
```

**Step 3: Verify files were moved**

Run:
```bash
ls .opencode/agent/*.md | wc -l && ls .opencode/agent/
```

**Expected output:**
```
16
backend-engineer-golang.md
backend-engineer-typescript.md
business-logic-reviewer.md
code-reviewer.md
codebase-explorer.md
devops-engineer.md
frontend-bff-engineer-typescript.md
frontend-designer.md
frontend-engineer.md
nil-safety-reviewer.md
prompt-quality-reviewer.md
qa-analyst.md
security-reviewer.md
sre.md
test-reviewer.md
write-plan.md
```

**If Task Fails:**

1. **No such file or directory:**
   - Check: `ls -la agent/` (verify source exists)
   - Fix: Ensure agent directory contains .md files
   - Rollback: `mv .opencode/agent/*.md agent/`

---

### Task 4: Move Skill Directories

**Files:**
- Move: `skill/*/` (30 directories) to `.opencode/skill/`

**Prerequisites:**
- Task 1 completed
- `.opencode/skill/` directory exists

**Step 1: Count current skill directories**

Run:
```bash
ls -d skill/*/ | wc -l
```

**Expected output:**
```
30
```

**Step 2: Move all skill directories**

Run:
```bash
mv skill/*/ .opencode/skill/
```

**Expected output:**
```
(no output - silent success)
```

**Step 3: Verify directories were moved**

Run:
```bash
ls -d .opencode/skill/*/ | wc -l
```

**Expected output:**
```
30
```

**Step 4: Verify key skill structure is intact**

Run:
```bash
ls .opencode/skill/test-driven-development/
```

**Expected output:**
```
SKILL.md
```

**If Task Fails:**

1. **mv: cannot move to non-directory:**
   - Check: `ls -la .opencode/skill` (verify it's a directory)
   - Fix: Remove conflicting file, recreate directory
   - Rollback: `mv .opencode/skill/*/ skill/`

2. **Partial move:**
   - Check: `ls skill/` and `ls .opencode/skill/`
   - Fix: Move remaining directories manually
   - Rollback: Move all back: `mv .opencode/skill/*/ skill/`

---

### Task 5: Move Plugin Files

**Files:**
- Move: `plugin/*.ts` to `.opencode/plugin/`
- Move: `plugin/package.json` to `.opencode/plugin/`
- Move: `plugin/README.md` to `.opencode/plugin/`
- Move: `plugin/utils/` to `.opencode/plugin/`
- Move: `plugin/tests/` to `.opencode/plugin/`

**Prerequisites:**
- Task 1 completed
- `.opencode/plugin/` directory exists

**Step 1: List current plugin files**

Run:
```bash
ls -la plugin/
```

**Expected output:**
```
(list of .ts files, package.json, README.md, utils/, tests/)
```

**Step 2: Move plugin contents (preserving structure)**

Run:
```bash
cp -r plugin/* .opencode/plugin/ && rm -rf plugin/*
```

**Expected output:**
```
(no output - silent success)
```

**Note:** We use cp + rm instead of mv to handle mixed files/directories safely.

**Step 3: Verify plugin files were moved**

Run:
```bash
ls .opencode/plugin/
```

**Expected output:**
```
context-injection.ts
doubt-resolver.ts
index.ts
notification.ts
outcome-inference.ts
package.json
README.md
session-outcome.ts
session-start.ts
task-completion-check.ts
test-plugins.test.ts
test-plugins.ts
tests
utils
```

**Step 4: Verify utils directory was moved**

Run:
```bash
ls .opencode/plugin/utils/
```

**Expected output:**
```
(utility files)
```

**If Task Fails:**

1. **Permission denied:**
   - Check: File permissions on plugin directory
   - Fix: `chmod -R u+rw plugin/`
   - Rollback: `cp -r .opencode/plugin/* plugin/`

---

### Task 6: Create background-tasks.json

**Files:**
- Create: `.opencode/background-tasks.json`

**Prerequisites:**
- `.opencode/` directory exists

**Step 1: Create the background-tasks.json file**

Create file `.opencode/background-tasks.json` with content:

```json
[]
```

**Explanation:** This file stores background task state. It starts empty and gets populated by the session-start plugin when background agents are dispatched. The empty array format matches oh-my-opencode convention.

**Step 2: Verify file was created**

Run:
```bash
cat .opencode/background-tasks.json
```

**Expected output:**
```
[]
```

**If Task Fails:**

1. **File creation failed:**
   - Check: Directory exists: `ls -la .opencode/`
   - Fix: Create directory first if missing
   - Rollback: `rm .opencode/background-tasks.json`

---

### Task 7: Update .opencode/ring.jsonc Location Reference

**Files:**
- No changes needed to `.opencode/ring.jsonc`

**Prerequisites:**
- None

**Step 1: Verify .opencode/ring.jsonc doesn't need path updates**

Run:
```bash
cat .opencode/ring.jsonc | grep -E "(command|agent|skill|plugin)"
```

**Expected output:**
```
(no output or only permission-related lines, no path references)
```

**Explanation:** The .opencode/ring.jsonc file contains configuration for permissions and agent settings, not path references. OpenCode discovers resources from `.opencode/` automatically.

**Step 2: Confirm .opencode/ring.jsonc is valid JSON**

Run:
```bash
node -e "console.log(JSON.parse(require('fs').readFileSync('.opencode/ring.jsonc')).name)"
```

**Expected output:**
```
ring-opencode
```

**If Task Fails:**

1. **Invalid JSON:**
   - Check: `cat .opencode/ring.jsonc`
   - Fix: Correct JSON syntax errors
   - Rollback: `git checkout .opencode/ring.jsonc`

---

### Task 8: Verify installer.sh Compatibility

**Files:**
- Review: `installer.sh` (no changes expected)

**Prerequisites:**
- Tasks 1-5 completed (files moved to .opencode/)

**Step 1: Verify installer.sh SOURCE_ROOT points to .opencode**

Run:
```bash
grep "SOURCE_ROOT=" installer.sh
```

**Expected output:**
```
SOURCE_ROOT="$SCRIPT_DIR/.opencode"
```

**Explanation:** The installer already expects source files in `.opencode/`. This confirms our migration aligns with the existing installer design.

**Step 2: Verify installer copies all four directories**

Run:
```bash
grep -E "copy_tree_no_delete|plugin|skill|command|agent" installer.sh | head -10
```

**Expected output:**
```
(lines showing copy_tree_no_delete function and for loop iterating plugin skill command agent)
```

**Step 3: Dry-run test installer source detection**

Run:
```bash
ls .opencode/plugin .opencode/skill .opencode/command .opencode/agent 2>&1 | head -5
```

**Expected output:**
```
(list of files in each directory, confirming they exist)
```

**If Task Fails:**

1. **SOURCE_ROOT points elsewhere:**
   - This would require updating installer.sh
   - Update line: `SOURCE_ROOT="$SCRIPT_DIR/.opencode"`
   - Rollback: `git checkout installer.sh`

---

### Task 9: Update README.md

**Files:**
- Modify: `README.md`

**Prerequisites:**
- None

**Step 1: Read current README.md structure**

The current README.md has installation instructions that reference `opencode/.opencode`. Since files now live at `.opencode/` (not in an `opencode` subdirectory), we need to update the paths.

**Step 2: Update README.md**

Replace the entire content of `README.md` with:

```markdown
# Ring for OpenCode

Ring skills library converted for native OpenCode use.

## Installation

### Option 1: Direct Copy (Recommended)

1. Clone this repository:
   ```bash
   git clone https://github.com/fredcamaral/ring-for-opencode.git
   cd ring-for-opencode
   ```

2. Run the installer to copy to your OpenCode config:
   ```bash
   ./installer.sh
   ```

   This copies Ring assets to `~/.config/opencode/` and installs dependencies.

### Option 2: Per-Project Installation

Copy `.opencode/` directory to your project root:
```bash
cp -r ring-for-opencode/.opencode /path/to/your/project/
```

Copy `.opencode/ring.jsonc` to your project root (optional, for customization):
```bash
cp ring-for-opencode/.opencode/ring.jsonc /path/to/your/project/
```

Copy `AGENTS.md` to your project root:
```bash
cp ring-for-opencode/AGENTS.md /path/to/your/project/
```

## Directory Structure

```
ring-for-opencode/
├── .opencode/              # OpenCode discovery root
│   ├── command/            # Slash commands (/commit, /codereview, etc.)
│   ├── agent/              # AI agents (@code-reviewer, @write-plan, etc.)
│   ├── skill/              # Development skills (TDD, debugging, etc.)
│   ├── plugin/             # TypeScript plugins for session management
│   ├── state/              # Runtime state storage
│   └── background-tasks.json
├── .opencode/ring.jsonc           # OpenCode configuration
├── AGENTS.md               # Agent discovery documentation
├── installer.sh            # Installation script
└── README.md
```

## Usage

### Skills
Skills are automatically discovered. Use via the skill tool:
```
skill({ name: "test-driven-development" })
```

### Agents
Agents can be @ mentioned:
```
@code-reviewer review this code
@codebase-explorer analyze the architecture
```

### Commands
Commands are invoked with `/`:
```
/commit
/codereview
/brainstorm
```

## Components

| Type | Count | Description |
|------|-------|-------------|
| **Skills** | 30 | Development workflow skills |
| **Agents** | 16 | Specialized AI agents |
| **Commands** | 16 | Utility commands |
| **Plugins** | 7 | Session management and workflow enhancement |

### Core Skills
- `test-driven-development` - RED-GREEN-REFACTOR methodology
- `brainstorming` - Socratic design refinement
- `requesting-code-review` - Parallel 5-reviewer code review
- `systematic-debugging` - 4-phase debugging methodology
- `executing-plans` - Batch task execution with checkpoints

### Core Agents
- `@code-reviewer` - Code quality, architecture, design patterns
- `@security-reviewer` - Security vulnerabilities, OWASP Top 10
- `@business-logic-reviewer` - Business rules, domain correctness
- `@test-reviewer` - Test quality, coverage, anti-patterns
- `@nil-safety-reviewer` - Nil/null pointer safety (Go and TypeScript)
- `@codebase-explorer` - Deep codebase analysis
- `@write-plan` - Implementation planning

### Dev-Team Agents
- `@backend-engineer-golang` - Go backend development
- `@backend-engineer-typescript` - TypeScript backend development
- `@frontend-engineer` - Frontend development
- `@devops-engineer` - Infrastructure and CI/CD
- `@qa-analyst` - Testing and quality assurance
- `@sre` - Site reliability engineering

## Plugins

| Plugin | Purpose |
|--------|---------|
| `session-start.ts` | Initializes Ring context and injects critical rules |
| `context-injection.ts` | Preserves Ring rules during context compaction |
| `notification.ts` | Desktop notifications (cross-platform) |
| `session-outcome.ts` | Prompts for session outcome grade |
| `task-completion-check.ts` | Notifies when all todos complete |
| `outcome-inference.ts` | Infers session outcome from todo state |
| `doubt-resolver.ts` | Structured choice prompts for ambiguity |

## Platform Compatibility

This conversion adapts Ring's Claude Code format to OpenCode's native format:

| Ring (Claude Code) | OpenCode |
|--------------------|----------|
| `model: opus` | `model: anthropic/claude-opus-4-5-20251101` |
| `model: sonnet` | `model: anthropic/claude-sonnet-4-20250514` |
| `type: reviewer` | `mode: subagent` |
| Bash hooks | TypeScript plugins |
| `TodoWrite` tool | Task tracking / todo management |
| `Skill` tool | Skill invocation |

## Source

Converted from [Ring](https://github.com/LerianStudio/ring) Claude Code plugin.

## License

MIT
```

**Step 3: Verify README.md was updated**

Run:
```bash
grep "\.opencode/" README.md | head -3
```

**Expected output:**
```
├── .opencode/              # OpenCode discovery root
cp -r ring-for-opencode/.opencode /path/to/your/project/
```

**If Task Fails:**

1. **File write failed:**
   - Check: File permissions
   - Fix: `chmod u+w README.md`
   - Rollback: `git checkout README.md`

---

### Task 10: Update AGENTS.md References

**Files:**
- Review: `AGENTS.md` (verify references are correct)

**Prerequisites:**
- None

**Step 1: Check current AGENTS.md references**

Run:
```bash
grep "\.opencode" AGENTS.md
```

**Expected output:**
```
Skills are loaded on-demand via the native `skill` tool. See `.opencode/skill/` for available skills.
  See `.opencode/skill/shared-patterns/doubt-resolver.md`.
```

**Explanation:** AGENTS.md already references `.opencode/skill/` which is correct for the new structure. No changes needed.

**Step 2: Verify the referenced path will exist**

Run:
```bash
ls .opencode/skill/shared-patterns/ 2>/dev/null || echo "shared-patterns not found"
```

**Expected output:**
```
(list of files in shared-patterns, OR "shared-patterns not found")
```

**Step 3: If shared-patterns exists, verify doubt-resolver.md**

Run:
```bash
ls .opencode/skill/shared-patterns/doubt-resolver.md 2>/dev/null || echo "File not found - update AGENTS.md reference"
```

**If doubt-resolver.md is not at that path:** Update AGENTS.md line 18 to point to correct location.

**If Task Fails:**

1. **Reference path doesn't exist:**
   - Check: Where doubt-resolver actually lives
   - Fix: Update AGENTS.md to correct path
   - Rollback: `git checkout AGENTS.md`

---

### Task 11: Run Verification Tests

**Files:**
- Test: `.opencode/plugin/test-plugins.ts`

**Prerequisites:**
- Tasks 1-6 completed
- bun installed

**Step 1: Verify bun is available**

Run:
```bash
bun --version
```

**Expected output:**
```
1.x.x (any version)
```

**Step 2: Change to plugin directory and run tests**

Run:
```bash
cd .opencode/plugin && bun test-plugins.ts
```

**Expected output:**
```
(test results - may show warnings but should not show errors)
```

**Step 3: If tests fail due to path issues, check test file**

Run:
```bash
grep -n "skill" .opencode/plugin/test-plugins.test.ts | head -5
```

**Expected output:**
```
(lines showing skill path references)
```

**Step 4: Update test file paths if needed**

If tests reference old paths, the test file may need updating. Check for:
- References to `../skill/` should be `./skill/` or `.opencode/skill/`
- References to `command/` should be `.opencode/command/`

**If Task Fails:**

1. **bun not installed:**
   - Skip this verification
   - Manual verification: `ls .opencode/plugin/*.ts`

2. **Tests fail with path errors:**
   - Note the errors for fixing
   - Proceed with migration, fix tests separately

---

### Task 12: Code Review Checkpoint

**Files:**
- All files modified in Tasks 1-11

**Prerequisites:**
- Tasks 1-11 completed

**Step 1: Review all changes**

Run:
```bash
git status
```

**Expected output:**
```
Changes not staged for commit:
  deleted:    command/*.md (16 files)
  deleted:    agent/*.md (16 files)
  deleted:    skill/*/ (30 directories)
  deleted:    plugin/* (plugin files)
  modified:   README.md

Untracked files:
  .opencode/
```

**Step 2: Verify structure is correct**

Run:
```bash
find .opencode -type f | wc -l
```

**Expected output:**
```
(approximately 80+ files - commands, agents, skills, plugins)
```

**Step 3: Spot check key files exist**

Run:
```bash
ls .opencode/command/commit.md .opencode/agent/code-reviewer.md .opencode/skill/test-driven-development/SKILL.md .opencode/plugin/index.ts
```

**Expected output:**
```
.opencode/command/commit.md
.opencode/agent/code-reviewer.md
.opencode/skill/test-driven-development/SKILL.md
.opencode/plugin/index.ts
```

**If any issues found:** Fix before proceeding to Task 13.

---

### Task 13: Clean Up Empty Directories

**Files:**
- Remove: `command/` (empty)
- Remove: `agent/` (empty)
- Remove: `skill/` (empty)
- Remove: `plugin/` (empty)

**Prerequisites:**
- Task 12 completed (code review passed)

**Step 1: Verify directories are empty**

Run:
```bash
ls command/ agent/ skill/ plugin/ 2>&1
```

**Expected output:**
```
(empty or "No such file or directory" for each)
```

**Step 2: Remove empty directories**

Run:
```bash
rmdir command agent skill plugin 2>/dev/null || echo "Some directories not empty or already removed"
```

**Expected output:**
```
(no output or "Some directories not empty or already removed")
```

**Step 3: Verify cleanup**

Run:
```bash
ls -d command agent skill plugin 2>&1
```

**Expected output:**
```
ls: command: No such file or directory
ls: agent: No such file or directory
ls: skill: No such file or directory
ls: plugin: No such file or directory
```

**If Task Fails:**

1. **Directory not empty:**
   - Check: `ls <directory>/`
   - Fix: Move remaining files to .opencode/ or delete if obsolete
   - Note: Some hidden files like .DS_Store may remain

---

### Task 14: Final Verification and Commit

**Files:**
- All changes from Tasks 1-13

**Prerequisites:**
- All previous tasks completed

**Step 1: Final structure verification**

Run:
```bash
tree -L 2 .opencode/ 2>/dev/null || find .opencode -maxdepth 2 -type d
```

**Expected output:**
```
.opencode/
├── agent/
├── command/
├── plugin/
│   ├── tests/
│   └── utils/
├── skill/
│   ├── brainstorming/
│   ├── defense-in-depth/
│   ├── ... (28 more)
├── state/
└── background-tasks.json
```

**Step 2: Stage all changes**

Run:
```bash
git add -A
```

**Expected output:**
```
(no output - silent success)
```

**Step 3: Review staged changes**

Run:
```bash
git status --short | head -20
```

**Expected output:**
```
D  agent/backend-engineer-golang.md
D  agent/... (more deletions)
D  command/brainstorm.md
D  command/... (more deletions)
D  skill/brainstorming/SKILL.md
D  skill/... (more deletions)
D  plugin/context-injection.ts
D  plugin/... (more deletions)
M  README.md
A  .opencode/agent/backend-engineer-golang.md
A  .opencode/... (more additions)
```

**Step 4: Create commit**

Run:
```bash
git commit -m "$(cat <<'EOF'
refactor: migrate to .opencode/ directory structure for OpenCode discovery

Move all Ring assets to .opencode/ directory following oh-my-opencode pattern:
- command/*.md -> .opencode/command/
- agent/*.md -> .opencode/agent/
- skill/*/ -> .opencode/skill/
- plugin/* -> .opencode/plugin/

Add .opencode/background-tasks.json for background agent state.
Update README.md with new installation instructions.
Installer.sh already expects this structure.

This enables native OpenCode plugin discovery from .opencode/ directory.
EOF
)"
```

**Expected output:**
```
[main abc1234] refactor: migrate to .opencode/ directory structure for OpenCode discovery
 XX files changed, XXX insertions(+), XXX deletions(-)
 rename command/brainstorm.md => .opencode/command/brainstorm.md (100%)
 ...
```

**Step 5: Verify commit**

Run:
```bash
git log --oneline -1
```

**Expected output:**
```
abc1234 refactor: migrate to .opencode/ directory structure for OpenCode discovery
```

**If Task Fails:**

1. **Commit fails:**
   - Check: `git status` for unstaged changes
   - Fix: Stage all changes with `git add -A`
   - Rollback: `git reset HEAD~1` (if commit succeeded but was wrong)

2. **Pre-commit hook fails:**
   - Check: Hook output for errors
   - Fix: Address hook errors (linting, formatting)
   - Retry: `git commit` again after fixes

---

## Post-Migration Verification

After completing all tasks, verify OpenCode discovers the commands:

**Test 1: Start OpenCode in the repository**

```bash
cd /path/to/ring-for-opencode
opencode
```

**Test 2: In OpenCode, try a command**

```
/commit --help
```

**Expected:** OpenCode recognizes the /commit command from `.opencode/command/commit.md`

**Test 3: Try mentioning an agent**

```
@code-reviewer list your capabilities
```

**Expected:** OpenCode finds @code-reviewer from `.opencode/agent/code-reviewer.md`

**Test 4: Try loading a skill**

```
skill({ name: "test-driven-development" })
```

**Expected:** OpenCode loads the skill from `.opencode/skill/test-driven-development/SKILL.md`

---

## Rollback Procedure

If migration needs to be reverted:

```bash
# Reset to state before migration
git reset --hard HEAD~1

# Verify original structure restored
ls command/ agent/ skill/ plugin/
```

---

## Summary

This plan migrates ring-for-opencode to use the `.opencode/` directory structure:

| Before | After |
|--------|-------|
| `command/*.md` | `.opencode/command/*.md` |
| `agent/*.md` | `.opencode/agent/*.md` |
| `skill/*/` | `.opencode/skill/*/` |
| `plugin/*` | `.opencode/plugin/*` |
| (none) | `.opencode/background-tasks.json` |
| (none) | `.opencode/state/` |

The installer.sh already expects this structure, so no installer changes are needed.
