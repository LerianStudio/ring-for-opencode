# SCR (Static Code Reviewer) CLI

Multi-phase static code analysis tool for comprehensive code review preparation.

## Installation

```bash
# Using go install
go install github.com/lerianstudio/ring/scripts/codereview/cmd/scr@latest

# From source
git clone https://github.com/lerianstudio/ring
cd ring/scripts/codereview
make build
```

## Quick Start

```bash
# Run all phases on uncommitted changes
scr run

# Run comparing branches
scr run --base=main --head=HEAD

# Run specific phase only
scr phase scope --base=main --head=HEAD
scr phase static
scr phase ast
```

## Commands

| Command | Description |
|---------|-------------|
| `scr run` | Execute all phases sequentially |
| `scr phase scope` | Phase 0: Detect changed files |
| `scr phase static` | Phase 1: Run linters |
| `scr phase ast` | Phase 2: Extract AST diffs |
| `scr phase callgraph` | Phase 3: Call graph analysis |
| `scr phase dataflow` | Phase 4: Data flow analysis |
| `scr phase context` | Phase 5: Compile review context |
| `scr version` | Print version info |

## Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--output-dir, -o` | Output directory for all artifacts | `.scr` |
| `--verbose, -v` | Enable verbose output | `false` |
| `--base` | Base git reference (commit/branch) | - |
| `--head` | Head git reference (commit/branch) | - |
| `--unstaged` | Analyze only unstaged and untracked files | `false` |
| `--files` | Comma-separated file patterns to analyze | - |
| `--files-from` | Path to file containing file patterns | - |

## Run Command Flags

| Flag | Description |
|------|-------------|
| `--skip` | Comma-separated list of phases to skip |

## Examples

### Analyze Current Branch vs Main

```bash
scr run --base=main --head=HEAD
```

### Analyze Specific Files

```bash
scr run --files=cmd/*.go,internal/**/*.go
```

### Skip Specific Phases

```bash
scr run --skip=static,dataflow
```

### Run Only Scope Detection

```bash
scr phase scope --base=main --head=HEAD
```

### Analyze Unstaged Changes

```bash
scr run --unstaged
```

### Custom Output Directory

```bash
scr run --output-dir=./review-output
```

## Output

All artifacts are written to the output directory (default: `.scr/`):

```
.scr/
├── scope.json              # Changed files and languages
├── static-analysis.json    # Lint findings
├── go-ast.json            # AST diffs (per language)
├── go-calls.json          # Call graph (per language)
├── go-flow.json           # Data flow (per language)
├── security-summary.md     # Security analysis summary
├── impact-summary.md       # Impact analysis summary
├── context-code-reviewer.md
├── context-security-reviewer.md
├── context-business-logic-reviewer.md
├── context-test-reviewer.md
└── context-nil-safety-reviewer.md
```

## Phase Details

### Phase 0: Scope Detection
Analyzes git diffs to detect changed files and determine the primary project language.

### Phase 1: Static Analysis
Runs linters based on detected language:
- **Go**: golangci-lint, staticcheck, gosec
- **TypeScript**: tsc, eslint
- **Python**: ruff, mypy, pylint, bandit

### Phase 2: AST Extraction
Parses source files and extracts structural changes including functions, types, and imports.

### Phase 3: Call Graph Analysis
Builds call graphs to identify:
- Direct callers of modified functions
- Transitive callers
- Affected tests
- Impact analysis

### Phase 4: Data Flow Analysis
Security-focused analysis detecting:
- Untrusted data sources (HTTP inputs, env vars, files)
- Sensitive data sinks (database, exec, response)
- Unsanitized data flows
- Nil/null safety issues

### Phase 5: Context Compilation
Aggregates all phase outputs into reviewer-specific markdown files.

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
scr completion bash > /etc/bash_completion.d/scr

# Zsh
scr completion zsh > "${fpath[1]}/_scr"

# Fish
scr completion fish > ~/.config/fish/completions/scr.fish
```

## Legacy Binaries

For backward compatibility, legacy binaries are still available:

```bash
make build-legacy
```

This builds the original separate binaries:
- `scope-detector`
- `static-analysis`
- `ast-extractor`
- `call-graph`
- `data-flow`
- `compile-context`
- `run-all`
