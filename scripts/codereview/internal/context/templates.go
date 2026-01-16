package context

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Template definitions for each reviewer's context file.

const codeReviewerTemplate = `# Pre-Analysis Context: Code Quality

## Static Analysis Findings ({{.FindingCount}} issues)

{{if .Findings}}
| Severity | Tool | File | Line | Message |
|----------|------|------|------|---------|
{{- range .Findings}}
| {{.Severity}} | {{.Tool}} | {{.File}} | {{.Line}} | {{.Message}} |
{{- end}}
{{else}}
No static analysis findings.
{{end}}

## Semantic Changes

{{if .HasSemanticChanges}}
### Functions Modified ({{len .ModifiedFunctions}})
{{range .ModifiedFunctions}}
#### ` + "`{{.Package}}.{{.Name}}`" + `
**File:** ` + "`{{.File}}:{{.After.LineStart}}-{{.After.LineEnd}}`" + `
{{if .Changes}}**Changes:** {{join .Changes ", "}}{{end}}

{{if signatureChanged .}}
` + "```diff" + `
- {{.Before.Signature}}
+ {{.After.Signature}}
` + "```" + `
{{end}}
{{end}}

### Functions Added ({{len .AddedFunctions}})
{{range .AddedFunctions}}
- ` + "`{{.Package}}.{{.Name}}`" + ` at ` + "`{{.File}}:{{.LineStart}}`" + `
{{end}}

### Types Modified ({{len .ModifiedTypes}})
{{range .ModifiedTypes}}
#### ` + "`{{.Name}}`" + `
**File:** ` + "`{{.File}}`" + `

| Field | Before | After |
|-------|--------|-------|
{{- range fieldChanges .}}
| {{.Name}} | {{.Before}} | {{.After}} |
{{- end}}
{{end}}
{{else}}
No semantic changes detected.
{{end}}

## Focus Areas

Based on analysis, pay special attention to:
{{range $i, $area := .FocusAreas}}
{{inc $i}}. **{{$area.Title}}** - {{$area.Description}}
{{- end}}
{{if not .FocusAreas}}
No specific focus areas identified.
{{end}}
`

const securityReviewerTemplate = `# Pre-Analysis Context: Security

## Security Scanner Findings ({{.FindingCount}} issues)

{{if .Findings}}
| Severity | Tool | Rule | File | Line | Message |
|----------|------|------|------|------|---------|
{{- range .Findings}}
| {{.Severity}} | {{.Tool}} | {{.Rule}} | {{.File}} | {{.Line}} | {{.Message}} |
{{- end}}
{{else}}
No security scanner findings.
{{end}}

## Data Flow Analysis

{{if .HasDataFlowAnalysis}}
### High Risk Flows ({{len .HighRiskFlows}})
{{range .HighRiskFlows}}
#### {{.ID}}: {{.Source.Type}} -> {{.Sink.Type}}
**File:** ` + "`{{.Source.File}}:{{.Source.Line}}`" + `
**Risk:** {{.Risk}}
**Notes:** {{.Notes}}

**Source:** ` + "`{{.Source.Expression}}`" + `
**Sink:** ` + "`{{.Sink.Expression}}`" + `
**Sanitized:** {{if .Sanitized}}Yes{{else}}No{{end}}
{{end}}

### Medium Risk Flows ({{len .MediumRiskFlows}})
{{range .MediumRiskFlows}}
- {{.Source.Type}} -> {{.Sink.Type}} at ` + "`{{.Source.File}}:{{.Source.Line}}`" + ` ({{.Notes}})
{{- end}}
{{else}}
No data flow analysis available.
{{end}}

## Focus Areas

Based on analysis, pay special attention to:
{{range $i, $area := .FocusAreas}}
{{inc $i}}. **{{$area.Title}}** - {{$area.Description}}
{{- end}}
{{if not .FocusAreas}}
No specific focus areas identified.
{{end}}
`

const businessLogicReviewerTemplate = `# Pre-Analysis Context: Business Logic

## Impact Analysis

{{if .HasCallGraph}}
### High Impact Changes

{{range .HighImpactFunctions}}
#### ` + "`{{.Function}}`" + `
**File:** ` + "`{{.File}}`" + `
**Risk Level:** {{riskLevel .}} ({{len .Callers}} direct callers)

**Direct Callers (signature change affects these):**
{{range $i, $caller := .Callers}}
{{inc $i}}. ` + "`{{$caller.Function}}`" + ` - ` + "`{{$caller.File}}:{{$caller.Line}}`" + `
{{- end}}

**Callees (this function depends on):**
{{range $i, $callee := .Callees}}
{{inc $i}}. ` + "`{{$callee.Function}}`" + `
{{- end}}
{{end}}
{{else}}
No call graph analysis available.
{{end}}

## Semantic Changes

{{if .HasSemanticChanges}}
### Functions with Logic Changes
{{range $i, $f := .ModifiedFunctions}}
{{inc $i}}. **` + "`{{$f.Package}}.{{$f.Name}}`" + `** - {{join $f.Changes ", "}}
{{- end}}
{{else}}
No semantic changes detected.
{{end}}

## Focus Areas

Based on analysis, pay special attention to:
{{range $i, $area := .FocusAreas}}
{{inc $i}}. **{{$area.Title}}** - {{$area.Description}}
{{- end}}
{{if not .FocusAreas}}
No specific focus areas identified.
{{end}}
`

const testReviewerTemplate = `# Pre-Analysis Context: Testing

## Test Coverage for Modified Code

{{if .HasCallGraph}}
| Function | File | Tests | Status |
|----------|------|-------|--------|
{{- range .AllModifiedFunctionsGraph}}
| ` + "`{{.Function}}`" + ` | {{.File}} | {{len .TestCoverage}} tests | {{testStatus .}} |
{{- end}}
{{else}}
No call graph analysis available for test coverage.
{{end}}

## Uncovered New Code

{{if .UncoveredFunctions}}
{{range .UncoveredFunctions}}
- ` + "`{{.Function}}`" + ` at ` + "`{{.File}}`" + ` - **No tests found**
{{- end}}
{{else}}
All modified code has test coverage.
{{end}}

## Focus Areas

Based on analysis, pay special attention to:
{{range $i, $area := .FocusAreas}}
{{inc $i}}. **{{$area.Title}}** - {{$area.Description}}
{{- end}}
{{if not .FocusAreas}}
No specific focus areas identified.
{{end}}
`

const nilSafetyReviewerTemplate = `# Pre-Analysis Context: Nil Safety

## Nil Source Analysis

{{if .HasNilSources}}
| Variable | File | Line | Checked? | Risk |
|----------|------|------|----------|------|
{{- range .NilSources}}
| ` + "`{{.Variable}}`" + ` | {{.File}} | {{.Line}} | {{if .Checked}}Yes{{else}}No{{end}} | {{.Risk}} |
{{- end}}
{{else}}
No nil sources detected in changed code.
{{end}}

## High Risk Nil Sources

{{range .HighRiskNilSources}}
### ` + "`{{.Variable}}`" + ` at ` + "`{{.File}}:{{.Line}}`" + `
**Expression:** ` + "`{{.Expression}}`" + `
**Checked:** {{if .Checked}}Yes (line {{.CheckLine}}){{else}}No{{end}}
**Notes:** {{.Notes}}
{{end}}

## Focus Areas

Based on analysis, pay special attention to:
{{range $i, $area := .FocusAreas}}
{{inc $i}}. **{{$area.Title}}** - {{$area.Description}}
{{- end}}
{{if not .FocusAreas}}
No specific focus areas identified.
{{end}}
`

// FocusArea represents a specific area requiring attention.
type FocusArea struct {
	Title       string
	Description string
}

// FieldChange represents a before/after field comparison.
type FieldChange struct {
	Name   string
	Before string
	After  string
}

// TemplateData holds data for template rendering.
type TemplateData struct {
	// Common fields
	FindingCount int
	Findings     []Finding
	FocusAreas   []FocusArea

	// Semantic changes (code-reviewer, business-logic-reviewer)
	HasSemanticChanges bool
	ModifiedFunctions  []FunctionDiff
	AddedFunctions     []FunctionInfo
	ModifiedTypes      []TypeDiff

	// Data flow (security-reviewer)
	HasDataFlowAnalysis bool
	HighRiskFlows       []DataFlow
	MediumRiskFlows     []DataFlow

	// Call graph (business-logic-reviewer, test-reviewer)
	HasCallGraph              bool
	HighImpactFunctions       []FunctionCallGraph
	AllModifiedFunctionsGraph []FunctionCallGraph // All modified functions with call graph data
	UncoveredFunctions        []FunctionCallGraph

	// Nil safety (nil-safety-reviewer)
	HasNilSources      bool
	NilSources         []NilSource
	HighRiskNilSources []NilSource
}

// templateFuncs provides custom functions for templates.
var templateFuncs = template.FuncMap{
	"inc": func(i int) int {
		return i + 1
	},
	"join": func(items []string, sep string) string {
		return strings.Join(items, sep)
	},
	"signatureChanged": func(f FunctionDiff) bool {
		return f.Before.Signature != f.After.Signature
	},
	"fieldChanges": func(t TypeDiff) []FieldChange {
		var changes []FieldChange
		beforeMap := make(map[string]string)
		for _, f := range t.Before.Fields {
			beforeMap[f.Name] = f.Type
		}
		afterMap := make(map[string]string)
		for _, f := range t.After.Fields {
			afterMap[f.Name] = f.Type
		}

		// Find modified and added fields
		for _, f := range t.After.Fields {
			if before, ok := beforeMap[f.Name]; ok {
				if before != f.Type {
					changes = append(changes, FieldChange{
						Name:   f.Name,
						Before: before,
						After:  f.Type,
					})
				}
			} else {
				changes = append(changes, FieldChange{
					Name:   f.Name,
					Before: "-",
					After:  f.Type + " (added)",
				})
			}
		}

		// Find deleted fields
		for _, f := range t.Before.Fields {
			if _, ok := afterMap[f.Name]; !ok {
				changes = append(changes, FieldChange{
					Name:   f.Name,
					Before: f.Type,
					After:  "(deleted)",
				})
			}
		}

		return changes
	},
	"riskLevel": func(f FunctionCallGraph) string {
		callerCount := len(f.Callers)
		if callerCount >= 5 {
			return "HIGH"
		}
		if callerCount >= 2 {
			return "MEDIUM"
		}
		return "LOW"
	},
	"testStatus": func(f FunctionCallGraph) string {
		if len(f.TestCoverage) == 0 {
			return "No tests"
		}
		return fmt.Sprintf("%d tests", len(f.TestCoverage))
	},
}

// RenderTemplate renders a template with the given data.
func RenderTemplate(templateStr string, data *TemplateData) (string, error) {
	tmpl, err := template.New("context").Funcs(templateFuncs).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// GetTemplateForReviewer returns the template string for a specific reviewer.
func GetTemplateForReviewer(reviewer string) (string, error) {
	switch reviewer {
	case "code-reviewer":
		return codeReviewerTemplate, nil
	case "security-reviewer":
		return securityReviewerTemplate, nil
	case "business-logic-reviewer":
		return businessLogicReviewerTemplate, nil
	case "test-reviewer":
		return testReviewerTemplate, nil
	case "nil-safety-reviewer":
		return nilSafetyReviewerTemplate, nil
	default:
		return "", fmt.Errorf("unknown reviewer: %s", reviewer)
	}
}
