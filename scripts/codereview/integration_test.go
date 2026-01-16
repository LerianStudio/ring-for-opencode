package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/lerianstudio/ring/scripts/codereview/internal/lint"
	"github.com/lerianstudio/ring/scripts/codereview/internal/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopeReader(t *testing.T) {
	scopePath := filepath.Join("testdata", "scope.json")
	s, err := scope.ReadScopeJSON(scopePath)

	require.NoError(t, err)
	assert.Equal(t, "main", s.BaseRef)
	assert.Equal(t, "HEAD", s.HeadRef)
	assert.Equal(t, "go", s.Language)
	assert.Equal(t, lint.LanguageGo, s.GetLanguage())

	files := s.GetAllFiles()
	assert.Len(t, files, 2)
	assert.Contains(t, files, "internal/handler/user.go")
	assert.Contains(t, files, "internal/service/notification.go")

	fileMap := s.GetAllFilesMap()
	assert.True(t, fileMap["internal/handler/user.go"])
	assert.True(t, fileMap["internal/service/notification.go"])
	assert.False(t, fileMap["nonexistent.go"])

	packages := s.GetPackages()
	assert.Len(t, packages, 2)
}

func TestLinterRegistry(t *testing.T) {
	ctx := context.Background()
	registry := lint.NewRegistry()

	// Register all linters
	registry.Register(lint.NewGolangciLint())
	registry.Register(lint.NewStaticcheck())
	registry.Register(lint.NewGosec())
	registry.Register(lint.NewTSC())
	registry.Register(lint.NewESLint())
	registry.Register(lint.NewRuff())
	registry.Register(lint.NewMypy())
	registry.Register(lint.NewPylint())
	registry.Register(lint.NewBandit())

	// Check Go linters registered
	goLinters := registry.GetLinters(lint.LanguageGo)
	assert.Len(t, goLinters, 3)

	// Check TS linters registered
	tsLinters := registry.GetLinters(lint.LanguageTypeScript)
	assert.Len(t, tsLinters, 2)

	// Check Python linters registered
	pyLinters := registry.GetLinters(lint.LanguagePython)
	assert.Len(t, pyLinters, 4)

	// Available linters depend on what's installed
	availableGo := registry.GetAvailableLinters(ctx, lint.LanguageGo)
	t.Logf("Available Go linters: %d", len(availableGo))
	for _, l := range availableGo {
		t.Logf("  - %s", l.Name())
	}
}

func TestResultAggregation(t *testing.T) {
	result := lint.NewResult()

	// Simulate findings from multiple tools
	result.AddFinding(lint.Finding{
		Tool:     "golangci-lint",
		Rule:     "SA1019",
		Severity: lint.SeverityWarning,
		File:     "internal/handler/user.go",
		Line:     45,
		Column:   12,
		Message:  "deprecated API",
		Category: lint.CategoryDeprecation,
	})

	result.AddFinding(lint.Finding{
		Tool:     "gosec",
		Rule:     "G401",
		Severity: lint.SeverityHigh,
		File:     "internal/handler/user.go",
		Line:     67,
		Column:   8,
		Message:  "weak crypto",
		Category: lint.CategorySecurity,
	})

	// Verify aggregation
	assert.Len(t, result.Findings, 2)
	assert.Equal(t, 0, result.Summary.Critical)
	assert.Equal(t, 1, result.Summary.High)
	assert.Equal(t, 1, result.Summary.Warning)
	assert.Equal(t, 0, result.Summary.Info)
	assert.Equal(t, 0, result.Summary.Unknown)

	// Test filtering
	fileMap := map[string]bool{
		"internal/handler/user.go": true,
	}
	filtered := result.FilterByFiles(fileMap)
	assert.Len(t, filtered.Findings, 2)

	// Filter to non-existent file
	fileMap2 := map[string]bool{
		"other.go": true,
	}
	filtered2 := result.FilterByFiles(fileMap2)
	assert.Len(t, filtered2.Findings, 0)
}
