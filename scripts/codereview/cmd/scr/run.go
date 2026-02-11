package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/pipeline"
	"github.com/spf13/cobra"
)

var skipPhases string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute all analysis phases",
	Long: `Execute all 6 analysis phases sequentially.

Phases are executed in order:
  1. scope    - Detect changed files and project language (30s timeout)
  2. static   - Run linters and static analyzers (5m timeout)
  3. ast      - Extract semantic AST diffs (2m timeout)
  4. callgraph - Analyze call relationships (3m timeout)
  5. dataflow - Security-focused data flow analysis (3m timeout)
  6. context  - Compile reviewer-specific context files (30s timeout)

Output artifacts are written to the output directory (default: .scr/).`,
	Example: `  # Run all phases on uncommitted changes
  scr run

  # Run comparing branches
  scr run --base=main --head=HEAD

  # Run with specific file patterns
  scr run --files=cmd/*.go,internal/**/*.go

  # Skip specific phases
  scr run --skip=static,dataflow

  # Custom output directory
  scr run --output-dir=./review-output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPipeline()
	},
}

func init() {
	runCmd.Flags().StringVar(&skipPhases, "skip", "", "Comma-separated list of phases to skip (e.g., 'static,callgraph')")
}

func runPipeline() error {
	// Validate flag combinations
	if err := validateFlagCombinations(); err != nil {
		return err
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create pipeline configuration
	cfg := &phases.Config{
		WorkDir:   getWorkDir(),
		OutputDir: globalCfg.OutputDir,
		Verbose:   globalCfg.Verbose,
		BaseRef:   globalCfg.BaseRef,
		HeadRef:   globalCfg.HeadRef,
		Unstaged:  globalCfg.Unstaged,
		Files:     globalCfg.Files,
		FilesFrom: globalCfg.FilesFrom,
	}

	// Set default refs if not using file patterns or unstaged mode
	if !globalCfg.HasFilePatterns() && !globalCfg.Unstaged {
		if cfg.BaseRef == "" {
			cfg.BaseRef = "main"
		}
		if cfg.HeadRef == "" {
			cfg.HeadRef = "HEAD"
		}
	}

	// Validate skip phases before running
	skipList := parseSkipList(skipPhases)
	for _, s := range skipList {
		if !phases.IsValidPhaseName(s) {
			return fmt.Errorf("invalid phase name in --skip: %q (valid: %v)", s, phases.PhaseNames())
		}
	}

	// Create and configure pipeline
	p := pipeline.New().
		WithSkip(skipList).
		WithVerbose(globalCfg.Verbose)

	// Run pipeline
	result, err := p.Run(ctx, cfg)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("execution interrupted")
		}
		return err
	}

	if result == nil {
		return fmt.Errorf("pipeline returned nil result")
	}

	if !result.Success() {
		return fmt.Errorf("one or more phases failed")
	}

	return nil
}

func validateFlagCombinations() error {
	hasFilePatterns := globalCfg.HasFilePatterns()
	hasRefs := globalCfg.BaseRef != "" || globalCfg.HeadRef != ""

	if globalCfg.Unstaged && hasFilePatterns {
		return fmt.Errorf("--unstaged cannot be used with --files/--files-from")
	}
	if globalCfg.Unstaged && hasRefs {
		return fmt.Errorf("--unstaged cannot be used with --base/--head")
	}
	if hasFilePatterns && hasRefs {
		return fmt.Errorf("--files/--files-from cannot be used with --base/--head")
	}
	return nil
}
