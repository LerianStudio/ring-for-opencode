// Package context implements Phase 5: context compilation for the unified CLI.
package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	ctxpkg "github.com/lerianstudio/ring/scripts/codereview/internal/context"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
)

// Phase implements the context compilation phase.
type Phase struct{}

// New creates a new context compilation phase.
func New() *Phase {
	return &Phase{}
}

// Name returns the canonical name of this phase.
func (p *Phase) Name() string {
	return "context"
}

// Timeout returns the default timeout for this phase.
func (p *Phase) Timeout() time.Duration {
	return 30 * time.Second
}

// ShouldSkip checks if this phase should be skipped.
func (p *Phase) ShouldSkip(cfg *phases.Config) (bool, string) {
	scopePath := cfg.ScopePath
	if scopePath == "" {
		scopePath = filepath.Join(cfg.OutputDir, "scope.json")
	}
	if _, err := os.Stat(scopePath); os.IsNotExist(err) {
		return true, "scope.json missing - skipping context phase"
	}
	return false, ""
}

// Run executes the context compilation phase.
func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Derive input directory from ScopePath (which is a file path) or OutputDir
	inputDir := cfg.OutputDir
	if cfg.ScopePath != "" {
		inputDir = filepath.Dir(cfg.ScopePath)
	}
	if inputDir == "" {
		inputDir = ".scr"
	}

	// Validate input directory exists
	if _, err := os.Stat(inputDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input directory does not exist: %s", inputDir)
		}
		return fmt.Errorf("cannot access input directory %s: %w", inputDir, err)
	}

	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = inputDir
	}

	if cfg.Verbose {
		logger.Debug("directories", "input", inputDir, "output", outputDir)
	}

	// Create compiler and execute
	compiler, err := ctxpkg.NewCompilerWithValidation(inputDir, outputDir)
	if err != nil {
		return fmt.Errorf("compiler initialization failed: %w", err)
	}

	if err := compiler.Compile(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	if cfg.Verbose {
		reviewers := ctxpkg.GetReviewerNames()
		logger.Debug("generated context files", "reviewers", reviewers)
	}

	fmt.Println("Context compilation complete.")
	return nil
}
