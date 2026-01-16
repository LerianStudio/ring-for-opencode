// Package scope provides language detection and file categorization for code review.
package scope

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/lerianstudio/ring/scripts/codereview/internal/git"
)

func TestLanguage_String(t *testing.T) {
	tests := []struct {
		name     string
		lang     Language
		expected string
	}{
		{"unknown", LanguageUnknown, "unknown"},
		{"go", LanguageGo, "go"},
		{"typescript", LanguageTypeScript, "typescript"},
		{"python", LanguagePython, "python"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.lang.String()
			if got != tt.expected {
				t.Errorf("Language.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"go file", "internal/service/user.go", ".go"},
		{"typescript file", "src/components/Button.ts", ".ts"},
		{"tsx file", "src/components/Button.tsx", ".tsx"},
		{"python file", "scripts/process.py", ".py"},
		{"no extension", "Makefile", ""},
		{"hidden file", ".gitignore", ""},
		{"nested path with dots", "path.to/file.go", ".go"},
		{"multiple extensions", "archive.tar.gz", ".gz"},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFileExtension(tt.path)
			if got != tt.expected {
				t.Errorf("getFileExtension(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		expected    Language
		expectError bool
	}{
		{
			name:     "empty files",
			files:    []string{},
			expected: LanguageUnknown,
		},
		{
			name:     "only go files",
			files:    []string{"main.go", "internal/service/user.go", "cmd/app/main.go"},
			expected: LanguageGo,
		},
		{
			name:     "only typescript files",
			files:    []string{"src/index.ts", "src/components/Button.tsx", "src/utils/helper.ts"},
			expected: LanguageTypeScript,
		},
		{
			name:     "only python files",
			files:    []string{"main.py", "scripts/process.py", "tests/test_main.py"},
			expected: LanguagePython,
		},
		{
			name:     "go with non-code files",
			files:    []string{"main.go", "README.md", "Makefile", ".gitignore", "go.mod"},
			expected: LanguageGo,
		},
		{
			name:     "typescript with config files",
			files:    []string{"src/index.ts", "package.json", "tsconfig.json", ".eslintrc"},
			expected: LanguageTypeScript,
		},
		{
			name:     "only non-code files",
			files:    []string{"README.md", "Makefile", ".gitignore", "LICENSE"},
			expected: LanguageUnknown,
		},
		{
			name:        "mixed go and typescript",
			files:       []string{"main.go", "src/index.ts"},
			expectError: true,
		},
		{
			name:        "mixed go and python",
			files:       []string{"main.go", "script.py"},
			expectError: true,
		},
		{
			name:        "mixed typescript and python",
			files:       []string{"index.ts", "main.py"},
			expectError: true,
		},
		{
			name:        "all three languages",
			files:       []string{"main.go", "index.ts", "script.py"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectLanguage(tt.files)
			if tt.expectError {
				if err == nil {
					t.Errorf("DetectLanguage() expected error, got nil")
				}
				if err != ErrMixedLanguages {
					t.Errorf("DetectLanguage() error = %v, want ErrMixedLanguages", err)
				}
				return
			}
			if err != nil {
				t.Errorf("DetectLanguage() unexpected error: %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("DetectLanguage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCategorizeFilesByStatus(t *testing.T) {
	tests := []struct {
		name            string
		files           []git.ChangedFile
		expectedMod     []string
		expectedAdded   []string
		expectedDeleted []string
	}{
		{
			name:            "empty files",
			files:           []git.ChangedFile{},
			expectedMod:     []string{},
			expectedAdded:   []string{},
			expectedDeleted: []string{},
		},
		{
			name: "only modified",
			files: []git.ChangedFile{
				{Path: "file1.go", Status: git.StatusModified},
				{Path: "file2.go", Status: git.StatusModified},
			},
			expectedMod:     []string{"file1.go", "file2.go"},
			expectedAdded:   []string{},
			expectedDeleted: []string{},
		},
		{
			name: "only added",
			files: []git.ChangedFile{
				{Path: "new1.go", Status: git.StatusAdded},
				{Path: "new2.go", Status: git.StatusAdded},
			},
			expectedMod:     []string{},
			expectedAdded:   []string{"new1.go", "new2.go"},
			expectedDeleted: []string{},
		},
		{
			name: "only deleted",
			files: []git.ChangedFile{
				{Path: "old1.go", Status: git.StatusDeleted},
				{Path: "old2.go", Status: git.StatusDeleted},
			},
			expectedMod:     []string{},
			expectedAdded:   []string{},
			expectedDeleted: []string{"old1.go", "old2.go"},
		},
		{
			name: "mixed statuses",
			files: []git.ChangedFile{
				{Path: "modified.go", Status: git.StatusModified},
				{Path: "added.go", Status: git.StatusAdded},
				{Path: "deleted.go", Status: git.StatusDeleted},
				{Path: "renamed.go", Status: git.StatusRenamed, OldPath: "old_name.go"},
				{Path: "copied.go", Status: git.StatusCopied, OldPath: "source.go"},
			},
			expectedMod:     []string{"modified.go", "renamed.go", "copied.go"},
			expectedAdded:   []string{"added.go"},
			expectedDeleted: []string{"deleted.go"},
		},
		{
			name: "unknown status treated as modified",
			files: []git.ChangedFile{
				{Path: "unknown.go", Status: git.StatusUnknown},
			},
			expectedMod:     []string{"unknown.go"},
			expectedAdded:   []string{},
			expectedDeleted: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod, added, deleted := CategorizeFilesByStatus(tt.files)

			if !slices.Equal(mod, tt.expectedMod) {
				t.Errorf("CategorizeFilesByStatus() modified = %v, want %v", mod, tt.expectedMod)
			}
			if !slices.Equal(added, tt.expectedAdded) {
				t.Errorf("CategorizeFilesByStatus() added = %v, want %v", added, tt.expectedAdded)
			}
			if !slices.Equal(deleted, tt.expectedDeleted) {
				t.Errorf("CategorizeFilesByStatus() deleted = %v, want %v", deleted, tt.expectedDeleted)
			}
		})
	}
}

func TestExtractPackages(t *testing.T) {
	tests := []struct {
		name     string
		lang     Language
		files    []string
		expected []string
	}{
		{
			name:     "empty files",
			lang:     LanguageGo,
			files:    []string{},
			expected: []string{},
		},
		{
			name:     "go files same package",
			lang:     LanguageGo,
			files:    []string{"internal/service/user.go", "internal/service/auth.go"},
			expected: []string{"internal/service"},
		},
		{
			name:     "go files different packages",
			lang:     LanguageGo,
			files:    []string{"internal/service/user.go", "internal/repository/user.go", "cmd/main.go"},
			expected: []string{"cmd", "internal/repository", "internal/service"},
		},
		{
			name:     "typescript files",
			lang:     LanguageTypeScript,
			files:    []string{"src/components/Button.tsx", "src/utils/helper.ts", "src/index.ts"},
			expected: []string{"src", "src/components", "src/utils"},
		},
		{
			name:     "python files",
			lang:     LanguagePython,
			files:    []string{"myapp/services/user.py", "myapp/models/user.py", "tests/test_user.py"},
			expected: []string{"myapp/models", "myapp/services", "tests"},
		},
		{
			name:     "root level files",
			lang:     LanguageGo,
			files:    []string{"main.go", "config.go"},
			expected: []string{"."},
		},
		{
			name:     "mixed root and nested",
			lang:     LanguageGo,
			files:    []string{"main.go", "internal/app/app.go"},
			expected: []string{".", "internal/app"},
		},
		{
			name:     "unknown language",
			lang:     LanguageUnknown,
			files:    []string{"file1.txt", "file2.md"},
			expected: []string{"."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPackages(tt.lang, tt.files)
			if !slices.Equal(got, tt.expected) {
				t.Errorf("ExtractPackages() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFilterByLanguage(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		lang     Language
		expected []string
	}{
		{
			name:     "empty files",
			files:    []string{},
			lang:     LanguageGo,
			expected: []string{},
		},
		{
			name:     "filter go files",
			files:    []string{"main.go", "README.md", "internal/app.go", "Makefile"},
			lang:     LanguageGo,
			expected: []string{"main.go", "internal/app.go"},
		},
		{
			name:     "filter typescript files",
			files:    []string{"index.ts", "Button.tsx", "package.json", "styles.css"},
			lang:     LanguageTypeScript,
			expected: []string{"index.ts", "Button.tsx"},
		},
		{
			name:     "filter python files",
			files:    []string{"main.py", "requirements.txt", "tests/test_main.py", "setup.cfg"},
			lang:     LanguagePython,
			expected: []string{"main.py", "tests/test_main.py"},
		},
		{
			name:     "no matching files",
			files:    []string{"README.md", "Makefile", "LICENSE"},
			lang:     LanguageGo,
			expected: []string{},
		},
		{
			name:     "unknown language returns all files",
			files:    []string{"file1.go", "file2.ts", "file3.py"},
			lang:     LanguageUnknown,
			expected: []string{"file1.go", "file2.ts", "file3.py"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByLanguage(tt.files, tt.lang)
			if !slices.Equal(got, tt.expected) {
				t.Errorf("FilterByLanguage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewDetector(t *testing.T) {
	tests := []struct {
		name    string
		workDir string
	}{
		{"empty workDir", ""},
		{"with workDir", "/path/to/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector(tt.workDir)
			if d == nil {
				t.Fatalf("NewDetector() returned nil")
			}
			if d.workDir != tt.workDir {
				t.Errorf("NewDetector().workDir = %q, want %q", d.workDir, tt.workDir)
			}
			if d.gitClient == nil {
				t.Error("NewDetector().gitClient is nil")
			}
		})
	}
}

func TestScopeResult_JSON(t *testing.T) {
	result := &ScopeResult{
		BaseRef:          "main",
		HeadRef:          "feature",
		Language:         "go",
		ModifiedFiles:    []string{"file1.go"},
		AddedFiles:       []string{"file2.go"},
		DeletedFiles:     []string{"file3.go"},
		TotalFiles:       3,
		TotalAdditions:   100,
		TotalDeletions:   50,
		PackagesAffected: []string{"internal/service"},
	}

	// Marshal the ScopeResult to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal into a map to verify JSON field names match struct tags
	var jsonMap map[string]any
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		t.Fatalf("json.Unmarshal into map failed: %v", err)
	}

	// Verify JSON field names match the struct tags (snake_case)
	expectedKeys := []string{
		"base_ref", "head_ref", "language", "modified", "added",
		"deleted", "total_files", "total_additions", "total_deletions",
		"packages_affected",
	}
	for _, key := range expectedKeys {
		if _, exists := jsonMap[key]; !exists {
			t.Errorf("expected JSON key %q not found in marshaled output", key)
		}
	}

	// Verify string field values
	if got, ok := jsonMap["base_ref"].(string); !ok || got != "main" {
		t.Errorf("base_ref: got %v, want %q", jsonMap["base_ref"], "main")
	}
	if got, ok := jsonMap["head_ref"].(string); !ok || got != "feature" {
		t.Errorf("head_ref: got %v, want %q", jsonMap["head_ref"], "feature")
	}
	if got, ok := jsonMap["language"].(string); !ok || got != "go" {
		t.Errorf("language: got %v, want %q", jsonMap["language"], "go")
	}

	// Verify numeric field values (JSON numbers unmarshal as float64)
	if got, ok := jsonMap["total_files"].(float64); !ok || int(got) != 3 {
		t.Errorf("total_files: got %v, want %d", jsonMap["total_files"], 3)
	}
	if got, ok := jsonMap["total_additions"].(float64); !ok || int(got) != 100 {
		t.Errorf("total_additions: got %v, want %d", jsonMap["total_additions"], 100)
	}
	if got, ok := jsonMap["total_deletions"].(float64); !ok || int(got) != 50 {
		t.Errorf("total_deletions: got %v, want %d", jsonMap["total_deletions"], 50)
	}

	// Verify slice field values
	verifyJSONSlice := func(key string, expected []string) {
		arr, ok := jsonMap[key].([]any)
		if !ok {
			t.Errorf("%s: expected array, got %T", key, jsonMap[key])
			return
		}
		if len(arr) != len(expected) {
			t.Errorf("%s: got %d elements, want %d", key, len(arr), len(expected))
			return
		}
		for i, v := range arr {
			if str, ok := v.(string); !ok || str != expected[i] {
				t.Errorf("%s[%d]: got %v, want %q", key, i, v, expected[i])
			}
		}
	}

	verifyJSONSlice("modified", []string{"file1.go"})
	verifyJSONSlice("added", []string{"file2.go"})
	verifyJSONSlice("deleted", []string{"file3.go"})
	verifyJSONSlice("packages_affected", []string{"internal/service"})

	// Verify round-trip: unmarshal back into ScopeResult and compare
	var roundTrip ScopeResult
	if err := json.Unmarshal(jsonData, &roundTrip); err != nil {
		t.Fatalf("json.Unmarshal round-trip failed: %v", err)
	}

	if roundTrip.BaseRef != result.BaseRef {
		t.Errorf("round-trip BaseRef: got %q, want %q", roundTrip.BaseRef, result.BaseRef)
	}
	if roundTrip.HeadRef != result.HeadRef {
		t.Errorf("round-trip HeadRef: got %q, want %q", roundTrip.HeadRef, result.HeadRef)
	}
	if roundTrip.Language != result.Language {
		t.Errorf("round-trip Language: got %q, want %q", roundTrip.Language, result.Language)
	}
	if roundTrip.TotalFiles != result.TotalFiles {
		t.Errorf("round-trip TotalFiles: got %d, want %d", roundTrip.TotalFiles, result.TotalFiles)
	}
	if roundTrip.TotalAdditions != result.TotalAdditions {
		t.Errorf("round-trip TotalAdditions: got %d, want %d", roundTrip.TotalAdditions, result.TotalAdditions)
	}
	if roundTrip.TotalDeletions != result.TotalDeletions {
		t.Errorf("round-trip TotalDeletions: got %d, want %d", roundTrip.TotalDeletions, result.TotalDeletions)
	}
	if !slices.Equal(roundTrip.ModifiedFiles, result.ModifiedFiles) {
		t.Errorf("round-trip ModifiedFiles: got %v, want %v", roundTrip.ModifiedFiles, result.ModifiedFiles)
	}
	if !slices.Equal(roundTrip.AddedFiles, result.AddedFiles) {
		t.Errorf("round-trip AddedFiles: got %v, want %v", roundTrip.AddedFiles, result.AddedFiles)
	}
	if !slices.Equal(roundTrip.DeletedFiles, result.DeletedFiles) {
		t.Errorf("round-trip DeletedFiles: got %v, want %v", roundTrip.DeletedFiles, result.DeletedFiles)
	}
	if !slices.Equal(roundTrip.PackagesAffected, result.PackagesAffected) {
		t.Errorf("round-trip PackagesAffected: got %v, want %v", roundTrip.PackagesAffected, result.PackagesAffected)
	}
}

// TestDetector_Integration tests the Detector with mock git client.
func TestDetector_DetectFromRefs(t *testing.T) {
	tests := []struct {
		name        string
		baseRef     string
		headRef     string
		mockResult  *git.DiffResult
		mockErr     error
		expectError bool
		checkResult func(*testing.T, *ScopeResult)
	}{
		{
			name:    "successful detection with go files",
			baseRef: "main",
			headRef: "HEAD",
			mockResult: &git.DiffResult{
				BaseRef: "main",
				HeadRef: "HEAD",
				Files: []git.ChangedFile{
					{Path: "internal/service/user.go", Status: git.StatusModified, Additions: 10, Deletions: 5},
					{Path: "internal/repository/user.go", Status: git.StatusAdded, Additions: 50, Deletions: 0},
					{Path: "old_file.go", Status: git.StatusDeleted, Additions: 0, Deletions: 30},
				},
				Stats: git.DiffStats{TotalFiles: 3, TotalAdditions: 60, TotalDeletions: 35},
			},
			checkResult: func(t *testing.T, r *ScopeResult) {
				if r.BaseRef != "main" {
					t.Errorf("BaseRef = %q, want %q", r.BaseRef, "main")
				}
				if r.HeadRef != "HEAD" {
					t.Errorf("HeadRef = %q, want %q", r.HeadRef, "HEAD")
				}
				if r.Language != "go" {
					t.Errorf("Language = %q, want %q", r.Language, "go")
				}
				if len(r.ModifiedFiles) != 1 {
					t.Errorf("ModifiedFiles len = %d, want 1", len(r.ModifiedFiles))
				}
				if len(r.AddedFiles) != 1 {
					t.Errorf("AddedFiles len = %d, want 1", len(r.AddedFiles))
				}
				if len(r.DeletedFiles) != 1 {
					t.Errorf("DeletedFiles len = %d, want 1", len(r.DeletedFiles))
				}
				if r.TotalFiles != 3 {
					t.Errorf("TotalFiles = %d, want 3", r.TotalFiles)
				}
				if r.TotalAdditions != 60 {
					t.Errorf("TotalAdditions = %d, want 60", r.TotalAdditions)
				}
				if r.TotalDeletions != 35 {
					t.Errorf("TotalDeletions = %d, want 35", r.TotalDeletions)
				}
				if len(r.PackagesAffected) != 3 {
					t.Errorf("PackagesAffected len = %d, want 3", len(r.PackagesAffected))
				}
			},
		},
		{
			name:    "successful detection with typescript files",
			baseRef: "develop",
			headRef: "feature/new",
			mockResult: &git.DiffResult{
				BaseRef: "develop",
				HeadRef: "feature/new",
				Files: []git.ChangedFile{
					{Path: "src/components/Button.tsx", Status: git.StatusModified, Additions: 20, Deletions: 10},
					{Path: "src/utils/helper.ts", Status: git.StatusAdded, Additions: 30, Deletions: 0},
				},
				Stats: git.DiffStats{TotalFiles: 2, TotalAdditions: 50, TotalDeletions: 10},
			},
			checkResult: func(t *testing.T, r *ScopeResult) {
				if r.Language != "typescript" {
					t.Errorf("Language = %q, want %q", r.Language, "typescript")
				}
			},
		},
		{
			name:    "empty diff",
			baseRef: "main",
			headRef: "main",
			mockResult: &git.DiffResult{
				BaseRef: "main",
				HeadRef: "main",
				Files:   []git.ChangedFile{},
				Stats:   git.DiffStats{},
			},
			checkResult: func(t *testing.T, r *ScopeResult) {
				if r.Language != "unknown" {
					t.Errorf("Language = %q, want %q", r.Language, "unknown")
				}
				if r.TotalFiles != 0 {
					t.Errorf("TotalFiles = %d, want 0", r.TotalFiles)
				}
			},
		},
		{
			name:    "mixed languages returns error",
			baseRef: "main",
			headRef: "HEAD",
			mockResult: &git.DiffResult{
				BaseRef: "main",
				HeadRef: "HEAD",
				Files: []git.ChangedFile{
					{Path: "main.go", Status: git.StatusModified},
					{Path: "index.ts", Status: git.StatusModified},
				},
				Stats: git.DiffStats{TotalFiles: 2},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector("")
			d.gitClient = &mockGitClient{
				diffResult: tt.mockResult,
				diffErr:    tt.mockErr,
			}

			result, err := d.DetectFromRefs(tt.baseRef, tt.headRef)

			if tt.expectError {
				if err == nil {
					t.Error("DetectFromRefs() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DetectFromRefs() unexpected error: %v", err)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestDetector_DetectAllChanges(t *testing.T) {
	tests := []struct {
		name        string
		mockResult  *git.DiffResult
		mockErr     error
		expectError bool
		checkResult func(*testing.T, *ScopeResult)
	}{
		{
			name: "successful detection",
			mockResult: &git.DiffResult{
				BaseRef: "HEAD",
				HeadRef: "working-tree",
				Files: []git.ChangedFile{
					{Path: "main.py", Status: git.StatusModified, Additions: 15, Deletions: 5},
				},
				Stats: git.DiffStats{TotalFiles: 1, TotalAdditions: 15, TotalDeletions: 5},
			},
			checkResult: func(t *testing.T, r *ScopeResult) {
				if r.Language != "python" {
					t.Errorf("Language = %q, want %q", r.Language, "python")
				}
				if r.BaseRef != "HEAD" {
					t.Errorf("BaseRef = %q, want %q", r.BaseRef, "HEAD")
				}
				if r.HeadRef != "working-tree" {
					t.Errorf("HeadRef = %q, want %q", r.HeadRef, "working-tree")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDetector("")
			d.gitClient = &mockGitClient{
				allChangesResult: tt.mockResult,
				allChangesErr:    tt.mockErr,
			}

			result, err := d.DetectAllChanges()

			if tt.expectError {
				if err == nil {
					t.Error("DetectAllChanges() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("DetectAllChanges() unexpected error: %v", err)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// mockGitClient implements git operations for testing.
type mockGitClient struct {
	diffResult       *git.DiffResult
	diffErr          error
	allChangesResult *git.DiffResult
	allChangesErr    error
}

func (m *mockGitClient) GetDiff(baseRef, headRef string) (*git.DiffResult, error) {
	if m.diffErr != nil {
		return nil, m.diffErr
	}
	return m.diffResult, nil
}

func (m *mockGitClient) GetAllChangesDiff() (*git.DiffResult, error) {
	if m.allChangesErr != nil {
		return nil, m.allChangesErr
	}
	return m.allChangesResult, nil
}
