// Package artifacts provides canonical paths for phase artifacts.
package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths provides canonical artifact paths for a given output directory.
type Paths struct {
	outputDir string
}

// New creates a new Paths instance for the given output directory.
func New(outputDir string) *Paths {
	return &Paths{outputDir: outputDir}
}

// OutputDir returns the base output directory.
func (p *Paths) OutputDir() string {
	return p.outputDir
}

// ScopeJSON returns the path to scope.json (Phase 0 output).
func (p *Paths) ScopeJSON() string {
	return filepath.Join(p.outputDir, "scope.json")
}

// StaticAnalysisJSON returns the path to static-analysis.json (Phase 1 output).
func (p *Paths) StaticAnalysisJSON() string {
	return filepath.Join(p.outputDir, "static-analysis.json")
}

// ASTJSON returns the path to the AST JSON file for a given language (Phase 2 output).
func (p *Paths) ASTJSON(language string) string {
	return filepath.Join(p.outputDir, fmt.Sprintf("%s-ast.json", language))
}

// ASTBatchJSON returns the path to the AST batch input file.
func (p *Paths) ASTBatchJSON() string {
	return filepath.Join(p.outputDir, "ast-batch.json")
}

// CallGraphJSON returns the path to the call graph JSON for a given language (Phase 3 output).
func (p *Paths) CallGraphJSON(language string) string {
	return filepath.Join(p.outputDir, fmt.Sprintf("%s-calls.json", language))
}

// DataFlowJSON returns the path to the data flow JSON for a given language (Phase 4 output).
func (p *Paths) DataFlowJSON(language string) string {
	return filepath.Join(p.outputDir, fmt.Sprintf("%s-flow.json", language))
}

// SecuritySummaryMD returns the path to the security summary markdown.
func (p *Paths) SecuritySummaryMD() string {
	return filepath.Join(p.outputDir, "security-summary.md")
}

// ImpactSummaryMD returns the path to the impact summary markdown.
func (p *Paths) ImpactSummaryMD() string {
	return filepath.Join(p.outputDir, "impact-summary.md")
}

// ContextMD returns the path to a reviewer context file (Phase 5 output).
func (p *Paths) ContextMD(reviewer string) string {
	return filepath.Join(p.outputDir, fmt.Sprintf("context-%s.md", reviewer))
}

// ReviewerContextFiles returns paths to all reviewer context files.
func (p *Paths) ReviewerContextFiles() []string {
	reviewers := []string{
		"code-reviewer",
		"security-reviewer",
		"business-logic-reviewer",
		"test-reviewer",
		"nil-safety-reviewer",
	}
	paths := make([]string, len(reviewers))
	for i, r := range reviewers {
		paths[i] = p.ContextMD(r)
	}
	return paths
}

// LanguagesFile returns the path to the languages JSON file for callgraph.
func (p *Paths) LanguagesFile() string {
	return filepath.Join(p.outputDir, "callgraph-languages.json")
}

// DetectASTOutputFile finds the first existing AST output file.
// Returns the path and language, or empty strings if none found.
func (p *Paths) DetectASTOutputFile() (string, string, bool) {
	candidates := []struct {
		lang string
		path string
	}{
		{"go", p.ASTJSON("go")},
		{"typescript", p.ASTJSON("typescript")},
		{"python", p.ASTJSON("python")},
		{"mixed", p.ASTJSON("mixed")},
	}

	for _, c := range candidates {
		if fileExists(c.path) {
			return c.path, c.lang, true
		}
	}
	return "", "", false
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := statFunc(path)
	return err == nil
}

// statFunc is the function used to check file existence (allows testing)
var statFunc = defaultStat

func defaultStat(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return true, nil
}
