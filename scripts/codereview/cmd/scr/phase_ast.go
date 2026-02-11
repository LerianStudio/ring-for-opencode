package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	astphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/ast"
	"github.com/spf13/cobra"
)

var (
	astBatchFile  string
	astScriptsDir string
	astLanguage   string
)

var phaseASTCmd = &cobra.Command{
	Use:   "ast",
	Short: "Phase 2: Extract semantic AST diffs",
	Long: `Extract semantic AST diffs from changed files.

This phase parses source files and extracts structural changes including
functions, types, imports, and other language constructs.

Supported languages: Go, TypeScript, Python`,
	Example: `  # Run with batch file from pipeline
  scr phase ast --batch=.scr/ast-batch.json

  # Force specific language
  scr phase ast --batch=batch.json --lang=go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runASTPhase()
	},
}

func init() {
	phaseASTCmd.Flags().StringVar(&astBatchFile, "batch", "", "Path to batch JSON file with file pairs")
	phaseASTCmd.Flags().StringVar(&astScriptsDir, "scripts", "", "Directory containing language scripts (ts/, py/)")
	phaseASTCmd.Flags().StringVar(&astLanguage, "lang", "", "Force language (go, typescript, python)")
}

func runASTPhase() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	batchFile := astBatchFile
	if batchFile == "" {
		batchFile = filepath.Join(globalCfg.OutputDir, "ast-batch.json")
	}

	cfg := &phases.Config{
		WorkDir:    getWorkDir(),
		OutputDir:  globalCfg.OutputDir,
		Verbose:    globalCfg.Verbose,
		BatchFile:  batchFile,
		ScriptsDir: astScriptsDir,
		Language:   astLanguage,
	}

	phase := astphase.New()
	phaseCtx, phaseCancel := context.WithTimeout(ctx, phase.Timeout())
	defer phaseCancel()

	return phase.Run(phaseCtx, cfg)
}
