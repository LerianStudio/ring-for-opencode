// Package main implements the data-flow binary for security-focused data flow analysis.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/dataflow"
)

// ScopeFile represents the scope.json structure from Phase 0.
// This is a simplified representation that supports both flat file lists
// and language-categorized files.
type ScopeFile struct {
	Files     []string            `json:"-"` // Populated after parsing
	Languages map[string][]string `json:"languages"`
}

// filesNested represents the nested files structure in scope.json.
type filesNested struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
}

const (
	// MaxFiles is the maximum number of files allowed in scope.json to prevent resource exhaustion.
	MaxFiles = 10000
)

var (
	scopePath = flag.String("scope", "scope.json", "Path to scope.json from Phase 0")
	outputDir = flag.String("output", ".", "Output directory for results")
	scriptDir = flag.String("scripts", "", "Path to scripts/codereview directory (auto-detect from executable path if not provided)")
	language  = flag.String("lang", "", "Analyze specific language only (go, python, typescript)")
	jsonOnly  = flag.Bool("json", false, "Output JSON only, no markdown summary")
	verbose   = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Data Flow Analyzer - Security-focused data flow analysis\n\n")
		fmt.Fprintf(os.Stderr, "Analyzes source code to detect:\n")
		fmt.Fprintf(os.Stderr, "  - Untrusted data sources (HTTP inputs, env vars, files)\n")
		fmt.Fprintf(os.Stderr, "  - Sensitive data sinks (database, exec, response)\n")
		fmt.Fprintf(os.Stderr, "  - Unsanitized data flows between sources and sinks\n")
		fmt.Fprintf(os.Stderr, "  - Nil/null safety issues\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Analyze all languages from scope.json:\n")
		fmt.Fprintf(os.Stderr, "  %s -scope .ring/codereview/scope.json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Analyze only Go files:\n")
		fmt.Fprintf(os.Stderr, "  %s -scope scope.json -lang go\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Output JSON only (no markdown):\n")
		fmt.Fprintf(os.Stderr, "  %s -scope scope.json -json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Supported languages: go, python, typescript\n")
	}
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get working directory early - needed for path validation
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Auto-detect script directory from executable path if not provided
	scriptsDir := *scriptDir
	if scriptsDir == "" {
		execPath, err := os.Executable()
		if err == nil {
			// Executable is typically at scripts/codereview/bin/data-flow
			// We want scripts/codereview/
			scriptsDir = filepath.Dir(filepath.Dir(execPath))
		}
		// Fallback to current directory if detection fails
		if scriptsDir == "" {
			scriptsDir = "."
		}
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Scope file: %s\n", *scopePath)
		fmt.Fprintf(os.Stderr, "Output directory: %s\n", *outputDir)
		fmt.Fprintf(os.Stderr, "Scripts directory: %s\n", scriptsDir)
		if *language != "" {
			fmt.Fprintf(os.Stderr, "Language filter: %s\n", *language)
		}
	}

	// Load scope.json with path validation
	scope, err := loadScope(*scopePath, workDir)
	if err != nil {
		return fmt.Errorf("failed to load scope: %w", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Loaded %d files from scope\n", len(scope.Files))
		for lang, files := range scope.Languages {
			fmt.Fprintf(os.Stderr, "  %s: %d files\n", lang, len(files))
		}
	}

	// Determine which languages to analyze
	langsToAnalyze := []string{"go", "python", "typescript"}
	if *language != "" {
		lang := normalizeLanguage(*language)
		if lang == "" {
			return fmt.Errorf("unsupported language: %s (supported: go, python, typescript)", *language)
		}
		langsToAnalyze = []string{lang}
	}

	// Initialize analyzers and run analysis
	results := make(map[string]*dataflow.FlowAnalysis)

	for _, lang := range langsToAnalyze {
		// Get files for this language
		files := getFilesForLanguage(scope, lang)
		if len(files) == 0 {
			if *verbose {
				fmt.Fprintf(os.Stderr, "No %s files to analyze, skipping\n", lang)
			}
			continue
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "Analyzing %d %s files...\n", len(files), lang)
		}

		// Create appropriate analyzer
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

		// Run analysis
		analysis, err := analyzer.Analyze(files)
		if err != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Warning: %s analysis failed: %v\n", lang, err)
			}
			continue
		}

		results[lang] = analysis

		// Write language-specific JSON output
		outputPath := filepath.Join(*outputDir, fmt.Sprintf("%s-flow.json", lang))
		if err := writeJSON(outputPath, analysis); err != nil {
			return fmt.Errorf("failed to write %s results: %w", lang, err)
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "Wrote %s analysis to %s\n", lang, outputPath)
		}
	}

	// Handle output mode
	if *jsonOnly {
		// Print all results as indented JSON
		output := struct {
			Languages map[string]*dataflow.FlowAnalysis `json:"languages"`
		}{
			Languages: results,
		}
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal results: %w", err)
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Generate security summary markdown
		summary := dataflow.GenerateSecuritySummary(results)
		summaryPath := filepath.Join(*outputDir, "security-summary.md")
		if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
			return fmt.Errorf("failed to write security summary: %w", err)
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "Wrote security summary to %s\n", summaryPath)
		}

		// Print summary stats
		printSummary(results)
	}

	return nil
}

// validateFilePath ensures a file path is within the working directory to prevent path traversal.
func validateFilePath(basePath, filePath string) (string, error) {
	// Get absolute paths
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("invalid base path: %w", err)
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("invalid file path: %w", err)
	}

	// Clean paths to resolve .. and symlinks
	cleanBase := filepath.Clean(absBase)
	cleanPath := filepath.Clean(absPath)

	// Ensure path is within base directory
	if !strings.HasPrefix(cleanPath, cleanBase) {
		return "", fmt.Errorf("path traversal detected: %s is outside %s", filePath, basePath)
	}

	return cleanPath, nil
}

// loadScope reads and parses scope.json, populating the Languages map if empty.
// workDir is used to validate that all file paths are within the working directory.
func loadScope(path string, workDir string) (*ScopeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading scope file: %w", err)
	}

	scope := &ScopeFile{
		Languages: make(map[string][]string),
	}

	// Parse into generic map to handle polymorphic "files" field
	var rawScope map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawScope); err != nil {
		return nil, fmt.Errorf("parsing scope file: %w", err)
	}

	// Extract files - handle both flat array and nested object formats
	var rawFiles []string
	if filesRaw, ok := rawScope["files"]; ok {
		// Try flat array first
		var flatFiles []string
		if err := json.Unmarshal(filesRaw, &flatFiles); err == nil {
			rawFiles = flatFiles
		} else {
			// Try nested object format
			var nestedFiles filesNested
			if err := json.Unmarshal(filesRaw, &nestedFiles); err == nil {
				rawFiles = append(rawFiles, nestedFiles.Modified...)
				rawFiles = append(rawFiles, nestedFiles.Added...)
			}
		}
	}

	// Check file count limit before processing
	if len(rawFiles) > MaxFiles {
		return nil, fmt.Errorf("too many files: %d (max %d)", len(rawFiles), MaxFiles)
	}

	// Validate all file paths to prevent path traversal
	validatedFiles := make([]string, 0, len(rawFiles))
	for _, f := range rawFiles {
		validPath, err := validateFilePath(workDir, f)
		if err != nil {
			// Log warning but continue (don't fail on single bad path)
			fmt.Fprintf(os.Stderr, "Warning: skipping invalid path: %v\n", err)
			continue
		}
		validatedFiles = append(validatedFiles, validPath)
	}
	scope.Files = validatedFiles

	// Extract languages if present
	if langRaw, ok := rawScope["languages"]; ok {
		if err := json.Unmarshal(langRaw, &scope.Languages); err != nil {
			// Ignore language parsing errors, will populate from files
			scope.Languages = make(map[string][]string)
		}
	}

	// Validate language file paths as well
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

	// Populate Languages map from Files if empty
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

// detectLanguage returns the language for a file based on extension.
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

// normalizeLanguage normalizes language names to canonical form.
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

// getFilesForLanguage returns files for a specific language from the scope.
func getFilesForLanguage(scope *ScopeFile, lang string) []string {
	// First, check the Languages map
	if files, ok := scope.Languages[lang]; ok && len(files) > 0 {
		return files
	}

	// Fall back to filtering Files by extension
	return filterFilesByLanguage(scope.Files, lang)
}

// filterFilesByLanguage filters files by language extension.
func filterFilesByLanguage(files []string, lang string) []string {
	var filtered []string
	for _, file := range files {
		if detectLanguage(file) == lang {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// writeJSON writes data to a JSON file with indentation.
func writeJSON(path string, data interface{}) error {
	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// printSummary prints analysis summary and risk assessment to stdout.
func printSummary(results map[string]*dataflow.FlowAnalysis) {
	// Aggregate statistics
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

	// Risk assessment
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
	fmt.Printf("  Output: %s/{lang}-flow.json\n", *outputDir)
	fmt.Printf("  Summary: %s/security-summary.md\n", *outputDir)
}
