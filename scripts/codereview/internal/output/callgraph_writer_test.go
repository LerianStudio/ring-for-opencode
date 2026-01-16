package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lerianstudio/ring/scripts/codereview/internal/callgraph"
	"github.com/lerianstudio/ring/scripts/codereview/internal/testutil"
)

var setupTestDir = testutil.SetupTestDir

func TestWriteJSON_NilResult(t *testing.T) {
	tmpDir := setupTestDir(t)

	err := WriteJSON(nil, tmpDir)
	if err == nil {
		t.Error("Expected error for nil result")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("Expected error message about nil, got: %v", err)
	}
}

func TestWriteJSON_CreatesCorrectFilename(t *testing.T) {
	tmpDir := setupTestDir(t)

	cgResult := &callgraph.CallGraphResult{
		Language:          "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{},
	}

	err := WriteJSON(cgResult, tmpDir)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	// Check file exists with correct name
	expectedPath := filepath.Join(tmpDir, "go-calls.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestWriteJSON_CreatesValidJSON(t *testing.T) {
	tmpDir := setupTestDir(t)

	cgResult := createTestCallGraphResult()

	err := WriteJSON(cgResult, tmpDir)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	// Read and validate JSON
	content, err := os.ReadFile(filepath.Join(tmpDir, "go-calls.json"))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var parsed callgraph.CallGraphResult
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("File contains invalid JSON: %v", err)
	}

	// Verify content
	if parsed.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", parsed.Language)
	}
	if len(parsed.ModifiedFunctions) != 3 {
		t.Errorf("Expected 3 modified functions, got %d", len(parsed.ModifiedFunctions))
	}
}

func TestWriteJSON_PrettyPrinted(t *testing.T) {
	tmpDir := setupTestDir(t)

	cgResult := &callgraph.CallGraphResult{
		Language:          "typescript",
		ModifiedFunctions: []callgraph.FunctionCallGraph{},
	}

	err := WriteJSON(cgResult, tmpDir)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "typescript-calls.json"))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check for indentation (pretty printed)
	if !strings.Contains(string(content), "\n") {
		t.Error("Expected pretty-printed JSON with newlines")
	}
	if !strings.Contains(string(content), "  ") {
		t.Error("Expected pretty-printed JSON with indentation")
	}
}

func TestWriteJSON_CreatesDirectory(t *testing.T) {
	tmpDir := setupTestDir(t)

	// Use nested path that doesn't exist
	nestedDir := filepath.Join(tmpDir, "nested", "deep", "output")

	cgResult := &callgraph.CallGraphResult{
		Language:          "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{},
	}

	err := WriteJSON(cgResult, nestedDir)
	if err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}

	// Check file exists
	expectedPath := filepath.Join(nestedDir, "go-calls.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestWriteImpactSummary_NilResult(t *testing.T) {
	tmpDir := setupTestDir(t)

	err := WriteImpactSummary(nil, tmpDir)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestWriteImpactSummary_CreatesFile(t *testing.T) {
	tmpDir := setupTestDir(t)

	cgResult := createTestCallGraphResult()

	err := WriteImpactSummary(cgResult, tmpDir)
	if err != nil {
		t.Fatalf("WriteImpactSummary returned error: %v", err)
	}

	// Check file exists
	expectedPath := filepath.Join(tmpDir, "impact-summary.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestWriteImpactSummary_ContainsMarkdown(t *testing.T) {
	tmpDir := setupTestDir(t)

	cgResult := createTestCallGraphResult()

	err := WriteImpactSummary(cgResult, tmpDir)
	if err != nil {
		t.Fatalf("WriteImpactSummary returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "impact-summary.md"))
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Verify markdown content
	if !strings.Contains(string(content), "# Impact Summary") {
		t.Error("Expected markdown header")
	}
	if !strings.Contains(string(content), "ProcessPayment") {
		t.Error("Expected function name in content")
	}
}

func TestWriteImpactSummary_CreatesDirectory(t *testing.T) {
	tmpDir := setupTestDir(t)

	nestedDir := filepath.Join(tmpDir, "nested", "path")

	cgResult := &callgraph.CallGraphResult{
		Language:          "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{},
	}

	err := WriteImpactSummary(cgResult, nestedDir)
	if err != nil {
		t.Fatalf("WriteImpactSummary returned error: %v", err)
	}

	// Check directory and file exist
	expectedPath := filepath.Join(nestedDir, "impact-summary.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestCallGraphWriter_WriteAll(t *testing.T) {
	tmpDir := setupTestDir(t)

	outputDir := filepath.Join(tmpDir, "output")
	writer := NewCallGraphWriter(outputDir)
	cgResult := createTestCallGraphResult()

	err := writer.WriteAll(cgResult)
	if err != nil {
		t.Fatalf("WriteAll returned error: %v", err)
	}

	// Check both files exist
	jsonPath := filepath.Join(outputDir, "go-calls.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("Expected JSON file to exist")
	}

	mdPath := filepath.Join(outputDir, "impact-summary.md")
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		t.Error("Expected markdown file to exist")
	}
}

func TestCallGraphWriter_EnsureDir(t *testing.T) {
	tmpDir := setupTestDir(t)

	outputDir := filepath.Join(tmpDir, "new", "dir")
	writer := NewCallGraphWriter(outputDir)

	// Directory shouldn't exist yet
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		t.Fatal("Directory should not exist before test")
	}

	err := writer.EnsureDir()
	if err != nil {
		t.Fatalf("EnsureDir returned error: %v", err)
	}

	// Directory should now exist
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("Directory should exist after EnsureDir")
	}
}

func TestNewCallGraphWriter(t *testing.T) {
	writer := NewCallGraphWriter("/some/path")
	if writer == nil {
		t.Error("NewCallGraphWriter returned nil")
	}
}

func TestCallGraphWriter_WriteResult(t *testing.T) {
	tmpDir := setupTestDir(t)

	writer := NewCallGraphWriter(tmpDir)
	cgResult := &callgraph.CallGraphResult{
		Language:          "python",
		ModifiedFunctions: []callgraph.FunctionCallGraph{},
	}

	err := writer.WriteResult(cgResult)
	if err != nil {
		t.Fatalf("WriteResult returned error: %v", err)
	}

	// Check file exists with correct name
	expectedPath := filepath.Join(tmpDir, "python-calls.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}

func TestCallGraphWriter_WriteSummary(t *testing.T) {
	tmpDir := setupTestDir(t)

	writer := NewCallGraphWriter(tmpDir)
	cgResult := createTestCallGraphResult()

	err := writer.WriteSummary(cgResult)
	if err != nil {
		t.Fatalf("WriteSummary returned error: %v", err)
	}

	// Check file exists
	expectedPath := filepath.Join(tmpDir, "impact-summary.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist", expectedPath)
	}
}
