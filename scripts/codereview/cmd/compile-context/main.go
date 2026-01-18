// Package main implements the compile-context binary for Phase 5 of the codereview system.
// It aggregates outputs from previous phases and generates reviewer-specific context files.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lerianstudio/ring/scripts/codereview/internal/context"
)

var version = "dev"

var (
	inputDir    = flag.String("input", ".ring/codereview", "Input directory containing phase outputs")
	outputDir   = flag.String("output", "", "Output directory for context files (default: same as input)")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
	showVersion = flag.Bool("version", false, "Show version and exit")
)

func main() {
	flag.Usage = printUsage
	flag.Parse()

	if *showVersion {
		fmt.Printf("compile-context version %s\n", version)
		os.Exit(0)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Validate input directory exists
	if _, err := os.Stat(*inputDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", *inputDir)
		}
		return fmt.Errorf("cannot access input directory %s: %w", *inputDir, err)
	}

	// Default output directory to input directory
	outDir := *outputDir
	if outDir == "" {
		outDir = *inputDir
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Input directory: %s\n", *inputDir)
		fmt.Fprintf(os.Stderr, "Output directory: %s\n", outDir)
	}

	// Create compiler and execute
	compiler, err := context.NewCompilerWithValidation(*inputDir, outDir)
	if err != nil {
		return fmt.Errorf("compiler initialization failed: %w", err)
	}

	if err := compiler.Compile(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	// Print success information
	if *verbose {
		reviewers := context.GetReviewerNames()
		fmt.Fprintf(os.Stderr, "\nGenerated context files:\n")
		for _, reviewer := range reviewers {
			fmt.Fprintf(os.Stderr, "  - context-%s.md\n", reviewer)
		}
		fmt.Fprintln(os.Stderr)
	}

	fmt.Println("Context compilation complete.")
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Context Compiler - Phase 5 of the codereview system\n\n")
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
