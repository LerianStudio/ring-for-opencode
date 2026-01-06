---
name: using-dev-team
description: |
  Guide to the dev-team plugin - specialist developer agents and 6-gate development cycle.
  Provides backend engineers (Go/TypeScript), frontend engineers, DevOps, SRE, and QA agents.
license: MIT
compatibility:
  platforms:
    - opencode

metadata:
  version: "1.0.0"
  author: "Lerian Studio"
  category: workflow
  trigger:
    - Development tasks requiring specialist implementation
    - Backend, frontend, or infrastructure code changes
    - Need for structured development workflow
---

# Using Dev-Team Plugin

## Overview

The dev-team plugin provides specialist developer agents for structured software development through a 6-gate workflow.

## Available Agents

| Agent | Specialization | Model |
|-------|----------------|-------|
| `backend-engineer-golang` | Go backend services, APIs | opus |
| `backend-engineer-typescript` | TypeScript backend services | opus |
| `frontend-engineer` | React/Next.js frontend | opus |
| `frontend-designer` | UI/UX design, styling | opus |
| `frontend-bff-engineer-typescript` | Backend-for-Frontend | opus |
| `devops-engineer` | Docker, CI/CD, infrastructure | opus |
| `sre` | Observability validation | opus |
| `qa-analyst` | Testing, coverage validation | opus |
| `prompt-quality-reviewer` | AI prompt analysis | opus |

## 6-Gate Development Cycle

| Gate | Name | Skill | Agent |
|------|------|-------|-------|
| 0 | Implementation | dev-implementation | backend-engineer-* / frontend-* |
| 1 | DevOps Setup | dev-devops | devops-engineer |
| 2 | SRE Validation | dev-sre | sre |
| 3 | Testing | dev-testing | qa-analyst |
| 4 | Code Review | requesting-code-review | code-reviewer, security-reviewer, business-logic-reviewer |
| 5 | Validation | dev-validation | User approval |

## Commands

| Command | Description |
|---------|-------------|
| `/dev-cycle` | Execute full 6-gate development cycle |
| `/dev-refactor` | Analyze codebase and generate refactoring tasks |
| `/dev-status` | Check current cycle status |
| `/dev-cancel` | Cancel current cycle |
| `/dev-report` | View feedback report from last cycle |

## Standards

All agents follow standards defined in:
- `dev-team/docs/standards/golang.md` - Go coding standards
- `dev-team/docs/standards/typescript.md` - TypeScript standards
- `dev-team/docs/standards/frontend.md` - Frontend standards
- `dev-team/docs/standards/devops.md` - DevOps standards
- `dev-team/docs/standards/sre.md` - SRE/observability standards

## TDD Requirement

All implementation follows Test-Driven Development:
1. **RED** - Write failing test first
2. **GREEN** - Write minimal code to pass
3. **REFACTOR** - Improve while tests pass

Gate 0 MUST produce test failure output before implementation.

## Usage Pattern

```
1. User provides task requirements
2. /dev-cycle invokes dev-cycle skill
3. dev-implementation dispatches appropriate specialist agent
4. Agent writes tests (RED) then implementation (GREEN)
5. dev-devops creates Docker/compose setup
6. dev-sre validates observability
7. dev-testing ensures coverage threshold
8. requesting-code-review dispatches 3 parallel reviewers
9. dev-validation requests user approval
10. dev-feedback-loop captures metrics for improvement
```

## When to Use

- Backend API development (Go or TypeScript)
- Frontend feature implementation
- Infrastructure setup
- Refactoring existing codebases to Ring standards
- Any task requiring structured development workflow
