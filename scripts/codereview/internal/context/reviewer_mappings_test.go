package context

import (
	"reflect"
	"testing"
)

func TestReviewerNames(t *testing.T) {
	expected := []string{
		"code-reviewer",
		"security-reviewer",
		"business-logic-reviewer",
		"test-reviewer",
		"nil-safety-reviewer",
	}

	names := GetReviewerNames()
	if !reflect.DeepEqual(names, expected) {
		t.Errorf("GetReviewerNames() = %v, want %v", names, expected)
	}
}

func TestGetReviewerDataSources(t *testing.T) {
	tests := []struct {
		reviewer string
		wantLen  int
	}{
		{"code-reviewer", 2},           // static-analysis, semantic-diff
		{"security-reviewer", 2},       // static-analysis (security), security-summary
		{"business-logic-reviewer", 2}, // semantic-diff, impact-summary
		{"test-reviewer", 1},           // impact-summary (test coverage)
		{"nil-safety-reviewer", 1},     // data-flow (nil_sources)
		{"unknown-reviewer", 0},        // unknown reviewer returns empty slice
	}

	for _, tt := range tests {
		t.Run(tt.reviewer, func(t *testing.T) {
			sources := GetReviewerDataSources(tt.reviewer)
			if len(sources) != tt.wantLen {
				t.Errorf("GetReviewerDataSources(%q) returned %d sources, want %d",
					tt.reviewer, len(sources), tt.wantLen)
			}
		})
	}
}

func TestFilterFindingsByCategory(t *testing.T) {
	findings := []Finding{
		{Category: "security", Severity: "high", Message: "SQL injection"},
		{Category: "style", Severity: "info", Message: "line too long"},
		{Category: "bug", Severity: "warning", Message: "unused var"},
		{Category: "security", Severity: "warning", Message: "weak hash"},
	}

	securityFindings := FilterFindingsByCategory(findings, "security")
	if len(securityFindings) != 2 {
		t.Errorf("FilterFindingsByCategory(security) returned %d, want 2", len(securityFindings))
	}

	styleFindings := FilterFindingsByCategory(findings, "style")
	if len(styleFindings) != 1 {
		t.Errorf("FilterFindingsByCategory(style) returned %d, want 1", len(styleFindings))
	}
}

func TestFilterFindingsBySeverity(t *testing.T) {
	findings := []Finding{
		{Severity: "critical", Message: "crash"},
		{Severity: "high", Message: "security"},
		{Severity: "warning", Message: "style"},
		{Severity: "info", Message: "hint"},
		{Severity: "high", Message: "another"},
	}

	// Filter high and above
	highAndAbove := FilterFindingsBySeverity(findings, "high")
	if len(highAndAbove) != 3 { // critical + 2 high
		t.Errorf("FilterFindingsBySeverity(high) returned %d, want 3", len(highAndAbove))
	}
}

func TestFilterFindingsBySeverity_UnknownThreshold(t *testing.T) {
	findings := []Finding{
		{Severity: "critical", Message: "crash"},
		{Severity: "info", Message: "hint"},
	}

	// Unknown severity threshold should return all findings
	result := FilterFindingsBySeverity(findings, "unknown")
	if len(result) != 2 {
		t.Errorf("FilterFindingsBySeverity(unknown) returned %d, want 2 (all findings)", len(result))
	}
}

func TestFilterNilSourcesByRisk(t *testing.T) {
	sources := []NilSource{
		{Variable: "a", Risk: "high", Checked: false},
		{Variable: "b", Risk: "low", Checked: true},
		{Variable: "c", Risk: "high", Checked: false},
		{Variable: "d", Risk: "medium", Checked: false},
	}

	highRisk := FilterNilSourcesByRisk(sources, "high")
	if len(highRisk) != 2 {
		t.Errorf("FilterNilSourcesByRisk(high) returned %d, want 2", len(highRisk))
	}
}

func TestFilterNilSourcesUnchecked(t *testing.T) {
	sources := []NilSource{
		{Variable: "a", Risk: "high", Checked: false},
		{Variable: "b", Risk: "low", Checked: true},
		{Variable: "c", Risk: "high", Checked: false},
		{Variable: "d", Risk: "medium", Checked: true},
	}

	unchecked := FilterNilSourcesUnchecked(sources)
	if len(unchecked) != 2 {
		t.Errorf("FilterNilSourcesUnchecked() returned %d, want 2", len(unchecked))
	}
}

func TestFilterFindingsForCodeReviewer(t *testing.T) {
	findings := []Finding{
		{Category: "security", Severity: "high", Message: "SQL injection"},
		{Category: "style", Severity: "info", Message: "line too long"},
		{Category: "bug", Severity: "warning", Message: "unused var"},
		{Category: "performance", Severity: "warning", Message: "slow loop"},
	}

	codeFindings := FilterFindingsForCodeReviewer(findings)
	// Should include style, bug, performance but not security
	if len(codeFindings) != 3 {
		t.Errorf("FilterFindingsForCodeReviewer() returned %d, want 3", len(codeFindings))
	}
}

func TestFilterFindingsForSecurityReviewer(t *testing.T) {
	findings := []Finding{
		{Category: "security", Severity: "high", Message: "SQL injection"},
		{Category: "style", Severity: "info", Message: "line too long"},
		{Category: "bug", Severity: "warning", Message: "unused var"},
		{Category: "security", Severity: "warning", Message: "weak hash"},
	}

	securityFindings := FilterFindingsForSecurityReviewer(findings)
	if len(securityFindings) != 2 {
		t.Errorf("FilterFindingsForSecurityReviewer() returned %d, want 2", len(securityFindings))
	}
}

func TestGetUncoveredFunctions(t *testing.T) {
	callGraph := &CallGraphData{
		ModifiedFunctions: []FunctionCallGraph{
			{
				Function:     "funcA",
				File:         "a.go",
				TestCoverage: []TestCoverage{{TestFunction: "TestA", File: "a_test.go", Line: 10}},
			},
			{
				Function:     "funcB",
				File:         "b.go",
				TestCoverage: nil,
			},
			{
				Function:     "funcC",
				File:         "c.go",
				TestCoverage: []TestCoverage{},
			},
		},
	}

	uncovered := GetUncoveredFunctions(callGraph)
	if len(uncovered) != 2 {
		t.Errorf("GetUncoveredFunctions() returned %d, want 2", len(uncovered))
	}
}

func TestGetHighImpactFunctions(t *testing.T) {
	callGraph := &CallGraphData{
		ModifiedFunctions: []FunctionCallGraph{
			{
				Function: "funcA",
				File:     "a.go",
				Callers:  []CallSite{{Function: "caller1"}, {Function: "caller2"}, {Function: "caller3"}},
			},
			{
				Function: "funcB",
				File:     "b.go",
				Callers:  []CallSite{{Function: "caller1"}},
			},
			{
				Function: "funcC",
				File:     "c.go",
				Callers:  []CallSite{{Function: "caller1"}, {Function: "caller2"}, {Function: "caller3"}, {Function: "caller4"}},
			},
		},
	}

	highImpact := GetHighImpactFunctions(callGraph, 3)
	if len(highImpact) != 2 {
		t.Errorf("GetHighImpactFunctions(threshold=3) returned %d, want 2", len(highImpact))
	}
}

func TestGetUncoveredFunctions_NilInput(t *testing.T) {
	result := GetUncoveredFunctions(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

func TestGetHighImpactFunctions_NilInput(t *testing.T) {
	result := GetHighImpactFunctions(nil, 3)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}
