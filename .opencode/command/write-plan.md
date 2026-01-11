---
description: Create detailed implementation plan with bite-sized tasks
agent: plan
subtask: true
---

Create a comprehensive implementation plan for a feature with exact file paths, complete code examples, and verification steps.

## Plan Requirements (Zero-Context Test)

Every plan must be executable with only the document - no assumptions:

- **Exact file paths** - Never "somewhere in src"
- **Complete code** - Never "add validation here"
- **Verification commands** - With expected output
- **Failure recovery** - What to do when things go wrong
- **Code review checkpoints** - After task batches
- **Agent recommendations** - Which agent for each task type

## Process

### Step 1: Dispatch Planning Agent
Explore the codebase to understand architecture, identify files needing modification, break feature into bite-sized tasks (2-5 minutes each).

### Step 2: Create Plan Document
Include:
- Header with goal, architecture, tech stack, prerequisites
- Bite-sized tasks with exact file paths
- Complete, copy-paste ready code for each task
- Exact verification commands with expected output
- Code review checkpoints after task batches
- Recommended agents for each task type
- Failure recovery steps

### Step 3: Save Plan
Save to: `docs/plans/YYYY-MM-DD-<feature-name>.md`

### Step 4: Choose Execution Mode
After plan is ready, ask user:
- **Execute now** - Start implementation immediately
- **Execute in parallel session** - Open new session in worktree
- **Save for later** - Keep plan for manual review

## Agent Selection

> **Note:** The specialized agents below are available when the dev-team skills are installed.
> Without them, use `@write-plan` for planning and `@codebase-explorer` for research.

| Task Type | Recommended Agent |
|-----------|-------------------|
| Backend (Go) | @backend-engineer-golang |
| Backend (TypeScript) | @backend-engineer-typescript |
| Frontend (BFF/API) | @frontend-bff-engineer-typescript |
| Infrastructure | @devops-engineer |
| Testing | @qa-analyst |
| Reliability | @sre |

$ARGUMENTS
