package ast

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

// TypeScriptExtractor implements AST extraction for TypeScript files
type TypeScriptExtractor struct {
	nodeExecutable string
	scriptPath     string
}

// NewTypeScriptExtractor creates a new TypeScript AST extractor
func NewTypeScriptExtractor(scriptDir string) *TypeScriptExtractor {
	return &TypeScriptExtractor{
		nodeExecutable: "node",
		scriptPath:     filepath.Join(scriptDir, "ts", "dist", "ast-extractor.js"),
	}
}

func (t *TypeScriptExtractor) Language() string {
	return "typescript"
}

func (t *TypeScriptExtractor) SupportedExtensions() []string {
	return []string{".ts", ".tsx", ".js", ".jsx"}
}

func (t *TypeScriptExtractor) ExtractDiff(ctx context.Context, beforePath, afterPath string) (*SemanticDiff, error) {
	before := beforePath
	if before == "" {
		before = `""`
	}
	after := afterPath
	if after == "" {
		after = `""`
	}

	cmd := exec.CommandContext(ctx, t.nodeExecutable, t.scriptPath, "--before", before, "--after", after)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("typescript extractor failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run typescript extractor: %w", err)
	}

	var diff SemanticDiff
	if err := json.Unmarshal(output, &diff); err != nil {
		return nil, fmt.Errorf("failed to parse typescript extractor output: %w", err)
	}

	return &diff, nil
}
