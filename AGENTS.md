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

### Development Workflow
1. Use `ring:brainstorming` to refine ideas into designs
2. Use `ring:test-driven-development` for implementation
3. Use `ring:requesting-code-review` before merging

## Available Agents

Agents are invoked via @ mention or automatically by primary agents.

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

## Compliance Rules

- TDD: Test must fail before implementation
- Review: All 5 reviewers must pass
- Commits: Use conventional commit format

## Key Principles

1. **DRY** - Don't Repeat Yourself
2. **YAGNI** - You Aren't Gonna Need It
3. **TDD** - Test-Driven Development
4. **Frequent commits** - Small, atomic changes
