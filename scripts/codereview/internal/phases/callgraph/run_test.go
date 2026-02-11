package callgraph

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
	if got := p.Name(); got != "callgraph" {
		t.Errorf("Name() = %q, want %q", got, "callgraph")
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
	tempDir := t.TempDir()

	// Create scope.json with empty files to test skip behavior
	scope := map[string]interface{}{
		"language": "go",
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

	skip, reason := p.ShouldSkip(cfg)
	if !skip {
		t.Error("ShouldSkip() = false, want true (no changed files)")
	}
	if reason == "" {
		t.Error("ShouldSkip() reason is empty")
	}
}

func TestPhase_ShouldSkip_UnknownLanguage(t *testing.T) {
	p := New()
	tempDir := t.TempDir()

	// Create scope.json with unknown language
	scope := map[string]interface{}{
		"language": "unknown",
		"files": map[string][]string{
			"modified": {"README.md"},
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
		t.Error("ShouldSkip() = false, want true (unknown language)")
	}
}

func TestPhase_ShouldSkip_GoLanguage(t *testing.T) {
	p := New()
	tempDir := t.TempDir()

	// Create scope.json with Go language
	scope := map[string]interface{}{
		"language": "go",
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
		t.Error("ShouldSkip() = true, want false (go language)")
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
