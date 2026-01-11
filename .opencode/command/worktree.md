---
description: Create isolated git worktree with interactive setup
agent: build
subtask: false
---

Set up an isolated git worktree workspace for feature development.

## Process

### Step 1: Get Feature Name
Ask for the feature/branch name (e.g., "auth-system", "user-profiles", "payment-integration").

### Step 2: Check for Existing Directories
Priority order:
1. `.worktrees/` (preferred, hidden)
2. `worktrees/` (alternative)
3. If none exists, ask user preference

### Step 3: Verify .gitignore
If using project-local directory:
- Check if directory is in .gitignore
- If NOT, add to .gitignore and commit immediately

### Step 4: Create Worktree
```bash
# Detect project name
PROJECT_NAME=$(basename "$(git rev-parse --show-toplevel)")

# Create worktree
git worktree add <path> -b <branch-name>

# Navigate to worktree
cd <path>
```

### Step 5: Run Project Setup
Auto-detect and run appropriate setup:
- Node.js: `npm install` (if package.json)
- Rust: `cargo build` (if Cargo.toml)
- Python: `pip install -r requirements.txt` or `poetry install`
- Go: `go mod download` (if go.mod)

### Step 6: Verify Clean Baseline
Run appropriate test command for the project:
- If tests fail: Report failures and ask whether to proceed
- If tests pass: Report ready

### Step 7: Report Completion
```
Worktree ready at <full-path>
Tests passing (N tests, 0 failures)
Ready to implement <feature-name>
```

$ARGUMENTS

Provide the feature/branch name for the worktree.
