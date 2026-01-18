// Package scope handles reading and parsing scope.json from Phase 0.
package scope

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/lint"
)

// ScopeJSON represents the scope.json structure from Phase 0.
type ScopeJSON struct {
	BaseRef   string        `json:"base_ref"`
	HeadRef   string        `json:"head_ref"`
	Language  string        `json:"language"` // Primary detected language
	Languages []string      `json:"languages,omitempty"`
	Files     FilesByStatus `json:"files"`
	Stats     StatsJSON     `json:"stats"`
	Packages  []string      `json:"packages_affected"`
}

func normalizeScopeJSON(scope *ScopeJSON) {
	if scope == nil {
		return
	}
	if scope.Files.Modified == nil {
		scope.Files.Modified = []string{}
	}
	if scope.Files.Added == nil {
		scope.Files.Added = []string{}
	}
	if scope.Files.Deleted == nil {
		scope.Files.Deleted = []string{}
	}
	if scope.Languages == nil {
		scope.Languages = []string{}
	}
	if scope.Packages == nil {
		scope.Packages = []string{}
	}
}

// FilesByStatus holds categorized file lists.
type FilesByStatus struct {
	Modified []string `json:"modified"`
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
}

// StatsJSON holds change statistics.
type StatsJSON struct {
	TotalFiles     int `json:"total_files"`
	TotalAdditions int `json:"total_additions"`
	TotalDeletions int `json:"total_deletions"`
}

// ReadScopeJSON reads and parses scope.json from the given path.
func ReadScopeJSON(scopePath string) (*ScopeJSON, error) {
	data, err := fileutil.ReadJSONFileWithLimit(scopePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scope.json: %w", err)
	}

	var scope ScopeJSON
	if err := json.Unmarshal(data, &scope); err != nil {
		return nil, fmt.Errorf("failed to parse scope.json: %w", err)
	}

	normalizeScopeJSON(&scope)
	return &scope, nil
}

// GetLanguage returns the primary language as a lint.Language.
func (s *ScopeJSON) GetLanguage() lint.Language {
	switch s.Language {
	case "go":
		return lint.LanguageGo
	case "typescript", "ts":
		return lint.LanguageTypeScript
	case "python", "py":
		return lint.LanguagePython
	case "mixed":
		return lint.LanguageMixed

	default:
		return lint.Language("")
	}
}

// GetAllFiles returns all changed files (modified + added) with normalized paths.
func (s *ScopeJSON) GetAllFiles() []string {
	all := make([]string, 0, len(s.Files.Modified)+len(s.Files.Added))
	for _, f := range s.Files.Modified {
		all = append(all, normalizeScopePath(f))
	}
	for _, f := range s.Files.Added {
		all = append(all, normalizeScopePath(f))
	}
	return all
}

// GetAllFilesMap returns a map of all changed files for quick lookup with normalized paths.
func (s *ScopeJSON) GetAllFilesMap() map[string]bool {
	fileMap := make(map[string]bool)
	for _, f := range s.Files.Modified {
		fileMap[normalizeScopePath(f)] = true
	}
	for _, f := range s.Files.Added {
		fileMap[normalizeScopePath(f)] = true
	}
	return fileMap
}

// NormalizeLanguage maps supported language aliases to canonical identifiers.
func NormalizeLanguage(lang string) lint.Language {
	switch strings.ToLower(lang) {
	case "go", "golang":
		return lint.LanguageGo
	case "typescript", "ts", "javascript", "js":
		return lint.LanguageTypeScript
	case "python", "py":
		return lint.LanguagePython
	case "mixed":
		return lint.LanguageMixed
	default:
		return lint.Language("")
	}
}

// normalizeScopePath normalizes file paths for consistent matching.
// Strips leading "./" or ".\\" and cleans path separators.
func normalizeScopePath(path string) string {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, ".\\")
	return path
}

// GetPackages returns the affected packages.
func (s *ScopeJSON) GetPackages() []string {
	return s.Packages
}

// DefaultScopePath returns the default scope.json path.
func DefaultScopePath(projectDir string) string {
	return filepath.Join(projectDir, ".ring", "codereview", "scope.json")
}
