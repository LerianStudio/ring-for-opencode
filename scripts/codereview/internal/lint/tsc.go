package lint

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// tscDiagnosticRegex matches TypeScript compiler diagnostic output.
// Format: "file.ts(line,col): error TSxxxx: message"
var tscDiagnosticRegex = regexp.MustCompile(`^(.+)\((\d+),(\d+)\):\s+(error|warning)\s+(TS\d+):\s+(.+)$`)

// TSC implements the TypeScript compiler type checker wrapper.
type TSC struct {
	executor *Executor
}

// NewTSC creates a new tsc wrapper.
func NewTSC() *TSC {
	return &TSC{
		executor: NewExecutor(),
	}
}

// Name returns the linter name.
func (t *TSC) Name() string {
	return "tsc"
}

// Language returns the supported language.
func (t *TSC) Language() Language {
	return LanguageTypeScript
}

// TargetKind declares tsc can work with explicit files (or defaults to project).
func (t *TSC) TargetKind() TargetKind {
	return TargetKindFiles
}

// Available checks if tsc is installed.
func (t *TSC) Available(ctx context.Context) bool {
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Prefer project-local tsc via npx; ensure it is runnable
	res := t.executor.Run(checkCtx, "", "npx", "tsc", "--version")
	if res.Err == nil && res.ExitCode == 0 {
		return true
	}

	// Fall back to globally installed tsc
	return t.executor.CommandAvailable(ctx, "tsc")
}

// Version returns the tsc version.
func (t *TSC) Version(ctx context.Context) (string, error) {
	// Try npx tsc first (project-local)
	version, err := t.executor.GetVersion(ctx, "npx", "tsc", "--version")
	if err != nil {
		// Fall back to global tsc
		version, err = t.executor.GetVersion(ctx, "tsc", "--version")
	}
	if err != nil {
		return "", err
	}
	// Extract version from "Version X.Y.Z"
	parts := strings.Fields(version)
	for i, p := range parts {
		if p == "Version" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return strings.TrimSpace(version), nil
}

// Run executes tsc type checking on the project. When files is non-empty, tsc
// is invoked only for the provided file paths; otherwise the whole project is
// checked.
func (t *TSC) Run(ctx context.Context, projectDir string, files []string) (*Result, error) {
	result := NewResult()

	version, err := t.Version(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("tsc version check failed: %v", err))
	} else {
		result.ToolVersions["typescript"] = version
	}

	// Run tsc --noEmit to type check without emitting files
	args := []string{"tsc", "--noEmit", "--pretty", "false"}
	if len(files) > 0 {
		args = append(args, files...)
	}

	execResult := t.executor.Run(ctx, projectDir, "npx", args...)
	if execResult.Err != nil {
		// Try global tsc
		execResult = t.executor.Run(ctx, projectDir, "tsc", args[1:]...)
		if execResult.Err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("tsc execution failed: %v", execResult.Err))
			return result, nil
		}
	}

	// Parse output line by line
	scanner := bufio.NewScanner(strings.NewReader(string(execResult.Stdout)))
	for scanner.Scan() {
		line := scanner.Text()
		matches := tscDiagnosticRegex.FindStringSubmatch(line)
		if len(matches) != 7 {
			continue
		}

		lineNum, _ := strconv.Atoi(matches[2])
		col, _ := strconv.Atoi(matches[3])

		finding := Finding{
			Tool:     t.Name(),
			Rule:     matches[5], // TSxxxx
			Severity: mapTSCSeverity(matches[4]),
			File:     normalizeFilePath(projectDir, matches[1]),
			Line:     lineNum,
			Column:   col,
			Message:  matches[6],
			Category: CategoryType,
		}
		result.AddFinding(finding)
	}

	return result, nil
}

// mapTSCSeverity maps tsc error/warning to common severity.
func mapTSCSeverity(level string) Severity {
	switch strings.ToLower(level) {
	case "error":
		return SeverityHigh
	case "warning":
		return SeverityWarning
	default:
		return SeverityInfo
	}
}
