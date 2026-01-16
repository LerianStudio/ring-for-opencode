package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapGolangciSeverity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Severity
	}{
		{
			name:     "error returns high severity",
			input:    "error",
			expected: SeverityHigh,
		},
		{
			name:     "uppercase error returns high severity",
			input:    "ERROR",
			expected: SeverityHigh,
		},
		{
			name:     "warning returns warning severity",
			input:    "warning",
			expected: SeverityWarning,
		},
		{
			name:     "uppercase warning returns warning severity",
			input:    "WARNING",
			expected: SeverityWarning,
		},
		{
			name:     "info returns info severity",
			input:    "info",
			expected: SeverityInfo,
		},
		{
			name:     "empty string defaults to info severity",
			input:    "",
			expected: SeverityInfo,
		},
		{
			name:     "unknown defaults to info severity",
			input:    "unknown",
			expected: SeverityInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapGolangciSeverity(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapGolangciCategory(t *testing.T) {
	tests := []struct {
		linter   string
		expected Category
	}{
		{"gosec", CategorySecurity},
		{"gocritic", CategoryStyle},
		{"staticcheck", CategoryBug},
		{"typecheck", CategoryBug},
		{"gofmt", CategoryStyle},
		{"goimports", CategoryStyle},
		{"govet", CategoryStyle},
		{"ineffassign", CategoryUnused},
		{"deadcode", CategoryUnused},
		{"unused", CategoryUnused},
		{"varcheck", CategoryUnused},
		{"gocyclo", CategoryComplexity},
		{"gocognit", CategoryComplexity},
		{"depguard", CategoryDeprecation},
		{"unknown", CategoryOther},
	}

	for _, tt := range tests {
		t.Run(tt.linter, func(t *testing.T) {
			result := mapGolangciCategory(tt.linter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeFilePath(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		filePath   string
		expected   string
	}{
		{
			name:       "relative path unchanged",
			projectDir: "/project",
			filePath:   "internal/handler.go",
			expected:   "internal/handler.go",
		},
		{
			name:       "absolute path converted",
			projectDir: "/project",
			filePath:   "/project/internal/handler.go",
			expected:   "internal/handler.go",
		},
		{
			name:       "outside project stays absolute",
			projectDir: "/project",
			filePath:   "/other/file.go",
			expected:   "../other/file.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeFilePath(tt.projectDir, tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGolangciLint_Name(t *testing.T) {
	g := NewGolangciLint()
	assert.Equal(t, "golangci-lint", g.Name())
}

func TestGolangciLint_Language(t *testing.T) {
	g := NewGolangciLint()
	assert.Equal(t, LanguageGo, g.Language())
}
