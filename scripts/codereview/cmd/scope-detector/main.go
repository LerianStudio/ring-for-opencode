// Package main provides the scope-detector CLI binary (deprecated wrapper).
// Use 'scr phase scope' instead.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	scopephase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/scope"
	"github.com/lerianstudio/ring/scripts/codereview/internal/recovery"
)

// version is set via ldflags during build.
var version = "dev"

var (
	baseRef     = flag.String("base", "", "Base reference (commit/branch). When both refs empty, detects all uncommitted changes")
	headRef     = flag.String("head", "", "Head reference (commit/branch). When both refs empty, detects all uncommitted changes")
	filesFlag   = flag.String("files", "", "Comma-separated file patterns to analyze (mutually exclusive with --base/--head)")
	filesFrom   = flag.String("files-from", "", "Path to file containing file patterns (one per line)")
	unstaged    = flag.Bool("unstaged", false, "Analyze only unstaged and untracked files")
	outputPath  = flag.String("output", "", "Output file path. Empty = write to stdout")
	workDir     = flag.String("workdir", "", "Working directory. Empty = current directory")
	showVersion = flag.Bool("version", false, "Show version and exit")
	verbose     = flag.Bool("v", false, "Enable verbose output")
)

func init() {
	flag.BoolVar(verbose, "verbose", false, "Enable verbose output")
}

func main() {
	os.Exit(recovery.WrapMain(realMain))
}

func realMain() {
	fmt.Fprintln(os.Stderr, "DEPRECATED: scope-detector is deprecated, use 'scr phase scope' instead")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: scope-detector [options]\n\n")
		fmt.Fprintf(os.Stderr, "Analyzes git diff to detect changed files and project language.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  scope-detector                             # All uncommitted changes\n")
		fmt.Fprintf(os.Stderr, "  scope-detector --base=main --head=HEAD     # Compare branches\n")
		fmt.Fprintf(os.Stderr, "  scope-detector --files=cmd/*.go,scripts/**/*.ts\n")
		fmt.Fprintf(os.Stderr, "  scope-detector --files-from=.ring/filelist.txt\n")
		fmt.Fprintf(os.Stderr, "  scope-detector --unstaged\n")
		fmt.Fprintf(os.Stderr, "  scope-detector --output=.ring/codereview/scope.json\n")
	}
	flag.Parse()

	if *verbose {
		logger.SetDefault(logger.NewLogger(logger.WithLevel(logger.LevelDebug)))
	}

	if err := run(); err != nil {
		logger.Error("error", "error", err)
		os.Exit(1)
	}
}

// run executes the main CLI logic.
func run() error {
	// Handle --version flag
	if *showVersion {
		fmt.Printf("scope-detector version %s\n", version)
		return nil
	}

	// Determine working directory
	wd := *workDir
	if wd == "" {
		var err error
		wd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Handle stdout output specially since the library always writes to file
	writeToStdout := *outputPath == ""

	// Build phases.Config from flags
	cfg := &phases.Config{
		WorkDir:   wd,
		Verbose:   *verbose,
		BaseRef:   *baseRef,
		HeadRef:   *headRef,
		Unstaged:  *unstaged,
		Files:     *filesFlag,
		FilesFrom: *filesFrom,
	}

	if writeToStdout {
		// Use a temp file, then read and write to stdout
		tmpDir, err := os.MkdirTemp("", "scope-detector-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg.ScopePath = filepath.Join(tmpDir, "scope.json")

		// Run the phase
		ctx := context.Background()
		if err := scopephase.New().Run(ctx, cfg); err != nil {
			return err
		}

		// Read the output and write to stdout
		data, err := os.ReadFile(cfg.ScopePath)
		if err != nil {
			return fmt.Errorf("failed to read scope output: %w", err)
		}
		_, err = os.Stdout.Write(data)
		return err
	}

	// Write to file directly
	cfg.ScopePath = *outputPath

	ctx := context.Background()
	if err := scopephase.New().Run(ctx, cfg); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Scope written to %s\n", *outputPath)
	return nil
}
