---
name: "ring:dev-feedback-loop"
description: |
  Captures metrics and learnings from completed development cycles.
  Enables continuous improvement by analyzing patterns across cycles.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: workflow
  trigger:
    - After ring:dev-cycle completes (success or failure)
    - After all gates have been processed
---

# Dev Feedback Loop

## Overview

Captures metrics and learnings from completed development cycles to enable continuous improvement.

## Metrics Captured

### Per-Gate Metrics

| Gate | Metrics |
|------|---------|
| Gate 0 (Implementation) | Duration, iterations, TDD compliance, standards gaps |
| Gate 1 (DevOps) | Duration, artifacts created, verification errors |
| Gate 2 (SRE) | Duration, instrumentation coverage %, validation errors |
| Gate 3 (Testing) | Duration, coverage %, iterations, test failures |
| Gate 4 (Review) | Duration, iterations, issues per reviewer |
| Gate 5 (Validation) | Duration, approval/rejection, rejection reasons |

### Aggregate Metrics

```json
{
  "total_duration_ms": 0,
  "total_iterations": 0,
  "gates_passed_first_try": 0,
  "gates_required_fixes": 0,
  "most_common_issues": [],
  "standards_compliance_rate": 0.0
}
```

## Learning Categories

### Technical Patterns

- Which standards are most frequently violated
- Which gates require most iterations
- Common test coverage gaps
- Recurring review issues

### Process Patterns

- Average cycle duration by task complexity
- Bottleneck gates
- Approval wait times
- Rejection reasons

## Output Format

```markdown
## Feedback Loop Summary

### Cycle Metrics
| Metric | Value |
|--------|-------|
| Total Duration | Xm Ys |
| Total Iterations | N |
| Gates Passed First Try | X/6 |

### Gate Performance
| Gate | Duration | Iterations | Issues |
|------|----------|------------|--------|
| Implementation | Xm | N | [count] |
| DevOps | Xm | N | [count] |
| SRE | Xm | N | [count] |
| Testing | Xm | N | [count] |
| Review | Xm | N | [count] |
| Validation | Xm | N | [result] |

### Standards Compliance
| Agent | Sections | Compliant | Gaps |
|-------|----------|-----------|------|
| backend-engineer-* | X | Y | Z |
| ring:devops-engineer | X | Y | Z |
| ring:sre | X | Y | Z |
| ring:qa-analyst | X | Y | Z |

### Top Issues
1. [Most common issue type]
2. [Second most common]
3. [Third most common]

### Recommendations
- [Suggestion based on patterns]
- [Process improvement]
```

## State Integration

Reads from ring:dev-cycle state file:
- `docs/ring:dev-cycle/current-cycle.json`
- `docs/ring:dev-refactor/current-cycle.json`

Extracts:
- `agent_outputs.*` for per-gate data
- `metrics.*` for timing data
- `tasks[*].gate_progress` for iteration counts

## Usage

Automatically invoked at end of ring:dev-cycle. Can also be manually triggered to analyze historical cycles.
