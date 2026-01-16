package ast

import (
	"context"
	"path/filepath"
	"testing"
)

func TestGoExtractor_ExtractDiff(t *testing.T) {
	extractor := NewGoExtractor()

	beforePath := filepath.Join("..", "..", "testdata", "go", "before.go")
	afterPath := filepath.Join("..", "..", "testdata", "go", "after.go")

	diff, err := extractor.ExtractDiff(context.Background(), beforePath, afterPath)
	if err != nil {
		t.Fatalf("ExtractDiff failed: %v", err)
	}

	// Verify language
	if diff.Language != "go" {
		t.Errorf("expected language 'go', got '%s'", diff.Language)
	}

	// Verify function changes
	funcChanges := make(map[string]ChangeType)
	for _, f := range diff.Functions {
		funcChanges[f.Name] = f.ChangeType
	}

	// Hello should be modified (signature changed)
	if ct, ok := funcChanges["Hello"]; !ok || ct != ChangeModified {
		t.Errorf("expected Hello to be modified, got %v", funcChanges["Hello"])
	}

	// FormatName should be removed
	if ct, ok := funcChanges["FormatName"]; !ok || ct != ChangeRemoved {
		t.Errorf("expected FormatName to be removed, got %v", funcChanges["FormatName"])
	}

	// NewGreeting should be added
	if ct, ok := funcChanges["NewGreeting"]; !ok || ct != ChangeAdded {
		t.Errorf("expected NewGreeting to be added, got %v", funcChanges["NewGreeting"])
	}

	// User.GetEmail should be added
	if ct, ok := funcChanges["*User.GetEmail"]; !ok || ct != ChangeAdded {
		t.Errorf("expected *User.GetEmail to be added, got %v", funcChanges["*User.GetEmail"])
	}

	// Verify type changes
	typeChanges := make(map[string]ChangeType)
	for _, ty := range diff.Types {
		typeChanges[ty.Name] = ty.ChangeType
	}

	// User should be modified (fields added)
	if ct, ok := typeChanges["User"]; !ok || ct != ChangeModified {
		t.Errorf("expected User to be modified, got %v", typeChanges["User"])
	}

	// Config should be added
	if ct, ok := typeChanges["Config"]; !ok || ct != ChangeAdded {
		t.Errorf("expected Config to be added, got %v", typeChanges["Config"])
	}

	// Verify import changes
	importChanges := make(map[string]ChangeType)
	for _, imp := range diff.Imports {
		importChanges[imp.Path] = imp.ChangeType
	}

	// strings should be removed
	if ct, ok := importChanges["strings"]; !ok || ct != ChangeRemoved {
		t.Errorf("expected 'strings' import to be removed, got %v", importChanges["strings"])
	}

	// context should be added
	if ct, ok := importChanges["context"]; !ok || ct != ChangeAdded {
		t.Errorf("expected 'context' import to be added, got %v", importChanges["context"])
	}

	// Verify summary
	if diff.Summary.FunctionsAdded < 2 {
		t.Errorf("expected at least 2 functions added, got %d", diff.Summary.FunctionsAdded)
	}
	if diff.Summary.FunctionsRemoved < 1 {
		t.Errorf("expected at least 1 function removed, got %d", diff.Summary.FunctionsRemoved)
	}
	if diff.Summary.TypesAdded < 1 {
		t.Errorf("expected at least 1 type added, got %d", diff.Summary.TypesAdded)
	}
}

func TestGoExtractor_NewFile(t *testing.T) {
	extractor := NewGoExtractor()

	afterPath := filepath.Join("..", "..", "testdata", "go", "after.go")

	diff, err := extractor.ExtractDiff(context.Background(), "", afterPath)
	if err != nil {
		t.Fatalf("ExtractDiff failed: %v", err)
	}

	// All functions should be added
	for _, f := range diff.Functions {
		if f.ChangeType != ChangeAdded {
			t.Errorf("expected function %s to be added, got %s", f.Name, f.ChangeType)
		}
	}
}

func TestGoExtractor_DeletedFile(t *testing.T) {
	extractor := NewGoExtractor()

	beforePath := filepath.Join("..", "..", "testdata", "go", "before.go")

	diff, err := extractor.ExtractDiff(context.Background(), beforePath, "")
	if err != nil {
		t.Fatalf("ExtractDiff failed: %v", err)
	}

	// All functions should be removed
	for _, f := range diff.Functions {
		if f.ChangeType != ChangeRemoved {
			t.Errorf("expected function %s to be removed, got %s", f.Name, f.ChangeType)
		}
	}
}

func TestGoExtractor_SupportedExtensions(t *testing.T) {
	extractor := NewGoExtractor()

	extensions := extractor.SupportedExtensions()
	if len(extensions) != 1 || extensions[0] != ".go" {
		t.Errorf("expected ['.go'], got %v", extensions)
	}
}

func TestGoExtractor_Language(t *testing.T) {
	extractor := NewGoExtractor()

	if extractor.Language() != "go" {
		t.Errorf("expected 'go', got '%s'", extractor.Language())
	}
}

func TestGoExtractor_ParseFile_InvalidPath(t *testing.T) {
	extractor := NewGoExtractor()

	_, err := extractor.ExtractDiff(context.Background(), "/nonexistent/file.go", "")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
