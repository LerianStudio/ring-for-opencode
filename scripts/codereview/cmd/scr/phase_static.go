package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	staticphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/static"
	"github.com/spf13/cobra"
)

var staticScopePath string

var phaseStaticCmd = &cobra.Command{
	Use:   "static",
	Short: "Phase 1: Run linters and static analyzers",
	Long: `Run linters and static analyzers on changed files.

This phase reads scope.json from Phase 0 and runs appropriate linters
based on the detected language. Findings are filtered to only include
issues in changed files.

Supported linters:
  Go:         golangci-lint, staticcheck, gosec
  TypeScript: tsc, eslint
  Python:     ruff, mypy, pylint, bandit`,
	Example: `  # Run with default scope path
  scr phase static

  # Custom scope path
  scr phase static --scope=.scr/scope.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStaticPhase()
	},
}

func init() {
	phaseStaticCmd.Flags().StringVar(&staticScopePath, "scope", "", "Path to scope.json (default: <output-dir>/scope.json)")
}

func runStaticPhase() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	scopePath := staticScopePath
	if scopePath == "" {
		scopePath = filepath.Join(globalCfg.OutputDir, "scope.json")
	}

	cfg := &phases.Config{
		WorkDir:   getWorkDir(),
		OutputDir: globalCfg.OutputDir,
		Verbose:   globalCfg.Verbose,
		ScopePath: scopePath,
	}

	phase := staticphase.New()
	phaseCtx, phaseCancel := context.WithTimeout(ctx, phase.Timeout())
	defer phaseCancel()

	return phase.Run(phaseCtx, cfg)
}
