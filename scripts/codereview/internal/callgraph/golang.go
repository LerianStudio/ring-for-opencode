package callgraph

import (
	"container/list"
	"context"
	"fmt"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// Resource protection limits.
const (
	maxModifiedFunctions = 500   // Maximum number of modified functions to analyze
	maxTransitiveCallers = 10000 // Maximum transitive callers before stopping BFS
	defaultTimeBudgetSec = 120   // Default time budget when not specified
)

// GoAnalyzer implements call graph analysis for Go code.
type GoAnalyzer struct {
	workDir        string
	loadPackagesFn func(ctx context.Context, patterns []string) ([]*packages.Package, []string, error)
}

// NewGoAnalyzer creates a new Go call graph analyzer.
// workDir is the root directory for package loading.
func NewGoAnalyzer(workDir string) *GoAnalyzer {
	analyzer := &GoAnalyzer{
		workDir: workDir,
	}
	analyzer.loadPackagesFn = analyzer.loadPackages
	return analyzer
}

// loadPackages loads Go packages from the working directory.
// Returns packages, any warnings encountered, and error if all packages failed to load.
func (g *GoAnalyzer) loadPackages(ctx context.Context, patterns []string) ([]*packages.Package, []string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Context: ctx,
		Dir:     g.workDir,
		Tests:   true, // Include test packages for test coverage analysis
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Check for package errors (but don't fail completely)
	var warnings []string
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			warnings = append(warnings, fmt.Sprintf("package %s: %v", pkg.PkgPath, err))
		}
	}

	if len(warnings) > 0 && len(pkgs) == 0 {
		return nil, warnings, fmt.Errorf("all packages failed to load: %v", warnings)
	}

	return pkgs, warnings, nil
}

// buildSSA builds SSA form for the loaded packages.
func (g *GoAnalyzer) buildSSA(pkgs []*packages.Package) *ssa.Program {
	// ssautil.AllPackages also returns an []*ssa.Package (ssaPkgs), but we only need the *ssa.Program for cha.CallGraph.
	prog, _ := ssautil.AllPackages(pkgs, ssa.InstantiateGenerics)
	if prog == nil {
		return nil
	}
	prog.Build()
	return prog
}

// getAffectedPackages extracts unique package paths from modified functions.
func getAffectedPackages(funcs []ModifiedFunction) []string {
	seen := make(map[string]bool)
	var result []string

	for _, fn := range funcs {
		if fn.Package != "" && !seen[fn.Package] {
			seen[fn.Package] = true
			result = append(result, fn.Package)
		}
	}

	return result
}

// Analyze implements the Analyzer interface for Go code.
func (g *GoAnalyzer) Analyze(modifiedFuncs []ModifiedFunction, timeBudgetSec int) (*CallGraphResult, error) {
	result := &CallGraphResult{
		Language:          "go",
		ModifiedFunctions: make([]FunctionCallGraph, 0, len(modifiedFuncs)),
		Warnings:          []string{},
		ImpactAnalysis: ImpactAnalysis{
			AffectedPackages: getAffectedPackages(modifiedFuncs),
		},
	}

	if len(modifiedFuncs) == 0 {
		return result, nil
	}

	// Truncate modified functions if exceeding limit
	if len(modifiedFuncs) > maxModifiedFunctions {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Truncated modified functions from %d to %d", len(modifiedFuncs), maxModifiedFunctions))
		modifiedFuncs = modifiedFuncs[:maxModifiedFunctions]
		result.PartialResults = true
	}

	// Apply default time budget if not specified
	if timeBudgetSec <= 0 {
		timeBudgetSec = defaultTimeBudgetSec
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeBudgetSec)*time.Second)
	defer cancel()

	// Determine patterns to load - use "./..." to load all packages in workspace
	patterns := []string{"./..."}

	// Load packages
	loadPackages := g.loadPackagesFn
	if loadPackages == nil {
		loadPackages = g.loadPackages
	}
	pkgs, pkgWarnings, err := loadPackages(ctx, patterns)
	if err != nil {
		// Check if context was cancelled (time budget exceeded)
		if ctx.Err() != nil {
			result.TimeBudgetExceeded = true
			result.PartialResults = true
			result.Warnings = append(result.Warnings, "Package loading timed out")
			return result, nil
		}
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	// Merge package loading warnings
	result.Warnings = append(result.Warnings, pkgWarnings...)

	if len(pkgs) == 0 {
		result.Warnings = append(result.Warnings, "No packages found")
		return result, nil
	}

	// Build SSA
	prog := g.buildSSA(pkgs)
	if prog == nil {
		result.Warnings = append(result.Warnings, "SSA program build failed")
		result.PartialResults = true
		return result, nil
	}

	// Build call graph using CHA (Class Hierarchy Analysis)
	// CHA is fast but conservative - it over-approximates call targets
	cg := cha.CallGraph(prog)

	// Track all unique direct callers and affected tests
	allDirectCallers := make(map[string]bool)
	allAffectedTests := make(map[string]bool)

	// Analyze each modified function
	for _, modFunc := range modifiedFuncs {
		// Check for timeout
		select {
		case <-ctx.Done():
			result.TimeBudgetExceeded = true
			result.PartialResults = true
			result.Warnings = append(result.Warnings, "Analysis timed out during function processing")
			return result, nil
		default:
		}

		fcg := g.analyzeFunction(prog, cg, modFunc, allDirectCallers, allAffectedTests)
		result.ModifiedFunctions = append(result.ModifiedFunctions, fcg)
	}

	// Calculate direct callers count
	result.ImpactAnalysis.DirectCallers = len(allDirectCallers)

	// Calculate transitive callers (excluding direct callers to avoid double-counting)
	transitiveCallers := g.findTransitiveCallers(cg, prog, modifiedFuncs, 10)
	result.ImpactAnalysis.TransitiveCallers = len(transitiveCallers) - result.ImpactAnalysis.DirectCallers
	if result.ImpactAnalysis.TransitiveCallers < 0 {
		result.ImpactAnalysis.TransitiveCallers = 0
	}

	// Calculate affected tests count
	result.ImpactAnalysis.AffectedTests = len(allAffectedTests)

	return result, nil
}

// analyzeFunction analyzes a single function and returns its call graph.
func (g *GoAnalyzer) analyzeFunction(
	prog *ssa.Program,
	cg *callgraph.Graph,
	modFunc ModifiedFunction,
	allDirectCallers map[string]bool,
	allAffectedTests map[string]bool,
) FunctionCallGraph {
	fcg := FunctionCallGraph{
		Function:     modFunc.Name,
		File:         modFunc.File,
		Callers:      make([]CallInfo, 0),
		Callees:      make([]CallInfo, 0),
		TestCoverage: make([]TestCoverage, 0),
	}

	// Find the SSA function
	ssaFunc := g.findSSAFunction(prog, modFunc)
	if ssaFunc == nil {
		return fcg
	}

	// Get the call graph node for this function
	node := cg.Nodes[ssaFunc]
	if node == nil {
		return fcg
	}

	// Find callers (incoming edges)
	for _, edge := range node.In {
		if edge.Caller == nil || edge.Caller.Func == nil {
			continue
		}

		callerFunc := edge.Caller.Func
		callerName := callerFunc.Name()
		if callerFunc.Signature.Recv() != nil {
			// Method - include receiver type
			callerName = callerFunc.Signature.Recv().Type().String() + "." + callerName
		}

		// Get position of the call site
		var callSite string
		var line int
		var file string

		if edge.Site != nil && edge.Site.Pos().IsValid() {
			pos := prog.Fset.Position(edge.Site.Pos())
			file = pos.Filename
			line = pos.Line
			callSite = fmt.Sprintf("%s:%d", filepath.Base(pos.Filename), pos.Line)
		} else if callerFunc.Pos().IsValid() {
			pos := prog.Fset.Position(callerFunc.Pos())
			file = pos.Filename
			line = pos.Line
		}

		// Track unique callers
		callerKey := fmt.Sprintf("%s:%s", file, callerName)
		allDirectCallers[callerKey] = true

		// Check if caller is a test function
		if isTestFunction(callerName) {
			allAffectedTests[callerKey] = true
			fcg.TestCoverage = append(fcg.TestCoverage, TestCoverage{
				TestFunction: callerName,
				File:         file,
				Line:         line,
			})
		}

		fcg.Callers = append(fcg.Callers, CallInfo{
			Function: callerName,
			File:     file,
			Line:     line,
			CallSite: callSite,
		})
	}

	// Find callees (outgoing edges)
	for _, edge := range node.Out {
		if edge.Callee == nil || edge.Callee.Func == nil {
			continue
		}

		calleeFunc := edge.Callee.Func
		calleeName := calleeFunc.Name()
		if calleeFunc.Signature.Recv() != nil {
			// Method - include receiver type
			calleeName = calleeFunc.Signature.Recv().Type().String() + "." + calleeName
		}

		// Get position
		var file string
		var line int
		var callSite string

		if edge.Site != nil && edge.Site.Pos().IsValid() {
			pos := prog.Fset.Position(edge.Site.Pos())
			callSite = fmt.Sprintf("%s:%d", filepath.Base(pos.Filename), pos.Line)
		}

		if calleeFunc.Pos().IsValid() {
			pos := prog.Fset.Position(calleeFunc.Pos())
			file = pos.Filename
			line = pos.Line
		}

		fcg.Callees = append(fcg.Callees, CallInfo{
			Function: calleeName,
			File:     file,
			Line:     line,
			CallSite: callSite,
		})
	}

	return fcg
}

// findSSAFunction locates the SSA function corresponding to a ModifiedFunction.
func (g *GoAnalyzer) findSSAFunction(prog *ssa.Program, modFunc ModifiedFunction) *ssa.Function {
	// Normalize the file path for comparison
	targetFile := filepath.Clean(modFunc.File)
	if !filepath.IsAbs(targetFile) && g.workDir != "" {
		targetFile = filepath.Join(g.workDir, targetFile)
	}

	// Construct the expected function name
	expectedName := modFunc.Name
	if modFunc.Receiver != "" {
		// Strip pointer indicator for matching
		receiver := strings.TrimPrefix(modFunc.Receiver, "*")
		expectedName = receiver + "." + modFunc.Name
	}

	// Search all packages for the function
	for _, pkg := range prog.AllPackages() {
		if pkg == nil {
			continue
		}

		// Check package-level functions
		for _, member := range pkg.Members {
			if fn, ok := member.(*ssa.Function); ok {
				if g.functionMatches(prog.Fset, fn, expectedName, targetFile, modFunc.Receiver) {
					return fn
				}
			}

			// Check methods on types
			if t, ok := member.(*ssa.Type); ok {
				// Get methods of the named type
				for i := 0; i < prog.MethodSets.MethodSet(t.Type()).Len(); i++ {
					sel := prog.MethodSets.MethodSet(t.Type()).At(i)
					fn := prog.MethodValue(sel)
					if fn != nil && g.functionMatches(prog.Fset, fn, expectedName, targetFile, modFunc.Receiver) {
						return fn
					}
				}

				// Also check pointer methods
				ptrType := prog.MethodSets.MethodSet(types.NewPointer(t.Type()))
				for i := 0; i < ptrType.Len(); i++ {
					sel := ptrType.At(i)
					fn := prog.MethodValue(sel)
					if fn != nil && g.functionMatches(prog.Fset, fn, expectedName, targetFile, modFunc.Receiver) {
						return fn
					}
				}
			}
		}
	}

	return nil
}

// functionMatches checks if an SSA function matches the expected name and file.
func (g *GoAnalyzer) functionMatches(fset *token.FileSet, fn *ssa.Function, expectedName, targetFile, receiver string) bool {
	// Build the full name for comparison
	fnName := fn.Name()
	if fn.Signature.Recv() != nil {
		recvType := fn.Signature.Recv().Type().String()
		// Extract just the type name without package path
		if idx := strings.LastIndex(recvType, "."); idx >= 0 {
			recvType = recvType[idx+1:]
		}
		recvType = strings.TrimPrefix(recvType, "*")
		fnName = recvType + "." + fn.Name()
	}

	// Check name match
	if fnName != expectedName && fn.Name() != expectedName {
		// Also try without receiver for simple function matching
		return false
	}

	// Check file match if we have position info
	if fn.Pos().IsValid() {
		fnFile := filepath.Clean(fset.Position(fn.Pos()).Filename)
		if fnFile != targetFile && !strings.HasSuffix(fnFile, targetFile) && !strings.HasSuffix(targetFile, filepath.Base(fnFile)) {
			return false
		}
	}

	return true
}

// isTestFunction checks if a function name indicates a test function.
func isTestFunction(name string) bool {
	// Remove receiver prefix if present
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}

	return strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") ||
		strings.HasPrefix(name, "Example") ||
		strings.HasPrefix(name, "Fuzz")
}

// findTransitiveCallers performs BFS to find all transitive callers up to maxDepth.
func (g *GoAnalyzer) findTransitiveCallers(
	cg *callgraph.Graph,
	prog *ssa.Program,
	modifiedFuncs []ModifiedFunction,
	maxDepth int,
) map[string]bool {
	transitiveCallers := make(map[string]bool)

	// Start with all modified functions
	var startNodes []*callgraph.Node
	for _, modFunc := range modifiedFuncs {
		ssaFunc := g.findSSAFunction(prog, modFunc)
		if ssaFunc == nil {
			continue
		}
		if node := cg.Nodes[ssaFunc]; node != nil {
			startNodes = append(startNodes, node)
		}
	}

	if len(startNodes) == 0 {
		return transitiveCallers
	}

	// BFS to find all transitive callers
	visited := make(map[*callgraph.Node]bool)
	queue := list.New()

	// Initialize queue with start nodes at depth 0
	for _, node := range startNodes {
		queue.PushBack(&nodeWithDepth{node: node, depth: 0})
		visited[node] = true
	}

	for queue.Len() > 0 {
		// Stop early if we've exceeded the transitive callers limit
		if len(transitiveCallers) >= maxTransitiveCallers {
			break
		}

		// Dequeue
		elem := queue.Front()
		current := elem.Value.(*nodeWithDepth)
		queue.Remove(elem)

		// Don't go beyond max depth
		if current.depth >= maxDepth {
			continue
		}

		// Process all callers (incoming edges)
		for _, edge := range current.node.In {
			if edge.Caller == nil || edge.Caller.Func == nil {
				continue
			}

			callerNode := edge.Caller
			if visited[callerNode] {
				continue
			}

			visited[callerNode] = true

			// Record this transitive caller
			callerFunc := callerNode.Func
			callerKey := callerFunc.String()
			if callerFunc.Pos().IsValid() {
				pos := prog.Fset.Position(callerFunc.Pos())
				callerKey = fmt.Sprintf("%s:%s", pos.Filename, callerFunc.Name())
			}
			transitiveCallers[callerKey] = true

			// Enqueue for further traversal
			queue.PushBack(&nodeWithDepth{
				node:  callerNode,
				depth: current.depth + 1,
			})
		}
	}

	return transitiveCallers
}

// nodeWithDepth is a helper struct for BFS traversal.
type nodeWithDepth struct {
	node  *callgraph.Node
	depth int
}
