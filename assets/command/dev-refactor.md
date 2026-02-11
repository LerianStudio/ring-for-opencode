---
name: "ring:dev-refactor"
description: Analyze existing codebase against standards and generate refactoring tasks for ring:dev-cycle
agent: build
---

Analyze existing codebase against standards and execute refactoring through ring:dev-cycle.

## ⛔ PRE-EXECUTION CHECK (EXECUTE FIRST)

**Before loading the skill, you MUST check:**

```
Does docs/PROJECT_RULES.md exist in the target project?
├── YES → Load skill: ring:dev-refactor
└── no  → Output blocker below and STOP
```

**If file does not exist, output this EXACT response:**

```markdown
## ⛔ HARD BLOCK: PROJECT_RULES.md Not Found

**Status:** BLOCKED - Cannot proceed

### Required Action
Create `docs/PROJECT_RULES.md` with your project's:
- Architecture patterns
- Code conventions
- Testing requirements
- DevOps standards

Then re-run `/ring:dev-refactor`.
```

**DO not:**
- Use "default" or "industry" standards
- Infer standards from existing code
- Proceed with partial analysis
- Offer to create the file

---

## Usage

```
/ring:dev-refactor [path] [options] [prompt]
/ring:dev-refactor [prompt]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `path` | No | Directory to analyze (default: current project root) |
| `prompt` | No | Direct instruction for refactoring focus (no quotes needed) |

## Options

| Option | Description | Example |
|--------|-------------|---------|
| `--standards PATH` | Custom standards file | `--standards docs/MY_PROJECT_RULES.md` |
| `--analyze-only` | Generate report without executing | `--analyze-only` |
| `--critical-only` | Limit execution/output to Critical and High (analysis still tracks all) | `--critical-only` |
| `--dry-run` | Show what would be analyzed | `--dry-run` |

## Examples

```bash
# Direct prompt - focus refactoring on specific area
/ring:dev-refactor Focus on multi-tenant patterns and ensure all repositories use poolmanager

# Direct prompt - another example
/ring:dev-refactor Check idempotency implementation against standards

# Analyze entire project and refactor
/ring:dev-refactor

# Analyze specific directory
/ring:dev-refactor src/domain

# Analyze with specific focus
/ring:dev-refactor src/domain Focus on error handling patterns

# Analysis only (no execution)
/ring:dev-refactor --analyze-only

# Only fix critical issues
/ring:dev-refactor --critical-only

# Use custom standards with focus
/ring:dev-refactor --standards docs/team-standards.md Prioritize security issues
```

## Workflow

**See skill `ring:dev-refactor` for the complete 13-step workflow with TodoWrite template.**

The skill defines all steps including: stack detection, ring:codebase-explorer dispatch, individual agent reports, finding mapping, and artifact generation.

## Analysis Dimensions

**⛔ All five dimensions are MANDATORY for analysis/tracking. The `--critical-only` flag filters execution/output only.**

| Dimension | What's Checked | Standards Reference |
|-----------|----------------|---------------------|
| **Architecture** | DDD patterns, layer separation, dependency direction, directory structure | `golang.md` § Architecture |
| **Code Quality** | Naming conventions, error handling, forbidden practices, security | `golang.md` § Error Handling |
| **Instrumentation** | Service method tracing, span naming, error classification, context propagation | `golang.md` § Distributed Tracing |
| **Testing** | Coverage percentage, test patterns, naming, missing tests | `golang.md` § Testing |
| **DevOps** | Dockerfile, docker-compose, env management, Helm charts | `golang.md` § DevOps |

**Analysis vs Execution:**
- **Analysis (always):** All five dimensions analyzed, all severities (Critical, High, Medium, Low) tracked
- **Execution (filterable):** `--critical-only` limits execution/prioritization to Critical and High severity issues

Example: `/ring:dev-refactor --critical-only` analyzes all issues but only executes fixes for Critical and High.

### Instrumentation Checklist (Quick Reference)

When analyzing services for instrumentation compliance, verify:

1. **Context extraction**: `libCommons.NewTrackingFromContext(ctx)` at method start
2. **Child span creation**: `tracer.Start(ctx, "layer.entity.operation")` with proper naming
3. **Span cleanup**: `defer span.End()` immediately after span creation
4. **Error classification**:
   - Business errors → `HandleSpanBusinessErrorEvent` (span stays OK)
   - Technical errors → `HandleSpanError` (span marked ERROR)
5. **Structured logging**: Use logger from context, not `log.Printf`

**Full details and code templates**: See `docs/standards/golang.md` § "Distributed Tracing Architecture"

## Output

**Timestamp format:** `{timestamp}` = `YYYY-MM-DDTHH:MM:SS` (e.g., `2026-02-07T22:30:45`)

**Analysis Report** (`docs/ring:dev-refactor/{timestamp}/analysis-report.md`):
- Summary table with issue counts by severity
- Detailed findings grouped by dimension
- Specific file locations and line numbers

**Tasks File** (`docs/ring:dev-refactor/{timestamp}/tasks.md`):
- Grouped refactoring tasks (REFACTOR-001, REFACTOR-002, etc.)
- Same format as PM Team output
- Compatible with ring:dev-cycle execution

## Severity Levels (all ARE MANDATORY)

| Level | Description | Priority | Tracking |
|-------|-------------|----------|----------|
| **Critical** | Security vulnerabilities, data loss risk | Fix immediately | **MANDATORY** |
| **High** | Architecture violations, major code smells | Fix in current sprint | **MANDATORY** |
| **Medium** | Convention violations, moderate gaps | Fix in next sprint | **MANDATORY** |
| **Low** | Style issues, minor gaps | Fix when capacity | **MANDATORY** |

**⛔ all severities are MANDATORY to track and fix. Low ≠ Optional. Low = Lower priority, still required.**

## Prerequisites

1. **⛔ PROJECT_RULES.md (MANDATORY)**: `docs/PROJECT_RULES.md` MUST exist - no defaults, no fallback
2. **Git repository**: Project should be under version control
3. **Readable codebase**: Access to source files

**If PROJECT_RULES.md does not exist:** This command will output a blocker message and terminate. The project owner must create the file first.

## Related Commands

| Command | Description |
|---------|-------------|
| `/ring:dev-cycle` | Execute development cycle (used after analysis) |
| `/ring:pre-dev-feature` | Plan new features (use instead for greenfield) |
| `/ring:codereview` | Manual code review (ring:dev-cycle includes this) |

---

## ⛔ MANDATORY: Load Full Skill

**After PROJECT_RULES.md check passes, load the skill:**

```
Use Skill tool: ring:dev-refactor
```

The skill contains the complete analysis workflow with:
- Anti-rationalization tables for codebase exploration
- Mandatory use of `ring:codebase-explorer` (not Bash/Explore)
- Standards coverage table requirements
- Finding → Task mapping gates
- Full agent dispatch prompts with `**MODE: ANALYSIS only**`

## Execution Context

Pass the following context to the skill:

| Parameter | Value |
|-----------|-------|
| `path` | First argument if it's a directory path (default: project root) |
| `prompt` | Remaining text after path and options (direct instruction for focus) |
| `--standards` | If provided, custom standards file path |
| `--analyze-only` | If provided, skip ring:dev-cycle execution |
| `--critical-only` | If provided, filter to Critical/High only |
| `--dry-run` | If provided, show what would be analyzed |

**Argument Parsing:**
- If first argument is a path (contains `/` or is a known directory) → treat as path
- If first argument is an option (`--*`) → no path specified
- Remaining non-option text → prompt (refactoring focus)

## User Approval (MANDATORY)

**Before executing ring:dev-cycle, you MUST ask:**

```yaml
AskUserQuestion:
  questions:
    - question: "Review refactoring plan. How to proceed?"
      header: "Approval"
      options:
        - label: "Approve all"
          description: "Proceed to ring:dev-cycle execution"
        - label: "Critical only"
          description: "Execute only Critical/High tasks"
        - label: "Cancel"
          description: "Keep analysis, skip execution"
```

## Quick Reference

See skill `ring:dev-refactor` for full details. Key rules:

- **All agents dispatch in parallel** - Single message, multiple Task calls
- **Specify model: "opus"** - All agents need opus for comprehensive analysis
- **MODE: ANALYSIS only** - Agents analyze, they DO NOT implement
- **Save artifacts** to `docs/ring:dev-refactor/{timestamp}/`
- **Get user approval** before executing ring:dev-cycle
- **Handoff**: `/ring:dev-cycle docs/ring:dev-refactor/{timestamp}/tasks.md`
