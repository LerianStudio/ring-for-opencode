package context

import "fmt"

// DataSource represents a data source for a reviewer.
type DataSource struct {
	Name        string // e.g., "static-analysis", "semantic-diff"
	Description string // What this data provides
}

// reviewerDataSources maps each reviewer to their primary data sources.
var reviewerDataSources = map[string][]DataSource{
	"code-reviewer": {
		{Name: "static-analysis", Description: "Lint findings, style issues, deprecations"},
		{Name: "semantic-diff", Description: "Function/type changes, signature modifications"},
	},
	"security-reviewer": {
		{Name: "static-analysis-security", Description: "Security scanner findings (gosec, bandit)"},
		{Name: "security-summary", Description: "Data flow analysis, taint tracking"},
	},
	"business-logic-reviewer": {
		{Name: "semantic-diff", Description: "Business logic changes in functions/methods"},
		{Name: "impact-summary", Description: "Call graph impact, affected callers"},
	},
	"test-reviewer": {
		{Name: "impact-summary", Description: "Test coverage for modified code"},
	},
	"nil-safety-reviewer": {
		{Name: "data-flow", Description: "Nil/None source analysis"},
	},
}

// reviewerOrder defines the canonical order of reviewers.
// This is the single source of truth for reviewer names and their order.
var reviewerOrder = []string{
	"code-reviewer",
	"security-reviewer",
	"business-logic-reviewer",
	"test-reviewer",
	"nil-safety-reviewer",
}

func init() {
	// Validate that all reviewers in reviewerOrder exist in reviewerDataSources.
	// This catches drift at startup rather than at runtime.
	for _, name := range reviewerOrder {
		if _, ok := reviewerDataSources[name]; !ok {
			panic(fmt.Sprintf("reviewer %q in reviewerOrder but not in reviewerDataSources", name))
		}
	}
}

// severityOrder defines severity ranking for filtering.
var severityOrder = map[string]int{
	"critical": 4,
	"high":     3,
	"medium":   2,
	"warning":  1,
	"info":     0,
}

// securityCategories lists categories relevant to security reviewer.
var securityCategories = map[string]bool{
	"security":       true,
	"vulnerability":  true,
	"injection":      true,
	"cryptography":   true,
	"authentication": true,
	"authorization":  true,
	"xss":            true,
	"sqli":           true,
	"secrets":        true,
}

// codeQualityCategories lists categories relevant to code reviewer.
var codeQualityCategories = map[string]bool{
	"bug":         true,
	"style":       true,
	"performance": true,
	"deprecation": true,
	"complexity":  true,
	"type":        true,
	"unused":      true,
	"other":       true,
}

// GetReviewerNames returns the list of all reviewer names in order.
func GetReviewerNames() []string {
	return reviewerOrder
}

// GetReviewerDataSources returns the data sources for a specific reviewer.
// Returns an empty slice if the reviewer is not found.
func GetReviewerDataSources(reviewer string) []DataSource {
	if sources, ok := reviewerDataSources[reviewer]; ok {
		return sources
	}
	return []DataSource{}
}

// FilterFindingsByCategory filters findings to only those in the specified category.
func FilterFindingsByCategory(findings []Finding, category string) []Finding {
	var filtered []Finding
	for _, f := range findings {
		if f.Category == category {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterFindingsBySeverity filters findings at or above the specified severity.
// If minSeverity is unknown, returns all findings.
func FilterFindingsBySeverity(findings []Finding, minSeverity string) []Finding {
	minLevel, ok := severityOrder[minSeverity]
	if !ok {
		// Unknown severity threshold - return all findings to be safe
		return findings
	}
	var filtered []Finding
	for _, f := range findings {
		level, ok := severityOrder[f.Severity]
		if !ok {
			// Include findings with unknown severity to avoid silent data loss
			filtered = append(filtered, f)
			continue
		}
		if level >= minLevel {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterFindingsForCodeReviewer filters findings relevant to code quality.
func FilterFindingsForCodeReviewer(findings []Finding) []Finding {
	var filtered []Finding
	for _, f := range findings {
		if codeQualityCategories[f.Category] {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterFindingsForSecurityReviewer filters findings relevant to security.
func FilterFindingsForSecurityReviewer(findings []Finding) []Finding {
	var filtered []Finding
	for _, f := range findings {
		if securityCategories[f.Category] {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterNilSourcesByRisk filters nil sources by risk level.
func FilterNilSourcesByRisk(sources []NilSource, risk string) []NilSource {
	var filtered []NilSource
	for _, s := range sources {
		if s.Risk == risk {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// FilterNilSourcesUnchecked filters nil sources that are not checked.
func FilterNilSourcesUnchecked(sources []NilSource) []NilSource {
	var filtered []NilSource
	for _, s := range sources {
		if !s.Checked {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// GetUncoveredFunctions returns functions with no test coverage.
func GetUncoveredFunctions(callGraph *CallGraphData) []FunctionCallGraph {
	if callGraph == nil {
		return nil
	}
	var uncovered []FunctionCallGraph
	for _, f := range callGraph.ModifiedFunctions {
		if len(f.TestCoverage) == 0 {
			uncovered = append(uncovered, f)
		}
	}
	return uncovered
}

// GetHighImpactFunctions returns functions with many callers.
func GetHighImpactFunctions(callGraph *CallGraphData, threshold int) []FunctionCallGraph {
	if callGraph == nil {
		return nil
	}
	var highImpact []FunctionCallGraph
	for _, f := range callGraph.ModifiedFunctions {
		if len(f.Callers) >= threshold {
			highImpact = append(highImpact, f)
		}
	}
	return highImpact
}
