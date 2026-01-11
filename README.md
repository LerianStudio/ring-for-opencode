# Ring for OpenCode

Ring skills library converted for native OpenCode use.

## Installation

1. Copy the `.opencode/` directory to your project root:
   ```bash
   cp -r opencode/.opencode /path/to/your/project/
   ```

2. Copy `opencode.json` to your project root (optional, for customization):
   ```bash
   cp opencode/opencode.json /path/to/your/project/
   ```

3. Copy `AGENTS.md` to your project root:
   ```bash
   cp opencode/AGENTS.md /path/to/your/project/
   ```

## Usage

### Skills
Skills are automatically discovered. Use via the skill tool:
```
skill({ name: "test-driven-development" })
```

### Agents
Agents can be @ mentioned:
```
@code-reviewer review this code
@codebase-explorer analyze the architecture
```

### Commands
Commands are invoked with `/`:
```
/commit
/codereview
/brainstorm
```

## Components

| Type | Count | Description |
|------|-------|-------------|
| **Skills** | 37 | Development workflow skills |
| **Agents** | 14 | Specialized AI agents |
| **Commands** | 19 | Utility commands |
| **Plugins** | 7 | Session management and workflow enhancement |

### Core Skills
- `test-driven-development` - RED-GREEN-REFACTOR methodology
- `brainstorming` - Socratic design refinement
- `requesting-code-review` - Parallel 5-reviewer code review
- `systematic-debugging` - 4-phase debugging methodology
- `executing-plans` - Batch task execution with checkpoints

### Core Agents
- `@code-reviewer` - Code quality, architecture, design patterns
- `@security-reviewer` - Security vulnerabilities, OWASP Top 10
- `@business-logic-reviewer` - Business rules, domain correctness
- `@test-reviewer` - Test quality, coverage, anti-patterns
- `@nil-safety-reviewer` - Nil/null pointer safety (Go and TypeScript)
- `@codebase-explorer` - Deep codebase analysis
- `@write-plan` - Implementation planning

### Dev-Team Agents
- `@backend-engineer-golang` - Go backend development
- `@backend-engineer-typescript` - TypeScript backend development
- `@frontend-engineer` - Frontend development
- `@devops-engineer` - Infrastructure and CI/CD
- `@qa-analyst` - Testing and quality assurance
- `@sre` - Site reliability engineering

## Plugins

| Plugin | Purpose |
|--------|---------|
| `session-start.ts` | Initializes Ring context and injects critical rules |
| `context-injection.ts` | Preserves Ring rules during context compaction |
| `notification.ts` | Desktop notifications (cross-platform) |
| `session-outcome.ts` | Prompts for session outcome grade |
| `task-completion-check.ts` | Notifies when all todos complete |
| `outcome-inference.ts` | Infers session outcome from todo state |
| `doubt-resolver.ts` | Structured choice prompts for ambiguity |

## Platform Compatibility

This conversion adapts Ring's Claude Code format to OpenCode's native format:

| Ring (Claude Code) | OpenCode |
|--------------------|----------|
| `model: opus` | `model: anthropic/claude-opus-4-5-20251101` |
| `model: sonnet` | `model: anthropic/claude-sonnet-4-20250514` |
| `type: reviewer` | `mode: subagent` |
| Bash hooks | TypeScript plugins |
| `TodoWrite` tool | Task tracking / todo management |
| `Skill` tool | Skill invocation |

## Source

Converted from [Ring](https://github.com/LerianStudio/ring) Claude Code plugin.

## License

MIT
