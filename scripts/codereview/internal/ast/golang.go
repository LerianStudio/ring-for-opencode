package ast

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
	"unicode"
)

// GoExtractor implements AST extraction for Go files
type GoExtractor struct{}

// NewGoExtractor creates a new Go AST extractor
func NewGoExtractor() *GoExtractor {
	return &GoExtractor{}
}

func (g *GoExtractor) Language() string {
	return "go"
}

func (g *GoExtractor) SupportedExtensions() []string {
	return []string{".go"}
}

// ParsedFile holds parsed AST information for a Go file
type ParsedFile struct {
	Fset      *token.FileSet
	File      *ast.File
	Functions map[string]*GoFunc
	Types     map[string]*GoType
	Imports   map[string]string // path -> alias
}

// GoFunc represents a parsed Go function
type GoFunc struct {
	Name       string
	Receiver   string
	Params     []Param
	Returns    []string
	IsExported bool
	StartLine  int
	EndLine    int
	BodyHash   string
}

// GoType represents a parsed Go type
type GoType struct {
	Name       string
	Kind       string // struct, interface, alias
	Fields     []GoField
	Methods    []string // interface methods
	IsExported bool
	StartLine  int
	EndLine    int
}

// GoField represents a struct field
type GoField struct {
	Name string
	Type string
	Tag  string
}

func (g *GoExtractor) parseFile(path string) (*ParsedFile, error) {
	if path == "" {
		return &ParsedFile{
			Functions: make(map[string]*GoFunc),
			Types:     make(map[string]*GoType),
			Imports:   make(map[string]string),
		}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	parsed := &ParsedFile{
		Fset:      fset,
		File:      file,
		Functions: make(map[string]*GoFunc),
		Types:     make(map[string]*GoType),
		Imports:   make(map[string]string),
	}

	// Extract imports
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		parsed.Imports[path] = alias
	}

	// Extract functions and types
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			fn := g.extractFunc(fset, node)
			key := fn.Name
			if fn.Receiver != "" {
				key = fn.Receiver + "." + fn.Name
			}
			parsed.Functions[key] = fn

		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				for _, spec := range node.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						t := g.extractType(fset, ts)
						parsed.Types[t.Name] = t
					}
				}
			}
		}
		return true
	})

	return parsed, nil
}

func (g *GoExtractor) extractFunc(fset *token.FileSet, fn *ast.FuncDecl) *GoFunc {
	isExported := false
	if len(fn.Name.Name) > 0 {
		isExported = unicode.IsUpper(rune(fn.Name.Name[0]))
	}

	goFn := &GoFunc{
		Name:       fn.Name.Name,
		IsExported: isExported,
		StartLine:  fset.Position(fn.Pos()).Line,
		EndLine:    fset.Position(fn.End()).Line,
	}

	// Extract receiver
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		goFn.Receiver = g.typeToString(fn.Recv.List[0].Type)
	}

	// Extract parameters
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			typeStr := g.typeToString(field.Type)
			if len(field.Names) == 0 {
				goFn.Params = append(goFn.Params, Param{Type: typeStr})
			} else {
				for _, name := range field.Names {
					goFn.Params = append(goFn.Params, Param{
						Name: name.Name,
						Type: typeStr,
					})
				}
			}
		}
	}

	// Extract return types
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			goFn.Returns = append(goFn.Returns, g.typeToString(field.Type))
		}
	}

	// Hash the body for change detection
	if fn.Body != nil {
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, fn.Body)
		goFn.BodyHash = fmt.Sprintf("%x", buf.Bytes())
	}

	return goFn
}

func (g *GoExtractor) extractType(fset *token.FileSet, ts *ast.TypeSpec) *GoType {
	goType := &GoType{
		Name:       ts.Name.Name,
		IsExported: unicode.IsUpper(rune(ts.Name.Name[0])),
		StartLine:  fset.Position(ts.Pos()).Line,
		EndLine:    fset.Position(ts.End()).Line,
	}

	switch t := ts.Type.(type) {
	case *ast.StructType:
		goType.Kind = "struct"
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				typeStr := g.typeToString(field.Type)
				tag := ""
				if field.Tag != nil {
					tag = field.Tag.Value
				}
				if len(field.Names) == 0 {
					// Embedded field
					goType.Fields = append(goType.Fields, GoField{
						Name: typeStr,
						Type: typeStr,
						Tag:  tag,
					})
				} else {
					for _, name := range field.Names {
						goType.Fields = append(goType.Fields, GoField{
							Name: name.Name,
							Type: typeStr,
							Tag:  tag,
						})
					}
				}
			}
		}

	case *ast.InterfaceType:
		goType.Kind = "interface"
		if t.Methods != nil {
			for _, method := range t.Methods.List {
				if len(method.Names) > 0 {
					goType.Methods = append(goType.Methods, method.Names[0].Name)
				}
			}
		}

	default:
		goType.Kind = "alias"
	}

	return goType
}

func (g *GoExtractor) typeToString(expr ast.Expr) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), expr)
	return buf.String()
}

// ExtractDiff compares two Go files and returns semantic differences
func (g *GoExtractor) ExtractDiff(ctx context.Context, beforePath, afterPath string) (*SemanticDiff, error) {
	before, err := g.parseFile(beforePath)
	if err != nil {
		return nil, fmt.Errorf("parsing before file: %w", err)
	}

	after, err := g.parseFile(afterPath)
	if err != nil {
		return nil, fmt.Errorf("parsing after file: %w", err)
	}

	diff := &SemanticDiff{
		Language: "go",
		FilePath: afterPath,
	}

	if afterPath == "" {
		diff.FilePath = beforePath
	}

	// Compare functions
	diff.Functions = g.compareFunctions(before.Functions, after.Functions)

	// Compare types
	diff.Types = g.compareTypes(before.Types, after.Types)

	// Compare imports
	diff.Imports = g.compareImports(before.Imports, after.Imports)

	// Compute summary
	diff.Summary = ComputeSummary(diff.Functions, diff.Types, diff.Imports)

	return diff, nil
}

func (g *GoExtractor) compareFunctions(before, after map[string]*GoFunc) []FunctionDiff {
	var diffs []FunctionDiff

	// Find removed and modified functions
	for name, beforeFn := range before {
		afterFn, exists := after[name]
		if !exists {
			diffs = append(diffs, FunctionDiff{
				Name:       name,
				ChangeType: ChangeRemoved,
				Before:     g.funcToSig(beforeFn),
			})
			continue
		}

		// Check if modified
		if g.funcChanged(beforeFn, afterFn) {
			diff := FunctionDiff{
				Name:       name,
				ChangeType: ChangeModified,
				Before:     g.funcToSig(beforeFn),
				After:      g.funcToSig(afterFn),
			}
			diff.BodyDiff = g.describeFuncChange(beforeFn, afterFn)
			diffs = append(diffs, diff)
		}
	}

	// Find added functions
	for name, afterFn := range after {
		if _, exists := before[name]; !exists {
			diffs = append(diffs, FunctionDiff{
				Name:       name,
				ChangeType: ChangeAdded,
				After:      g.funcToSig(afterFn),
			})
		}
	}

	return diffs
}

func (g *GoExtractor) funcToSig(fn *GoFunc) *FuncSig {
	return &FuncSig{
		Params:     fn.Params,
		Returns:    fn.Returns,
		Receiver:   fn.Receiver,
		IsExported: fn.IsExported,
		StartLine:  fn.StartLine,
		EndLine:    fn.EndLine,
	}
}

func (g *GoExtractor) funcChanged(before, after *GoFunc) bool {
	// Check signature changes
	if !g.paramsEqual(before.Params, after.Params) {
		return true
	}
	if !g.stringsEqual(before.Returns, after.Returns) {
		return true
	}
	if before.Receiver != after.Receiver {
		return true
	}
	// Check body changes
	if before.BodyHash != after.BodyHash {
		return true
	}
	return false
}

func (g *GoExtractor) describeFuncChange(before, after *GoFunc) string {
	var changes []string

	if !g.paramsEqual(before.Params, after.Params) {
		changes = append(changes, "parameters changed")
	}
	if !g.stringsEqual(before.Returns, after.Returns) {
		changes = append(changes, "return types changed")
	}
	if before.Receiver != after.Receiver {
		changes = append(changes, "receiver changed")
	}
	if before.BodyHash != after.BodyHash {
		changes = append(changes, "implementation changed")
	}

	return strings.Join(changes, ", ")
}

func (g *GoExtractor) compareTypes(before, after map[string]*GoType) []TypeDiff {
	var diffs []TypeDiff

	// Find removed and modified types
	for name, beforeType := range before {
		afterType, exists := after[name]
		if !exists {
			diffs = append(diffs, TypeDiff{
				Name:       name,
				Kind:       beforeType.Kind,
				ChangeType: ChangeRemoved,
				StartLine:  beforeType.StartLine,
				EndLine:    beforeType.EndLine,
			})
			continue
		}

		// Check if modified
		fieldDiffs := g.compareFields(beforeType.Fields, afterType.Fields)
		if len(fieldDiffs) > 0 || beforeType.Kind != afterType.Kind {
			diffs = append(diffs, TypeDiff{
				Name:       name,
				Kind:       afterType.Kind,
				ChangeType: ChangeModified,
				Fields:     fieldDiffs,
				StartLine:  afterType.StartLine,
				EndLine:    afterType.EndLine,
			})
		}
	}

	// Find added types
	for name, afterType := range after {
		if _, exists := before[name]; !exists {
			diffs = append(diffs, TypeDiff{
				Name:       name,
				Kind:       afterType.Kind,
				ChangeType: ChangeAdded,
				StartLine:  afterType.StartLine,
				EndLine:    afterType.EndLine,
			})
		}
	}

	return diffs
}

func (g *GoExtractor) compareFields(before, after []GoField) []FieldDiff {
	var diffs []FieldDiff

	beforeMap := make(map[string]GoField)
	for _, f := range before {
		beforeMap[f.Name] = f
	}

	afterMap := make(map[string]GoField)
	for _, f := range after {
		afterMap[f.Name] = f
	}

	// Find removed and modified fields
	for name, beforeField := range beforeMap {
		afterField, exists := afterMap[name]
		if !exists {
			diffs = append(diffs, FieldDiff{
				Name:       name,
				ChangeType: ChangeRemoved,
				OldType:    beforeField.Type,
			})
			continue
		}

		if beforeField.Type != afterField.Type {
			diffs = append(diffs, FieldDiff{
				Name:       name,
				ChangeType: ChangeModified,
				OldType:    beforeField.Type,
				NewType:    afterField.Type,
			})
		}
	}

	// Find added fields
	for name, afterField := range afterMap {
		if _, exists := beforeMap[name]; !exists {
			diffs = append(diffs, FieldDiff{
				Name:       name,
				ChangeType: ChangeAdded,
				NewType:    afterField.Type,
			})
		}
	}

	return diffs
}

func (g *GoExtractor) compareImports(before, after map[string]string) []ImportDiff {
	var diffs []ImportDiff

	for path, beforeAlias := range before {
		if afterAlias, exists := after[path]; !exists {
			diffs = append(diffs, ImportDiff{
				Path:       path,
				Alias:      beforeAlias,
				ChangeType: ChangeRemoved,
			})
		} else if beforeAlias != afterAlias {
			diffs = append(diffs, ImportDiff{
				Path:       path,
				Alias:      afterAlias,
				ChangeType: ChangeModified,
			})
		}
	}

	for path, alias := range after {
		if _, exists := before[path]; !exists {
			diffs = append(diffs, ImportDiff{
				Path:       path,
				Alias:      alias,
				ChangeType: ChangeAdded,
			})
		}
	}

	return diffs
}

func (g *GoExtractor) paramsEqual(a, b []Param) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Type != b[i].Type {
			return false
		}
	}
	return true
}

func (g *GoExtractor) stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
