package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func callgraphLanguagesFile(outputDir string) string {
	return filepath.Join(outputDir, "callgraph-languages.json")
}

func writeCallgraphLanguageFile(outputDir string, languages []string) error {
	if len(languages) == 0 {
		return fmt.Errorf("no callgraph languages provided")
	}
	payload := struct {
		Languages []string `json:"languages"`
	}{
		Languages: languages,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal callgraph languages: %w", err)
	}

	path := callgraphLanguagesFile(outputDir)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write callgraph languages: %w", err)
	}
	return nil
}

func languagesForCallgraph(outputDir string) ([]string, error) {
	scope, err := readScopeJSON(outputDir)
	if err != nil {
		return nil, err
	}

	languageSet := make(map[string]bool)
	if len(scope.Languages) > 0 {
		for _, lang := range scope.Languages {
			normalized := normalizeCallgraphLanguage(lang)
			if normalized != "" {
				languageSet[normalized] = true
			}
		}
	} else if scope.Language != "" {
		normalized := normalizeCallgraphLanguage(scope.Language)
		if normalized != "" {
			languageSet[normalized] = true
		}
	}

	if len(languageSet) == 0 {
		return nil, fmt.Errorf("no supported languages detected for callgraph")
	}

	languages := make([]string, 0, len(languageSet))
	for lang := range languageSet {
		languages = append(languages, lang)
	}
	orderCallgraphLanguages(languages)
	return languages, nil
}

func normalizeCallgraphLanguage(lang string) string {
	if lang == "" {
		return ""
	}
	normalized := strings.ToLower(strings.TrimSpace(lang))
	switch normalized {
	case "go", "golang":
		return "go"
	case "typescript", "ts", "javascript", "js":
		return "typescript"
	case "python", "py":
		return "python"
	case "mixed", "unknown":
		return ""
	default:
		return normalized
	}
}

func orderCallgraphLanguages(languages []string) {
	priority := map[string]int{"go": 0, "typescript": 1, "python": 2}
	sort.SliceStable(languages, func(i, j int) bool {
		pi, okI := priority[languages[i]]
		if !okI {
			pi = len(priority) + 1
		}
		pj, okJ := priority[languages[j]]
		if !okJ {
			pj = len(priority) + 1
		}
		if pi != pj {
			return pi < pj
		}
		return languages[i] < languages[j]
	})
}
