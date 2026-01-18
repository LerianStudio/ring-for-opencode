package ast

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathValidator(t *testing.T) {
	baseDir := t.TempDir()
	nestedDir := filepath.Join(baseDir, "nested")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	validator, err := NewPathValidator(baseDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	validPath := filepath.Join(nestedDir, "file.go")
	if err := os.WriteFile(validPath, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	validated, err := validator.ValidatePath(validPath)
	if err != nil {
		t.Fatalf("expected valid path, got error: %v", err)
	}
	resolvedValid, err := filepath.EvalSymlinks(validPath)
	if err != nil {
		t.Fatalf("failed to resolve path: %v", err)
	}
	if filepath.Clean(validated) != filepath.Clean(resolvedValid) {
		t.Errorf("expected validated path %s, got %s", resolvedValid, validated)
	}

	outsidePath := filepath.Join(os.TempDir(), "outside.go")
	if _, err := validator.ValidatePath(outsidePath); err == nil {
		t.Fatalf("expected error for outside path")
	}

	if _, err := validator.ValidatePath(""); err != nil {
		t.Fatalf("expected empty path to be allowed: %v", err)
	}
}

func TestPathValidator_NonexistentParent(t *testing.T) {
	baseDir := t.TempDir()
	validator, err := NewPathValidator(baseDir)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	nonexistent := filepath.Join(baseDir, "missing", "file.go")
	if _, err := validator.ValidatePath(nonexistent); err == nil {
		t.Fatalf("expected error for missing parent directory")
	}
}
