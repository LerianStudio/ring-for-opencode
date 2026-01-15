---
name: "ring:pre-dev-subtask-creation"
description: |
  Gate 8: Zero-context implementation steps - 2-5 minute atomic subtasks with
  complete code, exact commands, TDD pattern. Large Track only.
license: MIT
compatibility: opencode
metadata:
  trigger: |
    - Tasks passed Gate 7 validation
    - Need absolute implementation clarity
    - Creating work for engineers with zero codebase context
    - Large Track workflow (2+ day features)
  skip_when: |
    - Small Track workflow → execute tasks directly
    - Tasks simple enough without breakdown
    - Tasks not validated → complete Gate 7 first

sequence:
  after: [ring:pre-dev-task-breakdown]
  before: [ring:executing-plans, subagent-driven-development]
---

# Subtask Creation - Bite-Sized, Zero-Context Steps

## Overview

Write comprehensive implementation subtasks assuming the engineer has zero context for our codebase. Each subtask breaks down into 2-5 minute steps following RED-GREEN-REFACTOR. Provide step-by-step guidance to full task completion with exact commands and explicit verification, without pre-writing all code. **DRY. YAGNI. TDD. Frequent commits.**

**Save subtasks to:** `docs/pre-dev/{feature-name}/subtasks/T-[task-id]/ST-[task-id]-[number]-[description].md`

## Foundational Principle

**Every subtask must be completable by anyone with zero context about the system.**

Requiring context creates bottlenecks, onboarding friction, and integration failures.

**Subtasks answer**: Exactly what to change, where, and how to verify success, with precise steps and required code snippets.
**Subtasks never answer**: Why the system works this way (context is removed).

## Bite-Sized Step Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement the minimal code to make the test pass" - step
- "Run the tests and make sure they pass" - step
- "Commit" - step

## Subtask Document Structure

**Header (required):**

| Field | Content |
|-------|---------|
| Title | `# ST-[task-id]-[number]: [Subtask Name]` |
| Agent Note | `> **For Agents:** REQUIRED SUB-SKILL: Use ring:executing-plans` |
| Goal | One sentence describing what this builds |
| Prerequisites | Verification commands with expected output |
| Files | Create: `exact/path`, Modify: `exact/path:lines`, Test: `tests/path` |

**Step Structure (TDD Cycle):**

| Step | Content |
|------|---------|
| Step 1: Write failing test | Exact test cases and code snippets; full file only when creating a new file |
| Step 2: Run test to verify fail | Command + expected failure output |
| Step 3: Write minimal implementation | Exact file edits and code snippets; full file only when necessary |
| Step 4: Run test to verify pass | Command + expected success output |
| Step 5: Update exports (if needed) | Exact modification to index files |
| Step 6: Verify type checking | Command + expected output |
| Step 7: Commit | Exact git commands with message |
| Rollback | Exact commands to undo if issues |

## Explicit Rules

### ✅ DO Include in Subtasks
Exact file paths (absolute or from root), precise edit instructions with code snippets, full file contents only when creating a new file or when a full replacement is required, required imports for any provided snippets, step-by-step TDD cycle (numbered), verification commands (copy-pasteable), expected output (exact), rollback procedures (exact commands), prerequisites (what must exist first)

### ❌ NEVER Include in Subtasks
Placeholders: "...", "TODO", "implement here"; vague instructions: "update the service", "add validation"; assumptions: "assuming setup is done"; context requirements: "you need to understand X first"; incomplete guidance: "add the rest yourself"; missing imports: "import necessary packages"; undefined success: "make sure it works"; no verification: "test it manually"

## Rationalization Table

| Excuse | Reality |
|--------|---------|
| "The developer will figure out imports" | Imports are part of the change. Provide them for any snippet you include. |
| "TODO comments are fine for simple parts" | TODOs require decisions. Make them now. |
| "They'll know which service to update" | They won't. Specify the exact file path. |
| "The verification steps are obvious" | Obvious ≠ documented. Write exact commands. |
| "Rollback isn't needed for simple changes" | Simple changes fail too. Always provide rollback. |
| "This needs system understanding" | Then you haven't removed context. Simplify more. |
| "I'll provide the template, they fill it" | Templates are incomplete. Provide exact edits and snippets. |
| "The subtask description explains it" | Descriptions need interpretation. Give exact steps. |
| "They can look at similar code for reference" | That's context. Make subtask self-contained. |
| "This is too detailed, we're not that formal" | Detailed = parallelizable = faster. Be detailed. |
| "Steps are too small, feels like hand-holding" | Small steps = verifiable progress. Stay small. |

## Red Flags - STOP

If you catch yourself writing any of these in a subtask, **STOP and rewrite**:

- Code placeholders: `...`, `// TODO`, `// implement X here`
- Vague file references: "the user service", "the auth module"
- Assumption phrases: "assuming you have", "make sure you"
- Incomplete imports: "import required packages"
- Missing paths: Not specifying where files go
- Undefined verification: "test that it works"
- Steps longer than 5 minutes
- Context dependencies: "you need to understand X"
- No TDD cycle in implementation steps

**When you catch yourself**: Expand the subtask until it's completely self-contained.

## Gate 8 Validation Checklist

| Category | Requirements |
|----------|--------------|
| **Atomicity** | Each step 2-5 minutes; no system architecture understanding required; assignable to anyone |
| **Completeness** | All steps explicit; all file paths explicit; required code snippets and imports provided; full files only when needed; TDD cycle followed |
| **Verifiability** | Test commands copy-pasteable; expected output exact; commands run from project root |
| **Reversibility** | Rollback commands provided; rollback doesn't require system knowledge |

**Gate Result:** ✅ PASS → Ready for implementation | ⚠️ CONDITIONAL (add details) | ❌ FAIL (decompose further)

## Confidence Scoring

| Factor | Points | Criteria |
|--------|--------|----------|
| Step Atomicity | 0-30 | All 2-5 minutes: 30, Most sized right: 20, Too large/vague: 10 |
| Instruction Completeness | 0-30 | No missing steps or decisions: 30, Minor gaps: 15, Significant gaps: 5 |
| Context Independence | 0-25 | Anyone can execute: 25, Minor context: 15, Significant knowledge: 5 |
| TDD Coverage | 0-15 | All RED-GREEN-REFACTOR: 15, Most have tests: 10, Limited: 5 |

**Action:** 80+ autonomous | 50-79 present options | <50 ask about structure

## Execution Handoff

After creating subtasks, offer execution choice:

**"Subtasks complete. Two execution options:**
1. **Subagent-Driven** - Fresh subagent per subtask, review between, fast iteration → Use `subagent-driven-development`
2. **Parallel Session** - New session with ring:executing-plans, batch with checkpoints → Use `ring:executing-plans`

**Which approach?"**

## The Bottom Line

**If you wrote a subtask with "TODO" or "..." or "add necessary imports", delete it and rewrite with explicit steps and exact edits.**

Subtasks are not code dumps. Subtasks are complete, copy-pasteable steps that lead to a finished implementation following TDD.

- "Add validation" is not a step. [Exact file path + specific validation edits + test] is a step.
- "Update the service" is not a step. [Exact file path + exact edits + test] is a step.
- "Import necessary packages" is not a step. [Exact list of imports for your snippet] is a step.

Every subtask must be completable by someone who:
- Just joined the team yesterday
- Has never seen the codebase before
- Doesn't know the business domain
- Won't ask questions (you're unavailable)
- Follows TDD religiously

If they can't complete it with zero questions while following RED-GREEN-REFACTOR, **it's not atomic enough.**

**Remember: DRY. YAGNI. TDD. Frequent commits.**
