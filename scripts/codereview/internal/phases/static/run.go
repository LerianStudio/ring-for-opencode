// Package static implements Phase 1: static analysis for the unified CLI.
package static

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/lint"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/output"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/scope"
)

// Phase implements the static analysis phase.
type Phase struct{}

// New creates a new static analysis phase.
func New() *Phase {
	return &Phase{}
}

// Name returns the canonical name of this phase.
func (p *Phase) Name() string {
	return "static"
}

// Timeout returns the default timeout for this phase.
func (p *Phase) Timeout() time.Duration {
	return 5 * time.Minute
}

// Run executes the static analysis phase.
func (p *Phase) Run(ctx context.Context, cfg *phases.Config) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	projectDir := cfg.WorkDir
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	scopePath := cfg.ScopePath
	if scopePath == "" {
		scopePath = scope.DefaultScopePath(cfg.OutputDir)
	}
	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = output.DefaultOutputDir(projectDir)
	}

	if cfg.Verbose {
		logger.Debug("reading scope", "path", scopePath)
	}

	s, err := scope.ReadScopeJSON(scopePath)
	if err != nil {
		return fmt.Errorf("failed to read scope: %w", err)
	}

	lang := s.GetLanguage()
	if cfg.Verbose {
		logger.Debug("detected language", "language", lang)
	}

	registry := lint.NewRegistry()
	registerLinters(registry)

	linters := selectAvailableLinters(ctx, registry, lang, s)
	if len(linters) == 0 {
		logger.Warn("no linters available", "language", lang)
	}

	if cfg.Verbose {
		logger.Debug("available linters", "count", len(linters))
		for _, l := range linters {
			logger.Debug("linter", "name", l.Name())
		}
	}

	aggregateResult := lint.NewResult()
	changedFiles := s.GetAllFilesMap()

	for _, linter := range linters {
		// Check for context cancellation between linters
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if cfg.Verbose {
			logger.Debug("running linter", "name", linter.Name())
		}

		targets := selectTargets(linter, lang, s)

		result, err := linter.Run(ctx, projectDir, targets)
		if err != nil {
			logger.Warn("linter failed", "name", linter.Name(), "error", err)
			aggregateResult.Errors = append(aggregateResult.Errors, fmt.Sprintf("%s: %v", linter.Name(), err))
			continue
		}
		if result == nil {
			logger.Warn("linter returned nil result", "name", linter.Name())
			continue
		}

		filtered := result.FilterByFiles(changedFiles)
		aggregateResult.Merge(filtered)

		if cfg.Verbose {
			logger.Debug("linter findings", "name", linter.Name(), "count", len(filtered.Findings))
		}
	}

	deduplicateFindings(aggregateResult)

	writer := output.NewLintWriter(outputDir)
	if err := writer.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := writer.WriteResult(aggregateResult); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	if err := writer.WriteLanguageResult(lang, aggregateResult); err != nil {
		return fmt.Errorf("failed to write language result: %w", err)
	}

	if lang == lint.LanguageMixed {
		languages := s.Languages
		if len(languages) == 0 {
			languages = []string{"go", "typescript", "python"}
		}
		for _, language := range languages {
			normalized := scope.NormalizeLanguage(language)
			if normalized == "" {
				continue
			}
			if err := writer.WriteLanguageResult(normalized, aggregateResult); err != nil {
				logger.Warn("failed to write language result", "language", normalized, "error", err)
			}
		}
	}

	fmt.Printf("Static analysis complete:\n")
	fmt.Printf("  Files analyzed: %d\n", len(changedFiles))
	fmt.Printf("  Critical: %d\n", aggregateResult.Summary.Critical)
	fmt.Printf("  High: %d\n", aggregateResult.Summary.High)
	fmt.Printf("  Warning: %d\n", aggregateResult.Summary.Warning)
	fmt.Printf("  Info: %d\n", aggregateResult.Summary.Info)
	fmt.Printf("  Unknown: %d\n", aggregateResult.Summary.Unknown)
	fmt.Printf("  Output: %s\n", filepath.Join(outputDir, "static-analysis.json"))

	if len(aggregateResult.Errors) > 0 {
		fmt.Printf("\nWarnings during analysis:\n")
		for _, e := range aggregateResult.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	return nil
}

// selectTargets chooses files/packages for a linter based on its preference.
func selectTargets(linter lint.Linter, lang lint.Language, s *scope.ScopeJSON) []string {
	if selector, ok := linter.(lint.TargetSelector); ok {
		switch selector.TargetKind() {
		case lint.TargetKindPackages:
			if pkgs := s.GetPackages(); len(pkgs) > 0 {
				return pkgs
			}
			return nil
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

// selectAvailableLinters chooses available linters based on scope metadata.
func selectAvailableLinters(ctx context.Context, registry *lint.Registry, lang lint.Language, s *scope.ScopeJSON) []lint.Linter {
	if lang != lint.LanguageMixed {
		return registry.GetAvailableLinters(ctx, lang)
	}

	languageSet := s.Languages
	if len(languageSet) == 0 {
		return registry.GetAvailableLinters(ctx, lang)
	}

	var linters []lint.Linter
	seen := make(map[string]bool)
	for _, language := range languageSet {
		normalized := scope.NormalizeLanguage(language)
		if normalized == "" {
			continue
		}
		for _, linter := range registry.GetAvailableLinters(ctx, normalized) {
			name := linter.Name()
			if !seen[name] {
				linters = append(linters, linter)
				seen[name] = true
			}
		}
	}

	return linters
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
				logger.Warn("unknown severity", "severity", f.Severity, "file", f.File, "line", f.Line)
			}
		}
	}

	result.Findings = unique
}
