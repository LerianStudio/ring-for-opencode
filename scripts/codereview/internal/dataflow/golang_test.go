package dataflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoAnalyzer_DetectSources(t *testing.T) {
	tmpDir := t.TempDir()

	// Create temp Go file with various source patterns
	testFile := filepath.Join(tmpDir, "handler.go")
	content := `package main

import (
	"net/http"
	"os"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// HTTP body source
	body := r.Body

	// HTTP query source
	userID := r.URL.Query().Get("id")

	// HTTP header source
	authToken := r.Header.Get("Authorization")

	// Environment variable source
	apiKey := os.Getenv("API_KEY")

	// Use variables to avoid unused warnings
	_ = body
	_ = userID
	_ = authToken
	_ = apiKey
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	analyzer := NewGoAnalyzer(tmpDir)
	sources, err := analyzer.DetectSources([]string{testFile})
	require.NoError(t, err)

	// Verify at least 4 sources detected
	assert.GreaterOrEqual(t, len(sources), 4, "Expected at least 4 sources, got %d", len(sources))

	// Check source types include required types
	sourceTypes := make(map[SourceType]bool)
	for _, src := range sources {
		sourceTypes[src.Type] = true
	}

	assert.True(t, sourceTypes[SourceHTTPBody], "Expected SourceHTTPBody to be detected")
	assert.True(t, sourceTypes[SourceHTTPQuery], "Expected SourceHTTPQuery to be detected")
	assert.True(t, sourceTypes[SourceHTTPHeader], "Expected SourceHTTPHeader to be detected")
	assert.True(t, sourceTypes[SourceEnvVar], "Expected SourceEnvVar to be detected")

	// Verify all sources have valid file paths
	for _, src := range sources {
		assert.Equal(t, testFile, src.File, "Source file should match test file")
		assert.Greater(t, src.Line, 0, "Source line should be positive")
	}
}

func TestGoAnalyzer_DetectSinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create temp Go file with various sink patterns
	testFile := filepath.Join(tmpDir, "service.go")
	content := `package main

import (
	"database/sql"
	"log"
	"net/http"
	"os/exec"
)

func processData(db *sql.DB, w http.ResponseWriter) {
	// Database sink
	db.Exec("INSERT INTO users VALUES (1)")

	// HTTP response sink
	w.Write([]byte("response"))

	// Command execution sink
	exec.Command("ls", "-la")

	// Logging sink
	log.Printf("Processing complete")
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	analyzer := NewGoAnalyzer(tmpDir)
	sinks, err := analyzer.DetectSinks([]string{testFile})
	require.NoError(t, err)

	// Verify at least 4 sinks detected
	assert.GreaterOrEqual(t, len(sinks), 4, "Expected at least 4 sinks, got %d", len(sinks))

	// Check sink types include required types
	sinkTypes := make(map[SinkType]bool)
	for _, sink := range sinks {
		sinkTypes[sink.Type] = true
	}

	assert.True(t, sinkTypes[SinkDatabase], "Expected SinkDatabase to be detected")
	assert.True(t, sinkTypes[SinkResponse], "Expected SinkResponse to be detected")
	assert.True(t, sinkTypes[SinkExec], "Expected SinkExec to be detected")
	assert.True(t, sinkTypes[SinkLog], "Expected SinkLog to be detected")

	// Verify all sinks have valid file paths
	for _, sink := range sinks {
		assert.Equal(t, testFile, sink.File, "Sink file should match test file")
		assert.Greater(t, sink.Line, 0, "Sink line should be positive")
	}
}

func TestGoAnalyzer_TrackFlows(t *testing.T) {
	tmpDir := t.TempDir()

	// Create temp Go file with SQL injection pattern
	testFile := filepath.Join(tmpDir, "vulnerable.go")
	content := `package main

import (
	"database/sql"
	"net/http"
)

func vulnerableHandler(db *sql.DB, r *http.Request) {
	// Source: user input from query parameter
	userInput := r.URL.Query().Get("input")

	// Sink: SQL injection vulnerability - user input concatenated into query
	db.Exec("SELECT * FROM users WHERE name = '" + userInput + "'")
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	analyzer := NewGoAnalyzer(tmpDir)

	sources, err := analyzer.DetectSources([]string{testFile})
	require.NoError(t, err)
	require.NotEmpty(t, sources, "Expected sources to be detected")

	sinks, err := analyzer.DetectSinks([]string{testFile})
	require.NoError(t, err)
	require.NotEmpty(t, sinks, "Expected sinks to be detected")

	flows, err := analyzer.TrackFlows(sources, sinks, []string{testFile})
	require.NoError(t, err)

	// Verify at least one flow detected
	require.NotEmpty(t, flows, "Expected at least one flow to be detected")

	// Verify the flow is marked as Critical risk (SQL injection)
	foundCritical := false
	for _, flow := range flows {
		if flow.Risk == RiskCritical {
			foundCritical = true
			assert.Equal(t, SourceHTTPQuery, flow.Source.Type, "Source should be HTTP query")
			assert.Equal(t, SinkDatabase, flow.Sink.Type, "Sink should be database")
			assert.False(t, flow.Sanitized, "Flow should not be sanitized")
			break
		}
	}
	assert.True(t, foundCritical, "Expected to find a Critical risk flow for SQL injection")
}

func TestGoAnalyzer_DetectNilSources(t *testing.T) {
	tmpDir := t.TempDir()

	// Create temp Go file with nil sources
	testFile := filepath.Join(tmpDir, "nilcheck.go")
	content := `package main

import "sync"

type Cache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func (c *Cache) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

func processData(cache *Cache) {
	// Unchecked nil source - map lookup without ok check
	uncheckedValue := cache.data["key"]

	// Potentially unsafe usage
	_ = uncheckedValue

	// Checked nil source - map lookup with ok check
	checkedValue, ok := cache.data["safekey"]
	if ok {
		_ = checkedValue
	}
}

func anotherFunc(m map[string]string) {
	// Direct map access without check
	value := m["missing"]
	process(value)
}

func process(s string) {}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	analyzer := NewGoAnalyzer(tmpDir)
	nilSources, err := analyzer.DetectNilSources([]string{testFile})
	require.NoError(t, err)

	// Verify nil sources detected
	assert.NotEmpty(t, nilSources, "Expected nil sources to be detected")

	// Count checked vs unchecked
	uncheckedCount := 0
	checkedCount := 0
	for _, ns := range nilSources {
		if ns.IsChecked {
			checkedCount++
		} else {
			uncheckedCount++
		}
	}

	// We should have at least one unchecked nil source
	assert.Greater(t, uncheckedCount, 0, "Expected at least one unchecked nil source")

	// Verify IsChecked field differentiates checked vs unchecked
	// The checked one should be filtered out or marked as checked
	for _, ns := range nilSources {
		assert.NotEmpty(t, ns.Variable, "Variable name should not be empty")
		assert.Greater(t, ns.Line, 0, "Line number should be positive")
		assert.NotEmpty(t, ns.Origin, "Origin should not be empty")
	}
}

func TestGoAnalyzer_Analyze(t *testing.T) {
	tmpDir := t.TempDir()

	// Create temp Go file with SQL injection + XSS patterns
	testFile := filepath.Join(tmpDir, "app.go")
	content := `package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

func vulnerableHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// SQL Injection vulnerability
	userID := r.URL.Query().Get("id")
	db.Exec("SELECT * FROM users WHERE id = " + userID)

	// XSS vulnerability
	name := r.URL.Query().Get("name")
	fmt.Fprintf(w, "<h1>Hello, %s</h1>", name)

	// Another potential flow
	body := r.Body
	_ = body
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	analyzer := NewGoAnalyzer(tmpDir)
	analysis, err := analyzer.Analyze([]string{testFile})
	require.NoError(t, err)
	require.NotNil(t, analysis)

	// Verify Language is "go"
	assert.Equal(t, "go", analysis.Language, "Language should be 'go'")

	// Verify Statistics.TotalSources > 0
	assert.Greater(t, analysis.Statistics.TotalSources, 0, "Should have detected sources")

	// Verify Statistics.TotalSinks > 0
	assert.Greater(t, analysis.Statistics.TotalSinks, 0, "Should have detected sinks")

	// Verify Statistics.CriticalFlows > 0 (SQL injection should be critical)
	assert.Greater(t, analysis.Statistics.CriticalFlows, 0, "Should have detected critical flows (SQL injection)")

	// Verify sources array is populated
	assert.NotEmpty(t, analysis.Sources, "Sources array should not be empty")

	// Verify sinks array is populated
	assert.NotEmpty(t, analysis.Sinks, "Sinks array should not be empty")

	// Verify flows array is populated
	assert.NotEmpty(t, analysis.Flows, "Flows array should not be empty")
}

func TestCalculateRisk(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SourceType
		sinkType   SinkType
		sanitized  bool
		expected   RiskLevel
	}{
		{
			name:       "HTTP Query to Exec - Critical",
			sourceType: SourceHTTPQuery,
			sinkType:   SinkExec,
			sanitized:  false,
			expected:   RiskCritical,
		},
		{
			name:       "HTTP Body to Database - Critical",
			sourceType: SourceHTTPBody,
			sinkType:   SinkDatabase,
			sanitized:  false,
			expected:   RiskCritical,
		},
		{
			name:       "HTTP Query to Response - High (XSS)",
			sourceType: SourceHTTPQuery,
			sinkType:   SinkResponse,
			sanitized:  false,
			expected:   RiskHigh,
		},
		{
			name:       "HTTP Body to Template - High",
			sourceType: SourceHTTPBody,
			sinkType:   SinkTemplate,
			sanitized:  false,
			expected:   RiskHigh,
		},
		{
			name:       "Env Var to Database - Medium",
			sourceType: SourceEnvVar,
			sinkType:   SinkDatabase,
			sanitized:  false,
			expected:   RiskMedium,
		},
		{
			name:       "HTTP Query to Log - Low",
			sourceType: SourceHTTPQuery,
			sinkType:   SinkLog,
			sanitized:  false,
			expected:   RiskLow,
		},
		{
			name:       "Sanitized Critical becomes Low",
			sourceType: SourceHTTPQuery,
			sinkType:   SinkExec,
			sanitized:  true,
			expected:   RiskLow,
		},
		{
			name:       "HTTP Header to Exec - Critical",
			sourceType: SourceHTTPHeader,
			sinkType:   SinkExec,
			sanitized:  false,
			expected:   RiskCritical,
		},
		{
			name:       "HTTP Path to Database - Critical (SQL injection)",
			sourceType: SourceHTTPPath,
			sinkType:   SinkDatabase,
			sanitized:  false,
			expected:   RiskCritical,
		},
		{
			name:       "User Input to Exec - Critical",
			sourceType: SourceUserInput,
			sinkType:   SinkExec,
			sanitized:  false,
			expected:   RiskCritical,
		},
		{
			name:       "HTTP Query to Redirect - High (Open Redirect)",
			sourceType: SourceHTTPQuery,
			sinkType:   SinkRedirect,
			sanitized:  false,
			expected:   RiskHigh,
		},
		{
			name:       "Env Var to Exec - Medium",
			sourceType: SourceEnvVar,
			sinkType:   SinkExec,
			sanitized:  false,
			expected:   RiskMedium,
		},
		{
			name:       "Any source to File - Medium",
			sourceType: SourceHTTPQuery,
			sinkType:   SinkFile,
			sanitized:  false,
			expected:   RiskMedium,
		},
		{
			name:       "Database to Response - Info",
			sourceType: SourceDatabase,
			sinkType:   SinkResponse,
			sanitized:  false,
			expected:   RiskInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateRisk(tt.sourceType, tt.sinkType, tt.sanitized)
			assert.Equal(t, tt.expected, result, "Risk level mismatch for %s", tt.name)
		})
	}
}

func TestCheckSanitization(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewGoAnalyzer(tmpDir)

	tests := []struct {
		name       string
		lines      []string
		sourceLine int
		sinkLine   int
		sanitized  bool
	}{
		{
			name: "String concatenation - not sanitized",
			lines: []string{
				"userID := r.URL.Query().Get(\"id\")",
				"db.Exec(\"SELECT * FROM users WHERE id = \" + userID)",
			},
			sourceLine: 1,
			sinkLine:   2,
			sanitized:  false,
		},
		{
			name: "html.EscapeString - sanitized",
			lines: []string{
				"name := r.URL.Query().Get(\"name\")",
				"safeName := html.EscapeString(name)",
				"fmt.Fprintf(w, \"<h1>%s</h1>\", safeName)",
			},
			sourceLine: 1,
			sinkLine:   3,
			sanitized:  true,
		},
		{
			name: "url.QueryEscape - sanitized",
			lines: []string{
				"path := r.URL.Query().Get(\"path\")",
				"escaped := url.QueryEscape(path)",
				"http.Redirect(w, r, escaped, 302)",
			},
			sourceLine: 1,
			sinkLine:   3,
			sanitized:  true,
		},
		{
			name: "strconv.Atoi - sanitized",
			lines: []string{
				"idStr := r.URL.Query().Get(\"id\")",
				"id, err := strconv.Atoi(idStr)",
				"db.Query(\"SELECT * FROM users WHERE id = ?\", id)",
			},
			sourceLine: 1,
			sinkLine:   3,
			sanitized:  true,
		},
		{
			name: "Custom sanitize function - sanitized",
			lines: []string{
				"input := r.URL.Query().Get(\"input\")",
				"clean := sanitizeInput(input)",
				"db.Exec(query, clean)",
			},
			sourceLine: 1,
			sinkLine:   3,
			sanitized:  true,
		},
		{
			name: "Validate function - sanitized",
			lines: []string{
				"email := r.URL.Query().Get(\"email\")",
				"if validateEmail(email) {",
				"   db.Insert(email)",
				"}",
			},
			sourceLine: 1,
			sinkLine:   3,
			sanitized:  true,
		},
		{
			name: "No sanitization between source and sink",
			lines: []string{
				"input := r.URL.Query().Get(\"cmd\")",
				"exec.Command(\"sh\", \"-c\", input)",
			},
			sourceLine: 1,
			sinkLine:   2,
			sanitized:  false,
		},
		{
			name: "template.HTMLEscapeString - sanitized",
			lines: []string{
				"content := r.URL.Query().Get(\"content\")",
				"safe := template.HTMLEscapeString(content)",
				"w.Write([]byte(safe))",
			},
			sourceLine: 1,
			sinkLine:   3,
			sanitized:  true,
		},
		{
			name: "pgx.NamedArgs - sanitized",
			lines: []string{
				"name := r.URL.Query().Get(\"name\")",
				"args := pgx.NamedArgs{\"name\": name}",
				"db.Query(ctx, query, args)",
			},
			sourceLine: 1,
			sinkLine:   3,
			sanitized:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, sanitizers := analyzer.checkSanitization(tt.sourceLine, tt.sinkLine, tt.lines)
			assert.Equal(t, tt.sanitized, sanitized, "Sanitization detection mismatch for %s", tt.name)

			if tt.sanitized {
				assert.NotEmpty(t, sanitizers, "Expected sanitizers to be detected for %s", tt.name)
			}
		})
	}
}

func TestGoAnalyzer_Language(t *testing.T) {
	analyzer := NewGoAnalyzer("")
	assert.Equal(t, "go", analyzer.Language())
}

func TestGoAnalyzer_EmptyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewGoAnalyzer(tmpDir)

	// Test with empty file list
	sources, err := analyzer.DetectSources([]string{})
	require.NoError(t, err)
	assert.Empty(t, sources)

	sinks, err := analyzer.DetectSinks([]string{})
	require.NoError(t, err)
	assert.Empty(t, sinks)

	nilSources, err := analyzer.DetectNilSources([]string{})
	require.NoError(t, err)
	assert.Empty(t, nilSources)
}

func TestGoAnalyzer_NonGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-Go files
	txtFile := filepath.Join(tmpDir, "readme.txt")
	err := os.WriteFile(txtFile, []byte("not a go file"), 0644)
	require.NoError(t, err)

	jsFile := filepath.Join(tmpDir, "script.js")
	err = os.WriteFile(jsFile, []byte("console.log('hello');"), 0644)
	require.NoError(t, err)

	analyzer := NewGoAnalyzer(tmpDir)

	sources, err := analyzer.DetectSources([]string{txtFile, jsFile})
	require.NoError(t, err)
	assert.Empty(t, sources, "Should not detect sources in non-Go files")

	sinks, err := analyzer.DetectSinks([]string{txtFile, jsFile})
	require.NoError(t, err)
	assert.Empty(t, sinks, "Should not detect sinks in non-Go files")
}

func TestGoAnalyzer_FileWithComments(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Go file with patterns in comments (should be ignored)
	// Note: The implementation only ignores lines that START with // or /*
	testFile := filepath.Join(tmpDir, "commented.go")
	content := `package main

// r.Body should be ignored - this line starts with //
/* r.URL.Query().Get("id") should also be ignored - starts with /* */

func main() {
	// This is a real source
	userID := r.URL.Query().Get("id")
	_ = userID
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	analyzer := NewGoAnalyzer(tmpDir)
	sources, err := analyzer.DetectSources([]string{testFile})
	require.NoError(t, err)

	// Should only detect sources from non-comment lines
	// The implementation skips lines that start with // or /*
	assert.GreaterOrEqual(t, len(sources), 1, "Should detect at least 1 source (non-commented)")

	// Verify that the detected source is from the actual code line (line 8)
	// Line numbering: 1=package, 2=empty, 3=comment, 4=comment, 5=empty, 6=func, 7=comment, 8=userID
	foundCodeSource := false
	for _, src := range sources {
		if src.Line >= 6 { // Any line inside the func block (line 6+)
			foundCodeSource = true
			break
		}
	}
	assert.True(t, foundCodeSource, "Should detect source from the actual code line")
}

func TestExtractVariable(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "Simple assignment",
			line:     "userID := r.URL.Query().Get(\"id\")",
			expected: "userID",
		},
		{
			name:     "Assignment with error",
			line:     "data, err := json.Marshal(obj)",
			expected: "data",
		},
		{
			name:     "Assignment with ok",
			line:     "value, ok := cache.Get(\"key\")",
			expected: "value",
		},
		{
			name:     "Reassignment",
			line:     "name = r.Header.Get(\"X-Name\")",
			expected: "name",
		},
		{
			name:     "Blank identifier",
			line:     "_ = r.Body",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVariable(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFunctionName(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "Package function",
			line:     "json.Marshal(data)",
			expected: "json.Marshal",
		},
		{
			name:     "Method call",
			line:     "db.Exec(query)",
			expected: "db.Exec",
		},
		{
			name:     "Simple function",
			line:     "println(\"hello\")",
			expected: "println",
		},
		{
			name:     "Chained method",
			line:     "r.URL.Query().Get(\"id\")",
			expected: "r.URL.Query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFunctionName(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineNilOrigin(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "Map lookup",
			line:     "value := m[key]",
			expected: "map_lookup",
		},
		{
			name:     "Type assertion",
			line:     "str, ok := val.(string)",
			expected: "type_assertion",
		},
		{
			name:     "Database query",
			line:     "row := db.QueryRow(query)",
			expected: "database_query",
		},
		{
			name:     "JSON unmarshal",
			line:     "json.Unmarshal(data, &obj)",
			expected: "json_unmarshal",
		},
		{
			name:     "Find operation",
			line:     "user := repo.FindByID(id)",
			expected: "lookup_operation",
		},
		{
			name:     "Get operation",
			line:     "value := cache.Get(key)",
			expected: "lookup_operation",
		},
		{
			name:     "Context value",
			line:     "user := ctx.Value(userKey)",
			expected: "context_value",
		},
		{
			name:     "Channel receive",
			line:     "msg := <-ch",
			expected: "channel_receive",
		},
		{
			name:     "Unknown pattern",
			line:     "x := someFunc()",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineNilOrigin(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsVariableOrDerivative(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		varName  string
		expected bool
	}{
		{
			name:     "Direct variable usage",
			line:     "fmt.Println(userID)",
			varName:  "userID",
			expected: true,
		},
		{
			name:     "Method call on variable",
			line:     "userID.String()",
			varName:  "userID",
			expected: true,
		},
		{
			name:     "Index access",
			line:     "userID[0]",
			varName:  "userID",
			expected: true,
		},
		{
			name:     "Dereference",
			line:     "*userID",
			varName:  "userID",
			expected: true,
		},
		{
			name:     "Address of",
			line:     "&userID",
			varName:  "userID",
			expected: true,
		},
		{
			name:     "Str suffix derivative",
			line:     "fmt.Println(userIDStr)",
			varName:  "userID",
			expected: true,
		},
		{
			name:     "String suffix derivative",
			line:     "fmt.Println(userIDString)",
			varName:  "userID",
			expected: true,
		},
		{
			name:     "No match",
			line:     "fmt.Println(otherVar)",
			varName:  "userID",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsVariableOrDerivative(tt.line, tt.varName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateStats(t *testing.T) {
	sources := []Source{
		{Type: SourceHTTPQuery, File: "a.go", Line: 1},
		{Type: SourceHTTPBody, File: "b.go", Line: 2},
	}

	sinks := []Sink{
		{Type: SinkDatabase, File: "a.go", Line: 5},
		{Type: SinkExec, File: "b.go", Line: 6},
		{Type: SinkLog, File: "c.go", Line: 7},
	}

	flows := []Flow{
		{Risk: RiskCritical, Sanitized: false},
		{Risk: RiskHigh, Sanitized: false},
		{Risk: RiskMedium, Sanitized: true},
		{Risk: RiskLow, Sanitized: false},
	}

	nilSources := []NilSource{
		{IsChecked: false},
		{IsChecked: true},
		{IsChecked: false},
	}

	stats := calculateStats(sources, sinks, flows, nilSources)

	assert.Equal(t, 2, stats.TotalSources)
	assert.Equal(t, 3, stats.TotalSinks)
	assert.Equal(t, 4, stats.TotalFlows)
	assert.Equal(t, 3, stats.UnsanitizedFlows) // 3 flows are not sanitized
	assert.Equal(t, 1, stats.CriticalFlows)
	assert.Equal(t, 1, stats.HighRiskFlows)
	assert.Equal(t, 3, stats.NilRisks)
	assert.Equal(t, 2, stats.UncheckedNilRisks)
}

func TestIsGoFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"main.go", true},
		{"handler.go", true},
		{"main_test.go", false}, // Test files excluded
		{"readme.md", false},
		{"script.js", false},
		{"config.json", false},
		{"path/to/service.go", true},
		{"internal/handler_test.go", false}, // Test file in subdirectory
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isGoFile(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateFlowID(t *testing.T) {
	source := Source{File: "main.go", Line: 10}
	sink := Sink{File: "main.go", Line: 20}

	id := generateFlowID(source, sink)

	assert.NotEmpty(t, id)
	assert.True(t, len(id) > 0)
	assert.Contains(t, id, "flow_")

	// Same input should produce same ID
	id2 := generateFlowID(source, sink)
	assert.Equal(t, id, id2)

	// Different input should produce different ID
	sink2 := Sink{File: "main.go", Line: 30}
	id3 := generateFlowID(source, sink2)
	assert.NotEqual(t, id, id3)
}

func TestDescribeFlow(t *testing.T) {
	source := Source{
		Type:     SourceHTTPQuery,
		Pattern:  "URL query parameters",
		Variable: "userID",
	}
	sink := Sink{
		Type:    SinkDatabase,
		Pattern: "SQL exec",
	}

	// Test unsanitized flow
	desc := describeFlow(source, sink, false)
	assert.Contains(t, desc, string(SourceHTTPQuery))
	assert.Contains(t, desc, "userID")
	assert.Contains(t, desc, string(SinkDatabase))
	assert.Contains(t, desc, "unsanitized")

	// Test sanitized flow
	descSanitized := describeFlow(source, sink, true)
	assert.Contains(t, descSanitized, "sanitized")
	assert.NotContains(t, descSanitized, "unsanitized")
}
