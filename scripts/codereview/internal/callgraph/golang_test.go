package callgraph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAffectedPackages(t *testing.T) {
	tests := []struct {
		name     string
		funcs    []ModifiedFunction
		expected []string
	}{
		{
			name:     "empty input returns empty slice",
			funcs:    []ModifiedFunction{},
			expected: []string{},
		},
		{
			name:     "nil input returns empty slice",
			funcs:    nil,
			expected: []string{},
		},
		{
			name: "single package",
			funcs: []ModifiedFunction{
				{Name: "Foo", Package: "github.com/example/pkg"},
			},
			expected: []string{"github.com/example/pkg"},
		},
		{
			name: "multiple packages",
			funcs: []ModifiedFunction{
				{Name: "Foo", Package: "github.com/example/pkg1"},
				{Name: "Bar", Package: "github.com/example/pkg2"},
				{Name: "Baz", Package: "github.com/example/pkg3"},
			},
			expected: []string{
				"github.com/example/pkg1",
				"github.com/example/pkg2",
				"github.com/example/pkg3",
			},
		},
		{
			name: "duplicate packages are deduplicated",
			funcs: []ModifiedFunction{
				{Name: "Foo", Package: "github.com/example/pkg"},
				{Name: "Bar", Package: "github.com/example/pkg"},
				{Name: "Baz", Package: "github.com/example/pkg"},
			},
			expected: []string{"github.com/example/pkg"},
		},
		{
			name: "mixed duplicates and unique packages",
			funcs: []ModifiedFunction{
				{Name: "Foo", Package: "github.com/example/pkg1"},
				{Name: "Bar", Package: "github.com/example/pkg2"},
				{Name: "Baz", Package: "github.com/example/pkg1"},
				{Name: "Qux", Package: "github.com/example/pkg3"},
				{Name: "Quux", Package: "github.com/example/pkg2"},
			},
			expected: []string{
				"github.com/example/pkg1",
				"github.com/example/pkg2",
				"github.com/example/pkg3",
			},
		},
		{
			name: "empty package names are skipped",
			funcs: []ModifiedFunction{
				{Name: "Foo", Package: ""},
				{Name: "Bar", Package: "github.com/example/pkg"},
				{Name: "Baz", Package: ""},
			},
			expected: []string{"github.com/example/pkg"},
		},
		{
			name: "all empty package names returns empty slice",
			funcs: []ModifiedFunction{
				{Name: "Foo", Package: ""},
				{Name: "Bar", Package: ""},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAffectedPackages(tt.funcs)

			// Handle nil vs empty slice comparison
			if len(tt.expected) == 0 {
				if len(result) != 0 {
					t.Errorf("expected empty slice, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d packages, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			// Create a map for expected values to check presence
			expectedMap := make(map[string]bool)
			for _, pkg := range tt.expected {
				expectedMap[pkg] = true
			}

			for _, pkg := range result {
				if !expectedMap[pkg] {
					t.Errorf("unexpected package in result: %s", pkg)
				}
			}
		})
	}
}

func TestAnalyze_GoAnalyzer_Basic(t *testing.T) {
	workDir := t.TempDir()

	module := `module example.com/test

go 1.20
`
	if err := os.WriteFile(filepath.Join(workDir, "go.mod"), []byte(module), 0o644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	source := `package calc

func Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}

func UseAdd() int {
	return Add(1, 2)
}

func UseMultiply() int {
	return Multiply(2, 3)
}
`
	pkgDir := filepath.Join(workDir, "calc")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "calc.go"), []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write calc.go: %v", err)
	}

	testSource := `package calc

import "testing"

func TestUseAdd(t *testing.T) {
	if UseAdd() == 0 {
		t.Fatal("unexpected")
	}
}
`
	if err := os.WriteFile(filepath.Join(pkgDir, "calc_test.go"), []byte(testSource), 0o644); err != nil {
		t.Fatalf("failed to write calc_test.go: %v", err)
	}

	analyzer := NewGoAnalyzer(workDir)
	timeBudgetSec := 30
	result, err := analyzer.Analyze([]ModifiedFunction{{Name: "Add", File: filepath.Join("calc", "calc.go")}}, timeBudgetSec)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Analyze returned nil result")
	}

	if result.Language != "go" {
		t.Errorf("Language = %q, want %q", result.Language, "go")
	}

	if len(result.ModifiedFunctions) != 1 {
		t.Fatalf("expected 1 modified function, got %d", len(result.ModifiedFunctions))
	}

	fcg := result.ModifiedFunctions[0]
	if fcg.Function != "Add" {
		t.Errorf("Function = %q, want %q", fcg.Function, "Add")
	}

	if len(fcg.Callers) == 0 {
		t.Errorf("expected Add to have callers, got none")
	}

	hasUseAdd := false
	for _, caller := range fcg.Callers {
		if caller.Function == "UseAdd" {
			hasUseAdd = true
			break
		}
	}
	if !hasUseAdd {
		t.Errorf("expected UseAdd to be a caller of Add")
	}

	if len(fcg.TestCoverage) == 0 {
		t.Logf("Warning: expected test coverage for Add, got none")
	}

}

func TestIsTestFunction(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		expected bool
	}{
		// Test prefix cases
		{
			name:     "Test prefix - simple",
			funcName: "TestFoo",
			expected: true,
		},
		{
			name:     "Test prefix - with underscores",
			funcName: "Test_Foo_Bar",
			expected: true,
		},
		{
			name:     "Test prefix - exact match",
			funcName: "Test",
			expected: true,
		},
		// Benchmark prefix cases
		{
			name:     "Benchmark prefix - simple",
			funcName: "BenchmarkFoo",
			expected: true,
		},
		{
			name:     "Benchmark prefix - exact match",
			funcName: "Benchmark",
			expected: true,
		},
		// Example prefix cases
		{
			name:     "Example prefix - simple",
			funcName: "ExampleFoo",
			expected: true,
		},
		{
			name:     "Example prefix - exact match",
			funcName: "Example",
			expected: true,
		},
		// Fuzz prefix cases
		{
			name:     "Fuzz prefix - simple",
			funcName: "FuzzFoo",
			expected: true,
		},
		{
			name:     "Fuzz prefix - exact match",
			funcName: "Fuzz",
			expected: true,
		},
		// Regular functions (non-test)
		{
			name:     "regular function - simple",
			funcName: "Foo",
			expected: false,
		},
		{
			name:     "regular function - contains test word",
			funcName: "ContainsTest",
			expected: false,
		},
		{
			name:     "regular function - lowercase test prefix",
			funcName: "testFoo",
			expected: false,
		},
		{
			name:     "regular function - lowercase benchmark prefix",
			funcName: "benchmarkFoo",
			expected: false,
		},
		{
			name:     "regular function - empty string",
			funcName: "",
			expected: false,
		},
		{
			name:     "Test prefix - Testing matches because it starts with Test",
			funcName: "Testing",
			expected: true,
		},
		// Method receiver cases (with dot notation)
		{
			name:     "method with Test suffix after receiver",
			funcName: "SomeType.TestMethod",
			expected: true,
		},
		{
			name:     "method with Benchmark suffix after receiver",
			funcName: "SomeType.BenchmarkMethod",
			expected: true,
		},
		{
			name:     "method with regular name after receiver",
			funcName: "SomeType.RegularMethod",
			expected: false,
		},
		{
			name:     "method with pointer receiver - Test",
			funcName: "*SomeType.TestMethod",
			expected: true,
		},
		{
			name:     "nested package path with Test",
			funcName: "pkg.subpkg.TestFunc",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTestFunction(tt.funcName)
			if result != tt.expected {
				t.Errorf("isTestFunction(%q) = %v, expected %v", tt.funcName, result, tt.expected)
			}
		})
	}
}

func TestNewGoAnalyzer(t *testing.T) {
	tests := []struct {
		name            string
		workDir         string
		expectedNonNil  bool
		expectedWorkDir string
	}{
		{
			name:            "creates analyzer with valid work directory",
			workDir:         "/path/to/project",
			expectedNonNil:  true,
			expectedWorkDir: "/path/to/project",
		},
		{
			name:            "creates analyzer with empty work directory",
			workDir:         "",
			expectedNonNil:  true,
			expectedWorkDir: "",
		},
		{
			name:            "creates analyzer with relative path",
			workDir:         "./relative/path",
			expectedNonNil:  true,
			expectedWorkDir: "./relative/path",
		},
		{
			name:            "creates analyzer with current directory",
			workDir:         ".",
			expectedNonNil:  true,
			expectedWorkDir: ".",
		},
		{
			name:            "creates analyzer with nested path",
			workDir:         "/home/user/projects/myapp/src",
			expectedNonNil:  true,
			expectedWorkDir: "/home/user/projects/myapp/src",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewGoAnalyzer(tt.workDir)

			if tt.expectedNonNil && analyzer == nil {
				t.Error("NewGoAnalyzer returned nil, expected non-nil")
				return
			}

			if !tt.expectedNonNil && analyzer != nil {
				t.Error("NewGoAnalyzer returned non-nil, expected nil")
				return
			}

			if analyzer != nil && analyzer.workDir != tt.expectedWorkDir {
				t.Errorf("workDir = %q, expected %q", analyzer.workDir, tt.expectedWorkDir)
			}
		})
	}
}

// TestGoAnalyzerType verifies the GoAnalyzer struct fields and type.
func TestGoAnalyzerType(t *testing.T) {
	analyzer := NewGoAnalyzer("/test/dir")

	// Verify type
	if analyzer == nil {
		t.Fatal("NewGoAnalyzer returned nil")
	}

	// Type assertion should work
	var _ *GoAnalyzer = analyzer
}
