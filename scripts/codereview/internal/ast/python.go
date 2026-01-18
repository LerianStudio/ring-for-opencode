package ast

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// PythonExtractor implements AST extraction for Python files
type PythonExtractor struct {
	pythonExecutable string
	scriptPath       string
}

// NewPythonExtractor creates a new Python AST extractor
func NewPythonExtractor(scriptDir string) *PythonExtractor {
	return &PythonExtractor{
		pythonExecutable: "python3",
		scriptPath:       filepath.Join(scriptDir, "py", "ast_extractor.py"),
	}
}

func (p *PythonExtractor) Language() string {
	return "python"
}

func (p *PythonExtractor) SupportedExtensions() []string {
	return []string{".py", ".pyi"}
}

func (p *PythonExtractor) ExtractDiff(ctx context.Context, beforePath, afterPath string) (*SemanticDiff, error) {
	before := beforePath
	if before == "" {
		before = `""`
	}
	after := afterPath
	if after == "" {
		after = `""`
	}

	cmd := exec.CommandContext(ctx, p.pythonExecutable, p.scriptPath, "--before", before, "--after", after) // #nosec G204 - args are controlled
	cmd.Env = append([]string{"LC_ALL=C"}, os.Environ()...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("python extractor failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run python extractor: %w", err)
	}

	var diff SemanticDiff
	if err := json.Unmarshal(output, &diff); err != nil {
		return nil, fmt.Errorf("failed to parse python extractor output: %w", err)
	}

	return &diff, nil
}
