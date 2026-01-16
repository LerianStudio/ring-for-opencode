// Package callgraph provides call graph analysis for multiple languages.
package callgraph

// CallInfo represents a single caller or callee relationship.
type CallInfo struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	CallSite string `json:"call_site,omitempty"`
}

// TestCoverage represents a test that covers a function.
type TestCoverage struct {
	TestFunction string `json:"test_function"`
	File         string `json:"file"`
	Line         int    `json:"line"`
}

// FunctionCallGraph represents the call graph for a single modified function.
type FunctionCallGraph struct {
	Function     string         `json:"function"`
	File         string         `json:"file"`
	Callers      []CallInfo     `json:"callers"`
	Callees      []CallInfo     `json:"callees"`
	TestCoverage []TestCoverage `json:"test_coverage"`
}

// ImpactAnalysis summarizes the impact of all changes.
type ImpactAnalysis struct {
	DirectCallers     int      `json:"direct_callers"`
	TransitiveCallers int      `json:"transitive_callers"`
	AffectedTests     int      `json:"affected_tests"`
	AffectedPackages  []string `json:"affected_packages"`
}

// CallGraphResult is the output schema for call graph analysis.
type CallGraphResult struct {
	Language           string              `json:"language"`
	ModifiedFunctions  []FunctionCallGraph `json:"modified_functions"`
	ImpactAnalysis     ImpactAnalysis      `json:"impact_analysis"`
	TimeBudgetExceeded bool                `json:"time_budget_exceeded,omitempty"`
	PartialResults     bool                `json:"partial_results,omitempty"`
	Warnings           []string            `json:"warnings,omitempty"`
}

// ModifiedFunction represents a function from AST analysis (input).
type ModifiedFunction struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Package  string `json:"package"`
	Receiver string `json:"receiver,omitempty"`
}

// Analyzer defines the interface for language-specific call graph analyzers.
type Analyzer interface {
	// Analyze builds call graph for the given modified functions.
	// timeBudgetSec is the maximum time allowed (0 = no limit).
	Analyze(modifiedFuncs []ModifiedFunction, timeBudgetSec int) (*CallGraphResult, error)
}
