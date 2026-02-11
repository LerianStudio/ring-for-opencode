// Package callgraph implements Phase 3: call graph analysis for the unified CLI.
package callgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/callgraph"
	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/output"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/validate"
)

// Phase implements the call graph analysis phase.
type Phase struct{}

// New creates a new call graph analysis phase.
func New() *Phase {
	return &Phase{}
}

// Name returns the canonical name of this phase.
func (p *Phase) Name() string {
	return "callgraph"
}

// Timeout returns the default timeout for this phase.
func (p *Phase) Timeout() time.Duration {
	return 3 * time.Minute
}

// ShouldSkip checks if this phase should be skipped.
func (p *Phase) ShouldSkip(cfg *phases.Config) (bool, string) {
	// Skip if scope.json is missing or has no files/unknown language
	scopePath := cfg.ScopePath
	if scopePath == "" {
		scopePath = filepath.Join(cfg.OutputDir, "scope.json")
	}

	data, err := fileutil.ReadJSONFileWithLimit(scopePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, "scope.json missing - skipping callgraph phase"
		}
		return false, ""
	}

	var scope struct {
		Language string `json:"language"`
		Files    struct {
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

	language := strings.ToLower(strings.TrimSpace(scope.Language))
	if language == "" || language == "unknown" {
		return true, "Unknown language detected"
	}

	return false, ""
}

// FuncSig is a partial representation of ast.FuncSig.
type FuncSig struct {
	Receiver string `json:"receiver,omitempty"`
}

// FunctionDiff is a partial representation of ast.FunctionDiff.
type FunctionDiff struct {
	Name       string   `json:"name" validate:"required"`
	ChangeType string   `json:"change_type" validate:"required,oneof=added modified removed renamed"`
	Before     *FuncSig `json:"before,omitempty"`
	After      *FuncSig `json:"after,omitempty"`
}

// SemanticDiff is a partial representation of ast.SemanticDiff.
type SemanticDiff struct {
	Language  string         `json:"language" validate:"required"`
	FilePath  string         `json:"file_path" validate:"required"`
	Functions []FunctionDiff `json:"functions" validate:"dive"`
}

const changeTypeRemoved = "removed"

// Run executes the call graph analysis phase.
func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	astFile := cfg.ASTPath
	if astFile == "" {
		// Auto-detect AST file from output directory
		detected, _, found := detectASTOutputFile(cfg.OutputDir)
		if !found {
			return fmt.Errorf("no AST output file found in %s", cfg.OutputDir)
		}
		astFile = detected
	}

	validatedAST, err := fileutil.ValidatePath(astFile, ".")
	if err != nil {
		return fmt.Errorf("invalid AST file path: %w", err)
	}

	// Determine languages to analyze
	languages := []string{}
	if cfg.Language != "" {
		languages = append(languages, cfg.Language)
	}
	if len(languages) == 0 {
		languages = append(languages, detectLanguage(validatedAST))
	}

	astData, err := fileutil.ReadJSONFileWithLimit(validatedAST)
	if err != nil {
		return fmt.Errorf("failed to read AST file: %w", err)
	}

	// Parse AST input
	var diffs []SemanticDiff
	errArray := json.Unmarshal(astData, &diffs)
	if errArray != nil {
		var single SemanticDiff
		errSingle := json.Unmarshal(astData, &single)
		if errSingle != nil {
			return fmt.Errorf("failed to parse AST data: %w", errors.Join(errArray, errSingle))
		}
		diffs = []SemanticDiff{single}
	}

	if diffs == nil {
		diffs = []SemanticDiff{}
	}

	// Validate parsed AST data
	for i := range diffs {
		if err := validate.Validate(&diffs[i]); err != nil {
			return fmt.Errorf("invalid AST data at index %d: %w", i, err)
		}
	}

	if len(languages) == 0 || languages[0] == "" {
		languages = extractLanguagesFromDiffs(diffs)
	}
	languages = normalizeLanguages(languages)
	if len(languages) == 0 {
		return fmt.Errorf("could not detect language from AST data")
	}

	for _, lang := range languages {
		if !callgraph.IsSupported(lang) {
			return fmt.Errorf("unsupported language: %s (supported: %s)", lang, strings.Join(callgraph.SupportedLanguagesNormalized(), ", "))
		}
	}

	workDir := cfg.WorkDir
	if workDir == "" {
		workDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	var runErr error
	for _, lang := range languages {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		langDiffs := filterDiffsByLanguage(diffs, lang)
		if err := runCallgraphForLanguage(ctx, lang, langDiffs, workDir, cfg); err != nil {
			runErr = errors.Join(runErr, err)
		}
	}

	return runErr
}

func runCallgraphForLanguage(_ context.Context, lang string, diffs []SemanticDiff, workDir string, cfg *phases.Config) error {
	modifiedFuncs := buildModifiedFunctions(diffs)

	if cfg.Verbose {
		logger.Debug("analyzing call graph", "language", lang, "functions", len(modifiedFuncs))
	}

	analyzer, err := callgraph.NewAnalyzer(lang, workDir)
	if err != nil {
		return fmt.Errorf("failed to create analyzer for %s: %w", lang, err)
	}

	// Use 30 second time budget (can be made configurable later)
	timeBudgetSec := 30
	result, err := analyzer.Analyze(modifiedFuncs, timeBudgetSec)
	if err != nil {
		return fmt.Errorf("call graph analysis failed for %s: %w", lang, err)
	}

	// Write results
	if err := output.WriteJSON(result, cfg.OutputDir); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}

	if err := output.WriteImpactSummary(result, cfg.OutputDir); err != nil {
		return fmt.Errorf("failed to write impact summary: %w", err)
	}

	printSummary(result, cfg.OutputDir)
	return nil
}

func buildModifiedFunctions(diffs []SemanticDiff) []callgraph.ModifiedFunction {
	var funcs []callgraph.ModifiedFunction
	for _, diff := range diffs {
		for _, f := range diff.Functions {
			if f.ChangeType == changeTypeRemoved {
				continue
			}
			var receiver string
			if f.After != nil && f.After.Receiver != "" {
				receiver = f.After.Receiver
			} else if f.Before != nil && f.Before.Receiver != "" {
				receiver = f.Before.Receiver
			}
			pkg := extractPackageFromPath(diff.FilePath)
			funcs = append(funcs, callgraph.ModifiedFunction{
				Name:     f.Name,
				File:     diff.FilePath,
				Package:  pkg,
				Receiver: receiver,
			})
		}
	}
	return funcs
}

func extractPackageFromPath(filePath string) string {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return "main"
	}
	return filepath.Base(dir)
}

func detectASTOutputFile(outputDir string) (string, string, bool) {
	candidates := []struct {
		lang string
		path string
	}{
		{"go", filepath.Join(outputDir, "go-ast.json")},
		{"typescript", filepath.Join(outputDir, "typescript-ast.json")},
		{"python", filepath.Join(outputDir, "python-ast.json")},
		{"mixed", filepath.Join(outputDir, "mixed-ast.json")},
	}
	for _, c := range candidates {
		if _, err := os.Stat(c.path); err == nil {
			return c.path, c.lang, true
		}
	}
	return "", "", false
}

func detectLanguage(filename string) string {
	base := filepath.Base(filename)
	base = strings.ToLower(base)
	if strings.HasPrefix(base, "go-") || strings.HasPrefix(base, "golang-") {
		return callgraph.LangGo
	}
	if strings.HasPrefix(base, "ts-") || strings.HasPrefix(base, "typescript-") {
		return callgraph.LangTypeScript
	}
	if strings.HasPrefix(base, "py-") || strings.HasPrefix(base, "python-") {
		return callgraph.LangPython
	}
	return ""
}

func extractLanguagesFromDiffs(diffs []SemanticDiff) []string {
	if len(diffs) == 0 {
		return []string{}
	}
	counts := make(map[string]int)
	for _, diff := range diffs {
		lang := callgraph.NormalizeLanguage(diff.Language)
		if lang != "" && callgraph.IsSupported(lang) {
			counts[lang]++
		}
	}
	if len(counts) == 0 {
		return []string{}
	}

	priority := []string{callgraph.LangGo, callgraph.LangTypeScript, callgraph.LangPython}
	ordered := make([]string, 0, len(counts))
	for _, lang := range priority {
		if counts[lang] > 0 {
			ordered = append(ordered, lang)
		}
	}
	for lang := range counts {
		if !containsString(ordered, lang) {
			ordered = append(ordered, lang)
		}
	}
	return ordered
}

func normalizeLanguages(languages []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(languages))
	for _, lang := range languages {
		normalized := callgraph.NormalizeLanguage(lang)
		if normalized == "" || !callgraph.IsSupported(normalized) {
			continue
		}
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}
	return result
}

func filterDiffsByLanguage(diffs []SemanticDiff, lang string) []SemanticDiff {
	if len(diffs) == 0 {
		return diffs
	}
	normalized := callgraph.NormalizeLanguage(lang)
	if normalized == "" {
		return diffs
	}
	filtered := make([]SemanticDiff, 0, len(diffs))
	for _, diff := range diffs {
		if callgraph.NormalizeLanguage(diff.Language) == normalized {
			filtered = append(filtered, diff)
		}
	}
	return filtered
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func printSummary(result *callgraph.CallGraphResult, outputDir string) {
	fmt.Printf("Call graph analysis complete:\n")
	fmt.Printf("  Language: %s\n", result.Language)
	fmt.Printf("  Functions analyzed: %d\n", len(result.ModifiedFunctions))
	fmt.Printf("  Direct callers: %d\n", result.ImpactAnalysis.DirectCallers)
	fmt.Printf("  Transitive callers: %d\n", result.ImpactAnalysis.TransitiveCallers)
	fmt.Printf("  Affected tests: %d\n", result.ImpactAnalysis.AffectedTests)
	fmt.Printf("  Affected packages: %d\n", len(result.ImpactAnalysis.AffectedPackages))

	if result.TimeBudgetExceeded {
		fmt.Printf("  Warning: Time budget exceeded, results may be partial\n")
	}
	if result.PartialResults {
		fmt.Printf("  Warning: Partial results due to analysis limitations\n")
	}

	fmt.Printf("  Output: %s/%s-calls.json\n", outputDir, result.Language)
	fmt.Printf("  Summary: %s/impact-summary.md\n", outputDir)

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}
}
