package testutil

import (
	"os"
	"testing"
)

// setupTestDir is the internal implementation so callers inside this package can
// refer to the helper using a lower-case name when desired.
func setupTestDir(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "callgraph-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to clean temp dir %s: %v", tmpDir, err)
		}
	})

	return tmpDir
}

// SetupTestDir creates a temp directory and registers cleanup via t.Cleanup.
// It returns the temp directory path for callers to use in tests.
func SetupTestDir(t *testing.T) string {
	return setupTestDir(t)
}
