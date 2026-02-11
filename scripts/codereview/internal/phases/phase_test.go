package phases

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPhaseNames(t *testing.T) {
	names := PhaseNames()

	expected := []string{"scope", "static", "ast", "callgraph", "dataflow", "context"}
	if len(names) != len(expected) {
		t.Errorf("PhaseNames() returned %d phases, want %d", len(names), len(expected))
	}

	for i, name := range expected {
		if names[i] != name {
			t.Errorf("PhaseNames()[%d] = %q, want %q", i, names[i], name)
		}
	}
}

func TestIsValidPhaseName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"scope is valid", "scope", true},
		{"static is valid", "static", true},
		{"ast is valid", "ast", true},
		{"callgraph is valid", "callgraph", true},
		{"dataflow is valid", "dataflow", true},
		{"context is valid", "context", true},
		{"invalid phase", "invalid", false},
		{"empty string", "", false},
		{"uppercase", "SCOPE", false},
		{"similar name", "scopes", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPhaseName(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidPhaseName(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResult_Fields(t *testing.T) {
	result := Result{
		Name:       "test-phase",
		Duration:   5 * time.Second,
		Success:    true,
		Error:      nil,
		Skipped:    false,
		SkipReason: "",
	}

	if result.Name != "test-phase" {
		t.Errorf("Result.Name = %q, want %q", result.Name, "test-phase")
	}
	if result.Duration != 5*time.Second {
		t.Errorf("Result.Duration = %v, want %v", result.Duration, 5*time.Second)
	}
	if !result.Success {
		t.Error("Result.Success = false, want true")
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	if result.Skipped {
		t.Error("Result.Skipped = true, want false")
	}
}

func TestResult_WithError(t *testing.T) {
	err := errors.New("phase failed")
	result := Result{
		Name:    "failed-phase",
		Success: false,
		Error:   err,
	}

	if result.Success {
		t.Error("Result.Success = true, want false")
	}
	if result.Error == nil {
		t.Error("Result.Error = nil, want error")
	}
	if result.Error.Error() != "phase failed" {
		t.Errorf("Result.Error.Error() = %q, want %q", result.Error.Error(), "phase failed")
	}
}

func TestResult_Skipped(t *testing.T) {
	result := Result{
		Name:       "skipped-phase",
		Success:    true,
		Skipped:    true,
		SkipReason: "no files to analyze",
	}

	if !result.Skipped {
		t.Error("Result.Skipped = false, want true")
	}
	if result.SkipReason != "no files to analyze" {
		t.Errorf("Result.SkipReason = %q, want %q", result.SkipReason, "no files to analyze")
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{
		WorkDir:   "/project",
		OutputDir: ".scr",
		Verbose:   true,
		BaseRef:   "main",
		HeadRef:   "HEAD",
		Unstaged:  false,
		Files:     "*.go",
		FilesFrom: "files.txt",
		ScopePath: ".scr/scope.json",
		ASTPath:   ".scr/go-ast.json",
		Language:  "go",
		BatchFile: ".scr/ast-batch.json",
	}

	if cfg.WorkDir != "/project" {
		t.Errorf("Config.WorkDir = %q, want %q", cfg.WorkDir, "/project")
	}
	if cfg.OutputDir != ".scr" {
		t.Errorf("Config.OutputDir = %q, want %q", cfg.OutputDir, ".scr")
	}
	if !cfg.Verbose {
		t.Error("Config.Verbose = false, want true")
	}
	if cfg.BaseRef != "main" {
		t.Errorf("Config.BaseRef = %q, want %q", cfg.BaseRef, "main")
	}
	if cfg.HeadRef != "HEAD" {
		t.Errorf("Config.HeadRef = %q, want %q", cfg.HeadRef, "HEAD")
	}
	if cfg.Language != "go" {
		t.Errorf("Config.Language = %q, want %q", cfg.Language, "go")
	}
}

// mockPhase implements Phase for testing.
type mockPhase struct {
	name    string
	timeout time.Duration
	runErr  error
}

func (m *mockPhase) Name() string                            { return m.name }
func (m *mockPhase) Timeout() time.Duration                  { return m.timeout }
func (m *mockPhase) Run(_ context.Context, _ *Config) error { return m.runErr }

func TestPhaseInterface(t *testing.T) {
	phase := &mockPhase{
		name:    "mock",
		timeout: 30 * time.Second,
		runErr:  nil,
	}

	// Verify it implements Phase interface
	var _ Phase = phase

	if phase.Name() != "mock" {
		t.Errorf("Name() = %q, want %q", phase.Name(), "mock")
	}
	if phase.Timeout() != 30*time.Second {
		t.Errorf("Timeout() = %v, want %v", phase.Timeout(), 30*time.Second)
	}
	if err := phase.Run(context.Background(), &Config{}); err != nil {
		t.Errorf("Run() = %v, want nil", err)
	}
}

// mockSkipChecker implements both Phase and SkipChecker for testing.
type mockSkipChecker struct {
	mockPhase
	shouldSkip bool
	skipReason string
}

func (m *mockSkipChecker) ShouldSkip(_ *Config) (bool, string) {
	return m.shouldSkip, m.skipReason
}

func TestSkipCheckerInterface(t *testing.T) {
	phase := &mockSkipChecker{
		mockPhase: mockPhase{
			name:    "skippable",
			timeout: 1 * time.Minute,
		},
		shouldSkip: true,
		skipReason: "no changes detected",
	}

	// Verify it implements both interfaces
	var _ Phase = phase
	var _ SkipChecker = phase

	skip, reason := phase.ShouldSkip(&Config{})
	if !skip {
		t.Error("ShouldSkip() skip = false, want true")
	}
	if reason != "no changes detected" {
		t.Errorf("ShouldSkip() reason = %q, want %q", reason, "no changes detected")
	}
}
