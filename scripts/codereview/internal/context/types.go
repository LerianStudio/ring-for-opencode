// Package context provides context compilation for code review.
// It aggregates outputs from all analysis phases and generates
// reviewer-specific context files.
package context

// PhaseOutputs holds all the outputs from analysis phases.
type PhaseOutputs struct {
	Scope          *ScopeData          `json:"scope"`
	StaticAnalysis *StaticAnalysisData `json:"static_analysis"`
	AST            *ASTData            `json:"ast"`
	CallGraph      *CallGraphData      `json:"call_graph"`
	DataFlow       *DataFlowData       `json:"data_flow"`
	Errors         []string            `json:"errors,omitempty"`

	// Multi-language support: maps language code to data
	ASTByLanguage       map[string]*ASTData       `json:"ast_by_language,omitempty"`
	CallGraphByLanguage map[string]*CallGraphData `json:"call_graph_by_language,omitempty"`
	DataFlowByLanguage  map[string]*DataFlowData  `json:"data_flow_by_language,omitempty"`
}

// ScopeData represents Phase 0 output (scope.json).
type ScopeData struct {
	BaseRef          string   `json:"base_ref"`
	HeadRef          string   `json:"head_ref"`
	Language         string   `json:"language"`
	ModifiedFiles    []string `json:"modified"`
	AddedFiles       []string `json:"added"`
	DeletedFiles     []string `json:"deleted"`
	TotalFiles       int      `json:"total_files"`
	TotalAdditions   int      `json:"total_additions"`
	TotalDeletions   int      `json:"total_deletions"`
	PackagesAffected []string `json:"packages_affected"`
}

// StaticAnalysisData represents Phase 1 output (static-analysis.json).
type StaticAnalysisData struct {
	ToolVersions map[string]string `json:"tool_versions"`
	Findings     []Finding         `json:"findings"`
	Summary      FindingSummary    `json:"summary"`
	Errors       []string          `json:"errors,omitempty"`
}

// Finding represents a single lint/analysis finding.
type Finding struct {
	Tool       string `json:"tool"`
	Rule       string `json:"rule"`
	Severity   string `json:"severity"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Category   string `json:"category"`
}

// FindingSummary holds counts by severity.
type FindingSummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Warning  int `json:"warning"`
	Info     int `json:"info"`
}

// ASTData represents Phase 2 output ({lang}-ast.json).
type ASTData struct {
	Functions     FunctionChanges   `json:"functions"`
	Types         TypeChanges       `json:"types"`
	Interfaces    InterfaceChanges  `json:"interfaces,omitempty"`
	Classes       ClassChanges      `json:"classes,omitempty"`
	Imports       ImportChanges     `json:"imports"`
	ErrorHandling ErrorHandlingData `json:"error_handling,omitempty"`
}

// FunctionChanges holds function-level changes.
type FunctionChanges struct {
	Modified []FunctionDiff `json:"modified"`
	Added    []FunctionInfo `json:"added"`
	Deleted  []FunctionInfo `json:"deleted"`
}

// FunctionDiff represents a modified function.
type FunctionDiff struct {
	Name     string       `json:"name"`
	File     string       `json:"file"`
	Package  string       `json:"package,omitempty"`
	Module   string       `json:"module,omitempty"`
	Receiver string       `json:"receiver,omitempty"`
	Before   FunctionInfo `json:"before"`
	After    FunctionInfo `json:"after"`
	Changes  []string     `json:"changes"`
}

// FunctionInfo holds function details.
type FunctionInfo struct {
	Name      string           `json:"name,omitempty"`
	Signature string           `json:"signature"`
	File      string           `json:"file,omitempty"`
	Package   string           `json:"package,omitempty"`
	LineStart int              `json:"line_start"`
	LineEnd   int              `json:"line_end"`
	Params    []ParamInfo      `json:"params,omitempty"`
	Returns   []ReturnTypeInfo `json:"returns,omitempty"`
}

// ParamInfo holds parameter details.
type ParamInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
}

// ReturnTypeInfo holds return type information for functions.
type ReturnTypeInfo struct {
	Type string `json:"type"`
}

// TypeChanges holds struct/type changes.
type TypeChanges struct {
	Modified []TypeDiff       `json:"modified"`
	Added    []ReturnTypeInfo `json:"added"`
	Deleted  []ReturnTypeInfo `json:"deleted"`
}

// TypeDiff represents a modified type.
type TypeDiff struct {
	Name    string     `json:"name"`
	Kind    string     `json:"kind"`
	File    string     `json:"file"`
	Before  FieldsData `json:"before"`
	After   FieldsData `json:"after"`
	Changes []string   `json:"changes"`
}

// FieldsData holds struct field information.
type FieldsData struct {
	Fields []FieldInfo `json:"fields"`
}

// FieldInfo represents a struct field.
type FieldInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tags string `json:"tags,omitempty"`
}

// InterfaceChanges holds interface changes (Go/TS).
type InterfaceChanges struct {
	Modified []InterfaceDiff `json:"modified"`
	Added    []InterfaceInfo `json:"added"`
	Deleted  []InterfaceInfo `json:"deleted"`
}

// InterfaceDiff represents a modified interface.
type InterfaceDiff struct {
	Name    string        `json:"name"`
	File    string        `json:"file"`
	Before  InterfaceInfo `json:"before"`
	After   InterfaceInfo `json:"after"`
	Changes []string      `json:"changes"`
}

// InterfaceInfo holds interface details.
type InterfaceInfo struct {
	Name    string   `json:"name,omitempty"`
	Methods []string `json:"methods"`
}

// ClassChanges holds class changes (TS/Python).
type ClassChanges struct {
	Modified []ClassDiff `json:"modified"`
	Added    []ClassInfo `json:"added"`
	Deleted  []ClassInfo `json:"deleted"`
}

// ClassDiff represents a modified class.
type ClassDiff struct {
	Name    string    `json:"name"`
	File    string    `json:"file"`
	Before  ClassInfo `json:"before"`
	After   ClassInfo `json:"after"`
	Changes []string  `json:"changes"`
}

// ClassInfo holds class details.
type ClassInfo struct {
	Name       string   `json:"name,omitempty"`
	Methods    []string `json:"methods"`
	Attributes []string `json:"attributes,omitempty"`
}

// ImportChanges holds import changes.
type ImportChanges struct {
	Added   []ImportInfo `json:"added"`
	Removed []ImportInfo `json:"removed"`
}

// ImportInfo represents an import statement.
type ImportInfo struct {
	File   string   `json:"file"`
	Path   string   `json:"path,omitempty"`
	Module string   `json:"module,omitempty"`
	Names  []string `json:"names,omitempty"`
}

// ErrorHandlingData holds error handling changes (Go).
type ErrorHandlingData struct {
	NewErrorReturns    []ErrorReturn `json:"new_error_returns"`
	RemovedErrorChecks []ErrorCheck  `json:"removed_error_checks"`
}

// ErrorReturn represents a new error return path.
type ErrorReturn struct {
	Function  string `json:"function"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	ErrorType string `json:"error_type"`
	Message   string `json:"message"`
}

// ErrorCheck represents a removed error check.
type ErrorCheck struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// CallGraphData represents Phase 3 output ({lang}-calls.json).
type CallGraphData struct {
	ModifiedFunctions []FunctionCallGraph `json:"modified_functions"`
	ImpactAnalysis    ImpactSummary       `json:"impact_analysis"`
}

// FunctionCallGraph holds call graph data for a function.
type FunctionCallGraph struct {
	Function     string         `json:"function"`
	File         string         `json:"file"`
	Callers      []CallSite     `json:"callers"`
	Callees      []CallSite     `json:"callees"`
	TestCoverage []TestCoverage `json:"test_coverage"`
}

// CallSite represents a call site.
type CallSite struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	CallSite string `json:"call_site,omitempty"`
}

// TestCoverage represents test coverage for a function.
type TestCoverage struct {
	TestFunction string `json:"test_function"`
	File         string `json:"file"`
	Line         int    `json:"line"`
}

// ImpactSummary holds impact analysis summary.
type ImpactSummary struct {
	DirectCallers     int      `json:"direct_callers"`
	TransitiveCallers int      `json:"transitive_callers"`
	AffectedTests     int      `json:"affected_tests"`
	AffectedPackages  []string `json:"affected_packages,omitempty"`
	AffectedModules   []string `json:"affected_modules,omitempty"`
}

// DataFlowData represents Phase 4 output ({lang}-flow.json).
type DataFlowData struct {
	Flows      []DataFlow  `json:"flows"`
	NilSources []NilSource `json:"nil_sources,omitempty"`
	Summary    FlowSummary `json:"summary"`
}

// DataFlow represents a data flow path.
type DataFlow struct {
	ID        string     `json:"id"`
	Source    FlowSource `json:"source"`
	Path      []FlowStep `json:"path"`
	Sink      FlowSink   `json:"sink"`
	Sanitized bool       `json:"sanitized"`
	Risk      string     `json:"risk"`
	Notes     string     `json:"notes,omitempty"`
}

// FlowSource represents where data enters.
type FlowSource struct {
	Type       string `json:"type"`
	Subtype    string `json:"subtype,omitempty"`
	Framework  string `json:"framework,omitempty"`
	Variable   string `json:"variable"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Expression string `json:"expression"`
}

// FlowStep represents a step in data flow.
type FlowStep struct {
	Step          int    `json:"step"`
	File          string `json:"file"`
	Line          int    `json:"line"`
	Expression    string `json:"expression"`
	Operation     string `json:"operation"`
	SanitizerType string `json:"sanitizer_type,omitempty"`
}

// FlowSink represents where data exits.
type FlowSink struct {
	Type       string `json:"type"`
	Subtype    string `json:"subtype,omitempty"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Expression string `json:"expression"`
}

// NilSource represents a potential nil source.
type NilSource struct {
	Variable        string `json:"variable"`
	File            string `json:"file"`
	Line            int    `json:"line"`
	Expression      string `json:"expression"`
	Checked         bool   `json:"checked"`
	CheckLine       int    `json:"check_line,omitempty"`
	CheckExpression string `json:"check_expression,omitempty"`
	Risk            string `json:"risk,omitempty"`
	Notes           string `json:"notes,omitempty"`
}

// FlowSummary holds data flow summary.
type FlowSummary struct {
	TotalFlows       int `json:"total_flows"`
	SanitizedFlows   int `json:"sanitized_flows"`
	UnsanitizedFlows int `json:"unsanitized_flows"`
	HighRisk         int `json:"high_risk"`
	MediumRisk       int `json:"medium_risk"`
	LowRisk          int `json:"low_risk"`
	NilRisks         int `json:"nil_risks,omitempty"`
	NoneRisks        int `json:"none_risks,omitempty"`
}

// ReviewerContext holds the compiled context for a specific reviewer.
type ReviewerContext struct {
	ReviewerName string `json:"reviewer_name"`
	Title        string `json:"title"`
	Content      string `json:"content"`
}
