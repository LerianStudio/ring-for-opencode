---
name: "ring:resume-handoff"
description: Resume work from a handoff document with context analysis and validation
agent: build
subtask: false
---

# /ring:resume-handoff Command

Resumes work from a previously created handoff document.

> **Note:** If you just ran `/ring:create-handoff` followed by `/clear`, the handoff is **automatically loaded** - you don't need this command. Use this command to manually resume from **older handoffs** or **specific handoff files**.

## Usage

```
/ring:resume-handoff <path-to-handoff.md>
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `path` | Yes | Path to the handoff markdown file |

## How It Works

1. **Read the handoff file** at the specified path
2. **Parse all sections** to understand session state
3. **Acknowledge the context** by summarizing what was happening
4. **Identify next steps** from the handoff
5. **Continue work** from where the previous session left off

## Workflow

### Automatic Resume (Default)
```
# Previous session
User: /ring:create-handoff auth-refactor "OAuth integration"
Assistant: [Creates handoff file + breadcrumb]

# Clear context (handoff auto-loads if < 1 hour old)
User: /clear
Assistant: [Automatically detects handoff, presents context, continues work]
```

### Manual Resume (Older Handoffs)
```
# Resume from a specific older handoff
User: /ring:resume-handoff docs/handoffs/auth-refactor/2025-12-27_15-45-00_session.md
Assistant: [Reads handoff, acknowledges context, continues work]
```

## What to Do When Resuming

When this command is invoked:

1. **Read the entire handoff file**
2. **Summarize the context** back to the user:
   - What was being worked on
   - Current state
   - Key decisions made
3. **Identify blockers** from Open Questions section
4. **Propose next action** based on Next Steps section
5. **Ask for confirmation** before proceeding

## Response Template

When resuming, respond with:

```
## Resuming: {Session Name}

**Previous Session:** {timestamp from handoff}
**Status:** {status from handoff}

### Context Restored

{Brief summary of what was happening}

### Key Decisions (Already Made)

- {List key decisions so they're not revisited}

### Current State

{Where things stand}

### Recommended Next Action

Based on the handoff, the next step is: {first item from Next Steps}

Should I proceed with this, or would you like to adjust the approach?
```

## Examples

### Basic Resume
```
User: /ring:resume-handoff docs/handoffs/auth-refactor/2025-12-27_15-45-00_session.md
Assistant:

## Resuming: auth-refactor

**Previous Session:** 2025-12-27 15:45:00
**Status:** In Progress

### Context Restored

Working on OAuth provider integration for the authentication system.
Google OAuth was completed, currently implementing GitHub OAuth.

### Key Decisions (Already Made)

- Using passport.js for OAuth handling
- Storing tokens in Redis with 24h TTL
- Using refresh token rotation

### Current State

- Google OAuth: Complete
- GitHub OAuth: Partially implemented (callback handler pending)
- Tests: 12/15 passing

### Recommended Next Action

Based on the handoff, the next step is: Complete GitHub OAuth callback handler

Should I proceed with this, or would you like to adjust the approach?
```

## Error Handling

| Error | Cause | Resolution |
|-------|-------|------------|
| File not found | Invalid path | Check the path and try again |
| Invalid format | Not a handoff file | Ensure file was created with /ring:create-handoff |
| Missing sections | Incomplete handoff | Review handoff file and add missing sections |

## Tips

- **Keep handoffs fresh**: Create new handoffs frequently during long sessions
- **Be specific**: The more detail in the handoff, the better the resumption
- **Check Next Steps**: Ensure the Next Steps section is actionable
- **Review before clearing**: Verify the handoff captures everything important
- **Auto-resume works automatically**: After `/ring:create-handoff` + `/clear`, the handoff loads automatically within 1 hour. Beyond 1 hour, you'll be prompted. Use this command only for older or specific handoffs.
