// Package main provides unit tests for the call-graph CLI binary.
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lerianstudio/ring/scripts/codereview/internal/callgraph"
	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
)

func TestReadJSONFileWithLimit(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid_json_file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "valid.json")
				content := `{"language": "go", "file_path": "main.go", "functions": []}`
				if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: false,
		},
		{
			name: "empty_json_file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "empty.json")
				if err := os.WriteFile(filePath, []byte{}, 0o644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: false,
		},
		{
			name: "nonexistent_file",
			setup: func(t *testing.T) string {
				return "/nonexistent/file.json"
			},
			wantErr:   true,
			errSubstr: "failed to stat file",
		},
		{
			name: "semantic_diff_array",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "array.json")
				content := `[{"language": "go", "file_path": "main.go", "functions": []}]`
				if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return filePath
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup(t)
			data, err := fileutil.ReadJSONFileWithLimit(filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("fileutil.ReadJSONFileWithLimit(%q) expected error, got nil", filePath)
					return
				}
				if tt.errSubstr != "" && !containsSubstring(err.Error(), tt.errSubstr) {
					t.Errorf("fileutil.ReadJSONFileWithLimit(%q) error = %v, want error containing %q", filePath, err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("fileutil.ReadJSONFileWithLimit(%q) unexpected error: %v", filePath, err)
					return
				}
				if data == nil {
					t.Errorf("fileutil.ReadJSONFileWithLimit(%q) returned nil data without error", filePath)
				}
			}
		})
	}
}

func TestReadJSONFileWithLimit_ContentPreservation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "content.json")
	expectedContent := `{"language": "go", "functions": [{"name": "main"}]}`

	if err := os.WriteFile(filePath, []byte(expectedContent), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	data, err := fileutil.ReadJSONFileWithLimit(filePath)
	if err != nil {
		t.Fatalf("fileutil.ReadJSONFileWithLimit returned error: %v", err)
	}

	if string(data) != expectedContent {
		t.Errorf("fileutil.ReadJSONFileWithLimit content mismatch:\ngot:  %s\nwant: %s", string(data), expectedContent)
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "go_prefix",
			filename: "go-ast.json",
			want:     callgraph.LangGo,
		},
		{
			name:     "golang_prefix",
			filename: "golang-ast.json",
			want:     callgraph.LangGo,
		},
		{
			name:     "ts_prefix",
			filename: "ts-ast.json",
			want:     callgraph.LangTypeScript,
		},
		{
			name:     "typescript_prefix",
			filename: "typescript-ast.json",
			want:     callgraph.LangTypeScript,
		},
		{
			name:     "py_prefix",
			filename: "py-ast.json",
			want:     callgraph.LangPython,
		},
		{
			name:     "python_prefix",
			filename: "python-ast.json",
			want:     callgraph.LangPython,
		},
		{
			name:     "uppercase_GO",
			filename: "GO-ast.json",
			want:     callgraph.LangGo,
		},
		{
			name:     "mixed_case_TypeScript",
			filename: "TypeScript-ast.json",
			want:     callgraph.LangTypeScript,
		},
		{
			name:     "unknown_prefix",
			filename: "rust-ast.json",
			want:     "",
		},
		{
			name:     "no_prefix",
			filename: "ast.json",
			want:     "",
		},
		{
			name:     "full_path_go",
			filename: "/some/path/go-ast.json",
			want:     callgraph.LangGo,
		},
		{
			name:     "full_path_ts",
			filename: ".ring/codereview/ts-ast.json",
			want:     callgraph.LangTypeScript,
		},
		{
			name:     "empty_filename",
			filename: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLanguage(tt.filename)
			if got != tt.want {
				t.Errorf("detectLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestBuildModifiedFunctions(t *testing.T) {
	tests := []struct {
		name     string
		diffs    []SemanticDiff
		wantLen  int
		wantFunc []string // function names expected
	}{
		{
			name:     "empty_diffs",
			diffs:    []SemanticDiff{},
			wantLen:  0,
			wantFunc: []string{},
		},
		{
			name: "single_diff_single_function",
			diffs: []SemanticDiff{
				{
					Language: "go",
					FilePath: "internal/service/user.go",
					Functions: []FunctionDiff{
						{Name: "CreateUser", ChangeType: "added"},
					},
				},
			},
			wantLen:  1,
			wantFunc: []string{"CreateUser"},
		},
		{
			name: "skip_removed_functions",
			diffs: []SemanticDiff{
				{
					Language: "go",
					FilePath: "internal/service/user.go",
					Functions: []FunctionDiff{
						{Name: "CreateUser", ChangeType: "added"},
						{Name: "DeletedFunc", ChangeType: "removed"},
						{Name: "UpdateUser", ChangeType: "modified"},
					},
				},
			},
			wantLen:  2,
			wantFunc: []string{"CreateUser", "UpdateUser"},
		},
		{
			name: "multiple_diffs",
			diffs: []SemanticDiff{
				{
					Language: "go",
					FilePath: "internal/service/user.go",
					Functions: []FunctionDiff{
						{Name: "CreateUser", ChangeType: "added"},
					},
				},
				{
					Language: "go",
					FilePath: "internal/handler/auth.go",
					Functions: []FunctionDiff{
						{Name: "Authenticate", ChangeType: "modified"},
					},
				},
			},
			wantLen:  2,
			wantFunc: []string{"CreateUser", "Authenticate"},
		},
		{
			name: "with_receiver",
			diffs: []SemanticDiff{
				{
					Language: "go",
					FilePath: "internal/service/user.go",
					Functions: []FunctionDiff{
						{
							Name:       "Create",
							ChangeType: "added",
							After:      &FuncSig{Receiver: "UserService"},
						},
					},
				},
			},
			wantLen:  1,
			wantFunc: []string{"Create"},
		},
		{
			name: "receiver_from_before_when_after_nil",
			diffs: []SemanticDiff{
				{
					Language: "go",
					FilePath: "internal/service/user.go",
					Functions: []FunctionDiff{
						{
							Name:       "OldMethod",
							ChangeType: "modified",
							Before:     &FuncSig{Receiver: "OldService"},
							After:      nil,
						},
					},
				},
			},
			wantLen:  1,
			wantFunc: []string{"OldMethod"},
		},
		{
			name: "renamed_function",
			diffs: []SemanticDiff{
				{
					Language: "go",
					FilePath: "internal/service/user.go",
					Functions: []FunctionDiff{
						{Name: "NewName", ChangeType: "renamed"},
					},
				},
			},
			wantLen:  1,
			wantFunc: []string{"NewName"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildModifiedFunctions(tt.diffs)

			if len(got) != tt.wantLen {
				t.Errorf("buildModifiedFunctions() returned %d functions, want %d", len(got), tt.wantLen)
			}

			// Verify function names
			gotNames := make(map[string]bool)
			for _, f := range got {
				gotNames[f.Name] = true
			}

			for _, wantName := range tt.wantFunc {
				if !gotNames[wantName] {
					t.Errorf("buildModifiedFunctions() missing expected function %q", wantName)
				}
			}
		})
	}
}

func TestBuildModifiedFunctions_PackageExtraction(t *testing.T) {
	diffs := []SemanticDiff{
		{
			Language: "go",
			FilePath: "internal/service/user.go",
			Functions: []FunctionDiff{
				{Name: "CreateUser", ChangeType: "added"},
			},
		},
	}

	got := buildModifiedFunctions(diffs)
	if len(got) != 1 {
		t.Fatalf("Expected 1 function, got %d", len(got))
	}

	if got[0].Package != "service" {
		t.Errorf("Expected package 'service', got %q", got[0].Package)
	}

	if got[0].File != "internal/service/user.go" {
		t.Errorf("Expected file 'internal/service/user.go', got %q", got[0].File)
	}
}

func TestExtractPackageFromPath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "standard_go_path",
			filePath: "internal/service/user.go",
			want:     "service",
		},
		{
			name:     "nested_path",
			filePath: "pkg/api/v1/handlers/auth.go",
			want:     "handlers",
		},
		{
			name:     "root_file",
			filePath: "main.go",
			want:     "main",
		},
		{
			name:     "single_directory",
			filePath: "cmd/main.go",
			want:     "cmd",
		},
		{
			name:     "empty_path",
			filePath: "",
			want:     "main",
		},
		{
			name:     "dot_path",
			filePath: "./main.go",
			want:     "main",
		},
		{
			name:     "absolute_path",
			filePath: "/home/user/project/internal/repo/database.go",
			want:     "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPackageFromPath(tt.filePath)
			if got != tt.want {
				t.Errorf("extractPackageFromPath(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestPreviewASTData(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		maxLen  int
		wantLen int
		wantEnd string
	}{
		{
			name:    "short_data_under_512",
			input:   []byte(`{"key": "value"}`),
			maxLen:  512,
			wantLen: 16,
			wantEnd: "",
		},
		{
			name:    "exact_512_bytes",
			input:   make([]byte, 512),
			maxLen:  512,
			wantLen: 512,
			wantEnd: "",
		},
		{
			name:    "over_512_bytes",
			input:   make([]byte, 1000),
			maxLen:  512,
			wantLen: 512 + len("...(truncated)"),
			wantEnd: "...(truncated)",
		},
		{
			name:    "empty_data",
			input:   []byte{},
			maxLen:  512,
			wantLen: 0,
			wantEnd: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill byte arrays with recognizable content
			if len(tt.input) > 0 && tt.input[0] == 0 {
				for i := range tt.input {
					tt.input[i] = 'a'
				}
			}

			got := previewASTData(tt.input)

			if len(got) != tt.wantLen {
				t.Errorf("previewASTData() length = %d, want %d", len(got), tt.wantLen)
			}

			if tt.wantEnd != "" {
				if len(got) < len(tt.wantEnd) {
					t.Errorf("previewASTData() too short to contain ending %q", tt.wantEnd)
				} else if got[len(got)-len(tt.wantEnd):] != tt.wantEnd {
					t.Errorf("previewASTData() should end with %q, got ending %q",
						tt.wantEnd, got[len(got)-len(tt.wantEnd):])
				}
			}
		})
	}
}

func TestPreviewASTData_ContentPreservation(t *testing.T) {
	// Test that short content is preserved exactly
	input := `{"language": "go", "functions": []}`
	got := previewASTData([]byte(input))

	if got != input {
		t.Errorf("previewASTData() should preserve short content exactly:\ngot:  %s\nwant: %s", got, input)
	}
}

func TestPreviewASTData_TruncationPreservesStart(t *testing.T) {
	// Create content longer than 512 bytes
	prefix := `{"important": "data", "start": "`
	content := prefix + string(make([]byte, 600)) + `"}`

	got := previewASTData([]byte(content))

	// Verify the start is preserved
	if len(got) < len(prefix) {
		t.Fatalf("previewASTData() too short, got %d chars", len(got))
	}

	if got[:len(prefix)] != prefix {
		t.Errorf("previewASTData() should preserve start of content:\ngot:  %s\nwant: %s", got[:len(prefix)], prefix)
	}
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
