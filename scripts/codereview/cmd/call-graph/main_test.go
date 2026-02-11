// Package main provides unit tests for the call-graph CLI wrapper.
package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeprecationWarning(t *testing.T) {
	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "call-graph")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join(mustGetWd(t), ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, out)
	}

	// Run with --help to get deprecation warning without requiring valid AST file
	runCmd := exec.Command(binaryPath, "--help")
	var stderr bytes.Buffer
	runCmd.Stderr = &stderr

	_ = runCmd.Run() // Ignore exit code, we just want stderr

	if !strings.Contains(stderr.String(), "DEPRECATED: call-graph is deprecated") {
		t.Errorf("Expected deprecation warning in stderr, got: %s", stderr.String())
	}
}

func TestFlagsAreAccepted(t *testing.T) {
	// Build the binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "call-graph")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join(mustGetWd(t), ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, out)
	}

	// Create a minimal AST file
	astFile := filepath.Join(tmpDir, "go-ast.json")
	if err := os.WriteFile(astFile, []byte(`[]`), 0o644); err != nil {
		t.Fatalf("Failed to create AST file: %v", err)
	}

	// Test that all old flags are accepted (even if deprecated)
	runCmd := exec.Command(binaryPath,
		"--ast", astFile,
		"--output", tmpDir,
		"--timeout", "60",
		"--lang", "go",
		"--languages-file", "",
		"--output-suffix", "",
		"-v",
	)
	var stderr bytes.Buffer
	runCmd.Stderr = &stderr

	if err := runCmd.Run(); err != nil {
		// The command might fail for other reasons, but flags should be accepted
		if strings.Contains(stderr.String(), "flag provided but not defined") {
			t.Errorf("Flag not accepted: %s", stderr.String())
		}
	}
}

func mustGetWd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	return wd
}
