// Package main implements the static-analysis binary.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/lint"
	"github.com/lerianstudio/ring/scripts/codereview/internal/output"
	"github.com/lerianstudio/ring/scripts/codereview/internal/scope"
)

func main() {
	// Parse flags
	scopePath := flag.String("scope", "", "Path to scope.json (default: .ring/codereview/scope.json)")
	outputPath := flag.String("output", "", "Output directory (default: .ring/codereview/)")
	verbose := flag.Bool("v", false, "Verbose output")
	timeout := flag.Duration("timeout", 5*time.Minute, "Timeout for analysis")
	flag.Parse()

	// Determine project directory (current working directory)
	projectDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Set default paths
	if *scopePath == "" {
		*scopePath = scope.DefaultScopePath(projectDir)
	}
	if *outputPath == "" {
		*outputPath = output.DefaultOutputDir(projectDir)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Read scope
	if *verbose {
		log.Printf("Reading scope from: %s", *scopePath)
	}
	s, err := scope.ReadScopeJSON(*scopePath)
	if err != nil {
		log.Fatalf("Failed to read scope: %v", err)
	}

	// Get language
	lang := s.GetLanguage()
	if *verbose {
		log.Printf("Detected language: %s", lang)
	}

	// Initialize registry and register linters
	registry := lint.NewRegistry()
	registerLinters(registry)

	// Get available linters for detected language
	linters := registry.GetAvailableLinters(ctx, lang)
	if len(linters) == 0 {
		log.Printf("Warning: No linters available for language: %s", lang)
	}

	if *verbose {
		log.Printf("Available linters: %d", len(linters))
		for _, l := range linters {
			log.Printf("  - %s", l.Name())
		}
	}

	// Run all available linters
	aggregateResult := lint.NewResult()
	changedFiles := s.GetAllFilesMap()

	for _, linter := range linters {
		if *verbose {
			log.Printf("Running %s...", linter.Name())
		}

		// Get files/packages for this linter
		targets := selectTargets(linter, lang, s)

		result, err := linter.Run(ctx, projectDir, targets)
		if err != nil {
			log.Printf("Warning: %s failed: %v", linter.Name(), err)
			aggregateResult.Errors = append(aggregateResult.Errors, fmt.Sprintf("%s: %v", linter.Name(), err))
			continue
		}
		if result == nil {
			log.Printf("Warning: %s returned nil result", linter.Name())
			continue
		}

		// Filter to changed files only and merge
		filtered := result.FilterByFiles(changedFiles)
		aggregateResult.Merge(filtered)

		if *verbose {
			log.Printf("  %s: %d findings", linter.Name(), len(filtered.Findings))
		}
	}

	// Deduplicate findings (same file:line:message from different tools)
	deduplicateFindings(aggregateResult)

	// Ensure output directory exists
	writer := output.NewLintWriter(*outputPath)
	if err := writer.EnsureDir(); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Write results
	if err := writer.WriteResult(aggregateResult); err != nil {
		log.Fatalf("Failed to write results: %v", err)
	}

	// Write language-specific result
	if err := writer.WriteLanguageResult(lang, aggregateResult); err != nil {
		log.Fatalf("Failed to write language result: %v", err)
	}

	// Print summary
	fmt.Printf("Static analysis complete:\n")
	fmt.Printf("  Files analyzed: %d\n", len(changedFiles))
	fmt.Printf("  Critical: %d\n", aggregateResult.Summary.Critical)
	fmt.Printf("  High: %d\n", aggregateResult.Summary.High)
	fmt.Printf("  Warning: %d\n", aggregateResult.Summary.Warning)
	fmt.Printf("  Info: %d\n", aggregateResult.Summary.Info)
	fmt.Printf("  Unknown: %d\n", aggregateResult.Summary.Unknown)
	fmt.Printf("  Output: %s\n", filepath.Join(*outputPath, "static-analysis.json"))

	if len(aggregateResult.Errors) > 0 {
		fmt.Printf("\nWarnings during analysis:\n")
		for _, e := range aggregateResult.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}
}

// selectTargets chooses files/packages for a linter based on its preference.
// Falls back to prior language-based defaults when a linter does not specify.
func selectTargets(linter lint.Linter, lang lint.Language, s *scope.ScopeJSON) []string {
	if selector, ok := linter.(lint.TargetSelector); ok {
		switch selector.TargetKind() {
		case lint.TargetKindPackages:
			if pkgs := s.GetPackages(); len(pkgs) > 0 {
				return pkgs
			}
			return nil // let linter use its default
		case lint.TargetKindFiles:
			if files := s.GetAllFiles(); len(files) > 0 {
				return files
			}
			return nil
		case lint.TargetKindProject:
			return nil
		}
	}

	// Legacy behavior by language as a fallback.
	if lang == lint.LanguageGo {
		if pkgs := s.GetPackages(); len(pkgs) > 0 {
			return pkgs
		}
		return nil
	}

	if files := s.GetAllFiles(); len(files) > 0 {
		return files
	}

	return nil
}

// registerLinters adds all linters to the registry.
func registerLinters(r *lint.Registry) {
	// Go linters
	r.Register(lint.NewGolangciLint())
	r.Register(lint.NewStaticcheck())
	r.Register(lint.NewGosec())

	// TypeScript linters
	r.Register(lint.NewTSC())
	r.Register(lint.NewESLint())

	// Python linters
	r.Register(lint.NewRuff())
	r.Register(lint.NewMypy())
	r.Register(lint.NewPylint())
	r.Register(lint.NewBandit())
}

// deduplicateFindings removes duplicate findings based on file:line:message.
func deduplicateFindings(result *lint.Result) {
	seen := make(map[string]bool)
	unique := make([]lint.Finding, 0)

	// Reset summary
	result.Summary = lint.Summary{}

	for _, f := range result.Findings {
		key := fmt.Sprintf("%s:%d:%s", f.File, f.Line, f.Message)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, f)

			// Update summary
			switch f.Severity {
			case lint.SeverityCritical:
				result.Summary.Critical++
			case lint.SeverityHigh:
				result.Summary.High++
			case lint.SeverityWarning:
				result.Summary.Warning++
			case lint.SeverityInfo:
				result.Summary.Info++
			default:
				result.Summary.Unknown++
				msg := fmt.Sprintf("unknown severity %q for finding %s:%d (%s)", f.Severity, f.File, f.Line, f.Message)
				result.Errors = append(result.Errors, msg)
				log.Printf("Warning: %s", msg)
			}
		}
	}

	result.Findings = unique
}
