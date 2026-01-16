//go:build integration

package ast

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIntegration_GoExtractor(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "testdata")
	beforePath := filepath.Join(testdataDir, "go", "before.go")
	afterPath := filepath.Join(testdataDir, "go", "after.go")

	// Skip if testdata doesn't exist
	if _, err := os.Stat(beforePath); os.IsNotExist(err) {
		t.Skipf("testdata file not found: %s", beforePath)
	}
	if _, err := os.Stat(afterPath); os.IsNotExist(err) {
		t.Skipf("testdata file not found: %s", afterPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	extractor := NewGoExtractor()
	diff, err := extractor.ExtractDiff(ctx, beforePath, afterPath)
	if err != nil {
		t.Fatalf("ExtractDiff failed: %v", err)
	}

	if diff.Error != "" {
		t.Fatalf("diff contains error: %s", diff.Error)
	}

	// Verify expected changes from testdata:
	// - Hello: signature changed (params and returns)
	// - FormatName: removed
	// - NewGreeting: added
	// - User.GetEmail: added
	// - Config: new type added
	// - User: modified (added Email and IsActive fields)

	// Check functions added (NewGreeting, User.GetEmail)
	if diff.Summary.FunctionsAdded < 2 {
		t.Errorf("expected at least 2 functions added, got %d", diff.Summary.FunctionsAdded)
	}

	// Check functions removed (FormatName)
	if diff.Summary.FunctionsRemoved < 1 {
		t.Errorf("expected at least 1 function removed, got %d", diff.Summary.FunctionsRemoved)
	}

	// Check functions modified (Hello)
	if diff.Summary.FunctionsModified < 1 {
		t.Errorf("expected at least 1 function modified, got %d", diff.Summary.FunctionsModified)
	}

	// Check types added (Config)
	if diff.Summary.TypesAdded < 1 {
		t.Errorf("expected at least 1 type added, got %d", diff.Summary.TypesAdded)
	}

	// Check types modified (User)
	if diff.Summary.TypesModified < 1 {
		t.Errorf("expected at least 1 type modified, got %d", diff.Summary.TypesModified)
	}

	// Verify specific function changes
	funcMap := make(map[string]FunctionDiff)
	for _, fn := range diff.Functions {
		funcMap[fn.Name] = fn
	}

	// FormatName should be removed
	if fn, ok := funcMap["FormatName"]; ok {
		if fn.ChangeType != ChangeRemoved {
			t.Errorf("FormatName should be removed, got %s", fn.ChangeType)
		}
	} else {
		t.Error("FormatName not found in function diffs")
	}

	// NewGreeting should be added
	if fn, ok := funcMap["NewGreeting"]; ok {
		if fn.ChangeType != ChangeAdded {
			t.Errorf("NewGreeting should be added, got %s", fn.ChangeType)
		}
	} else {
		t.Error("NewGreeting not found in function diffs")
	}

	// Hello should be modified
	if fn, ok := funcMap["Hello"]; ok {
		if fn.ChangeType != ChangeModified {
			t.Errorf("Hello should be modified, got %s", fn.ChangeType)
		}
	} else {
		t.Error("Hello not found in function diffs")
	}

	// *User.GetEmail should be added
	if fn, ok := funcMap["*User.GetEmail"]; ok {
		if fn.ChangeType != ChangeAdded {
			t.Errorf("*User.GetEmail should be added, got %s", fn.ChangeType)
		}
	} else {
		t.Error("*User.GetEmail not found in function diffs")
	}

	// Verify markdown rendering doesn't panic
	md := RenderMarkdown(diff)
	if md == "" {
		t.Error("markdown render returned empty string")
	}

	// Verify JSON rendering
	jsonBytes, err := RenderJSON(diff)
	if err != nil {
		t.Errorf("JSON render failed: %v", err)
	}
	if len(jsonBytes) == 0 {
		t.Error("JSON render returned empty bytes")
	}

	t.Logf("Summary: +%d/-%d/~%d functions, +%d/-%d/~%d types",
		diff.Summary.FunctionsAdded, diff.Summary.FunctionsRemoved, diff.Summary.FunctionsModified,
		diff.Summary.TypesAdded, diff.Summary.TypesRemoved, diff.Summary.TypesModified)
}

func TestIntegration_Registry(t *testing.T) {
	scriptsDir := filepath.Join("..", "..")

	registry := NewRegistry()
	registry.Register(NewGoExtractor())
	registry.Register(NewTypeScriptExtractor(scriptsDir))
	registry.Register(NewPythonExtractor(scriptsDir))

	tests := []struct {
		ext      string
		wantLang string
	}{
		{".go", "go"},
		{".ts", "typescript"},
		{".tsx", "typescript"},
		{".js", "typescript"},
		{".jsx", "typescript"},
		{".py", "python"},
		{".pyi", "python"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			extractor, err := registry.GetExtractor("test" + tt.ext)
			if err != nil {
				t.Fatalf("GetExtractor failed: %v", err)
			}
			if extractor.Language() != tt.wantLang {
				t.Errorf("expected language %s, got %s", tt.wantLang, extractor.Language())
			}
		})
	}

	// Test unknown extension
	_, err := registry.GetExtractor("test.unknown")
	if err == nil {
		t.Error("expected error for unknown extension")
	}
}

func TestIntegration_ExtractAll(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "testdata")
	beforePath := filepath.Join(testdataDir, "go", "before.go")
	afterPath := filepath.Join(testdataDir, "go", "after.go")

	// Skip if testdata doesn't exist
	if _, err := os.Stat(beforePath); os.IsNotExist(err) {
		t.Skipf("testdata file not found: %s", beforePath)
	}
	if _, err := os.Stat(afterPath); os.IsNotExist(err) {
		t.Skipf("testdata file not found: %s", afterPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scriptsDir := filepath.Join("..", "..")
	registry := NewRegistry()
	registry.Register(NewGoExtractor())
	registry.Register(NewTypeScriptExtractor(scriptsDir))
	registry.Register(NewPythonExtractor(scriptsDir))

	pairs := []FilePair{
		{BeforePath: beforePath, AfterPath: afterPath},
	}

	diffs, err := registry.ExtractAll(ctx, pairs)
	if err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}

	diff := diffs[0]
	if diff.Error != "" {
		t.Fatalf("diff contains error: %s", diff.Error)
	}

	if diff.Language != "go" {
		t.Errorf("expected language 'go', got '%s'", diff.Language)
	}

	// Verify multiple markdown rendering
	md := RenderMultipleMarkdown(diffs)
	if md == "" {
		t.Error("RenderMultipleMarkdown returned empty string")
	}
	if !contains(md, "Overall Summary") {
		t.Error("RenderMultipleMarkdown missing 'Overall Summary' section")
	}
}

func TestIntegration_NewFileScenario(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "testdata")
	afterPath := filepath.Join(testdataDir, "go", "after.go")

	// Skip if testdata doesn't exist
	if _, err := os.Stat(afterPath); os.IsNotExist(err) {
		t.Skipf("testdata file not found: %s", afterPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	extractor := NewGoExtractor()
	// Empty beforePath = new file
	diff, err := extractor.ExtractDiff(ctx, "", afterPath)
	if err != nil {
		t.Fatalf("ExtractDiff failed for new file: %v", err)
	}

	if diff.Error != "" {
		t.Fatalf("diff contains error: %s", diff.Error)
	}

	// All functions should be "added"
	for _, fn := range diff.Functions {
		if fn.ChangeType != ChangeAdded {
			t.Errorf("function %s should be added in new file scenario, got %s", fn.Name, fn.ChangeType)
		}
	}

	// All types should be "added"
	for _, tp := range diff.Types {
		if tp.ChangeType != ChangeAdded {
			t.Errorf("type %s should be added in new file scenario, got %s", tp.Name, tp.ChangeType)
		}
	}

	t.Logf("New file scenario: %d functions added, %d types added",
		diff.Summary.FunctionsAdded, diff.Summary.TypesAdded)
}

func TestIntegration_DeletedFileScenario(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "testdata")
	beforePath := filepath.Join(testdataDir, "go", "before.go")

	// Skip if testdata doesn't exist
	if _, err := os.Stat(beforePath); os.IsNotExist(err) {
		t.Skipf("testdata file not found: %s", beforePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	extractor := NewGoExtractor()
	// Empty afterPath = deleted file
	diff, err := extractor.ExtractDiff(ctx, beforePath, "")
	if err != nil {
		t.Fatalf("ExtractDiff failed for deleted file: %v", err)
	}

	if diff.Error != "" {
		t.Fatalf("diff contains error: %s", diff.Error)
	}

	// All functions should be "removed"
	for _, fn := range diff.Functions {
		if fn.ChangeType != ChangeRemoved {
			t.Errorf("function %s should be removed in deleted file scenario, got %s", fn.Name, fn.ChangeType)
		}
	}

	// All types should be "removed"
	for _, tp := range diff.Types {
		if tp.ChangeType != ChangeRemoved {
			t.Errorf("type %s should be removed in deleted file scenario, got %s", tp.Name, tp.ChangeType)
		}
	}

	t.Logf("Deleted file scenario: %d functions removed, %d types removed",
		diff.Summary.FunctionsRemoved, diff.Summary.TypesRemoved)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
