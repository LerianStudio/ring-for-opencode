// Package main provides the scope-detector CLI binary for code review scope analysis.
// It analyzes git diffs to detect changed files, determine project language,
// and output structured JSON for downstream code review tools.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/output"
	"github.com/lerianstudio/ring/scripts/codereview/internal/scope"
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

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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

	// Create detector
	detector := scope.NewDetector(wd)

	// Detect scope based on refs or explicit files
	var result *scope.ScopeResult
	var err error

	patterns, patternsErr := resolveFilePatterns(*filesFlag, *filesFrom)
	if patternsErr != nil {
		return patternsErr
	}

	if len(patterns) > 0 {
		if *baseRef != "" || *headRef != "" || *unstaged {
			return fmt.Errorf("--files/--files-from cannot be used with --base/--head or --unstaged")
		}
		expanded, expandErr := scope.ExpandFilePatterns(wd, patterns)
		if expandErr != nil {
			return expandErr
		}
		if len(expanded) == 0 {
			fmt.Fprintln(os.Stderr, "Warning: no files matched the provided patterns")
			result = &scope.ScopeResult{
				BaseRef:          "",
				HeadRef:          "",
				Language:         scope.LanguageUnknown.String(),
				Languages:        []string{},
				ModifiedFiles:    []string{},
				AddedFiles:       []string{},
				DeletedFiles:     []string{},
				TotalFiles:       0,
				TotalAdditions:   0,
				TotalDeletions:   0,
				PackagesAffected: []string{},
			}
			err = nil
		} else {
			result, err = detector.DetectFromFiles("", expanded)
		}
	} else if *unstaged {
		if *baseRef != "" || *headRef != "" {
			return fmt.Errorf("--unstaged cannot be used with --base/--head")
		}
		result, err = detector.DetectUnstagedChanges()
	} else if *baseRef == "" && *headRef == "" {
		// No refs specified: detect all uncommitted changes (staged + unstaged)
		result, err = detector.DetectAllChanges()
	} else {
		// Refs specified: compare specific refs
		result, err = detector.DetectFromRefs(*baseRef, *headRef)
	}

	if err != nil {
		return fmt.Errorf("failed to detect scope: %w", err)
	}

	// Verbose output
	if *verbose {
		fmt.Fprintln(os.Stderr, "=== Scope Detector (Verbose) ===")
		fmt.Fprintf(os.Stderr, "Working directory: %s\n", wd)
		if *unstaged {
			fmt.Fprintln(os.Stderr, "Mode: unstaged + untracked")
		} else if *baseRef == "" {
			fmt.Fprintln(os.Stderr, "Base ref: (empty - detecting all uncommitted changes)")
		} else {
			fmt.Fprintf(os.Stderr, "Base ref: %s\n", *baseRef)
		}
		if *unstaged {
			fmt.Fprintln(os.Stderr, "Head ref: working tree")
		} else if *headRef == "" {
			fmt.Fprintln(os.Stderr, "Head ref: (empty - using working tree)")
		} else {
			fmt.Fprintf(os.Stderr, "Head ref: %s\n", *headRef)
		}
		fmt.Fprintf(os.Stderr, "Files found: %d\n", result.TotalFiles)
		fmt.Fprintf(os.Stderr, "Language detected: %s\n", result.Language)
		fmt.Fprintln(os.Stderr, "================================")
	}

	// Check for no changes
	if result.TotalFiles == 0 {
		fmt.Fprintln(os.Stderr, "No changes detected")
		// Still output empty result for consistency
	}

	// Create output wrapper
	// Defensive: NewScopeOutput returns nil only if result is nil, which cannot happen here
	// after successful detection. Kept for safety against future refactoring.
	scopeOutput := output.NewScopeOutput(result)
	if scopeOutput == nil {
		return fmt.Errorf("failed to create scope output: nil result")
	}

	// Write output
	if *outputPath != "" {
		validatedOutput, err := fileutil.ValidatePath(*outputPath, ".")
		if err != nil {
			return fmt.Errorf("invalid output path: %w", err)
		}
		// Write to file
		if err := scopeOutput.WriteToFile(validatedOutput); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Scope written to %s\n", validatedOutput)
	} else {
		// Write to stdout
		if err := scopeOutput.WriteToStdout(); err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}

	return nil
}
