// Package main provides integration tests for the scope-detector CLI binary.
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
const testBinaryName = "scope-detector-test"

// buildTestBinary builds the scope-detector binary for testing.
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

// isGitRepo checks if the current directory is inside a git repository.
func isGitRepo(t *testing.T) bool {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// hasGitHistory checks if HEAD~1 exists (has at least 2 commits).
func hasGitHistory(t *testing.T) bool {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD~1")
	return cmd.Run() == nil
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
	if !strings.HasPrefix(outputStr, "scope-detector version ") {
		t.Errorf("Expected version output to start with 'scope-detector version ', got: %s", outputStr)
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
		"scope-detector",
		"-base",
		"-head",
		"-output",
		"-workdir",
		"-version",
		"Options:",
		"Examples:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Help output missing expected string %q\nOutput: %s", expected, outputStr)
		}
	}
}

// TestMain_OutputToFile verifies that --output flag writes JSON to file.
func TestMain_OutputToFile(t *testing.T) {
	if !isGitRepo(t) {
		t.Skip("Skipping: not in a git repository")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "scope.json")

	// Run the binary with --output flag
	// Use current directory as workdir to ensure we're in a git repo
	cmd := exec.Command(binaryPath, "--output="+outputPath)

	// Use separate buffers for stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run with --output: %v\nStderr: %s", err, stderrBuf.String())
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created at %s", outputPath)
	}

	// Read and verify the file contains valid JSON
	fileContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify it's valid JSON by attempting to unmarshal
	var jsonData map[string]interface{}
	if err := json.Unmarshal(fileContent, &jsonData); err != nil {
		t.Errorf("Output file contains invalid JSON: %v\nContent: %s", err, string(fileContent))
	}

	// Verify stderr mentions the output file
	if !strings.Contains(stderrBuf.String(), "Scope written to") {
		t.Errorf("Expected stderr to mention output file, got: %s", stderrBuf.String())
	}
}

// TestMain_OutputToNestedPath verifies that --output creates parent directories.
func TestMain_OutputToNestedPath(t *testing.T) {
	if !isGitRepo(t) {
		t.Skip("Skipping: not in a git repository")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory with nested path
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "nested", "deeply", "scope.json")

	// Run the binary with nested output path
	cmd := exec.Command(binaryPath, "--output="+nestedPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run with nested --output: %v\nOutput: %s", err, string(output))
	}

	// Verify file was created
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created at nested path %s", nestedPath)
	}

	// Verify parent directories were created
	parentDir := filepath.Dir(nestedPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Errorf("Parent directories were not created: %s", parentDir)
	}
}

// scopeOutputJSON represents the expected JSON output structure for testing.
type scopeOutputJSON struct {
	BaseRef  string `json:"base_ref"`
	HeadRef  string `json:"head_ref"`
	Language string `json:"language"`
	Files    struct {
		Modified []string `json:"modified"`
		Added    []string `json:"added"`
		Deleted  []string `json:"deleted"`
	} `json:"files"`
	Stats struct {
		TotalFiles     int `json:"total_files"`
		TotalAdditions int `json:"total_additions"`
		TotalDeletions int `json:"total_deletions"`
	} `json:"stats"`
	Packages []string `json:"packages_affected"`
}

// TestMain_JSONStructure verifies that JSON output has the correct structure.
func TestMain_JSONStructure(t *testing.T) {
	if !isGitRepo(t) {
		t.Skip("Skipping: not in a git repository")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "scope.json")

	// Run the binary
	cmd := exec.Command(binaryPath, "--output="+outputPath)
	if _, err := cmd.CombinedOutput(); err != nil {
		// Even if there are no changes, should produce valid JSON
		t.Logf("Warning: command returned error (may have no changes): %v", err)
	}

	// Read output file
	fileContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Unmarshal into our expected structure
	var output scopeOutputJSON
	if err := json.Unmarshal(fileContent, &output); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v\nContent: %s", err, string(fileContent))
	}

	// Verify all required fields exist by checking the raw JSON
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(fileContent, &rawJSON); err != nil {
		t.Fatalf("Failed to unmarshal raw JSON: %v", err)
	}

	// Check top-level required fields
	requiredTopLevel := []string{"base_ref", "head_ref", "language", "files", "stats", "packages_affected"}
	for _, field := range requiredTopLevel {
		if _, exists := rawJSON[field]; !exists {
			t.Errorf("Missing required top-level field: %s", field)
		}
	}

	// Check files sub-structure
	if files, ok := rawJSON["files"].(map[string]interface{}); ok {
		requiredFiles := []string{"modified", "added", "deleted"}
		for _, field := range requiredFiles {
			if _, exists := files[field]; !exists {
				t.Errorf("Missing required files field: %s", field)
			}
		}
	} else {
		t.Error("files field is not an object")
	}

	// Check stats sub-structure
	if stats, ok := rawJSON["stats"].(map[string]interface{}); ok {
		requiredStats := []string{"total_files", "total_additions", "total_deletions"}
		for _, field := range requiredStats {
			if _, exists := stats[field]; !exists {
				t.Errorf("Missing required stats field: %s", field)
			}
		}
	} else {
		t.Error("stats field is not an object")
	}

	// Verify files arrays are actual arrays (not null)
	if output.Files.Modified == nil {
		t.Error("files.modified should be an array, not null")
	}
	if output.Files.Added == nil {
		t.Error("files.added should be an array, not null")
	}
	if output.Files.Deleted == nil {
		t.Error("files.deleted should be an array, not null")
	}
	if output.Packages == nil {
		t.Error("packages_affected should be an array, not null")
	}
}

// TestMain_StdoutOutput verifies that output goes to stdout when --output is not specified.
func TestMain_StdoutOutput(t *testing.T) {
	if !isGitRepo(t) {
		t.Skip("Skipping: not in a git repository")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Run without --output flag
	cmd := exec.Command(binaryPath)
	stdout, err := cmd.Output()
	if err != nil {
		// May have no changes, but should still produce JSON
		t.Logf("Warning: command returned error: %v", err)
	}

	// Verify stdout contains valid JSON
	if len(stdout) > 0 {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(stdout, &jsonData); err != nil {
			t.Errorf("stdout does not contain valid JSON: %v\nOutput: %s", err, string(stdout))
		}
	}
}

// TestMain_BaseHeadRefs verifies that --base and --head flags work correctly.
func TestMain_BaseHeadRefs(t *testing.T) {
	if !isGitRepo(t) {
		t.Skip("Skipping: not in a git repository")
	}
	if !hasGitHistory(t) {
		t.Skip("Skipping: repository does not have enough commit history (need HEAD~1)")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "scope.json")

	// Run with --base and --head flags
	cmd := exec.Command(binaryPath, "--base=HEAD~1", "--head=HEAD", "--output="+outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run with refs: %v\nOutput: %s", err, string(output))
	}

	// Read and verify output
	fileContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var scopeOutput scopeOutputJSON
	if err := json.Unmarshal(fileContent, &scopeOutput); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify refs are populated
	if scopeOutput.BaseRef == "" {
		t.Error("base_ref should not be empty when --base is specified")
	}
	if scopeOutput.HeadRef == "" {
		t.Error("head_ref should not be empty when --head is specified")
	}
}

// TestMain_WorkdirFlag verifies that --workdir flag changes the working directory.
func TestMain_WorkdirFlag(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "scope.json")

	// Get the repository root (assuming we're in a git repo)
	repoRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skip("Skipping: cannot determine git repository root")
	}

	// Run with --workdir pointing to repo root
	cmd := exec.Command(binaryPath, "--workdir="+strings.TrimSpace(string(repoRoot)), "--output="+outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run with --workdir: %v\nOutput: %s", err, string(output))
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}
}

// TestMain_InvalidWorkdir verifies that invalid --workdir produces an error.
func TestMain_InvalidWorkdir(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Run with non-existent workdir
	cmd := exec.Command(binaryPath, "--workdir=/nonexistent/path/that/should/not/exist")
	output, err := cmd.CombinedOutput()

	// Should fail with a non-zero exit code
	if err == nil {
		t.Fatal("Expected error for invalid workdir, but command succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected *exec.ExitError, got %T: %v", err, err)
	}

	if exitErr.ExitCode() == 0 {
		t.Errorf("Expected non-zero exit code, got 0")
	}

	// Error message should mention the invalid path to validate the failure reason
	outputStr := string(output)
	if !strings.Contains(outputStr, "/nonexistent/path") {
		t.Errorf("Expected error output to contain the invalid path '/nonexistent/path', got: %s", outputStr)
	}
}

// TestMain_InvalidRef verifies that invalid git ref produces an error.
func TestMain_InvalidRef(t *testing.T) {
	if !isGitRepo(t) {
		t.Skip("Skipping: not in a git repository")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Run with invalid ref
	cmd := exec.Command(binaryPath, "--base=invalid-nonexistent-ref-xyz123")
	output, err := cmd.CombinedOutput()

	// Should fail
	if err == nil {
		t.Error("Expected error for invalid ref, but command succeeded")
	}

	// Should produce error output
	if len(output) == 0 {
		t.Error("Expected some error output for invalid ref")
	}
}

// TestMain_ExitCodes verifies correct exit codes for various scenarios.
func TestMain_ExitCodes(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	tests := []struct {
		name         string
		args         []string
		expectExit0  bool
		skipIfNoRepo bool
	}{
		{
			name:        "version_exits_0",
			args:        []string{"--version"},
			expectExit0: true,
		},
		{
			name:         "valid_run_exits_0",
			args:         []string{},
			expectExit0:  true,
			skipIfNoRepo: true,
		},
		{
			name:        "invalid_workdir_exits_nonzero",
			args:        []string{"--workdir=/nonexistent"},
			expectExit0: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipIfNoRepo && !isGitRepo(t) {
				t.Skip("Skipping: not in a git repository")
			}

			cmd := exec.Command(binaryPath, tt.args...)
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

// TestMain_JSONPrettyPrint verifies that JSON output is properly formatted.
func TestMain_JSONPrettyPrint(t *testing.T) {
	if !isGitRepo(t) {
		t.Skip("Skipping: not in a git repository")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "scope.json")

	// Run the binary
	cmd := exec.Command(binaryPath, "--output="+outputPath)
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Logf("Warning: command returned error: %v", err)
	}

	// Read output file
	fileContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify JSON is pretty-printed (contains newlines and indentation)
	contentStr := string(fileContent)
	if !strings.Contains(contentStr, "\n") {
		t.Error("JSON output should be pretty-printed with newlines")
	}
	if !strings.Contains(contentStr, "  ") {
		t.Error("JSON output should be pretty-printed with indentation")
	}
}

// TestMain_ConsistentLanguageDetection verifies language detection consistency.
func TestMain_ConsistentLanguageDetection(t *testing.T) {
	if !isGitRepo(t) {
		t.Skip("Skipping: not in a git repository")
	}
	if !hasGitHistory(t) {
		t.Skip("Skipping: repository does not have enough commit history")
	}

	binaryPath := buildTestBinary(t)
	defer cleanupTestBinary(t, binaryPath)

	// Run twice with same parameters
	tempDir := t.TempDir()

	var results []scopeOutputJSON
	for i := 0; i < 2; i++ {
		outputPath := filepath.Join(tempDir, "scope"+string(rune('0'+i))+".json")

		cmd := exec.Command(binaryPath, "--base=HEAD~1", "--head=HEAD", "--output="+outputPath)
		if _, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Run %d failed: %v", i, err)
		}

		fileContent, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		var output scopeOutputJSON
		if err := json.Unmarshal(fileContent, &output); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		results = append(results, output)
	}

	// Verify language detection is consistent
	if results[0].Language != results[1].Language {
		t.Errorf("Language detection inconsistent: %s vs %s", results[0].Language, results[1].Language)
	}
}
