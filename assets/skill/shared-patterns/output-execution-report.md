# Execution Report Format

This file defines the standard execution report format for all development cycle gates.

## Standard Report Structure

Every gate skill MUST produce an execution report with these base metrics:

| Metric | Value | Description |
|--------|-------|-------------|
| Duration | Xm Ys | Time spent executing the gate |
| Iterations | N | Number of retry attempts |
| Result | PASS/FAIL/PARTIAL | Gate outcome |

## How to Use This File

Skills should reference this format and add gate-specific metrics:

```markdown
## Execution Report

Base metrics per [shared-patterns/output-execution-report.md](output-execution-report.md):

| Metric | Value |
|--------|-------|
| Duration | Xm Ys |
| Iterations | N |
| Result | PASS/FAIL/PARTIAL |

### Gate-Specific Details
[Add gate-specific metrics here]
```

## Gate-Specific Metric Examples

### Implementation Gate (0)
- tasks_completed: N/N
- files_created: N
- files_modified: N
- tests_added: N
- agent_used: {agent}

### DevOps Gate (1)
- dockerfile_action: CREATED/UPDATED/UNCHANGED
- compose_action: CREATED/UPDATED/UNCHANGED
- services_configured: N
- verification_passed: YES/no

### SRE Gate (2)
- logging_structured: YES/no
- tracing_enabled: YES/no/N/A

### Testing Gate (3)
- tests_written: N
- coverage: XX.X%
- criteria_covered: X/Y
- verdict: PASS/FAIL

### Review Gate (4)
- reviewers: 3 (parallel)
- findings: X Critical, Y High, Z Medium, W Low
- fixes_applied: N
- verdict: PASS/FAIL/NEEDS_DISCUSSION

### Validation Gate (5)
- criteria_validated: X/Y
- evidence_collected: X automated, Y manual
- user_decision: APPROVED/REJECTED
