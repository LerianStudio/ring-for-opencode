# AGENTS.md - Ring for OpenCode

This file provides guidance to OpenCode when working with code in this repository.

---

## Overview

Ring is a comprehensive skills library that enforces proven software engineering practices through mandatory workflows, parallel code review, and systematic development.

## Available Skills

Skills are loaded on-demand via the native `skill` tool. See `assets/skill/` for available skills.

### Core Skills
- `ring:test-driven-development` - RED-GREEN-REFACTOR methodology
- `ring:brainstorming` - Socratic design refinement
- `ring:requesting-code-review` - Parallel 5-reviewer code review
- `ring:executing-plans` - Batch task execution with checkpoints
- `ring:using-pm-team` - Pre-dev planning workflow (4-gate & 9-gate tracks)

### Development Workflow
1. Use `ring:brainstorming` (or `ring:pre-dev-feature`/`ring:pre-dev-full`) to plan
2. Use `ring:test-driven-development` for implementation
3. Use `ring:requesting-code-review` before merging

## Available Agents

Agents are invoked via @ mention or automatically by primary agents.

### Research Agents (subagents)
- `@ring:repo-research-analyst` - Codebase patterns and conventions
- `@ring:best-practices-researcher` - External best practices and industry standards
- `@ring:framework-docs-researcher` - Tech stack documentation and versions

### Reviewers (subagents)
- `@ring:code-reviewer` - Code quality, architecture, design patterns
- `@ring:business-logic-reviewer` - Business rules, domain correctness
- `@ring:security-reviewer` - Security vulnerabilities, data protection
- `@ring:test-reviewer` - Test quality, coverage, anti-patterns
- `@ring:nil-safety-reviewer` - Nil/null pointer safety (Go and TypeScript)

### Specialist Agents (subagents)
- `@ring:codebase-explorer` - Codebase exploration and analysis
- `@ring:write-plan` - Implementation planning

## Available Commands

Commands are invoked via `/ring:command-name`.

- `/ring:commit` - Atomic commits with intelligent grouping
- `/ring:codereview` - Dispatch all 5 reviewers in parallel
- `/ring:brainstorm` - Start design refinement session
- `/ring:pre-dev-feature` - Plan small feature (<2 days, 4 gates)
- `/ring:pre-dev-full` - Plan large feature (≥2 days, 9 gates)

## Compliance Rules

- TDD: Test must fail before implementation
- Review: All 5 reviewers must pass
- Commits: Use conventional commit format

## Key Principles

1. **DRY** - Don't Repeat Yourself
2. **YAGNI** - You Aren't Gonna Need It
3. **TDD** - Test-Driven Development
4. **Frequent commits** - Small, atomic changes
