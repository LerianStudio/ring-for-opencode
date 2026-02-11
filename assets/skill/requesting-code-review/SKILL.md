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
---

# Code Review (Gate 4)

## Overview

Dispatch all five reviewer subagents in **parallel** for fast, comprehensive feedback:

1. **ring:code-reviewer** - Architecture, design patterns, code quality
2. **ring:business-logic-reviewer** - Domain correctness, business rules, edge cases
3. **ring:security-reviewer** - Vulnerabilities, authentication, OWASP risks
4. **ring:test-reviewer** - Test quality, coverage, edge cases, anti-patterns
5. **ring:nil-safety-reviewer** - Nil/null pointer safety for Go and TypeScript

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
│ 📋 CODE REVIEW CONTEXT                                          │
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

## Step 2.5: Run Pre-Analysis Pipeline (MANDATORY)

**MANDATORY:** Run static analysis, AST extraction, and call graph analysis BEFORE dispatching reviewers. This provides critical context that significantly improves review quality.

**Skip Override:** The `skip_preanalysis` parameter allows bypassing this step ONLY when explicitly requested by the user. This is NOT recommended.

### Step 2.5.1: Detect Platform and Find Binary

```bash
# Detect platform
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
esac

# Binary search paths (in priority order)
PLUGIN_BIN="${CLAUDE_PLUGIN_ROOT:-}/lib/codereview/bin/${OS}_${ARCH}/run-all"
LOCAL_BIN="./default/lib/codereview/bin/${OS}_${ARCH}/run-all"

# Find binary
BINARY=""
CHECKSUM_FILE=""
if [[ -x "$PLUGIN_BIN" ]]; then
    BINARY="$PLUGIN_BIN"
    CHECKSUM_FILE="${CLAUDE_PLUGIN_ROOT:-}/lib/codereview/bin/${OS}_${ARCH}/CHECKSUMS.sha256"
elif [[ -x "$LOCAL_BIN" ]]; then
    BINARY="$LOCAL_BIN"
    CHECKSUM_FILE="./default/lib/codereview/bin/${OS}_${ARCH}/CHECKSUMS.sha256"
fi
```

### Step 2.5.2: Secure Binary Execution (Security)

```bash
# Secure execution: copy to temp, verify copy, execute copy
# This prevents TOCTOU race conditions by verifying the COPY we execute
secure_execute_binary() {
    local binary="$1"
    local checksum_file="$2"
    shift 2
    local args=("$@")

    # Create secure temporary copy
    local secure_copy=$(mktemp)
    trap "rm -f '$secure_copy'" EXIT

    # Copy binary to secure location
    if ! cp "$binary" "$secure_copy"; then
        echo "✗ Failed to create secure copy"
        return 1
    fi
    chmod 700 "$secure_copy"

    # Verify the COPY (not the original - prevents TOCTOU)
    local binary_name=$(basename "$binary")

    # Check checksum file exists
    if [[ ! -f "$checksum_file" ]]; then
        echo "✗ ERROR: Checksum file required for security verification"
        echo "  Set RING_ALLOW_UNVERIFIED=true to bypass (not recommended)"
        if [[ "${RING_ALLOW_UNVERIFIED:-false}" != "true" ]]; then
            return 1
        fi
        echo "⚠️ WARNING: Running in unverified mode"
    else
        # Get expected hash using exact match (prevents partial match attacks)
        local expected_hash=$(awk -v name="$binary_name" '$2 == name {print $1}' "$checksum_file")

        if [[ -z "$expected_hash" ]]; then
            echo "✗ ERROR: Binary '$binary_name' not found in checksum file"
            return 1
        fi

        # Compute hash of the COPY (macOS and Linux compatible)
        local actual_hash
        if command -v sha256sum &> /dev/null; then
            actual_hash=$(sha256sum "$secure_copy" | awk '{print $1}')
        elif command -v shasum &> /dev/null; then
            actual_hash=$(shasum -a 256 "$secure_copy" | awk '{print $1}')
        else
            echo "✗ ERROR: No sha256sum or shasum available"
            return 1
        fi

        if [[ "$expected_hash" != "$actual_hash" ]]; then
            echo "✗ CHECKSUM MISMATCH - Binary may be corrupted or tampered"
            echo "  Expected: $expected_hash"
            echo "  Actual:   $actual_hash"
            return 1
        fi

        echo "✓ Binary integrity verified"
    fi

    # Execute the verified copy
    "$secure_copy" "${args[@]}"
    local result=$?

    rm -f "$secure_copy"
    trap - EXIT

    return $result
}
```

### Step 2.5.3: Fallback to Build from Source

```bash
build_from_source() {
    echo "Attempting to build from source..."

    # Check if Go is available
    if ! command -v go &> /dev/null; then
        echo "✗ Go not installed. Cannot build from source."
        echo "  Install Go from https://go.dev/dl/ or use pre-built binaries."
        return 1
    fi

    # Find source directory
    local source_dir=""
    if [[ -d "./scripts/codereview" ]]; then
        source_dir="./scripts/codereview"
    elif [[ -d "${CLAUDE_PLUGIN_ROOT:-}/../../scripts/codereview" ]]; then
        source_dir="${CLAUDE_PLUGIN_ROOT:-}/../../scripts/codereview"
    fi

    if [[ -z "$source_dir" || ! -d "$source_dir" ]]; then
        echo "✗ Source directory not found. Cannot build from source."
        return 1
    fi

    # Build the binary
    local output_binary="/tmp/ring-codereview-run-all"
    echo "Building run-all from $source_dir..."

    if (cd "$source_dir" && go build -o "$output_binary" ./cmd/run-all/); then
        echo "✓ Built successfully: $output_binary"
        BINARY="$output_binary"
        return 0
    else
        echo "✗ Build failed"
        return 1
    fi
}
```

### Step 2.5.4: Execute with Verification

```bash
# Main execution flow using secure_execute_binary
# This ensures atomic verify-and-execute to prevent TOCTOU attacks
if [[ -n "$BINARY" ]]; then
    if secure_execute_binary "$BINARY" "$CHECKSUM_FILE" \
        --base="$BASE_SHA" --head="$HEAD_SHA" --output="docs/codereview" --verbose; then
        echo "Pre-analysis pipeline completed successfully"
    else
        echo "⚠️ Binary verification or execution failed"
        if build_from_source; then
            # Execute the newly built binary (no checksum for local builds)
            RING_ALLOW_UNVERIFIED=true secure_execute_binary "$BINARY" "" \
                --base="$BASE_SHA" --head="$HEAD_SHA" --output="docs/codereview" --verbose
        else
            echo "⚠️ DEGRADED MODE: Proceeding without pre-analysis"
            echo "  Reviewers will work without static analysis context."
            # Skip to Step 3 (dispatch reviewers)
        fi
    fi
else
    # No binary found - try building from source
    echo "No pre-built binary found for ${OS}_${ARCH}"
    if build_from_source; then
        RING_ALLOW_UNVERIFIED=true secure_execute_binary "$BINARY" "" \
            --base="$BASE_SHA" --head="$HEAD_SHA" --output="docs/codereview" --verbose
    else
        echo "⚠️ DEGRADED MODE: Pre-analysis binary not available"
        echo "  Reviewers will proceed WITHOUT static analysis context."
        # Skip to Step 3 (dispatch reviewers)
    fi
fi
```

- Timeout: Use `preanalysis_timeout` input (default 5 minutes)
- On success: Set `preanalysis_state.success = true`
- On failure: Display warning, set `preanalysis_state.success = false`, continue to Step 3

### Step 2.5.5: Read Context Files

If pipeline succeeded, read the 5 context files:

| Reviewer | Context File |
|----------|--------------|
| `ring:code-reviewer` | `docs/codereview/context-code-reviewer.md` |
| `ring:security-reviewer` | `docs/codereview/context-security-reviewer.md` |
| `ring:business-logic-reviewer` | `docs/codereview/context-business-logic-reviewer.md` |
| `ring:test-reviewer` | `docs/codereview/context-test-reviewer.md` |
| `ring:nil-safety-reviewer` | `docs/codereview/context-nil-safety-reviewer.md` |

Store each file's content in `preanalysis_state.context[reviewer_name]`.

If a context file is missing or empty, log warning and continue (reviewer will work without context).

```text
preanalysis_state = {
  enabled: true,
  success: false,
  context: {
    "ring:code-reviewer": null,
    "ring:security-reviewer": null,
    "ring:business-logic-reviewer": null,
    "ring:test-reviewer": null,
    "ring:nil-safety-reviewer": null
  }
}
```

## Step 3: Dispatch All 5 Reviewers in Parallel

**⛔ CRITICAL: All 5 reviewers MUST be dispatched in a SINGLE message with 5 Task calls.**

```yaml
# Task 1: Code Reviewer
Task:
  subagent_type: "ring:code-reviewer"
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

    ## Pre-Analysis Context

    **Static Analysis Results:**
    The following findings were automatically extracted by the pre-analysis pipeline.
    Use these to INFORM your review, not REPLACE your analysis.

    ---

    [IF preanalysis_state.context["ring:code-reviewer"] exists AND is not empty:]
    [INSERT the content of preanalysis_state.context["ring:code-reviewer"]]
    [ELSE:]
    _No pre-analysis context available. Perform standard review based on git diff._

    ---

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
  subagent_type: "ring:business-logic-reviewer"
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

    ## Pre-Analysis Context

    **Static Analysis Results:**
    The following findings were automatically extracted by the pre-analysis pipeline.
    Use these to INFORM your review, not REPLACE your analysis.

    ---

    [IF preanalysis_state.context["ring:business-logic-reviewer"] exists AND is not empty:]
    [INSERT the content of preanalysis_state.context["ring:business-logic-reviewer"]]
    [ELSE:]
    _No pre-analysis context available. Perform standard review based on git diff._

    ---

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
  subagent_type: "ring:security-reviewer"
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

    ## Pre-Analysis Context

    **Static Analysis Results:**
    The following findings were automatically extracted by the pre-analysis pipeline.
    Use these to INFORM your review, not REPLACE your analysis.

    ---

    [IF preanalysis_state.context["ring:security-reviewer"] exists AND is not empty:]
    [INSERT the content of preanalysis_state.context["ring:security-reviewer"]]
    [ELSE:]
    _No pre-analysis context available. Perform standard review based on git diff._

    ---

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
  subagent_type: "ring:test-reviewer"
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

    ## Pre-Analysis Context

    **Static Analysis Results:**
    The following findings were automatically extracted by the pre-analysis pipeline.
    Use these to INFORM your review, not REPLACE your analysis.

    ---

    [IF preanalysis_state.context["ring:test-reviewer"] exists AND is not empty:]
    [INSERT the content of preanalysis_state.context["ring:test-reviewer"]]
    [ELSE:]
    _No pre-analysis context available. Perform standard review based on git diff._

    ---

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
  subagent_type: "ring:nil-safety-reviewer"
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

    ## Pre-Analysis Context

    **Static Analysis Results:**
    The following findings were automatically extracted by the pre-analysis pipeline.
    Use these to INFORM your review, not REPLACE your analysis.

    ---

    [IF preanalysis_state.context["ring:nil-safety-reviewer"] exists AND is not empty:]
    [INSERT the content of preanalysis_state.context["ring:nil-safety-reviewer"]]
    [ELSE:]
    _No pre-analysis context available. Perform standard review based on git diff._

    ---

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

**⛔ CRITICAL: You are an ORCHESTRATOR. You CANNOT edit source files directly.**
**You MUST dispatch the implementation agent to fix ALL review issues.**

### Orchestrator Boundaries (HARD GATE)

**See [dev-team/skills/shared-patterns/standards-boundary-enforcement.md](../../dev-team/skills/shared-patterns/standards-boundary-enforcement.md) for core enforcement rules.**

**Key prohibition:** Edit/Write/Create on source files is FORBIDDEN. Always dispatch agent.

**If you catch yourself about to use Edit/Write/Create on source files → STOP. Dispatch agent.**

### Dispatch Implementation Agent

```yaml
Task:
  subagent_type: "[implementation_agent from Gate 0]"
  description: "Fix review issues for [unit_id]"
  prompt: |
    ⛔ FIX REQUIRED - Code Review Issues Found

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
    2. Run tests to verify fixes
    3. Commit fixes with descriptive message
    4. Return list of fixed issues with evidence

    ## For Low/Cosmetic Issues
    Add TODO/FIXME comments:
    - Low: `// TODO(review): [Issue] - [reviewer] on [date]`
    - Cosmetic: `// FIXME(nitpick): [Issue] - [reviewer] on [date]`
```

### Anti-Rationalization for Direct Editing

**See [shared-patterns/orchestrator-direct-editing-anti-rationalization.md](../shared-patterns/orchestrator-direct-editing-anti-rationalization.md) for complete anti-rationalization table.**

*Applies to: Step 6 (Fix dispatch after Ring reviewers) & Step 7.5.3 (Fix dispatch after CodeRabbit)*

## Step 7: Re-Run All Reviewers After Fixes

```text
After fixes committed:
1. Get new HEAD_SHA
2. Go back to Step 3 (dispatch all 5 reviewers again)

⛔ CRITICAL: Always re-run ALL 5 reviewers after fixes.
Do NOT cherry-pick reviewers.
```

## Step 7.5: CodeRabbit CLI Validation (Per-Subtask/Task)

**⛔ NEW APPROACH: CodeRabbit validates EACH subtask/task as it completes, accumulating findings to a file.**

### CodeRabbit Integration Overview

```text
┌─────────────────────────────────────────────────────────────────┐
│ CODERABBIT PER-UNIT VALIDATION FLOW                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ DURING REVIEW (after each subtask/task Ring reviewers pass):   │
│   1. Run CodeRabbit for that unit's files                      │
│   2. Append findings to .coderabbit-findings.md                │
│   3. Continue to next unit                                     │
│                                                                 │
│ BEFORE COMMIT (Step 8):                                        │
│   1. Display accumulated .coderabbit-findings.md               │
│   2. User decides: fix issues OR acknowledge and proceed       │
│                                                                 │
│ BENEFITS:                                                      │
│   • Catches issues close to when code was written              │
│   • Smaller scope = faster reviews (7-30 min per unit)         │
│   • Issues isolated to specific units, easier to fix           │
│   • Accumulated file provides audit trail                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Rate Limits (Official - per developer per repository per hour)

| Limit Type | Value | Notes |
|------------|-------|-------|
| Files reviewed | 200 files/hour | Per review |
| Reviews | 3 back-to-back, then 4/hour | **7 reviews possible in first hour** |
| Conversations | 25 back-to-back, then 50/hour | For follow-up questions |

**⏱️ TIMING:** Each CodeRabbit review takes **7-30+ minutes** depending on scope.
Run in background and check periodically for completion.

### Common Commands Reference

<a id="coderabbit-install-check"></a>
**CodeRabbit Installation Check:**
```bash
which coderabbit || which cr
```
> Used in Step 7.5.1 and after installation to verify CLI availability.

---

### ⚠️ PREREQUISITES & ENVIRONMENT REQUIREMENTS

**Before attempting Step 7.5, verify your environment supports the required operations:**

| Requirement | Local Dev | CI/CD | Containerized | Remote/SSH |
|-------------|-----------|-------|---------------|------------|
| `curl \| sh` install | ✅ Yes | ⚠️ May require elevated permissions | ❌ Often blocked | ⚠️ Depends on config |
| Browser auth (`coderabbit auth login`) | ✅ Yes | ❌ No browser | ❌ No browser | ❌ No browser |
| Write to `$HOME/.coderabbit/` | ✅ Yes | ⚠️ Ephemeral | ⚠️ Ephemeral | ✅ Usually |
| Internet access to `cli.coderabbit.ai` | ✅ Yes | ⚠️ Check firewall | ⚠️ Check firewall | ⚠️ Check firewall |

**⛔ HARD STOP CONDITIONS - Skip Step 7.5 if ANY apply:**
- Running in containerized environment without persistent storage
- CI/CD pipeline without pre-installed CodeRabbit CLI
- Non-interactive environment (no TTY for browser auth)
- Network restrictions blocking `cli.coderabbit.ai`
- Read-only filesystem

### Environment-Specific Guidance

#### Local Development (RECOMMENDED)
Standard flow works: `curl | sh` install + browser authentication.

#### CI/CD Pipelines
**Option A: Pre-install in CI image**
```dockerfile
# Add to your CI Dockerfile
RUN curl -fsSL https://cli.coderabbit.ai/install.sh | sh
```

**Option B: Use API token authentication (headless)**
```bash
# Set token via environment variable (add to CI secrets)
export CODERABBIT_API_TOKEN="your-api-token"
coderabbit auth login --token "$CODERABBIT_API_TOKEN"
```

**Option C: Skip CodeRabbit in CI, run locally**
```bash
# In CI config, set env var to auto-skip
export SKIP_CODERABBIT_REVIEW=true
```

#### Containerized/Docker Environments
```bash
# Option 1: Mount credentials from host
docker run -v ~/.coderabbit:/root/.coderabbit ...

# Option 2: Pass token as env var
docker run -e CODERABBIT_API_TOKEN="..." ...

# Option 3: Pre-bake into image (not recommended for tokens)
```

#### Non-Interactive/Headless Authentication
```bash
# Generate API token at: https://app.coderabbit.ai/settings/api-tokens
# Then authenticate without browser:
coderabbit auth login --token "cr_xxxxxxxxxxxxx"
```

---

### Step 7.5 Flow Logic

```text
┌─────────────────────────────────────────────────────────────────┐
│ ✅ ALL 3 RING REVIEWERS PASSED                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Checking CodeRabbit CLI availability...                         │
│                                                                 │
│ CodeRabbit provides additional AI-powered code review that      │
│ catches race conditions, memory leaks, security vulnerabilities,│
│ and edge cases that may complement Ring reviewers.              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**⛔ HARD GATE: CodeRabbit Execution Rules (NON-NEGOTIABLE)**

| Scenario | Rule | Action |
|----------|------|--------|
| **Installed & authenticated** | **MANDATORY** - CANNOT skip | Run CodeRabbit review, no prompt |
| **Not installed** | **MUST ask** user about installation | Present installation option |
| **User declines installation** | Optional - can proceed | Skip and continue to Step 8 |

**Why this distinction:**
- If CodeRabbit IS installed → User has committed to using it → MUST run
- If CodeRabbit is NOT installed → User choice to add it → MUST ask, but can decline

```text
FLOW:
1. Run CodeRabbit Installation Check
2. IF installed AND authenticated → Run CodeRabbit (MANDATORY, NO prompt, CANNOT skip)
3. IF installed BUT NOT authenticated → Guide authentication (REQUIRED before proceeding)
4. IF NOT installed → MUST ask user about installation (REQUIRED prompt)
5. IF user declines installation → Skip CodeRabbit, proceed to Step 8 (only valid skip path)
```

### Anti-Rationalization for CodeRabbit Execution

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "CodeRabbit is optional, I'll skip it" | If installed, it's MANDATORY. Optional only means installation is optional. | **Run CodeRabbit if installed** |
| "Ring reviewers passed, that's enough" | Different tools catch different issues. CodeRabbit complements Ring. | **Run CodeRabbit if installed** |
| "User didn't ask for CodeRabbit" | User installed it. Installation = consent to mandatory execution. | **Run CodeRabbit if installed** |
| "Takes too long, skip this time" | Time is irrelevant. Installed = mandatory. | **Run CodeRabbit if installed** |
| "I'll just proceed without asking about install" | MUST ask every user if they want to install. No silent skips. | **Ask user about installation** |

#### Step 7.5.1: Check CodeRabbit Installation

Run the [CodeRabbit Installation Check](#coderabbit-install-check) command.

**IF INSTALLED AND AUTHENTICATED → MANDATORY EXECUTION (CANNOT SKIP):**
```text
┌─────────────────────────────────────────────────────────────────┐
│ ✅ CodeRabbit CLI detected - MANDATORY EXECUTION                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ CodeRabbit CLI is installed and authenticated.                  │
│                                                                 │
│ ⛔ CodeRabbit review is MANDATORY when installed.               │
│    This step CANNOT be skipped. Proceeding automatically...     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```
→ Proceed directly to Step 7.5.2 (Run CodeRabbit Review) - **NO user prompt, NO skip option**

**IF NOT INSTALLED → MUST ASK USER (REQUIRED PROMPT):**

**⛔ You MUST present this prompt to the user. Silent skips are FORBIDDEN.**

```text
┌─────────────────────────────────────────────────────────────────┐
│ ⚠️  CodeRabbit CLI not found - INSTALLATION PROMPT REQUIRED     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ CodeRabbit CLI is not installed on your system.                 │
│                                                                 │
│ CodeRabbit provides additional AI-powered review that catches:  │
│   • Race conditions and concurrency issues                      │
│   • Memory leaks and resource management                        │
│   • Security vulnerabilities                                    │
│   • Edge cases missed by other reviewers                        │
│                                                                 │
│ ⛔ You MUST choose one of the following options:                │
│                                                                 │
│   (a) Yes, install CodeRabbit CLI (I'll guide you)              │
│   (b) No, skip CodeRabbit and proceed to Gate 5                 │
│                                                                 │
│ ⚠️  ENVIRONMENT CHECK:                                          │
│     • Interactive terminal with browser? → Standard install     │
│     • CI/headless? → Requires API token auth                    │
│     • Container? → See Environment-Specific Guidance above      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**If user selects (a) Yes, install:**
→ Proceed to Installation Flow below

**If user selects (b) No, skip:**
```text
→ Record: "CodeRabbit review: SKIPPED (not installed, user declined installation)"
→ Proceed to Step 8 (Success Output)
→ This is the ONLY valid path to skip CodeRabbit
```

#### Step 7.5.1a: CodeRabbit Installation Flow

```text
┌─────────────────────────────────────────────────────────────────┐
│ 📦 INSTALLING CODERABBIT CLI                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ ⚠️  ENVIRONMENT CHECK FIRST:                                    │
│                                                                 │
│ This installation requires:                                     │
│   • curl command available                                      │
│   • Write access to $HOME or /usr/local/bin                     │
│   • Internet access to cli.coderabbit.ai                        │
│   • Non-containerized environment (or persistent storage)       │
│                                                                 │
│ If in CI/container, see "Environment-Specific Guidance" above.  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Check environment before proceeding:**
```bash
# Verify prerequisites
curl --version && echo "curl: OK" || echo "curl: MISSING"
test -w "$HOME" && echo "HOME writable: OK" || echo "HOME writable: NO"
curl -sI https://cli.coderabbit.ai | head -1 | grep -q "200\|301\|302" && echo "Network: OK" || echo "Network: BLOCKED"
```

**If prerequisites pass, install:**
```text
┌─────────────────────────────────────────────────────────────────┐
│ 📦 Step 1: Installing CodeRabbit CLI...                         │
└─────────────────────────────────────────────────────────────────┘
```

```bash
# Step 1: Download and install CodeRabbit CLI
curl -fsSL https://cli.coderabbit.ai/install.sh | sh
```

**After installation, verify:** Run the [CodeRabbit Installation Check](#coderabbit-install-check) command.

**If installation successful:**
```text
┌─────────────────────────────────────────────────────────────────┐
│ ✅ CodeRabbit CLI installed successfully!                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Step 2: Authentication required                                 │
│                                                                 │
│ Choose your authentication method:                              │
│                                                                 │
│   (a) Browser login (interactive - opens browser)               │
│       → Best for: Local development with GUI                    │
│       → Command: coderabbit auth login                          │
│                                                                 │
│   (b) API token (headless - no browser needed)                  │
│       → Best for: CI/CD, containers, SSH sessions               │
│       → Get token: https://app.coderabbit.ai/settings/api-tokens│
│       → Command: coderabbit auth login --token "cr_xxx"         │
│                                                                 │
│   (c) Skip authentication and CodeRabbit review                 │
│                                                                 │
│ Note: Free tier allows 1 review/hour.                           │
│       Paid plans get enhanced reviews + higher limits.          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**If user selects (a) Browser login:**
```bash
# Step 2a: Authenticate with CodeRabbit (opens browser)
# ⚠️ Requires: GUI environment with default browser
coderabbit auth login
```

**If user selects (b) API token:**
```bash
# Step 2b: Authenticate with API token (headless)
# Get your token from: https://app.coderabbit.ai/settings/api-tokens
coderabbit auth login --token "cr_xxxxxxxxxxxxx"
```

**After authentication:**
```text
┌─────────────────────────────────────────────────────────────────┐
│ ✅ CodeRabbit CLI ready!                                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Installation: Complete                                          │
│ Authentication: Complete                                        │
│                                                                 │
│ Proceeding to CodeRabbit review...                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

→ Proceed to Step 7.5.2 (Run CodeRabbit Review)

**If installation failed:**
```text
┌─────────────────────────────────────────────────────────────────┐
│ ❌ CodeRabbit CLI installation failed                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Error: [error message from curl/sh]                             │
│                                                                 │
│ Troubleshooting:                                                │
│   • Check internet connection                                   │
│   • Try manual install: https://docs.coderabbit.ai/cli/overview │
│   • macOS/Linux only (Windows not supported yet)                │
│                                                                 │
│ Would you like to:                                              │
│   (a) Retry installation                                        │
│   (b) Skip CodeRabbit and proceed to Gate 5                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### Step 7.5.2: Run CodeRabbit Review

**⛔ GRANULAR VALIDATION: CodeRabbit MUST validate at the most granular level available.**

```text
DETERMINE VALIDATION SCOPE:
1. Check if current work has subtasks (from gate0_handoff or implementation context)
2. IF subtasks exist → Validate EACH SUBTASK separately
3. IF no subtasks → Validate the TASK as a whole

WHY GRANULAR VALIDATION:
- Subtask-level validation catches issues early
- Easier to pinpoint which subtask introduced problems
- Prevents "works for task A, breaks task B" scenarios
- Enables incremental fixes without re-running entire review
```

**Step 7.5.2a: Determine Validation Scope**

```text
validation_scope = {
  mode: null,  // "subtask" or "task"
  units: [],   // list of {id, files, commits} to validate
  current_index: 0
}

IF gate0_handoff.subtasks exists AND gate0_handoff.subtasks.length > 0:
  → validation_scope.mode = "subtask"
  → FOR EACH subtask in gate0_handoff.subtasks:
      → Get files changed by this subtask (from commits or file mapping)
      → Add to validation_scope.units: {
          id: subtask.id,
          name: subtask.name,
          files: [files touched by this subtask],
          base_sha: [sha before subtask],
          head_sha: [sha after subtask]
        }

  Display:
  ┌─────────────────────────────────────────────────────────────────┐
  │ 📋 CODERABBIT VALIDATION MODE: SUBTASK-LEVEL                    │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ Detected [N] subtasks. Will validate each separately:          │
  │                                                                 │
  │   1. [subtask-1-id]: [subtask-1-name]                          │
  │      Files: [file1.go, file2.go]                               │
  │                                                                 │
  │   2. [subtask-2-id]: [subtask-2-name]                          │
  │      Files: [file3.go, file4.go]                               │
  │                                                                 │
  │   ... (up to N subtasks)                                       │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘

ELSE:
  → validation_scope.mode = "task"
  → Add single unit: {
      id: unit_id,
      name: implementation_summary,
      files: implementation_files,
      base_sha: base_sha,
      head_sha: head_sha
    }

  Display:
  ┌─────────────────────────────────────────────────────────────────┐
  │ 📋 CODERABBIT VALIDATION MODE: TASK-LEVEL                       │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ No subtasks detected. Validating entire task:                  │
  │                                                                 │
  │   Task: [unit_id]                                              │
  │   Files: [N] files changed                                     │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘
```

**Step 7.5.2b: Run CodeRabbit for Each Validation Unit**

```text
coderabbit_results = {
  overall_status: "PASS",  // PASS only if ALL units pass
  units: []
}

FOR EACH unit IN validation_scope.units:
  Display:
  ┌─────────────────────────────────────────────────────────────────┐
  │ 🔍 VALIDATING: [unit.id] ([current]/[total])                    │
  ├─────────────────────────────────────────────────────────────────┤
  │ Name: [unit.name]                                              │
  │ Files: [unit.files.join(", ")]                                 │
  └─────────────────────────────────────────────────────────────────┘
```

```bash
# Run CodeRabbit review
# ⏱️ TIMING: 7-30+ minutes per review. Run in background if possible.

# Compare against base branch
coderabbit --prompt-only --type uncommitted --base [base_branch]

# Compare against specific commit on current branch
coderabbit --prompt-only --type uncommitted --base-commit [unit.base_sha]

# The command is synchronous - it completes when output is returned
```

```text
  Parse output and record:
  unit_result = {
    id: unit.id,
    status: "PASS" | "ISSUES_FOUND",
    issues: {
      critical: [list],
      high: [list],
      medium: [list],
      low: [list]
    }
  }

  coderabbit_results.units.push(unit_result)

  IF unit_result.issues.critical.length > 0 OR unit_result.issues.high.length > 0:
    → coderabbit_results.overall_status = "ISSUES_FOUND"

  ─────────────────────────────────────────────────────────────────
  ⛔ MANDATORY: APPEND FINDINGS TO .coderabbit-findings.md
  ─────────────────────────────────────────────────────────────────

  After EACH unit validation, append results to findings file:

  IF .coderabbit-findings.md does NOT exist:
    → Create file with header (see "Findings File Format" below)

  APPEND to .coderabbit-findings.md:
  ```
  ## Unit: [unit.id] - [unit.name]
  **Validated:** [timestamp]
  **Status:** [PASS | ISSUES_FOUND]
  **Files:** [unit.files.join(", ")]

  ### Issues Found
  | # | Severity | Description | File:Line | Recommendation |
  |---|----------|-------------|-----------|----------------|
  | 1 | [severity] | [description] | [file:line] | [recommendation] |
  | ... | ... | ... | ... | ... |

  ---
  ```

  This ensures ALL findings are accumulated for review before commit.

AFTER ALL UNITS VALIDATED:
  Display summary:
  ┌─────────────────────────────────────────────────────────────────┐
  │ 📊 CODERABBIT VALIDATION SUMMARY                                │
  ├─────────────────────────────────────────────────────────────────┤
  │ Mode: [SUBTASK-LEVEL | TASK-LEVEL]                             │
  │ Units Validated: [N]                                           │
  │ Overall Status: [PASS | ISSUES_FOUND]                          │
  │                                                                 │
  │ Per-Unit Results:                                              │
  │ ┌──────────────┬────────────┬──────┬──────┬────────┬─────┐     │
  │ │ Unit ID      │ Status     │ Crit │ High │ Medium │ Low │     │
  │ ├──────────────┼────────────┼──────┼──────┼────────┼─────┤     │
  │ │ [subtask-1]  │ ✅ PASS    │  0   │  0   │   0    │  1  │     │
  │ │ [subtask-2]  │ ❌ ISSUES  │  1   │  2   │   0    │  0  │     │
  │ └──────────────┴────────────┴──────┴──────┴────────┴─────┘     │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘
```

**Parse CodeRabbit output for:**
- Critical issues
- High severity issues
- Security vulnerabilities
- Performance concerns

### Findings File Format (.coderabbit-findings.md)

**This file accumulates ALL CodeRabbit findings across all validated units.**

```markdown
# CodeRabbit Findings

**Generated:** [initial timestamp]
**Last Updated:** [latest timestamp]
**Total Units Validated:** [N]
**Overall Status:** [PASS | ISSUES_FOUND]

## Summary

| Severity | Count | Status |
|----------|-------|--------|
| Critical | [N] | [N pending / N fixed] |
| High | [N] | [N pending / N fixed] |
| Medium | [N] | [N pending / N fixed] |
| Low | [N] | [N pending / N fixed] |

---

## Unit: [subtask-1-id] - [subtask-1-name]
**Validated:** [timestamp]
**Status:** [PASS | ISSUES_FOUND]
**Files:** [file1.go, file2.go]

### Issues Found
| # | Severity | Description | File:Line | Recommendation | Status |
|---|----------|-------------|-----------|----------------|--------|
| 1 | CRITICAL | Race condition in handler | handler.go:45 | Use sync.Mutex | PENDING |
| 2 | HIGH | Unchecked error return | repo.go:123 | Handle error | PENDING |

---

## Unit: [subtask-2-id] - [subtask-2-name]
**Validated:** [timestamp]
**Status:** PASS
**Files:** [file3.go]

### Issues Found
_No issues found._

---

[... additional units ...]
```

**File Location:** Project root (`.coderabbit-findings.md`)

**Lifecycle:**
1. Created when first CodeRabbit validation runs
2. Appended after each unit validation
3. Displayed before commit (Step 8)
4. User decides: fix issues or acknowledge and proceed
5. After commit, file can be deleted or kept for audit

#### Step 7.5.3: Handle CodeRabbit Findings

**⛔ CRITICAL: You are an ORCHESTRATOR. You CANNOT edit source files directly.**
**You MUST dispatch the implementation agent to fix issues.**

**⛔ GRANULAR FIX DISPATCH: Fixes MUST be dispatched per-unit (subtask or task).**

```text
IF coderabbit_results.overall_status == "ISSUES_FOUND":

  → FIRST: Display EACH issue in detail (REQUIRED before any action):
  ┌─────────────────────────────────────────────────────────────────┐
  │ ⚠️  CODERABBIT ISSUES FOUND - DETAILED DESCRIPTION               │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ UNIT: [subtask-1] - [subtask name]                             │
  │ ───────────────────────────────────────────────────────────────│
  │ Issue #1 [CRITICAL]                                            │
  │   Description: Race condition in concurrent request handler    │
  │   File: src/handler.go:45                                      │
  │   Code Context:                                                │
  │     43 | func (h *Handler) Process(ctx context.Context) {      │
  │     44 |     h.counter++  // ← NOT THREAD-SAFE                 │
  │     45 |     data := h.sharedMap[key]                          │
  │   Why it matters: Multiple goroutines can corrupt shared state │
  │   Recommendation: Use sync.Mutex or atomic operations          │
  │                                                                 │
  │ Issue #2 [HIGH]                                                │
  │   Description: Unchecked error return from database query      │
  │   File: src/repo.go:123                                        │
  │   Code Context:                                                │
  │     121 | func (r *Repo) GetUser(id string) (*User, error) {   │
  │     122 |     result, _ := r.db.Query(query, id)  // ← IGNORED │
  │     123 |     return parseUser(result), nil                    │
  │   Why it matters: Silent failures can cause data corruption    │
  │   Recommendation: Check and handle the error properly          │
  │                                                                 │
  │ UNIT: [subtask-2] - [subtask name]                             │
  │ ───────────────────────────────────────────────────────────────│
  │ Issue #3 [HIGH]                                                │
  │   Description: SQL injection vulnerability                     │
  │   File: src/query.go:89                                        │
  │   Code Context:                                                │
  │     87 | func BuildQuery(userInput string) string {            │
  │     88 |     return fmt.Sprintf("SELECT * FROM users WHERE     │
  │     89 |            name = '%s'", userInput)  // ← INJECTABLE  │
  │   Why it matters: Attacker can execute arbitrary SQL           │
  │   Recommendation: Use parameterized queries                    │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘

  → THEN: Ask user for action:
  "CodeRabbit found [N] issues in [M] units. What would you like to do?"
    (a) Fix all issues - dispatch implementation agent per unit
    (b) Proceed to Gate 5 (acknowledge risk)
    (c) Review findings in detail (show code context)

  IF user selects (a) Fix issues:
    → ⛔ DO NOT edit files directly
    → FOR EACH unit WITH issues (validation_scope.units where status == "ISSUES_FOUND"):

        Display:
        ┌─────────────────────────────────────────────────────────────────┐
        │ 🔧 DISPATCHING FIX: [unit.id] ([current]/[total with issues])   │
        ├─────────────────────────────────────────────────────────────────┤
        │ Unit: [unit.name]                                              │
        │ Critical Issues: [N]                                           │
        │ High Issues: [N]                                               │
        └─────────────────────────────────────────────────────────────────┘

        → DISPATCH implementation agent with unit-specific findings:

        Task:
          subagent_type: "[same agent used in Gate 0]"
          description: "Fix CodeRabbit issues for [unit.id]"
          prompt: |
            ## CodeRabbit Issues to Fix - [unit.id]

            **Scope:** This fix is for [subtask/task]: [unit.name]
            **Files in Scope:** [unit.files.join(", ")]

            The following issues were found by CodeRabbit CLI external review
            for THIS SPECIFIC [subtask/task].

            ⚠️ IMPORTANT: Only fix issues in files belonging to this unit:
            [unit.files list]

            ### Critical Issues
            [list from unit.issues.critical]

            ### High Issues
            [list from unit.issues.high]

            ## Requirements
            1. Fix each issue following Ring Standards
            2. Only modify files in scope: [unit.files]
            3. Run tests to verify fixes don't break functionality
            4. Commit fixes with message referencing unit: "fix([unit.id]): [description]"

        → Wait for agent to complete
        → Record fix result for this unit

        → VALIDATE EACH ISSUE INDIVIDUALLY:
        ┌─────────────────────────────────────────────────────────────────┐
        │ 🔍 VALIDATING FIXES FOR: [unit.id]                              │
        ├─────────────────────────────────────────────────────────────────┤
        │                                                                 │
        │ Each issue MUST be validated individually:                      │
        │                                                                 │
        │ Issue #1: [issue description]                                   │
        │   File: [file:line]                                            │
        │   Severity: CRITICAL                                           │
        │   Fix Applied: [description of fix]                            │
        │   Validation: ✅ RESOLVED / ❌ NOT RESOLVED                     │
        │   Evidence: [code snippet or test result]                      │
        │                                                                 │
        │ Issue #2: [issue description]                                   │
        │   File: [file:line]                                            │
        │   Severity: HIGH                                               │
        │   Fix Applied: [description of fix]                            │
        │   Validation: ✅ RESOLVED / ❌ NOT RESOLVED                     │
        │   Evidence: [code snippet or test result]                      │
        │                                                                 │
        │ ... (repeat for ALL issues)                                    │
        │                                                                 │
        └─────────────────────────────────────────────────────────────────┘

        → IF any issue NOT RESOLVED:
            → Identify the correct agent for re-dispatch:
              - Check gate0_handoff.implementation_agent (if available)
              - OR infer from file type:
                - *.go files → ring:backend-engineer-golang
                - *.ts files (backend) → ring:backend-engineer-typescript
                - *.ts/*.tsx files (frontend) → ring:frontend-engineer
                - *.yaml/*.yml (infra) → ring:devops-engineer

            → Re-dispatch ONLY unresolved issues to the correct agent:

            Task:
              subagent_type: "[correct agent based on file type or gate0_handoff]"
              model: "opus"
              description: "Retry fix for unresolved issues in [unit.id]"
              prompt: |
                ## RETRY: Unresolved CodeRabbit Issues - [unit.id]

                Previous fix attempt did NOT resolve these issues.
                This is attempt [N] of 2 maximum.

                ### Unresolved Issues (MUST FIX)
                | # | Severity | Description | File:Line | Previous Attempt | Why It Failed |
                |---|----------|-------------|-----------|------------------|---------------|
                | [issue.id] | [severity] | [description] | [file:line] | [what was tried] | [why not resolved] |

                ### Requirements
                1. Review the previous fix attempt and understand why it failed
                2. Apply a different/better solution
                3. Verify the fix resolves the issue
                4. Run relevant tests
                5. Commit with message: "fix([unit.id]): retry [issue description]"

            → Max 2 fix attempts per issue
            → IF issue still NOT RESOLVED after 2 attempts:
                → Mark as UNRESOLVED_ESCALATE
                → Add to escalation report for manual review

        → Record per-issue validation results:
        unit_validation = {
          id: unit.id,
          issues_validated: [
            {
              issue_id: 1,
              description: "[issue]",
              severity: "CRITICAL",
              file: "[file:line]",
              fix_applied: "[description]",
              status: "RESOLVED" | "NOT_RESOLVED",
              evidence: "[snippet or test]",
              attempts: 1
            },
            ...
          ],
          all_resolved: true | false
        }

    → AFTER ALL UNITS FIXED:
        Display:
        ┌─────────────────────────────────────────────────────────────────┐
        │ ✅ FIX DISPATCH COMPLETE                                        │
        ├─────────────────────────────────────────────────────────────────┤
        │ Units Fixed: [N] / [total with issues]                         │
        │ Total Issues Validated: [N]                                    │
        │ Issues Resolved: [N] / [N]                                     │
        │                                                                 │
        │ Per-Unit Fix Status:                                           │
        │ ┌──────────────┬────────────┬───────────────────────┐          │
        │ │ Unit ID      │ Status     │ Commit                │          │
        │ ├──────────────┼────────────┼───────────────────────┤          │
        │ │ [subtask-1]  │ ✅ FIXED   │ abc123                │          │
        │ │ [subtask-2]  │ ✅ FIXED   │ def456                │          │
        │ └──────────────┴────────────┴───────────────────────┘          │
        │                                                                 │
        │ Issue-Level Validation Details:                                │
        │ ┌──────────────────────────────────────────────────────────┐   │
        │ │ UNIT: [subtask-1]                                        │   │
        │ ├──────────────────────────────────────────────────────────┤   │
        │ │ #1 [CRITICAL] Race condition in handler                  │   │
        │ │    File: src/handler.go:45                               │   │
        │ │    Fix: Added mutex lock                                 │   │
        │ │    Status: ✅ RESOLVED                                   │   │
        │ │    Evidence: Test race_test.go passes                    │   │
        │ ├──────────────────────────────────────────────────────────┤   │
        │ │ #2 [HIGH] Unchecked error return                         │   │
        │ │    File: src/handler.go:67                               │   │
        │ │    Fix: Added error check with proper handling           │   │
        │ │    Status: ✅ RESOLVED                                   │   │
        │ │    Evidence: Error path verified in unit test            │   │
        │ └──────────────────────────────────────────────────────────┘   │
        │                                                                 │
        └─────────────────────────────────────────────────────────────────┘

LEGACY FLOW (when validation_scope.mode == "task"):
  IF CodeRabbit found CRITICAL or HIGH issues:
    → Display findings to user
    → Ask: "CodeRabbit found [N] critical/high issues. Fix now or proceed anyway?"
      (a) Fix issues - dispatch to implementation agent
      (b) Proceed to Gate 5 (acknowledge risk)
      (c) Review findings in detail

    IF user selects (a) Fix issues:
      → ⛔ DO NOT edit files directly
      → DISPATCH implementation agent with CodeRabbit findings:

      Task:
        subagent_type: "[same agent used in Gate 0]"
        model: "opus"
        description: "Fix CodeRabbit issues for [unit_id]"
        prompt: |
          ## CodeRabbit Issues to Fix

          The following issues were found by CodeRabbit CLI external review.
          Fix ALL Critical and High severity issues.

          ### Critical Issues
          [list from CodeRabbit output]

          ### High Issues
          [list from CodeRabbit output]

          ## Requirements
          1. Fix each issue following Ring Standards
          2. Run tests to verify fixes don't break functionality
          3. Commit fixes with descriptive message

    → After agent completes, re-run CodeRabbit: `coderabbit --prompt-only`
    → If CodeRabbit issues remain, repeat fix cycle (max 2 iterations for CodeRabbit)

    → ⛔ AFTER CodeRabbit passes, MUST re-run Ring reviewers:

    ┌─────────────────────────────────────────────────────────────────┐
    │ 🔄 RE-RUNNING RING REVIEWERS AFTER CODERABBIT FIXES             │
    ├─────────────────────────────────────────────────────────────────┤
    │                                                                 │
    │ CodeRabbit fixes may have introduced new issues detectable by   │
    │ Ring reviewers. Re-validation is MANDATORY before Gate 5.       │
    │                                                                 │
    └─────────────────────────────────────────────────────────────────┘

    Step 7.5.3a: Re-Run All 5 Ring Reviewers
    ─────────────────────────────────────────
    1. Get new HEAD_SHA after CodeRabbit fixes
    2. Dispatch all 5 reviewers in parallel (per Step 3):
       - ring:code-reviewer
       - ring:business-logic-reviewer
       - ring:security-reviewer
       - ring:test-reviewer
       - ring:nil-safety-reviewer
    3. Wait for all 5 to complete

    Step 7.5.3b: Handle Ring Reviewer Results
    ─────────────────────────────────────────
    IF all 5 Ring reviewers PASS:
      → Proceed to Step 8 (Success Output)

    IF any Ring reviewer finds CRITICAL/HIGH/MEDIUM issues:
      → Increment ring_revalidation_iterations counter
      → IF ring_revalidation_iterations >= 2:
          → ESCALATE: "Max iterations reached after CodeRabbit fixes"
          → Go to Step 9 (Escalate)
      → DISPATCH implementation agent to fix Ring reviewer issues
      → After fixes committed:
          → Re-run CodeRabbit: `coderabbit --prompt-only`
          → IF CodeRabbit passes:
              → Re-run all 5 Ring reviewers (loop back to Step 7.5.3a)
          → IF CodeRabbit finds issues:
              → Fix CodeRabbit issues first, then re-run Ring reviewers

    State tracking for CodeRabbit fix cycle:
    ```
    coderabbit_fix_state = {
      coderabbit_iterations: 0,      // max 2 for CodeRabbit-only fixes
      ring_revalidation_iterations: 0,  // max 2 for Ring reviewer re-runs
      total_max_iterations: 4        // absolute cap: 2 CR + 2 Ring
    }
    ```

IF CodeRabbit found only MEDIUM/LOW issues:
  → Display summary
  → ⛔ DO NOT edit files directly to add TODOs
  → DISPATCH implementation agent to add TODO comments:

  Task:
    subagent_type: "[same agent used in Gate 0]"
    description: "Add TODO comments for CodeRabbit findings"
    prompt: |
      Add TODO comments for these CodeRabbit findings:
      [list MEDIUM/LOW issues with file:line]

      Format: // TODO(coderabbit): [issue description]

  → After TODO comments added (code changed):
      → Re-run all 5 Ring reviewers (per Step 7.5.3a above)
      → IF Ring reviewers PASS: Proceed to Step 8
      → IF Ring reviewers find issues: Fix and re-run (max 2 iterations)

IF CodeRabbit found no issues:
  → Display: "✅ CodeRabbit review passed - no additional issues found"
  → No code changes made by CodeRabbit flow
  → Proceed directly to Step 8 (no Ring re-run needed)
```

### Anti-Rationalization for Direct Editing

**See [shared-patterns/orchestrator-direct-editing-anti-rationalization.md](../shared-patterns/orchestrator-direct-editing-anti-rationalization.md) - same table applies here.**

*Applies to: Step 6 (Fix dispatch after Ring reviewers) & Step 7.5.3 (Fix dispatch after CodeRabbit)*

#### Step 7.5.4: CodeRabbit Results Summary

```markdown
## CodeRabbit External Review
**Status:** [PASS|ISSUES_FOUND|SKIPPED]
**Validation Mode:** [SUBTASK-LEVEL|TASK-LEVEL]
**Units Validated:** [N]
**Total Issues Found:** [N]
**Issues Resolved:** [N]/[N]

### Per-Unit Validation Results
| Unit ID | Unit Name | Status | Critical | High | Medium | Low |
|---------|-----------|--------|----------|------|--------|-----|
| [subtask-1] | [name] | ✅ PASS | 0 | 0 | 0 | 1 |
| [subtask-2] | [name] | ✅ FIXED | 1→0 | 2→0 | 0 | 0 |
| [task-id] | [name] | ✅ PASS | 0 | 0 | 0 | 0 |

### Issues Found - Detailed Description (ALWAYS shown when issues exist)

#### Unit: [subtask-2]
| # | Severity | Description | File:Line | Code Context | Why It Matters | Recommendation |
|---|----------|-------------|-----------|--------------|----------------|----------------|
| 1 | CRITICAL | Race condition | handler.go:45 | `h.counter++` not thread-safe | Corrupts shared state | Use sync.Mutex |
| 2 | HIGH | Unchecked error | repo.go:123 | `result, _ := r.db.Query()` | Silent failures | Handle error |
| 3 | HIGH | SQL injection | query.go:89 | `fmt.Sprintf("...%s", input)` | Security breach | Parameterized query |

### Issue-Level Validation (REQUIRED after fixes are applied)

#### Unit: [subtask-2]
| # | Severity | Description | File:Line | Fix Applied | Status | Evidence |
|---|----------|-------------|-----------|-------------|--------|----------|
| 1 | CRITICAL | Race condition in concurrent handler | handler.go:45 | Added mutex lock around shared state | ✅ RESOLVED | race_test.go passes |
| 2 | HIGH | Unchecked error from DB query | repo.go:123 | Added error check with rollback | ✅ RESOLVED | Error path tested |
| 3 | HIGH | SQL injection vulnerability | query.go:89 | Used parameterized query | ✅ RESOLVED | Security test added |

#### Unit: [subtask-3] (if applicable)
| # | Severity | Description | File:Line | Fix Applied | Status | Evidence |
|---|----------|-------------|-----------|-------------|--------|----------|
| 1 | HIGH | Missing input validation | api.go:34 | Added validation middleware | ✅ RESOLVED | Fuzz test passes |

### Overall Summary by Severity
| Severity | Found | Resolved | Remaining | Action |
|----------|-------|----------|-----------|--------|
| Critical | [N] | [N] | 0 | Fixed |
| High | [N] | [N] | 0 | Fixed |
| Medium | [N] | [N] | 0 | TODO added |
| Low | [N] | - | [N] | TODO added |
```

### CodeRabbit Skip Scenarios (ONLY VALID PATHS)

**⛔ CodeRabbit can ONLY be skipped in these specific scenarios. Any other skip is a VIOLATION.**

| Scenario | Record As | Next Step | Why Valid |
|----------|-----------|-----------|-----------|
| CLI not installed, user **explicitly** declines install | `SKIPPED (not installed, user declined)` | Step 8 | User was asked and chose not to install |
| Installation failed after attempt, user skips | `SKIPPED (installation failed)` | Step 8 | Technical failure, user informed |
| Authentication failed after attempt, user skips | `SKIPPED (auth failed)` | Step 8 | Technical failure, user informed |
| Environment doesn't support (CI/container) | `SKIPPED (unsupported environment)` | Step 8 | Technical limitation |

**⛔ INVALID SKIP SCENARIOS (FORBIDDEN):**

| Invalid Scenario | Why FORBIDDEN | Required Action |
|------------------|---------------|-----------------|
| CLI installed but "skipped for speed" | Installed = MANDATORY | **Run CodeRabbit** |
| CLI installed but "Ring reviewers passed" | Complementary tools, both required | **Run CodeRabbit** |
| CLI not installed, no prompt shown | MUST ask user about installation | **Show installation prompt** |
| Silent skip without user interaction | All skips require explicit user choice | **Ask user** |

**⛔ NON-NEGOTIABLE:** When CodeRabbit CLI IS installed and authenticated, execution is **MANDATORY** - it CANNOT be skipped under any circumstance.

---

## Step 8: Display Accumulated Findings & Prepare Success Output

**⛔ BEFORE generating success output, MUST display accumulated CodeRabbit findings.**

### Step 8.1: Display Accumulated CodeRabbit Findings

```text
IF .coderabbit-findings.md exists:

  ┌─────────────────────────────────────────────────────────────────┐
  │ 📋 CODERABBIT FINDINGS - ACCUMULATED DURING REVIEW              │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ The following issues were identified by CodeRabbit during the  │
  │ review process. Review before proceeding to commit.            │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘

  → Display contents of .coderabbit-findings.md
  → Show summary table:

  ┌─────────────────────────────────────────────────────────────────┐
  │ 📊 CODERABBIT FINDINGS SUMMARY                                  │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ | Severity | Count | Status |                                  │
  │ |----------|-------|--------|                                  │
  │ | Critical | [N]   | [pending/fixed] |                         │
  │ | High     | [N]   | [pending/fixed] |                         │
  │ | Medium   | [N]   | [pending/fixed] |                         │
  │ | Low      | [N]   | [pending/fixed] |                         │
  │                                                                 │
  │ Total Issues: [N] | Fixed: [N] | Pending: [N]                  │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘

  → Ask user:
  ┌─────────────────────────────────────────────────────────────────┐
  │ ❓ ACTION REQUIRED                                              │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ [N] CodeRabbit issues are pending. What would you like to do?  │
  │                                                                 │
  │   (a) Fix all pending issues now (dispatch implementation agent)│
  │   (b) Review and fix issues one-by-one (interactive mode)      │
  │   (c) Acknowledge and proceed to commit (issues documented)    │
  │                                                                 │
  │ Note: Choosing (c) will include findings file in commit for    │
  │       tracking. Issues remain documented for future fixing.    │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘

  IF user selects (a) Fix all issues:
    → Dispatch implementation agent with ALL pending issues from findings file
    → After fixes, update .coderabbit-findings.md (mark issues as FIXED)
    → Re-run CodeRabbit validation for affected files
    → Loop back to Step 8.1 to display updated findings

  IF user selects (b) Interactive mode (one-by-one):
    → Go to Step 8.1.1 (Interactive Issue Review)

  IF user selects (c) Acknowledge and proceed:
    → Record: "CodeRabbit issues acknowledged by user"
    → Include .coderabbit-findings.md in commit (for audit trail)
    → Proceed to Step 8.2 (Success Output)

─────────────────────────────────────────────────────────────────
Step 8.1.1: Interactive Issue Review (One-by-One)
─────────────────────────────────────────────────────────────────

issues_to_fix = []
issues_to_skip = []

FOR EACH issue IN pending_issues (ordered by severity: CRITICAL → HIGH → MEDIUM → LOW):

  Display:
  ┌─────────────────────────────────────────────────────────────────┐
  │ 🔍 ISSUE [current]/[total] - [SEVERITY]                         │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ Unit: [unit.id] - [unit.name]                                  │
  │ File: [file:line]                                              │
  │                                                                 │
  │ Description:                                                   │
  │   [issue description]                                          │
  │                                                                 │
  │ Code Context:                                                  │
  │   [code snippet around the issue]                              │
  │                                                                 │
  │ Why it matters:                                                │
  │   [explanation of impact]                                      │
  │                                                                 │
  │ Recommendation:                                                │
  │   [suggested fix]                                              │
  │                                                                 │
  ├─────────────────────────────────────────────────────────────────┤
  │ What would you like to do with this issue?                     │
  │                                                                 │
  │   (f) Fix this issue                                           │
  │   (s) Skip this issue (acknowledge)                            │
  │   (a) Fix ALL remaining issues                                 │
  │   (k) Skip ALL remaining issues                                │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘

  IF user selects (f) Fix:
    → Add to issues_to_fix list
    → Continue to next issue

  IF user selects (s) Skip:
    → Add to issues_to_skip list
    → Continue to next issue

  IF user selects (a) Fix ALL remaining:
    → Add current + all remaining to issues_to_fix list
    → Break loop

  IF user selects (k) Skip ALL remaining:
    → Add current + all remaining to issues_to_skip list
    → Break loop

AFTER loop completes:
  Display summary:
  ┌─────────────────────────────────────────────────────────────────┐
  │ 📋 INTERACTIVE REVIEW COMPLETE                                  │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ Issues to fix: [N]                                             │
  │   [list of issues selected for fixing]                         │
  │                                                                 │
  │ Issues to skip: [N]                                            │
  │   [list of issues selected to skip]                            │
  │                                                                 │
  │ Proceed with this selection? (y/n)                             │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘

  IF user confirms (y):
    IF issues_to_fix.length > 0:
      → Dispatch implementation agent with ONLY issues_to_fix
      → After fixes, update .coderabbit-findings.md:
        - Mark fixed issues as FIXED
        - Mark skipped issues as ACKNOWLEDGED
      → Re-run CodeRabbit validation for affected files
      → Loop back to Step 8.1
    ELSE:
      → All issues skipped/acknowledged
      → Proceed to Step 8.2 (Success Output)

  IF user cancels (n):
    → Return to Step 8.1 main prompt

ELSE (no findings file exists):
  → CodeRabbit was skipped or found no issues
  → Proceed directly to Step 8.2 (Success Output)
```

### Step 8.2: Generate Success Output

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
| ring:code-reviewer | ✅ PASS | [count] |
| ring:business-logic-reviewer | ✅ PASS | [count] |
| ring:security-reviewer | ✅ PASS | [count] |
| ring:test-reviewer | ✅ PASS | [count] |
| ring:nil-safety-reviewer | ✅ PASS | [count] |

## Low/Cosmetic Issues (TODO/FIXME added)
[list with file locations]

## CodeRabbit Findings
**Findings File:** .coderabbit-findings.md
**Total Issues Found:** [N]
**Issues Fixed:** [N]
**Issues Acknowledged:** [N]
**Status:** [ALL_FIXED | ACKNOWLEDGED | NO_ISSUES]

## Handoff to Next Gate
- Review status: COMPLETE
- All blocking issues: RESOLVED
- Reviewers passed: 5/5
- CodeRabbit findings: [status]
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
| ring:code-reviewer | [PASS/FAIL] |
| ring:business-logic-reviewer | [PASS/FAIL] |
| ring:security-reviewer | [PASS/FAIL] |

## Handoff to Next Gate
- Review status: FAILED
- Unresolved blocking issues: [count]
- Ready for Gate 5: NO
- **Action Required:** User must manually resolve issues

⛔ ESCALATION: Max iterations (3) reached. Blocking issues remain.
```

---

## Pressure Resistance

See [dev-team/skills/shared-patterns/shared-pressure-resistance.md](../../dev-team/skills/shared-patterns/shared-pressure-resistance.md) for universal pressure scenarios.

| User Says | Your Response |
|-----------|---------------|
| "Skip review, code is simple" | "Simple code can have security issues. Dispatching all 5 reviewers." |
| "Just run ring:code-reviewer" | "All 5 reviewers run in parallel. No time saved by skipping." |
| "Fix later, merge now" | "Blocking issues (Critical/High/Medium) MUST be fixed before Gate 5." |

## Anti-Rationalization Table

See [dev-team/skills/shared-patterns/shared-anti-rationalization.md](../../dev-team/skills/shared-patterns/shared-anti-rationalization.md) for universal anti-rationalizations.

### Gate 4-Specific Anti-Rationalizations

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "Run reviewers one at a time" | Sequential = slow. Parallel = 5x faster. | **Dispatch all 5 in single message** |
| "Skip security for internal code" | Internal code can have vulnerabilities. | **Include ring:security-reviewer** |
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
| ring:code-reviewer | ✅/❌ |
| ring:business-logic-reviewer | ✅/❌ |
| ring:security-reviewer | ✅/❌ |
| ring:test-reviewer | ✅/❌ |
| ring:nil-safety-reviewer | ✅/❌ |

## CodeRabbit External Review (MANDATORY if installed, Optional to install)
**Status:** [PASS|ISSUES_FOUND|SKIPPED|NOT_INSTALLED]
**Validation Mode:** [SUBTASK-LEVEL|TASK-LEVEL]
**Units Validated:** [N]
**Units Passed:** [N]/[N]
**Issues Found:** [N]
**Issues Resolved:** [N]/[N]

### Per-Unit Results (if subtask-level)
| Unit ID | Status | Critical | High | Medium | Low |
|---------|--------|----------|------|--------|-----|
| [subtask-1] | ✅ PASS | 0 | 0 | 0 | 1 |
| [subtask-2] | ✅ FIXED | 0 | 0 | 0 | 0 |

### Issue-Level Validation (REQUIRED when issues were fixed)
| Unit | # | Severity | Description | Fix Applied | Status | Evidence |
|------|---|----------|-------------|-------------|--------|----------|
| subtask-2 | 1 | CRITICAL | Race condition | Mutex added | ✅ RESOLVED | Test passes |
| subtask-2 | 2 | HIGH | Unchecked error | Error handling added | ✅ RESOLVED | Test passes |

## Handoff to Next Gate
- Review status: [COMPLETE|FAILED]
- Blocking issues: [resolved|N remaining]
- CodeRabbit: [PASS|SKIPPED|N issues acknowledged]
- CodeRabbit validation: [N]/[N] units passed
- Ready for Gate 5: [YES|NO]
```
