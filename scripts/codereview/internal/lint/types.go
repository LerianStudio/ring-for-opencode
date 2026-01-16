// Package lint provides linter integrations for static analysis.
package lint

import "fmt"

// Severity represents the severity level of a finding.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// Category represents the category of a finding.
type Category string

const (
	CategorySecurity    Category = "security"
	CategoryBug         Category = "bug"
	CategoryStyle       Category = "style"
	CategoryPerformance Category = "performance"
	CategoryDeprecation Category = "deprecation"
	CategoryComplexity  Category = "complexity"
	CategoryType        Category = "type"
	CategoryUnused      Category = "unused"
	CategoryOther       Category = "other"
)

// Finding represents a single lint finding.
type Finding struct {
	Tool       string   `json:"tool"`
	Rule       string   `json:"rule"`
	Severity   Severity `json:"severity"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Column     int      `json:"column"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion,omitempty"`
	Category   Category `json:"category"`
}

// ToolVersions holds version information for all tools used.
type ToolVersions map[string]string

// Summary holds aggregated finding counts by severity.
type Summary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Warning  int `json:"warning"`
	Info     int `json:"info"`
	Unknown  int `json:"unknown"`
}

// Result is the aggregate output of static analysis.
type Result struct {
	ToolVersions ToolVersions `json:"tool_versions"`
	Findings     []Finding    `json:"findings"`
	Summary      Summary      `json:"summary"`
	Errors       []string     `json:"errors,omitempty"`
}

// NewResult creates a new Result with initialized fields.
func NewResult() *Result {
	return &Result{
		ToolVersions: make(ToolVersions),
		Findings:     make([]Finding, 0),
		Summary:      Summary{},
		Errors:       make([]string, 0),
	}
}

// AddFinding adds a finding and updates the summary.
func (r *Result) AddFinding(f Finding) {
	r.Findings = append(r.Findings, f)
	switch f.Severity {
	case SeverityCritical:
		r.Summary.Critical++
	case SeverityHigh:
		r.Summary.High++
	case SeverityWarning:
		r.Summary.Warning++
	case SeverityInfo:
		r.Summary.Info++
	default:
		r.Summary.Unknown++
		r.Errors = append(r.Errors, fmt.Sprintf("unknown severity %q for finding %s:%d (%s)", f.Severity, f.File, f.Line, f.Message))
	}
}

// Merge combines another Result into this one.
func (r *Result) Merge(other *Result) {
	if other == nil {
		return
	}
	for k, v := range other.ToolVersions {
		r.ToolVersions[k] = v
	}
	for _, f := range other.Findings {
		r.AddFinding(f)
	}
	r.Errors = append(r.Errors, other.Errors...)
}

// FilterByFiles returns findings only for the specified files.
func (r *Result) FilterByFiles(files map[string]bool) *Result {
	filtered := NewResult()
	for k, v := range r.ToolVersions {
		filtered.ToolVersions[k] = v
	}
	filtered.Errors = append(filtered.Errors, r.Errors...)

	for _, f := range r.Findings {
		if files[f.File] {
			filtered.AddFinding(f)
		}
	}
	return filtered
}
