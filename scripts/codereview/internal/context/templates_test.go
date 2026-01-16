package context

import (
	"strings"
	"testing"
)

// minTemplateLength is the minimum expected length for valid reviewer templates.
const minTemplateLength = 100

func TestRenderTemplate_CodeReviewer(t *testing.T) {
	data := &TemplateData{
		FindingCount: 2,
		Findings: []Finding{
			{Tool: "staticcheck", Rule: "SA1019", Severity: "warning", File: "user.go", Line: 45, Message: "deprecated API"},
			{Tool: "golangci-lint", Rule: "ineffassign", Severity: "info", File: "repo.go", Line: 23, Message: "unused var"},
		},
		HasSemanticChanges: true,
		ModifiedFunctions: []FunctionDiff{
			{
				Name:    "CreateUser",
				Package: "handler",
				File:    "user.go",
				Before:  FunctionInfo{Signature: "func CreateUser(ctx context.Context)", LineStart: 10, LineEnd: 20},
				After:   FunctionInfo{Signature: "func CreateUser(ctx context.Context, opts ...Option)", LineStart: 10, LineEnd: 30},
				Changes: []string{"added_param"},
			},
		},
		FocusAreas: []FocusArea{
			{Title: "Deprecated API", Description: "grpc.Dial needs update"},
		},
	}

	result, err := RenderTemplate(codeReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	// Verify key sections exist
	if !strings.Contains(result, "# Pre-Analysis Context: Code Quality") {
		t.Error("Missing title section")
	}
	if !strings.Contains(result, "Static Analysis Findings (2 issues)") {
		t.Error("Missing findings count")
	}
	if !strings.Contains(result, "staticcheck") {
		t.Error("Missing tool name in findings")
	}
	if !strings.Contains(result, "handler.CreateUser") {
		t.Error("Missing function name")
	}
	if !strings.Contains(result, "Deprecated API") {
		t.Error("Missing focus area")
	}
}

func TestRenderTemplate_SecurityReviewer(t *testing.T) {
	data := &TemplateData{
		FindingCount: 1,
		Findings: []Finding{
			{Tool: "gosec", Rule: "G401", Severity: "high", File: "crypto.go", Line: 23, Message: "weak crypto"},
		},
		HasDataFlowAnalysis: true,
		HighRiskFlows: []DataFlow{
			{
				ID:   "flow-1",
				Risk: "high",
				Source: FlowSource{
					Type:       "http_request",
					File:       "handler.go",
					Line:       45,
					Expression: "r.URL.Query().Get(\"id\")",
				},
				Sink: FlowSink{
					Type:       "database",
					File:       "repo.go",
					Line:       23,
					Expression: "db.Query(q, id)",
				},
				Sanitized: false,
				Notes:     "Query param without validation",
			},
		},
	}

	result, err := RenderTemplate(securityReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(result, "# Pre-Analysis Context: Security") {
		t.Error("Missing title section")
	}
	if !strings.Contains(result, "gosec") {
		t.Error("Missing security tool")
	}
	if !strings.Contains(result, "High Risk Flows") {
		t.Error("Missing high risk flows section")
	}
	if !strings.Contains(result, "http_request") {
		t.Error("Missing source type")
	}
}

func TestRenderTemplate_NilSafetyReviewer(t *testing.T) {
	data := &TemplateData{
		HasNilSources: true,
		NilSources: []NilSource{
			{Variable: "user", File: "handler.go", Line: 67, Checked: true, Risk: "low"},
			{Variable: "config", File: "service.go", Line: 23, Checked: false, Risk: "high", Notes: "env var may be empty"},
		},
		HighRiskNilSources: []NilSource{
			{Variable: "config", File: "service.go", Line: 23, Checked: false, Risk: "high", Expression: "os.Getenv(...)", Notes: "env var may be empty"},
		},
	}

	result, err := RenderTemplate(nilSafetyReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(result, "# Pre-Analysis Context: Nil Safety") {
		t.Error("Missing title section")
	}
	if !strings.Contains(result, "config") {
		t.Error("Missing variable name")
	}
	if !strings.Contains(result, "High Risk Nil Sources") {
		t.Error("Missing high risk section")
	}
}

func TestRenderTemplate_BusinessLogicReviewer(t *testing.T) {
	data := &TemplateData{
		HasCallGraph: true,
		HighImpactFunctions: []FunctionCallGraph{
			{
				Function: "ProcessPayment",
				File:     "payment.go",
				Callers: []CallSite{
					{Function: "HandleCheckout", File: "checkout.go", Line: 45},
					{Function: "HandleRefund", File: "refund.go", Line: 23},
				},
				Callees: []CallSite{
					{Function: "ValidateCard", File: "card.go", Line: 10},
				},
			},
		},
		HasSemanticChanges: true,
		ModifiedFunctions: []FunctionDiff{
			{
				Name:    "ProcessPayment",
				Package: "payment",
				Changes: []string{"body_changed", "error_handling_added"},
			},
		},
	}

	result, err := RenderTemplate(businessLogicReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(result, "# Pre-Analysis Context: Business Logic") {
		t.Error("Missing title section")
	}
	if !strings.Contains(result, "ProcessPayment") {
		t.Error("Missing function name")
	}
	if !strings.Contains(result, "HandleCheckout") {
		t.Error("Missing caller")
	}
}

func TestRenderTemplate_TestReviewer(t *testing.T) {
	data := &TemplateData{
		HasCallGraph: true,
		AllModifiedFunctionsGraph: []FunctionCallGraph{
			{
				Function: "CreateUser",
				File:     "user.go",
				TestCoverage: []TestCoverage{
					{TestFunction: "TestCreateUser", File: "user_test.go", Line: 10},
				},
			},
		},
		UncoveredFunctions: []FunctionCallGraph{
			{
				Function: "DeleteUser",
				File:     "user.go",
			},
		},
	}

	result, err := RenderTemplate(testReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(result, "# Pre-Analysis Context: Testing") {
		t.Error("Missing title section")
	}
	if !strings.Contains(result, "CreateUser") {
		t.Error("Missing covered function")
	}
	if !strings.Contains(result, "DeleteUser") {
		t.Error("Missing uncovered function")
	}
	if !strings.Contains(result, "No tests found") {
		t.Error("Missing no tests indicator")
	}
}

func TestGetTemplateForReviewer_ValidReviewers(t *testing.T) {
	tests := []struct {
		reviewer  string
		wantLen   int  // non-zero means template exists
		wantError bool // true if error expected
	}{
		{"code-reviewer", minTemplateLength, false},
		{"security-reviewer", minTemplateLength, false},
		{"business-logic-reviewer", minTemplateLength, false},
		{"test-reviewer", minTemplateLength, false},
		{"nil-safety-reviewer", minTemplateLength, false},
		{"unknown-reviewer", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.reviewer, func(t *testing.T) {
			tmpl, err := GetTemplateForReviewer(tt.reviewer)
			if tt.wantError {
				if err == nil {
					t.Errorf("GetTemplateForReviewer(%q) expected error, got nil", tt.reviewer)
				}
				return
			}
			if err != nil {
				t.Errorf("GetTemplateForReviewer(%q) unexpected error: %v", tt.reviewer, err)
				return
			}
			if len(tmpl) < tt.wantLen {
				t.Errorf("GetTemplateForReviewer(%q) returned len=%d, want len>=%d",
					tt.reviewer, len(tmpl), tt.wantLen)
			}
		})
	}
}

func TestRenderTemplate_EmptyData(t *testing.T) {
	data := &TemplateData{
		FindingCount:       0,
		Findings:           nil,
		HasSemanticChanges: false,
		FocusAreas:         nil,
	}

	result, err := RenderTemplate(codeReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed with empty data: %v", err)
	}

	if !strings.Contains(result, "No static analysis findings") {
		t.Error("Should show 'No static analysis findings' for empty findings")
	}
	if !strings.Contains(result, "No semantic changes detected") {
		t.Error("Should show 'No semantic changes detected' for empty changes")
	}
}

func TestTemplateFuncs_Inc(t *testing.T) {
	fn := templateFuncs["inc"].(func(int) int)
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"zero", 0, 1},
		{"positive", 5, 6},
		{"negative", -1, 0},
		{"large_negative", -100, -99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fn(tt.input); got != tt.expected {
				t.Errorf("inc(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTemplateFuncs_Join(t *testing.T) {
	fn := templateFuncs["join"].(func([]string, string) string)
	result := fn([]string{"a", "b", "c"}, ", ")
	if result != "a, b, c" {
		t.Errorf("join returned %q, want %q", result, "a, b, c")
	}
}

func TestTemplateFuncs_SignatureChanged(t *testing.T) {
	fn := templateFuncs["signatureChanged"].(func(FunctionDiff) bool)

	changed := FunctionDiff{
		Before: FunctionInfo{Signature: "func Foo()"},
		After:  FunctionInfo{Signature: "func Foo(x int)"},
	}
	if !fn(changed) {
		t.Error("signatureChanged should return true for different signatures")
	}

	unchanged := FunctionDiff{
		Before: FunctionInfo{Signature: "func Bar()"},
		After:  FunctionInfo{Signature: "func Bar()"},
	}
	if fn(unchanged) {
		t.Error("signatureChanged should return false for same signatures")
	}
}

func TestTemplateFuncs_RiskLevel(t *testing.T) {
	fn := templateFuncs["riskLevel"].(func(FunctionCallGraph) string)

	high := FunctionCallGraph{Callers: make([]CallSite, 5)}
	if fn(high) != "HIGH" {
		t.Error("5+ callers should be HIGH risk")
	}

	medium := FunctionCallGraph{Callers: make([]CallSite, 3)}
	if fn(medium) != "MEDIUM" {
		t.Error("2-4 callers should be MEDIUM risk")
	}

	low := FunctionCallGraph{Callers: make([]CallSite, 1)}
	if fn(low) != "LOW" {
		t.Error("0-1 callers should be LOW risk")
	}
}

func TestTemplateFuncs_TestStatus(t *testing.T) {
	fn := templateFuncs["testStatus"].(func(FunctionCallGraph) string)

	noTests := FunctionCallGraph{TestCoverage: nil}
	if fn(noTests) != "No tests" {
		t.Errorf("testStatus with no tests should return 'No tests', got %q", fn(noTests))
	}

	withTests := FunctionCallGraph{TestCoverage: []TestCoverage{{}, {}}}
	if fn(withTests) != "2 tests" {
		t.Errorf("testStatus with 2 tests should return '2 tests', got %q", fn(withTests))
	}
}

func TestTemplateFuncs_FieldChanges(t *testing.T) {
	fn := templateFuncs["fieldChanges"].(func(TypeDiff) []FieldChange)

	typeDiff := TypeDiff{
		Name: "User",
		Before: FieldsData{
			Fields: []FieldInfo{
				{Name: "ID", Type: "int"},
				{Name: "Name", Type: "string"},
				{Name: "Deleted", Type: "bool"},
			},
		},
		After: FieldsData{
			Fields: []FieldInfo{
				{Name: "ID", Type: "int64"},     // modified
				{Name: "Name", Type: "string"},  // unchanged (should not appear)
				{Name: "Email", Type: "string"}, // added
			},
		},
	}

	changes := fn(typeDiff)

	// Should have: ID modified, Email added, Deleted deleted = 3 changes
	if len(changes) != 3 {
		t.Errorf("fieldChanges returned %d changes, want 3", len(changes))
	}

	// Check for modified field
	foundModified := false
	foundAdded := false
	foundDeleted := false

	for _, c := range changes {
		if c.Name == "ID" && c.Before == "int" && c.After == "int64" {
			foundModified = true
		}
		if c.Name == "Email" && c.Before == "-" && strings.Contains(c.After, "added") {
			foundAdded = true
		}
		if c.Name == "Deleted" && c.After == "(deleted)" {
			foundDeleted = true
		}
	}

	if !foundModified {
		t.Error("fieldChanges should detect modified field 'ID'")
	}
	if !foundAdded {
		t.Error("fieldChanges should detect added field 'Email'")
	}
	if !foundDeleted {
		t.Error("fieldChanges should detect deleted field 'Deleted'")
	}
}

func TestRenderTemplate_InvalidTemplate(t *testing.T) {
	invalidTemplate := `{{.InvalidFunc | nonexistent}}`
	data := &TemplateData{}

	_, err := RenderTemplate(invalidTemplate, data)
	if err == nil {
		t.Error("RenderTemplate should return error for invalid template")
	}
}

func TestRenderTemplate_CodeReviewerWithSignatureChange(t *testing.T) {
	data := &TemplateData{
		FindingCount:       0,
		HasSemanticChanges: true,
		ModifiedFunctions: []FunctionDiff{
			{
				Name:    "Process",
				Package: "service",
				File:    "service.go",
				Before:  FunctionInfo{Signature: "func Process(ctx context.Context)", LineStart: 10, LineEnd: 20},
				After:   FunctionInfo{Signature: "func Process(ctx context.Context, id string)", LineStart: 10, LineEnd: 25},
				Changes: []string{"added_param"},
			},
		},
	}

	result, err := RenderTemplate(codeReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	// Verify the diff block appears for signature change
	if !strings.Contains(result, "```diff") {
		t.Error("Should show diff block for signature change")
	}
	if !strings.Contains(result, "- func Process(ctx context.Context)") {
		t.Error("Should show old signature in diff")
	}
	if !strings.Contains(result, "+ func Process(ctx context.Context, id string)") {
		t.Error("Should show new signature in diff")
	}
}

func TestRenderTemplate_SecurityReviewerWithMediumRiskFlows(t *testing.T) {
	data := &TemplateData{
		FindingCount:        0,
		HasDataFlowAnalysis: true,
		MediumRiskFlows: []DataFlow{
			{
				ID:   "flow-2",
				Risk: "medium",
				Source: FlowSource{
					Type: "config_file",
					File: "config.go",
					Line: 10,
				},
				Sink: FlowSink{
					Type: "logger",
				},
				Notes: "Config value logged",
			},
		},
	}

	result, err := RenderTemplate(securityReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(result, "Medium Risk Flows") {
		t.Error("Should show medium risk flows section")
	}
	if !strings.Contains(result, "config_file") {
		t.Error("Should show medium risk flow source type")
	}
}

func TestRenderTemplate_NilSafetyWithCheckedSource(t *testing.T) {
	data := &TemplateData{
		HasNilSources: true,
		NilSources: []NilSource{
			{Variable: "result", File: "handler.go", Line: 50, Checked: true, Risk: "low", CheckLine: 52},
		},
		HighRiskNilSources: []NilSource{
			{Variable: "result", File: "handler.go", Line: 50, Checked: true, Risk: "high", Expression: "db.Find(...)", CheckLine: 52, Notes: "Proper nil check exists"},
		},
	}

	result, err := RenderTemplate(nilSafetyReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(result, "Yes") {
		t.Error("Should show 'Yes' for checked nil source in table")
	}
	if !strings.Contains(result, "line 52") {
		t.Error("Should show check line number for checked source")
	}
}

func TestRenderTemplate_TestReviewerAllCovered(t *testing.T) {
	data := &TemplateData{
		HasCallGraph:       true,
		UncoveredFunctions: nil, // All functions covered
	}

	result, err := RenderTemplate(testReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(result, "All modified code has test coverage") {
		t.Error("Should show 'All modified code has test coverage' when no uncovered functions")
	}
}

func TestRenderTemplate_BusinessLogicNoCallGraph(t *testing.T) {
	data := &TemplateData{
		HasCallGraph: false,
	}

	result, err := RenderTemplate(businessLogicReviewerTemplate, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	if !strings.Contains(result, "No call graph analysis available") {
		t.Error("Should show 'No call graph analysis available' when HasCallGraph is false")
	}
}
