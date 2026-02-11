# Static Code Reviewer CLI (scr)

Multi-phase static code analysis tool for comprehensive code review preparation.

## Installation

### Using go install

```bash
go install github.com/lerianstudio/ring/scripts/codereview/cmd/scr@latest
```

### Building from source

```bash
git clone https://github.com/lerianstudio/ring.git
cd ring/scripts/codereview
make build
# Binary available at ./bin/scr
```

## Quick Start

```bash
# Run all phases on uncommitted changes
scr run

# Run comparing branches
scr run --base=main --head=HEAD

# Analyze only unstaged changes
scr run --unstaged

# Run specific phase only
scr phase scope --base=main --head=HEAD
scr phase static
scr phase ast

# Skip specific phases
scr run --skip=static,dataflow
```

## Commands

| Command | Description |
|---------|-------------|
| `scr run` | Execute all analysis phases sequentially |
| `scr phase scope` | Detect changed files and project language |
| `scr phase static` | Run linters and static analyzers |
| `scr phase ast` | Extract semantic AST diffs |
| `scr phase callgraph` | Analyze call relationships |
| `scr phase dataflow` | Security-focused data flow analysis |
| `scr phase context` | Compile reviewer-specific context files |
| `scr version` | Print version information |
| `scr completion` | Generate shell completion scripts |

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--output-dir` | `-o` | Output directory for artifacts (default: `.scr`) |
| `--verbose` | `-v` | Enable verbose output |
| `--base` | | Base git reference (commit/branch) |
| `--head` | | Head git reference (commit/branch) |
| `--unstaged` | | Analyze only unstaged and untracked files |
| `--files` | | Comma-separated file patterns to analyze |
| `--files-from` | | Path to file containing patterns (one per line) |

## Run Command Flags

| Flag | Description |
|------|-------------|
| `--skip` | Comma-separated list of phases to skip |

## Output

All artifacts are written to `.scr/` (configurable via `--output-dir`):

```
.scr/
├── scope.json                          # Changed files and detected languages
├── static-analysis.json                # Lint findings from all analyzers
├── go-ast.json                         # AST diffs (per language)
├── go-calls.json                       # Call graph analysis (per language)
├── go-flow.json                        # Data flow analysis (per language)
├── context-code-reviewer.md            # Context for code quality review
├── context-security-reviewer.md        # Context for security review
├── context-business-logic-reviewer.md  # Context for business logic review
├── context-test-reviewer.md            # Context for test coverage review
└── context-nil-safety-reviewer.md      # Context for nil/null safety review
```

## Shell Completion

Generate completion scripts for your shell:

```bash
# Bash (add to ~/.bashrc)
source <(scr completion bash)

# Zsh (add to ~/.zshrc)
source <(scr completion zsh)

# Fish
scr completion fish | source

# PowerShell
scr completion powershell | Out-String | Invoke-Expression
```

To load completions for each session permanently:

```bash
# Bash - Linux
scr completion bash > /etc/bash_completion.d/scr

# Bash - macOS
scr completion bash > $(brew --prefix)/etc/bash_completion.d/scr

# Zsh
scr completion zsh > "${fpath[1]}/_scr"

# Fish
scr completion fish > ~/.config/fish/completions/scr.fish
```

## Phase Execution Flow

```
scope → static → ast → callgraph → dataflow → context
  │        │       │        │           │         │
  │        │       │        │           │         └─> Reviewer context files
  │        │       │        │           └─> Security data flow analysis
  │        │       │        └─> Function call relationships
  │        │       └─> Semantic AST diffs
  │        └─> Linter findings (golangci-lint, eslint, etc.)
  └─> Changed files detection
```

Each phase reads artifacts from previous phases and produces its own artifacts.

## Backward Compatibility

Legacy binaries are still available with deprecation warnings:

- `scope-detector` → `scr phase scope`
- `static-analysis` → `scr phase static`
- `ast-extractor` → `scr phase ast`
- `call-graph` → `scr phase callgraph`
- `data-flow` → `scr phase dataflow`
- `compile-context` → `scr phase context`

These will be removed in a future version. Please migrate to `scr`.

## Development

```bash
# Build all binaries
make build-all

# Run tests
make test

# Run linter
make lint
```

## License

Part of the Ring project by Lerian Studio.
