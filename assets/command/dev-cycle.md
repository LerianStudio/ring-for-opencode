---
name: "ring:dev-cycle"
description: Execute the 6-gate development cycle for tasks in a markdown file
agent: build
---

Execute the development cycle for tasks in a markdown file.

## Usage

```
/dev-cycle [tasks-file] [options]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `tasks-file` | Yes* | Path to markdown file with tasks (e.g., `docs/tasks/sprint-001.md`) |

*Can be omitted if `docs/tasks/current.md` exists.

## Options

| Option | Description | Example |
|--------|-------------|---------|
| `--task ID` | Execute only specific task | `--task PROJ-123` |
| `--skip-gates` | Skip specific gates | `--skip-gates devops,review` |
| `--dry-run` | Validate tasks without executing | `--dry-run` |
| `--resume` | Resume interrupted cycle | `--resume` |

## Gates Executed

| Gate | Skill | Description |
|------|-------|-------------|
| 0 | `dev-implementation` | Implement code (TDD) |
| 1 | `dev-devops` | Create Docker/compose |
| 2 | `dev-sre` | Observability validation |
| 3 | `dev-testing` | Write and run tests (85%+ coverage) |
| 4 | `requesting-code-review` | Code review (5 reviewers in parallel) |
| 5 | `dev-validation` | Final validation |

## Execution Mode

Before executing, ask user for execution mode:
- **Manual per subtask**: Checkpoint after each subtask
- **Manual per task**: Checkpoint after each task
- **Automatic**: No checkpoints, run all gates

## Output

- **State file**: `docs/dev-cycle/current-cycle.json`
- **Feedback report**: `.ring/dev-team/feedback/cycle-YYYY-MM-DD.md`

## Quick Rules

- **ALL 6 gates execute** - Checkpoints affect pauses, not gates
- **Gates execute in order** - 0 -> 1 -> 2 -> 3 -> 4 -> 5
- **Gate 4 requires ALL 5 reviewers** - 3/5 = FAIL
- **Coverage threshold** - 85% minimum, no exceptions
- **State persisted** - Can resume with `--resume`

---

**Now loading skill: dev-cycle**

Pass arguments: $ARGUMENTS
