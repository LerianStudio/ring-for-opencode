---
name: "ring:dev-refactor"
description: Analyzes codebase against standards and generates refactoring tasks for dev-cycle.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: workflow
  trigger:
    - User wants to refactor existing project to follow standards
    - Legacy codebase needs modernization
    - Project audit requested
  skip_when:
    - Greenfield project - Use /pre-dev-* instead
    - Single file fix - Use dev-cycle directly
---

# Dev Refactor Skill

Analyzes existing codebase against Ring/Lerian standards and generates refactoring tasks compatible with dev-cycle.

## Mandatory Gap Principle

**ANY divergence from Ring standards = MANDATORY gap to implement.**

| Principle | Meaning |
|-----------|---------|
| **ALL divergences are gaps** | Every difference MUST be tracked as FINDING-XXX |
| **Severity affects PRIORITY, not TRACKING** | Low severity = lower priority, NOT optional |
| **NO filtering allowed** | You CANNOT decide which divergences "matter" |

## Workflow Steps

### Step 0: Validate PROJECT_RULES.md

Check if `docs/PROJECT_RULES.md` exists:

```yaml
If docs/PROJECT_RULES.md exists:
  - Load and continue to Step 1

If docs/PROJECT_RULES.md does NOT exist:
  - Check if {OPENCODE_CONFIG}/templates/PROJECT_RULES.md exists
  - Present option to user:
    "PROJECT_RULES.md not found. Would you like me to create one from the Ring template?"
    - YES: Copy {OPENCODE_CONFIG}/templates/PROJECT_RULES.md to docs/PROJECT_RULES.md
           Ask user to review and customize for their project
           Continue to Step 1
    - NO:  STOP with message:
           "Cannot proceed without PROJECT_RULES.md. Either:
           1. Create from template with /dev-refactor --init
           2. Run /pre-dev-full to generate project standards
           3. Create manually at docs/PROJECT_RULES.md"
```

### Step 1: Detect Project Stack
| File/Pattern | Stack | Agent |
|--------------|-------|-------|
| `go.mod` | Go Backend | backend-engineer-golang |
| `package.json` + React | Frontend | frontend-engineer |
| `package.json` + Express | TypeScript Backend | backend-engineer-typescript |

### Step 2: Read PROJECT_RULES.md
Extract project-specific conventions for agent context.

### Step 3: Generate Codebase Report
```yaml
Task:
  subagent_type: "codebase-explorer"
  description: "Generate codebase architecture report"
  prompt: |
    Generate comprehensive codebase report:
    - Project structure and directory layout
    - Architecture pattern (hexagonal, clean, etc.)
    - Technology stack from manifests
    - Code patterns: config, database, handlers, errors, telemetry, testing
```

### Step 4: Dispatch Specialist Agents (Parallel)

For Go projects:
- `backend-engineer-golang` - Go standards analysis
- `qa-analyst` - Test coverage analysis
- `devops-engineer` - DevOps analysis
- `sre` - Observability analysis

### Step 5: Generate findings.md
Every issue from agents becomes FINDING-XXX entry with:
- Severity
- Category
- Current Code (file:line)
- Ring Standard Reference
- Required Changes

### Step 6: Group Findings into Tasks
Create REFACTOR-XXX tasks grouping related findings.

### Step 7: Generate tasks.md
```markdown
## REFACTOR-001: [Task Name]
**Priority:** Critical | High | Medium | Low
**Effort:** [hours]h

### Findings Addressed
| Finding | Pattern | Severity |
|---------|---------|----------|

### Required Actions
1. [ ] [action from finding]
```

### Step 8: User Approval
Present plan for approval before execution.

### Step 9: Handoff to dev-cycle
```yaml
Skill tool:
  skill: "ring:dev-cycle"
```
Pass tasks file: `docs/refactor/{timestamp}/tasks.md`

## Output Artifacts

```
docs/refactor/{timestamp}/
├── codebase-report.md  (Step 3)
├── reports/            (Step 4)
│   ├── backend-engineer-golang-report.md
│   ├── qa-analyst-report.md
│   ├── devops-engineer-report.md
│   └── sre-report.md
├── findings.md         (Step 5)
└── tasks.md            (Step 7)
```

## Anti-Rationalization

| Rationalization | Why Wrong | Required Action |
|-----------------|-----------|-----------------|
| "Multiple similar issues can be one finding" | Distinct file:line = distinct finding | **One issue = One FINDING-XXX** |
| "Some findings are duplicates across agents" | Different perspectives matter | **Create separate findings** |
| "Team has approved this deviation" | Approval ≠ compliance | **Create FINDING-XXX, note decision** |
