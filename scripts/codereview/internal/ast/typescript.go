package ast

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
		scriptPath:     findTypeScriptASTExtractor(scriptDir),
	}
}

func findTypeScriptASTExtractor(scriptDir string) string {
	if scriptDir == "" {
		return filepath.Join("ts", "dist", "ast-extractor.js")
	}

	candidates := []string{
		filepath.Join(scriptDir, "ts", "dist", "ast-extractor.js"),
		filepath.Join(scriptDir, "ts", "ast-extractor.ts"),
		filepath.Join(scriptDir, "ts", "ast-extractor.js"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return filepath.Join(scriptDir, "ts", "dist", "ast-extractor.js")
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

	args := []string{t.scriptPath, "--before", before, "--after", after}
	cmd := exec.CommandContext(ctx, t.nodeExecutable, args...) // #nosec G204 - args are controlled
	cmd.Env = append([]string{"LC_ALL=C"}, os.Environ()...)
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
