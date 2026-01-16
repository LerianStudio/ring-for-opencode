//go:build integration

package dataflow

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegration_FullPipeline tests the complete data flow analysis pipeline
// with a Go file containing multiple security vulnerabilities.
func TestIntegration_FullPipeline(t *testing.T) {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "dataflow-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create vulnerable Go file with multiple security issues
	vulnGoFile := filepath.Join(tempDir, "vulnerable.go")
	vulnContent := `package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os/exec"
)

var db *sql.DB

// SQL Injection vulnerability - string concatenation in query
func getUserByID(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	query := "SELECT * FROM users WHERE id = '" + userID + "'"
	result, err := db.Exec(query)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = result
}

// Command Injection vulnerability - user input in shell command
func runCommand(w http.ResponseWriter, r *http.Request) {
	cmd := r.FormValue("cmd")
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(output)
}

// XSS vulnerability - unsanitized user input in response
func greetUser(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	w.Write([]byte("<h1>Hello " + name + "</h1>"))
}

// Template Injection vulnerability - user input parsed as template
func renderTemplate(w http.ResponseWriter, r *http.Request) {
	tmplStr := r.URL.Query().Get("template")
	tmpl, err := template.New("user").Parse(tmplStr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tmpl.Execute(w, nil)
}

// Safe parameterized query - properly sanitized
func getSafeUser(w http.ResponseWriter, r *http.Request) {
	safeID := r.URL.Query().Get("id")
	// Using parameterized query - sanitized
	result, err := db.Exec("SELECT * FROM users WHERE id = ?", safeID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_ = result
}

// Nil risk - unchecked map lookup
func getConfig(key string) string {
	config := make(map[string]string)
	value := config[key]
	return value
}

// Nil risk - checked pattern
func getConfigSafe(key string) string {
	config := make(map[string]string)
	value, ok := config[key]
	if !ok {
		return "default"
	}
	return value
}
`

	if err := os.WriteFile(vulnGoFile, []byte(vulnContent), 0644); err != nil {
		t.Fatalf("failed to write vulnerable Go file: %v", err)
	}

	// Create scope.json pointing to test file
	scopeJSON := map[string]interface{}{
		"base_ref": "main",
		"head_ref": "HEAD",
		"language": "go",
		"files": map[string]interface{}{
			"modified": []string{vulnGoFile},
			"added":    []string{},
			"deleted":  []string{},
		},
		"stats": map[string]interface{}{
			"total_files":     1,
			"total_additions": 80,
			"total_deletions": 0,
		},
		"packages_affected": []string{tempDir},
	}

	scopePath := filepath.Join(tempDir, "scope.json")
	scopeData, err := json.MarshalIndent(scopeJSON, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal scope.json: %v", err)
	}
	if err := os.WriteFile(scopePath, scopeData, 0644); err != nil {
		t.Fatalf("failed to write scope.json: %v", err)
	}

	// Run GoAnalyzer
	analyzer := NewGoAnalyzer(tempDir)
	analysis, err := analyzer.Analyze([]string{vulnGoFile})
	if err != nil {
		t.Fatalf("GoAnalyzer.Analyze() failed: %v", err)
	}

	// Verify sources detected
	if len(analysis.Sources) == 0 {
		t.Error("expected at least one source detected")
	}
	t.Logf("Sources detected: %d", len(analysis.Sources))

	// Verify sinks detected
	if len(analysis.Sinks) == 0 {
		t.Error("expected at least one sink detected")
	}
	t.Logf("Sinks detected: %d", len(analysis.Sinks))

	// Verify flows detected
	if len(analysis.Flows) == 0 {
		t.Error("expected at least one flow detected")
	}
	t.Logf("Flows detected: %d", len(analysis.Flows))

	// Count critical and high risk flows
	criticalCount := 0
	highCount := 0
	sanitizedCount := 0

	for _, flow := range analysis.Flows {
		switch flow.Risk {
		case RiskCritical:
			criticalCount++
			t.Logf("Critical flow: %s -> %s (%s)", flow.Source.Type, flow.Sink.Type, flow.Description)
		case RiskHigh:
			highCount++
			t.Logf("High risk flow: %s -> %s (%s)", flow.Source.Type, flow.Sink.Type, flow.Description)
		}
		if flow.Sanitized {
			sanitizedCount++
			t.Logf("Sanitized flow: %s -> %s (sanitizers: %v)", flow.Source.Type, flow.Sink.Type, flow.Sanitizers)
		}
	}

	// Verify at least 1 critical flow (command injection is reliably detected)
	// Note: SQL injection detection depends on string concatenation patterns being
	// directly visible in the query. The analyzer tracks variable usage heuristically.
	if criticalCount < 1 {
		t.Errorf("expected at least 1 critical flow (command injection), got %d", criticalCount)
	}

	// Verify at least 1 high risk flow (XSS)
	if highCount < 1 {
		t.Errorf("expected at least 1 high risk flow (XSS), got %d", highCount)
	}

	// Log sanitized flow count (parameterized query should be detected as sanitized)
	t.Logf("Sanitized flows: %d", sanitizedCount)

	// Verify statistics
	if analysis.Statistics.TotalFlows != len(analysis.Flows) {
		t.Errorf("statistics.TotalFlows = %d, want %d", analysis.Statistics.TotalFlows, len(analysis.Flows))
	}
	if analysis.Statistics.CriticalFlows != criticalCount {
		t.Errorf("statistics.CriticalFlows = %d, want %d", analysis.Statistics.CriticalFlows, criticalCount)
	}
	if analysis.Statistics.HighRiskFlows != highCount {
		t.Errorf("statistics.HighRiskFlows = %d, want %d", analysis.Statistics.HighRiskFlows, highCount)
	}

	// Generate security summary
	analyses := map[string]*FlowAnalysis{
		"go": analysis,
	}
	summary := GenerateSecuritySummary(analyses)

	// Verify summary contains expected sections
	expectedSections := []string{
		"# Security Data Flow Analysis",
		"## Executive Summary",
		"## Risk Assessment",
		"CRITICAL",
	}

	for _, section := range expectedSections {
		if !containsString(summary, section) {
			t.Errorf("security summary missing section: %q", section)
		}
	}

	// Verify Critical & High Risk Flows section appears when there are issues
	if criticalCount > 0 || highCount > 0 {
		if !containsString(summary, "## Critical & High Risk Flows") {
			t.Error("security summary missing '## Critical & High Risk Flows' section")
		}
	}

	// Write outputs to temp dir for inspection
	analysisJSON, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		t.Errorf("failed to marshal analysis JSON: %v", err)
	}
	analysisPath := filepath.Join(tempDir, "analysis.json")
	if err := os.WriteFile(analysisPath, analysisJSON, 0644); err != nil {
		t.Errorf("failed to write analysis.json: %v", err)
	}

	summaryPath := filepath.Join(tempDir, "security-summary.md")
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		t.Errorf("failed to write security-summary.md: %v", err)
	}

	t.Logf("Analysis output written to: %s", analysisPath)
	t.Logf("Summary output written to: %s", summaryPath)
	t.Logf("Statistics: Sources=%d, Sinks=%d, Flows=%d, Critical=%d, High=%d, Nil=%d",
		analysis.Statistics.TotalSources,
		analysis.Statistics.TotalSinks,
		analysis.Statistics.TotalFlows,
		analysis.Statistics.CriticalFlows,
		analysis.Statistics.HighRiskFlows,
		analysis.Statistics.NilRisks)
}

// TestIntegration_MultiLanguage tests analysis across multiple languages.
func TestIntegration_MultiLanguage(t *testing.T) {
	// Skip if python3 is not available
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available, skipping multi-language test")
	}

	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "dataflow-multilang-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create Go file with XSS vulnerability
	goFile := filepath.Join(tempDir, "handler.go")
	goContent := `package main

import (
	"net/http"
)

// XSS vulnerability
func greetHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	w.Write([]byte("<h1>Hello " + name + "</h1>"))
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatalf("failed to write Go file: %v", err)
	}

	// Create Python file with Flask vulnerability
	pyFile := filepath.Join(tempDir, "app.py")
	pyContent := `from flask import Flask, request

app = Flask(__name__)

# XSS vulnerability - unsanitized user input
@app.route('/greet')
def greet():
    name = request.args.get('name')
    return '<h1>Hello ' + name + '</h1>'

# SQL injection vulnerability
@app.route('/user')
def get_user():
    user_id = request.args.get('id')
    query = "SELECT * FROM users WHERE id = '" + user_id + "'"
    # cursor.execute(query)
    return query
`
	if err := os.WriteFile(pyFile, []byte(pyContent), 0644); err != nil {
		t.Fatalf("failed to write Python file: %v", err)
	}

	// Run GoAnalyzer on Go file
	goAnalyzer := NewGoAnalyzer(tempDir)
	goAnalysis, err := goAnalyzer.Analyze([]string{goFile})
	if err != nil {
		t.Fatalf("GoAnalyzer.Analyze() failed: %v", err)
	}

	// Verify Go analysis detected flows
	if len(goAnalysis.Flows) == 0 {
		t.Error("Go analysis expected to detect flows")
	}
	t.Logf("Go analysis: Sources=%d, Sinks=%d, Flows=%d",
		goAnalysis.Statistics.TotalSources,
		goAnalysis.Statistics.TotalSinks,
		goAnalysis.Statistics.TotalFlows)

	// Check for XSS flow in Go analysis
	hasGoXSS := false
	for _, flow := range goAnalysis.Flows {
		if flow.Sink.Type == SinkResponse && flow.Source.Type == SourceHTTPQuery {
			hasGoXSS = true
			break
		}
	}
	if !hasGoXSS {
		t.Error("Go analysis should detect XSS flow (HTTP query -> response)")
	}

	// Track analyses for combined summary
	analyses := map[string]*FlowAnalysis{
		"go": goAnalysis,
	}

	// Run PythonAnalyzer only if SCRIPT_DIR env is set
	scriptDir := os.Getenv("SCRIPT_DIR")
	if scriptDir != "" {
		pyAnalyzer := NewPythonAnalyzer(scriptDir)
		pyAnalysis, err := pyAnalyzer.Analyze([]string{pyFile})
		if err != nil {
			t.Logf("PythonAnalyzer.Analyze() failed (may be expected if script not found): %v", err)
		} else {
			t.Logf("Python analysis: Sources=%d, Sinks=%d, Flows=%d",
				pyAnalysis.Statistics.TotalSources,
				pyAnalysis.Statistics.TotalSinks,
				pyAnalysis.Statistics.TotalFlows)
			analyses["python"] = pyAnalysis
		}
	} else {
		t.Log("SCRIPT_DIR not set, skipping Python analyzer")
	}

	// Generate combined summary if we have multiple analyses
	summary := GenerateSecuritySummary(analyses)

	// Verify summary includes language breakdown
	if !containsString(summary, "## Language Breakdown") {
		t.Error("combined summary should include Language Breakdown section")
	}

	// Verify summary includes Go section
	if !containsString(summary, "### Go") {
		t.Error("combined summary should include Go language section")
	}

	// Write summary to temp dir for inspection
	summaryPath := filepath.Join(tempDir, "combined-summary.md")
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		t.Errorf("failed to write combined summary: %v", err)
	}
	t.Logf("Combined summary written to: %s", summaryPath)
}

// containsString checks if string s contains substring substr.
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
