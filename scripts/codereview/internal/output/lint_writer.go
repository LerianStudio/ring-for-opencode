// Package output handles writing analysis results to files.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerianstudio/ring/scripts/codereview/internal/lint"
)

// LintWriter handles writing lint analysis results.
type LintWriter struct {
	outputDir string
}

// NewLintWriter creates a new lint output writer.
func NewLintWriter(outputDir string) *LintWriter {
	return &LintWriter{
		outputDir: outputDir,
	}
}

// EnsureDir creates the output directory if it doesn't exist.
func (w *LintWriter) EnsureDir() error {
	return os.MkdirAll(w.outputDir, 0o700)
}

// WriteResult writes the analysis result to static-analysis.json.
func (w *LintWriter) WriteResult(result *lint.Result) error {
	return w.writeJSON("static-analysis.json", result)
}

// WriteLanguageResult writes a language-specific result file.
func (w *LintWriter) WriteLanguageResult(lang lint.Language, result *lint.Result) error {
	filename := fmt.Sprintf("%s-lint.json", lang)
	return w.writeJSON(filename, result)
}

// writeJSON writes data as formatted JSON to a file.
func (w *LintWriter) writeJSON(filename string, data interface{}) error {
	path := filepath.Join(w.outputDir, filename)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	dataWithNewline := append(jsonData, '\n')

	if err := os.WriteFile(path, dataWithNewline, 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

// DefaultOutputDir returns the default output directory.
func DefaultOutputDir(projectDir string) string {
	return filepath.Join(projectDir, ".ring", "codereview")
}
