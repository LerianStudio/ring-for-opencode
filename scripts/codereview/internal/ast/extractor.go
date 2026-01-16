package ast

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// Extractor defines the interface for language-specific AST extractors
type Extractor interface {
	// ExtractDiff compares two file versions and returns semantic differences
	ExtractDiff(ctx context.Context, beforePath, afterPath string) (*SemanticDiff, error)

	// SupportedExtensions returns file extensions this extractor handles
	SupportedExtensions() []string

	// Language returns the language name
	Language() string
}

// Registry holds all registered extractors
type Registry struct {
	extractors map[string]Extractor
}

// NewRegistry creates a new extractor registry
func NewRegistry() *Registry {
	return &Registry{
		extractors: make(map[string]Extractor),
	}
}

// Register adds an extractor to the registry
func (r *Registry) Register(e Extractor) {
	for _, ext := range e.SupportedExtensions() {
		r.extractors[ext] = e
	}
}

// GetExtractor returns the appropriate extractor for a file
func (r *Registry) GetExtractor(filePath string) (Extractor, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if e, ok := r.extractors[ext]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("no extractor registered for extension: %s", ext)
}

// ExtractAll processes multiple file pairs and returns semantic diffs
func (r *Registry) ExtractAll(ctx context.Context, pairs []FilePair) ([]SemanticDiff, error) {
	results := make([]SemanticDiff, 0, len(pairs))

	for _, pair := range pairs {
		path := pair.AfterPath
		if path == "" {
			path = pair.BeforePath
		}

		extractor, err := r.GetExtractor(path)
		if err != nil {
			results = append(results, SemanticDiff{
				FilePath: path,
				Error:    err.Error(),
			})
			continue
		}

		diff, err := extractor.ExtractDiff(ctx, pair.BeforePath, pair.AfterPath)
		if err != nil {
			results = append(results, SemanticDiff{
				FilePath: path,
				Language: extractor.Language(),
				Error:    err.Error(),
			})
			continue
		}

		results = append(results, *diff)
	}

	return results, nil
}

// ComputeSummary calculates the change summary from diffs
func ComputeSummary(funcs []FunctionDiff, types []TypeDiff, imports []ImportDiff) ChangeSummary {
	summary := ChangeSummary{}

	for _, f := range funcs {
		switch f.ChangeType {
		case ChangeAdded:
			summary.FunctionsAdded++
		case ChangeRemoved:
			summary.FunctionsRemoved++
		case ChangeModified, ChangeRenamed:
			summary.FunctionsModified++
		}
	}

	for _, t := range types {
		switch t.ChangeType {
		case ChangeAdded:
			summary.TypesAdded++
		case ChangeRemoved:
			summary.TypesRemoved++
		case ChangeModified, ChangeRenamed:
			summary.TypesModified++
		}
	}

	for _, i := range imports {
		switch i.ChangeType {
		case ChangeAdded:
			summary.ImportsAdded++
		case ChangeRemoved:
			summary.ImportsRemoved++
		}
	}

	return summary
}
