// Package dataflow implements Phase 4: data flow analysis for the unified CLI.
package dataflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/ast"
	"github.com/lerianstudio/ring/scripts/codereview/internal/dataflow"
	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
)

// Phase implements the data flow analysis phase.
type Phase struct{}

// New creates a new data flow analysis phase.
func New() *Phase {
	return &Phase{}
}

// Name returns the canonical name of this phase.
func (p *Phase) Name() string {
	return "dataflow"
}

// Timeout returns the default timeout for this phase.
func (p *Phase) Timeout() time.Duration {
	return 3 * time.Minute
}

// ShouldSkip checks if this phase should be skipped.
func (p *Phase) ShouldSkip(cfg *phases.Config) (bool, string) {
	scopePath := cfg.ScopePath
	if scopePath == "" {
		scopePath = filepath.Join(cfg.OutputDir, "scope.json")
	}
	if _, err := os.Stat(scopePath); os.IsNotExist(err) {
		return true, "scope.json missing - skipping dataflow phase"
	}
	return false, ""
}

// MaxFiles is the maximum number of files allowed in scope.json.
const MaxFiles = 10000

// ScopeFile represents the scope.json structure.
type ScopeFile struct {
	Files     []string
	Languages map[string][]string
}

// filesNested represents the nested files structure in scope.json.
type filesNested struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
}

// Run executes the data flow analysis phase.
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

	scriptsDir := cfg.ScriptsDir
	if scriptsDir == "" {
		execPath, err := os.Executable()
		if err == nil {
			scriptsDir = filepath.Dir(filepath.Dir(execPath))
		}
		if scriptsDir == "" {
			scriptsDir = "."
		}
	}

	scopePath := cfg.ScopePath
	if scopePath == "" {
		scopePath = filepath.Join(cfg.OutputDir, "scope.json")
	}

	if cfg.Verbose {
		logger.Debug("configuration", "scope", scopePath, "output", cfg.OutputDir, "scripts", scriptsDir)
	}

	scope, err := loadScope(scopePath, workDir)
	if err != nil {
		return fmt.Errorf("failed to load scope: %w", err)
	}

	if cfg.Verbose {
		logger.Debug("loaded scope", "files", len(scope.Files), "languages", scope.Languages)
	}

	// Determine which languages to analyze
	langsToAnalyze := []string{"go", "python", "typescript"}
	if cfg.Language != "" {
		lang := normalizeLanguage(cfg.Language)
		if lang == "" {
			return fmt.Errorf("unsupported language: %s (supported: go, python, typescript)", cfg.Language)
		}
		langsToAnalyze = []string{lang}
	}

	results := make(map[string]*dataflow.FlowAnalysis)

	for _, lang := range langsToAnalyze {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		files := getFilesForLanguage(scope, lang)
		if len(files) == 0 {
			if cfg.Verbose {
				logger.Debug("no files to analyze", "language", lang)
			}
			continue
		}

		if cfg.Verbose {
			logger.Debug("analyzing files", "language", lang, "count", len(files))
		}

		var analyzer dataflow.Analyzer
		switch lang {
		case "go":
			analyzer = dataflow.NewGoAnalyzer(workDir)
		case "python":
			analyzer = dataflow.NewPythonAnalyzer(scriptsDir)
		case "typescript":
			analyzer = dataflow.NewTypeScriptAnalyzer(scriptsDir)
		default:
			continue
		}

		analysis, err := analyzer.Analyze(files)
		if err != nil {
			if cfg.Verbose {
				logger.Warn("analysis failed", "language", lang, "error", err)
			}
			continue
		}

		results[lang] = analysis

		outputPath := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s-flow.json", lang))
		if err := writeJSON(outputPath, analysis); err != nil {
			return fmt.Errorf("failed to write %s results: %w", lang, err)
		}

		if cfg.Verbose {
			logger.Debug("wrote analysis", "language", lang, "path", outputPath)
		}
	}

	// Generate security summary markdown
	summary := dataflow.GenerateSecuritySummary(results)
	summaryPath := filepath.Join(cfg.OutputDir, "security-summary.md")
	if err := os.WriteFile(summaryPath, []byte(summary), 0o600); err != nil {
		return fmt.Errorf("failed to write security summary: %w", err)
	}

	if cfg.Verbose {
		logger.Debug("wrote security summary", "path", summaryPath)
	}

	printSummary(results, cfg.OutputDir)
	return nil
}

func validateFilePath(basePath, filePath string) (string, error) {
	validator, err := ast.NewPathValidator(basePath)
	if err != nil {
		return "", fmt.Errorf("invalid base path: %w", err)
	}
	validated, err := validator.ValidatePath(filePath)
	if err != nil {
		return "", fmt.Errorf("path traversal detected: %w", err)
	}
	return validated, nil
}

func loadScope(path string, workDir string) (*ScopeFile, error) {
	data, err := fileutil.ReadJSONFileWithLimit(path)
	if err != nil {
		return nil, fmt.Errorf("reading scope file: %w", err)
	}

	scope := &ScopeFile{
		Languages: make(map[string][]string),
	}

	var rawScope map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawScope); err != nil {
		return nil, fmt.Errorf("parsing scope file: %w", err)
	}

	var rawFiles []string
	if filesRaw, ok := rawScope["files"]; ok {
		var flatFiles []string
		if err := json.Unmarshal(filesRaw, &flatFiles); err == nil {
			rawFiles = flatFiles
		} else {
			var nestedFiles filesNested
			if err := json.Unmarshal(filesRaw, &nestedFiles); err == nil {
				rawFiles = append(rawFiles, nestedFiles.Modified...)
				rawFiles = append(rawFiles, nestedFiles.Added...)
			}
		}
	}
	if rawFiles == nil {
		rawFiles = []string{}
	}

	if len(rawFiles) > MaxFiles {
		return nil, fmt.Errorf("too many files: %d (max %d)", len(rawFiles), MaxFiles)
	}

	validatedFiles := make([]string, 0, len(rawFiles))
	for _, f := range rawFiles {
		validPath, err := validateFilePath(workDir, f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping invalid path: %v\n", err)
			continue
		}
		validatedFiles = append(validatedFiles, validPath)
	}
	scope.Files = validatedFiles

	if langRaw, ok := rawScope["languages"]; ok {
		if err := json.Unmarshal(langRaw, &scope.Languages); err != nil {
			scope.Languages = make(map[string][]string)
		}
	}
	if scope.Languages == nil {
		scope.Languages = make(map[string][]string)
	}

	for lang, files := range scope.Languages {
		if len(files) > MaxFiles {
			return nil, fmt.Errorf("too many files for language %s: %d (max %d)", lang, len(files), MaxFiles)
		}
		validatedLangFiles := make([]string, 0, len(files))
		for _, f := range files {
			validPath, err := validateFilePath(workDir, f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: skipping invalid path in %s: %v\n", lang, err)
				continue
			}
			validatedLangFiles = append(validatedLangFiles, validPath)
		}
		scope.Languages[lang] = validatedLangFiles
	}

	if len(scope.Languages) == 0 && len(scope.Files) > 0 {
		for _, file := range scope.Files {
			lang := detectLanguage(file)
			if lang != "" {
				scope.Languages[lang] = append(scope.Languages[lang], file)
			}
		}
	}

	return scope, nil
}

func detectLanguage(file string) string {
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".ts", ".tsx", ".js", ".jsx":
		return "typescript"
	default:
		return ""
	}
}

func normalizeLanguage(lang string) string {
	switch strings.ToLower(lang) {
	case "go", "golang":
		return "go"
	case "python", "py":
		return "python"
	case "typescript", "ts", "javascript", "js":
		return "typescript"
	default:
		return ""
	}
}

func getFilesForLanguage(scope *ScopeFile, lang string) []string {
	if files, ok := scope.Languages[lang]; ok && len(files) > 0 {
		return files
	}
	return filterFilesByLanguage(scope.Files, lang)
}

func filterFilesByLanguage(files []string, lang string) []string {
	var filtered []string
	for _, file := range files {
		if detectLanguage(file) == lang {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func writeJSON(path string, data interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonBytes, 0o600); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

func printSummary(results map[string]*dataflow.FlowAnalysis, outputDir string) {
	var totalStats struct {
		sources          int
		sinks            int
		flows            int
		unsanitizedFlows int
		criticalFlows    int
		highRiskFlows    int
		nilRisks         int
	}

	languages := make([]string, 0, len(results))
	for lang, analysis := range results {
		if analysis == nil {
			continue
		}
		languages = append(languages, lang)
		totalStats.sources += analysis.Statistics.TotalSources
		totalStats.sinks += analysis.Statistics.TotalSinks
		totalStats.flows += analysis.Statistics.TotalFlows
		totalStats.unsanitizedFlows += analysis.Statistics.UnsanitizedFlows
		totalStats.criticalFlows += analysis.Statistics.CriticalFlows
		totalStats.highRiskFlows += analysis.Statistics.HighRiskFlows
		totalStats.nilRisks += analysis.Statistics.NilRisks
	}

	fmt.Printf("Data flow analysis complete:\n")
	fmt.Printf("  Languages analyzed: %d (%s)\n", len(languages), strings.Join(languages, ", "))
	fmt.Printf("  Total sources: %d\n", totalStats.sources)
	fmt.Printf("  Total sinks: %d\n", totalStats.sinks)
	fmt.Printf("  Total flows: %d\n", totalStats.flows)
	fmt.Printf("  Unsanitized flows: %d\n", totalStats.unsanitizedFlows)
	fmt.Printf("  Nil/null risks: %d\n", totalStats.nilRisks)
	fmt.Println()

	if totalStats.criticalFlows > 0 {
		fmt.Printf("  CRITICAL: %d critical-risk flows detected!\n", totalStats.criticalFlows)
	}
	if totalStats.highRiskFlows > 0 {
		fmt.Printf("  WARNING: %d high-risk flows detected\n", totalStats.highRiskFlows)
	}
	if totalStats.criticalFlows == 0 && totalStats.highRiskFlows == 0 {
		fmt.Printf("  No critical or high-risk flows detected\n")
	}

	fmt.Println()
	fmt.Printf("  Output: %s/{lang}-flow.json\n", outputDir)
	fmt.Printf("  Summary: %s/security-summary.md\n", outputDir)
}
