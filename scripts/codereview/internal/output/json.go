// Package output provides JSON formatting for code review scope results.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerianstudio/ring/scripts/codereview/internal/scope"
)

// scopeJSON is the JSON output structure for scope results.
// Uses nested format for files and stats as required by the macro plan spec.
type scopeJSON struct {
	BaseRef  string    `json:"base_ref"`
	HeadRef  string    `json:"head_ref"`
	Language string    `json:"language"`
	Files    filesJSON `json:"files"`
	Stats    statsJSON `json:"stats"`
	Packages []string  `json:"packages_affected"`
}

// filesJSON groups file changes by status.
type filesJSON struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
}

// statsJSON contains aggregate statistics for the diff.
type statsJSON struct {
	TotalFiles     int `json:"total_files"`
	TotalAdditions int `json:"total_additions"`
	TotalDeletions int `json:"total_deletions"`
}

// ScopeOutput wraps a ScopeResult for JSON output.
type ScopeOutput struct {
	result *scope.ScopeResult
}

// NewScopeOutput creates a new ScopeOutput from a ScopeResult.
// Returns nil if result is nil to prevent panics in subsequent method calls.
func NewScopeOutput(result *scope.ScopeResult) *ScopeOutput {
	if result == nil {
		return nil
	}
	return &ScopeOutput{result: result}
}

// toScopeJSON converts the internal ScopeResult to the JSON output structure.
// Ensures slices are never nil (uses empty slices instead).
// Returns a zero-value scopeJSON if receiver or result is nil.
func (s *ScopeOutput) toScopeJSON() scopeJSON {
	// Defensive nil check to prevent panics
	if s == nil || s.result == nil {
		return scopeJSON{
			Files: filesJSON{
				Modified: []string{},
				Added:    []string{},
				Deleted:  []string{},
			},
			Packages: []string{},
		}
	}

	// Ensure slices are never nil
	modified := s.result.ModifiedFiles
	if modified == nil {
		modified = []string{}
	}

	added := s.result.AddedFiles
	if added == nil {
		added = []string{}
	}

	deleted := s.result.DeletedFiles
	if deleted == nil {
		deleted = []string{}
	}

	packages := s.result.PackagesAffected
	if packages == nil {
		packages = []string{}
	}

	return scopeJSON{
		BaseRef:  s.result.BaseRef,
		HeadRef:  s.result.HeadRef,
		Language: s.result.Language,
		Files: filesJSON{
			Modified: modified,
			Added:    added,
			Deleted:  deleted,
		},
		Stats: statsJSON{
			TotalFiles:     s.result.TotalFiles,
			TotalAdditions: s.result.TotalAdditions,
			TotalDeletions: s.result.TotalDeletions,
		},
		Packages: packages,
	}
}

// ToJSON returns compact JSON representation of the scope result.
// Returns an error if receiver or result is nil.
func (s *ScopeOutput) ToJSON() ([]byte, error) {
	if s == nil || s.result == nil {
		return nil, fmt.Errorf("cannot convert nil ScopeOutput to JSON")
	}
	data := s.toScopeJSON()
	return json.Marshal(data)
}

// ToPrettyJSON returns formatted JSON with indentation.
// Returns an error if receiver or result is nil.
func (s *ScopeOutput) ToPrettyJSON() ([]byte, error) {
	if s == nil || s.result == nil {
		return nil, fmt.Errorf("cannot convert nil ScopeOutput to JSON")
	}
	data := s.toScopeJSON()
	return json.MarshalIndent(data, "", "  ")
}

// WriteToFile writes pretty-printed JSON to the specified file path.
// Creates parent directories if they do not exist.
// Returns an error if receiver or result is nil.
//
// Security note: This function trusts the caller to provide safe paths.
// When accepting paths from untrusted input (e.g., CLI flags), the caller
// should validate the path is within an allowed directory.
// TODO(review): Consider adding path sanitization for untrusted inputs (security-reviewer, 2026-01-13, Low)
// TODO(review): Consider symlink protection via os.Lstat check (security-reviewer, 2026-01-13, Low)
func (s *ScopeOutput) WriteToFile(path string) error {
	if s == nil || s.result == nil {
		return fmt.Errorf("cannot write nil ScopeOutput to file")
	}

	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Get pretty JSON
	jsonBytes, err := s.ToPrettyJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// WriteToStdout writes pretty-printed JSON to stdout with a trailing newline.
// Returns an error if receiver or result is nil.
func (s *ScopeOutput) WriteToStdout() error {
	if s == nil || s.result == nil {
		return fmt.Errorf("cannot write nil ScopeOutput to stdout")
	}

	jsonBytes, err := s.ToPrettyJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write JSON with trailing newline
	_, err = fmt.Fprintf(os.Stdout, "%s\n", jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	return nil
}
