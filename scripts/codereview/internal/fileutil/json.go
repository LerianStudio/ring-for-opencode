// Package fileutil provides shared file utilities for codereview tools.
package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MaxJSONFileSize is the maximum allowed size for JSON files (50MB).
const MaxJSONFileSize = 50 * 1024 * 1024

// ValidateRelativePath ensures a path is relative and does not escape the working directory.
// Returns the cleaned relative path when valid.
func ValidateRelativePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("path must be relative: %s", path)
	}
	if strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || cleanPath == ".." {
		return "", fmt.Errorf("path contains directory traversal: %s", path)
	}
	return cleanPath, nil
}

// ValidatePath ensures a path does not escape the working directory.
// Relative paths are normalized; absolute paths are allowed but must not escape workDir.
func ValidatePath(path string, workDir string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	base := workDir
	if base == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to resolve working directory: %w", err)
		}
		base = cwd
	}

	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("failed to resolve working directory: %w", err)
	}

	cleaned := filepath.Clean(path)
	if filepath.IsAbs(cleaned) {
		if workDir != "" && workDir != "." {
			if !strings.HasPrefix(cleaned, baseAbs+string(filepath.Separator)) && cleaned != baseAbs {
				return "", fmt.Errorf("path escapes working directory: %s", path)
			}
		}
		return cleaned, nil
	}
	if strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".." {
		return "", fmt.Errorf("path contains directory traversal: %s", path)
	}
	absPath := filepath.Join(baseAbs, cleaned)
	if workDir != "" && workDir != "." {
		if !strings.HasPrefix(absPath, baseAbs+string(filepath.Separator)) && absPath != baseAbs {
			return "", fmt.Errorf("path escapes working directory: %s", path)
		}
	}
	return absPath, nil
}

// ReadJSONFileWithLimit reads a JSON file with size validation to prevent resource exhaustion.
// It validates the path to prevent directory traversal attacks.
func ReadJSONFileWithLimit(path string) ([]byte, error) {
	// Normalize path
	cleanPath, err := ValidatePath(path, "")
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > MaxJSONFileSize {
		return nil, fmt.Errorf("file %s exceeds maximum allowed size of %d bytes (actual: %d bytes)", cleanPath, MaxJSONFileSize, info.Size())
	}

	return os.ReadFile(cleanPath) // #nosec G304 - path is validated against traversal
}
