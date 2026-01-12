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
| `test-driven-development` | RED-GREEN-REFACTOR methodology |
| `requesting-code-review` | Parallel 5-reviewer system |
| `executing-plans` | Batch task execution with checkpoints |
| `brainstorming` | Socratic design refinement |
| `exploring-codebase` | Deep codebase analysis |
| `writing-plans` | Implementation planning with context |
| `handoff-tracking` | Session state preservation |
| `linting-codebase` | Parallel lint fixing |

## Available Agents

### Reviewers

| Agent | Focus |
|-------|-------|
| `@code-reviewer` | Code quality, architecture, design patterns |
| `@business-logic-reviewer` | Business rules, domain correctness |
| `@security-reviewer` | Security vulnerabilities, OWASP Top 10 |
| `@test-reviewer` | Test quality, coverage, anti-patterns |
| `@nil-safety-reviewer` | Nil/null pointer safety |

### Specialists

| Agent | Focus |
|-------|-------|
| `@codebase-explorer` | Deep codebase analysis |
| `@write-plan` | Implementation planning |
| `@backend-engineer-golang` | Go backend development |
| `@backend-engineer-typescript` | TypeScript backend development |
| `@frontend-engineer` | Frontend development |
| `@frontend-bff-engineer-typescript` | BFF/API layer development |
| `@frontend-designer` | UI/UX implementation |
| `@devops-engineer` | Infrastructure and CI/CD |
| `@qa-analyst` | Testing and quality assurance |
| `@sre` | Site reliability engineering |

## Commands

| Command | Description |
|---------|-------------|
| `/commit` | Atomic git commits with intelligent grouping |
| `/codereview` | Parallel 5-reviewer code review |
| `/brainstorm` | Socratic design refinement |
| `/execute-plan` | Batch task execution |
| `/write-plan` | Create implementation plans |
| `/lint` | Run lint checks with auto-fix |
| `/explore-codebase` | Deep codebase analysis |
| `/create-handoff` | Session state handoff |
| `/resume-handoff` | Resume from handoff |
| `/dev-cycle` | Full development cycle |
| `/dev-refactor` | Codebase refactoring |

> **Note**: Ring commands use a flat structure - nested commands (e.g., `/category/command`) are not supported. Commands must be placed directly in `assets/command/` or `.opencode/command/`. If you need hierarchical command organization, use OpenCode's native `.opencode/command/` directory which supports subdirectories.

## Configuration

Ring reads config from `~/.config/opencode/ring/config.jsonc` (user scope) and from project files at `.opencode/ring.jsonc` or `.ring/config.jsonc`.

Example `.opencode/ring.jsonc`:

```jsonc
{
  "$schema": "https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-config.schema.json",
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

Converted from [Ring](https://github.com/LerianStudio/ring) Claude Code plugin.

## License

MIT
