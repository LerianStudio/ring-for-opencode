package lint

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGolangciLintRun_ExecutionFailure(t *testing.T) {
	linter := NewGolangciLint()
	linter.versionFn = func(ctx context.Context) (string, error) {
		return "1.2.3", nil
	}
	executor := NewExecutor()
	executor.SetRunFn(func(ctx context.Context, dir string, name string, args ...string) *ExecResult {
		return &ExecResult{Err: errors.New("boom")}
	})
	linter.executor = executor

	result, err := linter.Run(context.Background(), "/tmp", []string{"./..."})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Errors)
}

func TestGolangciLintRun_ParseFailure(t *testing.T) {
	linter := NewGolangciLint()
	linter.versionFn = func(ctx context.Context) (string, error) {
		return "1.2.3", nil
	}
	executor := NewExecutor()
	executor.SetRunFn(func(ctx context.Context, dir string, name string, args ...string) *ExecResult {
		return &ExecResult{Stdout: []byte("{broken")}
	})
	linter.executor = executor

	result, err := linter.Run(context.Background(), "/tmp", []string{"./..."})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Errors)
	assert.Empty(t, result.Findings)
}

func TestGolangciLintRun_Success(t *testing.T) {
	linter := NewGolangciLint()
	linter.versionFn = func(ctx context.Context) (string, error) {
		return "1.2.3", nil
	}
	executor := NewExecutor()
	executor.SetRunFn(func(ctx context.Context, dir string, name string, args ...string) *ExecResult {
		return &ExecResult{Stdout: []byte(`{"Issues":[{"FromLinter":"gosec","Text":"oops","Severity":"warning","SourceLines":["line"],"Pos":{"Filename":"/project/main.go","Line":12,"Column":3}}]}`)}
	})
	linter.executor = executor

	result, err := linter.Run(context.Background(), "/project", []string{"./..."})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, SeverityWarning, result.Findings[0].Severity)
	assert.Equal(t, CategorySecurity, result.Findings[0].Category)
}
