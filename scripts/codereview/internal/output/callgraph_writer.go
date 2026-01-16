// Package output handles writing analysis results to files.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerianstudio/ring/scripts/codereview/internal/callgraph"
)

// CallGraphWriter handles writing call graph analysis results.
type CallGraphWriter struct {
	outputDir string
}

// NewCallGraphWriter creates a new call graph output writer.
func NewCallGraphWriter(outputDir string) *CallGraphWriter {
	return &CallGraphWriter{
		outputDir: outputDir,
	}
}

// EnsureDir creates the output directory if it doesn't exist.
func (w *CallGraphWriter) EnsureDir() error {
	return os.MkdirAll(w.outputDir, 0o755)
}

// WriteJSON writes the call graph result as a JSON file.
// Creates {language}-calls.json in the output directory.
func WriteJSON(result *callgraph.CallGraphResult, outputDir string) error {
	if result == nil {
		return fmt.Errorf("cannot write nil CallGraphResult to JSON")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Generate filename based on language
	filename := fmt.Sprintf("%s-calls.json", result.Language)
	path := filepath.Join(outputDir, filename)

	// Marshal with pretty printing
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal call graph result: %w", err)
	}

	// Write to file with trailing newline for better file handling.
	// Avoids an extra allocation vs append(jsonData, '\n').
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}

	if n, err := f.Write(jsonData); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to write %s: %w", path, err)
	} else if n != len(jsonData) {
		_ = f.Close()
		return fmt.Errorf("failed to write %s: short write (%d/%d)", path, n, len(jsonData))
	}

	if n, err := f.Write([]byte{'\n'}); err != nil {
		_ = f.Close()
		return fmt.Errorf("failed to write %s newline: %w", path, err)
	} else if n != 1 {
		_ = f.Close()
		return fmt.Errorf("failed to write %s newline: short write (%d/1)", path, n)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close %s: %w", path, err)
	}

	return nil
}

// WriteImpactSummary writes the impact summary as a Markdown file.
// Creates impact-summary.md in the output directory.
func WriteImpactSummary(result *callgraph.CallGraphResult, outputDir string) error {
	if result == nil {
		return fmt.Errorf("cannot write nil CallGraphResult to impact summary")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Generate markdown content
	markdown := RenderImpactSummary(result)

	// Write to file
	path := filepath.Join(outputDir, "impact-summary.md")
	if err := os.WriteFile(path, []byte(markdown), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	return nil
}

// WriteResult writes the call graph result to a JSON file using the writer's output directory.
func (w *CallGraphWriter) WriteResult(result *callgraph.CallGraphResult) error {
	return WriteJSON(result, w.outputDir)
}

// WriteSummary writes the impact summary markdown using the writer's output directory.
func (w *CallGraphWriter) WriteSummary(result *callgraph.CallGraphResult) error {
	return WriteImpactSummary(result, w.outputDir)
}

// WriteAll writes both JSON and Markdown output files.
//
// Note: directory creation is handled by the package-level helpers (WriteJSON,
// WriteImpactSummary) which call os.MkdirAll, so WriteAll does not redundantly
// ensure the directory itself.
func (w *CallGraphWriter) WriteAll(result *callgraph.CallGraphResult) error {
	if err := w.WriteResult(result); err != nil {
		return err
	}

	return w.WriteSummary(result)
}
