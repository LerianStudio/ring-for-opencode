---
description: Execute plan in batches with review checkpoints
agent: build
subtask: true
---

Execute an existing implementation plan with controlled checkpoints and code review between batches.

## Process

### Step 1: Load and Review Plan
- Read the plan file completely
- Review for any questions or concerns
- If you need the user to choose between viable options, ask the user directly
- Raise issues before starting
- Create task list to track progress

### Step 2: Choose Execution Mode (MANDATORY)

| Mode | Behavior |
|------|----------|
| **One-go (autonomous)** | Execute all batches continuously with code review between each |
| **Batch (with review)** | Execute one batch, pause for human feedback after review |

### Step 3: Execute Batch
- Default batch size: first 3 tasks
- Mark each task in_progress, execute, then complete
- Dispatch to specialized agents when available

### Step 4: Run Code Review
After each batch, run all 5 reviewers in parallel:
- @code-reviewer - Architecture and patterns
- @business-logic-reviewer - Requirements and edge cases
- @security-reviewer - OWASP and auth validation
- @test-reviewer - Test quality and coverage
- @nil-safety-reviewer - Nil/null pointer safety

**Issue handling by severity:**

| Severity | Action |
|----------|--------|
| Critical/High/Medium | Fix immediately, re-run all reviewers |
| Low | Add `TODO(review):` comment in code |
| Cosmetic/Nitpick | Add `FIXME(nitpick):` comment in code |

### Step 5: Report and Continue
- **One-go mode**: Continue to next batch, report at final completion
- **Batch mode**: Show summary, wait for feedback before proceeding

### Step 6: Complete Development
After all tasks complete:
- Verify tests pass
- Present options for branch completion

## Arguments

Provide path to plan file: `docs/plans/YYYY-MM-DD-feature-name.md`

$ARGUMENTS
