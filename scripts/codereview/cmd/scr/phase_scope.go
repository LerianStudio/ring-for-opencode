package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	scopephase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/scope"
	"github.com/spf13/cobra"
)

var phaseScopeCmd = &cobra.Command{
	Use:   "scope",
	Short: "Phase 0: Detect changed files and project language",
	Long: `Detect changed files and determine project language.

This phase analyzes git diffs to detect changed files, determines the primary
project language, and outputs structured JSON for downstream phases.`,
	Example: `  # Detect all uncommitted changes
  scr phase scope

  # Compare branches
  scr phase scope --base=main --head=HEAD

  # Analyze specific files
  scr phase scope --files=cmd/*.go,internal/**/*.go

  # Analyze unstaged changes only
  scr phase scope --unstaged`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScopePhase()
	},
}

func runScopePhase() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	phase := scopephase.New()
	phaseCtx, phaseCancel := context.WithTimeout(ctx, phase.Timeout())
	defer phaseCancel()

	return phase.Run(phaseCtx, cfg)
}
