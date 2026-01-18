// Package git provides utilities for interacting with git repositories.
// It wraps git CLI commands using exec.Command (no external dependencies).
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// FileStatus represents the status of a file in git diff output.
type FileStatus int

const (
	StatusUnknown FileStatus = iota
	StatusAdded
	StatusModified
	StatusDeleted
	StatusRenamed
	StatusCopied
)

// String returns the single-character git status code.
func (s FileStatus) String() string {
	switch s {
	case StatusAdded:
		return "A"
	case StatusModified:
		return "M"
	case StatusDeleted:
		return "D"
	case StatusRenamed:
		return "R"
	case StatusCopied:
		return "C"
	default:
		return "?"
	}
}

// ParseFileStatus converts a git status string to FileStatus.
// Handles both single-char ("M") and similarity-prefixed ("R100") formats.
func ParseFileStatus(s string) FileStatus {
	if len(s) == 0 {
		return StatusUnknown
	}
	switch s[0] {
	case 'A':
		return StatusAdded
	case 'M':
		return StatusModified
	case 'D':
		return StatusDeleted
	case 'R':
		return StatusRenamed
	case 'C':
		return StatusCopied
	default:
		return StatusUnknown
	}
}

// ChangedFile represents a single file change in a git diff.
type ChangedFile struct {
	Path      string     // Current path of the file
	Status    FileStatus // Type of change (A/M/D/R/C)
	OldPath   string     // Previous path (for renames/copies)
	Additions int        // Lines added
	Deletions int        // Lines deleted
}

// DiffStats contains aggregate statistics for a diff.
type DiffStats struct {
	TotalFiles     int
	TotalAdditions int
	TotalDeletions int
}

// DiffResult contains the complete result of a git diff operation.
type DiffResult struct {
	BaseRef    string
	HeadRef    string
	Files      []ChangedFile
	Stats      DiffStats
	StatsError string
}

// FileStats contains per-file additions/deletions.
type FileStats struct {
	Additions int
	Deletions int
}

// Client provides methods for interacting with git.
type Client struct {
	workDir string
	runner  func(workDir string, args ...string) ([]byte, error)
}

// FileExistsAtRef checks if a file exists at a given git ref.
func (c *Client) FileExistsAtRef(ref, path string) (bool, error) {
	if err := validateRef(ref); err != nil {
		return false, err
	}
	if path == "" {
		return false, fmt.Errorf("invalid file path: empty")
	}

	_, err := c.runGit("cat-file", "-e", ref+":"+path)
	if err != nil {
		if isMissingPathError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file at ref %s: %w", ref, err)
	}

	return true, nil
}

// NewClient creates a new git client for the specified directory.
// If workDir is empty, commands run in the current directory.
func NewClient(workDir string) *Client {
	return &Client{workDir: workDir, runner: runGitCommand}
}

func validateRef(ref string) error {
	if ref == "" {
		return nil
	}
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("invalid ref %q: cannot start with '-'", ref)
	}
	for _, r := range ref {
		if r == '.' || r == '/' || r == '_' || r == '-' || r == ':' || r == '@' || r == '~' || r == '^' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		return fmt.Errorf("invalid ref %q: contains unsupported character %q", ref, r)
	}
	return nil
}

// runGitCommand executes a git command and returns stdout.
func runGitCommand(workDir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...) // #nosec G204 - args validated by validateRef and internal usage
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Env = append([]string{"LC_ALL=C"}, os.Environ()...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s failed: %w\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// runGit executes a git command and returns stdout.
func (c *Client) runGit(args ...string) ([]byte, error) {
	runner := c.runner
	if runner == nil {
		runner = runGitCommand
	}
	return runner(c.workDir, args...)
}

// ShowFile returns file content from a git ref.
func (c *Client) ShowFile(ref, path string) ([]byte, error) {
	if err := validateRef(ref); err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("invalid path: empty")
	}
	if strings.HasPrefix(path, "-") {
		return nil, fmt.Errorf("invalid path: cannot start with '-'")
	}
	return c.runGit("show", ref+":"+path)
}

func isMissingPathError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "does not exist in") || strings.Contains(msg, "Path '")
}

// GetDiff returns the diff between two refs, or combined changes vs HEAD when refs are empty.
// baseRef: starting point (e.g., "main", commit SHA). Empty = HEAD.
// headRef: ending point (e.g., "HEAD", commit SHA). Empty = working tree.
// If baseRef is empty and headRef is set, the diff compares HEAD to headRef.
func (c *Client) GetDiff(baseRef, headRef string) (*DiffResult, error) {
	result := &DiffResult{
		BaseRef: baseRef,
		HeadRef: headRef,
		Files:   make([]ChangedFile, 0),
	}

	if err := validateRef(baseRef); err != nil {
		return nil, err
	}
	if err := validateRef(headRef); err != nil {
		return nil, err
	}

	args := []string{"diff", "--name-status", "-z"}

	switch {
	case baseRef == "" && headRef == "":
		args = append(args, "HEAD")
	case baseRef != "" && headRef == "":
		args = append(args, baseRef)
	case baseRef == "" && headRef != "":
		args = append(args, "HEAD", headRef)
	default:
		args = append(args, baseRef, headRef)
	}

	output, err := c.runGit(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff name-status: %w", err)
	}

	files, err := parseNameStatusOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff name-status: %w", err)
	}
	result.Files = files

	statsArgs := []string{"diff", "--numstat", "-z"}
	switch {
	case baseRef == "" && headRef == "":
		statsArgs = append(statsArgs, "HEAD")
	case baseRef != "" && headRef == "":
		statsArgs = append(statsArgs, baseRef)
	case baseRef == "" && headRef != "":
		statsArgs = append(statsArgs, "HEAD", headRef)
	default:
		statsArgs = append(statsArgs, baseRef, headRef)
	}

	statsOutput, err := c.runGit(statsArgs...)
	if err != nil {
		result.Stats.TotalFiles = len(result.Files)
		result.StatsError = fmt.Sprintf("failed to get diff numstat: %v", err)
		return result, nil
	}

	statsMap, err := parseNumstat(statsOutput)
	if err != nil {
		result.Stats.TotalFiles = len(result.Files)
		result.StatsError = fmt.Sprintf("failed to parse diff numstat: %v", err)
		statsMap = make(map[string]fileStats)
	}
	for i, f := range result.Files {
		if stats, ok := statsMap[f.Path]; ok {
			result.Files[i].Additions = stats.additions
			result.Files[i].Deletions = stats.deletions
			result.Stats.TotalAdditions += stats.additions
			result.Stats.TotalDeletions += stats.deletions
		}
	}
	result.Stats.TotalFiles = len(result.Files)

	return result, nil
}

// GetStagedDiff returns only staged changes (index vs HEAD).
func (c *Client) GetStagedDiff() (*DiffResult, error) {
	result := &DiffResult{
		BaseRef: "HEAD",
		HeadRef: "staged",
		Files:   make([]ChangedFile, 0),
	}

	output, err := c.runGit("diff", "--name-status", "--cached", "-z")
	if err != nil {
		return nil, fmt.Errorf("failed to get staged diff: %w", err)
	}

	files, err := parseNameStatusOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse staged name-status: %w", err)
	}
	result.Files = files

	statsOutput, err := c.runGit("diff", "--numstat", "-z", "--cached")
	if err != nil {
		result.Stats.TotalFiles = len(result.Files)
		result.StatsError = fmt.Sprintf("failed to get staged numstat: %v", err)
		return result, nil
	}

	statsMap, parseErr := parseNumstat(statsOutput)
	if parseErr != nil {
		result.Stats.TotalFiles = len(result.Files)
		result.StatsError = fmt.Sprintf("failed to parse staged numstat: %v", parseErr)
		statsMap = make(map[string]fileStats)
	}
	for i, f := range result.Files {
		if stats, ok := statsMap[f.Path]; ok {
			result.Files[i].Additions = stats.additions
			result.Files[i].Deletions = stats.deletions
			result.Stats.TotalAdditions += stats.additions
			result.Stats.TotalDeletions += stats.deletions
		}
	}
	result.Stats.TotalFiles = len(result.Files)

	return result, nil
}

// GetWorkingTreeDiff returns only unstaged changes (working tree vs index).
func (c *Client) GetWorkingTreeDiff() (*DiffResult, error) {
	result := &DiffResult{
		BaseRef: "index",
		HeadRef: "working-tree",
		Files:   make([]ChangedFile, 0),
	}

	output, err := c.runGit("diff", "--name-status", "-z")
	if err != nil {
		return nil, fmt.Errorf("failed to get working tree diff: %w", err)
	}

	files, err := parseNameStatusOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse working tree name-status: %w", err)
	}
	result.Files = files

	statsOutput, err := c.runGit("diff", "--numstat", "-z")
	if err != nil {
		result.Stats.TotalFiles = len(result.Files)
		result.StatsError = fmt.Sprintf("failed to get working tree numstat: %v", err)
		return result, nil
	}

	statsMap, parseErr := parseNumstat(statsOutput)
	if parseErr != nil {
		result.Stats.TotalFiles = len(result.Files)
		result.StatsError = fmt.Sprintf("failed to parse working tree numstat: %v", parseErr)
		statsMap = make(map[string]fileStats)
	}
	for i, f := range result.Files {
		if stats, ok := statsMap[f.Path]; ok {
			result.Files[i].Additions = stats.additions
			result.Files[i].Deletions = stats.deletions
			result.Stats.TotalAdditions += stats.additions
			result.Stats.TotalDeletions += stats.deletions
		}
	}
	result.Stats.TotalFiles = len(result.Files)

	return result, nil
}

// GetAllChangesDiff returns both staged and unstaged changes combined.
// It delegates to GetDiff("", "") which runs "git diff HEAD" to get the
// combined diff of index+working-tree vs HEAD.
func (c *Client) GetAllChangesDiff() (*DiffResult, error) {
	result, err := c.GetDiff("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get combined changes: %w", err)
	}

	// Set semantic refs for the combined diff
	result.BaseRef = "HEAD"
	result.HeadRef = "working-tree"

	return result, nil
}

// GetDiffStatsForFiles returns numstat data for a list of files against a base ref.
func (c *Client) GetDiffStatsForFiles(baseRef string, files []string) (DiffStats, map[string]FileStats, error) {
	if err := validateRef(baseRef); err != nil {
		return DiffStats{}, nil, err
	}

	if len(files) == 0 {
		return DiffStats{}, map[string]FileStats{}, nil
	}

	args := []string{"diff", "--numstat", "-z", baseRef, "--"}
	args = append(args, files...)

	output, err := c.runGit(args...)
	if err != nil {
		return DiffStats{TotalFiles: len(files)}, nil, fmt.Errorf("failed to get diff numstat: %w", err)
	}

	statsMap, err := parseNumstat(output)
	if err != nil {
		return DiffStats{TotalFiles: len(files)}, nil, fmt.Errorf("failed to parse diff numstat: %w", err)
	}

	result := make(map[string]FileStats, len(statsMap))
	stats := DiffStats{TotalFiles: len(files)}
	for path, fileStat := range statsMap {
		result[path] = FileStats{Additions: fileStat.additions, Deletions: fileStat.deletions}
		stats.TotalAdditions += fileStat.additions
		stats.TotalDeletions += fileStat.deletions
	}

	return stats, result, nil
}

// parseNameStatusOutput parses NUL-delimited output from git diff --name-status -z.
func parseNameStatusOutput(output []byte) ([]ChangedFile, error) {
	if len(output) == 0 {
		return []ChangedFile{}, nil
	}

	tokens := bytes.Split(output, []byte{0})
	for len(tokens) > 0 && len(tokens[len(tokens)-1]) == 0 {
		tokens = tokens[:len(tokens)-1]
	}

	files := make([]ChangedFile, 0)
	for i := 0; i < len(tokens); {
		if len(tokens[i]) == 0 {
			i++
			continue
		}

		statusToken := string(tokens[i])
		i++

		status := ParseFileStatus(statusToken)
		cf := ChangedFile{Status: status}

		if status == StatusRenamed || status == StatusCopied {
			if i+1 >= len(tokens) {
				return nil, fmt.Errorf("invalid rename/copy record: %s", statusToken)
			}
			cf.OldPath = string(tokens[i])
			cf.Path = string(tokens[i+1])
			i += 2
		} else {
			if i >= len(tokens) {
				return nil, fmt.Errorf("invalid name-status record: %s", statusToken)
			}
			cf.Path = string(tokens[i])
			i++
		}

		files = append(files, cf)
	}

	return files, nil
}

type fileStats struct {
	additions int
	deletions int
}

func normalizeNumstatPath(path string) string {
	if !strings.Contains(path, " => ") {
		return path
	}
	if strings.Contains(path, "{") && strings.Contains(path, "}") {
		start := strings.Index(path, "{")
		end := strings.LastIndex(path, "}")
		if start >= 0 && end > start {
			prefix := path[:start]
			suffix := path[end+1:]
			inner := path[start+1 : end]
			parts := strings.Split(inner, " => ")
			if len(parts) == 2 {
				return prefix + strings.TrimSpace(parts[1]) + suffix
			}
		}
	}
	parts := strings.Split(path, " => ")
	if len(parts) == 0 {
		return strings.TrimSpace(path)
	}
	return strings.TrimSpace(parts[len(parts)-1])
}

// parseNumstat parses git diff --numstat -z output (null-separated records).
// Format with -z: "add\tdel\tpath\0" for normal files
// For renames/copies: "add\tdel\t\0oldpath\0newpath\0" (path field empty)
// Binary files show as "-\t-\tpath\0"
func parseNumstat(output []byte) (map[string]fileStats, error) {
	result := make(map[string]fileStats)

	if len(output) == 0 {
		return result, nil
	}

	// Split on null bytes
	tokens := bytes.Split(output, []byte{0})

	// Remove trailing empty token if present (from final null terminator)
	for len(tokens) > 0 && len(tokens[len(tokens)-1]) == 0 {
		tokens = tokens[:len(tokens)-1]
	}

	for i := 0; i < len(tokens); {
		token := string(tokens[i])

		// Skip empty tokens
		if token == "" {
			i++
			continue
		}

		// Each record starts with "add\tdel\tpath" or "add\tdel\t" (for renames)
		parts := strings.SplitN(token, "\t", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid numstat record: expected 3 tab-separated fields, got %d in %q", len(parts), token)
		}

		addStr, delStr, pathPart := parts[0], parts[1], parts[2]

		// Binary files have "-" for additions and deletions - skip them (no stats to record)
		if addStr == "-" && delStr == "-" {
			i++
			continue
		}

		// Parse additions
		additions, err := strconv.Atoi(addStr)
		if err != nil {
			return nil, fmt.Errorf("invalid additions value %q in numstat record %q: %w", addStr, token, err)
		}

		// Parse deletions
		deletions, err := strconv.Atoi(delStr)
		if err != nil {
			return nil, fmt.Errorf("invalid deletions value %q in numstat record %q: %w", delStr, token, err)
		}

		var path string
		if pathPart == "" {
			// Rename/copy: path is empty, next two tokens are oldpath and newpath
			if i+2 >= len(tokens) {
				return nil, fmt.Errorf("invalid rename/copy record: missing paths after %q", token)
			}
			// oldPath := string(tokens[i+1]) // We don't need oldPath for stats
			path = string(tokens[i+2])
			i += 3
		} else {
			path = pathPart
			i++
		}

		path = normalizeNumstatPath(path)
		result[path] = fileStats{
			additions: additions,
			deletions: deletions,
		}
	}

	return result, nil
}
