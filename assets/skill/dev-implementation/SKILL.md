---
name: "ring:dev-implementation"
description: |
  Gate 0 of the development cycle. Executes code implementation using the appropriate
  specialized agent based on task content and project language. Handles TDD workflow
  with RED-GREEN phases. Follows project standards defined in docs/PROJECT_RULES.md.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: development
  trigger:
    - Gate 0 of development cycle
    - Tasks loaded at initialization
    - Ready to write code
  sequence:
    before: [dev-devops]
---

# Code Implementation (Gate 0)

## Overview

This skill executes the implementation phase of the development cycle:
- Selects the appropriate specialized agent based on task content
- Applies project standards from docs/PROJECT_RULES.md
- Follows TDD methodology (RED -> GREEN -> REFACTOR)
- Documents implementation decisions

## Role Clarification

**This skill ORCHESTRATES. Agents IMPLEMENT.**

| Who | Responsibility |
|-----|----------------|
| **This Skill** | Select agent, prepare prompts, track state, validate outputs |
| **Implementation Agent** | Write tests, write code, follow standards |

## Required Input

- `unit_id`: Task/subtask being implemented
- `requirements`: Acceptance criteria or task description
- `language`: go / typescript / python
- `service_type`: api / worker / batch / cli / frontend / bff

## Agent Selection

| Language | Service Type | Agent |
|----------|--------------|-------|
| Go | API, Worker, Batch, CLI | `backend-engineer-golang` |
| TypeScript | API, Worker | `backend-engineer-typescript` |
| TypeScript | Frontend | `frontend-engineer` |
| React/CSS | Design, Styling | `frontend-designer` |

## Gate 0.1: TDD-RED (Write Failing Test)

```yaml
Task:
  subagent_type: "[selected_agent]"
  description: "TDD-RED: Write failing test for [unit_id]"
  prompt: |
    TDD-RED PHASE: Write a FAILING Test

    ## Requirements
    [requirements]

    ## Your Task
    1. Write a test that captures the expected behavior
    2. The test MUST FAIL (no implementation exists yet)
    3. Run the test and capture the FAILURE output

    ## Required Output
    ### Test File
    **Path:** [path/to/test_file]

    ### Failure Output (MANDATORY)
    ```
    [actual test failure output]
    ```

    HARD GATE: Without failure output, TDD-RED is NOT complete.
```

## Gate 0.2: TDD-GREEN (Implementation)

**PREREQUISITE:** TDD-RED completed with failure output

```yaml
Task:
  subagent_type: "[selected_agent]"
  description: "TDD-GREEN: Implement code to pass test"
  prompt: |
    TDD-GREEN PHASE: Make the Test PASS

    ## TDD-RED Results
    - Test File: [test_file]
    - Failure Output: [failure_output]

    ## Your Task
    1. Write MINIMAL code to make the test pass
    2. Follow ALL Ring Standards (logging, tracing, error handling)
    3. Run the test and capture the PASS output

    ## Required Output
    ### Implementation Files
    | File | Action | Lines |
    |------|--------|-------|

    ### Pass Output (MANDATORY)
    ```
    [actual test pass output]
    ```

    ### Standards Compliance
    - Structured Logging: ✅/❌
    - OpenTelemetry Spans: ✅/❌
    - Error Handling: ✅/❌
    - Context Propagation: ✅/❌
```

## Output Schema

```markdown
## Implementation Summary
**Status:** PASS/FAIL/PARTIAL
**Unit ID:** [unit_id]
**Agent:** [selected_agent]
**Commit:** [sha]

## TDD Results
| Phase | Status | Output |
|-------|--------|--------|
| RED | ✅/❌ | [summary] |
| GREEN | ✅/❌ | [summary] |

## Files Changed
| File | Action | Lines |
|------|--------|-------|
| [path] | Created/Modified | +/-N |

## Standards Compliance
- Structured Logging: ✅
- OpenTelemetry Spans: ✅
- Error Handling: ✅
- Context Propagation: ✅

## Handoff to Next Gate
- Implementation status: COMPLETE
- Code compiles: ✅
- Tests pass: ✅
- Standards met: ✅
- Ready for Gate 1 (DevOps): YES
```

## Pressure Resistance

| User Says | Response |
|-----------|----------|
| "Skip TDD, just implement" | "TDD is MANDATORY. Dispatching agent for RED phase." |
| "Code exists, just add tests" | "DELETE existing code. TDD requires test-first." |
| "Add observability later" | "Observability is part of implementation. Agent MUST add it now." |

## Anti-Rationalization

| Rationalization | Why Wrong | Required Action |
|-----------------|-----------|-----------------|
| "Test passes on first run" | Passing test ≠ TDD. Test MUST fail first. | **Rewrite test to fail first** |
| "Skip RED, go straight to GREEN" | RED proves test validity | **Execute RED phase first** |
| "I'll add observability later" | Later = never. Observability is part of GREEN. | **Add logging + tracing NOW** |
| "DEFERRED to later tasks" | DEFERRED = FAILED. Standards are not deferrable. | **Implement ALL standards NOW** |
