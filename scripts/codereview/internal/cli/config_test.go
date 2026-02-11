package cli

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.OutputDir != ".scr" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, ".scr")
	}
	if cfg.Timeout != 10*time.Minute {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 10*time.Minute)
	}
	// Note: BaseRef and HeadRef defaults are applied conditionally in run.go,
	// not in DefaultConfig(), so they should be empty here.
	if cfg.BaseRef != "" {
		t.Errorf("BaseRef = %q, want empty (defaults applied in run.go)", cfg.BaseRef)
	}
	if cfg.HeadRef != "" {
		t.Errorf("HeadRef = %q, want empty (defaults applied in run.go)", cfg.HeadRef)
	}
	if cfg.Verbose {
		t.Error("Verbose = true, want false")
	}
	if cfg.Unstaged {
		t.Error("Unstaged = true, want false")
	}
	if cfg.FailFast {
		t.Error("FailFast = true, want false")
	}
}

func TestConfig_HasFilePatterns(t *testing.T) {
	tests := []struct {
		name      string
		files     string
		filesFrom string
		expected  bool
	}{
		{"both empty", "", "", false},
		{"files set", "*.go", "", true},
		{"filesFrom set", "", "files.txt", true},
		{"both set", "*.go", "files.txt", true},
		{"whitespace only files", "   ", "", false}, // whitespace-only is treated as empty
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Files:     tt.files,
				FilesFrom: tt.filesFrom,
			}
			got := cfg.HasFilePatterns()
			if got != tt.expected {
				t.Errorf("HasFilePatterns() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_ShouldSkip(t *testing.T) {
	tests := []struct {
		name     string
		skip     []string
		phase    string
		expected bool
	}{
		{"empty skip list", nil, "scope", false},
		{"phase in skip list", []string{"scope", "static"}, "scope", true},
		{"phase not in skip list", []string{"scope", "static"}, "ast", false},
		{"single phase skip", []string{"dataflow"}, "dataflow", true},
		{"case sensitive", []string{"SCOPE"}, "scope", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Skip: tt.skip}
			got := cfg.ShouldSkip(tt.phase)
			if got != tt.expected {
				t.Errorf("ShouldSkip(%q) = %v, want %v", tt.phase, got, tt.expected)
			}
		})
	}
}

func TestConfig_AllFields(t *testing.T) {
	cfg := &Config{
		WorkDir:     "/project",
		OutputDir:   ".output",
		Verbose:     true,
		BaseRef:     "develop",
		HeadRef:     "feature/new",
		Unstaged:    true,
		Files:       "cmd/*.go,internal/**/*.go",
		FilesFrom:   "changed-files.txt",
		Timeout:     5 * time.Minute,
		Skip:        []string{"static", "callgraph"},
		JSONSummary: "summary.json",
		FailFast:    true,
	}

	if cfg.WorkDir != "/project" {
		t.Errorf("WorkDir = %q, want %q", cfg.WorkDir, "/project")
	}
	if cfg.OutputDir != ".output" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, ".output")
	}
	if !cfg.Verbose {
		t.Error("Verbose = false, want true")
	}
	if cfg.BaseRef != "develop" {
		t.Errorf("BaseRef = %q, want %q", cfg.BaseRef, "develop")
	}
	if cfg.HeadRef != "feature/new" {
		t.Errorf("HeadRef = %q, want %q", cfg.HeadRef, "feature/new")
	}
	if !cfg.Unstaged {
		t.Error("Unstaged = false, want true")
	}
	if cfg.Files != "cmd/*.go,internal/**/*.go" {
		t.Errorf("Files = %q, want %q", cfg.Files, "cmd/*.go,internal/**/*.go")
	}
	if cfg.FilesFrom != "changed-files.txt" {
		t.Errorf("FilesFrom = %q, want %q", cfg.FilesFrom, "changed-files.txt")
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 5*time.Minute)
	}
	if len(cfg.Skip) != 2 {
		t.Errorf("len(Skip) = %d, want 2", len(cfg.Skip))
	}
	if cfg.JSONSummary != "summary.json" {
		t.Errorf("JSONSummary = %q, want %q", cfg.JSONSummary, "summary.json")
	}
	if !cfg.FailFast {
		t.Error("FailFast = false, want true")
	}
}
