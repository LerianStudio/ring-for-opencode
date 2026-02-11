package dataflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

func TestPhase_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "dataflow" {
		t.Errorf("Name() = %q, want %q", got, "dataflow")
	}
}

func TestPhase_Timeout(t *testing.T) {
	p := New()
	expected := 3 * time.Minute
	if got := p.Timeout(); got != expected {
		t.Errorf("Timeout() = %v, want %v", got, expected)
	}
}

func TestPhase_ImplementsInterface(t *testing.T) {
	p := New()
	var _ phases.Phase = p
	var _ phases.SkipChecker = p
}

func TestPhase_ShouldSkip_MissingScope(t *testing.T) {
	p := New()
	cfg := &phases.Config{
		OutputDir: t.TempDir(),
	}

	skip, reason := p.ShouldSkip(cfg)
	if !skip {
		t.Error("ShouldSkip() = false, want true (scope.json missing)")
	}
	if reason == "" {
		t.Error("ShouldSkip() reason is empty")
	}
}

func TestPhase_ShouldSkip_ScopeExists(t *testing.T) {
	p := New()
	tempDir := t.TempDir()

	// Create minimal scope.json
	if err := os.WriteFile(filepath.Join(tempDir, "scope.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write scope.json: %v", err)
	}

	cfg := &phases.Config{
		OutputDir: tempDir,
	}

	skip, _ := p.ShouldSkip(cfg)
	if skip {
		t.Error("ShouldSkip() = true, want false (scope.json exists)")
	}
}

func TestPhase_Run_CancelledContext(t *testing.T) {
	p := New()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cfg := &phases.Config{
		OutputDir: t.TempDir(),
	}

	err := p.Run(ctx, cfg)
	if err == nil {
		t.Error("Run() error = nil, want context.Canceled")
	}
}

func TestMaxFilesConstant(t *testing.T) {
	if MaxFiles != 10000 {
		t.Errorf("MaxFiles = %d, want 10000", MaxFiles)
	}
}
