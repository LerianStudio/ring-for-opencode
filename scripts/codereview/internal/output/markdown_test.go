package output

import (
	"strings"
	"testing"

	"github.com/lerianstudio/ring/scripts/codereview/internal/callgraph"
)

// createTestCallGraphResult creates a CallGraphResult for testing.
func createTestCallGraphResult() *callgraph.CallGraphResult {
	return &callgraph.CallGraphResult{
		Language: "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{
			{
				Function: "ProcessPayment",
				File:     "internal/service/payment.go",
				Callers: []callgraph.CallInfo{
					{Function: "HandleCheckout", File: "internal/handler/checkout.go", Line: 45},
					{Function: "HandleRefund", File: "internal/handler/refund.go", Line: 32},
					{Function: "BatchProcess", File: "internal/worker/batch.go", Line: 78},
				},
				Callees: []callgraph.CallInfo{
					{Function: "ValidateCard", File: "internal/service/validation.go", Line: 12},
					{Function: "ChargeAmount", File: "internal/gateway/stripe.go", Line: 55},
				},
				TestCoverage: []callgraph.TestCoverage{
					{TestFunction: "TestProcessPayment_Success", File: "internal/service/payment_test.go", Line: 15},
					{TestFunction: "TestProcessPayment_InvalidCard", File: "internal/service/payment_test.go", Line: 45},
				},
			},
			{
				Function: "ValidateInput",
				File:     "internal/service/validation.go",
				Callers: []callgraph.CallInfo{
					{Function: "ProcessPayment", File: "internal/service/payment.go", Line: 23},
				},
				Callees:      []callgraph.CallInfo{},
				TestCoverage: []callgraph.TestCoverage{},
			},
			{
				Function:     "UnusedHelper",
				File:         "internal/util/helpers.go",
				Callers:      []callgraph.CallInfo{},
				Callees:      []callgraph.CallInfo{},
				TestCoverage: []callgraph.TestCoverage{},
			},
		},
		ImpactAnalysis: callgraph.ImpactAnalysis{
			DirectCallers:     4,
			TransitiveCallers: 12,
			AffectedTests:     5,
			AffectedPackages:  []string{"internal/service", "internal/handler", "internal/worker"},
		},
	}
}

func TestRenderImpactSummary_NilResult(t *testing.T) {
	result := RenderImpactSummary(nil)

	if !strings.Contains(result, "# Impact Summary") {
		t.Error("Expected header in nil result")
	}
	if !strings.Contains(result, "No call graph analysis available") {
		t.Error("Expected nil message in result")
	}
}

func TestRenderImpactSummary_ContainsHeader(t *testing.T) {
	result := RenderImpactSummary(createTestCallGraphResult())

	if !strings.Contains(result, "# Impact Summary") {
		t.Error("Expected main header")
	}
	if !strings.Contains(result, "**Language:** go") {
		t.Error("Expected language field")
	}
}

func TestRenderImpactSummary_ContainsSummaryTable(t *testing.T) {
	result := RenderImpactSummary(createTestCallGraphResult())

	if !strings.Contains(result, "## Summary Metrics") {
		t.Error("Expected Summary Metrics section")
	}
	if !strings.Contains(result, "| Metric | Value |") {
		t.Error("Expected metrics table header")
	}
	if !strings.Contains(result, "| Direct Callers | 4 |") {
		t.Error("Expected direct callers count")
	}
	if !strings.Contains(result, "| Transitive Callers | 12 |") {
		t.Error("Expected transitive callers count")
	}
	if !strings.Contains(result, "| Affected Tests | 5 |") {
		t.Error("Expected affected tests count")
	}
}

func TestRenderImpactSummary_CategorizesByImpact(t *testing.T) {
	result := RenderImpactSummary(createTestCallGraphResult())

	// ProcessPayment has 3 callers -> HIGH
	if !strings.Contains(result, "## High Impact Functions") {
		t.Error("Expected High Impact section")
	}
	if !strings.Contains(result, "`ProcessPayment`") {
		t.Error("Expected ProcessPayment function")
	}

	// ValidateInput has 1 caller -> MEDIUM
	if !strings.Contains(result, "## Medium Impact Functions") {
		t.Error("Expected Medium Impact section")
	}
	if !strings.Contains(result, "`ValidateInput`") {
		t.Error("Expected ValidateInput function")
	}

	// UnusedHelper has 0 callers -> LOW
	if !strings.Contains(result, "## Low Impact Functions") {
		t.Error("Expected Low Impact section")
	}
	if !strings.Contains(result, "`UnusedHelper`") {
		t.Error("Expected UnusedHelper function")
	}
}

func TestRenderImpactSummary_ShowsTestCoverage(t *testing.T) {
	result := RenderImpactSummary(createTestCallGraphResult())

	// ProcessPayment has tests
	if !strings.Contains(result, "Has tests") {
		t.Error("Expected test coverage indicator")
	}

	// ValidateInput and UnusedHelper have no tests
	if !strings.Contains(result, "No tests found") {
		t.Error("Expected warning for functions without tests")
	}
}

func TestRenderImpactSummary_ShowsCallers(t *testing.T) {
	result := RenderImpactSummary(createTestCallGraphResult())

	if !strings.Contains(result, "**Direct Callers:**") {
		t.Error("Expected Direct Callers section")
	}
	if !strings.Contains(result, "`HandleCheckout`") {
		t.Error("Expected caller function name")
	}
	if !strings.Contains(result, "internal/handler/checkout.go:45") {
		t.Error("Expected caller file location")
	}
}

func TestRenderImpactSummary_ShowsCallees(t *testing.T) {
	result := RenderImpactSummary(createTestCallGraphResult())

	if !strings.Contains(result, "**Calls:**") {
		t.Error("Expected Calls section")
	}
	if !strings.Contains(result, "`ValidateCard`") {
		t.Error("Expected callee function name")
	}
}

func TestRenderImpactSummary_TruncatesLongCallerList(t *testing.T) {
	// Create result with many callers
	manyCallers := make([]callgraph.CallInfo, 15)
	for i := 0; i < 15; i++ {
		manyCallers[i] = callgraph.CallInfo{
			Function: "Caller" + string(rune('A'+i)),
			File:     "file.go",
			Line:     i + 1,
		}
	}

	cgResult := &callgraph.CallGraphResult{
		Language: "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{
			{
				Function: "ManyCallers",
				File:     "test.go",
				Callers:  manyCallers,
			},
		},
	}

	result := RenderImpactSummary(cgResult)

	// Should show "... and N more" for callers > 10
	if !strings.Contains(result, "... and 5 more") {
		t.Error("Expected truncation message for long caller list")
	}
}

func TestRenderImpactSummary_TruncatesLongCalleeList(t *testing.T) {
	// Create result with many callees
	manyCallees := make([]callgraph.CallInfo, 10)
	for i := 0; i < 10; i++ {
		manyCallees[i] = callgraph.CallInfo{
			Function: "Callee" + string(rune('A'+i)),
			File:     "file.go",
			Line:     i + 1,
		}
	}

	cgResult := &callgraph.CallGraphResult{
		Language: "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{
			{
				Function: "ManyCallees",
				File:     "test.go",
				Callees:  manyCallees,
			},
		},
	}

	result := RenderImpactSummary(cgResult)

	// Should show "... and N more" for callees > 5
	if !strings.Contains(result, "... and 5 more") {
		t.Error("Expected truncation message for long callee list")
	}
}

func TestRenderImpactSummary_ShowsWarnings(t *testing.T) {
	cgResult := &callgraph.CallGraphResult{
		Language:           "go",
		ModifiedFunctions:  []callgraph.FunctionCallGraph{},
		TimeBudgetExceeded: true,
		PartialResults:     true,
		Warnings:           []string{"Could not analyze some files", "External dependencies skipped"},
	}

	result := RenderImpactSummary(cgResult)

	if !strings.Contains(result, "## Warnings") {
		t.Error("Expected Warnings section")
	}
	if !strings.Contains(result, "Time Budget Exceeded") {
		t.Error("Expected time budget warning")
	}
	if !strings.Contains(result, "Partial Results") {
		t.Error("Expected partial results warning")
	}
	if !strings.Contains(result, "Could not analyze some files") {
		t.Error("Expected custom warning message")
	}
}

func TestRenderImpactSummary_NoWarningsSection_WhenNoWarnings(t *testing.T) {
	cgResult := &callgraph.CallGraphResult{
		Language:           "go",
		ModifiedFunctions:  []callgraph.FunctionCallGraph{},
		TimeBudgetExceeded: false,
		PartialResults:     false,
		Warnings:           nil,
	}

	result := RenderImpactSummary(cgResult)

	if strings.Contains(result, "## Warnings") {
		t.Error("Should not have Warnings section when no warnings")
	}
}

func TestRenderImpactSummary_EmptyFunctions(t *testing.T) {
	cgResult := &callgraph.CallGraphResult{
		Language:          "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{},
	}

	result := RenderImpactSummary(cgResult)

	if !strings.Contains(result, "No Modified Functions Analyzed") {
		t.Error("Expected message about no modified functions")
	}
}

func TestRenderImpactSummary_ShowsAffectedPackages(t *testing.T) {
	result := RenderImpactSummary(createTestCallGraphResult())

	if !strings.Contains(result, "### Affected Packages") {
		t.Error("Expected Affected Packages section")
	}
	if !strings.Contains(result, "`internal/service`") {
		t.Error("Expected package name in list")
	}
}

func TestCategorizeFunctions(t *testing.T) {
	functions := []callgraph.FunctionCallGraph{
		{Function: "High1", Callers: make([]callgraph.CallInfo, 5)},
		{Function: "High2", Callers: make([]callgraph.CallInfo, 3)},
		{Function: "Medium1", Callers: make([]callgraph.CallInfo, 2)},
		{Function: "Medium2", Callers: make([]callgraph.CallInfo, 1)},
		{Function: "Low1", Callers: []callgraph.CallInfo{}},
		{Function: "Low2"},
	}

	high, medium, low := categorizeFunctions(functions)

	if len(high) != 2 {
		t.Errorf("Expected 2 high impact functions, got %d", len(high))
	}
	if len(medium) != 2 {
		t.Errorf("Expected 2 medium impact functions, got %d", len(medium))
	}
	if len(low) != 2 {
		t.Errorf("Expected 2 low impact functions, got %d", len(low))
	}
}

func TestRenderFunctionImpact_IncludesAllFields(t *testing.T) {
	fcg := callgraph.FunctionCallGraph{
		Function: "TestFunc",
		File:     "test/file.go",
		Callers: []callgraph.CallInfo{
			{Function: "Caller1", File: "caller.go", Line: 10, CallSite: "line 10"},
		},
		Callees: []callgraph.CallInfo{
			{Function: "Callee1", File: "callee.go", Line: 20},
		},
		TestCoverage: []callgraph.TestCoverage{
			{TestFunction: "TestTestFunc", File: "test/file_test.go", Line: 5},
		},
	}

	result := renderFunctionImpact(fcg, "go", "HIGH")

	// Check function name
	if !strings.Contains(result, "### `TestFunc`") {
		t.Error("Expected function name header")
	}

	// Check file
	if !strings.Contains(result, "**File:** `test/file.go`") {
		t.Error("Expected file field")
	}

	// Check risk level
	if !strings.Contains(result, "**Risk Level:** HIGH") {
		t.Error("Expected risk level")
	}

	// Check caller count
	if !strings.Contains(result, "**Callers:** 1") {
		t.Error("Expected caller count")
	}

	// Check call site info
	if !strings.Contains(result, "(call site: line 10)") {
		t.Error("Expected call site info")
	}

	// Check separator
	if !strings.Contains(result, "---") {
		t.Error("Expected separator")
	}
}

// Table-driven tests for categorizeFunctions with edge cases
func TestCategorizeFunctions_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		functions    []callgraph.FunctionCallGraph
		wantHigh     int
		wantMedium   int
		wantLow      int
		descriptions map[string]string // function name -> expected category
	}{
		{
			name:       "empty input",
			functions:  []callgraph.FunctionCallGraph{},
			wantHigh:   0,
			wantMedium: 0,
			wantLow:    0,
		},
		{
			name: "boundary: exactly 3 callers is HIGH",
			functions: []callgraph.FunctionCallGraph{
				{Function: "ExactlyThree", Callers: make([]callgraph.CallInfo, 3)},
			},
			wantHigh:     1,
			wantMedium:   0,
			wantLow:      0,
			descriptions: map[string]string{"ExactlyThree": "HIGH"},
		},
		{
			name: "boundary: exactly 2 callers is MEDIUM",
			functions: []callgraph.FunctionCallGraph{
				{Function: "ExactlyTwo", Callers: make([]callgraph.CallInfo, 2)},
			},
			wantHigh:     0,
			wantMedium:   1,
			wantLow:      0,
			descriptions: map[string]string{"ExactlyTwo": "MEDIUM"},
		},
		{
			name: "boundary: exactly 1 caller is MEDIUM",
			functions: []callgraph.FunctionCallGraph{
				{Function: "ExactlyOne", Callers: make([]callgraph.CallInfo, 1)},
			},
			wantHigh:     0,
			wantMedium:   1,
			wantLow:      0,
			descriptions: map[string]string{"ExactlyOne": "MEDIUM"},
		},
		{
			name: "boundary: 0 callers is LOW",
			functions: []callgraph.FunctionCallGraph{
				{Function: "ZeroCallers", Callers: []callgraph.CallInfo{}},
			},
			wantHigh:     0,
			wantMedium:   0,
			wantLow:      1,
			descriptions: map[string]string{"ZeroCallers": "LOW"},
		},
		{
			name: "nil callers is LOW",
			functions: []callgraph.FunctionCallGraph{
				{Function: "NilCallers", Callers: nil},
			},
			wantHigh:     0,
			wantMedium:   0,
			wantLow:      1,
			descriptions: map[string]string{"NilCallers": "LOW"},
		},
		{
			name: "large caller count is HIGH",
			functions: []callgraph.FunctionCallGraph{
				{Function: "ManyCallers", Callers: make([]callgraph.CallInfo, 100)},
			},
			wantHigh:     1,
			wantMedium:   0,
			wantLow:      0,
			descriptions: map[string]string{"ManyCallers": "HIGH"},
		},
		{
			name: "mixed categories",
			functions: []callgraph.FunctionCallGraph{
				{Function: "High1", Callers: make([]callgraph.CallInfo, 5)},
				{Function: "High2", Callers: make([]callgraph.CallInfo, 3)},
				{Function: "Medium1", Callers: make([]callgraph.CallInfo, 2)},
				{Function: "Medium2", Callers: make([]callgraph.CallInfo, 1)},
				{Function: "Low1", Callers: []callgraph.CallInfo{}},
				{Function: "Low2", Callers: nil},
			},
			wantHigh:   2,
			wantMedium: 2,
			wantLow:    2,
		},
		{
			name: "all HIGH impact",
			functions: []callgraph.FunctionCallGraph{
				{Function: "A", Callers: make([]callgraph.CallInfo, 10)},
				{Function: "B", Callers: make([]callgraph.CallInfo, 5)},
				{Function: "C", Callers: make([]callgraph.CallInfo, 3)},
			},
			wantHigh:   3,
			wantMedium: 0,
			wantLow:    0,
		},
		{
			name: "all MEDIUM impact",
			functions: []callgraph.FunctionCallGraph{
				{Function: "A", Callers: make([]callgraph.CallInfo, 2)},
				{Function: "B", Callers: make([]callgraph.CallInfo, 1)},
			},
			wantHigh:   0,
			wantMedium: 2,
			wantLow:    0,
		},
		{
			name: "all LOW impact",
			functions: []callgraph.FunctionCallGraph{
				{Function: "A", Callers: []callgraph.CallInfo{}},
				{Function: "B", Callers: nil},
			},
			wantHigh:   0,
			wantMedium: 0,
			wantLow:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			high, medium, low := categorizeFunctions(tt.functions)

			if len(high) != tt.wantHigh {
				t.Errorf("high count = %d, want %d", len(high), tt.wantHigh)
			}
			if len(medium) != tt.wantMedium {
				t.Errorf("medium count = %d, want %d", len(medium), tt.wantMedium)
			}
			if len(low) != tt.wantLow {
				t.Errorf("low count = %d, want %d", len(low), tt.wantLow)
			}

			containsFn := func(list []callgraph.FunctionCallGraph, fn string) bool {
				for _, fcg := range list {
					if fcg.Function == fn {
						return true
					}
				}
				return false
			}

			for fn, expectedCat := range tt.descriptions {
				switch expectedCat {
				case "HIGH":
					if !containsFn(high, fn) {
						t.Errorf("function %q expected in HIGH category, but not found", fn)
					}
				case "MEDIUM":
					if !containsFn(medium, fn) {
						t.Errorf("function %q expected in MEDIUM category, but not found", fn)
					}
				case "LOW":
					if !containsFn(low, fn) {
						t.Errorf("function %q expected in LOW category, but not found", fn)
					}
				default:
					t.Errorf("function %q has unknown expected category %q", fn, expectedCat)
				}
			}
		})
	}
}

// Test renderCallers with no callers
func TestRenderCallers_Empty(t *testing.T) {
	var sb strings.Builder
	renderCallers(&sb, []callgraph.CallInfo{})

	result := sb.String()
	if !strings.Contains(result, "**Direct Callers:** None") {
		t.Error("Expected 'None' for empty callers")
	}
}

// Test renderCallers without call site info
func TestRenderCallers_WithoutCallSite(t *testing.T) {
	var sb strings.Builder
	callers := []callgraph.CallInfo{
		{Function: "TestCaller", File: "test.go", Line: 42},
	}
	renderCallers(&sb, callers)

	result := sb.String()
	if !strings.Contains(result, "`TestCaller` at `test.go:42`") {
		t.Error("Expected caller without call site")
	}
	if strings.Contains(result, "call site:") {
		t.Error("Should not have call site info when empty")
	}
}

// Test renderCallers with call site info
func TestRenderCallers_WithCallSite(t *testing.T) {
	var sb strings.Builder
	callers := []callgraph.CallInfo{
		{Function: "TestCaller", File: "test.go", Line: 42, CallSite: "myFunc()"},
	}
	renderCallers(&sb, callers)

	result := sb.String()
	if !strings.Contains(result, "(call site: myFunc())") {
		t.Error("Expected call site info")
	}
}

// Test renderCallees with no callees
func TestRenderCallees_Empty(t *testing.T) {
	var sb strings.Builder
	renderCallees(&sb, []callgraph.CallInfo{})

	result := sb.String()
	if !strings.Contains(result, "**Calls:** None") {
		t.Error("Expected 'None' for empty callees")
	}
}

// Test renderCallees without file info
func TestRenderCallees_WithoutFileInfo(t *testing.T) {
	var sb strings.Builder
	callees := []callgraph.CallInfo{
		{Function: "ExternalFunc"},
	}
	renderCallees(&sb, callees)

	result := sb.String()
	if !strings.Contains(result, "- `ExternalFunc`\n") {
		t.Error("Expected callee without file info")
	}
	if strings.Contains(result, "at `") {
		t.Error("Should not have file location when not provided")
	}
}

// Test renderCallees with file info
func TestRenderCallees_WithFileInfo(t *testing.T) {
	var sb strings.Builder
	callees := []callgraph.CallInfo{
		{Function: "LocalFunc", File: "local.go", Line: 10},
	}
	renderCallees(&sb, callees)

	result := sb.String()
	if !strings.Contains(result, "`LocalFunc` at `local.go:10`") {
		t.Error("Expected callee with file info")
	}
}

// Test renderTestCoverage with tests
func TestRenderTestCoverage_WithTests(t *testing.T) {
	var sb strings.Builder
	tests := []callgraph.TestCoverage{
		{TestFunction: "TestMyFunc", File: "my_test.go", Line: 15},
		{TestFunction: "TestMyFunc_Edge", File: "my_test.go", Line: 30},
	}
	renderTestCoverage(&sb, tests)

	result := sb.String()
	if !strings.Contains(result, "Has tests") {
		t.Error("Expected test coverage indicator")
	}

	if !strings.Contains(result, "<details>") {
		t.Error("Expected details element")
	}
	if !strings.Contains(result, "TestMyFunc") {
		t.Error("Expected test function name")
	}
	if !strings.Contains(result, "TestMyFunc_Edge") {
		t.Error("Expected second test function name")
	}
}

// Test renderTestCoverage without tests
func TestRenderTestCoverage_NoTests(t *testing.T) {
	var sb strings.Builder
	renderTestCoverage(&sb, []callgraph.TestCoverage{})

	result := sb.String()
	if !strings.Contains(result, "No tests found") {
		t.Error("Expected warning for no tests")
	}
}

// Test renderWarnings with only time budget exceeded
func TestRenderWarnings_OnlyTimeBudget(t *testing.T) {
	var sb strings.Builder
	result := &callgraph.CallGraphResult{
		TimeBudgetExceeded: true,
		PartialResults:     false,
		Warnings:           nil,
	}
	renderWarnings(&sb, result)

	output := sb.String()
	if !strings.Contains(output, "## Warnings") {
		t.Error("Expected Warnings header")
	}
	if !strings.Contains(output, "Time Budget Exceeded") {
		t.Error("Expected time budget warning")
	}
	if strings.Contains(output, "Partial Results") {
		t.Error("Should not have partial results warning")
	}
}

// Test renderWarnings with only partial results
func TestRenderWarnings_OnlyPartialResults(t *testing.T) {
	var sb strings.Builder
	result := &callgraph.CallGraphResult{
		TimeBudgetExceeded: false,
		PartialResults:     true,
		Warnings:           nil,
	}
	renderWarnings(&sb, result)

	output := sb.String()
	if !strings.Contains(output, "Partial Results") {
		t.Error("Expected partial results warning")
	}
	if strings.Contains(output, "Time Budget Exceeded") {
		t.Error("Should not have time budget warning")
	}
}

// Test renderWarnings with only custom warnings
func TestRenderWarnings_OnlyCustomWarnings(t *testing.T) {
	var sb strings.Builder
	result := &callgraph.CallGraphResult{
		TimeBudgetExceeded: false,
		PartialResults:     false,
		Warnings:           []string{"Custom warning 1", "Custom warning 2"},
	}
	renderWarnings(&sb, result)

	output := sb.String()
	if !strings.Contains(output, "Custom warning 1") {
		t.Error("Expected first custom warning")
	}
	if !strings.Contains(output, "Custom warning 2") {
		t.Error("Expected second custom warning")
	}
}

// Test renderWarnings with no warnings at all
func TestRenderWarnings_NoWarnings(t *testing.T) {
	var sb strings.Builder
	result := &callgraph.CallGraphResult{
		TimeBudgetExceeded: false,
		PartialResults:     false,
		Warnings:           nil,
	}
	renderWarnings(&sb, result)

	output := sb.String()
	if output != "" {
		t.Errorf("Expected empty output for no warnings, got: %s", output)
	}
}

// Test renderSummaryTable with zero values
func TestRenderSummaryTable_ZeroValues(t *testing.T) {
	var sb strings.Builder
	result := &callgraph.CallGraphResult{
		Language:          "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{},
		ImpactAnalysis: callgraph.ImpactAnalysis{
			DirectCallers:     0,
			TransitiveCallers: 0,
			AffectedTests:     0,
			AffectedPackages:  []string{},
		},
	}
	renderSummaryTable(&sb, result)

	output := sb.String()
	if !strings.Contains(output, "| Modified Functions | 0 |") {
		t.Error("Expected zero modified functions")
	}
	if !strings.Contains(output, "| Direct Callers | 0 |") {
		t.Error("Expected zero direct callers")
	}
	if strings.Contains(output, "### Affected Packages") {
		t.Error("Should not show affected packages when empty")
	}
}

// Test renderSummaryTable with multiple packages
func TestRenderSummaryTable_WithPackages(t *testing.T) {
	var sb strings.Builder
	result := &callgraph.CallGraphResult{
		Language:          "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{},
		ImpactAnalysis: callgraph.ImpactAnalysis{
			DirectCallers:     5,
			TransitiveCallers: 10,
			AffectedTests:     3,
			AffectedPackages:  []string{"pkg/a", "pkg/b", "pkg/c"},
		},
	}
	renderSummaryTable(&sb, result)

	output := sb.String()
	if !strings.Contains(output, "### Affected Packages") {
		t.Error("Expected affected packages section")
	}
	if !strings.Contains(output, "`pkg/a`") {
		t.Error("Expected first package")
	}
	if !strings.Contains(output, "`pkg/b`") {
		t.Error("Expected second package")
	}
	if !strings.Contains(output, "`pkg/c`") {
		t.Error("Expected third package")
	}
}

// Test RenderImpactSummary with only HIGH impact functions
func TestRenderImpactSummary_OnlyHighImpact(t *testing.T) {
	cgResult := &callgraph.CallGraphResult{
		Language: "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{
			{
				Function: "HighImpactFunc",
				File:     "high.go",
				Callers:  make([]callgraph.CallInfo, 5),
			},
		},
	}

	result := RenderImpactSummary(cgResult)

	if !strings.Contains(result, "## High Impact Functions") {
		t.Error("Expected High Impact section")
	}
	if strings.Contains(result, "## Medium Impact Functions") {
		t.Error("Should not have Medium Impact section")
	}
	if strings.Contains(result, "## Low Impact Functions") {
		t.Error("Should not have Low Impact section")
	}
}

// Test RenderImpactSummary with only MEDIUM impact functions
func TestRenderImpactSummary_OnlyMediumImpact(t *testing.T) {
	cgResult := &callgraph.CallGraphResult{
		Language: "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{
			{
				Function: "MediumImpactFunc",
				File:     "medium.go",
				Callers:  make([]callgraph.CallInfo, 2),
			},
		},
	}

	result := RenderImpactSummary(cgResult)

	if strings.Contains(result, "## High Impact Functions") {
		t.Error("Should not have High Impact section")
	}
	if !strings.Contains(result, "## Medium Impact Functions") {
		t.Error("Expected Medium Impact section")
	}
	if strings.Contains(result, "## Low Impact Functions") {
		t.Error("Should not have Low Impact section")
	}
}

// Test RenderImpactSummary with only LOW impact functions
func TestRenderImpactSummary_OnlyLowImpact(t *testing.T) {
	cgResult := &callgraph.CallGraphResult{
		Language: "go",
		ModifiedFunctions: []callgraph.FunctionCallGraph{
			{
				Function: "LowImpactFunc",
				File:     "low.go",
				Callers:  []callgraph.CallInfo{},
			},
		},
	}

	result := RenderImpactSummary(cgResult)

	if strings.Contains(result, "## High Impact Functions") {
		t.Error("Should not have High Impact section")
	}
	if strings.Contains(result, "## Medium Impact Functions") {
		t.Error("Should not have Medium Impact section")
	}
	if !strings.Contains(result, "## Low Impact Functions") {
		t.Error("Expected Low Impact section")
	}
}

// Test RenderImpactSummary with different languages
func TestRenderImpactSummary_DifferentLanguages(t *testing.T) {
	languages := []string{"go", "typescript", "python", "rust", "java"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			cgResult := &callgraph.CallGraphResult{
				Language:          lang,
				ModifiedFunctions: []callgraph.FunctionCallGraph{},
			}

			result := RenderImpactSummary(cgResult)

			expected := "**Language:** " + lang
			if !strings.Contains(result, expected) {
				t.Errorf("Expected language %s in output", lang)
			}
		})
	}
}

// Test renderFunctionImpact with no callers and no callees
func TestRenderFunctionImpact_NoCallersNoCallees(t *testing.T) {
	fcg := callgraph.FunctionCallGraph{
		Function:     "IsolatedFunc",
		File:         "isolated.go",
		Callers:      []callgraph.CallInfo{},
		Callees:      []callgraph.CallInfo{},
		TestCoverage: []callgraph.TestCoverage{},
	}

	result := renderFunctionImpact(fcg, "go", "LOW")

	if !strings.Contains(result, "**Direct Callers:** None") {
		t.Error("Expected 'None' for callers")
	}
	if !strings.Contains(result, "**Calls:** None") {
		t.Error("Expected 'None' for callees")
	}
	if !strings.Contains(result, "No tests found") {
		t.Error("Expected warning for no tests")
	}
}

// Test renderFunctionImpact with different risk levels
func TestRenderFunctionImpact_RiskLevels(t *testing.T) {
	tests := []struct {
		riskLevel string
	}{
		{"HIGH"},
		{"MEDIUM"},
		{"LOW"},
	}

	for _, tt := range tests {
		t.Run(tt.riskLevel, func(t *testing.T) {
			fcg := callgraph.FunctionCallGraph{
				Function: "TestFunc",
				File:     "test.go",
			}

			result := renderFunctionImpact(fcg, "go", tt.riskLevel)

			expected := "**Risk Level:** " + tt.riskLevel
			if !strings.Contains(result, expected) {
				t.Errorf("Expected risk level %s", tt.riskLevel)
			}
		})
	}
}
