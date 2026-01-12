---
name: "ring:dev-refactor"
description: Analyze existing codebase against standards and generate refactoring tasks for dev-cycle
agent: build
---

Analyze existing codebase against Ring/Lerian standards and execute refactoring through dev-cycle.

## PRE-EXECUTION CHECK (MANDATORY)

**Before loading the skill:**

```
Does docs/PROJECT_RULES.md exist?
├── YES → Load skill: ring:dev-refactor
└── NO  → Output blocker and STOP
```

**If file does NOT exist:**

```markdown
## HARD BLOCK: PROJECT_RULES.md Not Found

**Status:** BLOCKED - Cannot proceed

### Required Action
Create `docs/PROJECT_RULES.md` with your project's:
- Architecture patterns
- Code conventions
- Testing requirements
- DevOps standards

Then re-run `/dev-refactor`.
```

## Usage

```
/dev-refactor [path] [options]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `path` | No | Directory to analyze (default: project root) |

## Options

| Option | Description | Example |
|--------|-------------|---------|
| `--standards PATH` | Custom standards file | `--standards docs/MY_RULES.md` |
| `--analyze-only` | Generate report without executing | `--analyze-only` |
| `--critical-only` | Only Critical and High priority | `--critical-only` |
| `--dry-run` | Show what would be analyzed | `--dry-run` |

## Analysis Dimensions

| Dimension | What's Checked |
|-----------|----------------|
| **Architecture** | DDD patterns, layer separation, directory structure |
| **Code Quality** | Naming, error handling, forbidden practices |
| **Instrumentation** | Tracing, logging, context propagation |
| **Testing** | Coverage, test patterns, naming |
| **DevOps** | Dockerfile, docker-compose, env management |

## Severity Levels (ALL MANDATORY)

| Level | Priority | Tracking |
|-------|----------|----------|
| **Critical** | Fix immediately | **MANDATORY** |
| **High** | Fix in current sprint | **MANDATORY** |
| **Medium** | Fix in next sprint | **MANDATORY** |
| **Low** | Fix when capacity | **MANDATORY** |

**ALL severities are MANDATORY. Low = Lower priority, NOT optional.**

## Output

- **Analysis Report**: `docs/refactor/{timestamp}/analysis-report.md`
- **Tasks File**: `docs/refactor/{timestamp}/tasks.md`

## User Approval

Before executing dev-cycle, ask user:
- **Approve all**: Proceed to dev-cycle execution
- **Critical only**: Execute only Critical/High tasks
- **Cancel**: Keep analysis, skip execution

---

**Now loading skill: ring:dev-refactor**

Pass arguments: $ARGUMENTS
