package lint

import "context"

// Language represents a programming language.
type Language string

const (
	LanguageGo         Language = "go"
	LanguageTypeScript Language = "typescript"
	LanguagePython     Language = "python"
)

// TargetKind declares how a linter wants its targets expressed.
// Packages are typical for Go tools, while files suit TS/Python linters.
// Project allows the linter to choose its own default scope.
type TargetKind string

const (
	TargetKindAuto     TargetKind = "auto"     // Let caller decide based on language
	TargetKindPackages TargetKind = "packages" // Linter prefers package import paths
	TargetKindFiles    TargetKind = "files"    // Linter prefers explicit file paths
	TargetKindProject  TargetKind = "project"  // Linter wants to operate on the project root
)

// Linter defines the interface for all linter implementations.
type Linter interface {
	// Name returns the linter's name (e.g., "golangci-lint", "eslint").
	Name() string

	// Language returns the language this linter supports.
	Language() Language

	// Available checks if the linter is installed and available.
	Available(ctx context.Context) bool

	// Version returns the linter's version string.
	Version(ctx context.Context) (string, error)

	// Run executes the linter and returns findings.
	// projectDir is the root directory of the project.
	// files is the list of files/packages to analyze.
	Run(ctx context.Context, projectDir string, files []string) (*Result, error)
}

// TargetSelector is an optional interface for linters that want to control
// how targets are passed (packages vs files vs whole project).
type TargetSelector interface {
	TargetKind() TargetKind
}

// Registry holds all registered linters.
type Registry struct {
	linters map[Language][]Linter
}

// NewRegistry creates a new linter registry.
func NewRegistry() *Registry {
	return &Registry{
		linters: make(map[Language][]Linter),
	}
}

// Register adds a linter to the registry.
func (r *Registry) Register(l Linter) {
	lang := l.Language()
	r.linters[lang] = append(r.linters[lang], l)
}

// GetLinters returns all linters for a specific language.
func (r *Registry) GetLinters(lang Language) []Linter {
	return r.linters[lang]
}

// GetAvailableLinters returns only available linters for a language.
func (r *Registry) GetAvailableLinters(ctx context.Context, lang Language) []Linter {
	var available []Linter
	for _, l := range r.linters[lang] {
		if l.Available(ctx) {
			available = append(available, l)
		}
	}
	return available
}
