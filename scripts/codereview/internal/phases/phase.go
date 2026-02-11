// Package phases provides the phase interface and implementations for the scr unified binary.
package phases

import (
	"context"
	"time"
)

// Phase represents a single analysis phase that can be executed.
type Phase interface {
	// Name returns the canonical name of the phase.
	Name() string
	// Run executes the phase with the given configuration.
	Run(ctx context.Context, cfg *Config) error
	// Timeout returns the default timeout for this phase.
	Timeout() time.Duration
}

// SkipChecker is an optional interface for phases that can be skipped.
type SkipChecker interface {
	// ShouldSkip returns true and a reason if the phase should be skipped.
	ShouldSkip(cfg *Config) (bool, string)
}

// Config holds phase-specific configuration derived from CLI flags.
type Config struct {
	// Working directory
	WorkDir string
	// Output directory for artifacts
	OutputDir string
	// Verbose output
	Verbose bool

	// Git refs for comparison
	BaseRef  string
	HeadRef  string
	Unstaged bool

	// Explicit file patterns (mutually exclusive with refs)
	Files     string
	FilesFrom string

	// Phase-specific paths
	ScopePath  string // Path to scope.json
	ASTPath    string // Path to AST JSON
	ScriptsDir string // Directory containing language scripts (ts/, py/)

	// Phase-specific options
	Language  string // Language override
	BatchFile string // Batch file for AST
}

// Result captures the outcome of a phase execution.
type Result struct {
	Name       string
	Duration   time.Duration
	Success    bool
	Error      error
	Skipped    bool
	SkipReason string
}

// PhaseNames returns the canonical names of all phases in execution order.
func PhaseNames() []string {
	return []string{
		"scope",
		"static",
		"ast",
		"callgraph",
		"dataflow",
		"context",
	}
}

// IsValidPhaseName returns true if the given name is a valid phase name.
func IsValidPhaseName(name string) bool {
	for _, p := range PhaseNames() {
		if p == name {
			return true
		}
	}
	return false
}
