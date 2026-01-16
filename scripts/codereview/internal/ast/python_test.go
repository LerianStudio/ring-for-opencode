package ast

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPythonExtractor_SupportedExtensions(t *testing.T) {
	extractor := NewPythonExtractor("")

	extensions := extractor.SupportedExtensions()

	assert.Len(t, extensions, 2)
	assert.Contains(t, extensions, ".py")
	assert.Contains(t, extensions, ".pyi")
}

func TestPythonExtractor_Language(t *testing.T) {
	extractor := NewPythonExtractor("")

	assert.Equal(t, "python", extractor.Language())
}

func TestPythonExtractor_NewExtractor(t *testing.T) {
	scriptDir := "/path/to/scripts"
	extractor := NewPythonExtractor(scriptDir)

	assert.Equal(t, "python3", extractor.pythonExecutable)
	assert.Equal(t, filepath.Join(scriptDir, "py", "ast_extractor.py"), extractor.scriptPath)
}

func TestPythonExtractor_ExtractDiff(t *testing.T) {
	// Skip if python3 is not available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available, skipping Python extraction test")
	}

	// scriptDir is relative to this test file's location (internal/ast/)
	// The py directory is at ../../py from here
	scriptDir := filepath.Join("..", "..")
	extractor := NewPythonExtractor(scriptDir)

	beforePath := filepath.Join("..", "..", "testdata", "py", "before.py")
	afterPath := filepath.Join("..", "..", "testdata", "py", "after.py")

	diff, err := extractor.ExtractDiff(context.Background(), beforePath, afterPath)
	require.NoError(t, err, "ExtractDiff should succeed")

	// Verify language
	assert.Equal(t, "python", diff.Language)

	// Verify SemanticDiff structure fields exist
	assert.NotNil(t, diff.Functions, "Functions should not be nil")
	assert.NotNil(t, diff.Types, "Types should not be nil")
	assert.NotNil(t, diff.Imports, "Imports should not be nil")

	// Verify function changes
	funcChanges := make(map[string]ChangeType)
	for _, f := range diff.Functions {
		funcChanges[f.Name] = f.ChangeType
	}

	// greet should be modified (added parameter)
	if ct, ok := funcChanges["greet"]; ok {
		assert.Equal(t, ChangeModified, ct, "greet should be modified")
	}

	// format_name should be removed
	if ct, ok := funcChanges["format_name"]; ok {
		assert.Equal(t, ChangeRemoved, ct, "format_name should be removed")
	}

	// validate_email should be added
	if ct, ok := funcChanges["validate_email"]; ok {
		assert.Equal(t, ChangeAdded, ct, "validate_email should be added")
	}

	// Verify type/class changes
	typeChanges := make(map[string]ChangeType)
	for _, ty := range diff.Types {
		typeChanges[ty.Name] = ty.ChangeType
	}

	// User class should be modified (fields added)
	if ct, ok := typeChanges["User"]; ok {
		assert.Equal(t, ChangeModified, ct, "User should be modified")
	}

	// Config class should be added
	if ct, ok := typeChanges["Config"]; ok {
		assert.Equal(t, ChangeAdded, ct, "Config should be added")
	}

	// Verify import changes
	importChanges := make(map[string]ChangeType)
	for _, imp := range diff.Imports {
		importChanges[imp.Path] = imp.ChangeType
	}

	// os should be removed
	if ct, ok := importChanges["os"]; ok {
		assert.Equal(t, ChangeRemoved, ct, "os should be removed")
	}

	// logging should be added
	if ct, ok := importChanges["logging"]; ok {
		assert.Equal(t, ChangeAdded, ct, "logging should be added")
	}

	// Verify summary has reasonable values
	assert.GreaterOrEqual(t, diff.Summary.FunctionsAdded, 0, "FunctionsAdded should be >= 0")
	assert.GreaterOrEqual(t, diff.Summary.FunctionsRemoved, 0, "FunctionsRemoved should be >= 0")
	assert.GreaterOrEqual(t, diff.Summary.TypesAdded, 0, "TypesAdded should be >= 0")
}

func TestPythonExtractor_NewFile(t *testing.T) {
	// Skip if python3 is not available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available, skipping Python extraction test")
	}

	scriptDir := filepath.Join("..", "..")
	extractor := NewPythonExtractor(scriptDir)

	afterPath := filepath.Join("..", "..", "testdata", "py", "after.py")

	diff, err := extractor.ExtractDiff(context.Background(), "", afterPath)
	require.NoError(t, err, "ExtractDiff should succeed for new file")

	// All functions should be added
	for _, f := range diff.Functions {
		assert.Equal(t, ChangeAdded, f.ChangeType, "function %s should be added", f.Name)
	}

	// All types should be added
	for _, ty := range diff.Types {
		assert.Equal(t, ChangeAdded, ty.ChangeType, "type %s should be added", ty.Name)
	}
}

func TestPythonExtractor_DeletedFile(t *testing.T) {
	// Skip if python3 is not available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available, skipping Python extraction test")
	}

	scriptDir := filepath.Join("..", "..")
	extractor := NewPythonExtractor(scriptDir)

	beforePath := filepath.Join("..", "..", "testdata", "py", "before.py")

	diff, err := extractor.ExtractDiff(context.Background(), beforePath, "")
	require.NoError(t, err, "ExtractDiff should succeed for deleted file")

	// All functions should be removed
	for _, f := range diff.Functions {
		assert.Equal(t, ChangeRemoved, f.ChangeType, "function %s should be removed", f.Name)
	}

	// All types should be removed
	for _, ty := range diff.Types {
		assert.Equal(t, ChangeRemoved, ty.ChangeType, "type %s should be removed", ty.Name)
	}
}

func TestPythonExtractor_NonexistentFileTreatedAsEmpty(t *testing.T) {
	// Skip if python3 is not available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available, skipping Python extraction test")
	}

	scriptDir := filepath.Join("..", "..")
	extractor := NewPythonExtractor(scriptDir)

	// Nonexistent files are treated as empty (design decision for diff tools)
	// This allows comparing new files (empty before) and deleted files (empty after)
	diff, err := extractor.ExtractDiff(context.Background(), "/nonexistent/file.py", "")
	require.NoError(t, err, "nonexistent file should be treated as empty, not error")

	// Should return an empty diff since both sides are effectively empty
	assert.Empty(t, diff.Functions, "no functions expected from empty diff")
	assert.Empty(t, diff.Types, "no types expected from empty diff")
}

func TestPythonExtractor_InvalidScript(t *testing.T) {
	extractor := NewPythonExtractor("/nonexistent/path")

	_, err := extractor.ExtractDiff(context.Background(), "test.py", "")

	require.Error(t, err, "expected error for nonexistent script")
}
