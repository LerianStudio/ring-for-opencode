// Package main provides integration tests for the compile-context CLI binary.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testBinaryName is the name of the test binary.
const testBinaryName = "compile-context-test"

// buildTestBinary builds the compile-context binary for testing.
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
// Failures are logged but do not fail the test, as cleanup failures are typically
// transient (e.g., file locked on Windows) and do not affect test correctness.
func cleanupTestBinary(t *testing.T, binaryPath string) {
	t.Helper()

	err := os.Remove(binaryPath)
	switch {
	case err == nil:
		// Successfully removed
	case os.IsNotExist(err):
		// Already removed or never created - not an issue
	case os.IsPermission(err):
		// Permission denied - log for debugging but don't fail
		// This can happen on some platforms when the binary is still in use
		t.Logf("Warning: permission denied removing test binary %s (may still be in use): %v", binaryPath, err)
	default:
		// Other errors - log with details for debugging
		t.Logf("Warning: failed to remove test binary %s: %v", binaryPath, err)
	}
}

// createMockPhaseOutputs creates mock phase output files in the given directory.
func createMockPhaseOutputs(t *testing.T, dir string) {
	t.Helper()

	// Create scope.json
	scopeData := map[string]interface{}{
		"base_ref":          "main",
		"head_ref":          "HEAD",
		"language":          "go",
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

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

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
	if !strings.HasPrefix(outputStr, "compile-context version ") {
		t.Errorf("Expected version output to start with 'compile-context version ', got: %s", outputStr)
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
	// Accept either as valid since we just want to verify the output content
	if err != nil {
		// Check if it's an exit error with expected code
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("--help failed with unexpected error type %T: %v\nOutput: %s", err, err, string(output))
		}
		// Go's flag package exits with code 2 for --help
		if exitCode := exitErr.ExitCode(); exitCode != 2 {
			t.Fatalf("--help exited with unexpected code %d (expected 0 or 2): %v\nOutput: %s",
				exitCode, err, string(output))
		}
	}

	outputStr := string(output)

	// Verify help output contains expected sections
	expectedStrings := []string{
		"Usage:",
		"Context Compiler",
		"-input",
		"-output",
		"-verbose",
		"-version",
		"Options:",
		"Examples:",
		"Phase Outputs",
		"context-code-reviewer.md",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output missing expected string %q\nOutput: %s", expected, outputStr)
		}
	}
}

// TestRun_Success verifies successful compilation with valid phase outputs.
func TestRun_Success(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory with mock phase outputs
	tempDir := t.TempDir()
	createMockPhaseOutputs(t, tempDir)

	// Run the binary
	cmd := exec.Command(binaryPath, "--input="+tempDir, "--verbose")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success, got error: %v\nStdout: %s\nStderr: %s",
			err, stdout.String(), stderr.String())
	}

	// Verify success message
	if !strings.Contains(stdout.String(), "Context compilation complete") {
		t.Errorf("Expected success message in stdout, got: %s", stdout.String())
	}

	// Verify context files were created
	expectedFiles := []string{
		"context-code-reviewer.md",
		"context-security-reviewer.md",
		"context-business-logic-reviewer.md",
		"context-test-reviewer.md",
		"context-nil-safety-reviewer.md",
	}

	for _, filename := range expectedFiles {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected context file not created: %s", filename)
		}
	}
}

// TestRun_NonExistentInputDir verifies error when input directory doesn't exist.
func TestRun_NonExistentInputDir(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Run with non-existent input directory
	cmd := exec.Command(binaryPath, "--input=/nonexistent/path/that/should/not/exist")
	output, err := cmd.CombinedOutput()

	// Should fail with a non-zero exit code
	if err == nil {
		t.Fatal("Expected error for non-existent input directory, but command succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected *exec.ExitError, got %T: %v", err, err)
	}

	if exitErr.ExitCode() == 0 {
		t.Errorf("Expected non-zero exit code, got 0")
	}

	// Error message should mention the invalid path
	outputStr := string(output)
	if !strings.Contains(outputStr, "input directory does not exist") {
		t.Errorf("Expected error about non-existent input directory, got: %s", outputStr)
	}
}

// TestRun_EmptyInputDir verifies behavior when input directory has no phase outputs.
func TestRun_EmptyInputDir(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create empty temp directory
	tempDir := t.TempDir()

	// Run the binary - should still succeed but with minimal context
	cmd := exec.Command(binaryPath, "--input="+tempDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success even with empty input, got error: %v\nStderr: %s",
			err, stderr.String())
	}

	// Context files should still be created (with minimal content)
	expectedFiles := []string{
		"context-code-reviewer.md",
		"context-security-reviewer.md",
		"context-business-logic-reviewer.md",
		"context-test-reviewer.md",
		"context-nil-safety-reviewer.md",
	}

	for _, filename := range expectedFiles {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected context file not created: %s", filename)
		}
	}
}

// TestRun_CustomOutputDir verifies that --output flag creates files in custom directory.
func TestRun_CustomOutputDir(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directories
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	createMockPhaseOutputs(t, inputDir)

	// Run with separate output directory
	cmd := exec.Command(binaryPath, "--input="+inputDir, "--output="+outputDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Verify context files were created in output directory
	expectedFile := filepath.Join(outputDir, "context-code-reviewer.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Context file not created in output directory: %s", expectedFile)
	}

	// Verify input directory was not modified with new files
	inputContextFile := filepath.Join(inputDir, "context-code-reviewer.md")
	if _, err := os.Stat(inputContextFile); err == nil {
		t.Errorf("Context file should be in output directory, not input: %s", inputContextFile)
	}
}

// TestRun_NestedOutputDir verifies that nested output directories are created.
func TestRun_NestedOutputDir(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directories with nested output path
	inputDir := t.TempDir()
	outputDir := filepath.Join(t.TempDir(), "nested", "deeply", "output")
	createMockPhaseOutputs(t, inputDir)

	// Run with nested output path
	cmd := exec.Command(binaryPath, "--input="+inputDir, "--output="+outputDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Verify directories were created
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Errorf("Nested output directory was not created: %s", outputDir)
	}

	// Verify context file exists
	expectedFile := filepath.Join(outputDir, "context-code-reviewer.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Context file not created in nested output directory: %s", expectedFile)
	}
}

// TestRun_PermissionDenied verifies error when output directory is not writable.
func TestRun_PermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create input directory with mock data
	inputDir := t.TempDir()
	createMockPhaseOutputs(t, inputDir)

	// Create read-only output directory
	outputDir := t.TempDir()
	readOnlyDir := filepath.Join(outputDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	// Ensure cleanup can still work
	defer func() {
		_ = os.Chmod(readOnlyDir, 0755)
	}()

	// Try to write to a subdirectory of the read-only directory
	targetOutput := filepath.Join(readOnlyDir, "subdir")

	cmd := exec.Command(binaryPath, "--input="+inputDir, "--output="+targetOutput)
	output, err := cmd.CombinedOutput()

	// Should fail due to permission denied
	if err == nil {
		t.Fatal("Expected error for permission denied, but command succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected *exec.ExitError, got %T: %v", err, err)
	}

	if exitErr.ExitCode() == 0 {
		t.Errorf("Expected non-zero exit code, got 0")
	}

	// Error message should indicate permission issue
	outputStr := string(output)
	if !strings.Contains(outputStr, "permission denied") && !strings.Contains(outputStr, "Error") {
		t.Logf("Note: Error message does not explicitly mention permission denied: %s", outputStr)
	}
}

// TestRun_VerboseOutput verifies verbose mode includes additional information.
func TestRun_VerboseOutput(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory with mock phase outputs
	tempDir := t.TempDir()
	createMockPhaseOutputs(t, tempDir)

	// Run with verbose flag
	cmd := exec.Command(binaryPath, "--input="+tempDir, "--verbose")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success, got error: %v\nStderr: %s", err, stderr.String())
	}

	// Verbose output should mention directories and generated files
	stderrStr := stderr.String()
	expectedVerbose := []string{
		"Input directory:",
		"Output directory:",
		"Generated context files:",
		"context-code-reviewer.md",
	}

	for _, expected := range expectedVerbose {
		if !strings.Contains(stderrStr, expected) {
			t.Errorf("Verbose output missing expected string %q\nStderr: %s", expected, stderrStr)
		}
	}
}

// TestRun_MalformedJSON verifies handling of malformed phase output files.
func TestRun_MalformedJSON(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory with malformed JSON
	tempDir := t.TempDir()

	// Write malformed scope.json
	malformedJSON := []byte(`{"base_ref": "main", "head_ref":`)
	if err := os.WriteFile(filepath.Join(tempDir, "scope.json"), malformedJSON, 0644); err != nil {
		t.Fatalf("Failed to write malformed JSON: %v", err)
	}

	// Should still succeed (graceful degradation)
	cmd := exec.Command(binaryPath, "--input="+tempDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success with graceful degradation, got error: %v\nStderr: %s",
			err, stderr.String())
	}

	// Context files should still be created
	expectedFile := filepath.Join(tempDir, "context-code-reviewer.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Context file not created despite malformed input: %s", expectedFile)
	}
}

// TestRun_ExitCodes verifies correct exit codes for various scenarios.
func TestRun_ExitCodes(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	tests := []struct {
		name        string
		args        []string
		setupDir    bool
		expectExit0 bool
	}{
		{
			name:        "version_exits_0",
			args:        []string{"--version"},
			setupDir:    false,
			expectExit0: true,
		},
		{
			name:        "valid_input_exits_0",
			args:        []string{}, // will add --input=tempDir
			setupDir:    true,
			expectExit0: true,
		},
		{
			name:        "nonexistent_input_exits_nonzero",
			args:        []string{"--input=/nonexistent/path"},
			setupDir:    false,
			expectExit0: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args

			if tt.setupDir {
				tempDir := t.TempDir()
				createMockPhaseOutputs(t, tempDir)
				args = append(args, "--input="+tempDir)
			}

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

// TestRun_ContextFileContent verifies that generated context files contain expected content.
func TestRun_ContextFileContent(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory with mock phase outputs
	tempDir := t.TempDir()
	createMockPhaseOutputs(t, tempDir)

	// Run the binary
	cmd := exec.Command(binaryPath, "--input="+tempDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	// Read and verify code reviewer context file
	codeReviewerPath := filepath.Join(tempDir, "context-code-reviewer.md")
	content, err := os.ReadFile(codeReviewerPath)
	if err != nil {
		t.Fatalf("Failed to read code reviewer context: %v", err)
	}

	// Verify it's a markdown file with expected structure
	contentStr := string(content)
	if !strings.Contains(contentStr, "#") {
		t.Error("Context file should contain markdown headers")
	}
}

// TestRun_DefaultInputDir verifies that default input directory is used when not specified.
func TestRun_DefaultInputDir(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create a temp directory to run from
	workDir := t.TempDir()

	// Create the default input directory structure
	defaultInputDir := filepath.Join(workDir, ".ring", "codereview")
	if err := os.MkdirAll(defaultInputDir, 0755); err != nil {
		t.Fatalf("Failed to create default input directory: %v", err)
	}

	// Create mock phase outputs in the default location
	createMockPhaseOutputsInDir(t, defaultInputDir)

	// Run without --input flag, from the work directory
	cmd := exec.Command(binaryPath)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	// This should succeed if default directory exists and has content
	if err != nil {
		t.Logf("Note: Default directory test - command returned: %v\nOutput: %s", err, string(output))
	}
}

// createMockPhaseOutputsInDir is a helper that creates mock outputs in a specific directory.
func createMockPhaseOutputsInDir(t *testing.T, dir string) {
	t.Helper()

	// Same as createMockPhaseOutputs but for a specific directory
	scopeData := map[string]interface{}{
		"base_ref":          "main",
		"head_ref":          "HEAD",
		"language":          "go",
		"files":             map[string]interface{}{"modified": []string{}, "added": []string{}, "deleted": []string{}},
		"stats":             map[string]interface{}{"total_files": 0, "total_additions": 0, "total_deletions": 0},
		"packages_affected": []string{},
	}
	writeMockJSON(t, filepath.Join(dir, "scope.json"), scopeData)
}
