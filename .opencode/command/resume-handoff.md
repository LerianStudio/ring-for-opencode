---
description: Resume work from a handoff document with context analysis and validation
agent: build
subtask: false
---

Resume work from a handoff document through an interactive process. Handoffs contain critical context, learnings, and next steps from previous sessions.

## Process

### Step 1: Locate Handoff
- **If path provided**: Read the handoff file directly
- **If session-name provided**: Find most recent: `ls -t docs/handoffs/{session-name}/*.md | head -1`
- **If nothing provided**: List available: `find docs/handoffs -name "*.md" -type f | head -10`

### Step 2: Read and Analyze Handoff
Extract key sections:
- Task(s) and their statuses
- Recent changes
- Learnings (what worked, what failed)
- Critical references
- Action items and next steps

### Step 3: Verify Current State
```bash
# Check files mentioned in handoff exist
ls -la {files-from-handoff}

# Check git state
git log --oneline -5
git status
```

### Step 4: Present Analysis
```
I have analyzed the handoff from {date}.

**Original Tasks:**
- [Task 1]: {Status} -> {Current verification}

**Key Learnings:**
- {Learning} - {Still valid / Changed}

**Recommended Next Actions:**
1. {Most logical next step}
2. {Second priority}

Shall I proceed with {recommended action}?
```

### Step 5: Create Action Plan
Create task list from action items.

### Step 6: Begin Implementation
Start with first approved task, referencing learnings from handoff.

$ARGUMENTS

Provide path to handoff file or session name.
