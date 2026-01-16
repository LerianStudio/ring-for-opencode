package callgraph

import (
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
