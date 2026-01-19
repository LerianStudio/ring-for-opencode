// Package main implements the run-all orchestrator CLI for the codereview pre-analysis pipeline.
// It executes all 6 phases sequentially with timeout handling and graceful degradation.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/git"
)

// scopeJSON represents the structure of scope.json from Phase 0.
type scopeJSON struct {
	BaseRef   string        `json:"base_ref"`
	HeadRef   string        `json:"head_ref"`
	Language  string        `json:"language"`
	Languages []string      `json:"languages"`
	Files     filesByStatus `json:"files"`
}

// filesByStatus holds categorized file lists from scope.json.
type filesByStatus struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
}

func normalizeScopeJSON(scope *scopeJSON) {
	if scope == nil {
		return
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
	if scope.Languages == nil {
		scope.Languages = []string{}
	}
}

// filePair represents a before/after file pair for AST extraction.
type filePair struct {
	BeforePath string `json:"before_path"`
	AfterPath  string `json:"after_path"`
}

// astTempState holds temporary state for AST phase execution.
var astTempState struct {
	batchFile string
	tempDir   string
}

// readScopeJSON reads and parses scope.json from the output directory.
func readScopeJSON(outputDir string) (*scopeJSON, error) {
	scopePath := filepath.Join(outputDir, "scope.json")
	data, err := fileutil.ReadJSONFileWithLimit(scopePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scope.json: %w", err)
	}

	var scope scopeJSON
	if err := json.Unmarshal(data, &scope); err != nil {
		return nil, fmt.Errorf("failed to parse scope.json: %w", err)
	}
	normalizeScopeJSON(&scope)

	return &scope, nil
}

// shouldSkipForNoFiles skips phases if no files are detected.
func shouldSkipForNoFiles(cfg *config) (bool, string) {
	scope, err := readScopeJSON(cfg.outputDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, "scope.json missing - skipping dependent phase"
		}
		return false, ""
	}

	totalFiles := len(scope.Files.Modified) + len(scope.Files.Added) + len(scope.Files.Deleted)
	if totalFiles == 0 {
		return true, "No changed files detected"
	}
	return false, ""
}

// shouldSkipForUnknownLanguage skips phases when the language is unknown.
func shouldSkipForUnknownLanguage(cfg *config) (bool, string) {
	scope, err := readScopeJSON(cfg.outputDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, "scope.json missing - skipping dependent phase"
		}
		return false, ""
	}

	language := strings.ToLower(strings.TrimSpace(scope.Language))
	if language == "" || language == "unknown" {
		return true, "Unknown language detected"
	}
	return false, ""
}

// shouldSkipForMissingScope skips phases that require scope.json.
func shouldSkipForMissingScope(cfg *config) (bool, string) {
	_, err := readScopeJSON(cfg.outputDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, "scope.json missing - skipping dependent phase"
		}
		return false, "" // Let phase fail with its own error
	}
	return false, ""
}

func shouldSkipForNoFilesOrUnknownLanguage(cfg *config) (bool, string) {
	if skip, reason := shouldSkipForNoFiles(cfg); skip {
		return skip, reason
	}
	return shouldSkipForUnknownLanguage(cfg)
}

// generateASTBatchFile creates a batch JSON file for ast-extractor from scope data.
// It handles modified, added, and deleted files appropriately.
// For modified files, it extracts the "before" version from git to a temp directory.
func generateASTBatchFile(cfg *config) (string, string, error) {
	scope, err := readScopeJSON(cfg.outputDir)
	if err != nil {
		return "", "", err
	}

	// Create temp directory for "before" files
	tempDir, err := os.MkdirTemp("", "ast-before-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	var pairs []filePair

	// Handle modified files - need both before (from git) and after (current)
	for _, file := range scope.Files.Modified {
		beforePath, err := extractFileFromGit(cfg.baseRef, file, tempDir)
		if err != nil {
			return "", "", fmt.Errorf("failed to extract %s from %s: %w", file, cfg.baseRef, err)
		}
		pairs = append(pairs, filePair{
			BeforePath: beforePath,
			AfterPath:  file,
		})
	}

	// Handle added files - only after path
	for _, file := range scope.Files.Added {
		pairs = append(pairs, filePair{
			BeforePath: "",
			AfterPath:  file,
		})
	}

	// Handle deleted files - only before path (extract from git)
	for _, file := range scope.Files.Deleted {
		beforePath, err := extractFileFromGit(cfg.baseRef, file, tempDir)
		if err != nil {
			if cfg.verbose {
				fmt.Fprintf(os.Stderr, "Warning: could not extract deleted file %s from %s: %v\n", file, cfg.baseRef, err)
			}
			continue
		}
		pairs = append(pairs, filePair{
			BeforePath: beforePath,
			AfterPath:  "",
		})
	}

	// Write batch file
	batchPath := filepath.Join(cfg.outputDir, "ast-batch.json")
	data, err := json.MarshalIndent(pairs, "", "  ")
	if err != nil {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp directory %s: %v\n", tempDir, removeErr)
		}
		return "", "", fmt.Errorf("failed to marshal batch file: %w", err)
	}

	if err := os.WriteFile(batchPath, data, 0600); err != nil {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp directory %s: %v\n", tempDir, removeErr)
		}
		return "", "", fmt.Errorf("failed to write batch file: %w", err)
	}

	return batchPath, tempDir, nil
}

// extractFileFromGit extracts a file from a git ref to a temp directory.
// Returns the path to the extracted file.
func extractFileFromGit(ref, filePath, tempDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("invalid file path: empty")
	}

	// Reject absolute paths and traversal attempts
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

	// Create subdirectories in temp to preserve file structure
	destPath := filepath.Join(tempDir, relPath)
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create temp subdirectory: %w", err)
	}

	// Ensure destination remains within temp directory
	cleanTemp := filepath.Clean(tempDir)
	cleanDest := filepath.Clean(destPath)
	if !strings.HasPrefix(cleanDest, cleanTemp+string(filepath.Separator)) && cleanDest != cleanTemp {
		return "", fmt.Errorf("invalid file path: escapes temp directory")
	}

	// Use git show to extract file content
	client := git.NewClient("")
	output, err := client.ShowFile(ref, filePath)
	if err != nil {
		return "", fmt.Errorf("git show failed: %w", err)
	}

	// Write to temp file
	if err := os.WriteFile(destPath, output, 0600); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return destPath, nil
}

// detectASTOutputFile finds the AST output file from the output directory.
// Returns the path to the first found AST file (go-ast.json, typescript-ast.json, python-ast.json).
func detectASTOutputFile(outputDir string) (string, error) {
	// Check for language-specific AST files in priority order
	candidates := []string{
		filepath.Join(outputDir, "go-ast.json"),
		filepath.Join(outputDir, "typescript-ast.json"),
		filepath.Join(outputDir, "python-ast.json"),
		filepath.Join(outputDir, "mixed-ast.json"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no AST output file found in %s", outputDir)
}

// cleanupASTTempFiles removes temporary files created during AST phase.
func cleanupASTTempFiles() {
	if astTempState.tempDir != "" {
		if err := os.RemoveAll(astTempState.tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp directory %s: %v\n", astTempState.tempDir, err)
		}
		astTempState.tempDir = ""
	}
}

// version is set via ldflags during build.
var version = "dev"

// Phase represents a phase in the codereview pipeline.
type Phase struct {
	Name    string
	Binary  string
	Timeout time.Duration
	Args    func(cfg *config) []string
	Skip    func(cfg *config) (bool, string) // Optional skip condition
}

// config holds the parsed CLI configuration.
type config struct {
	baseRef     string
	headRef     string
	files       string
	filesFrom   string
	unstaged    bool
	outputDir   string
	binDir      string
	skip        string
	verbose     bool
	showVersion bool
}

// PhaseResult captures the outcome of a phase execution.
type PhaseResult struct {
	Name       string
	Duration   time.Duration
	Success    bool
	Error      error
	Skipped    bool
	SkipReason string // Reason for skipping (when Skipped is true)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags parses command-line flags and returns the configuration.
func parseFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.baseRef, "base", "main", "Base git reference (commit/branch)")
	flag.StringVar(&cfg.headRef, "head", "HEAD", "Head git reference (commit/branch)")
	flag.StringVar(&cfg.files, "files", "", "Comma-separated file patterns to analyze (mutually exclusive with --base/--head)")
	flag.StringVar(&cfg.filesFrom, "files-from", "", "Path to file containing file patterns (one per line)")
	flag.BoolVar(&cfg.unstaged, "unstaged", false, "Analyze only unstaged and untracked files")
	flag.StringVar(&cfg.outputDir, "output", ".ring/codereview", "Output directory for all phase results")
	flag.StringVar(&cfg.binDir, "bin-dir", "", "Directory containing phase binaries (auto-detect from executable path)")
	flag.StringVar(&cfg.skip, "skip", "", "Comma-separated list of phases to skip (e.g., 'static-analysis,callgraph')")
	flag.BoolVar(&cfg.verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&cfg.verbose, "v", false, "Enable verbose output (shorthand)")
	flag.BoolVar(&cfg.showVersion, "version", false, "Show version and exit")

	flag.Usage = printUsage
	flag.Parse()

	filesSelected := cfg.files != "" || cfg.filesFrom != ""
	baseSet := false
	headSet := false
	flag.CommandLine.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "base":
			baseSet = true
		case "head":
			headSet = true
		}
	})
	if cfg.unstaged && (filesSelected || baseSet || headSet) {
		fmt.Fprintln(os.Stderr, "Error: --unstaged cannot be used with --files/--files-from or --base/--head")
		os.Exit(2)
	}
	if filesSelected && (baseSet || headSet) {
		fmt.Fprintln(os.Stderr, "Error: --files/--files-from cannot be used with --base/--head")
		os.Exit(2)
	}

	return cfg
}

// printUsage prints comprehensive help information.
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: run-all [options]\n\n")
	fmt.Fprintf(os.Stderr, "Orchestrates all 6 phases of the codereview pre-analysis pipeline.\n\n")
	fmt.Fprintf(os.Stderr, "Phases (executed in order):\n")
	fmt.Fprintf(os.Stderr, "  1. scope          Detect changed files and project language (30s timeout)\n")
	fmt.Fprintf(os.Stderr, "  2. static-analysis Run linters and static analyzers (5m timeout)\n")
	fmt.Fprintf(os.Stderr, "  3. ast            Extract semantic diff from AST (2m timeout)\n")
	fmt.Fprintf(os.Stderr, "  4. callgraph      Analyze call relationships (3m timeout)\n")
	fmt.Fprintf(os.Stderr, "  5. dataflow       Security-focused data flow analysis (3m timeout)\n")
	fmt.Fprintf(os.Stderr, "  6. context        Compile reviewer context files (30s timeout)\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Run all phases comparing main to HEAD:\n")
	fmt.Fprintf(os.Stderr, "  run-all\n\n")
	fmt.Fprintf(os.Stderr, "  # Compare specific branches:\n")
	fmt.Fprintf(os.Stderr, "  run-all --base=develop --head=feature/my-branch\n\n")

	fmt.Fprintf(os.Stderr, "  # Analyze specific files:\n")
	fmt.Fprintf(os.Stderr, "  run-all --files=cmd/*.go,scripts/**/*.ts\n\n")

	fmt.Fprintf(os.Stderr, "  # Analyze files from a list:\n")
	fmt.Fprintf(os.Stderr, "  run-all --files-from=.ring/filelist.txt\n\n")

	fmt.Fprintf(os.Stderr, "  # Analyze unstaged + untracked files:\n")
	fmt.Fprintf(os.Stderr, "  run-all --unstaged\n\n")

	fmt.Fprintf(os.Stderr, "  # Custom output directory:\n")
	fmt.Fprintf(os.Stderr, "  run-all --output=./review-output\n\n")

	fmt.Fprintf(os.Stderr, "  # Skip specific phases:\n")
	fmt.Fprintf(os.Stderr, "  run-all --skip=static-analysis,dataflow\n\n")
	fmt.Fprintf(os.Stderr, "  # Verbose mode with custom binary directory:\n")
	fmt.Fprintf(os.Stderr, "  run-all --verbose --bin-dir=/path/to/binaries\n\n")
	fmt.Fprintf(os.Stderr, "Phase skip aliases:\n")
	fmt.Fprintf(os.Stderr, "  scope, static-analysis, ast, callgraph, dataflow, context\n\n")
	fmt.Fprintf(os.Stderr, "Output files (in output directory):\n")
	fmt.Fprintf(os.Stderr, "  scope.json            Changed files and language\n")
	fmt.Fprintf(os.Stderr, "  static-analysis.json  Linter findings\n")
	fmt.Fprintf(os.Stderr, "  {lang}-ast.json       Semantic AST diff\n")
	fmt.Fprintf(os.Stderr, "  {lang}-calls.json     Call graph analysis\n")
	fmt.Fprintf(os.Stderr, "  {lang}-flow.json      Data flow analysis\n")
	fmt.Fprintf(os.Stderr, "  context-*.md          Reviewer-specific context files\n")
}

// getPhases returns the ordered list of phases to execute.
func getPhases() []Phase {
	return []Phase{
		{
			Name:    "scope",
			Binary:  "scope-detector",
			Timeout: 30 * time.Second,
			Args: func(cfg *config) []string {
				args := []string{
					"--output", filepath.Join(cfg.outputDir, "scope.json"),
				}
				if cfg.unstaged {
					return append(args, "--unstaged")
				}
				if cfg.files != "" || cfg.filesFrom != "" {
					if cfg.files != "" {
						args = append(args, "--files", cfg.files)
					}
					if cfg.filesFrom != "" {
						args = append(args, "--files-from", cfg.filesFrom)
					}
					return args
				}

				return append(args, "--base", cfg.baseRef, "--head", cfg.headRef)

			},
		},
		{
			Name:    "static-analysis",
			Binary:  "static-analysis",
			Timeout: 5 * time.Minute,
			Args: func(cfg *config) []string {
				return []string{
					"--scope", filepath.Join(cfg.outputDir, "scope.json"),
					"--output", cfg.outputDir,
				}
			},
		},
		{
			Name:    "ast",
			Binary:  "ast-extractor",
			Timeout: 2 * time.Minute,
			Skip:    shouldSkipForNoFiles,
			Args: func(cfg *config) []string {
				// Generate batch file from scope.json
				// This reads scope.json, extracts before-files from git, and creates ast-batch.json
				batchPath, tempDir, err := generateASTBatchFile(cfg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to generate AST batch file: %v\n", err)
					return nil
				}
				// Store temp state for cleanup
				astTempState.batchFile = batchPath
				astTempState.tempDir = tempDir
				return []string{
					"--batch", batchPath,
					"--output", "json",
				}
			},
		},
		{
			Name:    "callgraph",
			Binary:  "call-graph",
			Timeout: 3 * time.Minute,
			Skip:    shouldSkipForNoFilesOrUnknownLanguage,
			Args: func(cfg *config) []string {
				// Detect which AST output file exists from previous phase
				astFile, err := detectASTOutputFile(cfg.outputDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
					return nil
				}

				callgraphLangs, langErr := languagesForCallgraph(cfg.outputDir)
				if langErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: %v\n", langErr)
					return nil
				}
				args := []string{
					"--ast", astFile,
					"--output", cfg.outputDir,
				}
				if len(callgraphLangs) > 1 {
					if err := writeCallgraphLanguageFile(cfg.outputDir, callgraphLangs); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
						return nil
					}
					args = append(args, "--languages-file", callgraphLanguagesFile(cfg.outputDir))
				}
				if len(callgraphLangs) == 1 {
					args = append(args, "--lang", callgraphLangs[0])
				}
				return args

			},
		},
		{
			Name:    "dataflow",
			Binary:  "data-flow",
			Timeout: 3 * time.Minute,
			Skip:    shouldSkipForMissingScope,
			Args: func(cfg *config) []string {
				return []string{
					"--scope", filepath.Join(cfg.outputDir, "scope.json"),
					"--output", cfg.outputDir,
				}
			},
		},
		{
			Name:    "context",
			Binary:  "compile-context",
			Timeout: 30 * time.Second,
			Args: func(cfg *config) []string {
				return []string{
					"--input", cfg.outputDir,
					"--output", cfg.outputDir,
				}
			},
		},
	}
}

// validateBinDir validates the bin-dir path for security.
// It uses proper path resolution to prevent path traversal attacks and verifies the directory exists.
func validateBinDir(binDir string) error {
	_, err := fileutil.ValidateDirectory(binDir, "")
	return err
}

// run executes the main CLI logic.
func run() error {
	cfg := parseFlags()

	// Handle --version flag
	if cfg.showVersion {
		fmt.Printf("run-all version %s\n", version)
		return nil
	}

	// Auto-detect bin directory from executable path if not provided
	if cfg.binDir == "" {
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine executable path: %w", err)
		}
		cfg.binDir = filepath.Dir(execPath)
	}

	// Validate bin-dir before executing any phases
	if err := validateBinDir(cfg.binDir); err != nil {
		return fmt.Errorf("bin-dir validation failed: %w", err)
	}

	if cfg.verbose {
		fmt.Fprintf(os.Stderr, "Configuration:\n")
		fmt.Fprintf(os.Stderr, "  Base ref: %s\n", cfg.baseRef)
		fmt.Fprintf(os.Stderr, "  Head ref: %s\n", cfg.headRef)
		if cfg.unstaged {
			fmt.Fprintln(os.Stderr, "  Mode: unstaged + untracked")
		}
		if cfg.files != "" {
			fmt.Fprintf(os.Stderr, "  Files: %s\n", cfg.files)
		}
		if cfg.filesFrom != "" {
			fmt.Fprintf(os.Stderr, "  Files from: %s\n", cfg.filesFrom)
		}
		fmt.Fprintf(os.Stderr, "  Output directory: %s\n", cfg.outputDir)
		fmt.Fprintf(os.Stderr, "  Binary directory: %s\n", cfg.binDir)
		if cfg.skip != "" {
			fmt.Fprintf(os.Stderr, "  Skipping phases: %s\n", cfg.skip)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Create output directory if needed
	validatedOutputDir, err := fileutil.ValidatePath(cfg.outputDir, ".")
	if err != nil {
		return fmt.Errorf("invalid output directory: %w", err)
	}
	cfg.outputDir = validatedOutputDir
	if err := os.MkdirAll(cfg.outputDir, 0o700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Parse skip list
	skipSet := parseSkipList(cfg.skip)

	// Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Get phases and execute
	phases := getPhases()
	results := make([]PhaseResult, 0, len(phases))
	startTime := time.Now()

	for _, phase := range phases {
		// Check for cancellation before starting each phase
		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "\nInterrupted. Stopping execution.\n")
			cleanupASTTempFiles() // Cleanup on interrupt
			interruptSummary := buildSummary(results)
			printSummary(interruptSummary, time.Since(startTime))
			return fmt.Errorf("execution interrupted")
		default:
		}

		result := executePhase(ctx, cfg, phase, skipSet)
		results = append(results, result)

		// Cleanup AST temp files after AST phase completes
		if phase.Name == "ast" {
			cleanupASTTempFiles()
		}

		// Print progress
		if result.Skipped {
			if result.SkipReason != "" {
				fmt.Fprintf(os.Stderr, "[SKIP] %s: %s\n", phase.Name, result.SkipReason)
			} else {
				fmt.Fprintf(os.Stderr, "[SKIP] %s\n", phase.Name)
			}
		} else if result.Success {
			fmt.Fprintf(os.Stderr, "[PASS] %s (%v)\n", phase.Name, result.Duration.Round(time.Millisecond))
		} else {
			fmt.Fprintf(os.Stderr, "[FAIL] %s (%v): %v\n", phase.Name, result.Duration.Round(time.Millisecond), result.Error)
		}
	}

	totalDuration := time.Since(startTime)

	// Print summary
	summary := buildSummary(results)
	printSummary(summary, totalDuration)

	// Determine exit code
	for _, result := range results {
		if !result.Success && !result.Skipped {
			return fmt.Errorf("one or more phases failed")
		}
	}

	return nil
}

// parseSkipList parses a comma-separated skip list into a set.
func parseSkipList(skip string) map[string]bool {
	skipSet := make(map[string]bool)
	if skip == "" {
		return skipSet
	}

	for _, name := range strings.Split(skip, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			skipSet[name] = true
		}
	}

	// Validate skip list against known phases
	knownPhases := map[string]bool{
		"scope":           true,
		"static-analysis": true,
		"ast":             true,
		"callgraph":       true,
		"dataflow":        true,
		"context":         true,
	}
	for phase := range skipSet {
		if !knownPhases[phase] {
			fmt.Fprintf(os.Stderr, "Warning: unknown phase '%s' in skip list\n", phase)
		}
	}

	return skipSet
}

// executePhase executes a single phase with timeout handling.
func executePhase(ctx context.Context, cfg *config, phase Phase, skipSet map[string]bool) PhaseResult {
	result := PhaseResult{
		Name: phase.Name,
	}

	// Check if phase should be skipped via CLI flag
	if skipSet[phase.Name] {
		result.Skipped = true
		result.Success = true
		result.SkipReason = "skipped via --skip flag"
		return result
	}

	// Check if phase has a Skip condition (e.g., unknown language)
	if phase.Skip != nil {
		if skip, reason := phase.Skip(cfg); skip {
			result.Skipped = true
			result.Success = true
			result.SkipReason = reason
			return result
		}
	}

	startTime := time.Now()

	// Build command arguments
	args := phase.Args(cfg)
	if args == nil {
		result.Duration = time.Since(startTime)
		result.Error = fmt.Errorf("failed to build phase arguments")
		return result
	}
	if cfg.verbose {
		args = append(args, "-v")
	}

	// Build binary path
	binaryPath := filepath.Join(cfg.binDir, phase.Binary)

	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		result.Duration = time.Since(startTime)
		result.Error = fmt.Errorf("binary not found: %s", binaryPath)
		return result
	}

	// Create a context with timeout for this phase
	// CommandContext will automatically kill the process when context is cancelled
	phaseCtx, phaseCancel := context.WithTimeout(ctx, phase.Timeout)
	defer phaseCancel()

	// Use CommandContext for automatic process cleanup on cancellation/timeout
	cmd := exec.CommandContext(phaseCtx, binaryPath, args...) // #nosec G204 - binary and args are internal/validated
	cmd.Env = append([]string{"LC_ALL=C"}, os.Environ()...)

	wd, err := os.Getwd()
	if err != nil {
		result.Duration = time.Since(startTime)
		result.Error = fmt.Errorf("failed to get working directory: %w", err)
		return result
	}
	cmd.Dir = wd

	// For AST phase, capture stdout to write to file (ast-extractor outputs JSON to stdout)
	var stdoutBuf *bytes.Buffer
	if phase.Name == "ast" {
		stdoutBuf = &bytes.Buffer{}
		cmd.Stdout = stdoutBuf
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	// Run the command - CommandContext handles cancellation automatically
	err = cmd.Run()
	result.Duration = time.Since(startTime)

	// Check if the error was due to context cancellation or timeout
	if phaseCtx.Err() != nil {
		if ctx.Err() != nil {
			// Parent context was cancelled (interrupt)
			result.Error = fmt.Errorf("interrupted")
		} else {
			// Phase context timed out
			result.Error = fmt.Errorf("timeout after %v", phase.Timeout)
		}
		return result
	}

	result.Error = err
	result.Success = err == nil

	// For AST phase, write captured stdout to language-specific AST file
	if phase.Name == "ast" && result.Success && stdoutBuf != nil {
		scope, scopeErr := readScopeJSON(cfg.outputDir)
		if scopeErr != nil {
			result.Error = fmt.Errorf("failed to read scope.json for AST output: %w", scopeErr)
			result.Success = false
		} else {
			language := scope.Language
			if language == "" {
				language = "unknown"
			}
			astOutputPath := filepath.Join(cfg.outputDir, fmt.Sprintf("%s-ast.json", language))
			if writeErr := os.WriteFile(astOutputPath, stdoutBuf.Bytes(), 0600); writeErr != nil {
				result.Error = fmt.Errorf("failed to write AST output to %s: %w", astOutputPath, writeErr)
				result.Success = false
			} else if cfg.verbose {
				fmt.Fprintf(os.Stderr, "  AST output written to: %s\n", astOutputPath)
			}
		}
	}

	return result
}

type summary struct {
	Passed       int
	Failed       int
	Skipped      int
	FailedPhases []PhaseResult
	Results      []PhaseResult
}

func buildSummary(results []PhaseResult) summary {
	result := summary{Results: results}
	for _, phase := range results {
		if phase.Skipped {
			result.Skipped++
			continue
		}
		if phase.Success {
			result.Passed++
			continue
		}
		result.Failed++
		result.FailedPhases = append(result.FailedPhases, phase)
	}
	return result
}

// printSummary prints the execution summary.
func printSummary(summary summary, totalDuration time.Duration) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "========================================\n")
	fmt.Fprintf(os.Stderr, "  Code Review Pre-Analysis Summary\n")
	fmt.Fprintf(os.Stderr, "========================================\n")
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintf(os.Stderr, "Phases: %d passed, %d failed, %d skipped\n", summary.Passed, summary.Failed, summary.Skipped)
	fmt.Fprintf(os.Stderr, "Total time: %v\n", totalDuration.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "\n")

	// Phase breakdown
	fmt.Fprintf(os.Stderr, "Phase Breakdown:\n")
	for _, result := range summary.Results {
		if result.Skipped {
			if result.SkipReason != "" {
				fmt.Fprintf(os.Stderr, "  [SKIP] %-20s %s\n", result.Name, result.SkipReason)
			} else {
				fmt.Fprintf(os.Stderr, "  [SKIP] %-20s\n", result.Name)
			}
		} else if result.Success {
			fmt.Fprintf(os.Stderr, "  [PASS] %-20s %v\n", result.Name, result.Duration.Round(time.Millisecond))
		} else {
			fmt.Fprintf(os.Stderr, "  [FAIL] %-20s %v\n", result.Name, result.Duration.Round(time.Millisecond))
		}
	}
	fmt.Fprintf(os.Stderr, "\n")

	// Failed phases detail
	if len(summary.FailedPhases) > 0 {
		fmt.Fprintf(os.Stderr, "Failed phases:\n")
		for _, phase := range summary.FailedPhases {
			fmt.Fprintf(os.Stderr, "  - %s: %v\n", phase.Name, phase.Error)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	if summary.Failed == 0 {
		fmt.Fprintf(os.Stderr, "All phases completed successfully.\n")
	} else {
		fmt.Fprintf(os.Stderr, "Some phases failed. Review output for details.\n")
	}
}
