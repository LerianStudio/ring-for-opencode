package ast

// ChangeType represents the kind of change detected
type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeRemoved  ChangeType = "removed"
	ChangeModified ChangeType = "modified"
	ChangeRenamed  ChangeType = "renamed"
)

// Param represents a function parameter
type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// FieldDiff represents a change in a struct/class field
type FieldDiff struct {
	Name       string     `json:"name"`
	ChangeType ChangeType `json:"change_type"`
	OldType    string     `json:"old_type,omitempty"`
	NewType    string     `json:"new_type,omitempty"`
}

// FuncSig represents a function signature
type FuncSig struct {
	Params     []Param  `json:"params"`
	Returns    []string `json:"returns"`
	Receiver   string   `json:"receiver,omitempty"`
	IsAsync    bool     `json:"is_async,omitempty"`
	Decorators []string `json:"decorators,omitempty"`
	IsExported bool     `json:"is_exported"`
	StartLine  int      `json:"start_line"`
	EndLine    int      `json:"end_line"`
}

// FunctionDiff represents a change in a function
type FunctionDiff struct {
	Name       string     `json:"name"`
	ChangeType ChangeType `json:"change_type"`
	Before     *FuncSig   `json:"before,omitempty"`
	After      *FuncSig   `json:"after,omitempty"`
	BodyDiff   string     `json:"body_diff,omitempty"`
}

// TypeDiff represents a change in a type definition
type TypeDiff struct {
	Name       string      `json:"name"`
	Kind       string      `json:"kind"`
	ChangeType ChangeType  `json:"change_type"`
	Fields     []FieldDiff `json:"fields,omitempty"`
	StartLine  int         `json:"start_line"`
	EndLine    int         `json:"end_line"`
}

// ImportDiff represents a change in imports
type ImportDiff struct {
	Path       string     `json:"path"`
	Alias      string     `json:"alias,omitempty"`
	ChangeType ChangeType `json:"change_type"`
}

// ChangeSummary provides counts of changes
type ChangeSummary struct {
	FunctionsAdded    int `json:"functions_added"`
	FunctionsRemoved  int `json:"functions_removed"`
	FunctionsModified int `json:"functions_modified"`
	TypesAdded        int `json:"types_added"`
	TypesRemoved      int `json:"types_removed"`
	TypesModified     int `json:"types_modified"`
	ImportsAdded      int `json:"imports_added"`
	ImportsRemoved    int `json:"imports_removed"`
}

// SemanticDiff represents the complete semantic diff for a file
type SemanticDiff struct {
	Language  string         `json:"language"`
	FilePath  string         `json:"file_path"`
	Functions []FunctionDiff `json:"functions"`
	Types     []TypeDiff     `json:"types"`
	Imports   []ImportDiff   `json:"imports"`
	Summary   ChangeSummary  `json:"summary"`
	Error     string         `json:"error,omitempty"`
}

// FilePair represents before/after versions of a file
type FilePair struct {
	BeforePath string `json:"before_path"`
	AfterPath  string `json:"after_path"`
	Language   string `json:"language,omitempty"`
}
