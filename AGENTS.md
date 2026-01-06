# AGENTS.md - Ring for OpenCode

This file provides guidance to OpenCode when working with code in this repository.

---

## Overview

Ring is a comprehensive skills library that enforces proven software engineering practices through mandatory workflows, parallel code review, and systematic development.

## Available Skills

Skills are loaded on-demand via the native `skill` tool. See `.opencode/skill/` for available skills.

## Interactive Tooling

- `ring_doubt` tool: when you have genuine uncertainty, present options and ask the user to choose.
  See `.opencode/skill/shared-patterns/doubt-resolver.md`.

### Core Skills
- `test-driven-development` - RED-GREEN-REFACTOR methodology
- `brainstorming` - Socratic design refinement
- `requesting-code-review` - Parallel 3-reviewer code review
- `systematic-debugging` - 4-phase debugging methodology
- `executing-plans` - Batch task execution with checkpoints

### Development Workflow
1. Use `brainstorming` to refine ideas into designs
2. Use `test-driven-development` for implementation
3. Use `requesting-code-review` before merging
4. Use `systematic-debugging` when issues arise

## Available Agents

Agents are invoked via @ mention or automatically by primary agents.

### Reviewers (subagents)
- `@code-reviewer` - Code quality, architecture, design patterns
- `@security-reviewer` - Security vulnerabilities, data protection
- `@business-logic-reviewer` - Business rules, domain correctness

### Specialist Agents (subagents)
- `@codebase-explorer` - Codebase exploration and analysis
- `@write-plan` - Implementation planning

## Available Commands

Commands are invoked via `/command-name`.

- `/commit` - Atomic commits with intelligent grouping
- `/codereview` - Dispatch all 3 reviewers in parallel
- `/brainstorm` - Start design refinement session

## Compliance Rules

- TDD: Test must fail before implementation
- Review: All 3 reviewers must pass
- Commits: Use conventional commit format

## Key Principles

1. **DRY** - Don't Repeat Yourself
2. **YAGNI** - You Aren't Gonna Need It
3. **TDD** - Test-Driven Development
4. **Frequent commits** - Small, atomic changes
