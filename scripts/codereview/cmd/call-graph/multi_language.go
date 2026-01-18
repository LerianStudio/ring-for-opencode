package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/callgraph"
	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/output"
)

type languagesPayload struct {
	Languages []string `json:"languages"`
}

func readLanguagesFile(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}

	validated, err := fileutil.ValidatePath(path, ".")
	if err != nil {
		return nil, fmt.Errorf("invalid languages file path: %w", err)
	}
	data, err := fileutil.ReadJSONFileWithLimit(validated)
	if err != nil {
		return nil, fmt.Errorf("failed to read languages file: %w", err)
	}

	var payload languagesPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse languages file: %w", err)
	}
	if payload.Languages == nil {
		return []string{}, nil
	}
	return payload.Languages, nil
}

func runCallgraphForLanguage(lang string, diffs []SemanticDiff, workDir, suffix string) error {
	lang = callgraph.NormalizeLanguage(lang)
	if lang == "" {
		return fmt.Errorf("invalid language")
	}

	outputDir := *outputDir
	if suffix != "" {
		outputDir = strings.TrimRight(outputDir, string(os.PathSeparator)) + suffix
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "AST input: %s\n", *astFile)
		fmt.Fprintf(os.Stderr, "Language: %s\n", lang)
		fmt.Fprintf(os.Stderr, "Output directory: %s\n", outputDir)
		fmt.Fprintf(os.Stderr, "Time budget: %d seconds\n", *timeout)
	}

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

	if len(modifiedFuncs) == 0 {
		if *verbose {
			fmt.Fprintf(os.Stderr, "No modified functions found, generating empty result\n")
		}
		result := &callgraph.CallGraphResult{
			Language:          lang,
			ModifiedFunctions: []callgraph.FunctionCallGraph{},
			ImpactAnalysis: callgraph.ImpactAnalysis{
				AffectedPackages: []string{},
			},
		}
		return writeResultsWithOutputDir(result, outputDir)
	}

	analyzer, err := callgraph.NewAnalyzer(lang, workDir)
	if err != nil {
		return fmt.Errorf("failed to create analyzer: %w", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Running call graph analysis...\n")
	}

	result, err := analyzer.Analyze(modifiedFuncs, *timeout)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	return writeResultsWithOutputDir(result, outputDir)
}

func writeResultsWithOutputDir(result *callgraph.CallGraphResult, outputDir string) error {
	if err := output.WriteJSON(result, outputDir); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	if err := output.WriteImpactSummary(result, outputDir); err != nil {
		return fmt.Errorf("failed to write markdown output: %w", err)
	}

	printSummaryWithOutputDir(result, outputDir)
	return nil
}

func printSummaryWithOutputDir(result *callgraph.CallGraphResult, outputDir string) {
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
