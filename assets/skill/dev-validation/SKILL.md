---
name: "ring:dev-validation"
description: |
  Development cycle validation gate (Gate 5) - validates all acceptance criteria are met
  and requires explicit user approval before completion.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: workflow
  trigger:
    - After review gate passes (Gate 4)
    - Implementation and tests complete
    - Need user sign-off on acceptance criteria
  sequence:
    after: [requesting-code-review]
---

# Dev Validation (Gate 5)

## Overview

Final validation gate requiring explicit user approval. Present evidence that each acceptance criterion is met and obtain APPROVED or REJECTED decision.

**Core principle:** Passing tests and code review do not guarantee requirements are met. User validation confirms implementation matches intent.

## Self-Approval Prohibition

**HARD GATE:** The agent that implemented code CANNOT approve validation for that same code.

| Scenario | Allowed? | Action |
|----------|----------|--------|
| Different agent/human approves | YES | Proceed |
| Same agent self-approves | NO | STOP - requires external approval |
| User explicitly approves | YES | User approval always valid |

## Severity Calibration

| Severity | Criteria | Action Required |
|----------|----------|-----------------|
| **CRITICAL** | AC completely unmet | MUST fix before approval |
| **HIGH** | AC partially met or degraded | MUST fix before approval |
| **MEDIUM** | Edge case gap | User decides |
| **LOW** | Quality issue, requirement met | User decides |

## Approval Format (MANDATORY)

User MUST respond with exactly one of:
- **"APPROVED"** - All criteria verified, proceed
- **"REJECTED: [reason]"** - Issues found, fix and revalidate

**NOT acceptable:**
- "Looks good" (vague)
- "👍" (ambiguous)
- Silence (not a response)
- "Approved with minor issues" (partial = REJECTED)

## Ambiguous Response Handling

| Response | Status | Action |
|----------|--------|--------|
| "Looks good" | AMBIGUOUS | Ask for explicit APPROVED |
| "Sure" / "Ok" | AMBIGUOUS | Ask for explicit APPROVED |
| "👍" / "✅" | AMBIGUOUS | Ask for explicit APPROVED |
| "APPROVED" | VALID | Proceed |
| "REJECTED: [reason]" | VALID | Return to Gate 0 |

## Workflow Steps

### Step 1: Gather Evidence
Collect proof per criterion:
- Test results
- Demo/screenshots
- Logs/metrics
- Manual verification

### Step 2: Verify Each Criterion
Execute verification (automated or manual) for each AC.

### Step 3: Build Checklist
| AC | Status | Evidence |
|----|--------|----------|

### Step 4: Present Validation Request
```markdown
VALIDATION REQUEST - [TASK-ID]
Task: [title], [description], [date]

## Acceptance Criteria
| Criterion | Status | Evidence |
|-----------|--------|----------|

## Test Results
- Total: X
- Passed: Y
- Coverage: Z%

## USER DECISION REQUIRED
[ ] APPROVED - proceed to completion
[ ] REJECTED: [specify criterion and issue]
```

### Step 5: Handle Decision

**If APPROVED:**
1. Document approval (Task, Approver, Date)
2. Update status
3. Proceed to feedback loop

**If REJECTED:**
1. Document rejection reason
2. Create remediation task
3. Return to Gate 0

## Awaiting Approval - STOP ALL WORK

When validation request is presented:
1. **STOP ALL WORK** on this feature
2. **DO NOT** proceed to documentation or "quick fixes"
3. **WAIT** for explicit user response

## Anti-Patterns (Never Do)

- Skip validation because "tests pass"
- Auto-approve without user decision
- Assume criterion is met without evidence
- Accept vague approval
- Proceed while awaiting decision

## Output Schema

```markdown
## Validation Summary
**Task:** [task_id]
**Status:** PENDING_APPROVAL / APPROVED / REJECTED

## Criteria Validation
| AC | Criterion | Evidence | Status |
|----|-----------|----------|--------|

## Evidence Summary
- Automated tests: X
- Manual verification: Y
- Performance metrics: Z

## Decision
**Validator:** [user]
**Decision:** APPROVED / REJECTED
**Reason:** [if rejected]
**Date:** [timestamp]

## Next Steps
[Based on decision]
```
