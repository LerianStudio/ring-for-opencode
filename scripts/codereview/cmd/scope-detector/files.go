package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
)

func resolveFilePatterns(filesFlag, filesFrom string) ([]string, error) {
	patterns := splitCSV(filesFlag)
	if filesFrom != "" {
		filePatterns, err := readPatternsFile(filesFrom)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, filePatterns...)
	}

	return normalizePatterns(patterns), nil
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func readPatternsFile(path string) ([]string, error) {
	cleaned, err := fileutil.ValidatePath(path, ".")
	if err != nil {
		return nil, fmt.Errorf("patterns file path invalid: %w", err)
	}

	file, err := os.Open(cleaned)
	if err != nil {
		return nil, fmt.Errorf("failed to read patterns file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var patterns []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read patterns file: %w", err)
	}
	return patterns, nil
}

func normalizePatterns(patterns []string) []string {
	result := make([]string, 0, len(patterns))
	seen := make(map[string]bool)
	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if trimmed == "" {
			continue
		}
		if !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, trimmed)
		}
	}
	return result
}
