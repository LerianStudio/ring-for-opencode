package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// staticcheckIssue represents a single staticcheck finding.
type staticcheckIssue struct {
	Code     string              `json:"code"`
	Severity string              `json:"severity"`
	Location staticcheckLocation `json:"location"`
	Message  string              `json:"message"`
	End      staticcheckLocation `json:"end"`
}

type staticcheckLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// Staticcheck implements the staticcheck wrapper.
type Staticcheck struct {
	executor *Executor
}

// NewStaticcheck creates a new staticcheck wrapper.
func NewStaticcheck() *Staticcheck {
	return &Staticcheck{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (s *Staticcheck) Name() string {
	return "staticcheck"
}

// Language returns the supported language.
func (s *Staticcheck) Language() Language {
	return LanguageGo
}

// TargetKind declares staticcheck wants package import paths.
func (s *Staticcheck) TargetKind() TargetKind {
	return TargetKindPackages
}

// Available checks if staticcheck is installed.
func (s *Staticcheck) Available(ctx context.Context) bool {
	return s.executor.CommandAvailable(ctx, "staticcheck")
}

// Version returns the staticcheck version.
func (s *Staticcheck) Version(ctx context.Context) (string, error) {
	version, err := s.executor.GetVersion(ctx, "staticcheck", "-version")
	if err != nil {
		return "", err
	}
	// Extract version from "staticcheck 2024.1.1 (v0.5.1)"
	parts := strings.Fields(version)
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return version, nil
}

// Run executes staticcheck on the specified packages.
func (s *Staticcheck) Run(ctx context.Context, projectDir string, packages []string) (*Result, error) {
	result := NewResult()

	version, err := s.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("staticcheck version check failed: %v", err))
	} else {
		result.ToolVersions["staticcheck"] = version
	}

	// Build arguments
	args := []string{"-f", "json"}

	// Add packages to analyze
	if len(packages) > 0 {
		args = append(args, packages...)
	} else {
		args = append(args, "./...")
	}

	execResult := s.executor.Run(ctx, projectDir, "staticcheck", args...)
	if msg, stop := staticcheckExecutionError(execResult); stop {
		result.Errors = append(result.Errors, msg)
		return result, nil
	}

	// Parse JSON lines output (one JSON object per line)
	lines := strings.Split(string(execResult.Stdout), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var issue staticcheckIssue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			continue // Skip malformed lines
		}

		finding := Finding{
			Tool:     s.Name(),
			Rule:     issue.Code,
			Severity: mapStaticcheckSeverity(issue.Code, issue.Severity),
			File:     normalizeFilePath(projectDir, issue.Location.File),
			Line:     issue.Location.Line,
			Column:   issue.Location.Column,
			Message:  issue.Message,
			Category: mapStaticcheckCategory(issue.Code),
		}
		result.AddFinding(finding)
	}

	return result, nil
}

// staticcheckExecutionError determines whether staticcheck execution should be treated as a failure.
// staticcheck exits 1 when findings exist, but 2+ (or 1 with no output) indicate execution issues.
func staticcheckExecutionError(execResult *ExecResult) (string, bool) {
	if execResult.Err != nil {
		return fmt.Sprintf("staticcheck execution failed: %v", execResult.Err), true
	}

	if execResult.ExitCode > 1 || (execResult.ExitCode == 1 && len(execResult.Stdout) == 0) {
		stderr := strings.TrimSpace(string(execResult.Stderr))
		if stderr == "" {
			stderr = "no output"
		}
		return fmt.Sprintf("staticcheck execution failed with exit code %d: %s", execResult.ExitCode, stderr), true
	}

	return "", false
}

// mapStaticcheckSeverity maps staticcheck codes to severity.
func mapStaticcheckSeverity(code, severity string) Severity {
	if strings.HasPrefix(code, "SA") {
		return SeverityWarning
	}
	if strings.HasPrefix(code, "S1") {
		return SeverityInfo
	}
	if strings.HasPrefix(code, "ST1") {
		return SeverityInfo
	}
	if severity == "error" {
		return SeverityHigh
	}
	return SeverityWarning
}

// mapStaticcheckCategory maps staticcheck codes to categories.
func mapStaticcheckCategory(code string) Category {
	switch {
	case strings.HasPrefix(code, "SA1"):
		return CategoryBug
	case strings.HasPrefix(code, "SA2"):
		return CategoryBug
	case strings.HasPrefix(code, "SA3"):
		return CategoryBug
	case strings.HasPrefix(code, "SA4"):
		return CategoryBug
	case strings.HasPrefix(code, "SA5"):
		return CategoryBug
	case strings.HasPrefix(code, "SA6"):
		return CategoryPerformance
	case strings.HasPrefix(code, "SA7"):
		return CategoryBug
	case strings.HasPrefix(code, "SA8"):
		return CategoryStyle
	case strings.HasPrefix(code, "SA9"):
		return CategorySecurity
	case strings.HasPrefix(code, "S1"):
		return CategoryStyle
	case strings.HasPrefix(code, "ST1"):
		return CategoryStyle
	case strings.HasPrefix(code, "QF"):
		return CategoryStyle
	default:
		return CategoryOther
	}
}
