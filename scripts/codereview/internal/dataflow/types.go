package dataflow

// SourceType categorizes where untrusted data originates
type SourceType string

const (
	SourceHTTPBody   SourceType = "http_body"
	SourceHTTPQuery  SourceType = "http_query"
	SourceHTTPHeader SourceType = "http_header"
	SourceHTTPPath   SourceType = "http_path"
	SourceEnvVar     SourceType = "env_var"
	SourceFile       SourceType = "file_read"
	SourceDatabase   SourceType = "database"
	SourceUserInput  SourceType = "user_input"
	SourceExternal   SourceType = "external_api"
)

// SinkType categorizes where data flows to
type SinkType string

const (
	SinkDatabase SinkType = "database"
	SinkExec     SinkType = "command_exec"
	SinkResponse SinkType = "http_response"
	SinkLog      SinkType = "logging"
	SinkFile     SinkType = "file_write"
	SinkTemplate SinkType = "template"
	SinkRedirect SinkType = "redirect"
)

// RiskLevel indicates severity of a flow
type RiskLevel string

const (
	RiskCritical RiskLevel = "critical"
	RiskHigh     RiskLevel = "high"
	RiskMedium   RiskLevel = "medium"
	RiskLow      RiskLevel = "low"
	RiskInfo     RiskLevel = "info"
)

// Source represents an untrusted data source
type Source struct {
	Type     SourceType `json:"type"`
	File     string     `json:"file"`
	Line     int        `json:"line"`
	Column   int        `json:"column,omitempty"`
	Variable string     `json:"variable"`
	Pattern  string     `json:"pattern"`
	Context  string     `json:"context,omitempty"`
}

// Sink represents a data destination
type Sink struct {
	Type     SinkType `json:"type"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Column   int      `json:"column,omitempty"`
	Function string   `json:"function"`
	Pattern  string   `json:"pattern"`
	Context  string   `json:"context,omitempty"`
}

// Flow represents a data path from source to sink
type Flow struct {
	ID          string    `json:"id"`
	Source      Source    `json:"source"`
	Sink        Sink      `json:"sink"`
	Path        []string  `json:"path"`
	Sanitized   bool      `json:"sanitized"`
	Sanitizers  []string  `json:"sanitizers,omitempty"`
	Risk        RiskLevel `json:"risk"`
	Description string    `json:"description"`
}

// NilSource tracks variables that may be nil/null
type NilSource struct {
	File      string    `json:"file"`
	Line      int       `json:"line"`
	Variable  string    `json:"variable"`
	Origin    string    `json:"origin"`
	IsChecked bool      `json:"is_checked"`
	CheckLine int       `json:"check_line,omitempty"`
	UsageLine int       `json:"usage_line,omitempty"`
	Risk      RiskLevel `json:"risk"`
}

// FlowAnalysis contains all analysis results for a language
type FlowAnalysis struct {
	Language   string      `json:"language"`
	Sources    []Source    `json:"sources"`
	Sinks      []Sink      `json:"sinks"`
	Flows      []Flow      `json:"flows"`
	NilSources []NilSource `json:"nil_sources"`
	Statistics Stats       `json:"statistics"`
}

// Stats provides summary statistics
type Stats struct {
	TotalSources      int `json:"total_sources"`
	TotalSinks        int `json:"total_sinks"`
	TotalFlows        int `json:"total_flows"`
	UnsanitizedFlows  int `json:"unsanitized_flows"`
	CriticalFlows     int `json:"critical_flows"`
	HighRiskFlows     int `json:"high_risk_flows"`
	NilRisks          int `json:"nil_risks"`
	UncheckedNilRisks int `json:"unchecked_nil_risks"`
}

// SecuritySummary aggregates results across languages
type SecuritySummary struct {
	Timestamp  string                  `json:"timestamp"`
	Languages  []string                `json:"languages"`
	Analyses   map[string]FlowAnalysis `json:"analyses"`
	TotalStats Stats                   `json:"total_stats"`
	TopRisks   []Flow                  `json:"top_risks"`
}

// Analyzer interface for language-specific implementations
type Analyzer interface {
	Language() string
	DetectSources(files []string) ([]Source, error)
	DetectSinks(files []string) ([]Sink, error)
	TrackFlows(sources []Source, sinks []Sink, files []string) ([]Flow, error)
	DetectNilSources(files []string) ([]NilSource, error)
	Analyze(files []string) (*FlowAnalysis, error)
}
