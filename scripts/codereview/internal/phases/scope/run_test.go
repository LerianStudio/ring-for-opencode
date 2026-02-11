package scope

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
	if got := p.Name(); got != "scope" {
		t.Errorf("Name() = %q, want %q", got, "scope")
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

func TestResolveFilePatterns(t *testing.T) {
	tests := []struct {
		name      string
		files     string
		filesFrom string
		expected  []string
		wantErr   bool
	}{
		{"both empty", "", "", nil, false},
		{"files only", "*.go,*.ts", "", []string{"*.go", "*.ts"}, false},
		{"with spaces", " *.go , *.ts ", "", []string{"*.go", "*.ts"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveFilePatterns(tt.files, tt.filesFrom)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveFilePatterns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("resolveFilePatterns() = %v, want %v", got, tt.expected)
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("resolveFilePatterns()[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestResolveFilePatterns_FromFile(t *testing.T) {
	tempDir := t.TempDir()
	patternsFile := filepath.Join(tempDir, "patterns.txt")

	content := "*.go\n# comment\n*.ts\n\n*.py"
	if err := os.WriteFile(patternsFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write patterns file: %v", err)
	}

	patterns, err := resolveFilePatterns("", patternsFile)
	if err != nil {
		t.Fatalf("resolveFilePatterns() error = %v", err)
	}

	expected := []string{"*.go", "*.ts", "*.py"}
	if len(patterns) != len(expected) {
		t.Fatalf("len(patterns) = %d, want %d", len(patterns), len(expected))
	}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("patterns[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestResolveFilePatterns_NonexistentFile(t *testing.T) {
	_, err := resolveFilePatterns("", "/nonexistent/patterns.txt")
	if err == nil {
		t.Error("resolveFilePatterns() error = nil, want error")
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

func TestPhase_Run_InvalidFlagCombination_FilesWithBase(t *testing.T) {
	p := New()
	cfg := &phases.Config{
		OutputDir: t.TempDir(),
		Files:     "*.go",
		BaseRef:   "main",
	}

	err := p.Run(context.Background(), cfg)
	if err == nil {
		t.Error("Run() error = nil, want error about invalid flag combination")
	}
}

func TestPhase_Run_InvalidFlagCombination_UnstagedWithBase(t *testing.T) {
	p := New()
	cfg := &phases.Config{
		OutputDir: t.TempDir(),
		Unstaged:  true,
		BaseRef:   "main",
	}

	err := p.Run(context.Background(), cfg)
	if err == nil {
		t.Error("Run() error = nil, want error about invalid flag combination")
	}
}

func TestReadPatternsFromFile(t *testing.T) {
	tempDir := t.TempDir()
	patternsFile := filepath.Join(tempDir, "patterns.txt")

	content := "*.go\n# this is a comment\n*.ts\n   \n*.py\n"
	if err := os.WriteFile(patternsFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write patterns file: %v", err)
	}

	patterns, err := readPatternsFromFile(patternsFile)
	if err != nil {
		t.Fatalf("readPatternsFromFile() error = %v", err)
	}

	expected := []string{"*.go", "*.ts", "*.py"}
	if len(patterns) != len(expected) {
		t.Fatalf("len(patterns) = %d, want %d", len(patterns), len(expected))
	}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("patterns[%d] = %q, want %q", i, p, expected[i])
		}
	}
}
