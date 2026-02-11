// Package main implements the call-graph binary (deprecated wrapper).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	callgraphphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/callgraph"

	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/recovery"
)

var (
	astFile   = flag.String("ast", "", "Path to {lang}-ast.json from Phase 2")
	outputDir = flag.String("output", ".ring/codereview", "Output directory")
	timeout   = flag.Int("timeout", 30, "Time budget in seconds (unused, kept for compatibility)")
	language  = flag.String("lang", "", "Language override (go, typescript, python)")
	verbose   = flag.Bool("v", false, "Enable verbose output")

	// Deprecated flags kept for compatibility
	_ = flag.String("languages-file", "", "DEPRECATED: no longer used")
	_ = flag.String("output-suffix", "", "DEPRECATED: no longer used")
)

func init() {
	flag.BoolVar(verbose, "verbose", false, "Enable verbose output")
}

func main() {
	os.Exit(recovery.WrapMain(realMain))
}

func realMain() {
	fmt.Fprintln(os.Stderr, "DEPRECATED: call-graph is deprecated, use 'scr phase callgraph' instead")

	flag.Parse()

	if *verbose {
		logger.SetDefault(logger.NewLogger(logger.WithLevel(logger.LevelDebug)))
	}

	// Suppress unused variable warning
	_ = timeout

	cfg := &phases.Config{
		OutputDir: *outputDir,
		ASTPath:   *astFile,
		Language:  *language,
		Verbose:   *verbose,
	}

	ctx := context.Background()
	if err := callgraphphase.New().Run(ctx, cfg); err != nil {
		logger.Error("error", "error", err)
		os.Exit(1)
	}
}
