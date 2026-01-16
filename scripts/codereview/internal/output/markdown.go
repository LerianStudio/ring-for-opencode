// Package output handles writing analysis results to files.
package output

import (
	"fmt"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/callgraph"
)

// ImpactLevel represents the risk level of a function based on caller count.
type ImpactLevel string

const (
	// ImpactHigh indicates a function with 3 or more callers.
	ImpactHigh ImpactLevel = "HIGH"
	// ImpactMedium indicates a function with 1-2 callers.
	ImpactMedium ImpactLevel = "MEDIUM"
	// ImpactLow indicates a function with 0 callers.
	ImpactLow ImpactLevel = "LOW"
)

const (
	maxCallersToShow = 10
	maxCalleesToShow = 5
)

// RenderImpactSummary generates a Markdown representation of the call graph analysis.
func RenderImpactSummary(result *callgraph.CallGraphResult) string {
	if result == nil {
		return "# Impact Summary\n\nNo call graph analysis available.\n"
	}

	var sb strings.Builder

	// Header
	sb.WriteString("# Impact Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Language:** %s\n\n", result.Language))

	// Warnings section
	renderWarnings(&sb, result)

	// Summary metrics table
	renderSummaryTable(&sb, result)

	// Categorize functions by impact
	high, medium, low := categorizeFunctions(result.ModifiedFunctions)

	// Render each category
	if len(high) > 0 {
		sb.WriteString("## High Impact Functions\n\n")
		sb.WriteString("Functions with 3 or more callers - changes here have wide-reaching effects.\n\n")
		for _, fcg := range high {
			sb.WriteString(renderFunctionImpact(fcg, result.Language, string(ImpactHigh)))
		}
	}

	if len(medium) > 0 {
		sb.WriteString("## Medium Impact Functions\n\n")
		sb.WriteString("Functions with 1-2 callers - changes affect a limited scope.\n\n")
		for _, fcg := range medium {
			sb.WriteString(renderFunctionImpact(fcg, result.Language, string(ImpactMedium)))
		}
	}

	if len(low) > 0 {
		sb.WriteString("## Low Impact Functions\n\n")
		sb.WriteString("Functions with no callers - may be entry points, tests, or dead code.\n\n")
		for _, fcg := range low {
			sb.WriteString(renderFunctionImpact(fcg, result.Language, string(ImpactLow)))
		}
	}

	if len(result.ModifiedFunctions) == 0 {
		sb.WriteString("## No Modified Functions Analyzed\n\n")
		sb.WriteString("No functions were found in the modified files for call graph analysis.\n\n")
	}

	return sb.String()
}

// renderWarnings adds the warnings section if there are any warnings or issues.
func renderWarnings(sb *strings.Builder, result *callgraph.CallGraphResult) {
	hasWarnings := result.TimeBudgetExceeded || result.PartialResults || len(result.Warnings) > 0

	if !hasWarnings {
		return
	}

	sb.WriteString("## Warnings\n\n")

	if result.TimeBudgetExceeded {
		sb.WriteString("- **Time Budget Exceeded:** Analysis was stopped before completion due to time constraints.\n")
	}

	if result.PartialResults {
		sb.WriteString("- **Partial Results:** Some functions could not be fully analyzed.\n")
	}

	for _, warning := range result.Warnings {
		sb.WriteString(fmt.Sprintf("- %s\n", warning))
	}

	sb.WriteString("\n")
}

// renderSummaryTable adds the summary metrics table.
func renderSummaryTable(sb *strings.Builder, result *callgraph.CallGraphResult) {
	sb.WriteString("## Summary Metrics\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Modified Functions | %d |\n", len(result.ModifiedFunctions)))
	sb.WriteString(fmt.Sprintf("| Direct Callers | %d |\n", result.ImpactAnalysis.DirectCallers))
	sb.WriteString(fmt.Sprintf("| Transitive Callers | %d |\n", result.ImpactAnalysis.TransitiveCallers))
	sb.WriteString(fmt.Sprintf("| Affected Tests | %d |\n", result.ImpactAnalysis.AffectedTests))
	sb.WriteString(fmt.Sprintf("| Affected Packages | %d |\n", len(result.ImpactAnalysis.AffectedPackages)))
	sb.WriteString("\n")

	// List affected packages if any
	if len(result.ImpactAnalysis.AffectedPackages) > 0 {
		sb.WriteString("### Affected Packages\n\n")
		for _, pkg := range result.ImpactAnalysis.AffectedPackages {
			sb.WriteString(fmt.Sprintf("- `%s`\n", pkg))
		}
		sb.WriteString("\n")
	}
}

// categorizeFunctions groups functions by their impact level based on caller count.
func categorizeFunctions(functions []callgraph.FunctionCallGraph) (high, medium, low []callgraph.FunctionCallGraph) {
	for _, fcg := range functions {
		callerCount := len(fcg.Callers)
		switch {
		case callerCount >= 3:
			high = append(high, fcg)
		case callerCount >= 1:
			medium = append(medium, fcg)
		default:
			low = append(low, fcg)
		}
	}
	return high, medium, low
}

// renderFunctionImpact generates the Markdown for a single function's impact.
func renderFunctionImpact(fcg callgraph.FunctionCallGraph, language, riskLevel string) string {
	var sb strings.Builder

	// Function header with risk badge
	sb.WriteString(fmt.Sprintf("### `%s`\n\n", fcg.Function))
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", fcg.File))
	sb.WriteString(fmt.Sprintf("**Risk Level:** %s\n", riskLevel))
	sb.WriteString(fmt.Sprintf("**Callers:** %d\n\n", len(fcg.Callers)))

	// Test coverage status
	renderTestCoverage(&sb, fcg.TestCoverage)

	// Direct callers
	renderCallers(&sb, fcg.Callers)

	// Callees
	renderCallees(&sb, fcg.Callees)

	sb.WriteString("---\n\n")
	return sb.String()
}

// renderTestCoverage adds test coverage information.
func renderTestCoverage(sb *strings.Builder, tests []callgraph.TestCoverage) {
	if len(tests) > 0 {
		sb.WriteString("**Test Coverage:** :white_check_mark: Has tests\n\n")
		sb.WriteString("<details>\n<summary>Tests covering this function</summary>\n\n")
		for _, test := range tests {
			sb.WriteString(fmt.Sprintf("- `%s` (%s:%d)\n", test.TestFunction, test.File, test.Line))
		}
		sb.WriteString("\n</details>\n\n")
	} else {
		sb.WriteString("**Test Coverage:** :warning: No tests found\n\n")
	}
}

// renderCallers adds the list of direct callers.
func renderCallers(sb *strings.Builder, callers []callgraph.CallInfo) {
	if len(callers) == 0 {
		sb.WriteString("**Direct Callers:** None\n\n")
		return
	}

	sb.WriteString("**Direct Callers:**\n\n")

	displayCount := len(callers)
	if displayCount > maxCallersToShow {
		displayCount = maxCallersToShow
	}

	for i := 0; i < displayCount; i++ {
		caller := callers[i]
		if caller.CallSite != "" {
			sb.WriteString(fmt.Sprintf("- `%s` at `%s:%d` (call site: %s)\n",
				caller.Function, caller.File, caller.Line, caller.CallSite))
		} else {
			sb.WriteString(fmt.Sprintf("- `%s` at `%s:%d`\n",
				caller.Function, caller.File, caller.Line))
		}
	}

	if len(callers) > maxCallersToShow {
		sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(callers)-maxCallersToShow))
	}

	sb.WriteString("\n")
}

// renderCallees adds the list of callees.
func renderCallees(sb *strings.Builder, callees []callgraph.CallInfo) {
	if len(callees) == 0 {
		sb.WriteString("**Calls:** None\n\n")
		return
	}

	sb.WriteString("**Calls:**\n\n")

	displayCount := len(callees)
	if displayCount > maxCalleesToShow {
		displayCount = maxCalleesToShow
	}

	for i := 0; i < displayCount; i++ {
		callee := callees[i]
		if callee.File != "" && callee.Line > 0 {
			sb.WriteString(fmt.Sprintf("- `%s` at `%s:%d`\n",
				callee.Function, callee.File, callee.Line))
		} else {
			sb.WriteString(fmt.Sprintf("- `%s`\n", callee.Function))
		}
	}

	if len(callees) > maxCalleesToShow {
		sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(callees)-maxCalleesToShow))
	}

	sb.WriteString("\n")
}
