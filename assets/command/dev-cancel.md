---
name: "ring:dev-cancel"
description: Cancel the current development cycle
agent: build
---

Cancel the current development cycle.

## Usage

```
/ring:dev-cancel [--force]
```

## Options

| Option | Description |
|--------|-------------|
| `--force` | Cancel without confirmation |

## Behavior

1. **Confirmation**: Asks for confirmation before canceling (unless `--force`)
2. **State preservation**: Saves current state for potential resume
3. **Cleanup**: Marks cycle as `cancelled` in state file
4. **Report**: Generates partial feedback report with completed tasks

## Confirmation Prompt

```
Cancel Development Cycle?

Cycle ID: 2024-01-15-143000
Progress: 3/5 tasks completed

This will:
- Stop the current cycle
- Save state for potential resume
- Generate partial feedback report

[Confirm Cancel] [Keep Running]
```

## After Confirmation

```
Cycle Cancelled

Cycle ID: 2024-01-15-143000
Status: cancelled
Completed: 3/5 tasks

State saved to: docs/ring:dev-cycle/current-cycle.json
Partial report: .ring/dev-team/feedback/cycle-2024-01-15-partial.md

To resume later:
  /ring:dev-cycle --resume
```

## When No Cycle is Running

```
No development cycle to cancel.

Check status with:
  /ring:dev-status
```

## Related Commands

| Command | Description |
|---------|-------------|
| `/ring:dev-cycle` | Start or resume cycle |
| `/ring:dev-status` | Check current status |
| `/ring:dev-report` | View feedback report |

---

**Now checking for active cycle to cancel...**

Read state from: `docs/ring:dev-cycle/current-cycle.json` or `docs/ring:dev-refactor/current-cycle.json`

Pass arguments: $ARGUMENTS
