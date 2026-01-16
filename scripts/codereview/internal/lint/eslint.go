package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// eslintOutput represents eslint JSON output (array of file results).
type eslintOutput []eslintFileResult

type eslintFileResult struct {
	FilePath string          `json:"filePath"`
	Messages []eslintMessage `json:"messages"`
}

type eslintMessage struct {
	RuleID   string `json:"ruleId"`
	Severity int    `json:"severity"` // 1 = warning, 2 = error
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

// ESLint implements the eslint wrapper.
type ESLint struct {
	executor *Executor
}

// NewESLint creates a new eslint wrapper.
func NewESLint() *ESLint {
	return &ESLint{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (e *ESLint) Name() string {
	return "eslint"
}

// Language returns the supported language.
func (e *ESLint) Language() Language {
	return LanguageTypeScript
}

// TargetKind declares eslint prefers explicit file paths.
func (e *ESLint) TargetKind() TargetKind {
	return TargetKindFiles
}

// Available checks if eslint is installed.
func (e *ESLint) Available(ctx context.Context) bool {
	// Prefer direct eslint binary when present.
	if e.executor.CommandAvailable(ctx, "eslint") {
		result := e.executor.Run(ctx, "", "eslint", "--version")
		return result.Err == nil && result.ExitCode == 0
	}

	// Fall back to npx invocation without installing packages.
	if !e.executor.CommandAvailable(ctx, "npx") {
		return false
	}

	result := e.executor.Run(ctx, "", "npx", "--no-install", "eslint", "--version")
	return result.Err == nil && result.ExitCode == 0
}

// Version returns the eslint version.
func (e *ESLint) Version(ctx context.Context) (string, error) {
	version, err := e.executor.GetVersion(ctx, "npx", "eslint", "--version")
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(strings.TrimSpace(version), "v"), nil
}

// Run executes eslint on the specified files.
func (e *ESLint) Run(ctx context.Context, projectDir string, files []string) (*Result, error) {
	result := NewResult()

	version, err := e.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("eslint version check failed: %v", err))
	} else {
		result.ToolVersions["eslint"] = version
	}

	// Build arguments
	args := []string{
		"eslint",
		"--format", "json",
		"--no-error-on-unmatched-pattern",
	}

	// Add files to lint
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	execResult := e.executor.Run(ctx, projectDir, "npx", args...)
	if execResult.Err != nil && (len(execResult.Stdout) == 0 || execResult.ExitCode == 2) {
		result.Errors = append(result.Errors, fmt.Sprintf("eslint execution failed: %v", execResult.Err))
		return result, nil
	}

	// Parse JSON output
	var output eslintOutput
	if err := json.Unmarshal(execResult.Stdout, &output); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("eslint output parse warning: %v", err))
		return result, nil
	}

	// Convert to common format
	for _, file := range output {
		for _, msg := range file.Messages {
			ruleID := msg.RuleID
			if ruleID == "" {
				ruleID = "parse-error"
			}

			finding := Finding{
				Tool:     e.Name(),
				Rule:     ruleID,
				Severity: mapESLintSeverity(msg.Severity),
				File:     normalizeFilePath(projectDir, file.FilePath),
				Line:     msg.Line,
				Column:   msg.Column,
				Message:  msg.Message,
				Category: mapESLintCategory(ruleID),
			}
			result.AddFinding(finding)
		}
	}

	return result, nil
}

// mapESLintSeverity maps eslint severity (1=warn, 2=error) to common severity.
func mapESLintSeverity(severity int) Severity {
	switch severity {
	case 2:
		return SeverityHigh
	case 1:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// mapESLintCategory maps eslint rule IDs to categories.
func mapESLintCategory(ruleID string) Category {
	switch {
	case strings.HasPrefix(ruleID, "@typescript-eslint/"):
		return CategoryType
	case strings.Contains(ruleID, "security"):
		return CategorySecurity
	case strings.Contains(ruleID, "no-unused"):
		return CategoryUnused
	case strings.HasPrefix(ruleID, "import/"):
		return CategoryStyle
	case strings.HasPrefix(ruleID, "react/") || strings.HasPrefix(ruleID, "react-"):
		return CategoryStyle
	case ruleID == "parse-error":
		return CategoryBug
	default:
		return CategoryStyle
	}
}
