# Unified CLI Implementation Plan

**Project:** static-code-reviewer-cli  
**Module:** github.com/lerianstudio/static-code-reviewer-cli  
**Binary:** `scr` (static-code-reviewer)

## Executive Summary

Transform 6 separate phase binaries + orchestrator into a single unified CLI with subcommands, making it easier for open source users to install and use.

---

## Phase 1: Foundation Setup (Day 1 - Morning)

### Task 1.1: Add Cobra Dependency

```bash
go get github.com/spf13/cobra@latest
```

**Files to modify:**
- `go.mod` - Add cobra dependency

### Task 1.2: Create CLI Directory Structure

```
cmd/
  scr/                          # NEW: Unified binary
    main.go                     # Entry point
    root.go                     # Root command + shared flags
    run.go                      # scr run (orchestrator)
    phase.go                    # scr phase parent command
    phase_scope.go              # scr phase scope
    phase_static.go             # scr phase static
    phase_ast.go                # scr phase ast
    phase_callgraph.go          # scr phase callgraph
    phase_dataflow.go           # scr phase dataflow
    phase_context.go            # scr phase context
    version.go                  # scr version
  scope-detector/               # KEEP: Backward compat wrapper
  static-analysis/              # KEEP: Backward compat wrapper
  ast-extractor/                # KEEP: Backward compat wrapper
  call-graph/                   # KEEP: Backward compat wrapper
  data-flow/                    # KEEP: Backward compat wrapper
  compile-context/              # KEEP: Backward compat wrapper
  run-all/                      # DEPRECATE: Replaced by scr run
```

### Task 1.3: Create Shared CLI Config Package

**New file:** `internal/cli/config.go`

```go
package cli

import "time"

// Config holds shared configuration for all commands.
type Config struct {
    // Common flags
    WorkDir   string
    OutputDir string
    Verbose   bool
    
    // Git refs
    BaseRef   string
    HeadRef   string
    Unstaged  bool
    
    // Execution control
    Timeout   time.Duration
    Skip      []string
    
    // Output control
    JSONSummary string
    FailFast    bool
}

// DefaultConfig returns config with sensible defaults.
func DefaultConfig() *Config {
    return &Config{
        OutputDir: ".scr",
        Timeout:   10 * time.Minute,
    }
}
```

---

## Phase 2: Phase Libraries (Day 1 - Afternoon)

Convert each phase from `main.go` logic to library packages with `Run(ctx, cfg) error`.

### Task 2.1: Create Phase Interface

**New file:** `internal/phases/phase.go`

```go
package phases

import (
    "context"
    "time"
)

// Phase represents a single analysis phase.
type Phase interface {
    Name() string
    Run(ctx context.Context, cfg *Config) error
    Timeout() time.Duration
}

// Config holds phase-specific configuration derived from CLI flags.
type Config struct {
    WorkDir     string
    OutputDir   string
    Verbose     bool
    BaseRef     string
    HeadRef     string
    Unstaged    bool
    Files       []string
    ScriptsDir  string
    
    // Phase-specific
    ScopePath   string   // Path to scope.json
    ASTPath     string   // Path to AST JSON
    Language    string   // Language override
    BatchFile   string   // Batch file for AST
}
```

### Task 2.2: Extract Scope Phase Library

**New file:** `internal/phases/scope/run.go`

Extract logic from `cmd/scope-detector/main.go` → `run()` function:

```go
package scope

import (
    "context"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
)

type Phase struct{}

func New() *Phase { return &Phase{} }

func (p *Phase) Name() string { return "scope" }

func (p *Phase) Timeout() time.Duration { return 2 * time.Minute }

func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
    // Move logic from cmd/scope-detector/main.go run() here
    // Key changes:
    // - Replace flag.* with cfg.*
    // - Use cfg.OutputDir for output path
    // - Honor ctx for cancellation
}
```

### Task 2.3: Extract Static Analysis Phase Library

**New file:** `internal/phases/static/run.go`

```go
package static

import (
    "context"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
)

type Phase struct{}

func New() *Phase { return &Phase{} }

func (p *Phase) Name() string { return "static-analysis" }

func (p *Phase) Timeout() time.Duration { return 5 * time.Minute }

func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
    // Move logic from cmd/static-analysis/main.go run() here
}
```

### Task 2.4: Extract AST Phase Library

**New file:** `internal/phases/ast/run.go`

```go
package astphase

import (
    "context"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
)

type Phase struct{}

func New() *Phase { return &Phase{} }

func (p *Phase) Name() string { return "ast" }

func (p *Phase) Timeout() time.Duration { return 3 * time.Minute }

func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
    // Move logic from cmd/ast-extractor/main.go run() here
}
```

### Task 2.5: Extract Call Graph Phase Library

**New file:** `internal/phases/callgraph/run.go`

```go
package callgraphphase

import (
    "context"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
)

type Phase struct{}

func New() *Phase { return &Phase{} }

func (p *Phase) Name() string { return "callgraph" }

func (p *Phase) Timeout() time.Duration { return 2 * time.Minute }

func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
    // Move logic from cmd/call-graph/main.go run() here
}
```

### Task 2.6: Extract Data Flow Phase Library

**New file:** `internal/phases/dataflow/run.go`

```go
package dataflowphase

import (
    "context"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
)

type Phase struct{}

func New() *Phase { return &Phase{} }

func (p *Phase) Name() string { return "dataflow" }

func (p *Phase) Timeout() time.Duration { return 3 * time.Minute }

func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
    // Move logic from cmd/data-flow/main.go run() here
}
```

### Task 2.7: Extract Context Phase Library

**New file:** `internal/phases/context/run.go`

```go
package contextphase

import (
    "context"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
)

type Phase struct{}

func New() *Phase { return &Phase{} }

func (p *Phase) Name() string { return "context" }

func (p *Phase) Timeout() time.Duration { return 1 * time.Minute }

func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
    // Move logic from cmd/compile-context/main.go run() here
}
```

---

## Phase 3: Pipeline Runner (Day 2 - Morning)

### Task 3.1: Create In-Process Pipeline Runner

**New file:** `internal/pipeline/pipeline.go`

```go
package pipeline

import (
    "context"
    "time"
    
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases/scope"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases/static"
    astphase "github.com/lerianstudio/static-code-reviewer-cli/internal/phases/ast"
    callgraphphase "github.com/lerianstudio/static-code-reviewer-cli/internal/phases/callgraph"
    dataflowphase "github.com/lerianstudio/static-code-reviewer-cli/internal/phases/dataflow"
    contextphase "github.com/lerianstudio/static-code-reviewer-cli/internal/phases/context"
)

// PhaseResult holds the outcome of a phase execution.
type PhaseResult struct {
    Name       string
    Success    bool
    Skipped    bool
    SkipReason string
    Duration   time.Duration
    Error      error
}

// Options configures pipeline execution.
type Options struct {
    Skip       map[string]bool
    FailFast   bool
    JSONOutput string
}

// DefaultPhases returns the ordered list of all phases.
func DefaultPhases() []phases.Phase {
    return []phases.Phase{
        scope.New(),
        static.New(),
        astphase.New(),
        callgraphphase.New(),
        dataflowphase.New(),
        contextphase.New(),
    }
}

// Run executes all phases in sequence with timeout handling.
func Run(ctx context.Context, cfg *phases.Config, opts Options) ([]PhaseResult, error) {
    allPhases := DefaultPhases()
    results := make([]PhaseResult, 0, len(allPhases))
    
    for _, phase := range allPhases {
        // Check skip list
        if opts.Skip[phase.Name()] {
            results = append(results, PhaseResult{
                Name:       phase.Name(),
                Skipped:    true,
                Success:    true,
                SkipReason: "skipped via --skip flag",
            })
            continue
        }
        
        // Execute with timeout
        phaseCtx, cancel := context.WithTimeout(ctx, phase.Timeout())
        start := time.Now()
        err := phase.Run(phaseCtx, cfg)
        cancel()
        
        result := PhaseResult{
            Name:     phase.Name(),
            Duration: time.Since(start),
            Success:  err == nil,
            Error:    err,
        }
        results = append(results, result)
        
        // Handle failure
        if err != nil && opts.FailFast {
            return results, err
        }
    }
    
    return results, nil
}
```

### Task 3.2: Create Artifacts Package

**New file:** `internal/artifacts/paths.go`

```go
package artifacts

import "path/filepath"

const (
    DefaultOutputDir = ".scr"
)

// Paths provides canonical artifact paths.
type Paths struct {
    OutputDir string
}

func New(outputDir string) *Paths {
    if outputDir == "" {
        outputDir = DefaultOutputDir
    }
    return &Paths{OutputDir: outputDir}
}

func (p *Paths) Scope() string         { return filepath.Join(p.OutputDir, "scope.json") }
func (p *Paths) StaticAnalysis() string { return filepath.Join(p.OutputDir, "static-analysis.json") }
func (p *Paths) AST(lang string) string { return filepath.Join(p.OutputDir, lang+"-ast.json") }
func (p *Paths) ASTBatch() string       { return filepath.Join(p.OutputDir, "ast-batch.json") }
func (p *Paths) CallGraph(lang string) string { return filepath.Join(p.OutputDir, lang+"-calls.json") }
func (p *Paths) DataFlow(lang string) string  { return filepath.Join(p.OutputDir, lang+"-flow.json") }
func (p *Paths) Context(reviewer string) string {
    return filepath.Join(p.OutputDir, "context-"+reviewer+".md")
}
func (p *Paths) Summary() string { return filepath.Join(p.OutputDir, "summary.json") }
```

---

## Phase 4: Cobra Commands (Day 2 - Afternoon)

### Task 4.1: Root Command

**New file:** `cmd/scr/root.go`

```go
package main

import (
    "os"
    
    "github.com/spf13/cobra"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/cli"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/logger"
)

var cfg = cli.DefaultConfig()

var rootCmd = &cobra.Command{
    Use:   "scr",
    Short: "Static Code Reviewer - Multi-phase code analysis tool",
    Long: `Static Code Reviewer (scr) analyzes code changes through multiple phases:

  Phase 0 (scope):    Detect changed files and languages
  Phase 1 (static):   Run static analysis / linters
  Phase 2 (ast):      Extract semantic AST diffs
  Phase 3 (callgraph): Analyze call relationships
  Phase 4 (dataflow): Security-focused data flow analysis
  Phase 5 (context):  Compile reviewer-specific context

Use 'scr run' to execute all phases, or 'scr phase <name>' for individual phases.`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        if cfg.Verbose {
            logger.SetDefault(logger.NewLogger(logger.WithLevel(logger.LevelDebug)))
        }
        if cfg.WorkDir == "" {
            wd, err := os.Getwd()
            if err != nil {
                return err
            }
            cfg.WorkDir = wd
        }
        return nil
    },
}

func init() {
    rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")
    rootCmd.PersistentFlags().StringVar(&cfg.OutputDir, "output-dir", ".scr", "Output directory for artifacts")
    rootCmd.PersistentFlags().StringVar(&cfg.WorkDir, "workdir", "", "Working directory (default: current)")
    rootCmd.PersistentFlags().StringVar(&cfg.BaseRef, "base", "", "Base git ref for comparison")
    rootCmd.PersistentFlags().StringVar(&cfg.HeadRef, "head", "", "Head git ref for comparison")
    rootCmd.PersistentFlags().BoolVar(&cfg.Unstaged, "unstaged", false, "Analyze only unstaged changes")
}
```

### Task 4.2: Run Command (Orchestrator)

**New file:** `cmd/scr/run.go`

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"
    
    "github.com/spf13/cobra"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/pipeline"
)

var skipPhases string

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Run all analysis phases",
    Long: `Execute all analysis phases in sequence with timeout handling and graceful degradation.

Phases are executed in order:
  0. scope         - Detect changed files and languages
  1. static        - Run static analysis / linters  
  2. ast           - Extract semantic AST diffs
  3. callgraph     - Analyze call relationships
  4. dataflow      - Security-focused data flow analysis
  5. context       - Compile reviewer-specific context`,
    Example: `  # Run all phases on uncommitted changes
  scr run

  # Run comparing branches
  scr run --base=main --head=HEAD

  # Skip specific phases
  scr run --skip=callgraph,dataflow

  # Custom output directory
  scr run --output-dir=./analysis`,
    RunE: runAllPhases,
}

func init() {
    runCmd.Flags().StringVar(&skipPhases, "skip", "", "Comma-separated phases to skip (scope,static,ast,callgraph,dataflow,context)")
    runCmd.Flags().DurationVar(&cfg.Timeout, "timeout", 10*time.Minute, "Overall timeout for all phases")
    runCmd.Flags().StringVar(&cfg.JSONSummary, "json-summary", "", "Path to write JSON summary")
    runCmd.Flags().BoolVar(&cfg.FailFast, "fail-fast", false, "Stop on first phase failure")
    
    rootCmd.AddCommand(runCmd)
}

func runAllPhases(cmd *cobra.Command, args []string) error {
    // Setup signal handling
    ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
    defer cancel()
    
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigCh
        cancel()
    }()
    
    // Parse skip list
    skipSet := make(map[string]bool)
    if skipPhases != "" {
        for _, s := range strings.Split(skipPhases, ",") {
            skipSet[strings.TrimSpace(s)] = true
        }
    }
    
    // Build phase config
    phaseCfg := &phases.Config{
        WorkDir:   cfg.WorkDir,
        OutputDir: cfg.OutputDir,
        Verbose:   cfg.Verbose,
        BaseRef:   cfg.BaseRef,
        HeadRef:   cfg.HeadRef,
        Unstaged:  cfg.Unstaged,
    }
    
    // Run pipeline
    opts := pipeline.Options{
        Skip:     skipSet,
        FailFast: cfg.FailFast,
    }
    
    results, err := pipeline.Run(ctx, phaseCfg, opts)
    
    // Print summary
    printPipelineSummary(results)
    
    // Write JSON summary if requested
    if cfg.JSONSummary != "" {
        if writeErr := writeSummary(cfg.JSONSummary, results); writeErr != nil {
            fmt.Fprintf(os.Stderr, "Warning: failed to write summary: %v\n", writeErr)
        }
    }
    
    return err
}

func printPipelineSummary(results []pipeline.PhaseResult) {
    // ... (similar to current run-all printSummary)
}
```

### Task 4.3: Phase Parent Command

**New file:** `cmd/scr/phase.go`

```go
package main

import "github.com/spf13/cobra"

var phaseCmd = &cobra.Command{
    Use:   "phase",
    Short: "Run individual analysis phases",
    Long: `Run a specific analysis phase independently.

Available phases:
  scope      - Phase 0: Detect changed files and languages
  static     - Phase 1: Run static analysis / linters
  ast        - Phase 2: Extract semantic AST diffs
  callgraph  - Phase 3: Analyze call relationships
  dataflow   - Phase 4: Security-focused data flow analysis
  context    - Phase 5: Compile reviewer-specific context`,
}

func init() {
    rootCmd.AddCommand(phaseCmd)
}
```

### Task 4.4: Individual Phase Commands

**New file:** `cmd/scr/phase_scope.go`

```go
package main

import (
    "context"
    "time"
    
    "github.com/spf13/cobra"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
    scopephase "github.com/lerianstudio/static-code-reviewer-cli/internal/phases/scope"
)

var scopeFiles string
var scopeFilesFrom string

var phaseScopeCmd = &cobra.Command{
    Use:   "scope",
    Short: "Phase 0: Detect changed files and languages",
    Long: `Analyzes git diff to detect changed files and project language.

Outputs scope.json with:
  - List of modified, added, and deleted files
  - Detected primary language
  - Package/module information`,
    Example: `  # Detect all uncommitted changes
  scr phase scope

  # Compare branches
  scr phase scope --base=main --head=HEAD

  # Analyze specific files
  scr phase scope --files=cmd/*.go,internal/**/*.go`,
    RunE: runScopePhase,
}

func init() {
    phaseScopeCmd.Flags().StringVar(&scopeFiles, "files", "", "Comma-separated file patterns")
    phaseScopeCmd.Flags().StringVar(&scopeFilesFrom, "files-from", "", "File containing patterns (one per line)")
    
    phaseCmd.AddCommand(phaseScopeCmd)
}

func runScopePhase(cmd *cobra.Command, args []string) error {
    phase := scopephase.New()
    
    ctx, cancel := context.WithTimeout(context.Background(), phase.Timeout())
    defer cancel()
    
    phaseCfg := &phases.Config{
        WorkDir:   cfg.WorkDir,
        OutputDir: cfg.OutputDir,
        Verbose:   cfg.Verbose,
        BaseRef:   cfg.BaseRef,
        HeadRef:   cfg.HeadRef,
        Unstaged:  cfg.Unstaged,
        // Phase-specific
        FilesPattern:   scopeFiles,
        FilesFromPath:  scopeFilesFrom,
    }
    
    return phase.Run(ctx, phaseCfg)
}
```

Similar files for other phases:
- `cmd/scr/phase_static.go`
- `cmd/scr/phase_ast.go`
- `cmd/scr/phase_callgraph.go`
- `cmd/scr/phase_dataflow.go`
- `cmd/scr/phase_context.go`

### Task 4.5: Version Command

**New file:** `cmd/scr/version.go`

```go
package main

import (
    "fmt"
    "runtime"
    
    "github.com/spf13/cobra"
)

var (
    Version   = "dev"
    GitCommit = "unknown"
    BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("scr version %s\n", Version)
        fmt.Printf("  commit:  %s\n", GitCommit)
        fmt.Printf("  built:   %s\n", BuildDate)
        fmt.Printf("  go:      %s\n", runtime.Version())
        fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

### Task 4.6: Main Entry Point

**New file:** `cmd/scr/main.go`

```go
package main

import (
    "os"
    
    "github.com/lerianstudio/static-code-reviewer-cli/internal/recovery"
)

func main() {
    os.Exit(recovery.WrapMain(func() {
        if err := rootCmd.Execute(); err != nil {
            os.Exit(1)
        }
    }))
}
```

---

## Phase 5: Backward Compatibility Wrappers (Day 3 - Morning)

### Task 5.1: Update Existing Binaries as Wrappers

**Modified file:** `cmd/scope-detector/main.go`

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    
    "github.com/lerianstudio/static-code-reviewer-cli/internal/phases"
    scopephase "github.com/lerianstudio/static-code-reviewer-cli/internal/phases/scope"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/logger"
    "github.com/lerianstudio/static-code-reviewer-cli/internal/recovery"
)

// Backward-compatible wrapper that calls the phase library directly.

var (
    baseRef     = flag.String("base", "", "Base reference")
    headRef     = flag.String("head", "", "Head reference")
    filesFlag   = flag.String("files", "", "Comma-separated file patterns")
    filesFrom   = flag.String("files-from", "", "Path to file with patterns")
    unstaged    = flag.Bool("unstaged", false, "Analyze unstaged changes only")
    outputPath  = flag.String("output", "", "Output file path")
    workDir     = flag.String("workdir", "", "Working directory")
    verbose     = flag.Bool("v", false, "Verbose output")
)

func main() {
    os.Exit(recovery.WrapMain(realMain))
}

func realMain() {
    flag.Parse()
    
    if *verbose {
        logger.SetDefault(logger.NewLogger(logger.WithLevel(logger.LevelDebug)))
    }
    
    wd := *workDir
    if wd == "" {
        var err error
        wd, err = os.Getwd()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    }
    
    phase := scopephase.New()
    ctx, cancel := context.WithTimeout(context.Background(), phase.Timeout())
    defer cancel()
    
    cfg := &phases.Config{
        WorkDir:       wd,
        OutputDir:     outputDirFromPath(*outputPath),
        Verbose:       *verbose,
        BaseRef:       *baseRef,
        HeadRef:       *headRef,
        Unstaged:      *unstaged,
        FilesPattern:  *filesFlag,
        FilesFromPath: *filesFrom,
    }
    
    if err := phase.Run(ctx, cfg); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func outputDirFromPath(path string) string {
    if path == "" {
        return ".scr"
    }
    return filepath.Dir(path)
}
```

Similar updates for all other `cmd/<phase>/main.go` files.

---

## Phase 6: Module Migration (Day 3 - Afternoon)

### Task 6.1: Update go.mod

```bash
# Change module path
go mod edit -module github.com/lerianstudio/static-code-reviewer-cli

# Update all imports
find . -name "*.go" -exec sed -i '' \
  's|github.com/lerianstudio/ring/scripts/codereview|github.com/lerianstudio/static-code-reviewer-cli|g' {} \;
```

### Task 6.2: Update Makefile

**Modified file:** `Makefile`

```makefile
MODULE := github.com/lerianstudio/static-code-reviewer-cli
BINARY := scr
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/scr

.PHONY: build-all
build-all: build
	# Build backward-compat wrappers
	go build -o bin/scope-detector ./cmd/scope-detector
	go build -o bin/static-analysis ./cmd/static-analysis
	go build -o bin/ast-extractor ./cmd/ast-extractor
	go build -o bin/call-graph ./cmd/call-graph
	go build -o bin/data-flow ./cmd/data-flow
	go build -o bin/compile-context ./cmd/compile-context

.PHONY: install
install:
	go install $(LDFLAGS) ./cmd/scr

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: lint
lint:
	golangci-lint run ./...
```

### Task 6.3: Update build-release.sh

```bash
#!/bin/bash
set -euo pipefail

VERSION="${1:-dev}"
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${PLATFORMS[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"
    
    output="dist/scr-${VERSION}-${GOOS}-${GOARCH}"
    [[ "$GOOS" == "windows" ]] && output+=".exe"
    
    echo "Building $output..."
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.Version=$VERSION" \
        -o "$output" \
        ./cmd/scr
done

# Generate checksums
cd dist && sha256sum scr-* > checksums.txt
```

---

## Phase 7: Documentation & Testing (Day 4)

### Task 7.1: Update README.md

```markdown
# Static Code Reviewer CLI

Multi-phase static code analysis tool for comprehensive code review preparation.

## Installation

```bash
# Using go install
go install github.com/lerianstudio/static-code-reviewer-cli/cmd/scr@latest

# Using Homebrew (coming soon)
brew install lerianstudio/tap/scr
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

## Output

All artifacts are written to `.scr/` (configurable via `--output-dir`):

```
.scr/
├── scope.json              # Changed files and languages
├── static-analysis.json    # Lint findings
├── go-ast.json            # AST diffs (per language)
├── go-calls.json          # Call graph (per language)
├── go-flow.json           # Data flow (per language)
├── context-code-reviewer.md
├── context-security-reviewer.md
├── context-business-logic-reviewer.md
├── context-test-reviewer.md
└── context-nil-safety-reviewer.md
```
```

### Task 7.2: Add Integration Tests

**New file:** `integration_test.go` (update existing)

```go
func TestUnifiedCLI(t *testing.T) {
    // Test scr run
    cmd := exec.Command("./bin/scr", "run", "--output-dir=testdata/output")
    // ...
    
    // Test scr phase scope
    cmd = exec.Command("./bin/scr", "phase", "scope", "--base=HEAD~1")
    // ...
}

func TestBackwardCompatibility(t *testing.T) {
    // Test old binary names still work
    cmd := exec.Command("./bin/scope-detector", "--base=HEAD~1")
    // ...
}
```

---

## Implementation Checklist

### Day 1 ✅ COMPLETE
- [x] Add Cobra dependency to go.mod
- [x] Create `internal/cli/config.go`
- [x] Create `internal/phases/phase.go` interface
- [x] Extract `internal/phases/scope/run.go`
- [x] Extract `internal/phases/static/run.go`
- [x] Extract `internal/phases/ast/run.go`
- [x] Extract `internal/phases/callgraph/run.go`
- [x] Extract `internal/phases/dataflow/run.go`
- [x] Extract `internal/phases/context/run.go`

### Day 2 ✅ COMPLETE
- [x] Create `internal/pipeline/pipeline.go`
- [x] Create `internal/artifacts/paths.go`
- [x] Create `cmd/scr/main.go`
- [x] Create `cmd/scr/root.go`
- [x] Create `cmd/scr/run.go`
- [x] Create `cmd/scr/phase.go`
- [x] Create `cmd/scr/phase_*.go` (6 files)
- [x] Create `cmd/scr/version.go`
- [x] Update `cmd/scope-detector/main.go` as wrapper (with deprecation warning)
- [x] Update `cmd/static-analysis/main.go` as wrapper (with deprecation warning)
- [x] Update `cmd/ast-extractor/main.go` as wrapper (with deprecation warning)
- [x] Update `cmd/call-graph/main.go` as wrapper (with deprecation warning)
- [x] Update `cmd/data-flow/main.go` as wrapper (with deprecation warning)
- [x] Update `cmd/compile-context/main.go` as wrapper (with deprecation warning)
- [x] Update Makefile

### Day 3 ✅ COMPLETE
- [x] Add tests for phase libraries (`internal/phases/*`)
- [x] Add tests for pipeline (`internal/pipeline/*`)
- [x] Add tests for unified CLI (`cmd/scr`) via integration tests
- [x] Update integration tests for backward compatibility
- [ ] Change module path in go.mod (optional - defer to separate PR)
- [x] Update build-release.sh

### Day 4 ✅ COMPLETE
- [x] Update README.md
- [x] Add shell completion support (cmd/scr/completion.go)
- [x] Test backward compatibility end-to-end (via integration tests)
- [x] Generate release binaries (VERSION=v1.0.0 ./build-release.sh)
- [ ] Create GitHub release (manual step)

---

## Risk Mitigation

1. **Phased rollout**: Keep old binaries working via wrappers
2. **Golden tests**: Add output fixture tests before refactoring
3. **Feature flags**: Allow falling back to exec-based pipeline if needed
4. **Deprecation warnings**: Old binaries print deprecation notice to stderr

---

## Future Enhancements

1. **Config file support**: `.scr.yaml` for persistent settings
2. **Parallel phases**: Run independent phases concurrently
3. **Plugin system**: Allow third-party analyzers
4. **Progress UI**: Rich terminal output with spinners
5. **Cache layer**: Skip unchanged file analysis
