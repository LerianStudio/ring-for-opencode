package callgraph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
)

// Resource protection limits for Python analysis.
const (
	pyMaxModifiedFunctions = 500              // Maximum number of modified functions to analyze
	pyDefaultTimeBudgetSec = 120              // Default time budget when not specified
	pyMaxOutputSize        = 50 * 1024 * 1024 // 50MB limit for subprocess output
)

// PythonAnalyzer implements call graph analysis for Python code.
type PythonAnalyzer struct {
	workDir string
}

// NewPythonAnalyzer creates a new Python call graph analyzer.
// workDir is the root directory for module resolution.
func NewPythonAnalyzer(workDir string) *PythonAnalyzer {
	return &PythonAnalyzer{
		workDir: workDir,
	}
}

// sanitizeFilePaths validates file paths to prevent command injection and path traversal.
func (p *PythonAnalyzer) sanitizeFilePaths(files []string) ([]string, error) {
	// Resolve workDir to absolute path for comparison
	absWorkDir, err := filepath.Abs(p.workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workDir: %w", err)
	}

	var sanitized []string
	for _, f := range files {
		// Check for dash prefix (command injection)
		if strings.HasPrefix(f, "-") {
			return nil, fmt.Errorf("invalid file path (starts with dash): %s", f)
		}
		// Check for null bytes
		if strings.ContainsRune(f, '\x00') {
			return nil, fmt.Errorf("invalid file path (contains null byte)")
		}

		// Resolve to absolute path
		absPath, err := filepath.Abs(f)
		if err != nil {
			return nil, fmt.Errorf("invalid file path (cannot resolve): %s", f)
		}

		// Evaluate symlinks to prevent escape via symlink
		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			// File might not exist yet, use the resolved absolute path
			realPath = absPath
		}

		// Verify path is within workDir (prevent path traversal)
		if !strings.HasPrefix(realPath, absWorkDir+string(filepath.Separator)) && realPath != absWorkDir {
			return nil, fmt.Errorf("invalid file path (outside work directory): %s", f)
		}

		sanitized = append(sanitized, f)
	}
	return sanitized, nil
}

// validPythonIdentifier is a regex pattern for valid Python identifiers.
// Matches: identifier or module.identifier or Class.method patterns.
var validPythonIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)*$`)

// sanitizeFunctionNames validates function names to prevent command injection.
func (p *PythonAnalyzer) sanitizeFunctionNames(names []string) ([]string, error) {
	var sanitized []string
	for _, name := range names {
		// Reject names starting with dash (command injection)
		if strings.HasPrefix(name, "-") {
			return nil, fmt.Errorf("invalid function name (starts with dash): %s", name)
		}
		// Validate against Python identifier pattern
		if !validPythonIdentifier.MatchString(name) {
			return nil, fmt.Errorf("invalid function name (not a valid identifier): %s", name)
		}
		sanitized = append(sanitized, name)
	}
	return sanitized, nil
}

// pyHelperOutput represents the output from the Python call-graph helper.
type pyHelperOutput struct {
	Functions []pyHelperFunction `json:"functions"`
	Error     string             `json:"error,omitempty"`
}

// pyHelperFunction represents function-level call information from the helper.
type pyHelperFunction struct {
	Name      string           `json:"name"`
	File      string           `json:"file"`
	Line      int              `json:"line"`
	EndLine   int              `json:"end_line"`
	CallSites []pyHelperCall   `json:"call_sites"`
	CalledBy  []pyHelperCaller `json:"called_by,omitempty"`
}

// pyHelperCall represents a call made from a function.
type pyHelperCall struct {
	Target   string `json:"target"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	IsMethod bool   `json:"is_method"`
}

// pyHelperCaller represents a caller of a function (when using --functions flag).
type pyHelperCaller struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Analyze implements the Analyzer interface for Python code.
func (p *PythonAnalyzer) Analyze(modifiedFuncs []ModifiedFunction, timeBudgetSec int) (*CallGraphResult, error) {
	result := &CallGraphResult{
		Language:          "python",
		ModifiedFunctions: make([]FunctionCallGraph, 0, len(modifiedFuncs)),
		Warnings:          []string{},
		ImpactAnalysis: ImpactAnalysis{
			AffectedPackages: p.getAffectedPackages(modifiedFuncs),
		},
	}

	if len(modifiedFuncs) == 0 {
		return result, nil
	}

	// Input validation: truncate if exceeding limit
	if len(modifiedFuncs) > pyMaxModifiedFunctions {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Truncated modified functions from %d to %d", len(modifiedFuncs), pyMaxModifiedFunctions))
		modifiedFuncs = modifiedFuncs[:pyMaxModifiedFunctions]
		result.PartialResults = true
	}

	// Apply default time budget if not specified
	if timeBudgetSec <= 0 {
		timeBudgetSec = pyDefaultTimeBudgetSec
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeBudgetSec)*time.Second)
	defer cancel()

	// Collect unique files to analyze
	files := p.collectUniqueFiles(modifiedFuncs)

	// Try to use the custom Python helper for detailed analysis
	helperResult, helperErr := p.runCallGraphPy(ctx, files, modifiedFuncs)
	if helperErr != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Python helper unavailable: %v", helperErr))
		// Fall back to pyan3
		return p.runPyan3(ctx, modifiedFuncs, result)
	}

	// Process helper results
	return p.processHelperResults(helperResult, modifiedFuncs, result)
}

// getAffectedPackages extracts unique package paths from modified functions.
func (p *PythonAnalyzer) getAffectedPackages(funcs []ModifiedFunction) []string {
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
func (p *PythonAnalyzer) collectUniqueFiles(funcs []ModifiedFunction) []string {
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

type pyHelperOutputTooLargeError struct {
	size  int
	limit int
}

func (e *pyHelperOutputTooLargeError) Error() string {
	return fmt.Sprintf("helper output exceeds size limit (%d > %d bytes)", e.size, e.limit)
}

func (p *PythonAnalyzer) runPythonHelperCommand(ctx context.Context, pythonBinary string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, pythonBinary, args...) // #nosec G204 - args sanitized
	cmd.Dir = p.workDir
	cmd.Env = append([]string{"LC_ALL=C"}, os.Environ()...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	limitedStdout := io.LimitReader(stdout, pyMaxOutputSize+1)
	output, readErr := io.ReadAll(limitedStdout)
	if readErr != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		return nil, readErr
	}

	// If we hit the limit, STOP: do not keep reading (risk hang). Kill + Wait to reap.
	if len(output) > pyMaxOutputSize {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		return nil, &pyHelperOutputTooLargeError{size: len(output), limit: pyMaxOutputSize}
	}

	waitErr := cmd.Wait()
	if exitErr, ok := waitErr.(*exec.ExitError); ok {
		// Preserve Cmd.Output() behavior: attach captured stderr to ExitError.
		exitErr.Stderr = stderr.Bytes()
	}
	if waitErr != nil {
		return nil, waitErr
	}

	return output, nil
}

// runCallGraphPy uses the custom Python call_graph.py script for detailed analysis.
func (p *PythonAnalyzer) runCallGraphPy(ctx context.Context, files []string, modifiedFuncs []ModifiedFunction) (*pyHelperOutput, error) {
	if len(files) == 0 {
		return &pyHelperOutput{}, nil
	}

	// Sanitize file paths to prevent command injection
	sanitizedFiles, err := p.sanitizeFilePaths(files)
	if err != nil {
		return nil, fmt.Errorf("file path validation failed: %w", err)
	}
	files = sanitizedFiles

	// Build function filter for --functions flag
	var funcNames []string
	for _, fn := range modifiedFuncs {
		funcNames = append(funcNames, fn.Name)
	}

	// Sanitize function names to prevent command injection
	if len(funcNames) > 0 {
		sanitizedFuncs, err := p.sanitizeFunctionNames(funcNames)
		if err != nil {
			return nil, fmt.Errorf("function name validation failed: %w", err)
		}
		funcNames = sanitizedFuncs
	}

	// Locate the helper script
	helperPath := p.findHelperScript()
	if helperPath == "" {
		return nil, fmt.Errorf("call_graph.py helper script not found")
	}
	validatedHelper, err := fileutil.ValidatePath(helperPath, ".")
	if err != nil {
		return nil, fmt.Errorf("helper script path invalid: %w", err)
	}
	helperPath = validatedHelper

	// Build command arguments
	args := []string{helperPath}
	args = append(args, files...)
	if len(funcNames) > 0 {
		args = append(args, "--functions", strings.Join(funcNames, ","))
	}

	// Try python3 first, then python
	output, err := p.runPythonHelperCommand(ctx, "python3", args)
	if err != nil {
		var tooLarge *pyHelperOutputTooLargeError
		if errors.As(err, &tooLarge) {
			return nil, err
		}
		// Try with python if python3 failed
		output, err = p.runPythonHelperCommand(ctx, "python", args)
		if err != nil {
			if errors.As(err, &tooLarge) {
				return nil, err
			}
			return nil, fmt.Errorf("failed to run Python helper: %w", err)
		}
	}

	// Check output size limit to prevent memory exhaustion
	if len(output) > pyMaxOutputSize {
		return nil, fmt.Errorf("helper output exceeds size limit (%d > %d bytes)", len(output), pyMaxOutputSize)
	}

	var helperOutput pyHelperOutput
	if err := json.Unmarshal(output, &helperOutput); err != nil {
		return nil, fmt.Errorf("failed to parse helper output: %w", err)
	}
	if helperOutput.Functions == nil {
		helperOutput.Functions = []pyHelperFunction{}
	}

	if helperOutput.Error != "" {
		return nil, fmt.Errorf("helper error: %s", helperOutput.Error)
	}

	return &helperOutput, nil
}

// findHelperScript locates the call_graph.py helper script.
func (p *PythonAnalyzer) findHelperScript() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	binDir := filepath.Dir(execPath)
	rootDir := filepath.Dir(binDir)
	searchPaths := []string{
		filepath.Join(rootDir, "py", "call_graph.py"),
		filepath.Join(rootDir, "scripts", "codereview", "py", "call_graph.py"),
	}

	for _, path := range searchPaths {
		cleaned := filepath.Clean(path)
		if !strings.HasPrefix(cleaned, rootDir+string(filepath.Separator)) && cleaned != rootDir {
			continue
		}
		if fileExists(cleaned) {
			return cleaned
		}
	}

	return ""
}

// processHelperResults converts Python helper output to CallGraphResult.
func (p *PythonAnalyzer) processHelperResults(helper *pyHelperOutput, modifiedFuncs []ModifiedFunction, result *CallGraphResult) (*CallGraphResult, error) {
	// Build lookup maps
	funcLookup := make(map[string]*pyHelperFunction)
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
				if isPythonTestFile(caller.File) || isPythonTestFunction(caller.Function) {
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

// runPyan3 uses pyan3 as fallback for call graph analysis.
func (p *PythonAnalyzer) runPyan3(ctx context.Context, modifiedFuncs []ModifiedFunction, result *CallGraphResult) (*CallGraphResult, error) {
	// Collect files to analyze
	files := p.collectUniqueFiles(modifiedFuncs)
	if len(files) == 0 {
		return result, nil
	}

	// Sanitize file paths
	sanitizedFiles, err := p.sanitizeFilePaths(files)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("file path validation failed: %v", err))
		return p.returnEmptyResults(modifiedFuncs, result), nil
	}

	// Try to run pyan3
	args := []string{"-e", "from pyan import main; main()", "--dot"}
	args = append(args, sanitizedFiles...)

	cmd := exec.CommandContext(ctx, "python3", args...) // #nosec G204 - args sanitized
	cmd.Dir = p.workDir
	cmd.Env = append([]string{"LC_ALL=C"}, os.Environ()...)

	output, err := cmd.Output()
	if err != nil {
		// Try with pip-installed pyan3
		pyanArgs := append([]string{"--dot"}, sanitizedFiles...)
		cmd = exec.CommandContext(ctx, "pyan3", pyanArgs...) // #nosec G204 - args sanitized
		cmd.Dir = p.workDir
		cmd.Env = append([]string{"LC_ALL=C"}, os.Environ()...)
		output, err = cmd.Output()
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("pyan3 not available: %v", err))
			return p.returnEmptyResults(modifiedFuncs, result), nil
		}
	}

	// Check output size limit
	if len(output) > pyMaxOutputSize {
		result.Warnings = append(result.Warnings, "pyan3 output exceeds size limit")
		return p.returnEmptyResults(modifiedFuncs, result), nil
	}

	// Parse DOT output (basic parsing for caller/callee relationships)
	return p.parsePyanOutput(string(output), modifiedFuncs, result)
}

// parsePyanOutput parses pyan3 DOT format output.
func (p *PythonAnalyzer) parsePyanOutput(dotOutput string, modifiedFuncs []ModifiedFunction, result *CallGraphResult) (*CallGraphResult, error) {
	// Build a map of edges from DOT output
	// DOT format: "caller" -> "callee";
	edges := make(map[string][]string) // caller -> callees

	lines := strings.Split(dotOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "->") {
			continue
		}

		// Parse edge line
		parts := strings.Split(line, "->")
		if len(parts) != 2 {
			continue
		}

		caller := strings.Trim(strings.TrimSpace(parts[0]), "\"")
		callee := strings.Trim(strings.TrimSpace(strings.TrimSuffix(parts[1], ";")), "\"")

		if caller != "" && callee != "" {
			edges[caller] = append(edges[caller], callee)
		}
	}

	// Build reverse map for callers
	callerMap := make(map[string][]string) // callee -> callers
	for caller, callees := range edges {
		for _, callee := range callees {
			callerMap[callee] = append(callerMap[callee], caller)
		}
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

		// Find callees
		if callees, ok := edges[modFunc.Name]; ok {
			for _, callee := range callees {
				fcg.Callees = append(fcg.Callees, CallInfo{
					Function: callee,
				})
			}
		}

		// Find callers
		if callers, ok := callerMap[modFunc.Name]; ok {
			for _, caller := range callers {
				allDirectCallers[caller] = true

				// Check if caller is a test
				if isPythonTestFunction(caller) {
					allAffectedTests[caller] = true
					fcg.TestCoverage = append(fcg.TestCoverage, TestCoverage{
						TestFunction: caller,
					})
				}

				fcg.Callers = append(fcg.Callers, CallInfo{
					Function: caller,
				})
			}
		}

		result.ModifiedFunctions = append(result.ModifiedFunctions, fcg)
	}

	result.ImpactAnalysis.DirectCallers = len(allDirectCallers)
	result.ImpactAnalysis.AffectedTests = len(allAffectedTests)

	return result, nil
}

// returnEmptyResults returns partial results without call graph information.
func (p *PythonAnalyzer) returnEmptyResults(modifiedFuncs []ModifiedFunction, result *CallGraphResult) *CallGraphResult {
	for _, modFunc := range modifiedFuncs {
		result.ModifiedFunctions = append(result.ModifiedFunctions, FunctionCallGraph{
			Function:     modFunc.Name,
			File:         modFunc.File,
			Callers:      make([]CallInfo, 0),
			Callees:      make([]CallInfo, 0),
			TestCoverage: make([]TestCoverage, 0),
		})
	}
	return result
}

// isPythonTestFile checks if a file path indicates a test file.
func isPythonTestFile(filePath string) bool {
	base := filepath.Base(filePath)

	// Check for common test file patterns
	if strings.HasPrefix(base, "test_") {
		return true
	}
	if strings.HasSuffix(base, "_test.py") {
		return true
	}

	// Check if in tests directory
	normalizedPath := filepath.ToSlash(filePath)
	return strings.Contains(normalizedPath, "tests/") || strings.Contains(normalizedPath, "test/")
}

// isPythonTestFunction checks if a function name indicates a test function.
func isPythonTestFunction(funcName string) bool {
	// Remove class prefix if present
	parts := strings.Split(funcName, ".")
	baseName := parts[len(parts)-1]

	// Python unittest/pytest patterns
	return strings.HasPrefix(baseName, "test_") ||
		strings.HasPrefix(baseName, "Test") ||
		strings.HasSuffix(baseName, "_test")
}
