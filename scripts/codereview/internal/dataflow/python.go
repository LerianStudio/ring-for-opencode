// Package dataflow provides data flow analysis for security review.
package dataflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PythonAnalyzer implements data flow analysis for Python and TypeScript
// by delegating to the py/data_flow.py script.
type PythonAnalyzer struct {
	scriptPath string
	language   string
}

// NewPythonAnalyzer creates a new analyzer for Python files.
func NewPythonAnalyzer(scriptDir string) *PythonAnalyzer {
	return &PythonAnalyzer{
		scriptPath: filepath.Join(scriptDir, "py", "data_flow.py"),
		language:   "python",
	}

}

// NewTypeScriptAnalyzer creates a new analyzer for TypeScript/JavaScript files.
func NewTypeScriptAnalyzer(scriptDir string) *PythonAnalyzer {
	return &PythonAnalyzer{
		scriptPath: filepath.Join(scriptDir, "py", "data_flow.py"),
		language:   "typescript",
	}

}

// Language returns the analyzer's target language.
func (p *PythonAnalyzer) Language() string {
	return p.language
}

// filterFiles filters the given files to only include those matching the analyzer's language.
func (p *PythonAnalyzer) filterFiles(files []string) []string {
	var filtered []string

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		switch p.language {
		case "python":
			if ext == ".py" {
				filtered = append(filtered, file)
			}
		case "typescript":
			if ext == ".ts" || ext == ".tsx" || ext == ".js" || ext == ".jsx" {
				filtered = append(filtered, file)
			}
		}
	}

	return filtered
}

// runScript executes the Python data flow analysis script and parses its output.
func (p *PythonAnalyzer) runScript(files []string) (*FlowAnalysis, error) {
	filteredFiles := p.filterFiles(files)
	if len(filteredFiles) == 0 {
		return &FlowAnalysis{
			Language:   p.language,
			Sources:    []Source{},
			Sinks:      []Sink{},
			Flows:      []Flow{},
			NilSources: []NilSource{},
			Statistics: Stats{},
		}, nil
	}

	// Build command arguments: python3 script.py <language> <files...>
	args := []string{p.scriptPath, p.language}
	args = append(args, filteredFiles...)

	cmd := exec.Command("python3", args...) // #nosec G204 - args are controlled
	cmd.Env = append([]string{"LC_ALL=C"}, os.Environ()...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("script execution failed: %w: %s", err, stderrStr)
		}
		return nil, fmt.Errorf("script execution failed: %w", err)
	}

	var analysis FlowAnalysis
	if err := json.Unmarshal(stdout.Bytes(), &analysis); err != nil {
		return nil, fmt.Errorf("parsing script output: %w", err)
	}

	return &analysis, nil
}

// DetectSources scans files for untrusted data sources.
func (p *PythonAnalyzer) DetectSources(files []string) ([]Source, error) {
	analysis, err := p.runScript(files)
	if err != nil {
		return nil, fmt.Errorf("detecting sources: %w", err)
	}
	return analysis.Sources, nil
}

// DetectSinks scans files for sensitive data sinks.
func (p *PythonAnalyzer) DetectSinks(files []string) ([]Sink, error) {
	analysis, err := p.runScript(files)
	if err != nil {
		return nil, fmt.Errorf("detecting sinks: %w", err)
	}
	return analysis.Sinks, nil
}

// TrackFlows traces data paths from sources to sinks.
// Note: This method re-runs the full analysis since the Python script
// performs integrated analysis. The sources and sinks parameters are
// ignored as the script determines them internally.
func (p *PythonAnalyzer) TrackFlows(sources []Source, sinks []Sink, files []string) ([]Flow, error) {
	analysis, err := p.runScript(files)
	if err != nil {
		return nil, fmt.Errorf("tracking flows: %w", err)
	}
	return analysis.Flows, nil
}

// DetectNilSources identifies variables that may be nil/null/undefined.
func (p *PythonAnalyzer) DetectNilSources(files []string) ([]NilSource, error) {
	analysis, err := p.runScript(files)
	if err != nil {
		return nil, fmt.Errorf("detecting nil sources: %w", err)
	}
	return analysis.NilSources, nil
}

// Analyze performs complete data flow analysis on the given files.
func (p *PythonAnalyzer) Analyze(files []string) (*FlowAnalysis, error) {
	return p.runScript(files)
}
