package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	dataflowphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/dataflow"
	"github.com/spf13/cobra"
)

var (
	dataflowScopePath  string
	dataflowScriptsDir string
	dataflowLanguage   string
)

var phaseDataflowCmd = &cobra.Command{
	Use:   "dataflow",
	Short: "Phase 4: Security-focused data flow analysis",
	Long: `Perform security-focused data flow analysis.

This phase analyzes source code to detect:
  - Untrusted data sources (HTTP inputs, env vars, files)
  - Sensitive data sinks (database, exec, response)
  - Unsanitized data flows between sources and sinks
  - Nil/null safety issues

Supported languages: Go, TypeScript, Python`,
	Example: `  # Run with default scope path
  scr phase dataflow

  # Custom scope path
  scr phase dataflow --scope=.scr/scope.json

  # Analyze only Go files
  scr phase dataflow --lang=go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDataflowPhase()
	},
}

func init() {
	phaseDataflowCmd.Flags().StringVar(&dataflowScopePath, "scope", "", "Path to scope.json (default: <output-dir>/scope.json)")
	phaseDataflowCmd.Flags().StringVar(&dataflowScriptsDir, "scripts", "", "Directory containing language scripts")
	phaseDataflowCmd.Flags().StringVar(&dataflowLanguage, "lang", "", "Analyze specific language only")
}

func runDataflowPhase() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	scopePath := dataflowScopePath
	if scopePath == "" {
		scopePath = filepath.Join(globalCfg.OutputDir, "scope.json")
	}

	cfg := &phases.Config{
		WorkDir:    getWorkDir(),
		OutputDir:  globalCfg.OutputDir,
		Verbose:    globalCfg.Verbose,
		ScopePath:  scopePath,
		ScriptsDir: dataflowScriptsDir,
		Language:   dataflowLanguage,
	}

	phase := dataflowphase.New()
	phaseCtx, phaseCancel := context.WithTimeout(ctx, phase.Timeout())
	defer phaseCancel()

	return phase.Run(phaseCtx, cfg)
}
