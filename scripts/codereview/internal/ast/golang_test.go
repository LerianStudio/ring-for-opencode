package ast

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func codereviewRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file path")
	}
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("failed to resolve codereview root: %v", err)
	}
	if _, err := os.Stat(absRoot); err != nil {
		t.Fatalf("codereview root does not exist: %v", err)
	}
	return absRoot
}

func setWorkingDir(t *testing.T, dir string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
}

func TestGoExtractor_ExtractDiff(t *testing.T) {
	extractor := NewGoExtractor()

	root := codereviewRoot(t)
	setWorkingDir(t, root)
	beforePath := filepath.Join("testdata", "go", "before.go")
	afterPath := filepath.Join("testdata", "go", "after.go")

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

	// Verify variable changes
	varChanges := make(map[string]ChangeType)
	for _, v := range diff.Variables {
		varChanges[v.Name] = v.ChangeType
	}

	if ct, ok := varChanges["DefaultTimeout"]; !ok || ct != ChangeAdded {
		t.Errorf("expected DefaultTimeout to be added, got %v", varChanges["DefaultTimeout"])
	}
	if ct, ok := varChanges["greetingPrefix"]; !ok || ct != ChangeAdded {
		t.Errorf("expected greetingPrefix to be added, got %v", varChanges["greetingPrefix"])
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
	if diff.Summary.VariablesAdded < 2 {
		t.Errorf("expected at least 2 variables added, got %d", diff.Summary.VariablesAdded)
	}
}

func TestGoExtractor_NewFile(t *testing.T) {
	extractor := NewGoExtractor()

	root := codereviewRoot(t)
	setWorkingDir(t, root)
	afterPath := filepath.Join("testdata", "go", "after.go")

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

	root := codereviewRoot(t)
	setWorkingDir(t, root)
	beforePath := filepath.Join("testdata", "go", "before.go")

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

func TestGoExtractor_ParseFile_RespectsScope(t *testing.T) {
	extractor := NewGoExtractor()
	path := filepath.Join(t.TempDir(), "sample.go")
	content := `package main

var Config = "Global"

type ConfigType struct {}

func init() {}
func init() {}

func main() {
	var Config = "Local"
	_ = Config
	type ConfigType struct {}
	_ = ConfigType{}
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	setWorkingDir(t, filepath.Dir(path))
	parsed, err := extractor.parseFile(path)
	if err != nil {
		t.Fatalf("parseFile failed: %v", err)
	}

	if parsed.Variables["Config"] == nil {
		t.Fatalf("expected global Config variable")
	}
	if parsed.Types["ConfigType"] == nil {
		t.Fatalf("expected global ConfigType")
	}
	if len(parsed.Functions) != 3 {
		t.Fatalf("expected 3 functions (main + 2 init), got %d", len(parsed.Functions))
	}
}
