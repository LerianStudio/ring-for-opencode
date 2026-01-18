package scope

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// ExpandFilePatterns expands glob patterns into a list of repo-relative file paths.
// Supports ** for matching multiple path segments.
func ExpandFilePatterns(workDir string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("no file patterns provided")
	}

	baseDir := workDir
	if baseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve working directory: %w", err)
		}
		baseDir = cwd
	}

	matches := make(map[string]bool)
	for _, raw := range patterns {
		pattern := strings.TrimSpace(raw)
		if pattern == "" {
			continue
		}

		if err := validatePattern(pattern); err != nil {
			return nil, err
		}

		if hasGlob(pattern) {
			found, err := expandGlobPattern(baseDir, pattern)
			if err != nil {
				return nil, err
			}
			if len(found) == 0 {
				continue
			}
			for _, match := range found {
				matches[match] = true
			}
			continue
		}

		cleaned := normalizePath(pattern)
		matches[cleaned] = true
	}

	if len(matches) == 0 {
		return []string{}, nil
	}

	result := make([]string, 0, len(matches))
	for file := range matches {
		cleaned := normalizePath(file)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}
	sort.Strings(result)
	return result, nil
}

func validatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	if filepath.IsAbs(pattern) {
		return fmt.Errorf("pattern must be relative: %s", pattern)
	}
	cleaned := filepath.Clean(pattern)
	if cleaned == "." {
		return fmt.Errorf("pattern must not be current directory")
	}

	for _, segment := range strings.Split(cleaned, string(filepath.Separator)) {
		if segment == ".." {
			return fmt.Errorf("pattern contains path traversal: %s", pattern)
		}
	}
	return nil
}

func normalizePath(value string) string {
	cleaned := filepath.Clean(value)
	cleaned = strings.TrimPrefix(cleaned, "./")
	cleaned = strings.TrimPrefix(cleaned, ".\\")
	return cleaned
}

func hasGlob(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

func expandGlobPattern(baseDir, pattern string) ([]string, error) {
	var results []string
	normalizedPattern := path.Clean(filepath.ToSlash(pattern))
	if strings.HasPrefix(normalizedPattern, "../") || normalizedPattern == ".." {
		return nil, fmt.Errorf("pattern contains path traversal: %s", pattern)
	}

	err := filepath.WalkDir(baseDir, func(fullPath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules", "vendor":
				return filepath.SkipDir
			default:
				return nil
			}
		}

		rel, err := filepath.Rel(baseDir, fullPath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		matched, err := matchGlob(normalizedPattern, rel)
		if err != nil {
			return err
		}
		if matched {
			results = append(results, normalizePath(rel))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}

func matchGlob(pattern, target string) (bool, error) {
	pattern = path.Clean(pattern)
	target = path.Clean(target)

	patternSegments := strings.Split(pattern, "/")
	targetSegments := strings.Split(target, "/")

	return matchGlobSegments(patternSegments, targetSegments)
}

func matchGlobSegments(patternSegments, targetSegments []string) (bool, error) {
	if len(patternSegments) == 0 {
		return len(targetSegments) == 0, nil
	}

	segment := patternSegments[0]
	if segment == "**" {
		if len(patternSegments) == 1 {
			return true, nil
		}
		for i := 0; i <= len(targetSegments); i++ {
			matched, err := matchGlobSegments(patternSegments[1:], targetSegments[i:])
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}

	if len(targetSegments) == 0 {
		return false, nil
	}

	ok, err := path.Match(segment, targetSegments[0])
	if err != nil {
		return false, fmt.Errorf("invalid glob pattern %q: %w", segment, err)
	}
	if !ok {
		return false, nil
	}

	return matchGlobSegments(patternSegments[1:], targetSegments[1:])
}
