// Package main provides integration and unit tests for the run-all CLI binary.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// testBinaryName is the name of the test binary.
const testBinaryName = "run-all-test"

// buildTestBinary builds the run-all binary for testing.
// Returns the path to the built binary.
func buildTestBinary(t *testing.T) string {
	t.Helper()

	// Build in the current directory (where main.go is)
	binaryPath := filepath.Join(".", testBinaryName)

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = "." // Ensure we're in the right directory

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
	}

	// Return absolute path for reliable execution
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	return absPath
}

// cleanupTestBinary removes the test binary.
func cleanupTestBinary(t *testing.T, binaryPath string) {
	t.Helper()

	if err := os.Remove(binaryPath); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: failed to remove test binary %s: %v", binaryPath, err)
	}
}

// ============================================================================
// parseSkipList Unit Tests
// ============================================================================

// TestParseSkipList_Empty verifies that empty skip string returns empty map.
func TestParseSkipList_Empty(t *testing.T) {
	result := parseSkipList("")

	if len(result) != 0 {
		t.Errorf("Expected empty map for empty string, got %d entries: %v", len(result), result)
	}
}

// TestParseSkipList_Single verifies parsing of a single phase.
func TestParseSkipList_Single(t *testing.T) {
	result := parseSkipList("scope")

	if len(result) != 1 {
		t.Errorf("Expected 1 entry, got %d: %v", len(result), result)
	}

	if !result["scope"] {
		t.Error("Expected 'scope' to be in skip list")
	}
}

// TestParseSkipList_Multiple verifies parsing of multiple phases.
func TestParseSkipList_Multiple(t *testing.T) {
	result := parseSkipList("scope,ast,callgraph")

	expectedPhases := []string{"scope", "ast", "callgraph"}
	if len(result) != len(expectedPhases) {
		t.Errorf("Expected %d entries, got %d: %v", len(expectedPhases), len(result), result)
	}

	for _, phase := range expectedPhases {
		if !result[phase] {
			t.Errorf("Expected '%s' to be in skip list", phase)
		}
	}
}

// TestParseSkipList_Whitespace verifies that whitespace is trimmed correctly.
func TestParseSkipList_Whitespace(t *testing.T) {
	result := parseSkipList(" scope , ast ")

	expectedPhases := []string{"scope", "ast"}
	if len(result) != len(expectedPhases) {
		t.Errorf("Expected %d entries, got %d: %v", len(expectedPhases), len(result), result)
	}

	// Verify trimmed names work
	if !result["scope"] {
		t.Error("Expected 'scope' (trimmed) to be in skip list")
	}
	if !result["ast"] {
		t.Error("Expected 'ast' (trimmed) to be in skip list")
	}

	// Verify whitespace versions are not present
	if result[" scope "] {
		t.Error("Whitespace should be trimmed from 'scope'")
	}
}

// TestParseSkipList_InvalidPhase verifies warning for unknown phase names.
func TestParseSkipList_InvalidPhase(t *testing.T) {
	// Capture stderr to check for warning
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	result := parseSkipList("scope,unknown-phase,ast")

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close stderr pipe: %v", err)
	}
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read stderr output: %v", err)
	}
	stderrOutput := buf.String()

	// Valid phases should still be in result
	if !result["scope"] || !result["ast"] {
		t.Error("Valid phases should still be parsed")
	}

	// Unknown phase should be in result (but with warning)
	if !result["unknown-phase"] {
		t.Error("Unknown phase should still be in skip list")
	}

	// Should have warning message
	if !strings.Contains(stderrOutput, "unknown phase") || !strings.Contains(stderrOutput, "unknown-phase") {
		t.Errorf("Expected warning about unknown phase, got: %s", stderrOutput)
	}
}

// TestParseSkipList_AllPhases verifies all valid phase names are accepted.
func TestParseSkipList_AllPhases(t *testing.T) {
	allPhases := "scope,static-analysis,ast,callgraph,dataflow,context"
	result := parseSkipList(allPhases)

	expectedPhases := []string{"scope", "static-analysis", "ast", "callgraph", "dataflow", "context"}
	if len(result) != len(expectedPhases) {
		t.Errorf("Expected %d entries, got %d: %v", len(expectedPhases), len(result), result)
	}

	for _, phase := range expectedPhases {
		if !result[phase] {
			t.Errorf("Expected '%s' to be in skip list", phase)
		}
	}
}

// TestParseSkipList_EmptyEntries verifies empty entries are ignored.
func TestParseSkipList_EmptyEntries(t *testing.T) {
	result := parseSkipList("scope,,ast,")

	if len(result) != 2 {
		t.Errorf("Expected 2 entries (empty entries ignored), got %d: %v", len(result), result)
	}

	if !result["scope"] || !result["ast"] {
		t.Error("Valid phases should be in skip list")
	}
}

// ============================================================================
// executePhase Unit Tests
// ============================================================================

// TestExecutePhase_BinaryNotFound verifies error when binary doesn't exist.
func TestExecutePhase_BinaryNotFound(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config{
		binDir:    "/nonexistent/binary/path",
		outputDir: filepath.Join(tempDir, "output"),
	}

	phase := Phase{
		Name:    "test-phase",
		Binary:  "nonexistent-binary",
		Timeout: 5 * time.Second,
		Args:    func(cfg *config) []string { return []string{} },
	}

	ctx := context.Background()
	result := executePhase(ctx, cfg, phase, map[string]bool{})

	if result.Success {
		t.Error("Expected failure for non-existent binary")
	}

	if result.Skipped {
		t.Error("Phase should not be marked as skipped")
	}

	if result.Error == nil {
		t.Error("Expected error for non-existent binary")
	}

	if !strings.Contains(result.Error.Error(), "binary not found") {
		t.Errorf("Expected 'binary not found' error, got: %v", result.Error)
	}
}

// TestExecutePhase_Skipped verifies skipped phases are handled correctly.
func TestExecutePhase_Skipped(t *testing.T) {
	binDir := t.TempDir()
	tempOutputDir := filepath.Join(t.TempDir(), "output")
	cfg := &config{
		binDir:    binDir,
		outputDir: tempOutputDir,
	}

	phase := Phase{
		Name:    "test-phase",
		Binary:  "test-binary",
		Timeout: 5 * time.Second,
		Args:    func(cfg *config) []string { return []string{} },
	}

	skipSet := map[string]bool{"test-phase": true}
	ctx := context.Background()
	result := executePhase(ctx, cfg, phase, skipSet)

	if !result.Skipped {
		t.Error("Expected phase to be marked as skipped")
	}

	if !result.Success {
		t.Error("Skipped phases should be marked as successful")
	}

	if result.Error != nil {
		t.Errorf("Skipped phases should not have errors, got: %v", result.Error)
	}
}

// TestExecutePhase_Timeout verifies that phases are killed on timeout.
func TestExecutePhase_Timeout(t *testing.T) {
	// Create a temporary directory for our test binary
	tempDir := t.TempDir()
	binDir := tempDir

	// Create a slow script/binary that sleeps
	var sleepBinary string
	var sleepBinaryPath string

	if runtime.GOOS == "windows" {
		// On Windows, create a batch file
		sleepBinary = "slow-phase.bat"
		sleepBinaryPath = filepath.Join(binDir, sleepBinary)
		script := "@echo off\nping -n 30 127.0.0.1 > nul\n"
		if err := os.WriteFile(sleepBinaryPath, []byte(script), 0o755); err != nil {
			t.Fatalf("Failed to create slow script: %v", err)
		}
	} else {
		// On Unix, create a shell script
		sleepBinary = "slow-phase"
		sleepBinaryPath = filepath.Join(binDir, sleepBinary)
		script := "#!/bin/sh\nsleep 30\n"
		if err := os.WriteFile(sleepBinaryPath, []byte(script), 0o755); err != nil {
			t.Fatalf("Failed to create slow script: %v", err)
		}
	}

	cfg := &config{
		binDir:    binDir,
		outputDir: filepath.Join(tempDir, "output"),
	}

	phase := Phase{
		Name:    "slow-phase",
		Binary:  sleepBinary,
		Timeout: 100 * time.Millisecond, // Very short timeout
		Args:    func(cfg *config) []string { return []string{} },
	}

	ctx := context.Background()
	start := time.Now()
	result := executePhase(ctx, cfg, phase, map[string]bool{})
	elapsed := time.Since(start)

	// Should have finished quickly due to timeout, not wait full 30 seconds
	if elapsed > 5*time.Second {
		t.Errorf("Expected timeout to trigger quickly, took %v", elapsed)
	}

	if result.Success {
		t.Error("Expected failure due to timeout")
	}

	if result.Error == nil {
		t.Error("Expected timeout error")
	}

	if !strings.Contains(result.Error.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", result.Error)
	}
}

// TestExecutePhase_Success verifies successful phase execution.
func TestExecutePhase_Success(t *testing.T) {
	// Create a temporary directory for our test binary
	tempDir := t.TempDir()
	binDir := tempDir

	// Create a fast script/binary that exits quickly
	var fastBinary string
	var fastBinaryPath string

	if runtime.GOOS == "windows" {
		fastBinary = "fast-phase.bat"
		fastBinaryPath = filepath.Join(binDir, fastBinary)
		script := "@echo off\necho success\n"
		if err := os.WriteFile(fastBinaryPath, []byte(script), 0o755); err != nil {
			t.Fatalf("Failed to create fast script: %v", err)
		}
	} else {
		fastBinary = "fast-phase"
		fastBinaryPath = filepath.Join(binDir, fastBinary)
		script := "#!/bin/sh\necho success\nexit 0\n"
		if err := os.WriteFile(fastBinaryPath, []byte(script), 0o755); err != nil {
			t.Fatalf("Failed to create fast script: %v", err)
		}
	}

	cfg := &config{
		binDir:    binDir,
		outputDir: filepath.Join(tempDir, "output"),
	}

	phase := Phase{
		Name:    "fast-phase",
		Binary:  fastBinary,
		Timeout: 30 * time.Second,
		Args:    func(cfg *config) []string { return []string{} },
	}

	ctx := context.Background()
	result := executePhase(ctx, cfg, phase, map[string]bool{})

	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}

	if result.Skipped {
		t.Error("Phase should not be marked as skipped")
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got: %v", result.Error)
	}
}

// TestExecutePhase_ContextCancellation verifies phases are killed on context cancellation.
func TestExecutePhase_ContextCancellation(t *testing.T) {
	// Create a temporary directory for our test binary
	tempDir := t.TempDir()
	binDir := tempDir

	// Create a slow script
	var sleepBinary string
	var sleepBinaryPath string

	if runtime.GOOS == "windows" {
		sleepBinary = "slow-cancel.bat"
		sleepBinaryPath = filepath.Join(binDir, sleepBinary)
		script := "@echo off\nping -n 30 127.0.0.1 > nul\n"
		if err := os.WriteFile(sleepBinaryPath, []byte(script), 0o755); err != nil {
			t.Fatalf("Failed to create slow script: %v", err)
		}
	} else {
		sleepBinary = "slow-cancel"
		sleepBinaryPath = filepath.Join(binDir, sleepBinary)
		script := "#!/bin/sh\nsleep 30\n"
		if err := os.WriteFile(sleepBinaryPath, []byte(script), 0o755); err != nil {
			t.Fatalf("Failed to create slow script: %v", err)
		}
	}

	cfg := &config{
		binDir:    binDir,
		outputDir: filepath.Join(tempDir, "output"),
	}

	phase := Phase{
		Name:    "slow-cancel",
		Binary:  sleepBinary,
		Timeout: 30 * time.Second, // Long timeout
		Args:    func(cfg *config) []string { return []string{} },
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result := executePhase(ctx, cfg, phase, map[string]bool{})
	elapsed := time.Since(start)

	// Should finish quickly due to cancellation
	if elapsed > 5*time.Second {
		t.Errorf("Expected cancellation to trigger quickly, took %v", elapsed)
	}

	if result.Success {
		t.Error("Expected failure due to cancellation")
	}

	if result.Error == nil {
		t.Error("Expected interruption error")
	}

	if !strings.Contains(result.Error.Error(), "interrupt") {
		t.Errorf("Expected interruption error, got: %v", result.Error)
	}
}

// TestExecutePhase_ContextPhaseSpecialHandling verifies the context phase runs as a binary.
func TestExecutePhase_ContextPhaseSpecialHandling(t *testing.T) {
	// Create temp directory with mock phase outputs
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	createMockPhaseOutputs(t, outputDir)

	compilePath := buildCompileContextBinary(t, tempDir)

	cfg := &config{
		binDir:    tempDir,
		outputDir: outputDir,
	}

	phase := Phase{
		Name:    "context",
		Binary:  filepath.Base(compilePath),
		Timeout: 30 * time.Second,
		Args: func(cfg *config) []string {
			return []string{"--input", cfg.outputDir, "--output", cfg.outputDir}
		},
	}

	ctx := context.Background()
	result := executePhase(ctx, cfg, phase, map[string]bool{})

	// Context phase uses external compiler, should succeed with mock data
	if !result.Success {
		t.Errorf("Expected success for context phase, got error: %v", result.Error)
	}

	// Verify context files were created
	expectedFile := filepath.Join(outputDir, "context-code-reviewer.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Context file not created: %s", expectedFile)
	}
}

func buildCompileContextBinary(t *testing.T, dir string) string {
	t.Helper()
	binaryPath := filepath.Join(dir, "compile-context")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/compile-context")
	buildCmd.Dir = filepath.Join(repoRoot(t), "scripts", "codereview")
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build compile-context binary: %v\nOutput: %s", err, string(output))
	}
	return binaryPath
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file path")
	}
	root := filepath.Join(filepath.Dir(filename), "..", "..", "..", "..")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	return absRoot
}

// ============================================================================
// Main Binary Integration Tests
// ============================================================================

// TestMain_Version verifies that --version flag outputs version string.
func TestMain_Version(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--version failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify version output format
	if !strings.HasPrefix(outputStr, "run-all version ") {
		t.Errorf("Expected version output to start with 'run-all version ', got: %s", outputStr)
	}

	// The default version is "dev" when not built with ldflags
	if !strings.Contains(outputStr, "dev") {
		t.Logf("Note: Version is not 'dev', likely built with ldflags: %s", outputStr)
	}
}

// TestMain_Help verifies that --help outputs usage information.
func TestMain_Help(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	// --help typically exits with code 0, but Go's flag package exits with 2
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() != 2 {
				t.Fatalf("--help exited with unexpected code %d: %v\nOutput: %s",
					exitErr.ExitCode(), err, string(output))
			}
		} else {
			t.Fatalf("--help failed unexpectedly: %v\nOutput: %s", err, string(output))
		}
	}

	outputStr := string(output)

	// Verify help output contains expected sections
	expectedStrings := []string{
		"Usage:",
		"run-all",
		"-base",
		"-head",
		"-files",
		"-files-from",
		"-output",
		"-skip",
		"-verbose",
		"-version",
		"Phases",
		"scope",
		"static-analysis",
		"ast",
		"callgraph",
		"dataflow",
		"context",
		"Examples:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output missing expected string %q\nOutput: %s", expected, outputStr)
		}
	}
}

// TestRun_AllPhasesSkipped verifies behavior when all phases are skipped.
func TestRun_AllPhasesSkipped(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	tempDir := t.TempDir()

	// Skip all phases
	cmd := exec.Command(binaryPath,
		"--output="+tempDir,
		"--skip=scope,static-analysis,ast,callgraph,dataflow,context",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success when all phases skipped, got error: %v\nStderr: %s",
			err, stderr.String())
	}

	stderrStr := stderr.String()

	// Should show all phases as skipped
	expectedSkips := []string{
		"[SKIP] scope", "[SKIP] static-analysis", "[SKIP] ast",
		"[SKIP] callgraph", "[SKIP] dataflow", "[SKIP] context",
	}

	for _, expected := range expectedSkips {
		if !strings.Contains(stderrStr, expected) {
			t.Errorf("Expected skip message for phase, missing: %s\nStderr: %s", expected, stderrStr)
		}
	}

	// Summary should show all skipped
	if !strings.Contains(stderrStr, "0 passed") || !strings.Contains(stderrStr, "6 skipped") {
		t.Errorf("Expected summary to show 0 passed, 6 skipped\nStderr: %s", stderrStr)
	}
}

// TestRun_PhaseFailure verifies that phase failures are reported but execution continues.
func TestRun_PhaseFailure(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	tempDir := t.TempDir()

	// Create a valid but empty bin-dir - phases will fail because binaries are missing
	emptyBinDir := t.TempDir()

	// Run without required binaries - phases will fail
	// Skip context phase since it uses internal compiler
	cmd := exec.Command(binaryPath,
		"--output="+tempDir,
		"--bin-dir="+emptyBinDir,
		"--skip=context", // Skip context since it doesn't use binaries
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Should exit with error due to failed phases
	err := cmd.Run()
	if err == nil {
		t.Log("Note: Command succeeded despite missing binaries")
	}

	stderrStr := stderr.String()

	// Should show failures for phases
	if !strings.Contains(stderrStr, "[FAIL]") {
		t.Errorf("Expected at least one [FAIL] message\nStderr: %s", stderrStr)
	}

	// Summary should show failures
	if !strings.Contains(stderrStr, "failed") {
		t.Errorf("Expected 'failed' in summary\nStderr: %s", stderrStr)
	}
}

// TestRun_OutputDirCreation verifies that output directory is created.
func TestRun_OutputDirCreation(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create nested output path that doesn't exist
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "nested", "deeply", "output")

	// Run with all phases skipped to avoid binary requirements
	cmd := exec.Command(binaryPath,
		"--output="+outputDir,
		"--skip=scope,static-analysis,ast,callgraph,dataflow,context",
	)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Verify output directory was created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Errorf("Output directory was not created: %s", outputDir)
	}
}

// TestRun_VerboseOutput verifies verbose mode includes additional information.
func TestRun_VerboseOutput(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	tempDir := t.TempDir()

	cmd := exec.Command(binaryPath,
		"--output="+tempDir,
		"--verbose",
		"--skip=scope,static-analysis,ast,callgraph,dataflow,context",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success, got error: %v\nStderr: %s", err, stderr.String())
	}

	stderrStr := stderr.String()

	// Verbose output should include configuration
	expectedVerbose := []string{
		"Configuration:",
		"Base ref:",
		"Head ref:",
		"Output directory:",
		"Binary directory:",
	}

	for _, expected := range expectedVerbose {
		if !strings.Contains(stderrStr, expected) {
			t.Errorf("Verbose output missing expected string %q\nStderr: %s", expected, stderrStr)
		}
	}
}

// TestRun_SummaryOutput verifies summary is printed at the end.
func TestRun_SummaryOutput(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	tempDir := t.TempDir()

	cmd := exec.Command(binaryPath,
		"--output="+tempDir,
		"--skip=scope,static-analysis,ast,callgraph,dataflow,context",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	stderrStr := stderr.String()

	// Summary should include expected sections
	expectedSummary := []string{
		"Code Review Pre-Analysis Summary",
		"Phases:",
		"Total time:",
		"Phase Breakdown:",
		"completed successfully",
	}

	for _, expected := range expectedSummary {
		if !strings.Contains(stderrStr, expected) {
			t.Errorf("Summary missing expected string %q\nStderr: %s", expected, stderrStr)
		}
	}
}

// TestRun_ExitCodes verifies correct exit codes for various scenarios.
func TestRun_ExitCodes(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create a valid but empty bin-dir for testing phase failures
	emptyBinDir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		expectExit0 bool
	}{
		{
			name:        "version_exits_0",
			args:        []string{"--version"},
			expectExit0: true,
		},
		{
			name:        "all_skipped_exits_0",
			args:        []string{"--skip=scope,static-analysis,ast,callgraph,dataflow,context"},
			expectExit0: true,
		},
		{
			name:        "missing_binaries_exits_nonzero",
			args:        []string{"--bin-dir=" + emptyBinDir, "--skip=context"},
			expectExit0: false,
		},
		{
			name:        "files_with_base_exits_nonzero",
			args:        []string{"--files=README.md", "--base=main"},
			expectExit0: false,
		},
		{
			name:        "files_from_with_head_exits_nonzero",
			args:        []string{"--files-from=README.md", "--head=HEAD"},
			expectExit0: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			args := append(tt.args, "--output="+tempDir)

			cmd := exec.Command(binaryPath, args...)
			err := cmd.Run()

			if tt.expectExit0 {
				if err != nil {
					t.Errorf("Expected exit code 0, but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("Expected non-zero exit code, but got 0")
				}
			}
		})
	}
}

// TestGetPhases verifies the phases are configured correctly.
func TestGetPhases(t *testing.T) {
	phases := getPhases()

	if len(phases) != 6 {
		t.Errorf("Expected 6 phases, got %d", len(phases))
	}

	// Verify phase order
	expectedOrder := []string{"scope", "static-analysis", "ast", "callgraph", "dataflow", "context"}
	for i, phase := range phases {
		if phase.Name != expectedOrder[i] {
			t.Errorf("Phase %d: expected %s, got %s", i, expectedOrder[i], phase.Name)
		}
	}

	// Verify each phase has required fields
	for _, phase := range phases {
		if phase.Name == "" {
			t.Error("Phase has empty name")
		}
		if phase.Binary == "" {
			t.Error("Phase has empty binary name")
		}
		if phase.Timeout == 0 {
			t.Errorf("Phase %s has zero timeout", phase.Name)
		}
		if phase.Args == nil {
			t.Errorf("Phase %s has nil Args function", phase.Name)
		}
	}
}

// TestPhaseArgs verifies phase argument generation.
func TestPhaseArgs(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config{
		baseRef:   "main",
		headRef:   "feature-branch",
		outputDir: filepath.Join(tempDir, "output"),
	}

	phases := getPhases()

	// Test scope phase args
	scopePhase := phases[0]
	if scopePhase.Name != "scope" {
		t.Fatal("First phase should be scope")
	}
	scopeArgs := scopePhase.Args(cfg)
	if !containsArg(scopeArgs, "--base") || !containsArg(scopeArgs, "--head") {
		t.Error("Scope phase should have --base and --head args")
	}
	if !containsArg(scopeArgs, "--output") {
		t.Error("Scope phase should have --output arg")
	}

	// Test context phase args
	contextPhase := phases[5]
	if contextPhase.Name != "context" {
		t.Fatal("Last phase should be context")
	}
	contextArgs := contextPhase.Args(cfg)
	if !containsArg(contextArgs, "--input") || !containsArg(contextArgs, "--output") {
		t.Error("Context phase should have --input and --output args")
	}
}

// ============================================================================
// Skip Condition & Git Extraction Tests
// ============================================================================

func TestShouldSkipForNoFiles_MissingScope(t *testing.T) {
	cfg := &config{outputDir: t.TempDir()}

	skip, reason := shouldSkipForNoFiles(cfg)
	if !skip {
		t.Fatal("Expected skip when scope.json is missing")
	}
	if !strings.Contains(reason, "scope.json") {
		t.Fatalf("Expected scope.json reason, got %q", reason)
	}
}

func TestShouldSkipForNoFiles_EmptyScope(t *testing.T) {
	tempDir := t.TempDir()
	writeScopeJSON(t, tempDir, "go", nil, nil, nil)
	cfg := &config{outputDir: tempDir}

	skip, reason := shouldSkipForNoFiles(cfg)
	if !skip {
		t.Fatal("Expected skip when no files are present")
	}
	if reason != "No changed files detected" {
		t.Fatalf("Unexpected skip reason: %q", reason)
	}
}

func TestShouldSkipForUnknownLanguage(t *testing.T) {
	tempDir := t.TempDir()
	writeScopeJSON(t, tempDir, "unknown", []string{"main.go"}, nil, nil)
	cfg := &config{outputDir: tempDir}

	skip, reason := shouldSkipForUnknownLanguage(cfg)
	if !skip {
		t.Fatal("Expected skip for unknown language")
	}
	if reason != "Unknown language detected" {
		t.Fatalf("Unexpected skip reason: %q", reason)
	}
}

func TestShouldSkipForMissingScope(t *testing.T) {
	cfg := &config{outputDir: t.TempDir()}

	skip, reason := shouldSkipForMissingScope(cfg)
	if !skip {
		t.Fatal("Expected skip when scope.json is missing")
	}
	if !strings.Contains(reason, "scope.json") {
		t.Fatalf("Expected scope.json reason, got %q", reason)
	}
}

func TestShouldSkipForNoFilesOrUnknownLanguage_UsesNoFilesReason(t *testing.T) {
	tempDir := t.TempDir()
	writeScopeJSON(t, tempDir, "unknown", nil, nil, nil)
	cfg := &config{outputDir: tempDir}

	skip, reason := shouldSkipForNoFilesOrUnknownLanguage(cfg)
	if !skip {
		t.Fatal("Expected skip when no files are present")
	}
	if reason != "No changed files detected" {
		t.Fatalf("Unexpected skip reason: %q", reason)
	}
}

func TestShouldSkipForNoFilesOrUnknownLanguage_UnknownLanguage(t *testing.T) {
	tempDir := t.TempDir()
	writeScopeJSON(t, tempDir, "unknown", []string{"main.go"}, nil, nil)
	cfg := &config{outputDir: tempDir}

	skip, reason := shouldSkipForNoFilesOrUnknownLanguage(cfg)
	if !skip {
		t.Fatal("Expected skip when language is unknown")
	}
	if reason != "Unknown language detected" {
		t.Fatalf("Unexpected skip reason: %q", reason)
	}
}

func TestExtractFileFromGit_Success(t *testing.T) {
	tempDir := t.TempDir()
	path, err := extractFileFromGit("HEAD", "scripts/codereview/cmd/run-all/main.go", tempDir)
	if err != nil {
		t.Fatalf("Expected success extracting file, got error: %v", err)
	}
	if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(tempDir)+string(filepath.Separator)) {
		t.Fatalf("Expected file in temp dir, got %s", path)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("Expected extracted file to exist: %v", statErr)
	}
}

func TestExtractFileFromGit_InvalidPaths(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		filePath  string
		errSubstr string
	}{
		{
			name:      "empty_path",
			filePath:  "",
			errSubstr: "invalid file path: empty",
		},
		{
			name:      "absolute_path",
			filePath:  filepath.Join(tempDir, "file.go"),
			errSubstr: "absolute paths are not allowed",
		},
		{
			name:      "traversal_path",
			filePath:  "../README.md",
			errSubstr: "contains path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractFileFromGit("HEAD", tt.filePath, tempDir)
			if err == nil {
				t.Fatalf("Expected error for %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Fatalf("Expected error containing %q, got %v", tt.errSubstr, err)
			}
		})
	}
}

func TestExtractFileFromGit_GitShowFailure(t *testing.T) {
	tempDir := t.TempDir()

	_, err := extractFileFromGit("HEAD", "no-such-file.txt", tempDir)
	if err == nil {
		t.Fatal("Expected error for missing file in git")
	}
	if !strings.Contains(err.Error(), "git show failed") {
		t.Fatalf("Expected git show failure, got %v", err)
	}
}

// containsArg checks if an argument is in the args slice.
func containsArg(args []string, arg string) bool {
	for _, a := range args {
		if a == arg {
			return true
		}
	}
	return false
}

// ============================================================================
// Helper Functions for Tests
// ============================================================================

func writeScopeJSON(t *testing.T, dir, language string, modified, added, deleted []string) {
	t.Helper()

	if modified == nil {
		modified = []string{}
	}
	if added == nil {
		added = []string{}
	}
	if deleted == nil {
		deleted = []string{}
	}

	scopeData := map[string]interface{}{
		"base_ref":  "main",
		"head_ref":  "HEAD",
		"language":  language,
		"languages": []string{language},
		"files": map[string]interface{}{
			"modified": modified,
			"added":    added,
			"deleted":  deleted,
		},
	}
	writeMockJSON(t, filepath.Join(dir, "scope.json"), scopeData)
}

// createMockPhaseOutputs creates mock phase output files in the given directory.
func createMockPhaseOutputs(t *testing.T, dir string) {
	t.Helper()

	// Create scope.json
	scopeData := map[string]interface{}{
		"base_ref":          "main",
		"head_ref":          "HEAD",
		"language":          "go",
		"languages":         []string{"go"},
		"files":             map[string]interface{}{"modified": []string{"main.go"}, "added": []string{}, "deleted": []string{}},
		"stats":             map[string]interface{}{"total_files": 1, "total_additions": 10, "total_deletions": 5},
		"packages_affected": []string{"main"},
	}
	writeMockJSON(t, filepath.Join(dir, "scope.json"), scopeData)

	// Create static-analysis.json
	staticData := map[string]interface{}{
		"findings": []map[string]interface{}{},
		"summary":  map[string]interface{}{"total": 0},
	}
	writeMockJSON(t, filepath.Join(dir, "static-analysis.json"), staticData)

	// Create go-ast.json
	astData := map[string]interface{}{
		"functions": map[string]interface{}{
			"modified": []map[string]interface{}{},
			"added":    []map[string]interface{}{},
			"deleted":  []map[string]interface{}{},
		},
		"types": map[string]interface{}{
			"modified": []map[string]interface{}{},
			"added":    []map[string]interface{}{},
		},
	}
	writeMockJSON(t, filepath.Join(dir, "go-ast.json"), astData)

	// Create go-calls.json
	callsData := map[string]interface{}{
		"modified_functions": []map[string]interface{}{},
		"impact_summary":     map[string]interface{}{},
	}
	writeMockJSON(t, filepath.Join(dir, "go-calls.json"), callsData)

	// Create go-flow.json
	flowData := map[string]interface{}{
		"flows":       []map[string]interface{}{},
		"nil_sources": []map[string]interface{}{},
	}
	writeMockJSON(t, filepath.Join(dir, "go-flow.json"), flowData)
}

// writeMockJSON writes a JSON file with the given data.
func writeMockJSON(t *testing.T, path string, data interface{}) {
	t.Helper()

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}
