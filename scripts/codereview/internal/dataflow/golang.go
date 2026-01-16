// Package dataflow provides data flow analysis for security review.
package dataflow

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MaxFileSize is the maximum file size to analyze (10MB - matches Python analyzer).
const MaxFileSize = 10 * 1024 * 1024

// sourcePattern defines a pattern for detecting untrusted data sources.
type sourcePattern struct {
	Type    SourceType
	Pattern *regexp.Regexp
	Desc    string
}

// sinkPattern defines a pattern for detecting sensitive data sinks.
type sinkPattern struct {
	Type    SinkType
	Pattern *regexp.Regexp
	Desc    string
}

// GoAnalyzer implements data flow analysis for Go code.
type GoAnalyzer struct {
	workDir         string
	sourcePatterns  []sourcePattern
	sinkPatterns    []sinkPattern
	sanitizerRegex  *regexp.Regexp
	nilPatterns     []*regexp.Regexp
	variableTracker map[string]Source
}

// NewGoAnalyzer creates a new Go data flow analyzer.
func NewGoAnalyzer(workDir string) *GoAnalyzer {
	return &GoAnalyzer{
		workDir:         workDir,
		sourcePatterns:  initSourcePatterns(),
		sinkPatterns:    initSinkPatterns(),
		sanitizerRegex:  initSanitizerRegex(),
		nilPatterns:     initNilPatterns(),
		variableTracker: make(map[string]Source),
	}
}

// initSourcePatterns returns patterns for detecting untrusted data sources.
func initSourcePatterns() []sourcePattern {
	return []sourcePattern{
		// HTTP Body
		{
			Type:    SourceHTTPBody,
			Pattern: regexp.MustCompile(`(?:r|req|request)\.Body`),
			Desc:    "HTTP request body",
		},
		{
			Type:    SourceHTTPBody,
			Pattern: regexp.MustCompile(`json\.(?:NewDecoder|Unmarshal)\s*\(`),
			Desc:    "JSON decode from request",
		},
		{
			Type:    SourceHTTPBody,
			Pattern: regexp.MustCompile(`ioutil\.ReadAll\s*\(\s*(?:r|req|request)\.Body`),
			Desc:    "Read all from request body",
		},
		{
			Type:    SourceHTTPBody,
			Pattern: regexp.MustCompile(`io\.ReadAll\s*\(\s*(?:r|req|request)\.Body`),
			Desc:    "Read all from request body",
		},
		{
			Type:    SourceHTTPBody,
			Pattern: regexp.MustCompile(`c\.(?:Body|BodyParser|Bind)\s*\(`),
			Desc:    "Fiber body binding",
		},

		// HTTP Query Parameters
		{
			Type:    SourceHTTPQuery,
			Pattern: regexp.MustCompile(`(?:r|req|request)\.URL\.Query\s*\(\)`),
			Desc:    "URL query parameters",
		},
		{
			Type:    SourceHTTPQuery,
			Pattern: regexp.MustCompile(`(?:r|req|request)\.FormValue\s*\(`),
			Desc:    "Form value from request",
		},
		{
			Type:    SourceHTTPQuery,
			Pattern: regexp.MustCompile(`(?:r|req|request)\.Form\.Get\s*\(`),
			Desc:    "Form.Get from request",
		},
		{
			Type:    SourceHTTPQuery,
			Pattern: regexp.MustCompile(`c\.Query\s*\(`),
			Desc:    "Fiber query parameter",
		},
		{
			Type:    SourceHTTPQuery,
			Pattern: regexp.MustCompile(`c\.QueryParam\s*\(`),
			Desc:    "Echo query parameter",
		},

		// HTTP Headers
		{
			Type:    SourceHTTPHeader,
			Pattern: regexp.MustCompile(`(?:r|req|request)\.Header\.Get\s*\(`),
			Desc:    "HTTP header value",
		},
		{
			Type:    SourceHTTPHeader,
			Pattern: regexp.MustCompile(`(?:r|req|request)\.Header\[`),
			Desc:    "HTTP header access",
		},
		{
			Type:    SourceHTTPHeader,
			Pattern: regexp.MustCompile(`c\.Get\s*\(\s*["']`),
			Desc:    "Fiber header get",
		},

		// HTTP Path Parameters
		{
			Type:    SourceHTTPPath,
			Pattern: regexp.MustCompile(`(?:mux\.)?Vars\s*\(\s*(?:r|req|request)\s*\)`),
			Desc:    "URL path variable (gorilla/mux)",
		},
		{
			Type:    SourceHTTPPath,
			Pattern: regexp.MustCompile(`chi\.URLParam\s*\(`),
			Desc:    "URL parameter (chi)",
		},
		{
			Type:    SourceHTTPPath,
			Pattern: regexp.MustCompile(`c\.Params\s*\(`),
			Desc:    "Fiber path parameter",
		},
		{
			Type:    SourceHTTPPath,
			Pattern: regexp.MustCompile(`c\.Param\s*\(`),
			Desc:    "Echo/Gin path parameter",
		},

		// Environment Variables
		{
			Type:    SourceEnvVar,
			Pattern: regexp.MustCompile(`os\.Getenv\s*\(`),
			Desc:    "Environment variable read",
		},
		{
			Type:    SourceEnvVar,
			Pattern: regexp.MustCompile(`os\.LookupEnv\s*\(`),
			Desc:    "Environment variable lookup",
		},
		{
			Type:    SourceEnvVar,
			Pattern: regexp.MustCompile(`viper\.(?:Get|GetString|GetInt)\s*\(`),
			Desc:    "Viper config read",
		},

		// File System
		{
			Type:    SourceFile,
			Pattern: regexp.MustCompile(`os\.(?:Open|ReadFile)\s*\(`),
			Desc:    "File open/read",
		},
		{
			Type:    SourceFile,
			Pattern: regexp.MustCompile(`ioutil\.ReadFile\s*\(`),
			Desc:    "File read (ioutil)",
		},
		{
			Type:    SourceFile,
			Pattern: regexp.MustCompile(`io\.(?:ReadAll|Copy)\s*\(`),
			Desc:    "IO read operation",
		},
		{
			Type:    SourceFile,
			Pattern: regexp.MustCompile(`bufio\.NewReader\s*\(`),
			Desc:    "Buffered reader",
		},

		// Database
		{
			Type:    SourceDatabase,
			Pattern: regexp.MustCompile(`\.Query(?:Row|Context)?\s*\(`),
			Desc:    "Database query result",
		},
		{
			Type:    SourceDatabase,
			Pattern: regexp.MustCompile(`\.Scan\s*\(`),
			Desc:    "Database scan result",
		},
		{
			Type:    SourceDatabase,
			Pattern: regexp.MustCompile(`\.Find(?:One|All|By)?\s*\(`),
			Desc:    "ORM find operation",
		},
		{
			Type:    SourceDatabase,
			Pattern: regexp.MustCompile(`\.First\s*\(`),
			Desc:    "GORM first query",
		},
		{
			Type:    SourceDatabase,
			Pattern: regexp.MustCompile(`collection\.(?:Find|FindOne)\s*\(`),
			Desc:    "MongoDB query",
		},

		// External API
		{
			Type:    SourceExternal,
			Pattern: regexp.MustCompile(`http\.(?:Get|Post|Do)\s*\(`),
			Desc:    "HTTP client request",
		},
		{
			Type:    SourceExternal,
			Pattern: regexp.MustCompile(`client\.(?:Get|Post|Do)\s*\(`),
			Desc:    "HTTP client call",
		},
		{
			Type:    SourceExternal,
			Pattern: regexp.MustCompile(`\.(?:Response|Body)\.`),
			Desc:    "HTTP response body",
		},

		// User Input
		{
			Type:    SourceUserInput,
			Pattern: regexp.MustCompile(`bufio\.NewScanner\s*\(\s*os\.Stdin`),
			Desc:    "Stdin scanner",
		},
		{
			Type:    SourceUserInput,
			Pattern: regexp.MustCompile(`fmt\.Scan(?:f|ln)?\s*\(`),
			Desc:    "Console input",
		},
	}
}

// initSinkPatterns returns patterns for detecting sensitive data sinks.
func initSinkPatterns() []sinkPattern {
	return []sinkPattern{
		// Database Operations
		{
			Type:    SinkDatabase,
			Pattern: regexp.MustCompile(`\.Exec(?:Context)?\s*\(`),
			Desc:    "SQL exec (potential injection)",
		},
		{
			Type:    SinkDatabase,
			Pattern: regexp.MustCompile(`\.Query(?:Row)?(?:Context)?\s*\([^,]*\+`),
			Desc:    "SQL query with concatenation",
		},
		{
			Type:    SinkDatabase,
			Pattern: regexp.MustCompile(`fmt\.Sprintf\s*\([^)]*(?:SELECT|INSERT|UPDATE|DELETE)`),
			Desc:    "SQL string formatting",
		},
		{
			Type:    SinkDatabase,
			Pattern: regexp.MustCompile(`collection\.(?:InsertOne|UpdateOne|DeleteOne)\s*\(`),
			Desc:    "MongoDB write operation",
		},

		// Command Execution
		{
			Type:    SinkExec,
			Pattern: regexp.MustCompile(`exec\.Command\s*\(`),
			Desc:    "OS command execution",
		},
		{
			Type:    SinkExec,
			Pattern: regexp.MustCompile(`exec\.CommandContext\s*\(`),
			Desc:    "OS command with context",
		},
		{
			Type:    SinkExec,
			Pattern: regexp.MustCompile(`os\.StartProcess\s*\(`),
			Desc:    "OS process start",
		},
		{
			Type:    SinkExec,
			Pattern: regexp.MustCompile(`syscall\.Exec\s*\(`),
			Desc:    "Syscall exec",
		},

		// HTTP Response
		{
			Type:    SinkResponse,
			Pattern: regexp.MustCompile(`(?:w|rw|writer|response)\.Write\s*\(`),
			Desc:    "HTTP response write",
		},
		{
			Type:    SinkResponse,
			Pattern: regexp.MustCompile(`(?:w|rw|writer|response)\.WriteString\s*\(`),
			Desc:    "HTTP response write string",
		},
		{
			Type:    SinkResponse,
			Pattern: regexp.MustCompile(`fmt\.Fprint(?:f|ln)?\s*\(\s*(?:w|rw|writer)`),
			Desc:    "Printf to response writer",
		},
		{
			Type:    SinkResponse,
			Pattern: regexp.MustCompile(`json\.NewEncoder\s*\(\s*(?:w|rw|writer)`),
			Desc:    "JSON encode to response",
		},
		{
			Type:    SinkResponse,
			Pattern: regexp.MustCompile(`c\.(?:JSON|Send|Write|Status)\s*\(`),
			Desc:    "Fiber response",
		},
		{
			Type:    SinkResponse,
			Pattern: regexp.MustCompile(`c\.(?:String|HTML|XML)\s*\(`),
			Desc:    "Echo/Gin response",
		},

		// Logging
		{
			Type:    SinkLog,
			Pattern: regexp.MustCompile(`log\.(?:Print|Printf|Println|Fatal|Fatalf)\s*\(`),
			Desc:    "Standard log output",
		},
		{
			Type:    SinkLog,
			Pattern: regexp.MustCompile(`logger\.(?:Info|Warn|Error|Debug|Fatal)(?:f|w)?\s*\(`),
			Desc:    "Structured logger",
		},
		{
			Type:    SinkLog,
			Pattern: regexp.MustCompile(`zap\.(?:L|S)\(\)\.(?:Info|Warn|Error|Debug)\s*\(`),
			Desc:    "Zap logger",
		},
		{
			Type:    SinkLog,
			Pattern: regexp.MustCompile(`logrus\.(?:Info|Warn|Error|Debug|Fatal)(?:f)?\s*\(`),
			Desc:    "Logrus logger",
		},
		{
			Type:    SinkLog,
			Pattern: regexp.MustCompile(`slog\.(?:Info|Warn|Error|Debug)\s*\(`),
			Desc:    "Slog logger",
		},

		// File Operations
		{
			Type:    SinkFile,
			Pattern: regexp.MustCompile(`os\.(?:WriteFile|Create)\s*\(`),
			Desc:    "File write/create",
		},
		{
			Type:    SinkFile,
			Pattern: regexp.MustCompile(`ioutil\.WriteFile\s*\(`),
			Desc:    "File write (ioutil)",
		},
		{
			Type:    SinkFile,
			Pattern: regexp.MustCompile(`(?:f|file)\.Write(?:String)?\s*\(`),
			Desc:    "File handle write",
		},
		{
			Type:    SinkFile,
			Pattern: regexp.MustCompile(`io\.(?:Copy|WriteString)\s*\(`),
			Desc:    "IO write operation",
		},

		// Template Rendering
		{
			Type:    SinkTemplate,
			Pattern: regexp.MustCompile(`template\.(?:HTML|JS|CSS)\s*\(`),
			Desc:    "Template type conversion",
		},
		{
			Type:    SinkTemplate,
			Pattern: regexp.MustCompile(`\.Execute(?:Template)?\s*\(`),
			Desc:    "Template execution",
		},
		{
			Type:    SinkTemplate,
			Pattern: regexp.MustCompile(`html/template.*Execute`),
			Desc:    "HTML template execute",
		},

		// Redirects
		{
			Type:    SinkRedirect,
			Pattern: regexp.MustCompile(`http\.Redirect\s*\(`),
			Desc:    "HTTP redirect",
		},
		{
			Type:    SinkRedirect,
			Pattern: regexp.MustCompile(`c\.Redirect\s*\(`),
			Desc:    "Fiber/Echo redirect",
		},
		{
			Type:    SinkRedirect,
			Pattern: regexp.MustCompile(`(?:w|rw)\.Header\(\)\.Set\s*\(\s*"Location"`),
			Desc:    "Location header redirect",
		},
	}
}

// initSanitizerRegex returns regex for detecting sanitization functions.
func initSanitizerRegex() *regexp.Regexp {
	sanitizers := []string{
		`html\.EscapeString`,
		`url\.QueryEscape`,
		`url\.PathEscape`,
		`strconv\.(?:Atoi|ParseInt|ParseFloat|ParseBool)`,
		`regexp\.MustCompile.*MatchString`,
		`strings\.(?:TrimSpace|ReplaceAll)`,
		`template\.(?:HTMLEscapeString|JSEscapeString)`,
		`sanitize\w*`,
		`escape\w*`,
		`validate\w*`,
		`clean\w*`,
		`filter\w*`,
		`whitelist\w*`,
		`sql\.Named`,
		`pgx\.NamedArgs`,
	}
	return regexp.MustCompile(`(?i)(?:` + strings.Join(sanitizers, "|") + `)`)
}

// initNilPatterns returns patterns for detecting nil sources.
func initNilPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		// Map lookups
		regexp.MustCompile(`(\w+)\s*(?:,\s*(?:ok|_))?\s*:?=\s*\w+\[`),
		// Database query results
		regexp.MustCompile(`(\w+)\s*,\s*err\s*:?=\s*(?:db|tx|conn)\.(?:Query|QueryRow|Exec)`),
		// Type assertions
		regexp.MustCompile(`(\w+)\s*(?:,\s*(?:ok|_))?\s*:?=\s*\w+\.\(\w+\)`),
		// JSON unmarshal
		regexp.MustCompile(`json\.Unmarshal\s*\([^,]+,\s*&?(\w+)\)`),
		// Pointer dereference setup
		regexp.MustCompile(`(\w+)\s*:?=\s*\*?\w+\.(?:Find|Get|Load|Lookup)\w*\s*\(`),
		// Interface type assertion
		regexp.MustCompile(`(\w+)\s*,\s*ok\s*:?=\s*\w+\.\(interface\{\}\)`),
		// Context value extraction
		regexp.MustCompile(`(\w+)\s*(?:,\s*(?:ok|_))?\s*:?=\s*(?:ctx|context)\.Value\s*\(`),
		// Channel receive
		regexp.MustCompile(`(\w+)\s*(?:,\s*(?:ok|_))?\s*:?=\s*<-\s*\w+`),
	}
}

// Language returns the analyzer's target language.
func (g *GoAnalyzer) Language() string {
	return "go"
}

// DetectSources scans files for untrusted data sources.
func (g *GoAnalyzer) DetectSources(files []string) ([]Source, error) {
	var sources []Source

	for _, filePath := range files {
		if !isGoFile(filePath) {
			continue
		}

		fileSources, err := g.detectSourcesInFile(filePath)
		if err != nil {
			continue // Skip files with errors
		}
		sources = append(sources, fileSources...)
	}

	return sources, nil
}

// detectSourcesInFile detects sources in a single file.
func (g *GoAnalyzer) detectSourcesInFile(filePath string) ([]Source, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), MaxFileSize)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sources []Source
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		for _, sp := range g.sourcePatterns {
			if matches := sp.Pattern.FindStringSubmatch(line); matches != nil {
				col := strings.Index(line, matches[0]) + 1
				variable := extractVariable(line)

				source := Source{
					Type:     sp.Type,
					File:     filePath,
					Line:     lineNum,
					Column:   col,
					Variable: variable,
					Pattern:  sp.Desc,
					Context:  strings.TrimSpace(line),
				}

				sources = append(sources, source)

				// Track variable for flow analysis
				if variable != "" {
					g.variableTracker[variable] = source
				}
			}
		}
	}

	return sources, scanner.Err()
}

// DetectSinks scans files for sensitive data sinks.
func (g *GoAnalyzer) DetectSinks(files []string) ([]Sink, error) {
	var sinks []Sink

	for _, filePath := range files {
		if !isGoFile(filePath) {
			continue
		}

		fileSinks, err := g.detectSinksInFile(filePath)
		if err != nil {
			continue // Skip files with errors
		}
		sinks = append(sinks, fileSinks...)
	}

	return sinks, nil
}

// detectSinksInFile detects sinks in a single file.
func (g *GoAnalyzer) detectSinksInFile(filePath string) ([]Sink, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), MaxFileSize)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sinks []Sink
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		for _, sp := range g.sinkPatterns {
			if matches := sp.Pattern.FindStringSubmatch(line); matches != nil {
				col := strings.Index(line, matches[0]) + 1
				funcName := extractFunctionName(line)

				sink := Sink{
					Type:     sp.Type,
					File:     filePath,
					Line:     lineNum,
					Column:   col,
					Function: funcName,
					Pattern:  sp.Desc,
					Context:  strings.TrimSpace(line),
				}

				sinks = append(sinks, sink)
			}
		}
	}

	return sinks, scanner.Err()
}

// TrackFlows traces data paths from sources to sinks.
func (g *GoAnalyzer) TrackFlows(sources []Source, sinks []Sink, files []string) ([]Flow, error) {
	var flows []Flow

	// Build file content map for flow analysis
	fileContents := make(map[string][]string)
	for _, filePath := range files {
		if !isGoFile(filePath) {
			continue
		}
		content, err := readFileLines(filePath)
		if err != nil {
			continue
		}
		fileContents[filePath] = content
	}

	// Analyze each source-sink pair within the same file
	for _, source := range sources {
		for _, sink := range sinks {
			// For simplicity, focus on same-file flows
			if source.File != sink.File {
				continue
			}

			// Check if source flows to sink
			if flow := g.analyzeFlow(source, sink, fileContents); flow != nil {
				flows = append(flows, *flow)
			}
		}
	}

	return flows, nil
}

// analyzeFlow determines if data flows from source to sink.
func (g *GoAnalyzer) analyzeFlow(source Source, sink Sink, fileContents map[string][]string) *Flow {
	lines, ok := fileContents[source.File]
	if !ok {
		return nil
	}

	// Check if source variable is used at sink
	sourceVar := source.Variable
	if sourceVar == "" {
		return nil
	}

	// Simple heuristic: source must come before sink
	if source.Line >= sink.Line {
		return nil
	}

	// Check if source variable appears in sink line
	if sink.Line > 0 && sink.Line <= len(lines) {
		sinkLine := lines[sink.Line-1]
		if !strings.Contains(sinkLine, sourceVar) {
			// Check for common transformations
			if !containsVariableOrDerivative(sinkLine, sourceVar) {
				return nil
			}
		}
	}

	// Build flow path
	path := buildFlowPath(source, sink, lines)

	// Check for sanitization between source and sink
	sanitized, sanitizers := g.checkSanitization(source.Line, sink.Line, lines)

	// Calculate risk level
	risk := calculateRisk(source.Type, sink.Type, sanitized)

	flow := &Flow{
		ID:          generateFlowID(source, sink),
		Source:      source,
		Sink:        sink,
		Path:        path,
		Sanitized:   sanitized,
		Sanitizers:  sanitizers,
		Risk:        risk,
		Description: describeFlow(source, sink, sanitized),
	}

	return flow
}

// DetectNilSources identifies variables that may be nil.
func (g *GoAnalyzer) DetectNilSources(files []string) ([]NilSource, error) {
	var nilSources []NilSource

	for _, filePath := range files {
		if !isGoFile(filePath) {
			continue
		}

		fileNils, err := g.detectNilSourcesInFile(filePath)
		if err != nil {
			continue // Skip files with errors
		}
		nilSources = append(nilSources, fileNils...)
	}

	return nilSources, nil
}

// detectNilSourcesInFile detects nil sources in a single file.
func (g *GoAnalyzer) detectNilSourcesInFile(filePath string) ([]NilSource, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), MaxFileSize)
	}

	content, err := readFileLines(filePath)
	if err != nil {
		return nil, err
	}

	var nilSources []NilSource
	detectedVars := make(map[string]NilSource)

	// First pass: detect variables that may be nil
	for lineNum, line := range content {
		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		for _, pattern := range g.nilPatterns {
			if matches := pattern.FindStringSubmatch(line); len(matches) > 1 {
				varName := matches[1]
				if varName == "" || varName == "_" {
					continue
				}

				origin := determineNilOrigin(line)
				nilSrc := NilSource{
					File:     filePath,
					Line:     lineNum + 1,
					Variable: varName,
					Origin:   origin,
					Risk:     RiskMedium,
				}

				// Check for ok pattern (map lookup, type assertion)
				if strings.Contains(line, ", ok") || strings.Contains(line, ", _") {
					nilSrc.IsChecked = true
				}

				detectedVars[varName] = nilSrc
			}
		}
	}

	// Second pass: check for nil checks and usage
	for lineNum, line := range content {
		lineNumber := lineNum + 1

		for varName, nilSrc := range detectedVars {
			// Skip if already marked as checked
			if nilSrc.IsChecked {
				continue
			}

			// Check for nil check
			nilCheckPattern := regexp.MustCompile(fmt.Sprintf(`if\s+%s\s*[!=]=\s*nil|if\s+nil\s*[!=]=\s*%s|%s\s*!=\s*nil\s*{`, varName, varName, varName))
			if nilCheckPattern.MatchString(line) && lineNumber > nilSrc.Line {
				nilSrc.IsChecked = true
				nilSrc.CheckLine = lineNumber
				detectedVars[varName] = nilSrc
				continue
			}

			// Check for usage without nil check
			usagePattern := regexp.MustCompile(fmt.Sprintf(`%s\.|\*%s|%s\[`, varName, varName, varName))
			if usagePattern.MatchString(line) && lineNumber > nilSrc.Line && !nilSrc.IsChecked {
				nilSrc.UsageLine = lineNumber
				nilSrc.Risk = RiskHigh
				detectedVars[varName] = nilSrc
			}
		}
	}

	// Collect results
	for _, nilSrc := range detectedVars {
		// Only report unchecked nil sources as risks
		if !nilSrc.IsChecked && nilSrc.UsageLine > 0 {
			nilSources = append(nilSources, nilSrc)
		} else if !nilSrc.IsChecked {
			// Potential nil source without observed usage
			nilSrc.Risk = RiskMedium
			nilSources = append(nilSources, nilSrc)
		}
	}

	return nilSources, nil
}

// Analyze performs complete data flow analysis on the given files.
func (g *GoAnalyzer) Analyze(files []string) (*FlowAnalysis, error) {
	sources, err := g.DetectSources(files)
	if err != nil {
		return nil, fmt.Errorf("detecting sources: %w", err)
	}

	sinks, err := g.DetectSinks(files)
	if err != nil {
		return nil, fmt.Errorf("detecting sinks: %w", err)
	}

	flows, err := g.TrackFlows(sources, sinks, files)
	if err != nil {
		return nil, fmt.Errorf("tracking flows: %w", err)
	}

	nilSources, err := g.DetectNilSources(files)
	if err != nil {
		return nil, fmt.Errorf("detecting nil sources: %w", err)
	}

	// Calculate statistics
	stats := calculateStats(sources, sinks, flows, nilSources)

	analysis := &FlowAnalysis{
		Language:   g.Language(),
		Sources:    sources,
		Sinks:      sinks,
		Flows:      flows,
		NilSources: nilSources,
		Statistics: stats,
	}

	return analysis, nil
}

// Helper functions

// isGoFile checks if a file is a Go source file.
func isGoFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

// readFileLines reads a file into lines.
func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// extractVariable extracts the variable name from an assignment line.
func extractVariable(line string) string {
	// Match patterns like: var := expr, var, err := expr, var = expr
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\w+)\s*(?:,\s*(?:err|ok|_))?\s*:?=`),
		regexp.MustCompile(`^\s*(\w+)\s*=`),
	}

	for _, p := range patterns {
		if matches := p.FindStringSubmatch(line); len(matches) > 1 {
			varName := matches[1]
			// Skip common non-variable patterns
			if varName != "if" && varName != "for" && varName != "switch" && varName != "_" {
				return varName
			}
		}
	}

	return ""
}

// extractFunctionName extracts the function name from a call expression.
func extractFunctionName(line string) string {
	// Match patterns like: pkg.Function(, obj.Method(, Function(
	pattern := regexp.MustCompile(`(\w+(?:\.\w+)*)\s*\(`)
	if matches := pattern.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// generateFlowID creates a unique identifier for a flow.
func generateFlowID(source Source, sink Sink) string {
	data := fmt.Sprintf("%s:%d:%s:%d", source.File, source.Line, sink.File, sink.Line)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("flow_%x", hash[:8])
}

// calculateRisk determines the risk level of a flow.
func calculateRisk(sourceType SourceType, sinkType SinkType, sanitized bool) RiskLevel {
	if sanitized {
		return RiskLow
	}

	// Critical: User input to command execution or database (SQL injection)
	if (sourceType == SourceHTTPBody || sourceType == SourceHTTPQuery ||
		sourceType == SourceHTTPHeader || sourceType == SourceHTTPPath ||
		sourceType == SourceUserInput) &&
		(sinkType == SinkExec || sinkType == SinkDatabase) {
		return RiskCritical
	}

	// High: User input to response (XSS) or template or redirect (open redirect)
	if (sourceType == SourceHTTPBody || sourceType == SourceHTTPQuery ||
		sourceType == SourceHTTPHeader || sourceType == SourceHTTPPath ||
		sourceType == SourceUserInput) &&
		(sinkType == SinkResponse || sinkType == SinkTemplate || sinkType == SinkRedirect) {
		return RiskHigh
	}

	// Medium: Environment vars to sensitive sinks, file writes
	if sourceType == SourceEnvVar && (sinkType == SinkDatabase || sinkType == SinkExec) {
		return RiskMedium
	}

	// Medium: File content to command execution or database (potential malicious file content)
	if sourceType == SourceFile && (sinkType == SinkExec || sinkType == SinkDatabase) {
		return RiskMedium
	}

	if sinkType == SinkFile {
		return RiskMedium
	}

	// Low: Logging, info exposure
	if sinkType == SinkLog {
		return RiskLow
	}

	return RiskInfo
}

// checkSanitization checks if data is sanitized between source and sink.
func (g *GoAnalyzer) checkSanitization(sourceLine, sinkLine int, lines []string) (bool, []string) {
	var sanitizers []string

	for i := sourceLine; i < sinkLine && i < len(lines); i++ {
		line := lines[i]
		if matches := g.sanitizerRegex.FindAllString(line, -1); matches != nil {
			sanitizers = append(sanitizers, matches...)
		}
	}

	return len(sanitizers) > 0, sanitizers
}

// describeFlow generates a human-readable description of a flow.
func describeFlow(source Source, sink Sink, sanitized bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Data from %s (%s)", source.Type, source.Pattern))

	if source.Variable != "" {
		sb.WriteString(fmt.Sprintf(" in variable '%s'", source.Variable))
	}

	sb.WriteString(fmt.Sprintf(" flows to %s (%s)", sink.Type, sink.Pattern))

	if sanitized {
		sb.WriteString(" [sanitized]")
	} else {
		sb.WriteString(" [unsanitized - potential vulnerability]")
	}

	return sb.String()
}

// buildFlowPath builds the data flow path between source and sink.
func buildFlowPath(source Source, sink Sink, lines []string) []string {
	var path []string

	path = append(path, fmt.Sprintf("%s:%d - Source: %s", filepath.Base(source.File), source.Line, source.Pattern))

	// Add intermediate transformations if any
	if source.Line < sink.Line-1 {
		path = append(path, fmt.Sprintf("... %d lines ...", sink.Line-source.Line-1))
	}

	path = append(path, fmt.Sprintf("%s:%d - Sink: %s", filepath.Base(sink.File), sink.Line, sink.Pattern))

	return path
}

// containsVariableOrDerivative checks if line contains variable or a derivative.
func containsVariableOrDerivative(line, varName string) bool {
	// Direct usage
	if strings.Contains(line, varName) {
		return true
	}

	// Check for common transformations
	derivatives := []string{
		varName + ".",   // Method call
		varName + "[",   // Index access
		"*" + varName,   // Dereference
		"&" + varName,   // Address
		varName + "Str", // Common suffixes
		varName + "String",
		varName + "Bytes",
	}

	for _, d := range derivatives {
		if strings.Contains(line, d) {
			return true
		}
	}

	return false
}

// determineNilOrigin categorizes the source of a potentially nil value.
func determineNilOrigin(line string) string {
	switch {
	case strings.Contains(line, "["):
		return "map_lookup"
	case strings.Contains(line, ".("):
		return "type_assertion"
	case strings.Contains(line, "Query"):
		return "database_query"
	case strings.Contains(line, "Unmarshal"):
		return "json_unmarshal"
	case strings.Contains(line, "Find") || strings.Contains(line, "Get") || strings.Contains(line, "Lookup"):
		return "lookup_operation"
	case strings.Contains(line, "Value("):
		return "context_value"
	case strings.Contains(line, "<-"):
		return "channel_receive"
	default:
		return "unknown"
	}
}

// calculateStats computes statistics for the analysis.
func calculateStats(sources []Source, sinks []Sink, flows []Flow, nilSources []NilSource) Stats {
	stats := Stats{
		TotalSources: len(sources),
		TotalSinks:   len(sinks),
		TotalFlows:   len(flows),
		NilRisks:     len(nilSources),
	}

	for _, flow := range flows {
		if !flow.Sanitized {
			stats.UnsanitizedFlows++
		}
		switch flow.Risk {
		case RiskCritical:
			stats.CriticalFlows++
		case RiskHigh:
			stats.HighRiskFlows++
		}
	}

	for _, ns := range nilSources {
		if !ns.IsChecked {
			stats.UncheckedNilRisks++
		}
	}

	return stats
}
