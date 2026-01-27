---
name: ring:pre-dev-task-breakdown
description: |
  Gate 7: Implementation tasks - value-driven decomposition into working increments
  that deliver measurable user value.

trigger: |
  - PRD passed Gate 1 (required)
  - TRD passed Gate 3 (required)
  - All Large Track gates passed (if applicable)
  - Ready to create sprint/iteration tasks

skip_when: |
  - PRD or TRD not validated → complete earlier gates
  - Tasks already exist → proceed to Subtask Creation
  - Trivial change → direct implementation

sequence:
  after: [pre-dev-trd-creation, pre-dev-dependency-map]
  before: [pre-dev-subtask-creation, executing-plans]
---

# Task Breakdown - Value-Driven Decomposition

## Foundational Principle

**Every task must deliver working software that provides measurable user value.**

Creating technical-only or oversized tasks creates:
- Work that doesn't ship until "everything is done"
- Teams working on pieces that don't integrate
- No early validation of value or technical approach
- Waterfall development disguised as iterative process

**Tasks answer**: What working increment will be delivered?
**Tasks never answer**: How to implement that increment (that's Subtasks).

## Mandatory Workflow

| Phase | Activities |
|-------|------------|
| **1. Task Identification** | Load PRD (Gate 1, required), TRD (Gate 3, required); optional: Feature Map, API Design, Data Model, Dependency Map; identify value streams |
| **2. Decomposition** | Per component/feature: define deliverable, set success criteria, map dependencies, estimate effort (S/M/L/XL max=2 weeks), plan testing, identify risks |
| **3. Gate 7 Validation** | All TRD components covered; every task delivers working software; measurable success criteria; correct dependencies; no task >2 weeks; testing strategy defined; risks with mitigations; delivery sequence optimizes value |

## Explicit Rules

### ✅ DO Include in Tasks
Task ID, title, type (Foundation/Feature/Integration/Polish), deliverable (what ships), user value (what users can do), technical value (what it enables), success criteria (testable/measurable), dependencies (blocks/requires/optional), effort estimate (S/M/L/XL with points), testing strategy, risk identification with mitigations, Definition of Done checklist

### ❌ NEVER Include in Tasks
Implementation details (file paths, code examples), step-by-step instructions (those go in subtasks), technical-only tasks with no user value, tasks exceeding 2 weeks (break them down), vague success criteria ("improve performance"), missing dependency information, undefined testing approach

### Task Sizing Rules

| Size | Points | Duration | Scope |
|------|--------|----------|-------|
| Small (S) | 1-3 | 1-3 days | Single component |
| Medium (M) | 5-8 | 3-5 days | Few dependencies |
| Large (L) | 13 | 1-2 weeks | Multiple components |
| XL (>2 weeks) | BREAK IT DOWN | Too large | Not atomic |

### Value Delivery Rules
- **Foundation**: Enables other work (database setup, core services)
- **Feature**: Delivers user-facing capabilities
- **Integration**: Connects to external systems
- **Polish**: Optimizes or enhances (nice-to-have)

## Rationalization Table

| Excuse | Reality |
|--------|---------|
| "This 3-week task is fine" | Tasks >2 weeks hide complexity. Break it down. |
| "Setup tasks don't need value" | Setup enables value. Define what it enables. |
| "Success criteria are obvious" | Obvious to you ≠ testable. Document explicitly. |
| "Dependencies will be clear later" | Later is too late. Map them now. |
| "We don't need detailed estimates" | Without estimates, no planning possible. Size them. |
| "Technical tasks can skip user value" | Even infrastructure enables users. Define the connection. |
| "Testing strategy can be decided during" | Testing affects design. Plan it upfront. |
| "Risks aren't relevant at task level" | Risks compound across tasks. Identify them early. |
| "DoD is the same for all tasks" | Different tasks need different criteria. Specify. |
| "We can combine multiple features" | Combining hides value delivery. Keep tasks focused. |

## Red Flags - STOP

If you catch yourself writing any of these in a task, **STOP**:

- Task estimates over 2 weeks
- Tasks named "Setup X" without defining what X enables
- Success criteria like "works" or "complete" (not measurable)
- No dependencies listed (every task depends on something)
- No testing strategy (how will you verify?)
- "Technical debt" as a task type (debt reduction must deliver value)
- Vague deliverables ("improve", "optimize", "refactor")
- Missing Definition of Done

**When you catch yourself**: Refine the task until it's concrete, valuable, and testable.

## Gate 7 Validation Checklist

| Category | Requirements |
|----------|--------------|
| **Task Completeness** | All TRD components have tasks; all PRD features have tasks; each task appropriately sized (no XL+); task boundaries clear |
| **Delivery Value** | Every task delivers working software; user value explicit; technical value clear; sequence optimizes value |
| **Technical Clarity** | Success criteria measurable/testable; dependencies correctly mapped; testing approach defined; DoD comprehensive |
| **Team Readiness** | Skills match capabilities; estimates realistic; capacity available; handoffs minimized |
| **Risk Management** | Risks identified per task; mitigations defined; high-risk tasks scheduled early; fallback plans exist |
| **Multi-Module** (if applicable) | All tasks have `target:` field; all tasks have `working_directory:`; per-module files generated (if doc_organization: per-module) |

**Gate Result:** ✅ PASS → Subtasks | ⚠️ CONDITIONAL (refine oversized/vague) | ❌ FAIL (re-decompose)

## Multi-Module Task Tagging

**If TopologyConfig exists in research.md frontmatter** (from Gate 0):

### Read Topology Configuration

```yaml
# From research.md frontmatter
topology:
  scope: fullstack
  structure: monorepo | multi-repo
  modules:
    backend:
      path: packages/api
      language: golang
    frontend:
      path: packages/web
      framework: nextjs
  doc_organization: unified | per-module
```

### Task Target Assignment

Each task MUST have `target:` and `working_directory:` fields when topology is multi-module.

**Agent assignment depends on both `target` and `api_pattern`:**

| Target | API Pattern | Task Type | Agent |
|--------|-------------|-----------|-------|
| `backend` | any | API endpoints, services, data layer, CLI | `ring:backend-engineer-golang` or `ring:backend-engineer-typescript` |
| `frontend` | `direct` | UI components, pages, forms, Server Components | `ring:frontend-engineer` |
| `frontend` | `direct` | Server Actions, data fetching hooks | `ring:frontend-engineer` |
| `frontend` | `bff` | API routes, data aggregation, transformation | `ring:frontend-bff-engineer-typescript` |
| `frontend` | `bff` | UI components, pages, forms | `ring:frontend-engineer` |
| `shared` | any | CI/CD, configs, docs, cross-module utilities | DevOps or general |

### How to Determine Agent for Frontend Tasks

**Read `api_pattern` from research.md frontmatter:**

```yaml
# From research.md
topology:
  scope: fullstack
  api_pattern: direct | bff | other
```

**Decision Flow:**

```
Is task target: frontend?
├─ NO → Use backend-engineer-* based on language
└─ YES → Check api_pattern
    ├─ direct → ALL frontend tasks use frontend-engineer
    └─ bff → Split tasks:
        ├─ API routes, aggregation, transformation → frontend-bff-engineer-typescript
        └─ UI components, pages, forms → frontend-engineer
```

### Task Format with Agent Assignment

```markdown
## T-003: User Login API Endpoint

**Target:** backend
**Working Directory:** packages/api
**Agent:** ring:backend-engineer-golang

**Deliverable:** Working login API that validates credentials and returns JWT token.

...rest of task...
```

```markdown
## T-004: User Dashboard Data Aggregation

**Target:** frontend
**Working Directory:** packages/web
**Agent:** ring:frontend-bff-engineer-typescript  # Because api_pattern: bff

**Deliverable:** BFF endpoint that aggregates user profile, recent activity, and notifications.

...rest of task...
```

```markdown
## T-005: User Dashboard UI

**Target:** frontend
**Working Directory:** packages/web
**Agent:** ring:frontend-engineer  # UI task, even with BFF pattern

**Deliverable:** Dashboard page component consuming aggregated data from BFF.

...rest of task...
```

### Validation for Agent Assignment

| Check | Requirement |
|-------|-------------|
| All tasks have `Agent:` field | MANDATORY |
| Agent matches api_pattern rules | If frontend + bff → check task type |
| BFF tasks clearly separated | Data aggregation vs UI clearly split |
| No mixed responsibilities | One task = one agent |

### Per-Module Output

**Document placement depends on topology.structure:**

#### Single-Repo

All tasks in one file:
```
docs/pre-dev/{feature}/
└── tasks.md           # All tasks with target tags
```

#### Monorepo (per-module placement)

Index at root, filtered tasks in module directories:
```
docs/pre-dev/{feature}/
└── tasks.md           # Index with ALL tasks (target tags included)

{backend.path}/docs/pre-dev/{feature}/
└── tasks.md           # Backend tasks only (target: backend)

{frontend.path}/docs/pre-dev/{feature}/
└── tasks.md           # Frontend tasks only (target: frontend)
```

#### Multi-Repo (distributed placement)

Tasks distributed to respective repositories:
```
{backend.path}/docs/pre-dev/{feature}/
└── tasks.md           # Backend tasks only

{frontend.path}/docs/pre-dev/{feature}/
└── tasks.md           # Frontend tasks only
```

**Note:** For multi-repo, there is no central index. Each repo contains only its relevant tasks.

### Task Splitting Logic

```python
def split_tasks_by_module(all_tasks: list, topology: dict) -> dict:
    """
    Split tasks into module-specific files.

    Returns dict with keys: 'index', 'backend', 'frontend'
    """
    structure = topology.get('structure', 'single-repo')
    modules = topology.get('modules', {})
    backend_path = modules.get('backend', {}).get('path', '.')
    frontend_path = modules.get('frontend', {}).get('path', '.')

    backend_tasks = [t for t in all_tasks if t.get('target') == 'backend']
    frontend_tasks = [t for t in all_tasks if t.get('target') == 'frontend']
    shared_tasks = [t for t in all_tasks if t.get('target') == 'shared']

    if structure == 'single-repo':
        return {
            'index': {
                'path': f"docs/pre-dev/{feature}/tasks.md",
                'tasks': all_tasks
            }
        }

    if structure == 'monorepo':
        return {
            'index': {
                'path': f"docs/pre-dev/{feature}/tasks.md",
                'tasks': all_tasks
            },
            'backend': {
                'path': f"{backend_path}/docs/pre-dev/{feature}/tasks.md",
                'tasks': backend_tasks + shared_tasks
            },
            'frontend': {
                'path': f"{frontend_path}/docs/pre-dev/{feature}/tasks.md",
                'tasks': frontend_tasks + shared_tasks
            }
        }

    if structure == 'multi-repo':
        return {
            'backend': {
                'path': f"{backend_path}/docs/pre-dev/{feature}/tasks.md",
                'tasks': backend_tasks + shared_tasks
            },
            'frontend': {
                'path': f"{frontend_path}/docs/pre-dev/{feature}/tasks.md",
                'tasks': frontend_tasks + shared_tasks
            }
        }
```

### Module-Specific Task File Header

Each module-specific tasks.md should include:

```markdown
---
feature: {feature-name}
module: backend | frontend
filtered_from: docs/pre-dev/{feature}/tasks.md  # (monorepo only)
total_tasks: N
---

# {Feature Name} - {Module} Tasks

This file contains tasks filtered for the **{module}** module.

**Full task list:** {link to index if monorepo, or note "distributed" if multi-repo}

---
```

### Validation for Multi-Module

| Check | Requirement |
|-------|-------------|
| All tasks have `target:` | If topology is monorepo or multi-repo |
| All tasks have `working_directory:` | If topology is monorepo or multi-repo |
| Target matches task content | Backend tasks have backend work, etc. |
| Working directory resolves correctly | Path exists or will be created |

---

## Task Template Structure

Output to `docs/pre-dev/{feature-name}/tasks.md`. Each task includes:

| Section | Content |
|---------|---------|
| **Header** | T-[XXX]: [Task Title - What It Delivers] |
| **Target** | backend \| frontend \| shared (if multi-module) |
| **Working Directory** | Path from topology config (if multi-module) |
| **Agent** | Recommended agent: ring:backend-engineer-*, ring:frontend-*-engineer-*, etc. |
| **Deliverable** | One sentence: what working software ships |
| **Scope** | Includes (specific capabilities), Excludes (future tasks with IDs) |
| **Success Criteria** | Testable items: Functional, Technical, Operational, Quality |
| **User/Technical Value** | What users can do; what this enables |
| **Technical Components** | From TRD, From Dependencies |
| **Dependencies** | Blocks (T-AAA), Requires (T-BBB), Optional (T-CCC) |
| **Effort Estimate** | Size (S/M/L/XL), Points, Duration, Team type |
| **Risks** | Per risk: Impact, Probability, Mitigation, Fallback |
| **Testing Strategy** | Unit, Integration, E2E, Performance, Security |
| **Definition of Done** | Code reviewed, tests passing, docs updated, security clean, performance met, deployed to staging, PO acceptance, monitoring configured |

## Common Violations

| Violation | Wrong | Correct |
|-----------|-------|---------|
| **Technical-Only Tasks** | "Setup PostgreSQL Database" with install/configure steps | "User Data Persistence Foundation" with deliverable (working DB layer <100ms), user value (enables T-002/T-003), success criteria (users table, pooling, migrations) |
| **Oversized Tasks** | "Complete User Management System" (6 weeks) with all auth features combined | Split into: T-005 Basic Auth (L), T-006 Password Mgmt (M), T-007 2FA (M), T-008 Permissions (L) |
| **Vague Success Criteria** | "Feature works, Tests pass, Code reviewed" | Functional (upload 100MB, formats), Technical (<2s response), Operational (99.5% success rate), Quality (90% coverage) |

## Delivery Sequencing

Optimize task order by sprint/phase with goals, critical path identification, and parallel work opportunities.

## Confidence Scoring

| Factor | Points | Criteria |
|--------|--------|----------|
| Task Decomposition | 0-30 | All appropriately sized: 30, Most well-scoped: 20, Too large/vague: 10 |
| Value Clarity | 0-25 | Every task delivers working software: 25, Most clear: 15, Unclear: 5 |
| Dependency Mapping | 0-25 | All documented: 25, Most clear: 15, Ambiguous: 5 |
| Estimation Quality | 0-20 | Based on past work: 20, Educated guesses: 12, Speculation: 5 |

**Action:** 80+ autonomous | 50-79 present options | <50 ask about velocity

## Output & After Approval

**Output to (depends on topology.structure):**

| Structure | Files Generated |
|-----------|-----------------|
| single-repo | `docs/pre-dev/{feature}/tasks.md` |
| monorepo | Index + `{backend.path}/docs/pre-dev/{feature}/tasks.md` + `{frontend.path}/docs/pre-dev/{feature}/tasks.md` |
| multi-repo | `{backend.path}/docs/pre-dev/{feature}/tasks.md` + `{frontend.path}/docs/pre-dev/{feature}/tasks.md` |

1. ✅ Tasks become sprint backlog
2. 🎯 Use as input for subtasks (`pre-dev-subtask-creation`)
3. 📊 Track progress per task (not per subtask)
4. 🚫 No implementation yet - that's in subtasks

## The Bottom Line

**If you created tasks that don't deliver working software, rewrite them.**

Tasks are not technical activities. Tasks are working increments.

"Setup database" is not a task. "User data persists correctly" is a task.
"Implement OAuth" is not a task. "Users can log in with Google" is a task.
"Write tests" is not a task. Tests are part of Definition of Done for other tasks.

Every task must answer: **"What working software can I demo to users?"**

If you can't demo it, it's not a task. It's subtask implementation detail.

**Deliver value. Ship working software. Make tasks demoable.**
