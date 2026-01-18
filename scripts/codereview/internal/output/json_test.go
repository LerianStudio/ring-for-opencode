// Package output provides JSON formatting for code review scope results.
package output

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lerianstudio/ring/scripts/codereview/internal/scope"
)

// createTestScopeResult creates a ScopeResult for testing.
func createTestScopeResult() *scope.ScopeResult {
	return &scope.ScopeResult{
		BaseRef:          "main",
		HeadRef:          "feature-branch",
		Language:         "go",
		Languages:        []string{"go"},
		ModifiedFiles:    []string{"internal/service/user.go", "internal/handler/auth.go"},
		AddedFiles:       []string{"internal/model/account.go"},
		DeletedFiles:     []string{"internal/deprecated/old.go"},
		TotalFiles:       4,
		TotalAdditions:   150,
		TotalDeletions:   25,
		PackagesAffected: []string{"internal/handler", "internal/model", "internal/service"},
	}
}

// createEmptyScopeResult creates a ScopeResult with empty slices.
func createEmptyScopeResult() *scope.ScopeResult {
	return &scope.ScopeResult{
		BaseRef:          "main",
		HeadRef:          "HEAD",
		Language:         "unknown",
		Languages:        []string{},
		ModifiedFiles:    []string{},
		AddedFiles:       []string{},
		DeletedFiles:     []string{},
		TotalFiles:       0,
		TotalAdditions:   0,
		TotalDeletions:   0,
		PackagesAffected: []string{},
	}
}

func TestNewScopeOutput(t *testing.T) {
	result := createTestScopeResult()
	output := NewScopeOutput(result)

	if output == nil {
		t.Fatal("NewScopeOutput returned nil")
	}
}

func TestToJSON_ReturnsValidCompactJSON(t *testing.T) {
	result := createTestScopeResult()
	output := NewScopeOutput(result)

	jsonBytes, err := output.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}

	// Verify it's valid JSON by unmarshaling
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("ToJSON returned invalid JSON: %v", err)
	}

	// Verify it's compact (no newlines)
	jsonStr := string(jsonBytes)
	if strings.Contains(jsonStr, "\n") {
		t.Error("ToJSON should return compact JSON without newlines")
	}

	// Verify required fields exist
	requiredFields := []string{"base_ref", "head_ref", "language", "languages", "files", "stats", "packages_affected"}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("ToJSON missing required field: %s", field)
		}
	}
}

func TestToJSON_HasCorrectStructure(t *testing.T) {
	result := createTestScopeResult()
	output := NewScopeOutput(result)

	jsonBytes, err := output.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}

	// Unmarshal into expected structure
	var parsed scopeJSON
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal into expected structure: %v", err)
	}

	// Verify top-level fields
	if parsed.BaseRef != "main" {
		t.Errorf("Expected base_ref 'main', got '%s'", parsed.BaseRef)
	}
	if parsed.HeadRef != "feature-branch" {
		t.Errorf("Expected head_ref 'feature-branch', got '%s'", parsed.HeadRef)
	}
	if parsed.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", parsed.Language)
	}
	if len(parsed.Languages) != 1 || parsed.Languages[0] != "go" {
		t.Errorf("Expected languages ['go'], got %v", parsed.Languages)
	}

	// Verify files structure
	if len(parsed.Files.Modified) != 2 {
		t.Errorf("Expected 2 modified files, got %d", len(parsed.Files.Modified))
	}
	if len(parsed.Files.Added) != 1 {
		t.Errorf("Expected 1 added file, got %d", len(parsed.Files.Added))
	}
	if len(parsed.Files.Deleted) != 1 {
		t.Errorf("Expected 1 deleted file, got %d", len(parsed.Files.Deleted))
	}

	// Verify stats structure
	if parsed.Stats.TotalFiles != 4 {
		t.Errorf("Expected total_files 4, got %d", parsed.Stats.TotalFiles)
	}
	if parsed.Stats.TotalAdditions != 150 {
		t.Errorf("Expected total_additions 150, got %d", parsed.Stats.TotalAdditions)
	}
	if parsed.Stats.TotalDeletions != 25 {
		t.Errorf("Expected total_deletions 25, got %d", parsed.Stats.TotalDeletions)
	}

	// Verify packages
	if len(parsed.Packages) != 3 {
		t.Errorf("Expected 3 packages, got %d", len(parsed.Packages))
	}
}

func TestToPrettyJSON_ReturnsFormattedJSON(t *testing.T) {
	result := createTestScopeResult()
	output := NewScopeOutput(result)

	jsonBytes, err := output.ToPrettyJSON()
	if err != nil {
		t.Fatalf("ToPrettyJSON returned error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("ToPrettyJSON returned invalid JSON: %v", err)
	}

	// Verify it contains newlines (formatted)
	jsonStr := string(jsonBytes)
	if !strings.Contains(jsonStr, "\n") {
		t.Error("ToPrettyJSON should return formatted JSON with newlines")
	}

	// Verify indentation exists
	if !strings.Contains(jsonStr, "  ") {
		t.Error("ToPrettyJSON should return indented JSON")
	}
}

func TestToJSON_SlicesNeverNull(t *testing.T) {
	// Create result with nil slices (simulating uninitialized state)
	result := &scope.ScopeResult{
		BaseRef:          "main",
		HeadRef:          "HEAD",
		Language:         "unknown",
		Languages:        nil,
		ModifiedFiles:    nil,
		AddedFiles:       nil,
		DeletedFiles:     nil,
		TotalFiles:       0,
		TotalAdditions:   0,
		TotalDeletions:   0,
		PackagesAffected: nil,
	}

	output := NewScopeOutput(result)
	jsonBytes, err := output.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify no null values in JSON output
	if strings.Contains(jsonStr, "null") {
		t.Errorf("JSON output contains null values, which should be empty arrays: %s", jsonStr)
	}

	// Verify arrays are present (empty arrays look like [])
	if !strings.Contains(jsonStr, `"languages":[]`) {
		t.Error("Expected empty array for languages")
	}
	if !strings.Contains(jsonStr, `"modified":[]`) {
		t.Error("Expected empty array for modified files")
	}
	if !strings.Contains(jsonStr, `"added":[]`) {
		t.Error("Expected empty array for added files")
	}
	if !strings.Contains(jsonStr, `"deleted":[]`) {
		t.Error("Expected empty array for deleted files")
	}
	if !strings.Contains(jsonStr, `"packages_affected":[]`) {
		t.Error("Expected empty array for packages_affected")
	}
}

func TestWriteToFile_CreatesFileWithValidJSON(t *testing.T) {
	result := createTestScopeResult()
	output := NewScopeOutput(result)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to clean temp dir: %v", err)
		}
	})

	filePath := filepath.Join(tmpDir, "scope.json")

	// Write to file
	err = output.WriteToFile(filePath)
	if err != nil {
		t.Fatalf("WriteToFile returned error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("WriteToFile did not create file")
	}

	// Verify content is valid JSON
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("File contains invalid JSON: %v", err)
	}

	// Verify it's pretty-printed (has newlines)
	if !strings.Contains(string(content), "\n") {
		t.Error("WriteToFile should write pretty-printed JSON")
	}
}

func TestWriteToFile_CreatesParentDirectories(t *testing.T) {
	result := createTestScopeResult()
	output := NewScopeOutput(result)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to clean temp dir: %v", err)
		}
	})

	// Use nested path that doesn't exist
	filePath := filepath.Join(tmpDir, "nested", "deep", "scope.json")

	// Verify parent doesn't exist
	parentDir := filepath.Dir(filePath)
	if _, err := os.Stat(parentDir); !os.IsNotExist(err) {
		t.Fatal("Parent directory should not exist before test")
	}

	// Write to file
	err = output.WriteToFile(filePath)
	if err != nil {
		t.Fatalf("WriteToFile returned error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("WriteToFile did not create file with nested directories")
	}

	// Verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("File contains invalid JSON: %v", err)
	}
}

func TestWriteToStdout(t *testing.T) {
	result := createTestScopeResult()
	output := NewScopeOutput(result)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() failed: %v", err)
	}
	os.Stdout = w

	err = output.WriteToStdout()

	// Restore stdout
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close stdout pipe writer: %v", err)
	}
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("WriteToStdout returned error: %v", err)
	}

	// Read captured output
	capturedBytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}
	captured := string(capturedBytes)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(captured)), &parsed); err != nil {
		t.Fatalf("WriteToStdout produced invalid JSON: %v\nCaptured: %s", err, captured)
	}

	// Verify trailing newline
	if !strings.HasSuffix(captured, "\n") {
		t.Error("WriteToStdout should add trailing newline")
	}
}

func TestToJSON_WithEmptyResult(t *testing.T) {
	result := createEmptyScopeResult()
	output := NewScopeOutput(result)

	jsonBytes, err := output.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}

	// Verify it's valid JSON
	var parsed scopeJSON
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("ToJSON returned invalid JSON: %v", err)
	}

	// Verify empty arrays (not nil)
	if parsed.Files.Modified == nil {
		t.Error("Modified should be empty array, not nil")
	}
	if parsed.Files.Added == nil {
		t.Error("Added should be empty array, not nil")
	}
	if parsed.Files.Deleted == nil {
		t.Error("Deleted should be empty array, not nil")
	}
	if parsed.Packages == nil {
		t.Error("Packages should be empty array, not nil")
	}

	// Verify values
	if len(parsed.Files.Modified) != 0 {
		t.Errorf("Expected 0 modified files, got %d", len(parsed.Files.Modified))
	}
	if parsed.Stats.TotalFiles != 0 {
		t.Errorf("Expected total_files 0, got %d", parsed.Stats.TotalFiles)
	}
}

func TestWriteToFile_OverwritesExistingFile(t *testing.T) {
	result := createTestScopeResult()
	output := NewScopeOutput(result)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to clean temp dir: %v", err)
		}
	})

	filePath := filepath.Join(tmpDir, "scope.json")

	// Create existing file with different content (use unique marker)
	existingContent := []byte(`{"previous_marker": "SHOULD_BE_REPLACED"}`)
	if err := os.WriteFile(filePath, existingContent, 0o644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Write new content
	err = output.WriteToFile(filePath)
	if err != nil {
		t.Fatalf("WriteToFile returned error: %v", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Should not contain old content marker
	if strings.Contains(string(content), "SHOULD_BE_REPLACED") {
		t.Error("WriteToFile should overwrite existing file")
	}

	// Should contain new content
	var parsed scopeJSON
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("File contains invalid JSON: %v", err)
	}
	if parsed.BaseRef != "main" {
		t.Error("WriteToFile did not write correct content")
	}
}
