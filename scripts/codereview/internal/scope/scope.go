// Package scope provides language detection and file categorization for code review.
package scope

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/git"
)

// Language represents the programming language of code files.
type Language int

const (
	LanguageUnknown Language = iota
	LanguageGo
	LanguageTypeScript
	LanguagePython
	LanguageMixed
)

// String returns the string representation of the language.
func (l Language) String() string {
	switch l {
	case LanguageGo:
		return "go"
	case LanguageTypeScript:
		return "typescript"
	case LanguagePython:
		return "python"
	case LanguageMixed:
		return "mixed"
	default:
		return "unknown"
	}
}

// extensionToLanguage maps file extensions to their respective languages.
var extensionToLanguage = map[string]Language{
	".go":  LanguageGo,
	".ts":  LanguageTypeScript,
	".tsx": LanguageTypeScript,
	".py":  LanguagePython,
}

// ScopeResult contains the analysis of changed files.
type ScopeResult struct {
	BaseRef          string   `json:"base_ref"`
	HeadRef          string   `json:"head_ref"`
	Language         string   `json:"language"`
	Languages        []string `json:"languages,omitempty"`
	ModifiedFiles    []string `json:"modified"`
	AddedFiles       []string `json:"added"`
	DeletedFiles     []string `json:"deleted"`
	TotalFiles       int      `json:"total_files"`
	TotalAdditions   int      `json:"total_additions"`
	TotalDeletions   int      `json:"total_deletions"`
	PackagesAffected []string `json:"packages_affected"`
}

// gitClientInterface defines the git operations needed by Detector.
type gitClientInterface interface {
	GetDiff(baseRef, headRef string) (*git.DiffResult, error)
	GetAllChangesDiff() (*git.DiffResult, error)
	GetDiffStatsForFiles(baseRef string, files []string) (git.DiffStats, map[string]git.FileStats, error)
	FileExistsAtRef(ref, path string) (bool, error)
}

// Detector analyzes git diffs to determine language and file categorization.
type Detector struct {
	workDir   string
	gitClient gitClientInterface
}

// NewDetector creates a new Detector for the specified working directory.
func NewDetector(workDir string) *Detector {
	return &Detector{
		workDir:   workDir,
		gitClient: git.NewClient(workDir),
	}
}

// DetectFromRefs analyzes changes between two git refs.
func (d *Detector) DetectFromRefs(baseRef, headRef string) (*ScopeResult, error) {
	diffResult, err := d.gitClient.GetDiff(baseRef, headRef)
	if err != nil {
		return nil, err
	}

	return d.buildScopeResult(diffResult)
}

// DetectAllChanges analyzes all staged and unstaged changes.
func (d *Detector) DetectAllChanges() (*ScopeResult, error) {
	diffResult, err := d.gitClient.GetAllChangesDiff()
	if err != nil {
		return nil, err
	}

	return d.buildScopeResult(diffResult)
}

// buildScopeResult creates a ScopeResult from a git DiffResult.
func (d *Detector) buildScopeResult(diffResult *git.DiffResult) (*ScopeResult, error) {
	// Extract all file paths
	var allPaths []string
	for _, f := range diffResult.Files {
		allPaths = append(allPaths, f.Path)
	}

	// Detect language
	lang, err := DetectLanguage(allPaths)
	if err != nil {
		return nil, err
	}

	// Categorize files by status
	modified, added, deleted := CategorizeFilesByStatus(diffResult.Files)

	// Extract packages from code files only
	codeFiles := FilterByLanguage(allPaths, lang)
	packages := ExtractPackages(codeFiles)

	languages := DetectLanguages(allPaths)

	return &ScopeResult{
		BaseRef:          diffResult.BaseRef,
		HeadRef:          diffResult.HeadRef,
		Language:         lang.String(),
		Languages:        languages,
		ModifiedFiles:    modified,
		AddedFiles:       added,
		DeletedFiles:     deleted,
		TotalFiles:       diffResult.Stats.TotalFiles,
		TotalAdditions:   diffResult.Stats.TotalAdditions,
		TotalDeletions:   diffResult.Stats.TotalDeletions,
		PackagesAffected: packages,
	}, nil
}

// DetectLanguage detects the primary programming language from a list of file paths.
// Returns LanguageMixed if multiple code languages are detected.
func DetectLanguage(files []string) (Language, error) {
	languagesFound := make(map[Language]bool)

	for _, f := range files {
		ext := strings.ToLower(getFileExtension(f))
		if lang, ok := extensionToLanguage[ext]; ok {
			languagesFound[lang] = true
		}
	}

	// Count detected code languages (not LanguageUnknown)
	count := len(languagesFound)

	if count == 0 {
		return LanguageUnknown, nil
	}

	if count > 1 {
		return LanguageMixed, nil
	}

	// Return the single detected language (count == 1 guarantees exactly one iteration)
	for lang := range languagesFound {
		return lang, nil
	}
	return LanguageUnknown, nil // Required by compiler; logically unreachable
}

// DetectLanguages returns all detected programming languages from a list of file paths.
// Results are returned as normalized language strings.
func DetectLanguages(files []string) []string {
	if len(files) == 0 {
		return []string{}
	}

	languagesFound := make(map[Language]bool)
	for _, f := range files {
		ext := strings.ToLower(getFileExtension(f))
		if lang, ok := extensionToLanguage[ext]; ok {
			languagesFound[lang] = true
		}
	}

	if len(languagesFound) == 0 {
		return []string{}
	}

	languages := make([]string, 0, len(languagesFound))
	for lang := range languagesFound {
		languages = append(languages, lang.String())
	}
	sort.Strings(languages)
	return languages
}

// getFileExtension returns the file extension including the dot.
// Returns empty string for files without extensions or hidden files.
func getFileExtension(path string) string {
	if path == "" {
		return ""
	}

	// Get the base name (last component of path)
	base := filepath.Base(path)

	// Handle hidden files (start with dot but no other extension)
	if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".") {
		return ""
	}

	ext := filepath.Ext(base)
	return ext
}

// CategorizeFilesByStatus separates files into modified, added, and deleted categories.
// Renamed and copied files are treated as modified.
// Unknown status files are treated as modified.
func CategorizeFilesByStatus(files []git.ChangedFile) (modified, added, deleted []string) {
	modified = make([]string, 0)
	added = make([]string, 0)
	deleted = make([]string, 0)

	for _, f := range files {
		switch f.Status {
		case git.StatusAdded:
			added = append(added, f.Path)
		case git.StatusDeleted:
			deleted = append(deleted, f.Path)
		default:
			// StatusModified, StatusRenamed, StatusCopied, StatusUnknown -> modified
			modified = append(modified, f.Path)
		}
	}

	return modified, added, deleted
}

// ExtractPackages returns unique parent directories (packages) from file paths.
// Results are sorted alphabetically.
func ExtractPackages(files []string) []string {
	if len(files) == 0 {
		return []string{}
	}

	packageSet := make(map[string]bool)

	for _, f := range files {
		dir := filepath.Dir(f)
		if dir == "" || dir == "." {
			dir = "."
		}
		packageSet[dir] = true
	}

	// Convert set to slice
	packages := make([]string, 0, len(packageSet))
	for pkg := range packageSet {
		packages = append(packages, pkg)
	}

	// Sort for consistent output
	sort.Strings(packages)

	return packages
}

// FilterByLanguage returns only files matching the specified language.
// If language is LanguageUnknown, returns all files.
func FilterByLanguage(files []string, lang Language) []string {
	if len(files) == 0 {
		return []string{}
	}

	// For unknown or mixed language, return all files
	if lang == LanguageUnknown || lang == LanguageMixed {
		result := make([]string, len(files))
		copy(result, files)
		return result
	}

	result := make([]string, 0)

	for _, f := range files {
		ext := getFileExtension(f)
		if fileLang, ok := extensionToLanguage[ext]; ok && fileLang == lang {
			result = append(result, f)
		}
	}

	return result
}
