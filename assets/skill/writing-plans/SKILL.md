---
name: ring:writing-plans
description: |
  Creates comprehensive implementation plans with exact file paths, complete code
  examples, and verification steps for engineers with zero codebase context.

trigger: |
  - Design phase complete (brainstorming/PRD/TRD validated)
  - Need to create executable task breakdown
  - Creating work for other engineers or AI agents

skip_when: |
  - Design not validated → use brainstorming first
  - Requirements still unclear → use pre-dev-prd-creation first
  - Already have a plan → use executing-plans

sequence:
  after: [brainstorming, pre-dev-trd-creation]
  before: [executing-plans, subagent-driven-development]

related:
  similar: [brainstorming]
---

# Writing Plans

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
3. Example: "Implement OAuth2 authentication!" → keywords: "oauth2 authentication"

**Safety:** Never include shell metacharacters (`;`, `$`, `` ` ``, `|`, `&`) in keywords.

Results inform the plan:
- `successful_handoffs` → Reference patterns that worked
- `failed_handoffs` → WARN in plan, design to avoid
- `relevant_plans` → Review for approach ideas

If `is_empty_index: true` → Proceed without precedent (normal for new projects).

**Step 2: Dispatch Write-Plan Agent**

Dispatch via `Task(subagent_type: "ring:write-plan", model: "opus")` with:
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
- `PASS` → Plan is safe to execute
- `WARNING` → Plan has >30% keyword overlap with past failures
  - Review the warnings in the output
  - Update plan to address the failure patterns
  - Re-run validation until PASS

**Note:** If artifact index is unavailable, validation passes automatically (nothing to check against).

**Step 4: Ask User About Execution**

Ask via `AskUserQuestion`: "Execute now?" Options:
1. Execute now → `subagent-driven-development`
2. Parallel session → user opens new session with `executing-plans`
3. Save for later → report location and end

## Why Use an Agent?

**Context preservation** (reading many files keeps supervisor clean) | **Model power** (Opus for comprehensive planning) | **Separation of concerns** (supervisor orchestrates, agent plans)

## What the Agent Does

**Query precedent** → Explore codebase → identify files → break into bite-sized tasks (2-5 min) → write complete code → include exact commands → add review checkpoints → **include Historical Precedent section** → verify Zero-Context Test → save to `docs/plans/YYYY-MM-DD-<feature>.md` → report back

## Requirements for Plans

Every plan: **Historical Precedent section** | Header (goal, architecture, tech stack) | Verification commands with expected output | Exact file paths (never "somewhere in src") | Complete code (never "add validation here") | Bite-sized steps with verification | Failure recovery | Review checkpoints | Zero-Context Test | **Recommended agents per task** | **Avoids known failure patterns**

### Multi-Module Task Requirements

**If TopologyConfig exists** (from pre-dev research.md frontmatter or user input):

Each task MUST include:
- **Target:** `backend` | `frontend` | `shared`
- **Working Directory:** Resolved path from topology configuration
- **Agent:** Recommended agent matching the target

**Task Format with Target:**

```markdown
## Task 3: Create User Login API

**Target:** backend
**Working Directory:** packages/api
**Agent:** ring:backend-engineer-golang

**Files to Create/Modify:**
- `packages/api/internal/handlers/auth.go`
- `packages/api/internal/services/auth_service.go`

...rest of task...
```

**Target Assignment Rules:**

| Target | When | Agent |
|--------|------|-------|
| `backend` | API endpoints, services, data layer, CLI | `backend-engineer-{golang,typescript}` |
| `frontend` | UI components, pages, BFF routes | See [Frontend Tasks (api_pattern aware)](#frontend-tasks-api_pattern-aware) |
| `shared` | CI/CD, configs, docs, cross-module | `devops-engineer` or `general-purpose` |

**Working Directory Resolution:**

| Topology Structure | Backend Path | Frontend Path |
|-------------------|--------------|---------------|
| `single-repo` | `.` | `.` |
| `monorepo` | `topology.modules.backend.path` | `topology.modules.frontend.path` |
| `multi-repo` | `topology.modules.backend.path` (absolute) | `topology.modules.frontend.path` (absolute) |

## Agent Selection

### Backend Tasks

| Task Type | Agent |
|-----------|-------|
| Go backend API/services | `backend-engineer-golang` |
| TypeScript backend API/services | `backend-engineer-typescript` |

### Frontend Tasks (api_pattern aware)

**Read `api_pattern` from topology configuration to determine correct agent:**

| API Pattern | Task Type | Agent |
|-------------|-----------|-------|
| `direct` | UI components, pages, forms | `frontend-engineer` |
| `direct` | Server Actions, data fetching | `frontend-engineer` |
| `direct` | Server Components with data loading | `frontend-engineer` |
| `bff` | API routes (`/api/*`) | `frontend-bff-engineer-typescript` |
| `bff` | Data aggregation, transformation | `frontend-bff-engineer-typescript` |
| `bff` | External service integration | `frontend-bff-engineer-typescript` |
| `bff` | UI components, pages, forms | `frontend-engineer` |
| `other` | Depends on pattern | Ask user or use frontend-engineer default |

### Decision Logic for Frontend Tasks

```
def get_frontend_agent(task, topology):
    api_pattern = topology.get('api_pattern', 'direct')

    if api_pattern == 'direct':
        return 'frontend-engineer'

    if api_pattern == 'bff':
        if is_bff_task(task):  # API routes, aggregation, transformation
            return 'frontend-bff-engineer-typescript'
        else:  # UI components, pages
            return 'frontend-engineer'

    return 'frontend-engineer'  # Default for 'other'

def is_bff_task(task):
    bff_indicators = [
        'API route', 'api route', '/api/',
        'aggregat', 'transform', 'BFF',
        'external service', 'backend service',
        'data layer', 'HTTP client'
    ]
    return any(ind in task.description for ind in bff_indicators)
```

### Infrastructure and Other Tasks

| Task Type | Agent |
|-----------|-------|
| Infra/CI/CD | `devops-engineer` |
| Testing | `qa-analyst` |
| Reliability | `sre` |
| Fallback | `general-purpose` (built-in, no prefix) |

### Task Format with api_pattern

When TopologyConfig includes `api_pattern`, include it in task metadata:

```markdown
## Task 3: Aggregate Dashboard Data

**Target:** frontend
**Working Directory:** packages/web
**API Pattern:** bff
**Agent:** ring:frontend-bff-engineer-typescript

**Files to Create/Modify:**
- `packages/web/app/api/dashboard/route.ts`
- `packages/web/lib/services/dashboard-aggregator.ts`

...rest of task...
```

## Execution Options Reference

| Option | Description |
|--------|-------------|
| **Execute now** | Fresh subagent per task, code review between tasks → `subagent-driven-development` |
| **Parallel session** | User opens new session, batch execution with human review → `executing-plans` |
| **Save for later** | Plan at `docs/plans/YYYY-MM-DD-<feature>.md`, manual review before execution |

## Required Patterns

This skill uses these universal patterns:
- **State Tracking:** See `skills/shared-patterns/state-tracking.md`
- **Failure Recovery:** See `skills/shared-patterns/failure-recovery.md`
- **Exit Criteria:** See `skills/shared-patterns/exit-criteria.md`
- **TodoWrite:** See `skills/shared-patterns/todowrite-integration.md`

Apply ALL patterns when using this skill.
