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
    after: [dev-implementation, dev-devops, dev-sre]
    before: [requesting-code-review]
---

# Dev Testing (Gate 3)

## Overview

Ensure every acceptance criterion has at least one **unit test** proving it works. Follow TDD methodology: RED (failing test) -> GREEN (implementation) -> REFACTOR.

**Core principle:** Untested acceptance criteria are unverified claims. Each criterion MUST map to at least one executable unit test.

**Coverage threshold:** 85% minimum (Ring standard). PROJECT_RULES.md can raise, not lower.

## Role Clarification

**This skill ORCHESTRATES. QA Analyst Agent EXECUTES.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Gather requirements, dispatch agent, track iterations |
| **QA Analyst Agent** | Write tests, run coverage, report results |

## Required Input

- `unit_id`: Task/subtask being tested
- `acceptance_criteria`: List of ACs to test
- `implementation_files`: Files from Gate 0
- `language`: go / typescript / python

## Dispatch QA Analyst Agent

```yaml
Task:
  subagent_type: "qa-analyst"
  description: "Write unit tests for [unit_id]"
  prompt: |
    WRITE UNIT TESTS for All Acceptance Criteria

    ## Context
    - Unit ID: [unit_id]
    - Language: [language]
    - Coverage Threshold: [coverage_threshold]%

    ## Acceptance Criteria to Test
    [list with AC-1, AC-2, etc.]

    ## Implementation Files to Test
    [implementation_files]

    ## Requirements

    ### Test Coverage
    - Minimum: [coverage_threshold]% branch coverage
    - Every AC MUST have at least one test
    - Edge cases REQUIRED

    ### Test Naming
    - Go: Test{Unit}_{Method}_{Scenario}
    - TS: describe('{Unit}', () => { it('should {scenario}', ...) })

    ### Edge Cases Required per AC Type
    | AC Type | Required Edge Cases | Minimum |
    |---------|---------------------|---------|
    | Input validation | null, empty, boundary | 3+ |
    | CRUD operations | not found, duplicate | 3+ |
    | Business logic | zero, negative, overflow | 3+ |

    ## Required Output

    ### Coverage Report
    | Package/File | Coverage |
    |--------------|----------|
    | **TOTAL** | **[X%]** |

    ### Traceability Matrix
    | AC ID | Criterion | Test File | Test Function | Status |
    |-------|-----------|-----------|---------------|--------|

    ### Quality Checks
    | Check | Status |
    |-------|--------|
    | No skipped tests | ✅/❌ |
    | Edge cases per AC | ✅/❌ |

    ### VERDICT
    **Coverage:** [X%] vs Threshold [Y%]
    **VERDICT:** PASS / FAIL
```

## Output Schema

```markdown
## Testing Summary
**Status:** PASS/FAIL
**Unit ID:** [unit_id]
**Iterations:** [N]

## Coverage Report
**Threshold:** [X%]
**Actual:** [Y%]
**Status:** ✅ PASS / ❌ FAIL

| Package/File | Coverage |
|--------------|----------|
| **TOTAL** | **[X%]** |

## Traceability Matrix
| AC ID | Criterion | Test | Status |
|-------|-----------|------|--------|

**Criteria Covered:** [X/Y]

## Quality Checks
| Check | Status |
|-------|--------|
| Coverage ≥ threshold | ✅/❌ |
| All ACs tested | ✅/❌ |
| No skipped tests | ✅/❌ |

## Handoff to Next Gate
- Testing status: COMPLETE
- Coverage: [X%]
- Ready for Gate 4 (Review): YES/NO
```

## Pressure Resistance

| User Says | Response |
|-----------|----------|
| "84% is close enough" | "85% is minimum threshold. 84% = FAIL. Adding more tests." |
| "Manual testing covers it" | "Gate 3 requires executable unit tests." |
| "Skip testing, deadline" | "Testing is MANDATORY. Untested code = unverified claims." |

## Anti-Rationalization

| Rationalization | Why Wrong | Required Action |
|-----------------|-----------|-----------------|
| "Tool shows 83% but real is 90%" | Tool output IS real | **Fix issue, re-measure** |
| "84.5% rounds to 85%" | Rounding not allowed | **Write more tests** |
| "Integration tests cover this" | Gate 3 = unit tests only | **Write unit tests** |

## Unit Test vs Integration Test

| Type | Characteristics | Gate 3? |
|------|----------------|---------|
| **Unit** | Mocks all external deps, tests single function | YES |
| **Integration** | Hits real database/API | NO |
