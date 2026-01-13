---
name: "ring:dev-cycle"
description: |
  Orchestrates the 6-gate development workflow for implementing tasks.
  Manages state, dispatches specialist agents, and enforces gate requirements.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: workflow
  trigger:
    - Tasks file provided (from pre-dev or dev-refactor)
    - User wants to implement development task
    - /dev-cycle command invoked
---

# Dev Cycle Skill

## Overview

Orchestrates the 6-gate development workflow:
- Gate 0: Implementation (TDD with specialist agent)
- Gate 1: DevOps (containerization, CI/CD)
- Gate 2: SRE (observability validation)
- Gate 3: Testing (coverage threshold)
- Gate 4: Review (3 parallel reviewers)
- Gate 5: Validation (user approval)

## State Management

State file: `docs/dev-cycle/current-cycle.json`

```json
{
  "version": "1.0",
  "state_path": "docs/dev-cycle/current-cycle.json",
  "current_task_index": 0,
  "current_gate": 0,
  "status": "running",
  "tasks": [
    {
      "id": "TASK-001",
      "title": "Task title",
      "status": "running",
      "gate_progress": {
        "implementation": {"status": "pending"},
        "devops": {"status": "pending"},
        "sre": {"status": "pending"},
        "testing": {"status": "pending"},
        "review": {"status": "pending"},
        "validation": {"status": "pending"}
      }
    }
  ]
}
```

## Workflow

### Step 0: Verify PROJECT_RULES.md

Check if `docs/PROJECT_RULES.md` exists:

```yaml
If docs/PROJECT_RULES.md exists:
  - Load and use for project standards
  - Continue to Step 1

If docs/PROJECT_RULES.md does NOT exist:
  - Check if {OPENCODE_CONFIG}/templates/PROJECT_RULES.md exists
  - Present option to user:
    "PROJECT_RULES.md not found. Would you like me to create one from the Ring template?"
    - YES: Copy {OPENCODE_CONFIG}/templates/PROJECT_RULES.md to docs/PROJECT_RULES.md
           Ask user to review and customize for their project
           Continue to Step 1
    - NO:  Ask if this is a legacy project that needs analysis
           If legacy: Generate PROJECT_RULES.md from codebase analysis
           If greenfield: Recommend running /pre-dev-full first
```

### Step 1: Initialize or Resume State

**New Cycle:**
```yaml
Read: tasks.md from input
Create: docs/dev-cycle/current-cycle.json
Set: current_task_index = 0, current_gate = 0
```

**Resume:**
```yaml
Read: docs/dev-cycle/current-cycle.json
Resume: from last checkpoint (current_task_index, current_gate)
```

### Step 2: Load Task

```yaml
task = state.tasks[current_task_index]
Extract: acceptance_criteria, technical_context, language
```

### Step 3: Execute Gate 0 (Implementation)

```yaml
Skill: ring:dev-implementation
Input:
  unit_id: task.id
  requirements: task.acceptance_criteria
  language: detected_language
  service_type: detected_type
```

**On completion:** Update state, write file, proceed to Gate 1.

### Step 4: Execute Gate 1 (DevOps)

```yaml
Skill: ring:dev-devops
Input:
  unit_id: task.id
  language: detected_language
  implementation_files: from Gate 0
```

### Step 5: Execute Gate 2 (SRE)

```yaml
Skill: ring:dev-sre
Input:
  unit_id: task.id
  language: detected_language
  implementation_files: from Gate 0
  implementation_agent: from Gate 0
```

### Step 6: Execute Gate 3 (Testing)

```yaml
Skill: ring:dev-testing
Input:
  unit_id: task.id
  acceptance_criteria: task.acceptance_criteria
  implementation_files: from Gate 0
  coverage_threshold: 85
```

### Step 7: Execute Gate 4 (Review)

```yaml
Skill: ring:requesting-code-review
# Dispatches 3 parallel reviewers:
# - ring:code-reviewer
# - ring:security-reviewer
# - ring:business-logic-reviewer
```

### Step 8: Execute Gate 5 (Validation)

```yaml
Skill: ring:dev-validation
# Presents validation checklist
# Requires explicit APPROVED/REJECTED
```

### Step 9: Task Complete or Next Task

**If APPROVED:**
- Mark task complete
- If more tasks: increment current_task_index, go to Step 2
- If no more tasks: proceed to feedback loop

**If REJECTED:**
- Create remediation task
- Return to Gate 0

### Step 10: Feedback Loop

```yaml
Skill: ring:dev-feedback-loop
# Captures metrics and learnings
```

## State Persistence Rule

**After EVERY gate:**
1. Update state object
2. Write to file (MANDATORY)
3. Verify persistence

```yaml
Write tool:
  file_path: docs/dev-cycle/current-cycle.json
  content: [full JSON state]
```

## Gate Transition Checkpoints

| After | Update | Write File |
|-------|--------|------------|
| Gate 0.1 (TDD-RED) | tdd_red.status | YES |
| Gate 0.2 (TDD-GREEN) | implementation.status | YES |
| Gate 1 (DevOps) | devops.status | YES |
| Gate 2 (SRE) | sre.status | YES |
| Gate 3 (Testing) | testing.status | YES |
| Gate 4 (Review) | review.status | YES |
| Gate 5 (Validation) | validation.status | YES |

## Pressure Resistance

| User Says | Response |
|-----------|----------|
| "Skip Gate X" | "All gates are MANDATORY. Cannot skip." |
| "Just implement, skip tests" | "TDD is REQUIRED. Gate 0 includes tests." |
| "Review takes too long" | "Review is part of quality. Dispatching reviewers now." |

## Anti-Rationalization

| Rationalization | Why Wrong | Required Action |
|-----------------|-----------|-----------------|
| "I'll save state at the end" | Crash loses ALL progress | **Save after EACH gate** |
| "Gate passed informally" | Informal ≠ verified | **Execute gate skill** |
| "User approved verbally" | Verbal ≠ documented | **Require explicit APPROVED** |
