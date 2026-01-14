---
name: "ring:dev-status"
description: Check the status of the current development cycle
agent: build
---

Check the status of the current development cycle.

## Usage

```
/ring:dev-status
```

## Output

Displays:
- Current cycle ID and start time
- Tasks: total, completed, in progress, pending
- Current task and gate being executed
- Assertiveness score (if tasks completed)
- Elapsed time

## Example Output

```
Development Cycle Status

Cycle ID: 2024-01-15-143000
Started: 2024-01-15 14:30:00
Status: in_progress

Tasks:
  Completed: 2/5
  In Progress: 1/5 (AUTH-003)
  Pending: 2/5

Current:
  Task: AUTH-003 - Implement refresh token
  Gate: 5/6 (ring:dev-testing)
  Iterations: 1

Metrics (completed tasks):
  Average Assertiveness: 89%
  Total Duration: 1h 45m

State file: docs/ring:dev-cycle/current-cycle.json
```

## When No Cycle is Running

```
No development cycle in progress.

Start a new cycle with:
  /ring:dev-cycle docs/tasks/your-tasks.md

Or resume an interrupted cycle:
  /ring:dev-cycle --resume
```

## Related Commands

| Command | Description |
|---------|-------------|
| `/ring:dev-cycle` | Start or resume cycle |
| `/ring:dev-cancel` | Cancel running cycle |
| `/ring:dev-report` | View feedback report |

---

**Now checking cycle status...**

Read state from: `docs/ring:dev-cycle/current-cycle.json` or `docs/ring:dev-refactor/current-cycle.json`
