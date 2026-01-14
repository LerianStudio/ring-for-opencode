---
name: "ring:create-handoff"
description: Create a handoff document capturing current session state for future resumption
agent: build
subtask: false
---

Create a detailed handoff document that preserves the current session's context, learnings, and next steps for future sessions.

## Process

### Step 1: Determine Handoff Location
```bash
SESSION_NAME="${1:-$(git branch --show-current | sed 's/^feature\///')}"
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
mkdir -p "docs/handoffs/${SESSION_NAME}"
# File: docs/handoffs/${SESSION_NAME}/${TIMESTAMP}_${DESC}.md
```

### Step 2: Gather Metadata
```bash
git rev-parse HEAD
git branch --show-current
git remote get-url origin 2>/dev/null || echo "local"
date -u +"%Y-%m-%dT%H:%M:%SZ"
```

### Step 3: Write Handoff Document
Include:
1. **Task Summary**: What were you working on? What's the status?
2. **Critical References**: What files MUST be read to continue?
3. **Recent Changes**: What files were modified (with line numbers)?
4. **Learnings**: What worked? What failed? What decisions were made?
5. **Action Items**: What's next?

### Step 4: Confirm Creation
```
Handoff created: docs/handoffs/{session-name}/{timestamp}_{desc}.md

Resume in a new session with:
/ring:resume-handoff docs/handoffs/{session-name}/{timestamp}_{desc}.md
```

$ARGUMENTS

Optionally provide: [session-name] [description]
If not provided, infer from recent plan files or current git branch.
