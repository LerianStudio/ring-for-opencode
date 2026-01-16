package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// banditOutput represents bandit JSON output.
type banditOutput struct {
	Results []banditResult `json:"results"`
	Metrics banditMetrics  `json:"metrics"`
}

type banditResult struct {
	Code            string `json:"code"`
	Filename        string `json:"filename"`
	IssueText       string `json:"issue_text"`
	IssueSeverity   string `json:"issue_severity"`
	IssueConfidence string `json:"issue_confidence"`
	LineNumber      int    `json:"line_number"`
	LineRange       []int  `json:"line_range"`
	MoreInfo        string `json:"more_info"`
	TestID          string `json:"test_id"`
	TestName        string `json:"test_name"`
}

// TODO(review): TotalIssues field is unused, consider removing or using (code-reviewer, 2026-01-13, Low)
type banditMetrics struct {
	TotalIssues int `json:"issues"`
}

// Bandit implements the bandit security scanner wrapper.
type Bandit struct {
	executor *Executor
}

// NewBandit creates a new bandit wrapper.
func NewBandit() *Bandit {
	return &Bandit{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (b *Bandit) Name() string {
	return "bandit"
}

// Language returns the supported language.
func (b *Bandit) Language() Language {
	return LanguagePython
}

// TargetKind declares bandit prefers explicit file paths.
func (b *Bandit) TargetKind() TargetKind {
	return TargetKindFiles
}

// Available checks if bandit is installed.
func (b *Bandit) Available(ctx context.Context) bool {
	return b.executor.CommandAvailable(ctx, "bandit")
}

// Version returns the bandit version.
func (b *Bandit) Version(ctx context.Context) (string, error) {
	version, err := b.executor.GetVersion(ctx, "bandit", "--version")
	if err != nil {
		return "", err
	}
	// Extract version from "bandit X.Y.Z"
	parts := strings.Fields(version)
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return strings.TrimSpace(version), nil
}

// Run executes bandit security analysis on the specified files.
func (b *Bandit) Run(ctx context.Context, projectDir string, files []string) (*Result, error) {
	result := NewResult()

	version, err := b.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("bandit version check failed: %v", err))
	} else {
		result.ToolVersions["bandit"] = version
	}

	// Build arguments
	args := []string{
		"-f", "json",
		"-q", // Quiet mode
	}

	// Add files to scan
	if len(files) > 0 {
		args = append(args, files...)
	} else {
		args = append(args, "-r", ".") // Recursive scan
	}

	execResult := b.executor.Run(ctx, projectDir, "bandit", args...)
	if execResult.Err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("bandit execution failed: %v", execResult.Err))
		return result, nil
	}
	if execResult.ExitCode > 1 {
		result.Errors = append(result.Errors, fmt.Sprintf("bandit execution failed with exit code %d: %s", execResult.ExitCode, strings.TrimSpace(string(execResult.Stderr))))
		return result, nil
	}

	// Parse JSON output
	var output banditOutput
	if err := json.Unmarshal(execResult.Stdout, &output); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("bandit output parse warning: %v", err))
		return result, nil
	}

	// Convert to common format
	for _, res := range output.Results {
		finding := Finding{
			Tool:       b.Name(),
			Rule:       res.TestID,
			Severity:   mapBanditSeverity(res.IssueSeverity, res.IssueConfidence),
			File:       normalizeFilePath(projectDir, res.Filename),
			Line:       res.LineNumber,
			Column:     1, // Bandit doesn't provide column info
			Message:    fmt.Sprintf("%s: %s", res.TestName, res.IssueText),
			Suggestion: res.MoreInfo,
			Category:   CategorySecurity,
		}
		result.AddFinding(finding)
	}

	return result, nil
}

// mapBanditSeverity maps bandit severity and confidence to common severity.
func mapBanditSeverity(severity, confidence string) Severity {
	sev := strings.ToUpper(severity)
	conf := strings.ToUpper(confidence)

	if sev == "HIGH" && conf == "HIGH" {
		return SeverityCritical
	}
	if sev == "HIGH" {
		return SeverityHigh
	}
	if sev == "MEDIUM" {
		return SeverityWarning
	}
	return SeverityInfo
}
