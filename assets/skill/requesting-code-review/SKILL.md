---
name: "ring:requesting-code-review"
description: |
  Gate 4 of development cycle - dispatches 5 specialized reviewers (code, business-logic,
  security, test, nil-safety) in parallel for comprehensive code review feedback.
license: MIT
compatibility: opencode
metadata:
  trigger: "Gate 4 of development cycle, after completing major feature, before merge to main, after fixing complex bug"
  skip_when: "Never - code review is always required for blocking issues"
  sequence_after: "ring:dev-testing"
  sequence_before: "ring:dev-validation"
  source: ring-default
---

# Code Review (Gate 4)

## Related Skills

**Complementary:** ring:dev-cycle, ring:dev-implementation, ring:dev-testing

## Input Schema

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| unit_id | string | optional | Task or subtask identifier (auto-generated if not provided) |
| base_sha | string | optional | Git SHA before implementation (auto-detected via git merge-base HEAD main) |
| head_sha | string | optional | Git SHA after implementation (auto-detected via git rev-parse HEAD) |
| implementation_summary | string | optional | Summary of what was implemented (auto-generated from git log if not provided) |
| requirements | string | optional | Requirements or acceptance criteria (reviewers will infer from code if not provided) |
| implementation_files | array | optional | List of files changed (auto-detected via git diff if not provided) |
| gate0_handoff | object | optional | Full handoff from Gate 0 (only when called from dev-cycle) |
| skip_reviewers | array | optional | Reviewers to skip: code-reviewer, business-logic-reviewer, security-reviewer, test-reviewer, nil-safety-reviewer (use sparingly) |

## Output Schema

**Format:** markdown

**Required sections:**
- `## Review Summary`
- `## Issues by Severity`
- `## Reviewer Verdicts`
- `## Handoff to Next Gate`

**Metrics:**
- result: PASS | FAIL | NEEDS_FIXES
- reviewers_passed: X/5 format
- issues_critical, issues_high, issues_medium, issues_low: integer counts
- iterations: integer

---

## Overview

Dispatch all five reviewer subagents in **parallel** for fast, comprehensive feedback:

1. **code-reviewer** - Architecture, design patterns, code quality
2. **business-logic-reviewer** - Domain correctness, business rules, edge cases
3. **security-reviewer** - Vulnerabilities, authentication, OWASP risks
4. **test-reviewer** - Test quality, coverage, edge cases, anti-patterns
5. **nil-safety-reviewer** - Nil/null pointer safety for Go and TypeScript

**Core principle:** All 5 reviewers run simultaneously in a single message with 5 Task tool calls.

## CRITICAL: Role Clarification

**This skill ORCHESTRATES. Reviewer Agents REVIEW.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Dispatch reviewers, aggregate findings, track iterations |
| **Reviewer Agents** | Analyze code, report issues with severity |
| **Implementation Agent** | Fix issues found by reviewers |

---

## Step 1: Gather Context (Auto-Detect if Not Provided)

```text
This skill supports TWO modes:
1. WITH INPUTS: Called by any skill/user that provides structured inputs (unit_id, base_sha, etc.)
2. STANDALONE: Called directly without inputs - auto-detects everything from git

FOR EACH INPUT, check if provided OR auto-detect:

1. unit_id:
   IF provided → use it
   ELSE → generate: "review-" + timestamp (e.g., "review-20241222-143052")

2. base_sha:
   IF provided → use it
   ELSE → Execute: git merge-base HEAD main
   IF git fails → Execute: git rev-parse HEAD~10 (fallback to last 10 commits)

3. head_sha:
   IF provided → use it
   ELSE → Execute: git rev-parse HEAD

4. implementation_files:
   IF provided → use it
   ELSE → Execute: git diff --name-only [base_sha] [head_sha]

5. implementation_summary:
   IF provided → use it
   ELSE → Execute: git log --oneline [base_sha]..[head_sha]
   Format as: "Changes: [list of commit messages]"

6. requirements:
   IF provided → use it
   ELSE → Set to: "Infer requirements from code changes and commit messages"
   (Reviewers will analyze code to understand intent)

AFTER AUTO-DETECTION, display context:
┌─────────────────────────────────────────────────────────────────┐
│ CODE REVIEW CONTEXT                                             │
├─────────────────────────────────────────────────────────────────┤
│ Unit ID: [unit_id]                                              │
│ Base SHA: [base_sha]                                            │
│ Head SHA: [head_sha]                                            │
│ Files Changed: [count] files                                    │
│ Commits: [count] commits                                        │
│                                                                 │
│ Dispatching 5 reviewers in parallel...                          │
└─────────────────────────────────────────────────────────────────┘
```

## Step 2: Initialize Review State

```text
review_state = {
  unit_id: [from input],
  base_sha: [from input],
  head_sha: [from input],
  reviewers: {
    code_reviewer: {verdict: null, issues: []},
    business_logic_reviewer: {verdict: null, issues: []},
    security_reviewer: {verdict: null, issues: []},
    test_reviewer: {verdict: null, issues: []},
    nil_safety_reviewer: {verdict: null, issues: []}
  },
  aggregated_issues: {
    critical: [],
    high: [],
    medium: [],
    low: [],
    cosmetic: []
  },
  iterations: 0,
  max_iterations: 3
}
```

## Step 3: Dispatch All 5 Reviewers in Parallel

**CRITICAL: All 5 reviewers MUST be dispatched in a SINGLE message with 5 Task calls.**

```yaml
# Task 1: Code Reviewer
Task:
  subagent_type: "code-reviewer"
  description: "Code review for [unit_id]"
  prompt: |
    ## Code Review Request

    **Unit ID:** [unit_id]
    **Base SHA:** [base_sha]
    **Head SHA:** [head_sha]

    ## What Was Implemented
    [implementation_summary]

    ## Requirements
    [requirements]

    ## Files Changed
    [implementation_files or "Use git diff"]

    ## Your Focus
    - Architecture and design patterns
    - Code quality and maintainability
    - Naming conventions
    - Error handling patterns
    - Performance concerns

    ## Required Output
    ### VERDICT: PASS / FAIL

    ### Issues Found
    | Severity | Description | File:Line | Recommendation |
    |----------|-------------|-----------|----------------|
    | [CRITICAL/HIGH/MEDIUM/LOW/COSMETIC] | [issue] | [location] | [fix] |

    ### What Was Done Well
    [positive observations]

# Task 2: Business Logic Reviewer
Task:
  subagent_type: "business-logic-reviewer"
  description: "Business logic review for [unit_id]"
  prompt: |
    ## Business Logic Review Request

    **Unit ID:** [unit_id]
    **Base SHA:** [base_sha]
    **Head SHA:** [head_sha]

    ## What Was Implemented
    [implementation_summary]

    ## Requirements
    [requirements]

    ## Your Focus
    - Domain correctness
    - Business rules implementation
    - Edge cases handling
    - Requirements coverage
    - Data validation

    ## Required Output
    ### VERDICT: PASS / FAIL

    ### Issues Found
    | Severity | Description | File:Line | Recommendation |
    |----------|-------------|-----------|----------------|
    | [CRITICAL/HIGH/MEDIUM/LOW/COSMETIC] | [issue] | [location] | [fix] |

    ### Requirements Traceability
    | Requirement | Status | Evidence |
    |-------------|--------|----------|
    | [req] | ✅/❌ | [file:line] |

# Task 3: Security Reviewer
Task:
  subagent_type: "security-reviewer"
  description: "Security review for [unit_id]"
  prompt: |
    ## Security Review Request

    **Unit ID:** [unit_id]
    **Base SHA:** [base_sha]
    **Head SHA:** [head_sha]

    ## What Was Implemented
    [implementation_summary]

    ## Requirements
    [requirements]

    ## Your Focus
    - Authentication and authorization
    - Input validation
    - SQL injection, XSS, CSRF
    - Sensitive data handling
    - OWASP Top 10 risks

    ## Required Output
    ### VERDICT: PASS / FAIL

    ### Issues Found
    | Severity | Description | File:Line | OWASP Category | Recommendation |
    |----------|-------------|-----------|----------------|----------------|
    | [CRITICAL/HIGH/MEDIUM/LOW] | [issue] | [location] | [A01-A10] | [fix] |

    ### Security Checklist
    | Check | Status |
    |-------|--------|
    | Input validation | ✅/❌ |
    | Auth checks | ✅/❌ |
    | No hardcoded secrets | ✅/❌ |

# Task 4: Test Reviewer
Task:
  subagent_type: "test-reviewer"
  description: "Test quality review for [unit_id]"
  prompt: |
    ## Test Quality Review Request

    **Unit ID:** [unit_id]
    **Base SHA:** [base_sha]
    **Head SHA:** [head_sha]

    ## What Was Implemented
    [implementation_summary]

    ## Requirements
    [requirements]

    ## Your Focus
    - Test coverage for business logic
    - Edge case testing (empty, null, boundary)
    - Error path coverage
    - Test independence and isolation
    - Assertion quality (not just "no error")
    - Test anti-patterns (testing mock behavior)

    ## Required Output
    ### VERDICT: PASS / FAIL

    ### Issues Found
    | Severity | Description | File:Line | Recommendation |
    |----------|-------------|-----------|----------------|
    | [CRITICAL/HIGH/MEDIUM/LOW] | [issue] | [location] | [fix] |

    ### Test Coverage Analysis
    | Test Type | Count | Coverage |
    |-----------|-------|----------|
    | Unit | [N] | [areas] |
    | Integration | [N] | [areas] |
    | E2E | [N] | [areas] |

# Task 5: Nil-Safety Reviewer
Task:
  subagent_type: "nil-safety-reviewer"
  description: "Nil/null safety review for [unit_id]"
  prompt: |
    ## Nil-Safety Review Request

    **Unit ID:** [unit_id]
    **Base SHA:** [base_sha]
    **Head SHA:** [head_sha]
    **Languages:** [Go|TypeScript|both - detect from files]

    ## What Was Implemented
    [implementation_summary]

    ## Requirements
    [requirements]

    ## Your Focus
    - Nil/null pointer risks in changed code
    - Missing nil guards before dereference
    - Map access without ok check (Go)
    - Type assertions without ok check (Go)
    - Optional chaining misuse (TypeScript)
    - Error-then-use patterns

    ## Required Output
    ### VERDICT: PASS / FAIL

    ### Issues Found
    | Severity | Description | File:Line | Recommendation |
    |----------|-------------|-----------|----------------|
    | [CRITICAL/HIGH/MEDIUM/LOW] | [issue] | [location] | [fix] |

    ### Nil Risk Trace
    [For each risk: Source → Propagation → Dereference point]
```

## Step 4: Wait for All Reviewers and Parse Output

```text
Wait for all 5 Task calls to complete.

For each reviewer:
1. Extract VERDICT (PASS/FAIL)
2. Extract Issues Found table
3. Categorize issues by severity

review_state.reviewers.code_reviewer = {
  verdict: [PASS/FAIL],
  issues: [parsed issues]
}
// ... same for other reviewers

Aggregate all issues by severity:
review_state.aggregated_issues.critical = [all critical from all reviewers]
review_state.aggregated_issues.high = [all high from all reviewers]
// ... etc
```

## Step 5: Handle Results by Severity

```text
Count blocking issues:
blocking_count = critical.length + high.length + medium.length

IF blocking_count == 0:
  → All reviewers PASS
  → Proceed to Step 8 (Success)

IF blocking_count > 0:
  → review_state.iterations += 1
  → IF iterations >= max_iterations: Go to Step 9 (Escalate)
  → Go to Step 6 (Dispatch Fixes)
```

## Step 6: Dispatch Fixes to Implementation Agent

```yaml
Task:
  subagent_type: "[implementation_agent from Gate 0]"
  description: "Fix review issues for [unit_id]"
  prompt: |
    FIX REQUIRED - Code Review Issues Found

    ## Context
    - **Unit ID:** [unit_id]
    - **Iteration:** [iterations] of [max_iterations]

    ## Critical Issues (MUST FIX)
    [list critical issues with file:line and recommendation]

    ## High Issues (MUST FIX)
    [list high issues]

    ## Medium Issues (MUST FIX)
    [list medium issues]

    ## Requirements
    1. Fix ALL Critical, High, and Medium issues
    2. Commit fixes
    3. Return list of fixed issues with evidence

    ## For Low/Cosmetic Issues
    Add TODO/FIXME comments:
    - Low: `// TODO(review): [Issue] - [reviewer] on [date]`
    - Cosmetic: `// FIXME(nitpick): [Issue] - [reviewer] on [date]`
```

## Step 7: Re-Run All Reviewers After Fixes

```text
After fixes committed:
1. Get new HEAD_SHA
2. Go back to Step 3 (dispatch all 5 reviewers again)

CRITICAL: Always re-run ALL 5 reviewers after fixes.
Do NOT cherry-pick reviewers.
```

## Step 8: Prepare Success Output

```text
Generate skill output:

## Review Summary
**Status:** PASS
**Unit ID:** [unit_id]
**Iterations:** [review_state.iterations]

## Issues by Severity
| Severity | Count |
|----------|-------|
| Critical | 0 |
| High | 0 |
| Medium | 0 |
| Low | [count] |
| Cosmetic | [count] |

## Reviewer Verdicts
| Reviewer | Verdict | Issues |
|----------|---------|--------|
| code-reviewer | ✅ PASS | [count] |
| business-logic-reviewer | ✅ PASS | [count] |
| security-reviewer | ✅ PASS | [count] |
| test-reviewer | ✅ PASS | [count] |
| nil-safety-reviewer | ✅ PASS | [count] |

## Low/Cosmetic Issues (TODO/FIXME added)
[list with file locations]

## Handoff to Next Gate
- Review status: COMPLETE
- All blocking issues: RESOLVED
- Reviewers passed: 5/5
- Ready for Gate 5 (Validation): YES
```

## Step 9: Escalate - Max Iterations Reached

```text
Generate skill output:

## Review Summary
**Status:** FAIL
**Unit ID:** [unit_id]
**Iterations:** [max_iterations] (MAX REACHED)

## Issues by Severity
| Severity | Count |
|----------|-------|
| Critical | [count] |
| High | [count] |
| Medium | [count] |

## Unresolved Issues
[list all Critical/High/Medium still open]

## Reviewer Verdicts
| Reviewer | Verdict |
|----------|---------|
| code-reviewer | [PASS/FAIL] |
| business-logic-reviewer | [PASS/FAIL] |
| security-reviewer | [PASS/FAIL] |
| test-reviewer | [PASS/FAIL] |
| nil-safety-reviewer | [PASS/FAIL] |

## Handoff to Next Gate
- Review status: FAILED
- Unresolved blocking issues: [count]
- Ready for Gate 5: NO
- **Action Required:** User must manually resolve issues

ESCALATION: Max iterations (3) reached. Blocking issues remain.
```

---

## Pressure Resistance

See [shared-patterns/pressure-resistance.md](../shared-patterns/pressure-resistance.md) for universal pressure scenarios.

| User Says | Your Response |
|-----------|---------------|
| "Skip review, code is simple" | "Simple code can have security issues. Dispatching all 5 reviewers." |
| "Just run code-reviewer" | "All 5 reviewers run in parallel. No time saved by skipping." |
| "Fix later, merge now" | "Blocking issues (Critical/High/Medium) MUST be fixed before Gate 5." |

## Anti-Rationalization Table

See [shared-patterns/anti-rationalization.md](../shared-patterns/anti-rationalization.md) for universal anti-rationalizations.

### Gate 4-Specific Anti-Rationalizations

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "Run reviewers one at a time" | Sequential = slow. Parallel = 5x faster. | **Dispatch all 5 in single message** |
| "Skip security for internal code" | Internal code can have vulnerabilities. | **Include security-reviewer** |
| "Critical issue is false positive" | Prove it with evidence, don't assume. | **Fix or provide evidence** |
| "Low issues don't need TODO" | TODOs ensure issues aren't forgotten. | **Add TODO comments** |
| "4 of 5 reviewers passed" | Gate 4 requires ALL 5. 4/5 = 0/5. | **Re-run ALL 5 reviewers** |
| "MEDIUM is not blocking" | MEDIUM = MUST FIX. Same as CRITICAL/HIGH. | **Fix MEDIUM issues NOW** |

---

## Execution Report Format

```markdown
## Review Summary
**Status:** [PASS|FAIL|NEEDS_FIXES]
**Unit ID:** [unit_id]
**Duration:** [Xm Ys]
**Iterations:** [N]

## Issues by Severity
| Severity | Count |
|----------|-------|
| Critical | [N] |
| High | [N] |
| Medium | [N] |
| Low | [N] |

## Reviewer Verdicts
| Reviewer | Verdict |
|----------|---------|
| code-reviewer | ✅/❌ |
| business-logic-reviewer | ✅/❌ |
| security-reviewer | ✅/❌ |
| test-reviewer | ✅/❌ |
| nil-safety-reviewer | ✅/❌ |

## Handoff to Next Gate
- Review status: [COMPLETE|FAILED]
- Blocking issues: [resolved|N remaining]
- Ready for Gate 5: [YES|NO]
```
