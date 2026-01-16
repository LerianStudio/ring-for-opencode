// Package main implements the call-graph binary for call graph analysis.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/callgraph"
	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/output"
)

// FuncSig is a partial representation of ast.FuncSig, extracting only fields
// needed for call graph analysis. This avoids importing the full ast package
// and allows flexible unmarshalling of Phase 2 output.
type FuncSig struct {
	Receiver string `json:"receiver,omitempty"`
}

// FunctionDiff is a partial representation of ast.FunctionDiff for Phase 2 output.
type FunctionDiff struct {
	Name       string   `json:"name"`
	ChangeType string   `json:"change_type"` // "added", "modified", "removed", "renamed"
	Before     *FuncSig `json:"before,omitempty"`
	After      *FuncSig `json:"after,omitempty"`
}

// SemanticDiff is a partial representation of ast.SemanticDiff for Phase 2 output.
// Only the fields needed for call graph analysis are included.
type SemanticDiff struct {
	Language  string         `json:"language"`
	FilePath  string         `json:"file_path"`
	Functions []FunctionDiff `json:"functions"`
}

// Change type constants matching ast.ChangeType values.
const (
	changeTypeRemoved = "removed"
)

var (
	astFile   = flag.String("ast", "", "Path to {lang}-ast.json from Phase 2 (required)")
	outputDir = flag.String("output", ".ring/codereview", "Output directory")
	timeout   = flag.Int("timeout", 30, "Time budget in seconds, 0 = no limit")
	language  = flag.String("lang", "", "Language override (go, typescript, python) - auto-detect from AST filename if not specified")
	verbose   = flag.Bool("v", false, "Enable verbose output")
)

func init() {
	flag.BoolVar(verbose, "verbose", false, "Enable verbose output")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Call Graph Analyzer - Analyze call relationships for modified functions\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Analyze Go code changes:\n")
		fmt.Fprintf(os.Stderr, "  %s -ast .ring/codereview/go-ast.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Analyze TypeScript with explicit language:\n")
		fmt.Fprintf(os.Stderr, "  %s -ast ast-output.json -lang typescript\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Custom output directory and timeout:\n")
		fmt.Fprintf(os.Stderr, "  %s -ast go-ast.json -output ./output -timeout 60\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Supported languages: %s\n", strings.Join(callgraph.SupportedLanguagesNormalized(), ", "))
	}
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Validate required flags
	if *astFile == "" {
		return fmt.Errorf("-ast flag is required")
	}

	// Detect or validate language
	lang := *language
	if lang == "" {
		lang = detectLanguage(*astFile)
		if lang == "" {
			return fmt.Errorf("could not detect language from filename %q, use -lang flag", filepath.Base(*astFile))
		}
	}
	lang = callgraph.NormalizeLanguage(lang)

	if !callgraph.IsSupported(lang) {
		return fmt.Errorf("unsupported language: %s (supported: %s)", lang, strings.Join(callgraph.SupportedLanguagesNormalized(), ", "))
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "AST input: %s\n", *astFile)
		fmt.Fprintf(os.Stderr, "Language: %s\n", lang)
		fmt.Fprintf(os.Stderr, "Output directory: %s\n", *outputDir)
		fmt.Fprintf(os.Stderr, "Time budget: %d seconds\n", *timeout)
	}

	// Read AST input with size limit
	astData, err := fileutil.ReadJSONFileWithLimit(*astFile)
	if err != nil {
		return fmt.Errorf("failed to read AST file: %w", err)
	}

	// Parse AST input - handle both single SemanticDiff and []SemanticDiff (batch mode)
	var diffs []SemanticDiff
	errArray := json.Unmarshal(astData, &diffs)
	if errArray != nil {
		// Try single object
		var single SemanticDiff
		errSingle := json.Unmarshal(astData, &single)
		if errSingle != nil {
			joined := errors.Join(
				fmt.Errorf("json.Unmarshal(astData, &diffs) into []SemanticDiff failed: %w", errArray),
				fmt.Errorf("json.Unmarshal(astData, &single) into SemanticDiff failed: %w", errSingle),
			)
			return fmt.Errorf(
				"failed to parse astData with json.Unmarshal as diffs ([]SemanticDiff) or single (SemanticDiff) (astData preview=%q): %w",
				previewASTData(astData),
				joined,
			)
		}
		diffs = []SemanticDiff{single}
	}

	// Ensure diffs is not nil (can be nil for empty JSON array or null)
	if diffs == nil {
		diffs = []SemanticDiff{}
	}

	// Build modified functions list from all diffs
	modifiedFuncs := buildModifiedFunctions(diffs)

	if *verbose {
		fmt.Fprintf(os.Stderr, "Modified functions: %d\n", len(modifiedFuncs))
		for _, f := range modifiedFuncs {
			if f.Receiver != "" {
				fmt.Fprintf(os.Stderr, "  - (%s).%s in %s\n", f.Receiver, f.Name, f.File)
			} else {
				fmt.Fprintf(os.Stderr, "  - %s in %s\n", f.Name, f.File)
			}
		}
	}

	// Check for empty input
	if len(modifiedFuncs) == 0 {
		if *verbose {
			fmt.Fprintf(os.Stderr, "No modified functions found, generating empty result\n")
		}
		// Generate empty result
		result := &callgraph.CallGraphResult{
			Language:          lang,
			ModifiedFunctions: []callgraph.FunctionCallGraph{},
			ImpactAnalysis: callgraph.ImpactAnalysis{
				AffectedPackages: []string{},
			},
		}
		return writeResults(result)
	}

	// Determine working directory for analysis
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create analyzer
	analyzer, err := callgraph.NewAnalyzer(lang, workDir)
	if err != nil {
		return fmt.Errorf("failed to create analyzer: %w", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Running call graph analysis...\n")
	}

	// Run analysis
	result, err := analyzer.Analyze(modifiedFuncs, *timeout)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Write results
	return writeResults(result)
}

// detectLanguage extracts language from AST filename prefix.
// Expected formats: go-ast.json, ts-ast.json, py-ast.json
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

// buildModifiedFunctions converts []SemanticDiff to []callgraph.ModifiedFunction.
// It extracts functions with change_type "added" or "modified" (not "removed").
func buildModifiedFunctions(diffs []SemanticDiff) []callgraph.ModifiedFunction {
	var funcs []callgraph.ModifiedFunction

	for _, diff := range diffs {
		for _, f := range diff.Functions {
			// Skip removed functions - they don't exist in the current codebase
			if f.ChangeType == changeTypeRemoved {
				continue
			}

			// Extract receiver from After signature (current state) or Before if After is nil
			var receiver string
			if f.After != nil && f.After.Receiver != "" {
				receiver = f.After.Receiver
			} else if f.Before != nil && f.Before.Receiver != "" {
				receiver = f.Before.Receiver
			}

			// Extract package from file path (directory containing the file)
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

// extractPackageFromPath extracts the package name from a file path.
// For Go files, this is typically the parent directory name.
func extractPackageFromPath(filePath string) string {
	dir := filepath.Dir(filePath)
	if dir == "." || dir == "" {
		return "main"
	}
	return filepath.Base(dir)
}

// writeResults writes call graph results to output files.
func writeResults(result *callgraph.CallGraphResult) error {
	// Write JSON output
	if err := output.WriteJSON(result, *outputDir); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}

	// Write impact summary markdown
	if err := output.WriteImpactSummary(result, *outputDir); err != nil {
		return fmt.Errorf("failed to write impact summary: %w", err)
	}

	// Print summary to stdout
	printSummary(result)

	return nil
}

func previewASTData(astData []byte) string {
	const max = 512
	if len(astData) <= max {
		return string(astData)
	}
	return string(astData[:max]) + "...(truncated)"
}

// printSummary prints analysis summary to stdout.
func printSummary(result *callgraph.CallGraphResult) {
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

	fmt.Printf("  Output: %s/%s-calls.json\n", *outputDir, result.Language)
	fmt.Printf("  Summary: %s/impact-summary.md\n", *outputDir)

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}
}
