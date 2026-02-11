// Package ast implements Phase 2: AST extraction for the unified CLI.
package ast

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/ast"
	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
)

// Phase implements the AST extraction phase.
type Phase struct{}

// New creates a new AST extraction phase.
func New() *Phase {
	return &Phase{}
}

// Name returns the canonical name of this phase.
func (p *Phase) Name() string {
	return "ast"
}

// Timeout returns the default timeout for this phase.
func (p *Phase) Timeout() time.Duration {
	return 2 * time.Minute
}

// ShouldSkip checks if this phase should be skipped.
func (p *Phase) ShouldSkip(cfg *phases.Config) (bool, string) {
	// Skip if no scope.json exists
	scopePath := cfg.ScopePath
	if scopePath == "" {
		scopePath = filepath.Join(cfg.OutputDir, "scope.json")
	}
	if _, err := os.Stat(scopePath); os.IsNotExist(err) {
		return true, "scope.json missing - skipping AST phase"
	}

	// Check if there are any files to process
	data, err := fileutil.ReadJSONFileWithLimit(scopePath)
	if err != nil {
		return true, fmt.Sprintf("failed to read scope.json: %v", err)
	}

	var scope struct {
		Files struct {
			Modified []string `json:"modified"`
			Added    []string `json:"added"`
			Deleted  []string `json:"deleted"`
		} `json:"files"`
	}
	if err := json.Unmarshal(data, &scope); err != nil {
		return true, fmt.Sprintf("failed to parse scope.json: %v", err)
	}

	totalFiles := len(scope.Files.Modified) + len(scope.Files.Added) + len(scope.Files.Deleted)
	if totalFiles == 0 {
		return true, "No changed files detected"
	}

	return false, ""
}

// Run executes the AST extraction phase.
func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	workDir := cfg.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Determine script directory for TypeScript/Python extractors
	scriptsPath := cfg.ScriptsDir
	if scriptsPath == "" {
		exe, err := os.Executable()
		if err == nil {
			scriptsPath = filepath.Join(filepath.Dir(exe), "..", "..")
		} else {
			scriptsPath = "."
		}
	}

	// Validate scripts directory before use
	if err := validateScriptsDir(scriptsPath); err != nil {
		return fmt.Errorf("scripts directory validation failed: %w", err)
	}

	if cfg.Verbose {
		logger.Debug("scripts directory", "path", scriptsPath)
	}

	validator, err := ast.NewPathValidator(workDir)
	if err != nil {
		return fmt.Errorf("failed to initialize path validator: %w", err)
	}

	// Create registry with all extractors
	registry := ast.NewRegistry()
	registry.Register(ast.NewGoExtractor())
	registry.Register(ast.NewTypeScriptExtractor(scriptsPath))
	registry.Register(ast.NewPythonExtractor(scriptsPath))

	// Handle batch mode (normal operation in pipeline)
	if cfg.BatchFile != "" {
		return p.processBatch(ctx, registry, cfg, validator)
	}

	return fmt.Errorf("batch file is required for AST phase in pipeline mode")
}

// processBatch processes a batch of file pairs.
func (p *Phase) processBatch(ctx context.Context, registry *ast.Registry, cfg *phases.Config, validator *ast.PathValidator) error {
	batchPath := cfg.BatchFile
	validatedBatch, err := validator.ValidatePath(batchPath)
	if err != nil {
		return fmt.Errorf("invalid batch file path: %w", err)
	}

	data, err := fileutil.ReadJSONFileWithLimit(validatedBatch)
	if err != nil {
		return fmt.Errorf("failed to read batch file: %w", err)
	}

	var pairs []ast.FilePair
	if err := json.Unmarshal(data, &pairs); err != nil {
		return fmt.Errorf("failed to parse batch file: %w", err)
	}

	if pairs == nil {
		pairs = []ast.FilePair{}
	}

	if cfg.Verbose {
		logger.Debug("processing batch", "pairs", len(pairs))
	}

	diffs, err := registry.ExtractAll(ctx, pairs)
	if err != nil {
		return fmt.Errorf("batch extraction failed: %w", err)
	}

	// Determine language from diffs
	language := "unknown"
	if len(diffs) > 0 {
		language = diffs[0].Language
	}

	// Write output to language-specific file
	outputPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s-ast.json", language))
	output, err := json.MarshalIndent(diffs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	if err := os.WriteFile(outputPath, output, 0600); err != nil {
		return fmt.Errorf("failed to write AST output: %w", err)
	}

	if cfg.Verbose {
		logger.Debug("AST output written", "path", outputPath, "diffs", len(diffs))
	}

	return nil
}

// validateScriptsDir validates the scripts directory path for security.
func validateScriptsDir(scriptsDir string) error {
	if scriptsDir == "" {
		return nil
	}
	if strings.Contains(scriptsDir, "..") {
		return fmt.Errorf("path traversal detected in scripts directory")
	}
	_, err := fileutil.ValidateDirectory(scriptsDir, "")
	return err
}
