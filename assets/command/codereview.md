---
name: "ring:codereview"
description: Run comprehensive parallel code review with all 5 specialized reviewers
agent: build
subtask: false
---

Dispatch all 5 specialized code reviewers in parallel, collect their reports, and provide a consolidated analysis.

## Review Process

### Step 0: Run Pre-Analysis Pipeline (MANDATORY)

**MANDATORY:** Before dispatching reviewers, run the pre-analysis pipeline to generate static analysis context.

```bash
# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

# Find binary (plugin path or dev repo)
BINARY="${CLAUDE_PLUGIN_ROOT}/lib/codereview/bin/${OS}_${ARCH}/run-all"
if [ ! -x "$BINARY" ]; then
  BINARY="./default/lib/codereview/bin/${OS}_${ARCH}/run-all"
fi

# Run help to understand best way to feed files to the pre-analysis app
$BINARY --help
```

**Scope:** Review what the user desires (unstaged files, HEAD x other branch, etc) and chooses the right params. If in doubt, ask the user.

**Output:** Creates 5 context files in `docs/codereview/`:
- `context-code-reviewer.md`
- `context-security-reviewer.md`
- `context-business-logic-reviewer.md`
- `context-test-reviewer.md`
- `context-nil-safety-reviewer.md`

⚠️ **DEGRADED MODE:** If binary not found, display warning and continue. Reviewers will work without pre-analysis context.

---

### Step 1: Dispatch All Five Reviewers in Parallel

**CRITICAL: Use a single message with 5 Task tool calls to launch all reviewers simultaneously.**

Gather the required context first:
- WHAT_WAS_IMPLEMENTED: Summary of changes made
- PLAN_OR_REQUIREMENTS: Original plan or requirements (if available)
- BASE_SHA: Base commit for comparison (if applicable)
- HEAD_SHA: Head commit for comparison (if applicable)
- DESCRIPTION: Additional context about the changes
- LANGUAGES: Go, TypeScript, or both (for ring:nil-safety-reviewer)

Then dispatch all 5 reviewers:

```
Task tool #1 (ring:code-reviewer):
  subagent_type: "ring:code-reviewer"
  model: "opus"
  description: "Review code quality and architecture"
  prompt: |
    WHAT_WAS_IMPLEMENTED: [summary of changes]
    PLAN_OR_REQUIREMENTS: [original plan/requirements]
    BASE_SHA: [base commit if applicable]
    HEAD_SHA: [head commit if applicable]
    DESCRIPTION: [additional context]

Task tool #2 (ring:business-logic-reviewer):
  subagent_type: "ring:business-logic-reviewer"
  model: "opus"
  description: "Review business logic correctness"
  prompt: |
    [Same parameters as above]

Task tool #3 (ring:security-reviewer):
  subagent_type: "ring:security-reviewer"
  model: "opus"
  description: "Review security vulnerabilities"
  prompt: |
    [Same parameters as above]

Task tool #4 (ring:test-reviewer):
  subagent_type: "ring:test-reviewer"
  model: "opus"
  description: "Review test quality and coverage"
  prompt: |
    [Same parameters as above]
    Focus: Edge cases, error paths, test independence, assertion quality.

Task tool #5 (ring:nil-safety-reviewer):
  subagent_type: "ring:nil-safety-reviewer"
  model: "opus"
  description: "Review nil/null pointer safety"
  prompt: |
    [Same parameters as above]
    LANGUAGES: [Go|TypeScript|both]
    Focus: Nil sources, propagation paths, missing guards.
```

**Wait for all five reviewers to complete their work.**

### Step 2: Collect and Aggregate Reports

Each reviewer returns:
- **Verdict:** PASS/FAIL/NEEDS_DISCUSSION
- **Strengths:** What was done well
- **Issues:** Categorized by severity (Critical/High/Medium/Low/Cosmetic)
- **Recommendations:** Specific actionable feedback

Consolidate all issues by severity across all five reviewers.

### Conflict Resolution

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
2. Add a "⚠️ Conflict" marker
3. Present to user for final decision
4. Do NOT automatically resolve conflicting recommendations

**Example:**
```
⚠️ Conflict Detected:
- ring:test-reviewer: "Add more mock isolation for external services"
- ring:code-reviewer: "Current mocking approach is sufficient"
- Resolution: User decision required - see both perspectives above
```

### Step 3: Provide Consolidated Report

Return a consolidated report in this format:

```markdown
# Full Review Report

## VERDICT: [PASS | FAIL | NEEDS_DISCUSSION]

## Executive Summary

[2-3 sentences about overall review across all gates]

**Total Issues:**
- Critical: [N across all gates]
- High: [N across all gates]
- Medium: [N across all gates]
- Low: [N across all gates]

---

## Code Quality Review (Foundation)

**Verdict:** [PASS | FAIL]
**Issues:** Critical [N], High [N], Medium [N], Low [N]

### Critical Issues
[List all critical code quality issues]

### High Issues
[List all high code quality issues]

[Medium/Low issues summary]

---

## Business Logic Review (Correctness)

**Verdict:** [PASS | FAIL]
**Issues:** Critical [N], High [N], Medium [N], Low [N]

### Critical Issues
[List all critical business logic issues]

### High Issues
[List all high business logic issues]

[Medium/Low issues summary]

---

## Security Review (Safety)

**Verdict:** [PASS | FAIL]
**Issues:** Critical [N], High [N], Medium [N], Low [N]

### Critical Vulnerabilities
[List all critical security vulnerabilities]

### High Vulnerabilities
[List all high security vulnerabilities]

[Medium/Low vulnerabilities summary]

---

## Test Quality Review (Coverage)

**Verdict:** [PASS | FAIL]
**Issues:** Critical [N], High [N], Medium [N], Low [N]

### Critical Issues
[Untested core logic, tests testing mock behavior]

### High Issues
[Missing edge cases, test anti-patterns]

[Medium/Low issues summary]

---

## Nil-Safety Review (Pointer Safety)

**Verdict:** [PASS | FAIL]
**Issues:** Critical [N], High [N], Medium [N], Low [N]

### Critical Issues
[Direct panic paths, unguarded nil dereference]

### High Issues
[Conditional nil risks, missing ok checks]

[Medium/Low issues summary]

---

## Consolidated Action Items

**MUST FIX (Critical):**
1. [Issue from any gate] - `file:line`
2. [Issue from any gate] - `file:line`

**SHOULD FIX (High):**
1. [Issue from any gate] - `file:line`
2. [Issue from any gate] - `file:line`

**CONSIDER (Medium/Low):**
[Brief list]

---

## Next Steps

**If PASS:**
- ✅ All 5 reviewers passed
- ✅ Ready for next step (merge/production)

**If FAIL:**
- ❌ Fix all Critical/High/Medium issues immediately
- ❌ Add TODO(review) comments for Low issues in code
- ❌ Add FIXME(nitpick) comments for Cosmetic/Nitpick issues in code
- ❌ Re-run all 5 reviewers in parallel after fixes

**If NEEDS_DISCUSSION:**
- 💬 [Specific discussion points across gates]
```

## Severity-Based Action Guide

After producing the consolidated report, provide clear guidance:

**Critical/High/Medium Issues:**
```
These issues MUST be fixed immediately:
1. [Issue description] - file.ext:line - [Reviewer]
2. [Issue description] - file.ext:line - [Reviewer]

Recommended approach:
- Dispatch fix subagent to address all Critical/High/Medium issues
- After fixes complete, re-run all 5 reviewers in parallel to verify
```

**Low Issues:**
```
Add TODO comments in the code for these issues:

// TODO(review): [Issue description]
// Reported by: [reviewer-name] on [date]
// Severity: Low
// Location: file.ext:line
```

**Cosmetic/Nitpick Issues:**
```
Add FIXME comments in the code for these issues:

// FIXME(nitpick): [Issue description]
// Reported by: [reviewer-name] on [date]
// Severity: Cosmetic
// Location: file.ext:line
```

## Reviewer Failure Handling

If any reviewer fails during execution (timeout, error, incomplete output):

### Single Reviewer Failure

1. **Do NOT aggregate partial results** - Wait for all 5 reviewers
2. **Retry the failed reviewer once:**
   ```
   Task tool (retry failed reviewer):
     model: "opus"
     description: "Retry [reviewer-name] review"
     prompt: [same parameters as original]
   ```
3. **If retry fails:** Report which reviewer failed and why
4. **Continue with available results** only if user explicitly approves

### Multiple Reviewer Failures

1. **Stop and report** - Do not provide partial review
2. **Investigate root cause:**
   - Large codebase? Consider chunking files
   - Timeout? Increase timeout or reduce scope
   - Error? Check file paths and permissions
3. **Retry all failed reviewers** after addressing root cause

### Incomplete Output Detection

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

**Action:** Re-dispatch the reviewer with explicit instruction to include all required sections.

## Remember

1. **All reviewers are independent** - They run in parallel, not sequentially
2. **Dispatch all 5 reviewers in parallel** - Single message, 5 Task calls
3. **Specify model: "opus"** - All reviewers need opus for comprehensive analysis
4. **Wait for all to complete** - Don't aggregate until all reports received
5. **Consolidate findings by severity** - Group all issues across reviewers
6. **Provide clear action guidance** - Tell user exactly what to fix vs. document
7. **Overall FAIL if any reviewer fails** - One failure means work needs fixes
8. **Retry failed reviewers once** - Don't give up on first failure

---

## MANDATORY: Load Full Skill

**This command MUST load the skill for complete workflow execution.**

```
Use Skill tool: ring:requesting-code-review
```

The skill contains the complete workflow with:
- Auto-detection of git context (base_sha, head_sha, files)
- Parallel dispatch of all 5 reviewers in single message
- Issue aggregation by severity
- Iteration loop with fix dispatching
- Escalation handling at max iterations
- Anti-rationalization tables
- Pressure resistance scenarios
