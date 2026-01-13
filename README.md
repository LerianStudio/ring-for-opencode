# Ring for OpenCode

A comprehensive skills library enforcing proven software engineering practices.

## Quick Start

```bash
# Clone and install globally
git clone https://github.com/LerianStudio/ring-for-opencode.git
cd ring-for-opencode
./installer.sh

# Or install per-project (copies assets/ to your project's .opencode/)
cp -r assets/ /your/project/.opencode/

# Optional: add project Ring config
# Create `.opencode/ring.jsonc` or `.ring/config.jsonc` in your project root
```

## What Ring Provides

| Component | Count | Location |
|-----------|-------|----------|
| Agents | 15 | `assets/agent/` |
| Skills | 23 | `assets/skill/` |
| Commands | 15 | `assets/command/` |
| Plugin | 1 (unified) | `plugin/` |

## Architecture

Ring uses a unified plugin architecture with four core components:

- **Unified Plugin**: Single entry point (`RingUnifiedPlugin`) that injects all components into OpenCode
- **Hook System**: Event-driven middleware for session lifecycle (session start, context injection, task completion)
- **Config Injection**: Automatically loads agents, skills, and commands from `.opencode/` directories

## Core Skills

| Skill | Description |
|-------|-------------|
| `ring:test-driven-development` | RED-GREEN-REFACTOR methodology |
| `ring:requesting-code-review` | Parallel 5-reviewer system |
| `ring:executing-plans` | Batch task execution with checkpoints |
| `ring:brainstorming` | Socratic design refinement |
| `ring:exploring-codebase` | Deep codebase analysis |
| `ring:writing-plans` | Implementation planning with context |
| `ring:handoff-tracking` | Session state preservation |
| `ring:linting-codebase` | Parallel lint fixing |

## Available Agents

### Reviewers

| Agent | Focus |
|-------|-------|
| `@ring:code-reviewer` | Code quality, architecture, design patterns |
| `@ring:business-logic-reviewer` | Business rules, domain correctness |
| `@ring:security-reviewer` | Security vulnerabilities, OWASP Top 10 |
| `@ring:test-reviewer` | Test quality, coverage, anti-patterns |
| `@ring:nil-safety-reviewer` | Nil/null pointer safety |

### Specialists

| Agent | Focus |
|-------|-------|
| `@ring:codebase-explorer` | Deep codebase analysis |
| `@ring:write-plan` | Implementation planning |
| `@ring:backend-engineer-golang` | Go backend development |
| `@ring:backend-engineer-typescript` | TypeScript backend development |
| `@ring:frontend-engineer` | Frontend development |
| `@ring:frontend-bff-engineer-typescript` | BFF/API layer development |
| `@ring:frontend-designer` | UI/UX implementation |
| `@ring:devops-engineer` | Infrastructure and CI/CD |
| `@ring:qa-analyst` | Testing and quality assurance |
| `@ring:sre` | Site reliability engineering |

## Commands

| Command | Description |
|---------|-------------|
| `/ring:commit` | Atomic git commits with intelligent grouping |
| `/ring:codereview` | Parallel 5-reviewer code review |
| `/ring:brainstorm` | Socratic design refinement |
| `/ring:execute-plan` | Batch task execution |
| `/ring:write-plan` | Create implementation plans |
| `/ring:lint` | Run lint checks with auto-fix |
| `/ring:explore-codebase` | Deep codebase analysis |
| `/ring:create-handoff` | Session state handoff |
| `/ring:resume-handoff` | Resume from handoff |
| `/ring:dev-cycle` | Full development cycle |
| `/ring:dev-refactor` | Codebase refactoring |

> **Note**: Ring commands use a flat structure - nested commands (e.g., `/category/command`) are not supported. Commands must be placed directly in `assets/command/` or `.opencode/command/`. If you need hierarchical command organization, use OpenCode's native `.opencode/command/` directory which supports subdirectories.

## Configuration

Ring reads config from `~/.config/opencode/ring/config.jsonc` (user scope) and from project files at `.opencode/ring.jsonc` or `.ring/config.jsonc`.

Example `.opencode/ring.jsonc`:

```jsonc
{
  "$schema": "https://raw.githubusercontent.com/LerianStudio/ring-for-opencode/main/assets/ring-config.schema.json",
  "disabled_agents": []
}
```

Features:
- Schema validation for IDE autocomplete
- Per-agent model configuration
- Permission management
- Plugin configuration

## Directory Structure

```
assets/              # Source assets (installed to user's .opencode/)
├── agent/           # 15 specialized agents
├── command/         # 15 slash commands
├── skill/           # 23 skills with workflows
└── *.schema.json    # JSON schemas for validation

plugin/              # Unified plugin system
├── ring-unified.ts  # Main entry point
├── loaders/         # Component loaders
├── hooks/           # Hook system (session, context, tasks)
├── config/          # Configuration management
├── lifecycle/       # Session lifecycle handlers
├── tools/           # Custom Ring tools
└── utils/           # Shared utilities
```

## Development

```bash
# Install dependencies
bun install

# Build plugin
bun run build

# Lint
bun run lint

# Test
bun test
```

## Source

Converted from Ring Claude Code plugin to OpenCode.

## License

MIT
