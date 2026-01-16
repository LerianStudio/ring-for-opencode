package ast

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrPathTraversal indicates a path traversal attack was detected.
var ErrPathTraversal = errors.New("path traversal detected")

// PathValidator provides path validation against a base directory to prevent
// path traversal attacks.
type PathValidator struct {
	baseDir string
}

// NewPathValidator creates a new PathValidator with the given base directory.
// The base directory is resolved to an absolute path.
func NewPathValidator(baseDir string) (*PathValidator, error) {
	if baseDir == "" {
		// Default to current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		baseDir = cwd
	}

	// Resolve to absolute path
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}

	// Evaluate symlinks to prevent symlink-based traversal
	realBase, err := filepath.EvalSymlinks(absBase)
	if err != nil {
		// If the path doesn't exist yet, use the absolute path
		if os.IsNotExist(err) {
			realBase = absBase
		} else {
			return nil, fmt.Errorf("failed to evaluate symlinks for base directory: %w", err)
		}
	}

	return &PathValidator{baseDir: realBase}, nil
}

// ValidatePath checks if the given path is within the allowed base directory.
// Returns the resolved absolute path if valid, or an error if the path escapes
// the base directory.
func (v *PathValidator) ValidatePath(path string) (string, error) {
	if path == "" {
		// Empty path is allowed (represents no file)
		return "", nil
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Evaluate symlinks to prevent symlink-based traversal
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If file doesn't exist, check the parent directory
		if os.IsNotExist(err) {
			// For non-existent files, validate the parent directory exists
			// and that the constructed path would be within bounds
			parentDir := filepath.Dir(absPath)
			realParent, err := filepath.EvalSymlinks(parentDir)
			if err != nil {
				if os.IsNotExist(err) {
					return "", fmt.Errorf("parent directory does not exist: %s", parentDir)
				}
				return "", fmt.Errorf("failed to evaluate symlinks: %w", err)
			}
			// Reconstruct the path with the real parent
			realPath = filepath.Join(realParent, filepath.Base(absPath))
		} else {
			return "", fmt.Errorf("failed to evaluate symlinks: %w", err)
		}
	}

	// Clean the paths for consistent comparison
	cleanBase := filepath.Clean(v.baseDir)
	cleanPath := filepath.Clean(realPath)

	// Check if the path is within the base directory
	// The path must either equal the base or be a child of it
	if !strings.HasPrefix(cleanPath, cleanBase+string(filepath.Separator)) && cleanPath != cleanBase {
		return "", fmt.Errorf("%w: path %q escapes base directory %q", ErrPathTraversal, path, v.baseDir)
	}

	return realPath, nil
}

// ValidatePaths validates multiple paths and returns them all if valid.
func (v *PathValidator) ValidatePaths(paths ...string) ([]string, error) {
	result := make([]string, len(paths))
	for i, path := range paths {
		validated, err := v.ValidatePath(path)
		if err != nil {
			return nil, err
		}
		result[i] = validated
	}
	return result, nil
}

// BaseDir returns the base directory used for validation.
func (v *PathValidator) BaseDir() string {
	return v.baseDir
}

// ValidatePath is a convenience function that validates a path against a base directory.
// If baseDir is empty, it uses the current working directory.
func ValidatePath(path, baseDir string) (string, error) {
	validator, err := NewPathValidator(baseDir)
	if err != nil {
		return "", err
	}
	return validator.ValidatePath(path)
}
