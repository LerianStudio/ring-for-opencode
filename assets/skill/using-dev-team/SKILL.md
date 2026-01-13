---
name: "ring:using-dev-team"
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

| Agent | Specialization |
|-------|----------------|
| `backend-engineer-golang` | Go backend services, APIs |
| `backend-engineer-typescript` | TypeScript backend services |
| `frontend-engineer` | React/Next.js frontend |
| `frontend-designer` | UI/UX design, styling |
| `devops-engineer` | Docker, CI/CD, infrastructure |
| `sre` | Observability validation |
| `qa-analyst` | Testing, coverage validation |

## 6-Gate Development Cycle

| Gate | Name | Skill | Agent |
|------|------|-------|-------|
| 0 | Implementation | ring:dev-implementation | backend-engineer-* / frontend-* |
| 1 | DevOps Setup | ring:dev-devops | devops-engineer |
| 2 | SRE Validation | ring:dev-sre | sre |
| 3 | Testing | ring:dev-testing | qa-analyst |
| 4 | Code Review | ring:requesting-code-review | ring:code-reviewer, ring:security-reviewer, ring:business-logic-reviewer |
| 5 | Validation | ring:dev-validation | User approval |

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
2. /dev-cycle invokes ring:dev-cycle skill
3. ring:dev-implementation dispatches appropriate specialist agent
4. Agent writes tests (RED) then implementation (GREEN)
5. ring:dev-devops creates Docker/compose setup
6. ring:dev-sre validates observability
7. ring:dev-testing ensures coverage threshold
8. ring:requesting-code-review dispatches 3 parallel reviewers
9. ring:dev-validation requests user approval
10. ring:dev-feedback-loop captures metrics for improvement
```

## When to Use

- Backend API development (Go or TypeScript)
- Frontend feature implementation
- Infrastructure setup
- Refactoring existing codebases to Ring standards
- Any task requiring structured development workflow
