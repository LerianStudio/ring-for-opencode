---
name: "ring:writing-plans"
description: |
  Creates comprehensive implementation plans with exact file paths, complete code
  examples, and verification steps for engineers with zero codebase context.
license: MIT
compatibility: opencode
metadata:
  trigger: "Design phase complete, need executable task breakdown, creating work for others"
  skip_when: "Design not validated, requirements unclear, already have a plan"
  sequence_after: "ring:brainstorming, ring:pre-dev-trd-creation"
  sequence_before: "ring:executing-plans"
---

# Writing Plans

## Related Skills

**Similar:** ring:brainstorming

---

## Overview

This skill dispatches a specialized agent to write comprehensive implementation plans for engineers with zero codebase context.

**Announce at start:** "I'm using the ring:writing-plans skill to create the implementation plan."

**Context:** This should be run in a dedicated worktree (created by ring:brainstorming skill).

## The Process

**Step 1: Dispatch Write-Plan Agent**

Dispatch via `Task(subagent_type: "ring:write-plan")` with:
- Instructions to create bite-sized tasks (2-5 min each)
- Include exact file paths, complete code, verification steps
- Save to `docs/plans/YYYY-MM-DD-<feature-name>.md`

**Step 2: Ask User About Execution**

Ask via `question tool`: "Execute now?" Options:
1. Parallel session -> user opens new session with `ring:executing-plans`
2. Save for later -> report location and end

## Why Use an Agent?

**Context preservation** (reading many files keeps supervisor clean) | **Separation of concerns** (supervisor orchestrates, agent plans)

## What the Agent Does

**Explore codebase -> identify files -> break into bite-sized tasks (2-5 min) -> write complete code -> include exact commands -> add review checkpoints -> verify Zero-Context Test -> save to `docs/plans/YYYY-MM-DD-<feature>.md` -> report back

## Requirements for Plans

Every plan: Header (goal, architecture, tech stack) | Verification commands with expected output | Exact file paths (never "somewhere in src") | Complete code (never "add validation here") | Bite-sized steps with verification | Failure recovery | Review checkpoints | Zero-Context Test | **Recommended agents per task**

## Agent Selection

| Task Type | Agent |
|-----------|-------|
| Backend API/services | `backend-engineer-{golang,typescript}` |
| Frontend | `ring:frontend-engineer` |
| Infra/CI/CD | `ring:devops-engineer` |
| Testing | `ring:qa-analyst` |
| Reliability | `ring:sre` |
| Fallback | `general-purpose` (built-in, no prefix) |

## Execution Options Reference

| Option | Description |
|--------|-------------|
| **Parallel session** | User opens new session, batch execution with human review -> `ring:executing-plans` |
| **Save for later** | Plan at `docs/plans/YYYY-MM-DD-<feature>.md`, manual review before execution |

## Required Patterns

This skill uses these universal patterns:
- **State Tracking:** See `shared-patterns/state-tracking.md`
- **Failure Recovery:** See `shared-patterns/failure-recovery.md`
- **Exit Criteria:** See `shared-patterns/exit-criteria.md`
- **todowrite tool:** See `shared-patterns/todowrite-integration.md`

Apply ALL patterns when using this skill.
