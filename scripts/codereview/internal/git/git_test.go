package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\noutput: %s", strings.Join(args, " "), err, output)
	}

	return strings.TrimSpace(string(output))
}

func writeFile(t *testing.T, dir, name, contents string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

func containsArg(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}

func nulJoin(parts ...string) []byte {
	return []byte(strings.Join(parts, "\x00") + "\x00")
}

func requireGit(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
}

func setupTestRepo(t *testing.T) string {
	t.Helper()

	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	writeFile(t, dir, "README.md", "base\n")
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")

	return dir
}

func TestFileStatusString(t *testing.T) {
	tests := []struct {
		name     string
		status   FileStatus
		expected string
	}{
		{"Added", StatusAdded, "A"},
		{"Modified", StatusModified, "M"},
		{"Deleted", StatusDeleted, "D"},
		{"Renamed", StatusRenamed, "R"},
		{"Copied", StatusCopied, "C"},
		{"Unknown", StatusUnknown, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("FileStatus.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseFileStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FileStatus
	}{
		{"Added", "A", StatusAdded},
		{"Modified", "M", StatusModified},
		{"Deleted", "D", StatusDeleted},
		{"Renamed", "R100", StatusRenamed},
		{"Renamed partial", "R075", StatusRenamed},
		{"Copied", "C", StatusCopied},
		{"Unknown", "X", StatusUnknown},
		{"Empty", "", StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseFileStatus(tt.input); got != tt.expected {
				t.Errorf("ParseFileStatus(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestClientGetDiff(t *testing.T) {
	repo := setupTestRepo(t)
	writeFile(t, repo, "README.md", "base\nupdated\n")
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-m", "update")

	client := NewClient(repo)
	result, err := client.GetDiff("HEAD~1", "HEAD")
	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}

	if result.BaseRef != "HEAD~1" {
		t.Errorf("BaseRef = %q, want %q", result.BaseRef, "HEAD~1")
	}
	if result.HeadRef != "HEAD" {
		t.Errorf("HeadRef = %q, want %q", result.HeadRef, "HEAD")
	}
	if len(result.Files) != 1 {
		t.Fatalf("Files length = %d, want %d", len(result.Files), 1)
	}
	if result.Files[0].Path != "README.md" {
		t.Errorf("Path = %q, want %q", result.Files[0].Path, "README.md")
	}
	if result.Files[0].Additions != 1 {
		t.Errorf("Additions = %d, want %d", result.Files[0].Additions, 1)
	}
	if result.Files[0].Deletions != 0 {
		t.Errorf("Deletions = %d, want %d", result.Files[0].Deletions, 0)
	}
	if result.Stats.TotalAdditions != 1 {
		t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 1)
	}
	if result.Stats.TotalDeletions != 0 {
		t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 0)
	}
	if result.Stats.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 1)
	}
	if result.StatsError != "" {
		t.Errorf("StatsError = %q, want empty", result.StatsError)
	}
}

func TestClientGetDiffInvalidRef(t *testing.T) {
	repo := setupTestRepo(t)
	client := NewClient(repo)

	if _, err := client.GetDiff("missing-ref", "HEAD"); err == nil {
		t.Fatal("expected error for missing ref")
	} else if !strings.Contains(err.Error(), "failed to get diff name-status") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientGetDiffRejectsOptionRef(t *testing.T) {
	repo := setupTestRepo(t)
	client := NewClient(repo)

	if _, err := client.GetDiff("-bad", "HEAD"); err == nil {
		t.Fatal("expected error for option-like ref")
	} else if !strings.Contains(err.Error(), "invalid ref") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientGetDiffRunGitFailure(t *testing.T) {
	client := &Client{runner: func(_ string, args ...string) ([]byte, error) {
		return nil, errors.New("runGit failed")
	}}

	if _, err := client.GetDiff("HEAD", "HEAD~1"); err == nil {
		t.Fatal("expected error when runGit fails")
	} else if !strings.Contains(err.Error(), "failed to get diff name-status") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientGetDiffNumstatFailure(t *testing.T) {
	client := &Client{runner: func(_ string, args ...string) ([]byte, error) {
		if containsArg(args, "--name-status") {
			return nulJoin("M", "file.txt"), nil
		}
		if containsArg(args, "--numstat") {
			return nil, errors.New("numstat failure")
		}
		return nil, errors.New("unexpected args")
	}}

	result, err := client.GetDiff("HEAD", "HEAD~1")
	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("Files length = %d, want %d", len(result.Files), 1)
	}
	if result.Stats.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 1)
	}
	if result.StatsError == "" {
		t.Fatal("expected StatsError to be set")
	}
	if !strings.Contains(result.StatsError, "numstat") {
		t.Fatalf("StatsError = %q, want numstat context", result.StatsError)
	}
	if result.Files[0].Additions != 0 || result.Files[0].Deletions != 0 {
		t.Fatalf("expected file stats to remain zero on numstat failure")
	}
	if result.Stats.TotalAdditions != 0 {
		t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 0)
	}
	if result.Stats.TotalDeletions != 0 {
		t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 0)
	}
}

func TestClientGetDiffNameStatusParseFailure(t *testing.T) {
	client := &Client{runner: func(_ string, args ...string) ([]byte, error) {
		if containsArg(args, "--name-status") {
			return nulJoin("M"), nil
		}
		if containsArg(args, "--numstat") {
			return []byte("1\t0\tfile.txt\x00"), nil
		}
		return nil, errors.New("unexpected args")
	}}

	if _, err := client.GetDiff("HEAD", "HEAD~1"); err == nil {
		t.Fatal("expected error for malformed name-status output")
	} else if !strings.Contains(err.Error(), "failed to parse diff name-status") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientGetDiffBinaryNumstat(t *testing.T) {
	client := &Client{runner: func(_ string, args ...string) ([]byte, error) {
		if containsArg(args, "--name-status") {
			return nulJoin("A", "image.png"), nil
		}
		if containsArg(args, "--numstat") {
			return []byte("-\t-\timage.png\x00"), nil
		}
		return nil, errors.New("unexpected args")
	}}

	result, err := client.GetDiff("HEAD", "HEAD~1")
	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("Files length = %d, want %d", len(result.Files), 1)
	}
	if result.Files[0].Additions != 0 {
		t.Errorf("Additions = %d, want %d", result.Files[0].Additions, 0)
	}
	if result.Files[0].Deletions != 0 {
		t.Errorf("Deletions = %d, want %d", result.Files[0].Deletions, 0)
	}
	if result.Stats.TotalAdditions != 0 {
		t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 0)
	}
	if result.Stats.TotalDeletions != 0 {
		t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 0)
	}
	if result.Stats.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 1)
	}
	if result.StatsError != "" {
		t.Errorf("StatsError = %q, want empty", result.StatsError)
	}
}

func TestClientGetDiffMixedNumstat(t *testing.T) {
	client := &Client{runner: func(_ string, args ...string) ([]byte, error) {
		if containsArg(args, "--name-status") {
			return nulJoin("A", "binary.png", "M", "file.txt"), nil
		}
		if containsArg(args, "--numstat") {
			return []byte("-\t-\tbinary.png\x002\t1\tfile.txt\x00"), nil
		}
		return nil, errors.New("unexpected args")
	}}

	result, err := client.GetDiff("HEAD", "HEAD~1")
	if err != nil {
		t.Fatalf("GetDiff() error = %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("Files length = %d, want %d", len(result.Files), 2)
	}

	filesByPath := make(map[string]ChangedFile, len(result.Files))
	for _, file := range result.Files {
		filesByPath[file.Path] = file
	}

	binary := filesByPath["binary.png"]
	if binary.Additions != 0 || binary.Deletions != 0 {
		t.Fatalf("binary file stats should be zero")
	}
	text := filesByPath["file.txt"]
	if text.Additions != 2 || text.Deletions != 1 {
		t.Fatalf("text stats = %d/%d, want 2/1", text.Additions, text.Deletions)
	}
	if result.Stats.TotalAdditions != 2 {
		t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 2)
	}
	if result.Stats.TotalDeletions != 1 {
		t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 1)
	}
	if result.StatsError != "" {
		t.Errorf("StatsError = %q, want empty", result.StatsError)
	}
}

func TestClientGetDiffRefCombos(t *testing.T) {
	t.Run("Empty refs", func(t *testing.T) {
		repo := setupTestRepo(t)
		writeFile(t, repo, "staged.txt", "staged\n")
		runGit(t, repo, "add", "staged.txt")
		writeFile(t, repo, "README.md", "base\nworking\n")

		client := NewClient(repo)
		result, err := client.GetDiff("", "")
		if err != nil {
			t.Fatalf("GetDiff() error = %v", err)
		}

		if result.BaseRef != "" {
			t.Errorf("BaseRef = %q, want empty", result.BaseRef)
		}
		if result.HeadRef != "" {
			t.Errorf("HeadRef = %q, want empty", result.HeadRef)
		}
		if len(result.Files) != 2 {
			t.Fatalf("Files length = %d, want %d", len(result.Files), 2)
		}

		filesByPath := make(map[string]ChangedFile, len(result.Files))
		for _, file := range result.Files {
			filesByPath[file.Path] = file
		}
		for _, path := range []string{"README.md", "staged.txt"} {
			file, ok := filesByPath[path]
			if !ok {
				t.Fatalf("missing file %q", path)
			}
			if file.Additions != 1 {
				t.Errorf("Additions for %s = %d, want %d", path, file.Additions, 1)
			}
			if file.Deletions != 0 {
				t.Errorf("Deletions for %s = %d, want %d", path, file.Deletions, 0)
			}
		}
		if result.Stats.TotalAdditions != 2 {
			t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 2)
		}
		if result.Stats.TotalDeletions != 0 {
			t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 0)
		}
		if result.Stats.TotalFiles != 2 {
			t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 2)
		}
	})

	t.Run("HeadRef empty", func(t *testing.T) {
		repo := setupTestRepo(t)
		writeFile(t, repo, "staged.txt", "staged\n")
		runGit(t, repo, "add", "staged.txt")
		writeFile(t, repo, "README.md", "base\nworking\n")

		client := NewClient(repo)
		result, err := client.GetDiff("HEAD", "")
		if err != nil {
			t.Fatalf("GetDiff() error = %v", err)
		}

		if result.BaseRef != "HEAD" {
			t.Errorf("BaseRef = %q, want %q", result.BaseRef, "HEAD")
		}
		if result.HeadRef != "" {
			t.Errorf("HeadRef = %q, want empty", result.HeadRef)
		}
		if len(result.Files) != 2 {
			t.Fatalf("Files length = %d, want %d", len(result.Files), 2)
		}

		filesByPath := make(map[string]ChangedFile, len(result.Files))
		for _, file := range result.Files {
			filesByPath[file.Path] = file
		}
		for _, path := range []string{"README.md", "staged.txt"} {
			file, ok := filesByPath[path]
			if !ok {
				t.Fatalf("missing file %q", path)
			}
			if file.Additions != 1 {
				t.Errorf("Additions for %s = %d, want %d", path, file.Additions, 1)
			}
			if file.Deletions != 0 {
				t.Errorf("Deletions for %s = %d, want %d", path, file.Deletions, 0)
			}
		}
		if result.Stats.TotalAdditions != 2 {
			t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 2)
		}
		if result.Stats.TotalDeletions != 0 {
			t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 0)
		}
		if result.Stats.TotalFiles != 2 {
			t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 2)
		}
	})

	t.Run("BaseRef empty", func(t *testing.T) {
		repo := setupTestRepo(t)
		writeFile(t, repo, "README.md", "base\nupdated\n")
		runGit(t, repo, "add", "README.md")
		runGit(t, repo, "commit", "-m", "update")

		client := NewClient(repo)
		result, err := client.GetDiff("", "HEAD~1")
		if err != nil {
			t.Fatalf("GetDiff() error = %v", err)
		}

		if result.BaseRef != "" {
			t.Errorf("BaseRef = %q, want empty", result.BaseRef)
		}
		if result.HeadRef != "HEAD~1" {
			t.Errorf("HeadRef = %q, want %q", result.HeadRef, "HEAD~1")
		}
		if len(result.Files) != 1 {
			t.Fatalf("Files length = %d, want %d", len(result.Files), 1)
		}
		if result.Files[0].Path != "README.md" {
			t.Errorf("Path = %q, want %q", result.Files[0].Path, "README.md")
		}
		if result.Files[0].Additions != 0 {
			t.Errorf("Additions = %d, want %d", result.Files[0].Additions, 0)
		}
		if result.Files[0].Deletions != 1 {
			t.Errorf("Deletions = %d, want %d", result.Files[0].Deletions, 1)
		}
		if result.Stats.TotalAdditions != 0 {
			t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 0)
		}
		if result.Stats.TotalDeletions != 1 {
			t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 1)
		}
		if result.Stats.TotalFiles != 1 {
			t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 1)
		}
	})

	t.Run("HeadRef branch", func(t *testing.T) {
		repo := setupTestRepo(t)
		writeFile(t, repo, "README.md", "base\nupdated\n")
		runGit(t, repo, "add", "README.md")
		runGit(t, repo, "commit", "-m", "update")
		runGit(t, repo, "branch", "feature/test", "HEAD~1")

		client := NewClient(repo)
		result, err := client.GetDiff("", "feature/test")
		if err != nil {
			t.Fatalf("GetDiff() error = %v", err)
		}

		if result.HeadRef != "feature/test" {
			t.Errorf("HeadRef = %q, want %q", result.HeadRef, "feature/test")
		}
		if len(result.Files) != 1 {
			t.Fatalf("Files length = %d, want %d", len(result.Files), 1)
		}
		if result.Files[0].Additions != 0 {
			t.Errorf("Additions = %d, want %d", result.Files[0].Additions, 0)
		}
		if result.Files[0].Deletions != 1 {
			t.Errorf("Deletions = %d, want %d", result.Files[0].Deletions, 1)
		}
		if result.Stats.TotalDeletions != 1 {
			t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 1)
		}
	})

	t.Run("BaseRef branch", func(t *testing.T) {
		repo := setupTestRepo(t)
		writeFile(t, repo, "README.md", "base\nupdated\n")
		runGit(t, repo, "add", "README.md")
		runGit(t, repo, "commit", "-m", "update")
		runGit(t, repo, "branch", "feature/test", "HEAD~1")

		client := NewClient(repo)
		result, err := client.GetDiff("feature/test", "HEAD")
		if err != nil {
			t.Fatalf("GetDiff() error = %v", err)
		}

		if result.BaseRef != "feature/test" {
			t.Errorf("BaseRef = %q, want %q", result.BaseRef, "feature/test")
		}
		if result.HeadRef != "HEAD" {
			t.Errorf("HeadRef = %q, want %q", result.HeadRef, "HEAD")
		}
		if len(result.Files) != 1 {
			t.Fatalf("Files length = %d, want %d", len(result.Files), 1)
		}
		if result.Files[0].Additions != 1 {
			t.Errorf("Additions = %d, want %d", result.Files[0].Additions, 1)
		}
		if result.Files[0].Deletions != 0 {
			t.Errorf("Deletions = %d, want %d", result.Files[0].Deletions, 0)
		}
		if result.Stats.TotalAdditions != 1 {
			t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 1)
		}
	})
}

func TestClientGetStagedDiff(t *testing.T) {
	repo := setupTestRepo(t)
	writeFile(t, repo, "staged.txt", "staged\n")
	runGit(t, repo, "add", "staged.txt")

	client := NewClient(repo)
	result, err := client.GetStagedDiff()
	if err != nil {
		t.Fatalf("GetStagedDiff() error = %v", err)
	}

	if result.BaseRef != "HEAD" {
		t.Errorf("BaseRef = %q, want %q", result.BaseRef, "HEAD")
	}
	if result.HeadRef != "staged" {
		t.Errorf("HeadRef = %q, want %q", result.HeadRef, "staged")
	}
	if len(result.Files) != 1 {
		t.Fatalf("Files length = %d, want %d", len(result.Files), 1)
	}
	if result.Files[0].Path != "staged.txt" {
		t.Errorf("Path = %q, want %q", result.Files[0].Path, "staged.txt")
	}
	if result.Files[0].Additions != 1 {
		t.Errorf("Additions = %d, want %d", result.Files[0].Additions, 1)
	}
	if result.Files[0].Deletions != 0 {
		t.Errorf("Deletions = %d, want %d", result.Files[0].Deletions, 0)
	}
	if result.Stats.TotalAdditions != 1 {
		t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 1)
	}
	if result.Stats.TotalDeletions != 0 {
		t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 0)
	}
	if result.Stats.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 1)
	}
	if result.StatsError != "" {
		t.Errorf("StatsError = %q, want empty", result.StatsError)
	}
}

func TestClientGetWorkingTreeDiff(t *testing.T) {
	repo := setupTestRepo(t)
	writeFile(t, repo, "README.md", "base\nline\n")
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-m", "add line")
	writeFile(t, repo, "README.md", "base\nupdated\n")

	client := NewClient(repo)
	result, err := client.GetWorkingTreeDiff()
	if err != nil {
		t.Fatalf("GetWorkingTreeDiff() error = %v", err)
	}

	if result.BaseRef != "index" {
		t.Errorf("BaseRef = %q, want %q", result.BaseRef, "index")
	}
	if result.HeadRef != "working-tree" {
		t.Errorf("HeadRef = %q, want %q", result.HeadRef, "working-tree")
	}
	if len(result.Files) != 1 {
		t.Fatalf("Files length = %d, want %d", len(result.Files), 1)
	}
	if result.Files[0].Path != "README.md" {
		t.Errorf("Path = %q, want %q", result.Files[0].Path, "README.md")
	}
	if result.Files[0].Additions != 1 {
		t.Errorf("Additions = %d, want %d", result.Files[0].Additions, 1)
	}
	if result.Files[0].Deletions != 1 {
		t.Errorf("Deletions = %d, want %d", result.Files[0].Deletions, 1)
	}
	if result.Stats.TotalAdditions != 1 {
		t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 1)
	}
	if result.Stats.TotalDeletions != 1 {
		t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 1)
	}
	if result.Stats.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 1)
	}
	if result.StatsError != "" {
		t.Errorf("StatsError = %q, want empty", result.StatsError)
	}
}

func TestClientGetAllChangesDiff(t *testing.T) {
	repo := setupTestRepo(t)
	writeFile(t, repo, "extra.txt", "one\n")
	runGit(t, repo, "add", "extra.txt")
	runGit(t, repo, "commit", "-m", "add extra")
	writeFile(t, repo, "README.md", "base\nstaged\n")
	runGit(t, repo, "add", "README.md")
	writeFile(t, repo, "README.md", "base\nstaged\nunstaged\n")
	writeFile(t, repo, "extra.txt", "one\nupdated\n")

	client := NewClient(repo)
	result, err := client.GetAllChangesDiff()
	if err != nil {
		t.Fatalf("GetAllChangesDiff() error = %v", err)
	}

	if result.BaseRef != "HEAD" {
		t.Errorf("BaseRef = %q, want %q", result.BaseRef, "HEAD")
	}
	if result.HeadRef != "working-tree" {
		t.Errorf("HeadRef = %q, want %q", result.HeadRef, "working-tree")
	}
	if len(result.Files) != 2 {
		t.Fatalf("Files length = %d, want %d", len(result.Files), 2)
	}

	filesByPath := make(map[string]ChangedFile, len(result.Files))
	for _, file := range result.Files {
		filesByPath[file.Path] = file
	}

	readme, ok := filesByPath["README.md"]
	if !ok {
		t.Fatalf("missing README.md")
	}
	// README.md: HEAD="base\n" → working tree="base\nstaged\nunstaged\n"
	// Combined diff shows +2 additions (staged, unstaged lines)
	if readme.Additions != 2 {
		t.Errorf("README.md additions = %d, want %d", readme.Additions, 2)
	}
	if readme.Deletions != 0 {
		t.Errorf("README.md deletions = %d, want %d", readme.Deletions, 0)
	}

	extra, ok := filesByPath["extra.txt"]
	if !ok {
		t.Fatalf("missing extra.txt")
	}
	if extra.Additions != 1 {
		t.Errorf("extra.txt additions = %d, want %d", extra.Additions, 1)
	}
	if extra.Deletions != 0 {
		t.Errorf("extra.txt deletions = %d, want %d", extra.Deletions, 0)
	}

	// Total: README.md (+2) + extra.txt (+1) = 3 additions
	if result.Stats.TotalAdditions != 3 {
		t.Errorf("TotalAdditions = %d, want %d", result.Stats.TotalAdditions, 3)
	}
	if result.Stats.TotalDeletions != 0 {
		t.Errorf("TotalDeletions = %d, want %d", result.Stats.TotalDeletions, 0)
	}
	if result.Stats.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want %d", result.Stats.TotalFiles, 2)
	}
	if result.StatsError != "" {
		t.Errorf("StatsError = %q, want empty", result.StatsError)
	}
}

func TestParseNameStatusOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected ChangedFile
		wantErr  bool
	}{
		{
			name:  "Modified file",
			input: nulJoin("M", "internal/handler/user.go"),
			expected: ChangedFile{
				Path:   "internal/handler/user.go",
				Status: StatusModified,
			},
		},
		{
			name:  "Added file",
			input: nulJoin("A", "new/file.go"),
			expected: ChangedFile{
				Path:   "new/file.go",
				Status: StatusAdded,
			},
		},
		{
			name:  "Deleted file",
			input: nulJoin("D", "old/file.go"),
			expected: ChangedFile{
				Path:   "old/file.go",
				Status: StatusDeleted,
			},
		},
		{
			name:  "Renamed file",
			input: nulJoin("R100", "old/path.go", "new/path.go"),
			expected: ChangedFile{
				Path:    "new/path.go",
				OldPath: "old/path.go",
				Status:  StatusRenamed,
			},
		},
		{
			name:  "Copied file",
			input: nulJoin("C100", "original.go", "copy.go"),
			expected: ChangedFile{
				Path:    "copy.go",
				OldPath: "original.go",
				Status:  StatusCopied,
			},
		},
		{
			name:    "Invalid rename missing new path",
			input:   nulJoin("R100", "old/path.go"),
			wantErr: true,
		},
		{
			name:    "Invalid copy missing new path",
			input:   nulJoin("C100", "old/path.go"),
			wantErr: true,
		},
		{
			name:    "Invalid record missing path",
			input:   nulJoin("M"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := parseNameStatusOutput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseNameStatusOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(files) != 1 {
				t.Fatalf("files length = %d, want %d", len(files), 1)
			}
			got := files[0]
			if got.Path != tt.expected.Path {
				t.Errorf("Path = %q, want %q", got.Path, tt.expected.Path)
			}
			if got.OldPath != tt.expected.OldPath {
				t.Errorf("OldPath = %q, want %q", got.OldPath, tt.expected.OldPath)
			}
			if got.Status != tt.expected.Status {
				t.Errorf("Status = %v, want %v", got.Status, tt.expected.Status)
			}
		})
	}

	multi := nulJoin("M", "one.go", "A", "two.go")
	files, err := parseNameStatusOutput(multi)
	if err != nil {
		t.Fatalf("parseNameStatusOutput() unexpected error for multi-record: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("multi-record files length = %d, want %d", len(files), 2)
	}
	if files[0].Path != "one.go" || files[1].Path != "two.go" {
		t.Fatalf("multi-record paths = %q, %q", files[0].Path, files[1].Path)
	}
	if files[0].Status != StatusModified || files[1].Status != StatusAdded {
		t.Fatalf("multi-record statuses = %v, %v", files[0].Status, files[1].Status)
	}
}

func TestParseNameStatusOutputEmpty(t *testing.T) {
	files, err := parseNameStatusOutput(nil)
	if err != nil {
		t.Fatalf("parseNameStatusOutput() error = %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("files length = %d, want %d", len(files), 0)
	}
}

func TestParseNumstat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]fileStats
		wantErr  bool
	}{
		{
			name:  "Single file",
			input: "10\t5\tpath/to/file.go\x00",
			expected: map[string]fileStats{
				"path/to/file.go": {additions: 10, deletions: 5},
			},
		},
		{
			name:  "Multiple files",
			input: "10\t5\tfile1.go\x0020\t3\tfile2.go\x00",
			expected: map[string]fileStats{
				"file1.go": {additions: 10, deletions: 5},
				"file2.go": {additions: 20, deletions: 3},
			},
		},
		{
			name:     "Binary file (skip)",
			input:    "-\t-\timage.png\x00",
			expected: map[string]fileStats{},
		},
		{
			name:  "Binary and text",
			input: "-\t-\timage.png\x001\t0\ttext.go\x00",
			expected: map[string]fileStats{
				"text.go": {additions: 1, deletions: 0},
			},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: map[string]fileStats{},
		},
		{
			name:  "File with spaces",
			input: "10\t5\tpath/to/file with spaces.go\x00",
			expected: map[string]fileStats{
				"path/to/file with spaces.go": {additions: 10, deletions: 5},
			},
		},
		{
			name:  "Rename path with arrow notation",
			input: "5\t1\told/path.go => new/path.go\x00",
			expected: map[string]fileStats{
				"new/path.go": {additions: 5, deletions: 1},
			},
		},
		{
			name:  "Rename with braces",
			input: "1\t0\tsrc/{old => new}/file.go\x00",
			expected: map[string]fileStats{
				"src/new/file.go": {additions: 1, deletions: 0},
			},
		},
		{
			name:  "Rename with -z format (empty path, oldpath, newpath)",
			input: "5\t1\t\x00old/path.go\x00new/path.go\x00",
			expected: map[string]fileStats{
				"new/path.go": {additions: 5, deletions: 1},
			},
		},
		{
			name:    "Invalid additions value",
			input:   "x\t5\tbad.go\x00",
			wantErr: true,
		},
		{
			name:    "Invalid deletions value",
			input:   "10\ty\talso-bad.go\x00",
			wantErr: true,
		},
		{
			name:    "Too few fields",
			input:   "10\t5\x00",
			wantErr: true,
		},
		{
			name:    "Rename missing newpath",
			input:   "5\t1\t\x00old/path.go\x00",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNumstat([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseNumstat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("parseNumstat() returned %d entries, want %d", len(got), len(tt.expected))
			}
			for path, expectedStats := range tt.expected {
				if gotStats, ok := got[path]; !ok {
					t.Errorf("parseNumstat() missing path %q", path)
				} else if gotStats != expectedStats {
					t.Errorf("parseNumstat()[%q] = %+v, want %+v", path, gotStats, expectedStats)
				}
			}
		})
	}
}

func TestParseNumstatLongPath(t *testing.T) {
	// With null-separated parsing, very long paths are handled correctly
	longPath := strings.Repeat("a", 70000)
	result, err := parseNumstat([]byte("1\t0\t" + longPath + "\x00"))
	if err != nil {
		t.Fatalf("parseNumstat() error = %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("parseNumstat() returned %d entries, want 1", len(result))
	}
	if stats, ok := result[longPath]; !ok {
		t.Error("parseNumstat() missing long path")
	} else if stats.additions != 1 || stats.deletions != 0 {
		t.Errorf("parseNumstat() stats = %+v, want {additions: 1, deletions: 0}", stats)
	}
}

func TestParseNumstatMalformedRecord(t *testing.T) {
	// Test that malformed records return errors
	tests := []struct {
		name  string
		input string
	}{
		{"single field", "invalid\x00"},
		{"two fields", "1\t2\x00"},
		{"non-numeric additions", "abc\t0\tfile.go\x00"},
		{"non-numeric deletions", "0\txyz\tfile.go\x00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseNumstat([]byte(tt.input))
			if err == nil {
				t.Error("expected error for malformed numstat record")
			}
		})
	}
}
