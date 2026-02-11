package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	callgraphphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/callgraph"
	"github.com/spf13/cobra"
)

var (
	callgraphASTPath  string
	callgraphLanguage string
)

var phaseCallgraphCmd = &cobra.Command{
	Use:   "callgraph",
	Short: "Phase 3: Analyze call relationships",
	Long: `Analyze call graph relationships for modified functions.

This phase reads AST output from Phase 2 and builds call graphs to identify
direct and transitive callers, affected tests, and impact analysis.

Supported languages: Go, TypeScript, Python`,
	Example: `  # Run with auto-detected AST file
  scr phase callgraph

  # Custom AST file path
  scr phase callgraph --ast=.scr/go-ast.json

  # Force specific language
  scr phase callgraph --lang=go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCallgraphPhase()
	},
}

func init() {
	phaseCallgraphCmd.Flags().StringVar(&callgraphASTPath, "ast", "", "Path to AST JSON file from Phase 2")
	phaseCallgraphCmd.Flags().StringVar(&callgraphLanguage, "lang", "", "Force language (go, typescript, python)")
}

func runCallgraphPhase() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := &phases.Config{
		WorkDir:   getWorkDir(),
		OutputDir: globalCfg.OutputDir,
		Verbose:   globalCfg.Verbose,
		ASTPath:   callgraphASTPath,
		ScopePath: filepath.Join(globalCfg.OutputDir, "scope.json"),
		Language:  callgraphLanguage,
	}

	phase := callgraphphase.New()
	phaseCtx, phaseCancel := context.WithTimeout(ctx, phase.Timeout())
	defer phaseCancel()

	return phase.Run(phaseCtx, cfg)
}
