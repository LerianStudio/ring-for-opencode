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

// ReadJSONFileWithLimit reads a JSON file with size validation to prevent resource exhaustion.
// It validates the path to prevent directory traversal attacks.
func ReadJSONFileWithLimit(path string) ([]byte, error) {
	// Normalize path
	cleanPath := filepath.Clean(path)

	// Prevent directory traversal by rejecting paths with ".." after cleaning
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("path contains directory traversal: %s", path)
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
