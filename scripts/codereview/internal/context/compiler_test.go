package context

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompiler_Compile(t *testing.T) {
	// Create temp directories
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create sample phase outputs
	createSamplePhaseOutputs(t, inputDir)

	// Create compiler and run
	compiler := NewCompiler(inputDir, outputDir)
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Verify all context files were created
	reviewers := GetReviewerNames()
	for _, reviewer := range reviewers {
		contextPath := filepath.Join(outputDir, "context-"+reviewer+".md")
		if _, err := os.Stat(contextPath); os.IsNotExist(err) {
			t.Errorf("Context file not created for %s", reviewer)
		}

		// Read and verify content
		content, err := os.ReadFile(contextPath)
		if err != nil {
			t.Errorf("Failed to read context file for %s: %v", reviewer, err)
			continue
		}

		// Verify structure - context files should have required sections
		contentStr := string(content)

		// Verify required sections exist (all reviewers have these)
		requiredSections := []string{
			"# Pre-Analysis Context:",
			"## Focus Areas",
		}
		for _, section := range requiredSections {
			if !strings.Contains(contentStr, section) {
				t.Errorf("Context file for %s missing required section: %s", reviewer, section)
			}
		}

		// Verify minimum content length (ensures non-trivial output)
		if len(contentStr) < 100 {
			t.Errorf("Context file for %s is too short (%d bytes), expected substantial content", reviewer, len(contentStr))
		}
	}
}

func TestCompiler_ReviewerContexts(t *testing.T) {
	tests := []struct {
		name             string
		reviewer         string
		expectedTitle    string
		expectedSections []string
	}{
		{
			name:             "code reviewer has quality sections",
			reviewer:         "code-reviewer",
			expectedTitle:    "Code Quality",
			expectedSections: []string{"Static Analysis Findings"},
		},
		{
			name:             "security reviewer has security sections",
			reviewer:         "security-reviewer",
			expectedTitle:    "Security",
			expectedSections: []string{"Data Flow Analysis"},
		},
		{
			name:             "business logic reviewer has impact sections",
			reviewer:         "business-logic-reviewer",
			expectedTitle:    "Business Logic",
			expectedSections: []string{"Impact Analysis"},
		},
		{
			name:             "test reviewer has coverage sections",
			reviewer:         "test-reviewer",
			expectedTitle:    "Testing",
			expectedSections: []string{"Test Coverage"},
		},
		{
			name:             "nil safety reviewer has nil analysis sections",
			reviewer:         "nil-safety-reviewer",
			expectedTitle:    "Nil Safety",
			expectedSections: []string{"Nil Source Analysis"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputDir := t.TempDir()
			outputDir := t.TempDir()
			createSamplePhaseOutputs(t, inputDir)

			compiler := NewCompiler(inputDir, outputDir)
			if err := compiler.Compile(); err != nil {
				t.Fatalf("Compile() error = %v", err)
			}

			content, err := os.ReadFile(filepath.Join(outputDir, "context-"+tt.reviewer+".md"))
			if err != nil {
				t.Fatalf("Failed to read %s context: %v", tt.reviewer, err)
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, tt.expectedTitle) {
				t.Errorf("Missing %q title", tt.expectedTitle)
			}
			for _, section := range tt.expectedSections {
				if !strings.Contains(contentStr, section) {
					t.Errorf("Missing section: %s", section)
				}
			}
		})
	}
}

func TestCompiler_MissingInputs(t *testing.T) {
	// Test with empty input directory
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	compiler := NewCompiler(inputDir, outputDir)

	// Should not fail, just produce minimal output
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() should not fail with missing inputs: %v", err)
	}

	// Context files should still be created (with "no data" messages)
	contextPath := filepath.Join(outputDir, "context-code-reviewer.md")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		t.Error("Context file should be created even with missing inputs")
	}
}

func TestCompiler_PartialInputs(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create only scope.json
	scope := ScopeData{
		BaseRef:  "main",
		HeadRef:  "HEAD",
		Language: "go",
		Files: ScopeFiles{
			Modified: []string{},
			Added:    []string{},
			Deleted:  []string{},
		},
		Stats: ScopeStats{},
	}
	scopeData, err := json.Marshal(scope)
	if err != nil {
		t.Fatalf("failed to marshal scope: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "scope.json"), scopeData, 0644); err != nil {
		t.Fatalf("Failed to write scope.json: %v", err)
	}

	compiler := NewCompiler(inputDir, outputDir)
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() error with partial inputs: %v", err)
	}

	// Verify files were created
	for _, reviewer := range GetReviewerNames() {
		contextPath := filepath.Join(outputDir, "context-"+reviewer+".md")
		if _, err := os.Stat(contextPath); os.IsNotExist(err) {
			t.Errorf("Context file not created for %s with partial inputs", reviewer)
		}
	}
}

func TestCompiler_SameInputOutputDir(t *testing.T) {
	dir := t.TempDir()
	createSamplePhaseOutputs(t, dir)

	compiler := NewCompiler(dir, dir)
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() error with same input/output dir: %v", err)
	}

	// Verify context files exist alongside input files
	if _, err := os.Stat(filepath.Join(dir, "scope.json")); os.IsNotExist(err) {
		t.Error("Original scope.json should still exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "context-code-reviewer.md")); os.IsNotExist(err) {
		t.Error("Context file should be created")
	}
}

func TestCompiler_FocusAreasGenerated(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	createSamplePhaseOutputs(t, inputDir)

	compiler := NewCompiler(inputDir, outputDir)
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	// Check code reviewer focus areas (should have deprecation and signature change)
	content, err := os.ReadFile(filepath.Join(outputDir, "context-code-reviewer.md"))
	if err != nil {
		t.Fatalf("failed to read code-reviewer context: %v", err)
	}
	contentStr := string(content)
	if !strings.Contains(contentStr, "Focus Areas") {
		t.Error("Missing Focus Areas section in code-reviewer context")
	}
	if !strings.Contains(contentStr, "Deprecated API Usage") {
		t.Error("Missing deprecation focus area")
	}
	if !strings.Contains(contentStr, "Signature change in CreateUser") {
		t.Error("Missing signature change focus area")
	}
}

func TestCompiler_HighRiskNilSources(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	createSamplePhaseOutputs(t, inputDir)

	compiler := NewCompiler(inputDir, outputDir)
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "context-nil-safety-reviewer.md"))
	if err != nil {
		t.Fatalf("failed to read nil-safety-reviewer context: %v", err)
	}
	contentStr := string(content)

	// Should have high risk nil sources section
	if !strings.Contains(contentStr, "High Risk Nil Sources") {
		t.Error("Missing High Risk Nil Sources section")
	}
	// Should contain the high-risk config variable
	if !strings.Contains(contentStr, "config") {
		t.Error("Missing high-risk 'config' variable")
	}
}

func TestCompiler_UncoveredFunctions(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create phase outputs with an uncovered function
	scope := ScopeData{
		BaseRef:  "main",
		HeadRef:  "HEAD",
		Language: "go",
		Files: ScopeFiles{
			Modified: []string{},
			Added:    []string{},
			Deleted:  []string{},
		},
		Stats: ScopeStats{},
	}
	writeJSON(t, inputDir, "scope.json", scope)

	calls := CallGraphData{
		ModifiedFunctions: []FunctionCallGraph{
			{Function: "handler.CreateUser", File: "user.go", Callers: []CallSite{{Function: "router.ServeHTTP"}}, TestCoverage: nil}, // No tests
		},
	}
	writeJSON(t, inputDir, "go-calls.json", calls)

	compiler := NewCompiler(inputDir, outputDir)
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "context-test-reviewer.md"))
	if err != nil {
		t.Fatalf("failed to read test-reviewer context: %v", err)
	}
	contentStr := string(content)

	if !strings.Contains(contentStr, "Uncovered") {
		t.Error("Missing uncovered functions mention")
	}
}

func TestCompiler_InvalidJSONHandling(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(inputDir, "scope.json"), []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid scope.json: %v", err)
	}

	compiler := NewCompiler(inputDir, outputDir)
	// Should not fail, but should continue with other files
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() should handle invalid JSON gracefully: %v", err)
	}

	// Context files should still be created
	contextPath := filepath.Join(outputDir, "context-code-reviewer.md")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		t.Error("Context file should be created even with invalid JSON")
	}

	// Verify context file indicates parse error was handled
	// The compiler should produce output even when parsing fails
	content, err := os.ReadFile(contextPath)
	if err != nil {
		t.Fatalf("Failed to read context file: %v", err)
	}

	// Context file should still have structure (with "no data" messages for missing data)
	contentStr := string(content)
	if !strings.Contains(contentStr, "# Pre-Analysis Context") {
		t.Error("Context file should contain header even with invalid JSON input")
	}
}

// Helper function to create sample phase outputs for testing
func createSamplePhaseOutputs(t *testing.T, dir string) {
	t.Helper()

	// scope.json
	scope := ScopeData{
		BaseRef:  "main",
		HeadRef:  "HEAD",
		Language: "go",
		Files: ScopeFiles{
			Modified: []string{"internal/handler/user.go"},
			Added:    []string{"internal/service/notification.go"},
			Deleted:  []string{},
		},
		Stats: ScopeStats{
			TotalFiles:     2,
			TotalAdditions: 100,
			TotalDeletions: 10,
		},
		PackagesAffected: []string{"internal/handler", "internal/service"},
	}

	writeJSON(t, dir, "scope.json", scope)

	// static-analysis.json
	static := StaticAnalysisData{
		ToolVersions: map[string]string{"golangci-lint": "1.56.0"},
		Findings: []Finding{
			{Tool: "staticcheck", Rule: "SA1019", Severity: "warning", File: "user.go", Line: 45, Message: "deprecated", Category: "deprecation"},
			{Tool: "gosec", Rule: "G401", Severity: "high", File: "crypto.go", Line: 23, Message: "weak crypto", Category: "security"},
		},
		Summary: FindingSummary{High: 1, Warning: 1},
	}
	writeJSON(t, dir, "static-analysis.json", static)

	// go-ast.json
	ast := ASTData{
		Functions: FunctionChanges{
			Modified: []FunctionDiff{
				{
					Name:    "CreateUser",
					Package: "handler",
					File:    "user.go",
					Before:  FunctionInfo{Signature: "func CreateUser(ctx context.Context)", LineStart: 10, LineEnd: 20},
					After:   FunctionInfo{Signature: "func CreateUser(ctx context.Context, opts ...Option)", LineStart: 10, LineEnd: 30},
					Changes: []string{"added_param"},
				},
			},
			Added: []FunctionInfo{
				{Name: "NotifyUser", Package: "service", File: "notification.go", Signature: "func NotifyUser(ctx context.Context)", LineStart: 10, LineEnd: 40},
			},
		},
	}
	writeJSON(t, dir, "go-ast.json", ast)

	// go-calls.json
	calls := CallGraphData{
		ModifiedFunctions: []FunctionCallGraph{
			{
				Function: "handler.CreateUser",
				File:     "user.go",
				Callers: []CallSite{
					{Function: "router.ServeHTTP", File: "router.go", Line: 89},
				},
				TestCoverage: []TestCoverage{
					{TestFunction: "TestCreateUser", File: "user_test.go", Line: 23},
				},
			},
		},
		ImpactAnalysis: ImpactSummary{DirectCallers: 1, AffectedTests: 1},
	}
	writeJSON(t, dir, "go-calls.json", calls)

	// go-flow.json
	flow := DataFlowData{
		Flows: []DataFlow{
			{
				ID:        "flow-1",
				Risk:      "medium",
				Sanitized: false,
				Source:    FlowSource{Type: "http_request", File: "user.go", Line: 23, Expression: "r.Body"},
				Sink:      FlowSink{Type: "database", File: "repo.go", Line: 45, Expression: "db.Exec()"},
			},
		},
		NilSources: []NilSource{
			{Variable: "user", File: "user.go", Line: 67, Checked: true, Risk: "low"},
			{Variable: "config", File: "service.go", Line: 23, Checked: false, Risk: "high", Notes: "env var"},
		},
		Summary: FlowSummary{TotalFlows: 1, MediumRisk: 1, NilRisks: 1},
	}
	writeJSON(t, dir, "go-flow.json", flow)
}

func writeJSON(t *testing.T, dir, filename string, data interface{}) {
	t.Helper()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal %s: %v", filename, err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), jsonData, 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", filename, err)
	}
}

func TestCompiler_TestReviewerNewErrorPaths(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// Create scope
	scope := ScopeData{
		Language: "go",
		Files: ScopeFiles{
			Modified: []string{"user.go"},
			Added:    []string{},
			Deleted:  []string{},
		},
		Stats: ScopeStats{},
	}
	scopeData, err := json.Marshal(scope)
	if err != nil {
		t.Fatalf("failed to marshal scope: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "scope.json"), scopeData, 0644); err != nil {
		t.Fatalf("failed to write scope.json: %v", err)
	}

	// Create AST with NewErrorReturns
	ast := ASTData{
		ErrorHandling: ErrorHandlingData{
			NewErrorReturns: []ErrorReturn{
				{Function: "CreateUser", File: "user.go", Line: 45, ErrorType: "error", Message: "user creation failed"},
			},
		},
	}
	astData, err := json.Marshal(ast)
	if err != nil {
		t.Fatalf("failed to marshal ast: %v", err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "go-ast.json"), astData, 0644); err != nil {
		t.Fatalf("failed to write go-ast.json: %v", err)
	}

	compiler := NewCompiler(inputDir, outputDir)
	if err := compiler.Compile(); err != nil {
		t.Fatalf("Compile() error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outputDir, "context-test-reviewer.md"))
	if err != nil {
		t.Fatalf("failed to read test-reviewer context: %v", err)
	}

	if !strings.Contains(string(content), "New Error Paths") {
		t.Error("expected 'New Error Paths' focus area in test-reviewer context")
	}
}
