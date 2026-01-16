package callgraph

import (
	"fmt"
	"strings"
)

// Supported language identifiers for call graph analysis.
const (
	LangGo         = "go"
	LangGolang     = "golang"
	LangTypeScript = "typescript"
	LangTS         = "ts"
	LangPython     = "python"
	LangPy         = "py"
)

var languageNormalization = map[string]string{
	// Go
	LangGo:     LangGo,
	LangGolang: LangGo,

	// TypeScript
	LangTypeScript: LangTypeScript,
	LangTS:         LangTypeScript,

	// Python
	LangPython: LangPython,
	LangPy:     LangPython,
}

// NewAnalyzer creates a language-specific call graph analyzer.
// Supported languages:
//   - "go", "golang" for Go
//   - "typescript", "ts" for TypeScript
//   - "python", "py" for Python
//
// workDir is the root directory for the analysis (used for package/module resolution).
func NewAnalyzer(language, workDir string) (Analyzer, error) {
	switch NormalizeLanguage(language) {
	case LangGo:
		return NewGoAnalyzer(workDir), nil
	case LangTypeScript:
		return NewTypeScriptAnalyzer(workDir), nil
	case LangPython:
		return NewPythonAnalyzer(workDir), nil
	default:
		return nil, fmt.Errorf("unsupported language: %s (supported: %s)", language, strings.Join(SupportedLanguages(), ", "))
	}
}

// SupportedLanguages returns a slice of all supported language identifiers.
func SupportedLanguages() []string {
	return []string{
		LangGo,
		LangGolang,
		LangTypeScript,
		LangTS,
		LangPython,
		LangPy,
	}
}

// SupportedLanguagesNormalized returns a slice of normalized language names
// (primary identifiers only, without aliases).
func SupportedLanguagesNormalized() []string {
	return []string{
		LangGo,
		LangTypeScript,
		LangPython,
	}
}

// NormalizeLanguage converts language aliases to their primary identifier.
// Returns the input unchanged if it's not a known alias.
func NormalizeLanguage(language string) string {
	normalized, ok := languageNormalization[strings.ToLower(language)]
	if !ok {
		return language
	}

	return normalized
}

// IsSupported checks if a language is supported for call graph analysis.
func IsSupported(language string) bool {
	_, ok := languageNormalization[strings.ToLower(language)]
	return ok
}
