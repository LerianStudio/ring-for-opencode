package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// pylintOutput represents pylint JSON output (array of messages).
type pylintOutput []pylintMessage

type pylintMessage struct {
	Type      string `json:"type"` // convention, refactor, warning, error, fatal
	Module    string `json:"module"`
	Obj       string `json:"obj"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Path      string `json:"path"`
	Symbol    string `json:"symbol"`
	Message   string `json:"message"`
	MessageID string `json:"message-id"`
}

// Pylint implements the pylint wrapper.
type Pylint struct {
	executor *Executor
}

// NewPylint creates a new pylint wrapper.
func NewPylint() *Pylint {
	return &Pylint{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (p *Pylint) Name() string {
	return "pylint"
}

// Language returns the supported language.
func (p *Pylint) Language() Language {
	return LanguagePython
}

// TargetKind declares pylint prefers explicit file paths.
func (p *Pylint) TargetKind() TargetKind {
	return TargetKindFiles
}

// Available checks if pylint is installed.
func (p *Pylint) Available(ctx context.Context) bool {
	return p.executor.CommandAvailable(ctx, "pylint")
}

// Version returns the pylint version.
func (p *Pylint) Version(ctx context.Context) (string, error) {
	version, err := p.executor.GetVersion(ctx, "pylint", "--version")
	if err != nil {
		return "", err
	}
	// Extract version from "pylint X.Y.Z\n..."
	lines := strings.Split(version, "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}
	return strings.TrimSpace(version), nil
}

// Run executes pylint on the specified files.
func (p *Pylint) Run(ctx context.Context, projectDir string, files []string) (*Result, error) {
	result := NewResult()

	version, err := p.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("pylint version check failed: %v", err))
	} else {
		result.ToolVersions["pylint"] = version
	}

	// Build arguments
	args := []string{
		"--output-format=json",
		"--exit-zero", // Don't fail on findings
	}

	// Add files to lint
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	execResult := p.executor.Run(ctx, projectDir, "pylint", args...)
	if execResult.Err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("pylint execution failed: %v", execResult.Err))
		return result, nil
	}

	// Parse JSON output
	var output pylintOutput
	if err := json.Unmarshal(execResult.Stdout, &output); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("pylint output parse warning: %v", err))
		return result, nil
	}

	// Convert to common format
	for _, msg := range output {
		finding := Finding{
			Tool:     p.Name(),
			Rule:     msg.MessageID,
			Severity: mapPylintSeverity(msg.Type),
			File:     normalizeFilePath(projectDir, msg.Path),
			Line:     msg.Line,
			Column:   msg.Column,
			Message:  fmt.Sprintf("%s: %s", msg.Symbol, msg.Message),
			Category: mapPylintCategory(msg.Type, msg.MessageID),
		}
		result.AddFinding(finding)
	}

	return result, nil
}

// mapPylintSeverity maps pylint message types to severity.
func mapPylintSeverity(msgType string) Severity {
	switch strings.ToLower(msgType) {
	case "fatal", "error":
		return SeverityHigh
	case "warning":
		return SeverityWarning
	case "refactor", "convention":
		return SeverityInfo
	default:
		return SeverityInfo
	}
}

// mapPylintCategory maps pylint message types and IDs to categories.
func mapPylintCategory(msgType, msgID string) Category {
	switch strings.ToLower(msgType) {
	case "fatal", "error":
		return CategoryBug
	case "warning":
		if strings.HasPrefix(msgID, "W0611") || strings.HasPrefix(msgID, "W0612") {
			return CategoryUnused
		}
		return CategoryBug
	case "refactor":
		return CategoryComplexity
	case "convention":
		return CategoryStyle
	default:
		return CategoryOther
	}
}
