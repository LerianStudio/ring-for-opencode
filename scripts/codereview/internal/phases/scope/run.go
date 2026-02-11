// Package scope implements Phase 0: scope detection for the unified CLI.
package scope

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/output"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/scope"
)

// Phase implements the scope detection phase.
type Phase struct{}

// New creates a new scope detection phase.
func New() *Phase {
	return &Phase{}
}

// Name returns the canonical name of this phase.
func (p *Phase) Name() string {
	return "scope"
}

// Timeout returns the default timeout for this phase.
func (p *Phase) Timeout() time.Duration {
	return 30 * time.Second
}

// Run executes the scope detection phase.
func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Determine working directory
	wd := cfg.WorkDir
	if wd == "" {
		var err error
		wd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Create detector
	detector := scope.NewDetector(wd)

	// Detect scope based on configuration
	var result *scope.ScopeResult
	var err error

	patterns, patternsErr := resolveFilePatterns(cfg.Files, cfg.FilesFrom)
	if patternsErr != nil {
		return patternsErr
	}

	if len(patterns) > 0 {
		if cfg.BaseRef != "" || cfg.HeadRef != "" || cfg.Unstaged {
			return fmt.Errorf("--files/--files-from cannot be used with --base/--head or --unstaged")
		}
		expanded, expandErr := scope.ExpandFilePatterns(wd, patterns)
		if expandErr != nil {
			return expandErr
		}
		if len(expanded) == 0 {
			logger.Warn("no files matched the provided patterns")
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
		} else {
			result, err = detector.DetectFromFiles("", expanded)
		}
	} else if cfg.Unstaged {
		if cfg.BaseRef != "" || cfg.HeadRef != "" {
			return fmt.Errorf("--unstaged cannot be used with --base/--head")
		}
		result, err = detector.DetectUnstagedChanges()
	} else if cfg.BaseRef == "" && cfg.HeadRef == "" {
		// No refs specified: detect all uncommitted changes (staged + unstaged)
		result, err = detector.DetectAllChanges()
	} else {
		// Refs specified: compare specific refs
		result, err = detector.DetectFromRefs(cfg.BaseRef, cfg.HeadRef)
	}

	if err != nil {
		return fmt.Errorf("failed to detect scope: %w", err)
	}

	if cfg.Verbose {
		mode := "refs"
		if cfg.Unstaged {
			mode = "unstaged + untracked"
		} else if cfg.BaseRef == "" && cfg.HeadRef == "" {
			mode = "all uncommitted changes"
		}
		logger.Debug("scope detection", "workdir", wd, "mode", mode, "base", cfg.BaseRef, "head", cfg.HeadRef, "files_found", result.TotalFiles, "language", result.Language)
	}

	// Check for no changes
	if result.TotalFiles == 0 {
		fmt.Fprintln(os.Stderr, "No changes detected")
	}

	// Create output wrapper
	scopeOutput := output.NewScopeOutput(result)
	if scopeOutput == nil {
		return fmt.Errorf("failed to create scope output: nil result")
	}

	// Write output to file
	outputPath := cfg.ScopePath
	if outputPath == "" {
		// Use OutputDir directly as the parent directory for scope.json
		outputPath = filepath.Join(cfg.OutputDir, "scope.json")
	}

	validatedOutput, err := fileutil.ValidatePath(outputPath, ".")
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	if err := scopeOutput.WriteToFile(validatedOutput); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "Scope written to %s\n", validatedOutput)
	}

	return nil
}

// resolveFilePatterns resolves file patterns from flags.
func resolveFilePatterns(files, filesFrom string) ([]string, error) {
	var patterns []string

	if files != "" {
		for _, p := range strings.Split(files, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				patterns = append(patterns, p)
			}
		}
	}

	if filesFrom != "" {
		filePatterns, err := readPatternsFromFile(filesFrom)
		if err != nil {
			return nil, fmt.Errorf("failed to read patterns from file: %w", err)
		}
		patterns = append(patterns, filePatterns...)
	}

	return patterns, nil
}

// readPatternsFromFile reads file patterns from a file, one per line.
func readPatternsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return patterns, nil
}
