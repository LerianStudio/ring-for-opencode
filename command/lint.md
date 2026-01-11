---
description: Run lint checks and dispatch parallel agents to fix all issues
agent: build
subtask: true
---

Run linting tools, analyze results, and dispatch parallel agents to fix all issues until the codebase is clean.

## Process

### Phase 1: Lint Execution
- Run `make lint` or detect appropriate lint command
- Capture all output and exit codes
- Parse errors, warnings, and their locations

### Phase 2: Stream Analysis
- Group issues by file or logical component
- Identify independent streams that can be fixed in parallel
- Create fix assignments for each stream

### Phase 3: Parallel Agent Dispatch
- Launch N agents simultaneously (one per stream)
- Each agent fixes issues in their assigned scope
- Agents work independently without conflicts

### Phase 4: Verification Loop
- Re-run lint after all agents complete
- If issues remain, analyze and dispatch new agents
- Continue until lint passes completely

## Critical Constraints

- **NO AUTOMATED SCRIPTS** - Agents fix code directly, never create automation scripts
- **NO DOCUMENTATION** - Agents fix lint issues only, don't add docs/comments
- **DIRECT FIXES ONLY** - Each issue is fixed manually by editing the source

## Examples

```
/lint                    # Lint entire codebase
/lint src/services/      # Lint specific path
```

$ARGUMENTS

Optionally specify a path to lint (defaults to entire codebase).
