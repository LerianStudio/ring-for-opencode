// Package pipeline provides in-process orchestration for all analysis phases.
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/git"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	astphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/ast"
	callgraphphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/callgraph"
	contextphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/context"
	dataflowphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/dataflow"
	scopephase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/scope"
	staticphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/static"
)

// Pipeline orchestrates the execution of all phases.
type Pipeline struct {
	phases  []phases.Phase
	skipSet map[string]bool
	verbose bool
	stderr  io.Writer
}

// New creates a new Pipeline with all phases.
func New() *Pipeline {
	return &Pipeline{
		phases: []phases.Phase{
			scopephase.New(),
			staticphase.New(),
			astphase.New(),
			callgraphphase.New(),
			dataflowphase.New(),
			contextphase.New(),
		},
		skipSet: make(map[string]bool),
		stderr:  os.Stderr,
	}
}

// NewWithPhases creates a new Pipeline with custom phases (for testing).
func NewWithPhases(ph []phases.Phase) *Pipeline {
	return &Pipeline{
		phases:  ph,
		skipSet: make(map[string]bool),
		stderr:  os.Stderr,
	}
}

// WithSkip sets phases to skip.
func (p *Pipeline) WithSkip(skip []string) *Pipeline {
	for _, s := range skip {
		p.skipSet[strings.TrimSpace(s)] = true
	}
	return p
}

// WithVerbose enables verbose output.
func (p *Pipeline) WithVerbose(verbose bool) *Pipeline {
	p.verbose = verbose
	return p
}

// WithStderr sets a custom stderr writer (for testing).
func (p *Pipeline) WithStderr(w io.Writer) *Pipeline {
	p.stderr = w
	return p
}

// Result captures the outcome of a pipeline execution.
type Result struct {
	PhaseResults  []phases.Result
	TotalDuration time.Duration
	Passed        int
	Failed        int
	Skipped       int
	FailedPhases  []string
}

// Success returns true if all phases passed (or were skipped).
func (r *Result) Success() bool {
	return r.Failed == 0
}

// Run executes all phases in sequence.
func (p *Pipeline) Run(ctx context.Context, cfg *phases.Config) (*Result, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config")
	}
	if cfg.OutputDir == "" {
		return nil, fmt.Errorf("output directory is required")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	result := &Result{
		PhaseResults: make([]phases.Result, 0, len(p.phases)),
	}
	startTime := time.Now()

	// Track AST temp state for cleanup
	var astTempDir string
	defer func() {
		if astTempDir != "" {
			if err := os.RemoveAll(astTempDir); err != nil {
				fmt.Fprintf(p.stderr, "Warning: failed to clean up temp directory %s: %v\n", astTempDir, err)
			}
		}
	}()

	for _, phase := range p.phases {
		// Check for cancellation
		select {
		case <-ctx.Done():
			p.printInterruptSummary(result, time.Since(startTime))
			return result, ctx.Err()
		default:
		}

		phaseResult := p.executePhase(ctx, phase, cfg, &astTempDir)
		result.PhaseResults = append(result.PhaseResults, phaseResult)

		// Update counters
		if phaseResult.Skipped {
			result.Skipped++
		} else if phaseResult.Success {
			result.Passed++
		} else {
			result.Failed++
			result.FailedPhases = append(result.FailedPhases, phaseResult.Name)
		}

		// Print progress
		p.printPhaseResult(phaseResult)

		// Cleanup AST temp files after AST phase
		if phase.Name() == "ast" && astTempDir != "" {
			if err := os.RemoveAll(astTempDir); err != nil {
				fmt.Fprintf(p.stderr, "Warning: failed to clean up temp directory %s: %v\n", astTempDir, err)
			}
			astTempDir = ""
		}
	}

	result.TotalDuration = time.Since(startTime)
	p.printSummary(result)

	return result, nil
}

func (p *Pipeline) executePhase(ctx context.Context, phase phases.Phase, cfg *phases.Config, astTempDir *string) phases.Result {
	result := phases.Result{
		Name: phase.Name(),
	}

	// Check if phase should be skipped via CLI flag
	if p.skipSet[phase.Name()] {
		result.Skipped = true
		result.Success = true
		result.SkipReason = "skipped via --skip flag"
		return result
	}

	// Check if phase has a Skip condition
	if skipper, ok := phase.(phases.SkipChecker); ok {
		if skip, reason := skipper.ShouldSkip(cfg); skip {
			result.Skipped = true
			result.Success = true
			result.SkipReason = reason
			return result
		}
	}

	startTime := time.Now()

	// Build phase-specific config
	phaseCfg := &phases.Config{
		WorkDir:    cfg.WorkDir,
		OutputDir:  cfg.OutputDir,
		Verbose:    cfg.Verbose,
		BaseRef:    cfg.BaseRef,
		HeadRef:    cfg.HeadRef,
		Unstaged:   cfg.Unstaged,
		Files:      cfg.Files,
		FilesFrom:  cfg.FilesFrom,
		ScopePath:  filepath.Join(cfg.OutputDir, "scope.json"),
		ScriptsDir: cfg.ScriptsDir,
		Language:   cfg.Language,
	}

	// Special handling for AST phase - generate batch file
	if phase.Name() == "ast" {
		batchPath, tempDir, err := generateASTBatchFile(cfg)
		if err != nil {
			result.Duration = time.Since(startTime)
			result.Error = fmt.Errorf("failed to generate AST batch file: %w", err)
			return result
		}
		phaseCfg.BatchFile = batchPath
		*astTempDir = tempDir
	}

	// Create context with timeout
	phaseCtx, cancel := context.WithTimeout(ctx, phase.Timeout())
	defer cancel()

	// Run the phase
	err := phase.Run(phaseCtx, phaseCfg)
	result.Duration = time.Since(startTime)

	if phaseCtx.Err() != nil {
		if ctx.Err() != nil {
			result.Error = fmt.Errorf("interrupted")
		} else {
			result.Error = fmt.Errorf("timeout after %v", phase.Timeout())
		}
		return result
	}

	result.Error = err
	result.Success = err == nil
	return result
}

func (p *Pipeline) printPhaseResult(result phases.Result) {
	if result.Skipped {
		if result.SkipReason != "" {
			fmt.Fprintf(p.stderr, "[SKIP] %s: %s\n", result.Name, result.SkipReason)
		} else {
			fmt.Fprintf(p.stderr, "[SKIP] %s\n", result.Name)
		}
	} else if result.Success {
		fmt.Fprintf(p.stderr, "[PASS] %s (%v)\n", result.Name, result.Duration.Round(time.Millisecond))
	} else {
		fmt.Fprintf(p.stderr, "[FAIL] %s (%v): %v\n", result.Name, result.Duration.Round(time.Millisecond), result.Error)
	}
}

func (p *Pipeline) printSummary(result *Result) {
	fmt.Fprintf(p.stderr, "\n")
	fmt.Fprintf(p.stderr, "========================================\n")
	fmt.Fprintf(p.stderr, "  Static Code Reviewer Summary\n")
	fmt.Fprintf(p.stderr, "========================================\n")
	fmt.Fprintf(p.stderr, "\n")

	fmt.Fprintf(p.stderr, "Phases: %d passed, %d failed, %d skipped\n", result.Passed, result.Failed, result.Skipped)
	fmt.Fprintf(p.stderr, "Total time: %v\n", result.TotalDuration.Round(time.Millisecond))
	fmt.Fprintf(p.stderr, "\n")

	fmt.Fprintf(p.stderr, "Phase Breakdown:\n")
	for _, pr := range result.PhaseResults {
		if pr.Skipped {
			if pr.SkipReason != "" {
				fmt.Fprintf(p.stderr, "  [SKIP] %-20s %s\n", pr.Name, pr.SkipReason)
			} else {
				fmt.Fprintf(p.stderr, "  [SKIP] %-20s\n", pr.Name)
			}
		} else if pr.Success {
			fmt.Fprintf(p.stderr, "  [PASS] %-20s %v\n", pr.Name, pr.Duration.Round(time.Millisecond))
		} else {
			fmt.Fprintf(p.stderr, "  [FAIL] %-20s %v\n", pr.Name, pr.Duration.Round(time.Millisecond))
		}
	}
	fmt.Fprintf(p.stderr, "\n")

	if len(result.FailedPhases) > 0 {
		fmt.Fprintf(p.stderr, "Failed phases:\n")
		for _, pr := range result.PhaseResults {
			if !pr.Success && !pr.Skipped {
				fmt.Fprintf(p.stderr, "  - %s: %v\n", pr.Name, pr.Error)
			}
		}
		fmt.Fprintf(p.stderr, "\n")
	}

	if result.Failed == 0 {
		fmt.Fprintf(p.stderr, "All phases completed successfully.\n")
	} else {
		fmt.Fprintf(p.stderr, "Some phases failed. Review output for details.\n")
	}
}

func (p *Pipeline) printInterruptSummary(result *Result, duration time.Duration) {
	fmt.Fprintf(p.stderr, "\nInterrupted. Partial results:\n")
	fmt.Fprintf(p.stderr, "Phases: %d passed, %d failed, %d skipped\n", result.Passed, result.Failed, result.Skipped)
	fmt.Fprintf(p.stderr, "Time before interrupt: %v\n", duration.Round(time.Millisecond))
}

// scopeJSON represents the structure of scope.json from Phase 0.
type scopeJSON struct {
	BaseRef   string        `json:"base_ref"`
	HeadRef   string        `json:"head_ref"`
	Language  string        `json:"language"`
	Languages []string      `json:"languages"`
	Files     filesByStatus `json:"files"`
}

type filesByStatus struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
}

type filePair struct {
	BeforePath string `json:"before_path"`
	AfterPath  string `json:"after_path"`
}

func generateASTBatchFile(cfg *phases.Config) (string, string, error) {
	scopePath := filepath.Join(cfg.OutputDir, "scope.json")
	data, err := fileutil.ReadJSONFileWithLimit(scopePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read scope.json: %w", err)
	}

	var scope scopeJSON
	if err := json.Unmarshal(data, &scope); err != nil {
		return "", "", fmt.Errorf("failed to parse scope.json: %w", err)
	}

	if scope.Files.Modified == nil {
		scope.Files.Modified = []string{}
	}
	if scope.Files.Added == nil {
		scope.Files.Added = []string{}
	}
	if scope.Files.Deleted == nil {
		scope.Files.Deleted = []string{}
	}

	tempDir, err := os.MkdirTemp("", "ast-before-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	baseRef := cfg.BaseRef
	if baseRef == "" {
		baseRef = scope.BaseRef
	}

	var pairs []filePair

	// Handle modified files
	for _, file := range scope.Files.Modified {
		beforePath, err := extractFileFromGit(baseRef, file, tempDir)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", "", fmt.Errorf("failed to extract %s from %s: %w", file, baseRef, err)
		}
		pairs = append(pairs, filePair{
			BeforePath: beforePath,
			AfterPath:  file,
		})
	}

	// Handle added files
	for _, file := range scope.Files.Added {
		pairs = append(pairs, filePair{
			BeforePath: "",
			AfterPath:  file,
		})
	}

	// Handle deleted files
	for _, file := range scope.Files.Deleted {
		beforePath, err := extractFileFromGit(baseRef, file, tempDir)
		if err != nil {
			if cfg.Verbose {
				logger.Warn("could not extract deleted file", "file", file, "ref", baseRef, "error", err)
			}
			continue
		}
		pairs = append(pairs, filePair{
			BeforePath: beforePath,
			AfterPath:  "",
		})
	}

	batchPath := filepath.Join(cfg.OutputDir, "ast-batch.json")
	batchData, err := json.MarshalIndent(pairs, "", "  ")
	if err != nil {
		os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("failed to marshal batch file: %w", err)
	}

	if err := os.WriteFile(batchPath, batchData, 0600); err != nil {
		os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("failed to write batch file: %w", err)
	}

	return batchPath, tempDir, nil
}

func extractFileFromGit(ref, filePath, tempDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("invalid file path: empty")
	}

	if filepath.IsAbs(filePath) {
		return "", fmt.Errorf("invalid file path: absolute paths are not allowed")
	}
	cleaned := filepath.Clean(filePath)
	if strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return "", fmt.Errorf("invalid file path: contains path traversal")
	}

	relPath, err := filepath.Rel(".", cleaned)
	if err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}
	if strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." {
		return "", fmt.Errorf("invalid file path: contains path traversal")
	}

	destPath := filepath.Join(tempDir, relPath)
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create temp subdirectory: %w", err)
	}

	cleanTemp := filepath.Clean(tempDir)
	cleanDest := filepath.Clean(destPath)
	if !strings.HasPrefix(cleanDest, cleanTemp+string(filepath.Separator)) && cleanDest != cleanTemp {
		return "", fmt.Errorf("invalid file path: escapes temp directory")
	}

	client := git.NewClient("")
	output, err := client.ShowFile(ref, filePath)
	if err != nil {
		return "", fmt.Errorf("git show failed: %w", err)
	}

	if err := os.WriteFile(destPath, output, 0600); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return destPath, nil
}
