package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
)

// highImpactCallerThreshold is the minimum number of callers for a function
// to be considered high-impact. Used consistently across all analysis.
const highImpactCallerThreshold = 3

// maxJSONFileSize is the maximum allowed size for JSON input files (50MB).
const maxJSONFileSize = 50 * 1024 * 1024

// reviewerDataBuilder is a function that populates template data for a specific reviewer.
type reviewerDataBuilder func(c *Compiler, data *TemplateData, outputs *PhaseOutputs)

// reviewerDataBuilders maps reviewer names to their data builder functions.
var reviewerDataBuilders = map[string]reviewerDataBuilder{
	"code-reviewer":           (*Compiler).buildCodeReviewerData,
	"security-reviewer":       (*Compiler).buildSecurityReviewerData,
	"business-logic-reviewer": (*Compiler).buildBusinessLogicReviewerData,
	"test-reviewer":           (*Compiler).buildTestReviewerData,
	"nil-safety-reviewer":     (*Compiler).buildNilSafetyReviewerData,
}

// Compiler aggregates phase outputs and generates reviewer context files.
type Compiler struct {
	inputDir  string
	outputDir string
	language  string
}

// validatePath validates a directory path for security.
// It prevents path traversal attacks and optionally verifies the directory exists.
func validatePath(path string, mustExist bool) error {
	validated, err := fileutil.ValidatePath(path, ".")
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if mustExist {
		info, err := os.Stat(validated)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", validated)
			}
			return fmt.Errorf("failed to stat path: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", validated)
		}
	}

	return nil
}

// readJSONFileWithLimit reads a JSON file with a size limit to prevent resource exhaustion.
func readJSONFileWithLimit(path string) ([]byte, error) {
	return fileutil.ReadJSONFileWithLimit(path)
}

// NewCompiler creates a new context compiler.
// inputDir: directory containing phase outputs (e.g., .ring/codereview/)
// outputDir: directory to write context files (typically same as inputDir)
// Returns an error if paths contain traversal sequences.
func NewCompiler(inputDir, outputDir string) *Compiler {
	return &Compiler{
		inputDir:  inputDir,
		outputDir: outputDir,
	}
}

// NewCompilerWithValidation creates a new context compiler with path validation.
// inputDir: directory containing phase outputs (e.g., .ring/codereview/)
// outputDir: directory to write context files (typically same as inputDir)
// Returns an error if paths contain traversal sequences or are invalid.
func NewCompilerWithValidation(inputDir, outputDir string) (*Compiler, error) {
	// Validate input directory (must exist since we read from it)
	if err := validatePath(inputDir, true); err != nil {
		return nil, fmt.Errorf("invalid input directory: %w", err)
	}

	// Validate output directory (may not exist yet, will be created)
	if err := validatePath(outputDir, false); err != nil {
		return nil, fmt.Errorf("invalid output directory: %w", err)
	}

	return &Compiler{
		inputDir:  inputDir,
		outputDir: outputDir,
	}, nil
}

// Compile reads all phase outputs and generates reviewer context files.
func (c *Compiler) Compile() error {
	// Read all phase outputs
	outputs, err := c.readPhaseOutputs()
	if err != nil {
		return fmt.Errorf("failed to read phase outputs: %w", err)
	}

	// Determine language from scope
	if outputs.Scope != nil {
		c.language = outputs.Scope.Language
	}

	// Generate context for each reviewer
	reviewers := GetReviewerNames()
	for _, reviewer := range reviewers {
		if err := c.generateReviewerContext(reviewer, outputs); err != nil {
			return fmt.Errorf("failed to generate context for %s: %w", reviewer, err)
		}
	}

	return nil
}

// readPhaseOutputs reads all phase outputs from the input directory.
func (c *Compiler) readPhaseOutputs() (*PhaseOutputs, error) {
	outputs := &PhaseOutputs{}

	// Read scope.json (Phase 0)
	scopePath := filepath.Join(c.inputDir, "scope.json")
	if data, err := readJSONFileWithLimit(scopePath); err == nil {
		var scope ScopeData
		if err := json.Unmarshal(data, &scope); err == nil {
			applyScopeDerivedFields(&scope)
			outputs.Scope = &scope
		} else {
			outputs.Errors = append(outputs.Errors, fmt.Sprintf("scope.json parse error: %v", err))
		}
	} else if !os.IsNotExist(err) {
		outputs.Errors = append(outputs.Errors, fmt.Sprintf("scope.json read error: %v", err))
	}

	// Read static-analysis.json (Phase 1)
	staticPath := filepath.Join(c.inputDir, "static-analysis.json")
	if data, err := readJSONFileWithLimit(staticPath); err == nil {
		var static StaticAnalysisData
		if err := json.Unmarshal(data, &static); err == nil {
			outputs.StaticAnalysis = &static
		} else {
			outputs.Errors = append(outputs.Errors, fmt.Sprintf("static-analysis.json parse error: %v", err))
		}
	} else if !os.IsNotExist(err) {
		outputs.Errors = append(outputs.Errors, fmt.Sprintf("static-analysis.json read error: %v", err))
	}

	// Read language-specific AST (Phase 2) - support multi-language projects
	outputs.ASTByLanguage = make(map[string]*ASTData)
	for _, lang := range []string{"go", "typescript", "python"} {
		astPath := filepath.Join(c.inputDir, fmt.Sprintf("%s-ast.json", lang))
		if data, err := readJSONFileWithLimit(astPath); err == nil {
			var ast ASTData
			if err := json.Unmarshal(data, &ast); err == nil {
				outputs.ASTByLanguage[lang] = &ast
				// Keep backward compatibility: first language found becomes primary
				if outputs.AST == nil {
					outputs.AST = &ast
				}
			} else {
				outputs.Errors = append(outputs.Errors, fmt.Sprintf("%s-ast.json parse error: %v", lang, err))
			}
		} else if !os.IsNotExist(err) {
			outputs.Errors = append(outputs.Errors, fmt.Sprintf("%s-ast.json read error: %v", lang, err))
		}
	}

	// Read language-specific call graph (Phase 3) - support multi-language projects
	outputs.CallGraphByLanguage = make(map[string]*CallGraphData)
	for _, lang := range []string{"go", "typescript", "python"} {
		callsPath := filepath.Join(c.inputDir, fmt.Sprintf("%s-calls.json", lang))
		if data, err := readJSONFileWithLimit(callsPath); err == nil {
			var calls CallGraphData
			if err := json.Unmarshal(data, &calls); err == nil {
				outputs.CallGraphByLanguage[lang] = &calls
				// Keep backward compatibility: first language found becomes primary
				if outputs.CallGraph == nil {
					outputs.CallGraph = &calls
				}
			} else {
				outputs.Errors = append(outputs.Errors, fmt.Sprintf("%s-calls.json parse error: %v", lang, err))
			}
		} else if !os.IsNotExist(err) {
			outputs.Errors = append(outputs.Errors, fmt.Sprintf("%s-calls.json read error: %v", lang, err))
		}
	}

	// Read language-specific data flow (Phase 4) - support multi-language projects
	outputs.DataFlowByLanguage = make(map[string]*DataFlowData)
	for _, lang := range []string{"go", "typescript", "python"} {
		flowPath := filepath.Join(c.inputDir, fmt.Sprintf("%s-flow.json", lang))
		if data, err := readJSONFileWithLimit(flowPath); err == nil {
			var flow DataFlowData
			if err := json.Unmarshal(data, &flow); err == nil {
				outputs.DataFlowByLanguage[lang] = &flow
				// Keep backward compatibility: first language found becomes primary
				if outputs.DataFlow == nil {
					outputs.DataFlow = &flow
				}
			} else {
				outputs.Errors = append(outputs.Errors, fmt.Sprintf("%s-flow.json parse error: %v", lang, err))
			}
		} else if !os.IsNotExist(err) {
			outputs.Errors = append(outputs.Errors, fmt.Sprintf("%s-flow.json read error: %v", lang, err))
		}
	}

	return outputs, nil
}

func applyScopeDerivedFields(scope *ScopeData) {
	if scope == nil {
		return
	}
	scope.ModifiedFiles = scope.Files.Modified
	scope.AddedFiles = scope.Files.Added
	scope.DeletedFiles = scope.Files.Deleted
	scope.TotalFiles = scope.Stats.TotalFiles
	scope.TotalAdditions = scope.Stats.TotalAdditions
	scope.TotalDeletions = scope.Stats.TotalDeletions
	if scope.Languages == nil {
		scope.Languages = []string{}
	}
}

// generateReviewerContext generates the context file for a specific reviewer.
func (c *Compiler) generateReviewerContext(reviewer string, outputs *PhaseOutputs) error {
	// Build template data based on reviewer
	data := c.buildTemplateData(reviewer, outputs)

	// Get and render template
	templateStr, err := GetTemplateForReviewer(reviewer)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	content, err := RenderTemplate(templateStr, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Write context file
	outputPath := filepath.Join(c.outputDir, fmt.Sprintf("context-%s.md", reviewer))
	if err := os.MkdirAll(c.outputDir, 0o700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

// buildTemplateData constructs the template data for a specific reviewer.
func (c *Compiler) buildTemplateData(reviewer string, outputs *PhaseOutputs) *TemplateData {
	data := &TemplateData{}
	builder, ok := reviewerDataBuilders[reviewer]
	if !ok {
		data.Findings = []Finding{}
		data.FocusAreas = []FocusArea{{
			Title:       "Unknown reviewer",
			Description: fmt.Sprintf("No template data builder registered for reviewer %s", reviewer),
		}}
		return data
	}
	builder(c, data, outputs)
	return data
}

// buildCodeReviewerData populates data for the code reviewer.
func (c *Compiler) buildCodeReviewerData(data *TemplateData, outputs *PhaseOutputs) {
	// Static analysis findings (non-security)
	if outputs.StaticAnalysis != nil {
		data.Findings = FilterFindingsForCodeReviewer(outputs.StaticAnalysis.Findings)
		data.FindingCount = len(data.Findings)
	}

	// Semantic changes from AST
	if outputs.AST != nil {
		data.HasSemanticChanges = true
		data.ModifiedFunctions = outputs.AST.Functions.Modified
		data.AddedFunctions = outputs.AST.Functions.Added
		data.ModifiedTypes = outputs.AST.Types.Modified
	}

	// Build focus areas
	data.FocusAreas = c.buildCodeReviewerFocusAreas(outputs)
}

// buildSecurityReviewerData populates data for the security reviewer.
func (c *Compiler) buildSecurityReviewerData(data *TemplateData, outputs *PhaseOutputs) {
	// Security-specific findings
	if outputs.StaticAnalysis != nil {
		data.Findings = FilterFindingsForSecurityReviewer(outputs.StaticAnalysis.Findings)
		data.FindingCount = len(data.Findings)
	}

	// Data flow analysis
	if outputs.DataFlow != nil {
		data.HasDataFlowAnalysis = true
		for _, flow := range outputs.DataFlow.Flows {
			switch flow.Risk {
			case "high", "critical":
				data.HighRiskFlows = append(data.HighRiskFlows, flow)
			case "medium":
				data.MediumRiskFlows = append(data.MediumRiskFlows, flow)
			}
		}
	}

	// Build focus areas
	data.FocusAreas = c.buildSecurityReviewerFocusAreas(outputs)
}

// buildBusinessLogicReviewerData populates data for the business logic reviewer.
func (c *Compiler) buildBusinessLogicReviewerData(data *TemplateData, outputs *PhaseOutputs) {
	// Call graph for impact analysis
	if outputs.CallGraph != nil {
		data.HasCallGraph = true
		data.HighImpactFunctions = GetHighImpactFunctions(outputs.CallGraph, highImpactCallerThreshold)
	}

	// Semantic changes
	if outputs.AST != nil {
		data.HasSemanticChanges = true
		data.ModifiedFunctions = outputs.AST.Functions.Modified
	}

	// Build focus areas
	data.FocusAreas = c.buildBusinessLogicReviewerFocusAreas(outputs)
}

// buildTestReviewerData populates data for the test reviewer.
func (c *Compiler) buildTestReviewerData(data *TemplateData, outputs *PhaseOutputs) {
	// Call graph for test coverage
	if outputs.CallGraph != nil {
		data.HasCallGraph = true
		// Use AllModifiedFunctionsGraph for template (holds FunctionCallGraph)
		data.AllModifiedFunctionsGraph = outputs.CallGraph.ModifiedFunctions
		data.UncoveredFunctions = GetUncoveredFunctions(outputs.CallGraph)
	}

	// Build focus areas
	data.FocusAreas = c.buildTestReviewerFocusAreas(outputs)
}

// buildNilSafetyReviewerData populates data for the nil safety reviewer.
func (c *Compiler) buildNilSafetyReviewerData(data *TemplateData, outputs *PhaseOutputs) {
	// Nil sources from data flow
	if outputs.DataFlow != nil && len(outputs.DataFlow.NilSources) > 0 {
		data.HasNilSources = true
		data.NilSources = outputs.DataFlow.NilSources
		data.HighRiskNilSources = FilterNilSourcesByRisk(outputs.DataFlow.NilSources, "high")
	}

	// Build focus areas
	data.FocusAreas = c.buildNilSafetyReviewerFocusAreas(outputs)
}

// Focus area builders

func (c *Compiler) buildCodeReviewerFocusAreas(outputs *PhaseOutputs) []FocusArea {
	var areas []FocusArea

	// Check for deprecation warnings
	if outputs.StaticAnalysis != nil {
		deprecations := FilterFindingsByCategory(outputs.StaticAnalysis.Findings, "deprecation")
		if len(deprecations) > 0 {
			areas = append(areas, FocusArea{
				Title:       "Deprecated API Usage",
				Description: fmt.Sprintf("%d deprecated API calls need updating", len(deprecations)),
			})
		}
	}

	// Check for signature changes
	if outputs.AST != nil {
		for _, f := range outputs.AST.Functions.Modified {
			if f.Before.Signature != f.After.Signature {
				areas = append(areas, FocusArea{
					Title:       fmt.Sprintf("Signature change in %s", f.Name),
					Description: "Function signature modified - verify caller compatibility",
				})
			}
		}
	}

	return areas
}

func (c *Compiler) buildSecurityReviewerFocusAreas(outputs *PhaseOutputs) []FocusArea {
	var areas []FocusArea

	// Check for high-risk data flows
	if outputs.DataFlow != nil {
		highRisk := 0
		for _, flow := range outputs.DataFlow.Flows {
			if (flow.Risk == "high" || flow.Risk == "critical") && !flow.Sanitized {
				highRisk++
			}
		}
		if highRisk > 0 {
			areas = append(areas, FocusArea{
				Title:       "Unsanitized High-Risk Flows",
				Description: fmt.Sprintf("%d data flows without sanitization", highRisk),
			})
		}
	}

	// Check for security findings
	if outputs.StaticAnalysis != nil {
		critical := FilterFindingsBySeverity(
			FilterFindingsForSecurityReviewer(outputs.StaticAnalysis.Findings),
			"high",
		)
		if len(critical) > 0 {
			areas = append(areas, FocusArea{
				Title:       "Critical Security Findings",
				Description: fmt.Sprintf("%d high/critical security issues detected", len(critical)),
			})
		}
	}

	return areas
}

func (c *Compiler) buildBusinessLogicReviewerFocusAreas(outputs *PhaseOutputs) []FocusArea {
	var areas []FocusArea

	// Check for high-impact changes
	if outputs.CallGraph != nil {
		highImpact := GetHighImpactFunctions(outputs.CallGraph, highImpactCallerThreshold)
		if len(highImpact) > 0 {
			areas = append(areas, FocusArea{
				Title:       "High-Impact Functions",
				Description: fmt.Sprintf("%d functions with %d+ callers modified", len(highImpact), highImpactCallerThreshold),
			})
		}
	}

	// Check for new functions
	if outputs.AST != nil && len(outputs.AST.Functions.Added) > 0 {
		areas = append(areas, FocusArea{
			Title:       "New Functions",
			Description: fmt.Sprintf("%d new functions added - verify business requirements", len(outputs.AST.Functions.Added)),
		})
	}

	return areas
}

func (c *Compiler) buildTestReviewerFocusAreas(outputs *PhaseOutputs) []FocusArea {
	var areas []FocusArea

	// Check for uncovered functions
	if outputs.CallGraph != nil {
		uncovered := GetUncoveredFunctions(outputs.CallGraph)
		if len(uncovered) > 0 {
			areas = append(areas, FocusArea{
				Title:       "Uncovered Code",
				Description: fmt.Sprintf("%d modified functions without test coverage", len(uncovered)),
			})
		}
	}

	// Check for new error paths
	if outputs.AST != nil && outputs.AST.ErrorHandling.NewErrorReturns != nil {
		if len(outputs.AST.ErrorHandling.NewErrorReturns) > 0 {
			areas = append(areas, FocusArea{
				Title:       "New Error Paths",
				Description: fmt.Sprintf("%d new error return paths need negative tests", len(outputs.AST.ErrorHandling.NewErrorReturns)),
			})
		}
	}

	return areas
}

func (c *Compiler) buildNilSafetyReviewerFocusAreas(outputs *PhaseOutputs) []FocusArea {
	var areas []FocusArea

	// Check for unchecked nil sources
	if outputs.DataFlow != nil {
		unchecked := FilterNilSourcesUnchecked(outputs.DataFlow.NilSources)
		if len(unchecked) > 0 {
			areas = append(areas, FocusArea{
				Title:       "Unchecked Nil Sources",
				Description: fmt.Sprintf("%d potential nil values without checks", len(unchecked)),
			})
		}

		// Check for high-risk nil sources
		highRisk := FilterNilSourcesByRisk(outputs.DataFlow.NilSources, "high")
		if len(highRisk) > 0 {
			areas = append(areas, FocusArea{
				Title:       "High-Risk Nil Sources",
				Description: fmt.Sprintf("%d high-risk nil sources require immediate attention", len(highRisk)),
			})
		}
	}

	return areas
}
