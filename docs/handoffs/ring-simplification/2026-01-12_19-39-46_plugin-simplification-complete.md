---
date: 2026-01-12T22:39:46Z
session_name: ring-simplification
git_commit: b40c18940df018708b70646fea0d539b7758fc5b
branch: main
repository: ring-for-opencode
topic: "Ring Plugin Simplification"
tags: [refactoring, simplification, prompts, externalization]
status: complete
outcome: UNKNOWN
root_span_id:
turn_span_id:
---

# Handoff: Ring Plugin Simplification - Remove Orchestrator/Background/Notifications

## Task Summary

Major simplification of the Ring plugin by removing complex orchestration features and externalizing hardcoded prompts.

**Completed:**
1. Removed `plugin/background/` module (async task management, concurrency control)
2. Removed `plugin/orchestrator/` module (worker pools, job registry, workflow engine, task tools)
3. Removed notification hook (cross-platform desktop notifications)
4. Removed task-completion hook (todo state tracking)
5. Removed ledger features from session-start and context-injection hooks
6. Externalized 8 hardcoded prompts to `assets/prompts/` folder
7. Made skills/commands/agents references programmatic (generated from actual loaded assets)

## Critical References
- `plugin/ring-unified.ts` - Main plugin entry point (simplified)
- `plugin/hooks/factories/context-injection.ts` - Programmatic reference generation
- `assets/prompts/` - Externalized prompt files

## Recent Changes

### Deleted Folders
- `plugin/background/` - Entire module removed
- `plugin/orchestrator/` - Entire module removed
- `__tests__/plugin/background/` - Tests for removed module
- `__tests__/plugin/orchestrator/` - Tests for removed module

### Deleted Files
- `plugin/hooks/factories/notification.ts`
- `plugin/hooks/factories/task-completion.ts`
- `__tests__/plugin/hooks/task-completion-notification.integration.test.ts`
- `assets/prompts/context-injection/agents-reference.txt` (now programmatic)
- `assets/prompts/context-injection/skills-reference.txt` (now programmatic)
- `assets/prompts/context-injection/commands-reference.txt` (now programmatic)

### Modified Files
- `plugin/ring-unified.ts` - Removed background manager, simplified event handling
- `plugin/tools/index.ts` - Returns empty object (no orchestrator tools)
- `plugin/config/schema.ts` - Removed BackgroundTaskConfig, NotificationConfig, orchestrator schemas
- `plugin/config/loader.ts` - Removed getBackgroundTaskConfig(), getNotificationConfig()
- `plugin/config/index.ts` - Cleaned up exports
- `plugin/hooks/factories/index.ts` - Removed notification/task-completion exports
- `plugin/hooks/factories/session-start.ts` - Removed ledger loading, externalized prompts
- `plugin/hooks/factories/context-injection.ts` - Removed ledger summary, externalized prompts, programmatic references
- `plugin/hooks/types.ts` - Removed notification, task-completion from HookName
- `__tests__/plugin/security-hardening.test.ts` - Removed notification/ledger tests

### Created Files
- `assets/prompts/session-start/critical-rules.txt`
- `assets/prompts/session-start/agent-reminder.txt`
- `assets/prompts/session-start/duplication-guard.txt`
- `assets/prompts/session-start/doubt-questions.txt`
- `assets/prompts/context-injection/compact-rules.txt`

## Learnings

### What Worked
- **Incremental deletion approach** - Removing modules one by one and fixing imports after each helped catch issues early
- **Subagent delegation** - Using subagents for multi-file changes was efficient
- **Prompt externalization pattern** - Content in .txt files, XML wrappers in TypeScript code provides good separation

### What Failed
- **Initial test run** - Forgot to delete test files for removed modules, causing test failures
- **Security test file** - Had to manually update to remove references to deleted notification functions

### Key Decisions
- Decision: **Centralized prompts in `assets/prompts/`** instead of distributed per-module
  - Alternatives: `plugin/hooks/factories/prompts/` folder
  - Reason: Consistent with existing `assets/` structure (agents, skills, commands)

- Decision: **Programmatic reference generation** for skills/commands/agents
  - Alternatives: Keep static text files
  - Reason: References auto-update when assets change, includes user customizations

- Decision: **Keep XML wrapper tags in TypeScript**
  - Alternatives: Include tags in .txt files
  - Reason: Enforces structure, allows graceful fallback if file missing

## Files Modified

**Deleted (~2,500+ lines removed):**
- `plugin/background/` - 4 files (manager, concurrency, types, index)
- `plugin/orchestrator/` - 8+ files (config, profiles, jobs, worker-pool, workflow/engine, tools/task-tools, types, index)
- `plugin/hooks/factories/notification.ts`
- `plugin/hooks/factories/task-completion.ts`

**Created:**
- `assets/prompts/session-start/critical-rules.txt`
- `assets/prompts/session-start/agent-reminder.txt`
- `assets/prompts/session-start/duplication-guard.txt`
- `assets/prompts/session-start/doubt-questions.txt`
- `assets/prompts/context-injection/compact-rules.txt`

**Modified:**
- `plugin/ring-unified.ts` - Simplified (removed ~50 lines)
- `plugin/tools/index.ts` - Now returns empty object
- `plugin/config/schema.ts` - Removed 3 schemas
- `plugin/hooks/factories/context-injection.ts` - Added programmatic generators

## Action Items & Next Steps

1. **Consider further simplification** - Review remaining hooks for necessity
2. **Update documentation** - AGENTS.md may need updates to reflect removed features
3. **Test in real usage** - Verify the simplified plugin works correctly with opencode
4. **Consider removing utils/state.ts functions** - Some functions (findMostRecentFile, readFileSafe) may now be unused

## Other Notes

**Final structure of assets/prompts/:**
```
assets/prompts/
├── session-start/
│   ├── critical-rules.txt
│   ├── agent-reminder.txt
│   ├── duplication-guard.txt
│   └── doubt-questions.txt
└── context-injection/
    └── compact-rules.txt
    (skills/commands/agents refs are now programmatic)
```

**Verification commands:**
```bash
npm run lint   # 48 files, no issues
npm run build  # Bundles successfully
npm test       # 9 tests pass
```
