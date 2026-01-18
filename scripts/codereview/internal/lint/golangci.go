package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// golangciLintOutput represents golangci-lint JSON output.
type golangciLintOutput struct {
	Issues []golangciIssue `json:"Issues"`
}

type golangciIssue struct {
	FromLinter  string           `json:"FromLinter"`
	Text        string           `json:"Text"`
	Severity    string           `json:"Severity"`
	SourceLines []string         `json:"SourceLines"`
	Pos         golangciPosition `json:"Pos"`
}

type golangciPosition struct {
	Filename string `json:"Filename"`
	Line     int    `json:"Line"`
	Column   int    `json:"Column"`
}

// GolangciLint implements the golangci-lint wrapper.
type GolangciLint struct {
	executor *Executor
}

// NewGolangciLint creates a new golangci-lint wrapper.
func NewGolangciLint() *GolangciLint {
	return &GolangciLint{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (g *GolangciLint) Name() string {
	return "golangci-lint"
}

// Language returns the supported language.
func (g *GolangciLint) Language() Language {
	return LanguageGo
}

// TargetKind declares golangci-lint wants package import paths.
func (g *GolangciLint) TargetKind() TargetKind {
	return TargetKindPackages
}

// Available checks if golangci-lint is installed.
func (g *GolangciLint) Available(ctx context.Context) bool {
	return g.executor.CommandAvailable(ctx, "golangci-lint")
}

// Version returns the golangci-lint version.
func (g *GolangciLint) Version(ctx context.Context) (string, error) {
	version, err := g.executor.GetVersion(ctx, "golangci-lint", "version", "--format", "short")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(version, "v"), nil
}

// Run executes golangci-lint on the specified packages.
func (g *GolangciLint) Run(ctx context.Context, projectDir string, packages []string) (*Result, error) {
	result := NewResult()

	version, err := g.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("golangci-lint version check failed: %v", err))
	} else {
		result.ToolVersions["golangci-lint"] = version
	}

	// Build arguments
	args := []string{
		"run",
		"--out-format=json",
		"--issues-exit-code=0", // Don't fail on findings
	}

	// Add packages to analyze
	if len(packages) > 0 {
		args = append(args, packages...)
	} else {
		args = append(args, "./...")
	}

	execResult := g.executor.Run(ctx, projectDir, "golangci-lint", args...)
	if execResult.Err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("golangci-lint execution failed: %v", execResult.Err))
		return result, nil
	}

	trimmed := strings.TrimSpace(string(execResult.Stdout))
	if trimmed == "" {
		return result, nil
	}

	// Parse JSON output
	var output golangciLintOutput
	if err := json.Unmarshal([]byte(trimmed), &output); err != nil {
		// Try to parse partial output
		result.Errors = append(result.Errors, fmt.Sprintf("golangci-lint output parse warning: %v", err))
		return result, nil
	}

	// Convert to common format
	for _, issue := range output.Issues {
		finding := Finding{
			Tool:     g.Name(),
			Rule:     issue.FromLinter,
			Severity: mapGolangciSeverity(issue.Severity),
			File:     normalizeFilePath(projectDir, issue.Pos.Filename),
			Line:     issue.Pos.Line,
			Column:   issue.Pos.Column,
			Message:  issue.Text,
			Category: mapGolangciCategory(issue.FromLinter),
		}
		result.AddFinding(finding)
	}

	return result, nil
}

// mapGolangciSeverity maps golangci-lint severity to common severity.
func mapGolangciSeverity(severity string) Severity {
	switch strings.ToLower(severity) {
	case "error":
		return SeverityHigh
	case "warning":
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// mapGolangciCategory maps linter name to category.
func mapGolangciCategory(linter string) Category {
	switch linter {
	case "gosec":
		return CategorySecurity
	case "staticcheck", "typecheck":
		return CategoryBug
	case "gofmt", "goimports", "govet", "gocritic":
		return CategoryStyle
	case "ineffassign", "deadcode", "unused", "varcheck":
		return CategoryUnused
	case "gocyclo", "gocognit":
		return CategoryComplexity
	case "depguard":
		return CategoryDeprecation
	default:
		return CategoryOther
	}
}

// normalizeFilePath converts absolute paths to relative paths.
func normalizeFilePath(projectDir, filePath string) string {
	if filepath.IsAbs(filePath) {
		if rel, err := filepath.Rel(projectDir, filePath); err == nil {
			return rel
		}
	}
	return filePath
}
