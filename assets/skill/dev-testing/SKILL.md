---
name: "ring:dev-testing"
description: |
  Gate 3 of development cycle - ensures unit test coverage meets threshold (85%+)
  for all acceptance criteria using TDD methodology.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: development
  trigger:
    - After implementation and SRE complete (Gate 0/1/2)
    - Task has acceptance criteria requiring test coverage
    - Need to verify implementation meets requirements
  sequence:
    after: [ring:dev-implementation, ring:dev-devops, ring:dev-sre]
    before: [ring:requesting-code-review]
---

# Dev Testing (Consolidated: Gates 3-7)

This skill consolidates the 5 testing gates from the Ring development cycle. In the official Ring, these are separate skills (dev-unit-testing, dev-fuzz-testing, dev-property-testing, dev-integration-testing, dev-chaos-testing). This consolidated version provides all testing gate content in a single file for OpenCode compatibility.

---

# Gate 3: Unit Testing


# Dev Unit Testing (Gate 3)

## Overview

Ensure every acceptance criterion has at least one **unit test** proving it works. Follow TDD methodology: RED (failing test) -> GREEN (implementation) -> REFACTOR.

**Core principle:** Untested acceptance criteria are unverified claims. Each criterion MUST map to at least one executable unit test.

<block_condition>
- Coverage below 85% = FAIL
- Any acceptance criterion without test = FAIL
</block_condition>

**Coverage threshold:** 85% minimum (Ring standard). PROJECT_RULES.md can raise, not lower.

## CRITICAL: Role Clarification

**This skill ORCHESTRATES. QA Analyst Agent EXECUTES.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Gather requirements, dispatch agent, track iterations |
| **QA Analyst Agent** | Write tests, run coverage, report results |

---

## Step 1: Validate Input

```text
REQUIRED INPUT (from ring:dev-cycle orchestrator):
<verify_before_proceed>
- unit_id exists
- acceptance_criteria is not empty
- implementation_files is not empty
- language is valid (go|typescript|python)
</verify_before_proceed>

```text
- unit_id: [task/subtask being tested]
- acceptance_criteria: [list of ACs to test]
- implementation_files: [files from Gate 0]
- language: [go|typescript|python]

OPTIONAL INPUT:
- coverage_threshold: [default 85.0, cannot be lower]
- gate0_handoff: [full Gate 0 output]
- existing_tests: [existing test files]

if any REQUIRED input is missing:
  → STOP and report: "Missing required input: [field]"
  → Return to orchestrator with error

if coverage_threshold < 85:
  → STOP and report: "Coverage threshold cannot be below Ring minimum (85%)"
  → Use 85% as threshold
```

## Step 2: Initialize Testing State

```text
testing_state = {
  unit_id: [from input],
  coverage_threshold: max(85, [from input]),
  coverage_actual: null,
  verdict: null,
  iterations: 0,
  max_iterations: 3,
  traceability_matrix: [],
  tests_written: 0
}
```

## Step 3: Dispatch QA Analyst Agent

<dispatch_required agent="ring:qa-analyst">
Write unit tests for all acceptance criteria with 85%+ coverage.
</dispatch_required>

```yaml
Task:
  subagent_type: "ring:qa-analyst"
  description: "Write unit tests for [unit_id]"
  prompt: |
    ⛔ WRITE UNIT TESTS for All Acceptance Criteria

    ## Input Context
    - **Unit ID:** [unit_id]
    - **Language:** [language]
    - **Coverage Threshold:** [coverage_threshold]%

    ## Acceptance Criteria to Test
    [list acceptance_criteria with AC-1, AC-2, etc.]

    ## Implementation Files to Test
    [list implementation_files]

    ## Standards Reference
    For Go: https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/golang.md
    For TS: https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/typescript.md

    Focus on: Testing Patterns section

    ## Requirements

    ### Test Coverage
    - Minimum: [coverage_threshold]% branch coverage
    - Every AC MUST have at least one test
    - Edge cases REQUIRED (null, empty, boundary, error conditions)

    ### Test Naming
    - Go: `Test{Unit}_{Method}_{Scenario}`
    - TS: `describe('{Unit}', () => { it('should {scenario}', ...) })`

    ### Test Structure
    - One behavior per test
    - Arrange-Act-Assert pattern
    - Mock all external dependencies
    - no database/API calls (unit tests only)

    ### Edge Cases Required per AC Type

    <cannot_skip>
    - Minimum 3 edge cases per AC type
    - null, empty, boundary conditions required
    - Error conditions required
    </cannot_skip>

    | AC Type | Required Edge Cases | Minimum |
    |---------|---------------------|---------|
    | Input validation | null, empty, boundary, invalid format | 3+ |
    | CRUD operations | not found, duplicate, concurrent | 3+ |
    | Business logic | zero, negative, overflow, boundary | 3+ |
    | Error handling | timeout, connection failure, retry | 2+ |

    ## Required Output Format

    ### Test Files Created
    | File | Tests | Lines |
    |------|-------|-------|
    | [path] | [count] | +N |

    ### Coverage Report
    **Command:** [coverage command]
    **Result:**
    ```
    [paste actual coverage output]
    ```

    | Package/File | Coverage |
    |--------------|----------|
    | [name] | [X%] |
    | **TOTAL** | **[X%]** |

    ### Traceability Matrix
    | AC ID | Criterion | Test File | Test Function | Status |
    |-------|-----------|-----------|---------------|--------|
    | AC-1 | [criterion text] | [file] | [function] | ✅/❌ |
    | AC-2 | [criterion text] | [file] | [function] | ✅/❌ |

    ### Quality Checks
    | Check | Status |
    |-------|--------|
    | No skipped tests | ✅/❌ |
    | No assertion-less tests | ✅/❌ |
    | Edge cases per AC | ✅/❌ |
    | Test isolation | ✅/❌ |

    ### VERDICT
    **Coverage:** [X%] vs Threshold [Y%]
    **VERDICT:** PASS / FAIL
    
    If FAIL:
    - **Gap Analysis:** [what needs more tests]
    - **Files needing coverage:** [list with line numbers]
```

## Step 4: Parse QA Analyst Output

```text
Parse agent output:

1. Extract coverage percentage from Coverage Report
2. Extract traceability matrix
3. Extract verdict

testing_state.coverage_actual = [extracted coverage]
testing_state.traceability_matrix = [extracted matrix]
testing_state.tests_written = [count from Test Files Created]

if verdict == "PASS" and coverage_actual >= coverage_threshold:
  → testing_state.verdict = "PASS"
  → Proceed to Step 6

if verdict == "FAIL" or coverage_actual < coverage_threshold:
  → testing_state.verdict = "FAIL"
  → testing_state.iterations += 1
  → if iterations >= max_iterations: Go to Step 7 (Escalate)
  → Go to Step 5 (Dispatch Fix)
```

## Step 5: Dispatch Fix to Implementation Agent

**Coverage below threshold → Return to Gate 0 for more tests**

```yaml
Task:
  subagent_type: "[implementation_agent from Gate 0]"  # e.g., "ring:backend-engineer-golang"
  description: "Add tests to meet coverage threshold for [unit_id]"
  prompt: |
    ⛔ COVERAGE BELOW THRESHOLD - Add More Tests

    ## Current Status
    - **Coverage Actual:** [coverage_actual]%
    - **Coverage Threshold:** [coverage_threshold]%
    - **Gap:** [threshold - actual]%
    - **Iteration:** [iterations] of [max_iterations]

    ## Gap Analysis (from QA)
    [paste gap analysis from QA output]

    ## Files Needing Coverage
    [paste files list from QA output]

    ## Requirements
    1. Add tests to cover the identified gaps
    2. Focus on edge cases and error paths
    3. Run coverage after each addition
    4. Stop when coverage >= [threshold]%

    ## Required Output
    - Tests added: [list]
    - New coverage: [X%]
    - Coverage command output
```

After fix → Go back to Step 3 (Re-dispatch QA Analyst)

## Step 6: Prepare Success Output

```text
Generate skill output:

## Testing Summary
**Status:** PASS
**Unit ID:** [unit_id]
**Iterations:** [testing_state.iterations]

## Coverage Report
**Threshold:** [coverage_threshold]%
**Actual:** [coverage_actual]%
**Status:** ✅ PASS

| Package/File | Coverage |
|--------------|----------|
[from QA output]
| **TOTAL** | **[coverage_actual]%** |

## Traceability Matrix
| AC ID | Criterion | Test | Status |
|-------|-----------|------|--------|
[from testing_state.traceability_matrix]

**Criteria Covered:** [X]/[Y] (100%)

## Quality Checks
| Check | Status |
|-------|--------|
| Coverage ≥ threshold | ✅ |
| All ACs tested | ✅ |
| No skipped tests | ✅ |
| Edge cases present | ✅ |

## Handoff to Next Gate
- Testing status: COMPLETE
- Coverage: [coverage_actual]% (threshold: [coverage_threshold]%)
- All criteria tested: ✅
- Ready for Gate 4 (Review): YES
```

## Step 7: Escalate - Max Iterations Reached

```text
Generate skill output:

## Testing Summary
**Status:** FAIL
**Unit ID:** [unit_id]
**Iterations:** [max_iterations] (MAX REACHED)

## Coverage Report
**Threshold:** [coverage_threshold]%
**Actual:** [coverage_actual]%
**Gap:** [threshold - actual]%
**Status:** ❌ FAIL

## Gap Analysis
[from last QA output]

## Files Still Needing Coverage
[from last QA output]

## Handoff to Next Gate
- Testing status: FAILED
- Ready for Gate 4: no
- **Action Required:** User must manually add tests or adjust scope

⛔ ESCALATION: Max iterations (3) reached. Coverage still below threshold.
User intervention required.
```

---

## Pressure Resistance

See [shared-patterns/shared-pressure-resistance.md](../shared-patterns/shared-pressure-resistance.md) for universal pressure scenarios.

| User Says | Your Response |
|-----------|---------------|
| "84% is close enough" | "85% is minimum threshold. 84% = FAIL. Adding more tests." |
| "Manual testing covers it" | "Gate 3 requires executable unit tests. Dispatching QA analyst." |
| "Skip testing, deadline" | "Testing is MANDATORY. Untested code = unverified claims." |

---

## Anti-Rationalization Table

See [shared-patterns/shared-anti-rationalization.md](../shared-patterns/shared-anti-rationalization.md) for universal anti-rationalizations.

### Gate 3-Specific Anti-Rationalizations

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "Tool shows 83% but real is 90%" | Tool output IS real. Your belief is not. | **Fix issue, re-measure** |
| "Excluding dead code gets us to 85%" | Delete dead code, don't exclude it. | **Delete dead code** |
| "84.5% rounds to 85%" | Rounding is not allowed. 84.5% < 85%. | **Write more tests** |
| "Close enough with all AC tested" | "Close enough" is not passing. | **Meet exact threshold** |
| "Integration tests cover this" | Gate 3 = unit tests only. Different scope. | **Write unit tests** |

## Unit Test vs Integration Test

| Type | Characteristics | Gate 3? |
|------|----------------|---------|
| **Unit** ✅ | Mocks all external deps, tests single function | YES |
| **Integration** ❌ | Hits real database/API/filesystem | no |

---

## Execution Report Format

```markdown
## Testing Summary
**Status:** [PASS|FAIL]
**Unit ID:** [unit_id]
**Duration:** [Xm Ys]
**Iterations:** [N]

## Coverage Report
**Threshold:** [X%]
**Actual:** [Y%]
**Status:** [✅ PASS | ❌ FAIL]

## Traceability Matrix
| AC ID | Criterion | Test | Status |
|-------|-----------|------|--------|
| AC-1 | [text] | [test] | ✅/❌ |

**Criteria Covered:** [X/Y]

## Handoff to Next Gate
- Testing status: [COMPLETE|FAILED]
- Coverage: [X%]
- Ready for Gate 4: [YES|no]
```

---

# Gate 4: Fuzz Testing


# Dev Fuzz Testing (Gate 4)

## Overview

Ensure critical parsing and input handling code has **fuzz tests** to discover crashes and edge cases through random input generation.

**Core principle:** Fuzz tests find bugs you didn't think to test for. They're mandatory for all code that handles external input.

<block_condition>
- No fuzz functions = FAIL
- Seed corpus < 5 entries = FAIL
- Any crash found = FAIL (fix and re-run)
</block_condition>

## CRITICAL: Role Clarification

**This skill ORCHESTRATES. QA Analyst Agent (fuzz mode) EXECUTES.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Gather requirements, dispatch agent, track iterations |
| **QA Analyst Agent** | Write fuzz tests, generate corpus, run fuzz |

---

## Standards Reference

**MANDATORY:** Load testing-fuzz.md standards via WebFetch.

<fetch_required>
https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/golang/testing-fuzz.md
</fetch_required>

---

## Step 1: Validate Input

```text
REQUIRED INPUT:
- unit_id: [task/subtask being tested]
- implementation_files: [files from Gate 0]
- language: [go only for native fuzz]

OPTIONAL INPUT:
- gate3_handoff: [full Gate 3 output]

if any REQUIRED input is missing:
  → STOP and report: "Missing required input: [field]"

if language != "go":
  → STOP and report: "Native fuzz testing only supported for Go (Go 1.18+)"
```

## Step 2: Dispatch QA Analyst Agent (Fuzz Mode)

```text
Task tool:
  subagent_type: "ring:qa-analyst"
  model: "opus"
  prompt: |
    **MODE:** FUZZ TESTING (Gate 4)

    **Standards:** Load testing-fuzz.md

    **Input:**
    - Unit ID: {unit_id}
    - Implementation Files: {implementation_files}
    - Language: {language}

    **Requirements:**
    1. Create fuzz functions (FuzzXxx naming)
    2. Add seed corpus (minimum 5 entries per function)
    3. Run fuzz tests for 30 seconds
    4. Report any crashes found

    **Output Sections Required:**
    - ## Fuzz Testing Summary
    - ## Corpus Report
    - ## Handoff to Next Gate
```

## Step 3: Evaluate Results

```text
Parse agent output:

if "Status: PASS" in output:
  → Gate 4 PASSED
  → Return success with metrics

if "Status: FAIL" in output:
  → Dispatch fix to implementation agent
  → Re-run fuzz tests (max 3 iterations)
  → If still failing: ESCALATE to user
```

## Step 4: Generate Output

```text
## Fuzz Testing Summary
**Status:** {PASS|FAIL}
**Fuzz Functions:** {count}
**Corpus Entries:** {count}
**Crashes Found:** {count}

## Corpus Report
| Function | Entries | Crashes |
|----------|---------|---------|
| {function_name} | {count} | {count} |

## Handoff to Next Gate
- Ready for Gate 5 (Property Testing): {YES|NO}
- Iterations: {count}
```

---

## Anti-Rationalization Table

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "Unit tests cover edge cases" | You can't test what you don't think of. Fuzz finds unknowns. | **Write fuzz tests** |
| "Code is too simple for fuzz" | Simple code can still crash on malformed input. | **Write fuzz tests** |
| "Fuzz testing is slow" | 30 seconds per function. Crashes in production are slower. | **Write fuzz tests** |
| "We validate input anyway" | Validation can have bugs too. Fuzz tests the validators. | **Write fuzz tests** |

---

---

# Gate 5: Property-Based Testing


# Dev Property Testing (Gate 5)

## Overview

Ensure domain logic has **property-based tests** to verify invariants hold for all randomly generated inputs.

**Core principle:** Property tests verify universal truths about your domain. If "balance is never negative" is a rule, test it with thousands of random inputs.

<block_condition>
- No property functions = FAIL
- Any counterexample found = FAIL (fix and re-run)
- No quick.Check usage = FAIL
</block_condition>

## CRITICAL: Role Clarification

**This skill ORCHESTRATES. QA Analyst Agent (property mode) EXECUTES.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Gather requirements, dispatch agent, track iterations |
| **QA Analyst Agent** | Write property tests, run quick.Check, report counterexamples |

---

## Standards Reference

**MANDATORY:** Load testing-property.md standards via WebFetch.

<fetch_required>
https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/golang/testing-property.md
</fetch_required>

---

## Step 1: Validate Input

```text
REQUIRED INPUT:
- unit_id: [task/subtask being tested]
- implementation_files: [files from Gate 0]
- language: [go]

OPTIONAL INPUT:
- domain_invariants: [list of invariants to verify]
- gate4_handoff: [full Gate 4 output]

if any REQUIRED input is missing:
  → STOP and report: "Missing required input: [field]"
```

## Step 2: Dispatch QA Analyst Agent (Property Mode)

```text
Task tool:
  subagent_type: "ring:qa-analyst"
  model: "opus"
  prompt: |
    **MODE:** PROPERTY-BASED TESTING (Gate 5)

    **Standards:** Load testing-property.md

    **Input:**
    - Unit ID: {unit_id}
    - Implementation Files: {implementation_files}
    - Language: {language}
    - Domain Invariants: {domain_invariants}

    **Requirements:**
    1. Identify domain invariants from code
    2. Create property functions (TestProperty_{Subject}_{Property} naming)
    3. Use testing/quick.Check for verification
    4. Report any counterexamples found

    **Output Sections Required:**
    - ## Property Testing Summary
    - ## Properties Report
    - ## Handoff to Next Gate
```

## Step 3: Evaluate Results

```text
Parse agent output:

if "Status: PASS" in output:
  → Gate 5 PASSED
  → Return success with metrics

if "Status: FAIL" in output:
  → Dispatch fix to implementation agent
  → Re-run property tests (max 3 iterations)
  → If still failing: ESCALATE to user
```

## Step 4: Generate Output

```text
## Property Testing Summary
**Status:** {PASS|FAIL}
**Properties Tested:** {count}
**Properties Passed:** {count}
**Counterexamples Found:** {count}

## Properties Report
| Property | Subject | Status |
|----------|---------|--------|
| {property_name} | {subject} | {PASS|FAIL} |

## Handoff to Next Gate
- Ready for Gate 6 (Integration Testing): {YES|NO}
- Iterations: {count}
```

---

## Common Properties to Test

| Domain | Example Properties |
|--------|-------------------|
| Money/Currency | Amount never negative, currency always valid, addition commutative |
| User/Account | Email always valid format, password meets policy, status transitions valid |
| Order/Transaction | Total equals sum of items, quantity always positive, state machine valid |
| Date/Time | Start before end, duration always positive, timezone valid |

---

## Anti-Rationalization Table

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "Unit tests verify logic" | Unit tests verify SPECIFIC cases. Properties verify ALL cases. | **Write property tests** |
| "No domain invariants" | Every domain has rules. "ID is unique", "amount > 0", etc. | **Identify and test invariants** |
| "Too abstract" | Properties are concrete: "user.age >= 0 for all users". | **Write property tests** |
| "quick.Check is slow" | Milliseconds to find bugs that would take hours to discover. | **Write property tests** |

---

---

# Gate 6: Integration Testing


# Dev Integration Testing (Gate 6)

## Overview

Ensure every integration scenario has at least one **integration test** proving real external dependencies work correctly. Use testcontainers for all external services.

**Core principle:** Unit tests mock dependencies, integration tests verify real behavior. Both are required.

<block_condition>
- Any integration scenario without test = FAIL
- Any test using production services = FAIL
- Any test with hardcoded ports = FAIL
- Any flaky test (fails on retry) = FAIL
</block_condition>

## CRITICAL: Role Clarification

**This skill ORCHESTRATES. QA Analyst Agent (integration mode) EXECUTES.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Gather scenarios, check if needed, dispatch agent, validate output |
| **QA Analyst Agent** | Write tests, run coverage, verify quality gates |

---

## Step 0: Detect External Dependencies (Auto-Detection)

**MANDATORY:** When `external_dependencies` is empty or not provided, scan the codebase to detect them automatically before validation.

```text
if external_dependencies is empty or not provided:

  detected_dependencies = []

  1. Scan docker-compose.yml / docker-compose.yaml for service images:
     - Grep tool: pattern "postgres" in docker-compose* files → add "postgres"
     - Grep tool: pattern "mongo" in docker-compose* files → add "mongodb"
     - Grep tool: pattern "valkey" in docker-compose* files → add "valkey"
     - Grep tool: pattern "redis" in docker-compose* files → add "redis"
     - Grep tool: pattern "rabbitmq" in docker-compose* files → add "rabbitmq"

  2. Scan dependency manifests:
     if language == "go":
       - Grep tool: pattern "github.com/lib/pq" in go.mod → add "postgres"
       - Grep tool: pattern "github.com/jackc/pgx" in go.mod → add "postgres"
       - Grep tool: pattern "go.mongodb.org/mongo-driver" in go.mod → add "mongodb"
       - Grep tool: pattern "github.com/redis/go-redis" in go.mod → add "redis"
       - Grep tool: pattern "github.com/valkey-io/valkey-go" in go.mod → add "valkey"
       - Grep tool: pattern "github.com/rabbitmq/amqp091-go" in go.mod → add "rabbitmq"

     if language == "typescript":
       - Grep tool: pattern "\"pg\"" in package.json → add "postgres"
       - Grep tool: pattern "@prisma/client" in package.json → add "postgres"
       - Grep tool: pattern "\"mongodb\"" in package.json → add "mongodb"
       - Grep tool: pattern "\"mongoose\"" in package.json → add "mongodb"
       - Grep tool: pattern "\"redis\"" in package.json → add "redis"
       - Grep tool: pattern "\"ioredis\"" in package.json → add "redis"
       - Grep tool: pattern "@valkey" in package.json → add "valkey"
       - Grep tool: pattern "\"amqplib\"" in package.json → add "rabbitmq"
       - Grep tool: pattern "amqp-connection-manager" in package.json → add "rabbitmq"

  3. Deduplicate detected_dependencies
  4. Set external_dependencies = detected_dependencies

  Log: "Auto-detected external dependencies: [detected_dependencies]"
```

<auto_detect_reason>
PM team task files often omit external_dependencies. If the codebase uses postgres, mongodb, valkey, or rabbitmq, these are external dependencies that MUST have integration tests. Auto-detection prevents silent skips.
</auto_detect_reason>

---

## Step 1: Validate Input

```text
REQUIRED INPUT (from ring:dev-cycle orchestrator):
<verify_before_proceed>
- unit_id exists
- language is valid (go|typescript)
</verify_before_proceed>

OPTIONAL INPUT (determines if Gate 6 runs or skips):
- integration_scenarios: [list of scenarios] - if provided and non-empty, Gate 6 runs
- external_dependencies: [list of deps] (from input OR auto-detected in Step 0) - if non-empty, Gate 6 runs
- gate3_handoff: [full Gate 3 output]
- implementation_files: [files from Gate 0]

EXECUTION LOGIC:
1. if any REQUIRED input is missing:
   -> STOP and report: "Missing required input: [field]"
   -> Return to orchestrator with error

2. if integration_scenarios is empty AND external_dependencies is empty (AFTER auto-detection in Step 0):
   -> Gate 6 SKIP (document reason: "No integration scenarios or external dependencies found after codebase scan")
   -> Return skip result with status: "skipped"

3. Otherwise:
   -> Gate 6 REQUIRED - proceed to Step 2
```

## Step 2: Check If Integration Tests Needed

**Decision Tree:**

```text
1. Task has external_dependencies list (from input or auto-detected)?
   |
   +-- YES -> Gate 6 REQUIRED
   |
   +-- NO -> Continue to #2

2. Task has integration_scenarios?
   |
   +-- YES -> Gate 6 REQUIRED
   |
   +-- NO -> Continue to #3

3. Task acceptance criteria mention "integration", "database", "queue"?
   |
   +-- YES -> Gate 6 REQUIRED
   |
   +-- NO -> Gate 6 SKIP (with reason)
```

**If SKIP:**
```text
Return:
  status: SKIP
  skip_reason: "No external dependencies (after codebase scan) or integration scenarios identified"
  ready_for_gate7: YES
```

## Step 3: Initialize Testing State

```text
integration_state = {
  unit_id: [from input],
  scenarios: [from integration_scenarios or derived from external_dependencies],
  dependencies: [from external_dependencies],
  verdict: null,
  iterations: 0,
  max_iterations: 3,
  tests_passed: 0,
  tests_failed: 0,
  flaky_detected: 0
}
```

## Step 4: Dispatch QA Analyst Agent (Integration Mode)

<dispatch_required agent="ring:qa-analyst">
Write integration tests for all scenarios using testcontainers.
</dispatch_required>

```yaml
Task:
  subagent_type: "ring:qa-analyst"
  description: "Integration testing for [unit_id]"
  prompt: |
    **test_mode: integration**

    ## Input Context
    - **Unit ID:** [unit_id]
    - **Language:** [language]

    ## Integration Scenarios to Test
    [list integration_scenarios with IS-1, IS-2, etc.]

    ## External Dependencies
    [list external_dependencies with container requirements]

    ## Standards Reference
    WebFetch: https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/golang/testing-integration.md

    Focus on: All sections, especially INT-5 (Build Tags), INT-6 (Testcontainers), INT-7 (No t.Parallel())

    ## Requirements

    ### File Naming
    - Pattern: `*_integration_test.go`
    - Build tag: `//go:build integration` (MANDATORY at top of file)

    ### Function Naming
    - Pattern: `TestIntegration_{Component}_{Scenario}`
    - Example: `TestIntegration_UserRepository_Create`

    ### Container Usage
    - Use testcontainers for ALL external dependencies
    - Versions MUST match infra/docker-compose
    - Use t.Cleanup() for container termination

    ### Quality Rules
    - No t.Parallel() - integration tests run sequentially
    - No hardcoded ports - use dynamic ports from containers
    - No production services - all deps containerized
    - Each scenario MUST have at least one test

    ## Required Output Format

    ### Test Files Created
    | File | Tests | Lines |
    |------|-------|-------|
    | [path] | [count] | +N |

    ### Scenario Coverage
    | IS ID | Scenario | Test File | Test Function | Status |
    |-------|----------|-----------|---------------|--------|
    | IS-1 | [scenario text] | [file] | [function] | PASS/FAIL |
    | IS-2 | [scenario text] | [file] | [function] | PASS/FAIL |

    ### Quality Gate Results
    | Check | Status | Evidence |
    |-------|--------|----------|
    | Build tags present | PASS/FAIL | [file count] |
    | No hardcoded ports | PASS/FAIL | [grep result] |
    | Testcontainers used | PASS/FAIL | [imports] |
    | No t.Parallel() | PASS/FAIL | [grep result] |
    | Cleanup present | PASS/FAIL | [t.Cleanup count] |

    ### VERDICT
    **Tests:** [X passed, Y failed]
    **Quality Gate:** PASS / FAIL
    **VERDICT:** PASS / FAIL

    If FAIL:
    - **Gap Analysis:** [what needs more tests or fixes]
    - **Files needing attention:** [list with issues]
```

## Step 5: Parse QA Analyst Output

```text
Parse agent output:

1. Extract scenario coverage from Scenario Coverage table
2. Extract quality gate results
3. Extract verdict

integration_state.tests_passed = [count from output]
integration_state.tests_failed = [count from output]

if verdict == "PASS" and quality_gate == "PASS":
  -> integration_state.verdict = "PASS"
  -> Proceed to Step 7 (Success)

if verdict == "FAIL" or quality_gate == "FAIL":
  -> integration_state.verdict = "FAIL"
  -> integration_state.iterations += 1
  -> if iterations >= max_iterations: Go to Step 8 (Escalate)
  -> Go to Step 6 (Dispatch Fix)
```

## Step 6: Dispatch Fix to Implementation Agent

**Quality gate failed or tests failing -> Return to implementation agent**

```yaml
Task:
  subagent_type: "[implementation_agent from Gate 0]"
  description: "Fix integration test issues for [unit_id]"
  prompt: |
    Integration Test Issues - Fix Required

    ## Current Status
    - **Tests Passed:** [tests_passed]
    - **Tests Failed:** [tests_failed]
    - **Quality Gate:** FAIL
    - **Iteration:** [iterations] of [max_iterations]

    ## Issues Found (from QA)
    [paste gap analysis from QA output]

    ## Files Needing Attention
    [paste files list from QA output]

    ## Requirements
    1. Fix the identified issues
    2. Ensure all containers use testcontainers
    3. Remove any t.Parallel() from integration tests
    4. Add missing t.Cleanup() calls
    5. Replace hardcoded ports with dynamic ports

    ## Required Output
    - Issues fixed: [list]
    - Files modified: [list]
```

After fix -> Go back to Step 4 (Re-dispatch QA Analyst)

## Step 7: Prepare Success Output

```text
Generate skill output:

## Integration Testing Summary
**Status:** PASS
**Unit ID:** [unit_id]
**Iterations:** [integration_state.iterations]

## Scenario Coverage
| IS ID | Scenario | Test | Status |
|-------|----------|------|--------|
[from integration_state]

**Scenarios Covered:** [X]/[Y] (100%)

## Quality Gate Results
| Check | Status |
|-------|--------|
| Build tags | PASS |
| No hardcoded ports | PASS |
| Testcontainers | PASS |
| No t.Parallel() | PASS |
| Cleanup present | PASS |
| No flaky tests | PASS |

## Handoff to Next Gate
- Integration testing status: COMPLETE
- Tests passed: [tests_passed]
- Tests failed: 0
- Flaky tests: 0
- Ready for Gate 7 (Chaos Testing): YES
```

## Step 8: Escalate - Max Iterations Reached

```text
Generate skill output:

## Integration Testing Summary
**Status:** FAIL
**Unit ID:** [unit_id]
**Iterations:** [max_iterations] (MAX REACHED)

## Gap Analysis
[from last QA output]

## Files Still Needing Fixes
[from last QA output]

## Handoff to Next Gate
- Integration testing status: FAILED
- Ready for Gate 4: NO
- **Action Required:** User must manually fix integration tests

ESCALATION: Max iterations (3) reached. Integration tests still failing.
User intervention required.
```

---

## Pressure Resistance

See [shared-patterns/shared-pressure-resistance.md](../shared-patterns/shared-pressure-resistance.md) for universal pressure scenarios.

| User Says | Your Response |
|-----------|---------------|
| "Unit tests cover this" | "Unit tests mock dependencies. Integration tests verify real behavior. Both required." |
| "Testcontainers is too slow" | "Correctness > speed. Real dependencies catch real bugs." |
| "CI doesn't have Docker" | "Docker is baseline infrastructure. Fix CI before skipping integration tests." |
| "Skip integration, deadline" | "Integration bugs cost 10x more in production. Testing is non-negotiable." |

---

## Anti-Rationalization Table

See [shared-patterns/shared-anti-rationalization.md](../shared-patterns/shared-anti-rationalization.md) for universal anti-rationalizations.

### Gate 6-Specific Anti-Rationalizations

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "Database already tested in unit tests" | Unit tests use mocks, not real DB | **Write integration tests** |
| "Testcontainers setup is complex" | Complexity is one-time. Bugs are recurring. | **Use testcontainers** |
| "Integration tests are flaky" | Flaky = poorly written. Fix isolation. | **Fix the tests** |
| "No external dependencies" | Check task requirements. Often implicit. | **Verify with decision tree** |
| "t.Parallel() makes CI faster" | Faster but flaky. Flaky = worthless. | **Remove t.Parallel()** |
| "Hardcoded port works locally" | Fails in CI when port is taken. | **Use dynamic ports** |
| "Production DB is more realistic" | Production DB is dangerous and unreliable for tests. | **Use testcontainers** |

---

## Execution Report Format

```markdown
## Integration Testing Summary
**Status:** [PASS|FAIL|SKIP]
**Unit ID:** [unit_id]
**Duration:** [Xm Ys]
**Iterations:** [N]

## Scenario Coverage
| IS ID | Scenario | Test | Status |
|-------|----------|------|--------|
| IS-1 | [text] | [test] | PASS/FAIL |

**Scenarios Covered:** [X/Y]

## Quality Gate Results
| Check | Status |
|-------|--------|
| Build tags | PASS/FAIL |
| No hardcoded ports | PASS/FAIL |
| Testcontainers | PASS/FAIL |
| No t.Parallel() | PASS/FAIL |
| Cleanup present | PASS/FAIL |
| No flaky tests | PASS/FAIL |

## Handoff to Next Gate
- Integration testing status: [COMPLETE|FAILED|SKIPPED]
- Ready for Gate 7: [YES|NO]
```

---

## Skip Conditions (Documented)

**When Gate 6 can be skipped (MUST document reason):**

| Condition | Skip Reason |
|-----------|-------------|
| No external dependencies | "Task has no database, API, or queue interactions" |
| Pure business logic | "Task is pure function/logic with no I/O" |
| Library/utility code | "Task is internal utility with no external calls" |
| Already covered | "Integration tests exist and pass (verified)" |

**When Gate 6 CANNOT be skipped:**

| Condition | Why Required |
|-----------|--------------|
| Task touches database | Database queries need real verification |
| Task calls external APIs | HTTP behavior varies from mocks |
| Task uses message queues | Pub/sub requires real broker testing |
| Task has transactions | ACID guarantees need real DB |
| Task has migrations | Schema changes need integration verification |

---

# Gate 7: Chaos Testing


# Dev Chaos Testing (Gate 7)

## Overview

Ensure code handles **failure conditions gracefully** by injecting faults using Toxiproxy. Verify connection loss, latency, and network partitions don't cause crashes.

**Core principle:** All infrastructure fails. Chaos testing ensures your code handles it gracefully.

<block_condition>
- No chaos tests = FAIL
- Any dependency without failure test = FAIL
- Recovery not verified = FAIL
- System crashes on failure = FAIL
</block_condition>

## CRITICAL: Role Clarification

**This skill ORCHESTRATES. QA Analyst Agent (chaos mode) EXECUTES.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Gather requirements, dispatch agent, track iterations |
| **QA Analyst Agent** | Write chaos tests, setup Toxiproxy, verify recovery |

---

## Standards Reference

**MANDATORY:** Load testing-chaos.md standards via WebFetch.

<fetch_required>
https://raw.githubusercontent.com/LerianStudio/ring/main/dev-team/docs/standards/golang/testing-chaos.md
</fetch_required>

---

## Step 0: Detect External Dependencies (Auto-Detection)

**MANDATORY:** When `external_dependencies` is empty or not provided, scan the codebase to detect them automatically before validation.

```text
if external_dependencies is empty or not provided:

  detected_dependencies = []

  1. Scan docker-compose.yml / docker-compose.yaml for service images:
     - Grep tool: pattern "postgres" in docker-compose* files → add "postgres"
     - Grep tool: pattern "mongo" in docker-compose* files → add "mongodb"
     - Grep tool: pattern "valkey" in docker-compose* files → add "valkey"
     - Grep tool: pattern "redis" in docker-compose* files → add "redis"
     - Grep tool: pattern "rabbitmq" in docker-compose* files → add "rabbitmq"

  2. Scan dependency manifests:
     if language == "go":
       - Grep tool: pattern "github.com/lib/pq" in go.mod → add "postgres"
       - Grep tool: pattern "github.com/jackc/pgx" in go.mod → add "postgres"
       - Grep tool: pattern "go.mongodb.org/mongo-driver" in go.mod → add "mongodb"
       - Grep tool: pattern "github.com/redis/go-redis" in go.mod → add "redis"
       - Grep tool: pattern "github.com/valkey-io/valkey-go" in go.mod → add "valkey"
       - Grep tool: pattern "github.com/rabbitmq/amqp091-go" in go.mod → add "rabbitmq"

     if language == "typescript":
       - Grep tool: pattern "\"pg\"" in package.json → add "postgres"
       - Grep tool: pattern "@prisma/client" in package.json → add "postgres"
       - Grep tool: pattern "\"mongodb\"" in package.json → add "mongodb"
       - Grep tool: pattern "\"mongoose\"" in package.json → add "mongodb"
       - Grep tool: pattern "\"redis\"" in package.json → add "redis"
       - Grep tool: pattern "\"ioredis\"" in package.json → add "redis"
       - Grep tool: pattern "@valkey" in package.json → add "valkey"
       - Grep tool: pattern "\"amqplib\"" in package.json → add "rabbitmq"
       - Grep tool: pattern "amqp-connection-manager" in package.json → add "rabbitmq"

  3. Deduplicate detected_dependencies
  4. Set external_dependencies = detected_dependencies

  Log: "Auto-detected external dependencies: [detected_dependencies]"
```

<auto_detect_reason>
PM team task files often omit external_dependencies. If the codebase uses postgres, mongodb, valkey, or rabbitmq, these are external dependencies that MUST have chaos tests. Auto-detection prevents silent skips.
</auto_detect_reason>

---

## Step 1: Validate Input

```text
REQUIRED INPUT:
- unit_id: [task/subtask being tested]
- external_dependencies: [postgres, mongodb, valkey, redis, rabbitmq, etc.] (from input OR auto-detected in Step 0)
- language: [go|typescript]

OPTIONAL INPUT:
- gate6_handoff: [full Gate 6 output]

if any REQUIRED input is missing:
  → STOP and report: "Missing required input: [field]"

if external_dependencies is empty (AFTER auto-detection in Step 0):
  → STOP and report: "No external dependencies found after codebase scan - chaos testing requires dependencies"
```

## Step 2: Dispatch QA Analyst Agent (Chaos Mode)

```text
Task tool:
  subagent_type: "ring:qa-analyst"
  model: "opus"
  prompt: |
    **MODE:** CHAOS TESTING (Gate 7)

    **Standards:** Load testing-chaos.md

    **Input:**
    - Unit ID: {unit_id}
    - External Dependencies: {external_dependencies}
    - Language: {language}

    **Requirements:**
    1. Setup Toxiproxy infrastructure in tests/utils/chaos/
    2. Create chaos tests (TestIntegration_Chaos_{Component}_{Scenario} naming)
    3. Use dual-gate pattern (CHAOS=1 env + testing.Short())
    4. Test failure scenarios: Connection Loss, High Latency, Network Partition
    5. Verify 5-phase structure: Normal → Inject → Verify → Restore → Recovery

    **Output Sections Required:**
    - ## Chaos Testing Summary
    - ## Failure Scenarios
    - ## Handoff to Next Gate
```

## Step 3: Evaluate Results

```text
Parse agent output:

if "Status: PASS" in output:
  → Gate 7 PASSED
  → Return success with metrics

if "Status: FAIL" in output:
  → Dispatch fix to implementation agent
  → Re-run chaos tests (max 3 iterations)
  → If still failing: ESCALATE to user
```

## Step 4: Generate Output

```text
## Chaos Testing Summary
**Status:** {PASS|FAIL}
**Dependencies Tested:** {count}
**Scenarios Tested:** {count}
**Recovery Verified:** {Yes|No}

## Failure Scenarios
| Component | Scenario | Status | Recovery |
|-----------|----------|--------|----------|
| {component} | {scenario} | {PASS|FAIL} | {Yes|No} |

## Handoff to Next Gate
- Ready for Gate 8 (Code Review): {YES|NO}
- Iterations: {count}
```

---

## Failure Scenarios by Dependency

| Dependency | Required Scenarios |
|------------|-------------------|
| PostgreSQL | Connection Loss, High Latency, Network Partition |
| MongoDB | Connection Loss, High Latency, Network Partition |
| Valkey | Connection Loss, High Latency, Timeout |
| Redis | Connection Loss, High Latency, Timeout |
| RabbitMQ | Connection Loss, Network Partition, Slow Consumer |
| HTTP APIs | Timeout, 5xx Errors, Connection Refused |

---

## Anti-Rationalization Table

| Rationalization | Why It's WRONG | Required Action |
|-----------------|----------------|-----------------|
| "Infrastructure is reliable" | AWS, GCP, Azure all have outages. Your code must handle them. | **Write chaos tests** |
| "Integration tests cover failures" | Integration tests verify happy path. Chaos tests verify failure handling. | **Write chaos tests** |
| "Toxiproxy is complex" | One container. 20 minutes setup. Prevents production incidents. | **Write chaos tests** |
| "We have monitoring" | Monitoring detects problems. Chaos testing prevents them. | **Write chaos tests** |
| "Circuit breakers handle it" | Circuit breakers need testing too. Chaos tests verify they work. | **Write chaos tests** |

---
