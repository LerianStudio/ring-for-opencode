package context

import (
	"context"
	"os"
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
	if got := p.Name(); got != "context" {
		t.Errorf("Name() = %q, want %q", got, "context")
	}
}

func TestPhase_Timeout(t *testing.T) {
	p := New()
	expected := 30 * time.Second
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

func TestPhase_Run_MissingInputDir(t *testing.T) {
	p := New()
	cfg := &phases.Config{
		OutputDir: "/nonexistent/directory",
	}

	err := p.Run(context.Background(), cfg)
	if err == nil {
		t.Error("Run() error = nil, want error about missing input directory")
	}
}

func TestPhase_Run_DefaultOutputDir(t *testing.T) {
	p := New()
	tempDir := t.TempDir()

	// Create the input directory with required scope.json
	if err := os.WriteFile(tempDir+"/scope.json", []byte(`{"language":"go","files":{"modified":[],"added":[],"deleted":[]}}`), 0644); err != nil {
		t.Fatalf("failed to write scope.json: %v", err)
	}

	cfg := &phases.Config{
		ScopePath: tempDir,
		OutputDir: tempDir,
	}

	// This will likely fail because there's no real data, but it should not fail due to input dir
	err := p.Run(context.Background(), cfg)
	// We expect some error because the context compiler needs more data,
	// but it should not be a "directory does not exist" error
	if err != nil && err.Error() == "input directory does not exist: "+tempDir {
		t.Error("unexpected input directory error")
	}
}
