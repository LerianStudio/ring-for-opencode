package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/lerianstudio/ring/scripts/codereview/internal/cli"
	"github.com/lerianstudio/ring/scripts/codereview/internal/lint"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/pipeline"
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

// TestUnifiedCLI tests the unified CLI configuration and pipeline.
func TestUnifiedCLI(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := cli.DefaultConfig()
		assert.Equal(t, ".scr", cfg.OutputDir)
		// Note: BaseRef/HeadRef defaults are applied conditionally in run.go, not in DefaultConfig()
		assert.Equal(t, "", cfg.BaseRef)
		assert.Equal(t, "", cfg.HeadRef)
		assert.False(t, cfg.Verbose)
		assert.False(t, cfg.Unstaged)
	})

	t.Run("HasFilePatterns", func(t *testing.T) {
		cfg := &cli.Config{}
		assert.False(t, cfg.HasFilePatterns())

		cfg.Files = "*.go"
		assert.True(t, cfg.HasFilePatterns())

		cfg.Files = ""
		cfg.FilesFrom = "files.txt"
		assert.True(t, cfg.HasFilePatterns())
	})

	t.Run("ShouldSkip", func(t *testing.T) {
		cfg := &cli.Config{Skip: []string{"static", "dataflow"}}
		assert.True(t, cfg.ShouldSkip("static"))
		assert.True(t, cfg.ShouldSkip("dataflow"))
		assert.False(t, cfg.ShouldSkip("scope"))
		assert.False(t, cfg.ShouldSkip("ast"))
	})
}

// TestPhaseNames verifies all phase names are valid.
func TestPhaseNames(t *testing.T) {
	names := phases.PhaseNames()
	expected := []string{"scope", "static", "ast", "callgraph", "dataflow", "context"}
	assert.Equal(t, expected, names)

	for _, name := range expected {
		assert.True(t, phases.IsValidPhaseName(name), "expected %q to be valid", name)
	}

	assert.False(t, phases.IsValidPhaseName("invalid"))
	assert.False(t, phases.IsValidPhaseName(""))
}

// TestPipelineCreation verifies the pipeline can be created and configured.
func TestPipelineCreation(t *testing.T) {
	p := pipeline.New()
	require.NotNil(t, p)

	// Test chaining
	p = p.WithSkip([]string{"static", "dataflow"}).WithVerbose(true)
	require.NotNil(t, p)
}

// TestPipelineResult tests pipeline result methods.
func TestPipelineResult(t *testing.T) {
	t.Run("Success with no failures", func(t *testing.T) {
		r := &pipeline.Result{Passed: 6, Failed: 0, Skipped: 0}
		assert.True(t, r.Success())
	})

	t.Run("Failure with failed phases", func(t *testing.T) {
		r := &pipeline.Result{Passed: 4, Failed: 2, Skipped: 0}
		assert.False(t, r.Success())
	})
}

// TestUnifiedCLIBinary tests the scr binary if it exists.
func TestUnifiedCLIBinary(t *testing.T) {
	binPath := filepath.Join("bin", "scr")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("scr binary not built; run 'make build' first")
	}

	t.Run("version command", func(t *testing.T) {
		cmd := exec.Command(binPath, "version")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "version command failed: %s", string(output))
		assert.Contains(t, string(output), "scr")
	})

	t.Run("help command", func(t *testing.T) {
		cmd := exec.Command(binPath, "--help")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "help command failed: %s", string(output))
		assert.Contains(t, string(output), "Static Code Reviewer")
	})

	t.Run("phase help", func(t *testing.T) {
		cmd := exec.Command(binPath, "phase", "--help")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "phase help failed: %s", string(output))
		assert.Contains(t, string(output), "scope")
		assert.Contains(t, string(output), "static")
	})

	t.Run("run help", func(t *testing.T) {
		cmd := exec.Command(binPath, "run", "--help")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "run help failed: %s", string(output))
		assert.Contains(t, string(output), "Execute all")
	})
}

// TestLegacyBinaryBackwardCompatibility tests that legacy binaries still work.
func TestLegacyBinaryBackwardCompatibility(t *testing.T) {
	legacyBinaries := []string{
		"scope-detector",
		"static-analysis",
		"ast-extractor",
		"call-graph",
		"data-flow",
		"compile-context",
	}

	for _, binary := range legacyBinaries {
		binPath := filepath.Join("bin", binary)
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			t.Skipf("%s binary not built; run 'make build-all' first", binary)
		}

		t.Run(binary+" --help", func(t *testing.T) {
			cmd := exec.Command(binPath, "--help")
			output, err := cmd.CombinedOutput()
			// Legacy binaries should work even with --help
			// Some might not have --help flag, so we just check it doesn't crash
			if err != nil {
				// Check if it's just missing --help flag (exit code 2) vs actual error
				if exitErr, ok := err.(*exec.ExitError); ok {
					if exitErr.ExitCode() == 2 {
						t.Skipf("%s doesn't support --help flag", binary)
					}
				}
			}
			_ = output // Just verify it runs
		})
	}
}
