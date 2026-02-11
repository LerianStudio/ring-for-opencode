package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	contextphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/context"
	"github.com/spf13/cobra"
)

var phaseContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Phase 5: Compile reviewer-specific context files",
	Long: `Compile reviewer-specific context files from all phase outputs.

This phase aggregates outputs from previous phases and generates markdown
files tailored for different reviewer perspectives:
  - context-code-reviewer.md         - Code quality and style analysis
  - context-security-reviewer.md     - Security vulnerabilities and data flows
  - context-business-logic-reviewer.md - Business logic and impact analysis
  - context-test-reviewer.md         - Test coverage and gaps
  - context-nil-safety-reviewer.md   - Nil/null safety analysis`,
	Example: `  # Run with default paths
  scr phase context

  # Custom output directory
  scr phase context --output-dir=./review-output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runContextPhase()
	},
}

func runContextPhase() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := &phases.Config{
		WorkDir:   getWorkDir(),
		OutputDir: globalCfg.OutputDir,
		Verbose:   globalCfg.Verbose,
	}

	phase := contextphase.New()
	phaseCtx, phaseCancel := context.WithTimeout(ctx, phase.Timeout())
	defer phaseCancel()

	return phase.Run(phaseCtx, cfg)
}
