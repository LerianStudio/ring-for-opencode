package static

import (
	"context"
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
	if got := p.Name(); got != "static" {
		t.Errorf("Name() = %q, want %q", got, "static")
	}
}

func TestPhase_Timeout(t *testing.T) {
	p := New()
	expected := 5 * time.Minute
	if got := p.Timeout(); got != expected {
		t.Errorf("Timeout() = %v, want %v", got, expected)
	}
}

func TestPhase_ImplementsInterface(t *testing.T) {
	p := New()
	var _ phases.Phase = p
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

func TestPhase_Run_MissingScopeFile(t *testing.T) {
	p := New()
	cfg := &phases.Config{
		WorkDir:   t.TempDir(),
		OutputDir: t.TempDir(),
		ScopePath: "/nonexistent/scope.json",
	}

	err := p.Run(context.Background(), cfg)
	if err == nil {
		t.Error("Run() error = nil, want error about missing scope file")
	}
}
