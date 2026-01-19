package scope

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerianstudio/ring/scripts/codereview/internal/git"
)

// DetectFromFiles analyzes an explicit file list (with optional base ref) for scope.
func (d *Detector) DetectFromFiles(baseRef string, files []string) (*ScopeResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	cleanFiles := normalizeFileList(files)
	if len(cleanFiles) == 0 {
		return nil, fmt.Errorf("no valid files provided")
	}

	workDir := d.workDir
	if workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		workDir = cwd
	}
	if baseRef == "" {
		baseRef = "HEAD"
	}

	return d.buildScopeResultFromFiles(baseRef, cleanFiles)
}

func resolveFileStatus(client gitClientInterface, workDir, baseRef, file string) (git.FileStatus, error) {
	if baseRef == "" {
		baseRef = "HEAD"
	}

	inBase, err := client.FileExistsAtRef(baseRef, file)
	if err != nil {
		return git.StatusUnknown, err
	}

	inWorktree, err := fileExists(filepath.Join(workDir, file))
	if err != nil {
		return git.StatusUnknown, err
	}

	switch {
	case inBase && inWorktree:
		return git.StatusModified, nil
	case inBase && !inWorktree:
		return git.StatusDeleted, nil
	case !inBase && inWorktree:
		return git.StatusAdded, nil
	default:
		return git.StatusUnknown, fmt.Errorf("file not found in base or working tree: %s", file)
	}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func findFileStats(statsByFile map[string]git.FileStats, file string) git.FileStats {
	if len(statsByFile) == 0 {
		return git.FileStats{}
	}
	if stats, ok := statsByFile[file]; ok {
		return stats
	}
	cleaned := filepath.Clean(file)
	for path, stats := range statsByFile {
		if filepath.Clean(path) == cleaned {
			return stats
		}
	}
	return git.FileStats{}
}

func normalizeFileList(files []string) []string {
	result := make([]string, 0, len(files))
	seen := make(map[string]bool)
	for _, file := range files {
		if file == "" {
			continue
		}
		cleaned := filepath.Clean(file)
		if cleaned == "." {
			continue
		}
		if !seen[cleaned] {
			seen[cleaned] = true
			result = append(result, cleaned)
		}
	}
	return result
}
