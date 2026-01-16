package ast

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderMarkdown(t *testing.T) {
	diff := &SemanticDiff{
		Language: "go",
		FilePath: "pkg/service/user.go",
		Functions: []FunctionDiff{
			{
				Name:       "GetUser",
				ChangeType: ChangeAdded,
				After: &FuncSig{
					Params:     []Param{{Name: "id", Type: "int"}},
					Returns:    []string{"*User", "error"},
					IsExported: true,
					StartLine:  10,
					EndLine:    20,
				},
			},
			{
				Name:       "DeleteUser",
				ChangeType: ChangeRemoved,
				Before: &FuncSig{
					Params:     []Param{{Name: "id", Type: "int"}},
					Returns:    []string{"error"},
					IsExported: true,
					StartLine:  25,
					EndLine:    30,
				},
			},
			{
				Name:       "UpdateUser",
				ChangeType: ChangeModified,
				Before: &FuncSig{
					Params:     []Param{{Name: "id", Type: "int"}},
					Returns:    []string{"error"},
					IsExported: true,
					StartLine:  35,
					EndLine:    45,
				},
				After: &FuncSig{
					Params:     []Param{{Name: "id", Type: "int"}, {Name: "data", Type: "UserData"}},
					Returns:    []string{"*User", "error"},
					IsExported: true,
					StartLine:  35,
					EndLine:    50,
				},
				BodyDiff: "Added validation logic",
			},
		},
		Types: []TypeDiff{
			{
				Name:       "User",
				Kind:       "struct",
				ChangeType: ChangeModified,
				Fields: []FieldDiff{
					{Name: "Email", ChangeType: ChangeAdded, NewType: "string"},
					{Name: "Age", ChangeType: ChangeRemoved, OldType: "int"},
				},
				StartLine: 5,
				EndLine:   15,
			},
			{
				Name:       "Config",
				Kind:       "struct",
				ChangeType: ChangeAdded,
				StartLine:  50,
				EndLine:    55,
			},
		},
		Imports: []ImportDiff{
			{Path: "context", ChangeType: ChangeAdded},
			{Path: "fmt", ChangeType: ChangeRemoved},
			{Path: "log", Alias: "logger", ChangeType: ChangeAdded},
		},
		Summary: ChangeSummary{
			FunctionsAdded:    1,
			FunctionsRemoved:  1,
			FunctionsModified: 1,
			TypesAdded:        1,
			TypesRemoved:      0,
			TypesModified:     1,
			ImportsAdded:      2,
			ImportsRemoved:    1,
		},
	}

	output := RenderMarkdown(diff)

	// Verify header
	assert.Contains(t, output, "# Semantic Changes: pkg/service/user.go")
	assert.Contains(t, output, "**Language:** go")

	// Verify summary table
	assert.Contains(t, output, "## Summary")
	assert.Contains(t, output, "| Category | Added | Removed | Modified |")
	assert.Contains(t, output, "| Functions | 1 | 1 | 1 |")
	assert.Contains(t, output, "| Types | 1 | 0 | 1 |")
	assert.Contains(t, output, "| Imports | 2 | 1 | - |")

	// Verify functions section
	assert.Contains(t, output, "## Functions")
	assert.Contains(t, output, "+ `GetUser`")
	assert.Contains(t, output, "- `DeleteUser`")
	assert.Contains(t, output, "~ `UpdateUser`")
	assert.Contains(t, output, "**Status:** Added")
	assert.Contains(t, output, "**Status:** Removed")
	assert.Contains(t, output, "**Status:** Modified")
	assert.Contains(t, output, "Added validation logic")

	// Verify types section
	assert.Contains(t, output, "## Types")
	assert.Contains(t, output, "~ `User` (struct)")
	assert.Contains(t, output, "+ `Config` (struct)")
	assert.Contains(t, output, "**Field Changes:**")
	assert.Contains(t, output, "| Email | added |")
	assert.Contains(t, output, "| Age | removed |")

	// Verify imports section
	assert.Contains(t, output, "## Imports")
	assert.Contains(t, output, "+ `context`")
	assert.Contains(t, output, "- `fmt`")
	assert.Contains(t, output, "+ `log` as logger")
}

func TestRenderMarkdown_EmptyDiff(t *testing.T) {
	diff := &SemanticDiff{
		Language:  "go",
		FilePath:  "pkg/empty.go",
		Functions: []FunctionDiff{},
		Types:     []TypeDiff{},
		Imports:   []ImportDiff{},
		Summary:   ChangeSummary{},
	}

	output := RenderMarkdown(diff)

	// Should have header and summary but no function/type/import sections
	assert.Contains(t, output, "# Semantic Changes: pkg/empty.go")
	assert.Contains(t, output, "**Language:** go")
	assert.Contains(t, output, "## Summary")

	// Should not have function/type/import sections when empty
	assert.NotContains(t, output, "## Functions")
	assert.NotContains(t, output, "## Types")
	assert.NotContains(t, output, "## Imports")
}

func TestGetChangeIcon(t *testing.T) {
	tests := []struct {
		name       string
		changeType ChangeType
		expected   string
	}{
		{
			name:       "added change",
			changeType: ChangeAdded,
			expected:   "+",
		},
		{
			name:       "removed change",
			changeType: ChangeRemoved,
			expected:   "-",
		},
		{
			name:       "modified change",
			changeType: ChangeModified,
			expected:   "~",
		},
		{
			name:       "renamed change",
			changeType: ChangeRenamed,
			expected:   ">",
		},
		{
			name:       "unknown change type",
			changeType: ChangeType("unknown"),
			expected:   "?",
		},
		{
			name:       "empty change type",
			changeType: ChangeType(""),
			expected:   "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := getChangeIcon(tt.changeType)
			assert.Equal(t, tt.expected, icon)
		})
	}
}

func TestRenderJSON(t *testing.T) {
	diff := &SemanticDiff{
		Language: "go",
		FilePath: "test.go",
		Functions: []FunctionDiff{
			{Name: "Foo", ChangeType: ChangeAdded},
		},
		Summary: ChangeSummary{FunctionsAdded: 1},
	}

	output, err := RenderJSON(diff)
	assert.NoError(t, err)

	// Verify it's valid JSON with expected content
	outputStr := string(output)
	assert.Contains(t, outputStr, `"language": "go"`)
	assert.Contains(t, outputStr, `"file_path": "test.go"`)
	assert.Contains(t, outputStr, `"name": "Foo"`)
	assert.Contains(t, outputStr, `"change_type": "added"`)
	assert.Contains(t, outputStr, `"functions_added": 1`)
}

func TestRenderMultipleMarkdown(t *testing.T) {
	diffs := []SemanticDiff{
		{
			Language: "go",
			FilePath: "pkg/a.go",
			Functions: []FunctionDiff{
				{Name: "FuncA", ChangeType: ChangeAdded},
			},
			Summary: ChangeSummary{FunctionsAdded: 1},
		},
		{
			Language: "go",
			FilePath: "pkg/b.go",
			Functions: []FunctionDiff{
				{Name: "FuncB", ChangeType: ChangeRemoved},
			},
			Summary: ChangeSummary{FunctionsRemoved: 1},
		},
	}

	output := RenderMultipleMarkdown(diffs)

	// Verify overall report structure
	assert.Contains(t, output, "# Semantic Diff Report")
	assert.Contains(t, output, "## Overall Summary")
	assert.Contains(t, output, "**Files analyzed:** 2")

	// Verify aggregated summary
	assert.Contains(t, output, "| Functions | 1 | 1 | 0 |")

	// Verify individual file diffs are included
	assert.Contains(t, output, "# Semantic Changes: pkg/a.go")
	assert.Contains(t, output, "# Semantic Changes: pkg/b.go")

	// Verify separators between files
	assert.True(t, strings.Count(output, "---") >= 2, "should have separators between files")
}

func TestFormatSignature(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		sig      *FuncSig
		expected string
	}{
		{
			name:     "simple function",
			funcName: "greet",
			sig: &FuncSig{
				Params:  []Param{{Name: "name", Type: "string"}},
				Returns: []string{"string"},
			},
			expected: "func greet(name: string) -> string\n",
		},
		{
			name:     "async function",
			funcName: "fetchData",
			sig: &FuncSig{
				Params:  []Param{{Name: "url", Type: "string"}},
				Returns: []string{"Data"},
				IsAsync: true,
			},
			expected: "async func fetchData(url: string) -> Data\n",
		},
		{
			name:     "method with receiver",
			funcName: "GetUser",
			sig: &FuncSig{
				Params:   []Param{{Name: "id", Type: "int"}},
				Returns:  []string{"*User", "error"},
				Receiver: "*UserService",
			},
			expected: "(*UserService) func GetUser(id: int) -> *User, error\n",
		},
		{
			name:     "void return",
			funcName: "log",
			sig: &FuncSig{
				Params:  []Param{{Name: "msg", Type: "string"}},
				Returns: []string{},
			},
			expected: "func log(msg: string) -> void\n",
		},
		{
			name:     "parameter without type",
			funcName: "process",
			sig: &FuncSig{
				Params:  []Param{{Name: "data"}},
				Returns: []string{},
			},
			expected: "func process(data) -> void\n",
		},
		{
			name:     "multiple parameters",
			funcName: "add",
			sig: &FuncSig{
				Params:  []Param{{Name: "a", Type: "int"}, {Name: "b", Type: "int"}},
				Returns: []string{"int"},
			},
			expected: "func add(a: int, b: int) -> int\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSignature(tt.funcName, tt.sig)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"added", "Added"},
		{"removed", "Removed"},
		{"modified", "Modified"},
		{"", ""},
		{"A", "A"},
		{"aBC", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := capitalizeFirst(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
