// Package main provides a deprecated wrapper for the compile-context binary.
// Use 'scr phase context' instead.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	contextphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/context"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/recovery"
)

var version = "dev"

var (
	inputDir    = flag.String("input", ".ring/codereview", "Input directory containing phase outputs")
	outputDir   = flag.String("output", "", "Output directory for context files (default: same as input)")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
	showVersion = flag.Bool("version", false, "Show version and exit")
)

func main() {
	os.Exit(recovery.WrapMain(realMain))
}

func realMain() {
	flag.Usage = printUsage
	flag.Parse()

	if *showVersion {
		fmt.Printf("compile-context version %s\n", version)
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "DEPRECATED: compile-context is deprecated, use 'scr phase context' instead")

	if *verbose {
		logger.SetDefault(logger.NewLogger(logger.WithLevel(logger.LevelDebug)))
	}

	// Map legacy flags to phases.Config
	// ScopePath is used as input directory, OutputDir is used for output
	outDir := *outputDir
	if outDir == "" {
		outDir = *inputDir
	}

	cfg := &phases.Config{
		ScopePath: *inputDir,
		OutputDir: outDir,
		Verbose:   *verbose,
	}

	if err := contextphase.New().Run(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Context Compiler - Phase 5 of the codereview system\n\n")
	fmt.Fprintf(os.Stderr, "DEPRECATED: Use 'scr phase context' instead.\n\n")
	fmt.Fprintf(os.Stderr, "Aggregates outputs from previous phases and generates reviewer-specific\n")
	fmt.Fprintf(os.Stderr, "context files in Markdown format.\n\n")
	fmt.Fprintf(os.Stderr, "Expected Phase Outputs (in input directory):\n")
	fmt.Fprintf(os.Stderr, "  Phase 0: scope.json         - Changed files and language detection\n")
	fmt.Fprintf(os.Stderr, "  Phase 1: static-analysis.json - Lint findings and code quality issues\n")
	fmt.Fprintf(os.Stderr, "  Phase 2: {lang}-ast.json    - AST extraction and semantic changes\n")
	fmt.Fprintf(os.Stderr, "  Phase 3: {lang}-calls.json  - Call graph and impact analysis\n")
	fmt.Fprintf(os.Stderr, "  Phase 4: {lang}-flow.json   - Data flow and security analysis\n\n")
	fmt.Fprintf(os.Stderr, "Generated Context Files:\n")
	fmt.Fprintf(os.Stderr, "  context-code-reviewer.md         - Code quality and style analysis\n")
	fmt.Fprintf(os.Stderr, "  context-security-reviewer.md     - Security vulnerabilities and data flows\n")
	fmt.Fprintf(os.Stderr, "  context-business-logic-reviewer.md - Business logic and impact analysis\n")
	fmt.Fprintf(os.Stderr, "  context-test-reviewer.md         - Test coverage and gaps\n")
	fmt.Fprintf(os.Stderr, "  context-nil-safety-reviewer.md   - Nil/null safety analysis\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Compile context using default directories:\n")
	fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # Compile from custom input directory:\n")
	fmt.Fprintf(os.Stderr, "  %s --input /path/to/codereview --verbose\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  # Output to different directory:\n")
	fmt.Fprintf(os.Stderr, "  %s --input .ring/codereview --output ./review-context\n", os.Args[0])
}
