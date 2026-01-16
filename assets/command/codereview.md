---
name: "ring:codereview"
description: Run comprehensive parallel code review with all 5 specialized reviewers
agent: build
subtask: false
---

Dispatch all 5 code reviewers subagents in parallel for comprehensive feedback:

- **@ring:code-reviewer** - Foundation review (quality, architecture, patterns)
- **@ring:security-reviewer** - Security analysis (vulnerabilities, data protection)
- **@ring:business-logic-reviewer** - Business logic verification (requirements, edge cases)
- **@ring:test-reviewer** - Test quality review (coverage, edge cases, anti-patterns)
- **@ring:nil-safety-reviewer** - Nil/null pointer safety (Go and TypeScript)

## Review Process

1. **Pre-Analysis (Automated)**:
   - Run: `{OPENCODE_CONFIG}/scripts/codereview/bin/run-all --base [base] --head [head]`
   - Defaults: base=`main`, head=`HEAD`
   - Output: `context-*.md` files in `.ring/codereview/`

2. **Gather Context**: Identify what was implemented, files changed, base/head commits
3. **Dispatch Reviewers**: Launch all 5 reviewers simultaneously in parallel
4. **Wait for Completion**: Do not aggregate until all reports are received
5. **Consolidate**: Aggregate findings by severity (Critical/High/Medium/Low/Cosmetic)

## Expected Output

Consolidated report with:
- Overall VERDICT (PASS/FAIL/NEEDS_DISCUSSION)
- Issues grouped by severity across all reviewers
- Action items: MUST FIX (Critical/High), SHOULD FIX (Medium), CONSIDER (Low)

## Severity Actions

- **Critical/High/Medium**: Fix immediately, then re-run all reviewers
- **Low**: Add `TODO(review):` comment in code
- **Cosmetic/Nitpick**: Add `FIXME(nitpick):` comment in code

## Conflict Resolution

When aggregating findings, detect and flag conflicting recommendations between reviewers:

| Conflict Type | Resolution | Priority |
|--------------|------------|----------|
| Security vs Performance | Security recommendation wins | CRITICAL |
| More tests vs Over-testing | Defer to ring:test-reviewer for test scope | MEDIUM |
| More mocks vs Less mocks | Evaluate based on ring:test-reviewer guidance | MEDIUM |
| Refactor vs Keep simple | Defer to ring:code-reviewer for architecture decisions | MEDIUM |

**Flagging Conflicts:**
When reviewers provide contradictory guidance:
1. Include BOTH recommendations in consolidated report
2. Add a "Conflict" marker
3. Present to user for final decision
4. Do NOT automatically resolve conflicting recommendations

## Failure Handling

If any reviewer fails:
1. Retry the failed reviewer once
2. If retry fails, report which reviewer failed and why
3. Continue with available results only if user explicitly approves

## Incomplete Output Detection

Signs that a reviewer produced incomplete output:

| Pattern | Detection Method | Action |
|---------|-----------------|--------|
| Missing VERDICT | Output lacks "## VERDICT:" or "**Verdict:**" | Re-dispatch reviewer |
| Empty Issues section | "## Issues Found" followed by no content or "None" only | Verify this is intentional (PASS case) |
| Missing required sections | Check against output_schema in agent definition | Re-dispatch with explicit section reminder |
| Truncated output | Ends mid-sentence or lacks closing sections | Re-dispatch with smaller scope |
| Generic responses | Only contains boilerplate without file-specific analysis | Re-dispatch with explicit file list |

**Validation Regex Patterns:**
- Verdict present: `/^##?\s*VERDICT:?\s*(PASS|FAIL|NEEDS_DISCUSSION)/im`
- Issues section: `/^##?\s*Issues Found/im`
- Summary present: `/^##?\s*(Summary|Executive Summary)/im`

$ARGUMENTS

Wait for all 5 reviewers to complete, then aggregate findings by severity.
