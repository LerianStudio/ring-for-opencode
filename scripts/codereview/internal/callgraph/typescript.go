package callgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Resource protection limits for TypeScript analysis.
const (
	tsMaxModifiedFunctions = 500              // Maximum number of modified functions to analyze
	tsDefaultTimeBudgetSec = 120              // Default time budget when not specified
	tsMaxOutputSize        = 50 * 1024 * 1024 // 50MB limit for subprocess output
)

// TypeScriptAnalyzer implements call graph analysis for TypeScript code.
type TypeScriptAnalyzer struct {
	workDir string
}

// NewTypeScriptAnalyzer creates a new TypeScript call graph analyzer.
// workDir is the root directory for module resolution.
func NewTypeScriptAnalyzer(workDir string) *TypeScriptAnalyzer {
	return &TypeScriptAnalyzer{
		workDir: workDir,
	}
}

// sanitizeFilePaths validates file paths to prevent command injection.
func (t *TypeScriptAnalyzer) sanitizeFilePaths(files []string) ([]string, error) {
	var sanitized []string
	for _, f := range files {
		if strings.HasPrefix(f, "-") {
			return nil, fmt.Errorf("invalid file path (starts with dash): %s", f)
		}
		if strings.ContainsRune(f, '\x00') {
			return nil, fmt.Errorf("invalid file path (contains null byte)")
		}
		sanitized = append(sanitized, f)
	}
	return sanitized, nil
}

// depCruiserOutput represents the top-level output from dependency-cruiser.
type depCruiserOutput struct {
	Modules []depCruiserModule `json:"modules"`
}

// depCruiserModule represents a module in the dependency-cruiser output.
type depCruiserModule struct {
	Source       string                 `json:"source"`
	Dependencies []depCruiserDependency `json:"dependencies"`
}

// depCruiserDependency represents a dependency relationship.
type depCruiserDependency struct {
	Module       string `json:"module"`
	Resolved     string `json:"resolved"`
	ModuleSystem string `json:"moduleSystem"`
	Dynamic      bool   `json:"dynamic"`
}

// tsHelperOutput represents the output from the TypeScript call-graph helper.
type tsHelperOutput struct {
	Functions []tsHelperFunction `json:"functions"`
	Error     string             `json:"error,omitempty"`
}

// tsHelperFunction represents function-level call information from the helper.
type tsHelperFunction struct {
	Name      string           `json:"name"`
	File      string           `json:"file"`
	Line      int              `json:"line"`
	CallSites []tsHelperCall   `json:"call_sites"`
	CalledBy  []tsHelperCaller `json:"called_by,omitempty"`
}

// tsHelperCall represents a call made from a function.
type tsHelperCall struct {
	Target   string `json:"target"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	IsMethod bool   `json:"is_method"`
}

// tsHelperCaller represents a caller of a function (when using --functions flag).
type tsHelperCaller struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Analyze implements the Analyzer interface for TypeScript code.
func (t *TypeScriptAnalyzer) Analyze(modifiedFuncs []ModifiedFunction, timeBudgetSec int) (*CallGraphResult, error) {
	result := &CallGraphResult{
		Language:          "typescript",
		ModifiedFunctions: make([]FunctionCallGraph, 0, len(modifiedFuncs)),
		ImpactAnalysis: ImpactAnalysis{
			AffectedPackages: t.getAffectedPackages(modifiedFuncs),
		},
	}

	if len(modifiedFuncs) == 0 {
		return result, nil
	}

	// Input validation: truncate if exceeding limit
	if len(modifiedFuncs) > tsMaxModifiedFunctions {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Truncated modified functions from %d to %d", len(modifiedFuncs), tsMaxModifiedFunctions))
		modifiedFuncs = modifiedFuncs[:tsMaxModifiedFunctions]
		result.PartialResults = true
	}

	// Apply default time budget if not specified
	if timeBudgetSec <= 0 {
		timeBudgetSec = tsDefaultTimeBudgetSec
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeBudgetSec)*time.Second)
	defer cancel()

	// Collect unique files to analyze
	files := t.collectUniqueFiles(modifiedFuncs)

	// Try to use the TypeScript helper for detailed analysis
	helperResult, helperErr := t.analyzeWithTSHelper(ctx, files, modifiedFuncs)
	if helperErr != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("TypeScript helper unavailable: %v", helperErr))
		// Fall back to dependency-cruiser based analysis
		return t.analyzeWithDepCruiser(ctx, modifiedFuncs, result)
	}

	// Process helper results
	return t.processHelperResults(helperResult, modifiedFuncs, result)
}

// getAffectedPackages extracts unique package paths from modified functions.
func (t *TypeScriptAnalyzer) getAffectedPackages(funcs []ModifiedFunction) []string {
	seen := make(map[string]bool)
	var result []string

	for _, fn := range funcs {
		pkg := fn.Package
		if pkg == "" {
			// Extract package from file path
			pkg = filepath.Dir(fn.File)
		}
		if pkg != "" && !seen[pkg] {
			seen[pkg] = true
			result = append(result, pkg)
		}
	}

	return result
}

// collectUniqueFiles returns unique file paths from modified functions.
func (t *TypeScriptAnalyzer) collectUniqueFiles(funcs []ModifiedFunction) []string {
	seen := make(map[string]bool)
	var files []string

	for _, fn := range funcs {
		if fn.File != "" && !seen[fn.File] {
			seen[fn.File] = true
			files = append(files, fn.File)
		}
	}

	return files
}

// analyzeWithTSHelper uses the Node.js TypeScript helper for detailed analysis.
func (t *TypeScriptAnalyzer) analyzeWithTSHelper(ctx context.Context, files []string, modifiedFuncs []ModifiedFunction) (*tsHelperOutput, error) {
	if len(files) == 0 {
		return &tsHelperOutput{}, nil
	}

	// Sanitize file paths to prevent command injection
	sanitizedFiles, err := t.sanitizeFilePaths(files)
	if err != nil {
		return nil, fmt.Errorf("file path validation failed: %w", err)
	}
	files = sanitizedFiles

	// Build function filter for --functions flag
	var funcNames []string
	for _, fn := range modifiedFuncs {
		funcNames = append(funcNames, fn.Name)
	}

	// Locate the helper script - try relative to workDir first, then common locations
	helperPath := t.findHelperScript()
	if helperPath == "" {
		return nil, fmt.Errorf("call-graph.ts helper script not found")
	}

	// Build command arguments
	args := []string{helperPath}
	args = append(args, files...)
	if len(funcNames) > 0 {
		args = append(args, "--functions", strings.Join(funcNames, ","))
	}

	// Try npx ts-node first, then node with compiled JS
	var cmd *exec.Cmd
	if strings.HasSuffix(helperPath, ".ts") {
		cmd = exec.CommandContext(ctx, "npx", append([]string{"ts-node"}, args...)...)
	} else {
		cmd = exec.CommandContext(ctx, "node", args...)
	}
	cmd.Dir = t.workDir

	output, err := cmd.Output()
	if err != nil {
		// Try with compiled version if ts-node failed
		jsPath := strings.TrimSuffix(helperPath, ".ts") + ".js"
		distPath := filepath.Join(filepath.Dir(helperPath), "dist", "call-graph.js")

		for _, altPath := range []string{jsPath, distPath} {
			// Include files and --functions flag in retry
			altArgs := append([]string{altPath}, files...)
			if len(funcNames) > 0 {
				altArgs = append(altArgs, "--functions", strings.Join(funcNames, ","))
			}
			cmd = exec.CommandContext(ctx, "node", altArgs...)
			cmd.Dir = t.workDir
			output, err = cmd.Output()
			if err == nil {
				break
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to run TypeScript helper: %w", err)
		}
	}

	// Check output size limit to prevent memory exhaustion
	if len(output) > tsMaxOutputSize {
		return nil, fmt.Errorf("helper output exceeds size limit (%d > %d bytes)", len(output), tsMaxOutputSize)
	}

	var helperOutput tsHelperOutput
	if err := json.Unmarshal(output, &helperOutput); err != nil {
		return nil, fmt.Errorf("failed to parse helper output: %w", err)
	}

	if helperOutput.Error != "" {
		return nil, fmt.Errorf("helper error: %s", helperOutput.Error)
	}

	return &helperOutput, nil
}

// findHelperScript locates the call-graph.ts helper script.
func (t *TypeScriptAnalyzer) findHelperScript() string {
	// Search paths relative to workDir and common locations
	searchPaths := []string{
		filepath.Join(t.workDir, "ts", "call-graph.ts"),
		filepath.Join(t.workDir, "ts", "dist", "call-graph.js"),
		filepath.Join(t.workDir, "scripts", "codereview", "ts", "call-graph.ts"),
		filepath.Join(t.workDir, "scripts", "codereview", "ts", "dist", "call-graph.js"),
		filepath.Join(t.workDir, "..", "ts", "call-graph.ts"),
		filepath.Join(t.workDir, "..", "ts", "dist", "call-graph.js"),
	}

	for _, path := range searchPaths {
		if fileExists(path) {
			return path
		}
	}

	return ""
}

// processHelperResults converts TypeScript helper output to CallGraphResult.
func (t *TypeScriptAnalyzer) processHelperResults(helper *tsHelperOutput, modifiedFuncs []ModifiedFunction, result *CallGraphResult) (*CallGraphResult, error) {
	// Build lookup maps
	funcLookup := make(map[string]*tsHelperFunction)
	for i := range helper.Functions {
		fn := &helper.Functions[i]
		key := fmt.Sprintf("%s:%s", fn.File, fn.Name)
		funcLookup[key] = fn
	}

	allDirectCallers := make(map[string]bool)
	allAffectedTests := make(map[string]bool)

	for _, modFunc := range modifiedFuncs {
		fcg := FunctionCallGraph{
			Function:     modFunc.Name,
			File:         modFunc.File,
			Callers:      make([]CallInfo, 0),
			Callees:      make([]CallInfo, 0),
			TestCoverage: make([]TestCoverage, 0),
		}

		// Find matching function in helper output
		key := fmt.Sprintf("%s:%s", modFunc.File, modFunc.Name)
		helperFunc := funcLookup[key]

		if helperFunc != nil {
			// Add callees from call sites
			for _, call := range helperFunc.CallSites {
				fcg.Callees = append(fcg.Callees, CallInfo{
					Function: call.Target,
					File:     modFunc.File,
					Line:     call.Line,
					CallSite: fmt.Sprintf("%s:%d", filepath.Base(modFunc.File), call.Line),
				})
			}

			// Add callers
			for _, caller := range helperFunc.CalledBy {
				callerKey := fmt.Sprintf("%s:%s", caller.File, caller.Function)
				allDirectCallers[callerKey] = true

				// Check if caller is a test
				if isTypeScriptTestFile(caller.File) || isTypeScriptTestFunction(caller.Function) {
					allAffectedTests[callerKey] = true
					fcg.TestCoverage = append(fcg.TestCoverage, TestCoverage{
						TestFunction: caller.Function,
						File:         caller.File,
						Line:         caller.Line,
					})
				}

				fcg.Callers = append(fcg.Callers, CallInfo{
					Function: caller.Function,
					File:     caller.File,
					Line:     caller.Line,
				})
			}
		}

		result.ModifiedFunctions = append(result.ModifiedFunctions, fcg)
	}

	result.ImpactAnalysis.DirectCallers = len(allDirectCallers)
	result.ImpactAnalysis.AffectedTests = len(allAffectedTests)

	return result, nil
}

// analyzeWithDepCruiser performs module-level analysis using dependency-cruiser.
func (t *TypeScriptAnalyzer) analyzeWithDepCruiser(ctx context.Context, modifiedFuncs []ModifiedFunction, result *CallGraphResult) (*CallGraphResult, error) {
	// Collect files to analyze
	files := t.collectUniqueFiles(modifiedFuncs)
	if len(files) == 0 {
		return result, nil
	}

	// Run dependency-cruiser for each file
	depGraph, err := t.runDepCruiser(ctx, files)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("dependency-cruiser unavailable: %v", err))
		// Return partial results without dependency information
		for _, modFunc := range modifiedFuncs {
			result.ModifiedFunctions = append(result.ModifiedFunctions, FunctionCallGraph{
				Function:     modFunc.Name,
				File:         modFunc.File,
				Callers:      make([]CallInfo, 0),
				Callees:      make([]CallInfo, 0),
				TestCoverage: make([]TestCoverage, 0),
			})
		}
		return result, nil
	}

	// Build module dependency map
	moduleDeps := make(map[string]*depCruiserModule)
	for i := range depGraph.Modules {
		mod := &depGraph.Modules[i]
		moduleDeps[mod.Source] = mod
	}

	allDirectCallers := make(map[string]bool)
	allAffectedTests := make(map[string]bool)

	for _, modFunc := range modifiedFuncs {
		fcg := FunctionCallGraph{
			Function:     modFunc.Name,
			File:         modFunc.File,
			Callers:      make([]CallInfo, 0),
			Callees:      make([]CallInfo, 0),
			TestCoverage: make([]TestCoverage, 0),
		}

		// Find callees (modules this file imports)
		callees := t.findCallees(moduleDeps, modFunc.File)
		for _, callee := range callees {
			fcg.Callees = append(fcg.Callees, CallInfo{
				Function: filepath.Base(callee),
				File:     callee,
			})
		}

		// Find callers (modules that import this file)
		callers := t.findCallers(moduleDeps, modFunc.File)
		for _, caller := range callers {
			callerKey := caller
			allDirectCallers[callerKey] = true

			// Check if caller is a test file
			if isTypeScriptTestFile(caller) {
				allAffectedTests[callerKey] = true
				fcg.TestCoverage = append(fcg.TestCoverage, TestCoverage{
					TestFunction: filepath.Base(caller),
					File:         caller,
				})
			}

			fcg.Callers = append(fcg.Callers, CallInfo{
				Function: filepath.Base(caller),
				File:     caller,
			})
		}

		// Also find test coverage specifically
		testCoverage := t.findTestCoverage(moduleDeps, modFunc.File)
		for _, testFile := range testCoverage {
			if !allAffectedTests[testFile] {
				allAffectedTests[testFile] = true
				fcg.TestCoverage = append(fcg.TestCoverage, TestCoverage{
					TestFunction: filepath.Base(testFile),
					File:         testFile,
				})
			}
		}

		result.ModifiedFunctions = append(result.ModifiedFunctions, fcg)
	}

	result.ImpactAnalysis.DirectCallers = len(allDirectCallers)
	result.ImpactAnalysis.AffectedTests = len(allAffectedTests)

	return result, nil
}

// runDepCruiser executes dependency-cruiser and returns the parsed output.
func (t *TypeScriptAnalyzer) runDepCruiser(ctx context.Context, files []string) (*depCruiserOutput, error) {
	// Sanitize file paths to prevent command injection
	sanitizedFiles, err := t.sanitizeFilePaths(files)
	if err != nil {
		return nil, fmt.Errorf("file path validation failed: %w", err)
	}

	// Try npx depcruise first, then global depcruise
	args := []string{"depcruise", "--output-type", "json"}
	args = append(args, sanitizedFiles...)

	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Dir = t.workDir

	output, err := cmd.Output()
	if err != nil {
		// Try global installation
		cmd = exec.CommandContext(ctx, "depcruise", append([]string{"--output-type", "json"}, sanitizedFiles...)...)
		cmd.Dir = t.workDir
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("dependency-cruiser not available: %w", err)
		}
	}

	var result depCruiserOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse depcruise output: %w", err)
	}

	return &result, nil
}

// findCallers finds all modules that import the target file.
func (t *TypeScriptAnalyzer) findCallers(moduleDeps map[string]*depCruiserModule, targetFile string) []string {
	var callers []string
	normalizedTarget := normalizeFilePath(targetFile)

	for source, mod := range moduleDeps {
		for _, dep := range mod.Dependencies {
			resolved := normalizeFilePath(dep.Resolved)
			if resolved == normalizedTarget || strings.HasSuffix(resolved, normalizedTarget) {
				callers = append(callers, source)
				break
			}
		}
	}

	return callers
}

// findCallees finds all modules that the target file imports.
func (t *TypeScriptAnalyzer) findCallees(moduleDeps map[string]*depCruiserModule, targetFile string) []string {
	var callees []string
	normalizedTarget := normalizeFilePath(targetFile)

	mod, exists := moduleDeps[normalizedTarget]
	if !exists {
		// Try matching with different path formats
		for source, m := range moduleDeps {
			if strings.HasSuffix(source, normalizedTarget) || strings.HasSuffix(normalizedTarget, source) {
				mod = m
				break
			}
		}
	}

	if mod != nil {
		for _, dep := range mod.Dependencies {
			if dep.Resolved != "" {
				callees = append(callees, dep.Resolved)
			}
		}
	}

	return callees
}

// findTestCoverage finds test files that import the target file.
func (t *TypeScriptAnalyzer) findTestCoverage(moduleDeps map[string]*depCruiserModule, targetFile string) []string {
	callers := t.findCallers(moduleDeps, targetFile)
	var testFiles []string

	for _, caller := range callers {
		if isTypeScriptTestFile(caller) {
			testFiles = append(testFiles, caller)
		}
	}

	return testFiles
}

// isTypeScriptTestFile checks if a file path indicates a test file.
func isTypeScriptTestFile(filePath string) bool {
	base := filepath.Base(filePath)

	// Check for common test file patterns
	patterns := []string{
		".test.",
		".spec.",
		"_test.",
		"_spec.",
	}

	for _, pattern := range patterns {
		if strings.Contains(base, pattern) {
			return true
		}
	}

	// Check if in __tests__ directory
	return strings.Contains(filePath, "__tests__") ||
		strings.Contains(filePath, "/test/") ||
		strings.Contains(filePath, "/tests/")
}

// isTypeScriptTestFunction checks if a function name indicates a test function.
func isTypeScriptTestFunction(funcName string) bool {
	lowerName := strings.ToLower(funcName)
	// Exact matches for test framework functions
	exactMatches := []string{"it", "test", "describe", "beforeeach", "aftereach", "beforeall", "afterall"}
	for _, match := range exactMatches {
		if lowerName == match {
			return true
		}
	}
	// Prefix matches for custom test functions
	return strings.HasPrefix(lowerName, "test_") ||
		strings.HasPrefix(lowerName, "spec_")
}

// normalizeFilePath normalizes a file path for comparison.
func normalizeFilePath(path string) string {
	// Remove leading ./ and normalize separators
	path = strings.TrimPrefix(path, "./")
	path = filepath.Clean(path)
	return path
}
