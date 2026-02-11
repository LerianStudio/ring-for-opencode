// Package main implements the static-analysis binary (deprecated wrapper).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	staticphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/static"
	"github.com/lerianstudio/ring/scripts/codereview/internal/recovery"
)

var (
	scopePath  = flag.String("scope", "", "Path to scope.json (default: .ring/codereview/scope.json)")
	outputPath = flag.String("output", "", "Output directory (default: .ring/codereview/)")
	verbose    = flag.Bool("v", false, "Verbose output")
	timeout    = flag.Duration("timeout", 5*time.Minute, "Timeout for analysis")
)

func main() {
	os.Exit(recovery.WrapMain(realMain))
}

func realMain() {
	fmt.Fprintln(os.Stderr, "DEPRECATED: static-analysis is deprecated, use 'scr phase static' instead")

	flag.Parse()

	if *verbose {
		logger.SetDefault(logger.NewLogger(logger.WithLevel(logger.LevelDebug)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	cfg := &phases.Config{
		ScopePath: *scopePath,
		OutputDir: *outputPath,
		Verbose:   *verbose,
	}

	if err := staticphase.New().Run(ctx, cfg); err != nil {
		logger.Error("error", "error", err)
		os.Exit(1)
	}
}
