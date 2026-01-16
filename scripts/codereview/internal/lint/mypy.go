package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type mypyMessage struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// Mypy implements the mypy type checker wrapper.
type Mypy struct {
	executor *Executor
}

// NewMypy creates a new mypy wrapper.
func NewMypy() *Mypy {
	return &Mypy{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (m *Mypy) Name() string {
	return "mypy"
}

// Language returns the supported language.
func (m *Mypy) Language() Language {
	return LanguagePython
}

// TargetKind declares mypy prefers explicit file paths.
func (m *Mypy) TargetKind() TargetKind {
	return TargetKindFiles
}

// Available checks if mypy is installed.
func (m *Mypy) Available(ctx context.Context) bool {
	return m.executor.CommandAvailable(ctx, "mypy")
}

// Version returns the mypy version.
func (m *Mypy) Version(ctx context.Context) (string, error) {
	version, err := m.executor.GetVersion(ctx, "mypy", "--version")
	if err != nil {
		return "", err
	}
	// Extract version from "mypy X.Y.Z (compiled: yes)"
	parts := strings.Fields(version)
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return strings.TrimSpace(version), nil
}

// Run executes mypy type checking on the specified files.
func (m *Mypy) Run(ctx context.Context, projectDir string, files []string) (*Result, error) {
	result := NewResult()

	version, err := m.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("mypy version check failed: %v", err))
	} else {
		result.ToolVersions["mypy"] = version
	}

	// Build arguments
	args := []string{
		"--output", "json",
		"--no-error-summary",
		"--show-error-codes",
	}

	// Add files to check
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, ".")
	}

	execResult := m.executor.Run(ctx, projectDir, "mypy", args...)
	// mypy returns non-zero exit code when type errors are found
	// Only treat as failure if there's no output to parse
	if execResult.Err != nil && len(execResult.Stdout) == 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("mypy execution failed: %v", execResult.Err))
		return result, nil
	}

	// mypy JSON output is one JSON object per line
	lines := strings.Split(string(execResult.Stdout), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var msg mypyMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue // Skip malformed lines
		}

		finding := Finding{
			Tool:     m.Name(),
			Rule:     msg.Code,
			Severity: mapMypySeverity(msg.Severity),
			File:     normalizeFilePath(projectDir, msg.File),
			Line:     msg.Line,
			Column:   msg.Column,
			Message:  msg.Message,
			Category: CategoryType,
		}
		result.AddFinding(finding)
	}

	return result, nil
}

// mapMypySeverity maps mypy severity to common severity.
func mapMypySeverity(severity string) Severity {
	switch strings.ToLower(severity) {
	case "error":
		return SeverityHigh
	case "warning":
		return SeverityWarning
	case "note":
		return SeverityInfo
	default:
		return SeverityWarning
	}
}
