package pipeline

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/testutil"
)

// mockPhase implements phases.Phase for testing.
type mockPhase struct {
	name       string
	timeout    time.Duration
	runErr     error
	runDelay   time.Duration
	shouldSkip bool
	skipReason string
}

func (m *mockPhase) Name() string           { return m.name }
func (m *mockPhase) Timeout() time.Duration { return m.timeout }

func (m *mockPhase) Run(ctx context.Context, _ *phases.Config) error {
	if m.runDelay > 0 {
		select {
		case <-time.After(m.runDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return m.runErr
}

func (m *mockPhase) ShouldSkip(_ *phases.Config) (bool, string) {
	return m.shouldSkip, m.skipReason
}

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if len(p.phases) != 6 {
		t.Errorf("len(phases) = %d, want 6", len(p.phases))
	}
	if p.skipSet == nil {
		t.Error("skipSet is nil")
	}
	if p.verbose {
		t.Error("verbose = true, want false")
	}
}

func TestPipeline_WithSkip(t *testing.T) {
	p := New().WithSkip([]string{"static", "callgraph", " dataflow "})

	if !p.skipSet["static"] {
		t.Error("skipSet[static] = false, want true")
	}
	if !p.skipSet["callgraph"] {
		t.Error("skipSet[callgraph] = false, want true")
	}
	if !p.skipSet["dataflow"] {
		t.Error("skipSet[dataflow] = false, want true")
	}
	if p.skipSet["scope"] {
		t.Error("skipSet[scope] = true, want false")
	}
}

func TestPipeline_WithVerbose(t *testing.T) {
	p := New().WithVerbose(true)
	if !p.verbose {
		t.Error("verbose = false, want true")
	}
}

func TestResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		failed   int
		expected bool
	}{
		{"no failures", 0, true},
		{"one failure", 1, false},
		{"multiple failures", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Result{Failed: tt.failed}
			if got := r.Success(); got != tt.expected {
				t.Errorf("Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPipeline_RunWithMockPhases(t *testing.T) {
	tempDir := testutil.SetupTestDir(t)

	mockPhases := []phases.Phase{
		&mockPhase{name: "phase1", timeout: 1 * time.Minute},
		&mockPhase{name: "phase2", timeout: 1 * time.Minute},
	}

	p := &Pipeline{
		phases:  mockPhases,
		skipSet: make(map[string]bool),
		stderr:  io.Discard,
	}

	cfg := &phases.Config{
		WorkDir:   tempDir,
		OutputDir: tempDir,
	}

	result, err := p.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success() {
		t.Error("Result.Success() = false, want true")
	}
	if result.Passed != 2 {
		t.Errorf("Passed = %d, want 2", result.Passed)
	}
	if result.Failed != 0 {
		t.Errorf("Failed = %d, want 0", result.Failed)
	}
	if result.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", result.Skipped)
	}
}

func TestPipeline_RunWithFailingPhase(t *testing.T) {
	tempDir := testutil.SetupTestDir(t)

	mockPhases := []phases.Phase{
		&mockPhase{name: "phase1", timeout: 1 * time.Minute},
		&mockPhase{name: "phase2", timeout: 1 * time.Minute, runErr: errors.New("phase2 failed")},
		&mockPhase{name: "phase3", timeout: 1 * time.Minute},
	}

	p := &Pipeline{
		phases:  mockPhases,
		skipSet: make(map[string]bool),
		stderr:  io.Discard,
	}

	cfg := &phases.Config{
		WorkDir:   tempDir,
		OutputDir: tempDir,
	}

	result, err := p.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Success() {
		t.Error("Result.Success() = true, want false")
	}
	if result.Passed != 2 {
		t.Errorf("Passed = %d, want 2", result.Passed)
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1", result.Failed)
	}
	if len(result.FailedPhases) != 1 || result.FailedPhases[0] != "phase2" {
		t.Errorf("FailedPhases = %v, want [phase2]", result.FailedPhases)
	}
}

func TestPipeline_RunWithSkippedPhase(t *testing.T) {
	tempDir := testutil.SetupTestDir(t)

	mockPhases := []phases.Phase{
		&mockPhase{name: "phase1", timeout: 1 * time.Minute},
		&mockPhase{name: "phase2", timeout: 1 * time.Minute},
		&mockPhase{name: "phase3", timeout: 1 * time.Minute},
	}

	p := &Pipeline{
		phases:  mockPhases,
		skipSet: map[string]bool{"phase2": true},
		stderr:  io.Discard,
	}

	cfg := &phases.Config{
		WorkDir:   tempDir,
		OutputDir: tempDir,
	}

	result, err := p.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success() {
		t.Error("Result.Success() = false, want true")
	}
	if result.Passed != 2 {
		t.Errorf("Passed = %d, want 2", result.Passed)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}

	// Verify phase2 was skipped
	var phase2Result *phases.Result
	for i := range result.PhaseResults {
		if result.PhaseResults[i].Name == "phase2" {
			phase2Result = &result.PhaseResults[i]
			break
		}
	}
	if phase2Result == nil {
		t.Fatal("phase2 result not found")
	}
	if !phase2Result.Skipped {
		t.Error("phase2.Skipped = false, want true")
	}
	if phase2Result.SkipReason != "skipped via --skip flag" {
		t.Errorf("phase2.SkipReason = %q, want %q", phase2Result.SkipReason, "skipped via --skip flag")
	}
}

func TestPipeline_RunWithShouldSkipPhase(t *testing.T) {
	tempDir := testutil.SetupTestDir(t)

	mockPhases := []phases.Phase{
		&mockPhase{name: "phase1", timeout: 1 * time.Minute},
		&mockPhase{name: "phase2", timeout: 1 * time.Minute, shouldSkip: true, skipReason: "no files to analyze"},
	}

	p := &Pipeline{
		phases:  mockPhases,
		skipSet: make(map[string]bool),
		stderr:  io.Discard,
	}

	cfg := &phases.Config{
		WorkDir:   tempDir,
		OutputDir: tempDir,
	}

	result, err := p.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Passed != 1 {
		t.Errorf("Passed = %d, want 1", result.Passed)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}

	// Verify phase2 was skipped with custom reason
	for _, pr := range result.PhaseResults {
		if pr.Name == "phase2" {
			if !pr.Skipped {
				t.Error("phase2.Skipped = false, want true")
			}
			if pr.SkipReason != "no files to analyze" {
				t.Errorf("phase2.SkipReason = %q, want %q", pr.SkipReason, "no files to analyze")
			}
			break
		}
	}
}

func TestPipeline_RunWithCancellation(t *testing.T) {
	tempDir := testutil.SetupTestDir(t)

	mockPhases := []phases.Phase{
		&mockPhase{name: "phase1", timeout: 1 * time.Minute, runDelay: 500 * time.Millisecond},
		&mockPhase{name: "phase2", timeout: 1 * time.Minute},
	}

	p := &Pipeline{
		phases:  mockPhases,
		skipSet: make(map[string]bool),
		stderr:  io.Discard,
	}

	cfg := &phases.Config{
		WorkDir:   tempDir,
		OutputDir: tempDir,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := p.Run(ctx, cfg)
	if err == nil {
		t.Error("Run() error = nil, want context error")
	}

	// Should have at least one phase result (the interrupted one)
	if len(result.PhaseResults) == 0 {
		t.Error("PhaseResults is empty, expected at least one result")
	}
}

func TestPipeline_RunWithTimeout(t *testing.T) {
	tempDir := testutil.SetupTestDir(t)

	// Phase with very short timeout but long execution
	mockPhases := []phases.Phase{
		&mockPhase{name: "slow-phase", timeout: 50 * time.Millisecond, runDelay: 200 * time.Millisecond},
	}

	p := &Pipeline{
		phases:  mockPhases,
		skipSet: make(map[string]bool),
		stderr:  io.Discard,
	}

	cfg := &phases.Config{
		WorkDir:   tempDir,
		OutputDir: tempDir,
	}

	result, err := p.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Success() {
		t.Error("Result.Success() = true, want false (timeout)")
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1", result.Failed)
	}

	// Check the error mentions timeout
	if len(result.PhaseResults) > 0 {
		pr := result.PhaseResults[0]
		if pr.Error == nil {
			t.Error("phase error = nil, want timeout error")
		}
	}
}

func TestPipeline_TotalDuration(t *testing.T) {
	tempDir := testutil.SetupTestDir(t)

	mockPhases := []phases.Phase{
		&mockPhase{name: "phase1", timeout: 1 * time.Minute, runDelay: 10 * time.Millisecond},
		&mockPhase{name: "phase2", timeout: 1 * time.Minute, runDelay: 10 * time.Millisecond},
	}

	p := &Pipeline{
		phases:  mockPhases,
		skipSet: make(map[string]bool),
		stderr:  io.Discard,
	}

	cfg := &phases.Config{
		WorkDir:   tempDir,
		OutputDir: tempDir,
	}

	result, err := p.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Total duration should be at least 20ms (two 10ms phases)
	if result.TotalDuration < 20*time.Millisecond {
		t.Errorf("TotalDuration = %v, want >= 20ms", result.TotalDuration)
	}
}

func TestPipeline_RunWithNilConfig(t *testing.T) {
	p := New()

	result, err := p.Run(context.Background(), nil)
	if err == nil {
		t.Error("Run() error = nil, want error for nil config")
	}
	if result != nil {
		t.Error("Run() result should be nil on error")
	}
	if err.Error() != "nil config" {
		t.Errorf("Run() error = %q, want %q", err.Error(), "nil config")
	}
}

func TestPipeline_RunWithEmptyOutputDir(t *testing.T) {
	p := New()
	cfg := &phases.Config{
		OutputDir: "",
	}

	result, err := p.Run(context.Background(), cfg)
	if err == nil {
		t.Error("Run() error = nil, want error for empty OutputDir")
	}
	if result != nil {
		t.Error("Run() result should be nil on error")
	}
	if err.Error() != "output directory is required" {
		t.Errorf("Run() error = %q, want %q", err.Error(), "output directory is required")
	}
}
