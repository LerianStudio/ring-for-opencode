---
description: Run comprehensive parallel code review with all 3 specialized reviewers
agent: build
subtask: true
---

Dispatch all 3 code reviewers in parallel for comprehensive feedback:

- **@code-reviewer** - Foundation review (quality, architecture, patterns)
- **@security-reviewer** - Security analysis (vulnerabilities, data protection)
- **@business-logic-reviewer** - Business logic verification (requirements, edge cases)

## Review Process

1. **Gather Context**: Identify what was implemented, files changed, base/head commits
2. **Dispatch Reviewers**: Launch all 3 reviewers simultaneously in parallel
3. **Wait for Completion**: Do not aggregate until all reports are received
4. **Consolidate**: Aggregate findings by severity (Critical/High/Medium/Low/Cosmetic)

## Expected Output

Consolidated report with:
- Overall VERDICT (PASS/FAIL/NEEDS_DISCUSSION)
- Issues grouped by severity across all reviewers
- Action items: MUST FIX (Critical/High), SHOULD FIX (Medium), CONSIDER (Low)

## Severity Actions

- **Critical/High/Medium**: Fix immediately, then re-run all reviewers
- **Low**: Add `TODO(review):` comment in code
- **Cosmetic/Nitpick**: Add `FIXME(nitpick):` comment in code

## Failure Handling

If any reviewer fails:
1. Retry the failed reviewer once
2. If retry fails, report which reviewer failed and why
3. Continue with available results only if user explicitly approves

$ARGUMENTS

Wait for all reviewers to complete, then aggregate findings by severity.
