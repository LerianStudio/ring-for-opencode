---
name: "ring:create-handoff"
description: Create a handoff document capturing current session state for future resumption
agent: build
subtask: false
---

# /ring:create-handoff Command

Creates a comprehensive handoff document that captures the current session's context, progress, decisions, and next steps. After creating the handoff, simply run `/clear` - the handoff will be **automatically loaded** in the new session (no need to run `/ring:resume-handoff` manually).

## Usage

```
/ring:create-handoff [session-name] [description]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `session-name` | No | Short identifier for the session (defaults to feature/task name) |
| `description` | No | Brief description of current state |

## What Gets Captured

The handoff document includes:

1. **Session Summary** - What was being worked on in this entire session (organize per sub-topics)
2. **Current State** - Where things stand right now
3. **Completed Work** - What was accomplished
4. **In-Progress Work** - What's partially done
5. **Key Decisions** - Important choices made and why
6. **What Worked** - Successful approaches
7. **What Didn't Work** - Failed approaches to avoid
8. **Open Questions** - Unresolved items needing attention
9. **Next Steps** - Clear actions for resumption
10. **Relevant Files** - Key files touched or to be modified

## Output Location

Handoffs are saved to: `docs/handoffs/{session-name}/{timestamp}_{description}.md`

Example: `docs/handoffs/auth-refactor/2025-12-27_15-45-00_oauth-integration.md`

## Workflow

1. **Create handoff**: `/ring:create-handoff auth-refactor "OAuth integration complete"`
2. **Clear context**: Run `/clear` to reset conversation
3. **Auto-resume**: The handoff is automatically loaded in the new session

> **Note:** If more than 1 hour passes between creating the handoff and running `/clear`, you'll be asked whether to resume instead of auto-loading. For manual resume of older handoffs, use `/ring:resume-handoff <path>`.

## Handoff Template

When invoked, create a file with this structure:

```markdown
# Handoff: {Session Name}

**Created:** {timestamp}
**Status:** {In Progress | Blocked | Ready for Review | Complete}

## Summary

{1-2 sentence overview of what this session was about}

## Current State

{Where things stand right now - be specific about what's done vs pending}

## Completed Work

- {List of completed items with file references where relevant}

## In-Progress Work

- {Partially completed items - describe exactly where you stopped}

## Key Decisions

| Decision | Rationale | Alternatives Considered |
|----------|-----------|------------------------|
| {decision} | {why} | {what else was considered} |

## What Worked

- {Successful approaches, patterns, or solutions worth reusing}

## What Didn't Work

- {Failed approaches - document to avoid repeating}

## Open Questions

- [ ] {Unresolved questions that need answers}
- [ ] {Blockers or dependencies}

## Next Steps

1. {First thing to do when resuming}
2. {Second priority}
3. {Third priority}

## Relevant Files

| File | Purpose | Status |
|------|---------|--------|
| `path/to/file` | {what it does} | {modified/created/to-modify} |

## Context for Resumption

{Any additional context the next session needs - gotchas, environment setup, etc.}
```

## Auto-Resume Breadcrumb

After creating the handoff file, you MUST also write the breadcrumb file for auto-resume:

**File:** `docs/handoffs/.pending`

**Content (exactly 2 lines):**
```
{absolute-path-to-the-handoff-file}
{unix-timestamp-in-seconds}
```

**Example:**
```
/Users/dev/project/docs/handoffs/auth-refactor/2025-12-27_15-45-00_oauth-integration.md
1735311900
```

**How to generate the timestamp:** Use Bash tool with `date +%s` to get the current Unix timestamp.

**How to get the absolute path:** Use Bash tool with `realpath` or `pwd` to construct the full path.

**IMPORTANT:** The `.pending` file MUST be written AFTER the handoff file is created. This breadcrumb enables the SessionStart hook to auto-detect and load the handoff when the user runs `/clear`.

Also ensure `docs/handoffs/.gitignore` exists with:
```
.pending
```
This prevents the ephemeral breadcrumb from being committed.

## Examples

### Basic Usage
```
User: /ring:create-handoff
Assistant: I'll create a handoff document for the current session.
[Creates docs/handoffs/current-session/2025-12-27_15-45-00_session.md with filled template]
```

### With Session Name
```
User: /ring:create-handoff auth-refactor "OAuth provider integration"
Assistant: I'll create a handoff document for the auth-refactor session.
[Creates docs/handoffs/auth-refactor/2025-12-27_15-45-00_session.md with filled template]

Summary:
- Created: docs/handoffs/auth-refactor/2025-12-27_15-45-00_session.md
- Session ID: auth-refactor_2025-12-27_15-45-00
- Title: OAuth provider integration
```
