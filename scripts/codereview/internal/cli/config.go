// Package cli provides shared CLI configuration for the scr unified binary.
package cli

import (
	"strings"
	"time"
)

// Config holds shared configuration for all commands.
type Config struct {
	// Common flags
	WorkDir   string
	OutputDir string
	Verbose   bool

	// Git refs
	BaseRef  string
	HeadRef  string
	Unstaged bool

	// File patterns
	Files     string
	FilesFrom string

	// Execution control
	Timeout time.Duration
	Skip    []string

	// Output control
	JSONSummary string
	FailFast    bool
}

// DefaultConfig returns config with sensible defaults.
// Note: BaseRef and HeadRef defaults are applied conditionally in run.go
// when neither --files/--files-from nor --unstaged are specified.
func DefaultConfig() *Config {
	return &Config{
		OutputDir: ".scr",
		Timeout:   10 * time.Minute,
	}
}

// HasFilePatterns returns true if file patterns were specified.
func (c *Config) HasFilePatterns() bool {
	return strings.TrimSpace(c.Files) != "" || strings.TrimSpace(c.FilesFrom) != ""
}

// ShouldSkip returns true if the given phase name is in the skip list.
func (c *Config) ShouldSkip(phase string) bool {
	for _, s := range c.Skip {
		if s == phase {
			return true
		}
	}
	return false
}
