---
name: writing-plans
description: |
  Creates comprehensive implementation plans with exact file paths, complete code
  examples, and verification steps for engineers with zero codebase context.
license: MIT
compatibility: opencode
metadata:
  trigger: "Design phase complete, need executable task breakdown, creating work for others"
  skip_when: "Design not validated, requirements unclear, already have a plan"
  sequence_after: "brainstorming, pre-dev-trd-creation"
  sequence_before: "executing-plans"
  source: ring-default
---

# Writing Plans

## Related Skills

**Similar:** brainstorming

---

## Overview

This skill dispatches a specialized agent to write comprehensive implementation plans for engineers with zero codebase context.

**Announce at start:** "I'm using the writing-plans skill to create the implementation plan."

**Context:** This should be run in a dedicated worktree (created by brainstorming skill).

## The Process

**Step 1: Query Historical Precedent (MANDATORY)**

Before planning, query the artifact index for historical context:

```bash
# Keywords MUST be alphanumeric with spaces only (sanitize before use)
python3 default/lib/artifact-index/artifact_query.py --mode planning "keywords" --json
```

**Keyword Extraction (MANDATORY):**
1. Extract topic keywords from the planning request
2. Sanitize keywords: Keep only alphanumeric characters, hyphens, and spaces
3. Example: "Implement OAuth2 authentication!" -> keywords: "oauth2 authentication"

**Safety:** Never include shell metacharacters (`;`, `$`, `` ` ``, `|`, `&`) in keywords.

Results inform the plan:
- `successful_handoffs` -> Reference patterns that worked
- `failed_handoffs` -> WARN in plan, design to avoid
- `relevant_plans` -> Review for approach ideas

If `is_empty_index: true` -> Proceed without precedent (normal for new projects).

**Step 2: Dispatch Write-Plan Agent**

Dispatch via `Task(subagent_type: "write-plan")` with:
- Instructions to create bite-sized tasks (2-5 min each)
- Include exact file paths, complete code, verification steps
- **Include Historical Precedent section** in plan with query results
- Save to `docs/plans/YYYY-MM-DD-<feature-name>.md`

**Step 3: Validate Plan Against Failure Patterns**

After the plan is saved, validate it against known failures:

```bash
python3 default/lib/validate-plan-precedent.py docs/plans/YYYY-MM-DD-<feature>.md
```

**Interpretation:**
- `PASS` -> Plan is safe to execute
- `WARNING` -> Plan has >30% keyword overlap with past failures
  - Review the warnings in the output
  - Update plan to address the failure patterns
  - Re-run validation until PASS

**Note:** If artifact index is unavailable, validation passes automatically (nothing to check against).

**Step 4: Ask User About Execution**

Ask via `AskUserQuestion`: "Execute now?" Options:
1. Parallel session -> user opens new session with `executing-plans`
2. Save for later -> report location and end

## Why Use an Agent?

**Context preservation** (reading many files keeps supervisor clean) | **Model power** (Opus for comprehensive planning) | **Separation of concerns** (supervisor orchestrates, agent plans)

## What the Agent Does

**Query precedent** -> Explore codebase -> identify files -> break into bite-sized tasks (2-5 min) -> write complete code -> include exact commands -> add review checkpoints -> **include Historical Precedent section** -> verify Zero-Context Test -> save to `docs/plans/YYYY-MM-DD-<feature>.md` -> report back

## Requirements for Plans

Every plan: **Historical Precedent section** | Header (goal, architecture, tech stack) | Verification commands with expected output | Exact file paths (never "somewhere in src") | Complete code (never "add validation here") | Bite-sized steps with verification | Failure recovery | Review checkpoints | Zero-Context Test | **Recommended agents per task** | **Avoids known failure patterns**

## Agent Selection

| Task Type | Agent |
|-----------|-------|
| Backend API/services | `backend-engineer-{golang,typescript}` |
| Frontend | `frontend-engineer` |
| Infra/CI/CD | `devops-engineer` |
| Testing | `qa-analyst` |
| Reliability | `sre` |
| Fallback | `general-purpose` (built-in, no prefix) |

## Execution Options Reference

| Option | Description |
|--------|-------------|
| **Parallel session** | User opens new session, batch execution with human review -> `executing-plans` |
| **Save for later** | Plan at `docs/plans/YYYY-MM-DD-<feature>.md`, manual review before execution |

## Required Patterns

This skill uses these universal patterns:
- **State Tracking:** See `shared-patterns/state-tracking.md`
- **Failure Recovery:** See `shared-patterns/failure-recovery.md`
- **Exit Criteria:** See `shared-patterns/exit-criteria.md`
- **TodoWrite:** See `shared-patterns/todowrite-integration.md`

Apply ALL patterns when using this skill.
