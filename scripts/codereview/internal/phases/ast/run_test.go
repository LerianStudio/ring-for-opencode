package ast

import (
	"context"
	"encoding/json"
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
	if got := p.Name(); got != "ast" {
		t.Errorf("Name() = %q, want %q", got, "ast")
	}
}

func TestPhase_Timeout(t *testing.T) {
	p := New()
	expected := 2 * time.Minute
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

func TestPhase_ShouldSkip_EmptyFiles(t *testing.T) {
	p := New()
	tempDir := t.TempDir()

	// Create empty scope.json
	scope := map[string]interface{}{
		"files": map[string][]string{
			"modified": {},
			"added":    {},
			"deleted":  {},
		},
	}
	data, _ := json.Marshal(scope)
	if err := os.WriteFile(filepath.Join(tempDir, "scope.json"), data, 0644); err != nil {
		t.Fatalf("failed to write scope.json: %v", err)
	}

	cfg := &phases.Config{
		OutputDir: tempDir,
	}

	skip, _ := p.ShouldSkip(cfg)
	if !skip {
		t.Error("ShouldSkip() = false, want true (no files)")
	}
}

func TestPhase_ShouldSkip_WithFiles(t *testing.T) {
	p := New()
	tempDir := t.TempDir()

	// Create scope.json with files
	scope := map[string]interface{}{
		"files": map[string][]string{
			"modified": {"main.go"},
			"added":    {},
			"deleted":  {},
		},
	}
	data, _ := json.Marshal(scope)
	if err := os.WriteFile(filepath.Join(tempDir, "scope.json"), data, 0644); err != nil {
		t.Fatalf("failed to write scope.json: %v", err)
	}

	cfg := &phases.Config{
		OutputDir: tempDir,
	}

	skip, _ := p.ShouldSkip(cfg)
	if skip {
		t.Error("ShouldSkip() = true, want false (has files)")
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
