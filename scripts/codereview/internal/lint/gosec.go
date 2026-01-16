package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// gosecOutput represents gosec JSON output.
type gosecOutput struct {
	Issues []gosecIssue `json:"Issues"`
}

type gosecIssue struct {
	Severity   string `json:"severity"`
	Confidence string `json:"confidence"`
	RuleID     string `json:"rule_id"`
	Details    string `json:"details"`
	File       string `json:"file"`
	Line       string `json:"line"`
	Column     string `json:"column"`
	Code       string `json:"code"`
}

// Gosec implements the gosec wrapper.
type Gosec struct {
	executor *Executor
}

// NewGosec creates a new gosec wrapper.
func NewGosec() *Gosec {
	return &Gosec{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (g *Gosec) Name() string {
	return "gosec"
}

// Language returns the supported language.
func (g *Gosec) Language() Language {
	return LanguageGo
}

// TargetKind declares gosec wants package import paths.
func (g *Gosec) TargetKind() TargetKind {
	return TargetKindPackages
}

// Available checks if gosec is installed.
func (g *Gosec) Available(ctx context.Context) bool {
	return g.executor.CommandAvailable(ctx, "gosec")
}

// Version returns the gosec version.
func (g *Gosec) Version(ctx context.Context) (string, error) {
	version, err := g.executor.GetVersion(ctx, "gosec", "-version")
	if err != nil {
		return "", err
	}
	// Extract version from "Version: X.Y.Z" or similar
	for _, line := range strings.Split(version, "\n") {
		if strings.Contains(line, "Version:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}
	return strings.TrimSpace(version), nil
}

// Run executes gosec on the specified packages.
func (g *Gosec) Run(ctx context.Context, projectDir string, packages []string) (*Result, error) {
	result := NewResult()

	version, err := g.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("gosec version check failed: %v", err))
	} else {
		result.ToolVersions["gosec"] = version
	}

	// Build arguments
	args := []string{
		"-fmt=json",
		"-quiet",
		"-no-fail", // Don't exit non-zero on findings
	}

	// Add packages to analyze
	if len(packages) > 0 {
		args = append(args, packages...)
	} else {
		args = append(args, "./...")
	}

	execResult := g.executor.Run(ctx, projectDir, "gosec", args...)
	if execResult.Err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("gosec execution failed: %v", execResult.Err))
		return result, nil
	}

	// Parse JSON output
	var output gosecOutput
	if err := json.Unmarshal(execResult.Stdout, &output); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("gosec output parse warning: %v", err))
		return result, nil
	}

	// Convert to common format
	for _, issue := range output.Issues {
		line, err := strconv.Atoi(issue.Line)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("gosec output parse warning: file=%s line=%q column=%q err=%v", issue.File, issue.Line, issue.Column, err))
			line = 0
		}

		col, err := strconv.Atoi(issue.Column)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("gosec output parse warning: file=%s line=%q column=%q err=%v", issue.File, issue.Line, issue.Column, err))
			col = 0
		}

		finding := Finding{
			Tool:     g.Name(),
			Rule:     issue.RuleID,
			Severity: mapGosecSeverity(issue.Severity, issue.Confidence),
			File:     normalizeFilePath(projectDir, issue.File),
			Line:     line,
			Column:   col,
			Message:  issue.Details,
			Category: CategorySecurity,
		}
		result.AddFinding(finding)
	}

	return result, nil
}

// mapGosecSeverity maps gosec severity and confidence to common severity.
func mapGosecSeverity(severity, confidence string) Severity {
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
