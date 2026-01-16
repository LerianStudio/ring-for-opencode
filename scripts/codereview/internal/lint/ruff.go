package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ruffOutput represents ruff JSON output (array of diagnostics).
type ruffOutput []ruffDiagnostic

type ruffDiagnostic struct {
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Location ruffLocation `json:"location"`
	Fix      *ruffFix     `json:"fix"`
	Filename string       `json:"filename"`
}

type ruffLocation struct {
	Row    int `json:"row"`
	Column int `json:"column"`
}

type ruffFix struct {
	Message string `json:"message"`
}

// Ruff implements the ruff linter wrapper.
type Ruff struct {
	executor *Executor
}

// NewRuff creates a new ruff wrapper.
func NewRuff() *Ruff {
	return &Ruff{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (r *Ruff) Name() string {
	return "ruff"
}

// Language returns the supported language.
func (r *Ruff) Language() Language {
	return LanguagePython
}

// TargetKind declares ruff prefers explicit file paths.
func (r *Ruff) TargetKind() TargetKind {
	return TargetKindFiles
}

// Available checks if ruff is installed.
func (r *Ruff) Available(ctx context.Context) bool {
	return r.executor.CommandAvailable(ctx, "ruff")
}

// Version returns the ruff version.
func (r *Ruff) Version(ctx context.Context) (string, error) {
	version, err := r.executor.GetVersion(ctx, "ruff", "--version")
	if err != nil {
		return "", err
	}
	// Extract version from "ruff X.Y.Z"
	parts := strings.Fields(version)
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return strings.TrimSpace(version), nil
}

// Run executes ruff on the specified files.
func (r *Ruff) Run(ctx context.Context, projectDir string, files []string) (*Result, error) {
	result := NewResult()

	version, err := r.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("ruff version check failed: %v", err))
	} else {
		result.ToolVersions["ruff"] = version
	}

	// Build arguments
	args := []string{
		"check",
		"--output-format", "json",
		"--exit-zero", // Don't fail on findings
	}

	// Add files to lint
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	execResult := r.executor.Run(ctx, projectDir, "ruff", args...)
	if execResult.Err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("ruff execution failed: %v", execResult.Err))
		return result, nil
	}

	// Parse JSON output
	var output ruffOutput
	if strings.TrimSpace(string(execResult.Stdout)) == "" {
		output = ruffOutput{}
	} else if err := json.Unmarshal(execResult.Stdout, &output); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("ruff output parse warning: %v", err))
		return result, nil
	}

	// Convert to common format
	for _, diag := range output {
		suggestion := ""
		if diag.Fix != nil {
			suggestion = diag.Fix.Message
		}

		finding := Finding{
			Tool:       r.Name(),
			Rule:       diag.Code,
			Severity:   mapRuffSeverity(diag.Code),
			File:       normalizeFilePath(projectDir, diag.Filename),
			Line:       diag.Location.Row,
			Column:     diag.Location.Column,
			Message:    diag.Message,
			Suggestion: suggestion,
			Category:   mapRuffCategory(diag.Code),
		}
		result.AddFinding(finding)
	}

	return result, nil
}

// mapRuffSeverity maps ruff codes to severity.
func mapRuffSeverity(code string) Severity {
	switch {
	case strings.HasPrefix(code, "S"):
		return SeverityHigh // Security
	case strings.HasPrefix(code, "E"):
		return SeverityWarning // Pycodestyle style (e.g., line length); map to warning intentionally
	case strings.HasPrefix(code, "F"):
		return SeverityWarning // Pyflakes
	case strings.HasPrefix(code, "W"):
		return SeverityWarning // Warnings
	case strings.HasPrefix(code, "B"):
		return SeverityWarning // Bugbear
	default:
		return SeverityInfo
	}
}

// mapRuffCategory maps ruff codes to categories.
func mapRuffCategory(code string) Category {
	switch {
	case strings.HasPrefix(code, "S"):
		return CategorySecurity
	case strings.HasPrefix(code, "F"):
		return CategoryBug
	case strings.HasPrefix(code, "E"):
		return CategoryStyle
	case strings.HasPrefix(code, "W"):
		return CategoryStyle
	case strings.HasPrefix(code, "B"):
		return CategoryBug
	case strings.HasPrefix(code, "I"):
		return CategoryStyle // Import sorting
	case strings.HasPrefix(code, "UP"):
		return CategoryDeprecation // Pyupgrade
	case strings.HasPrefix(code, "C"):
		return CategoryComplexity
	default:
		return CategoryOther
	}
}
