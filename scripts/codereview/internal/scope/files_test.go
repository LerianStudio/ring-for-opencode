package scope

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExpandFilePatterns(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping glob tests on Windows path handling")
	}

	baseDir := t.TempDir()
	mustWriteFile(t, filepath.Join(baseDir, "cmd", "app", "main.go"))
	mustWriteFile(t, filepath.Join(baseDir, "cmd", "app", "util.go"))
	mustWriteFile(t, filepath.Join(baseDir, "web", "app.ts"))
	mustWriteFile(t, filepath.Join(baseDir, "README.md"))

	files, err := ExpandFilePatterns(baseDir, []string{"cmd/**/*.go", "web/*.ts"})
	if err != nil {
		t.Fatalf("ExpandFilePatterns() error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(files), files)
	}
}

func TestExpandFilePatterns_NoMatches(t *testing.T) {
	baseDir := t.TempDir()
	files, err := ExpandFilePatterns(baseDir, []string{"missing/*.go"})
	if err != nil {
		t.Fatalf("unexpected error for unmatched pattern: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty result for unmatched pattern, got %v", files)
	}
}

func TestExpandFilePatterns_PathTraversal(t *testing.T) {
	baseDir := t.TempDir()
	_, err := ExpandFilePatterns(baseDir, []string{"../*.go"})
	if err == nil {
		t.Fatal("expected error for traversal pattern")
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}
