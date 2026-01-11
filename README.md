# Ring for OpenCode

A comprehensive skills library enforcing proven software engineering practices.

## Quick Start

```bash
# Clone and install globally
git clone https://github.com/LerianStudio/ring-for-opencode.git
cd ring-for-opencode
./installer.sh

# Or install per-project
cp -r .opencode/ /your/project/
cp opencode.json /your/project/
```

## What Ring Provides

| Component | Count | Location |
|-----------|-------|----------|
| Agents | 16 | `.opencode/agent/` |
| Skills | 30 | `.opencode/skill/` |
| Commands | 16 | `.opencode/command/` |
| Plugin | 1 (unified) | `plugin/` |

## Architecture

Ring uses a unified plugin architecture with four core components:

- **Unified Plugin**: Single entry point (`RingUnifiedPlugin`) that injects all components into OpenCode
- **Hook System**: Event-driven middleware for session lifecycle (session start, context injection, task completion)
- **Config Injection**: Automatically loads agents, skills, and commands from `.opencode/` directories
- **Background Tasks**: Parallel agent execution manager for concurrent subagent operations

## Core Skills

| Skill | Description |
|-------|-------------|
| `test-driven-development` | RED-GREEN-REFACTOR methodology |
| `requesting-code-review` | Parallel 5-reviewer system |
| `executing-plans` | Batch task execution with checkpoints |
| `systematic-debugging` | 4-phase debugging methodology |
| `brainstorming` | Socratic design refinement |
| `exploring-codebase` | Deep codebase analysis |
| `writing-plans` | Implementation planning with context |
| `dispatching-parallel-agents` | Concurrent agent execution |
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
| `@prompt-quality-reviewer` | Prompt engineering analysis |

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
| `/interview-me` | Requirements gathering |
| `/create-handoff` | Session state handoff |
| `/resume-handoff` | Resume from handoff |
| `/dev-cycle` | Full development cycle |
| `/dev-refactor` | Codebase refactoring |

## Configuration

Example `opencode.json`:

```json
{
  "$schema": ".opencode/schema.json",
  "permissions": {
    "allow": ["Write(*)", "Edit(*)", "Bash(*)"]
  },
  "agents": {
    "code-reviewer": { "model": "anthropic/claude-sonnet-4-20250514" },
    "write-plan": { "model": "anthropic/claude-opus-4-5-20251101" }
  }
}
```

Features:
- Schema validation for IDE autocomplete
- Per-agent model configuration
- Permission management
- Plugin configuration

## Directory Structure

```
.opencode/
├── agent/           # 16 specialized agents
├── command/         # 16 slash commands
├── skill/           # 30 skills with workflows
└── state/           # Runtime state (gitignored)

plugin/              # Unified plugin system
├── ring-unified.ts  # Main entry point
├── loaders/         # Component loaders
├── hooks/           # Hook system (session, context, tasks)
├── config/          # Configuration management
├── lifecycle/       # Session lifecycle handlers
├── background/      # Background task manager
├── tools/           # Custom tools (ring_doubt)
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
