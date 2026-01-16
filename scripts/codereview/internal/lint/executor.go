package lint

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout is the default timeout for linter execution.
const DefaultTimeout = 5 * time.Minute

// ExecResult holds the result of command execution.
type ExecResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Err      error
}

// Executor runs external commands.
type Executor struct {
	timeout time.Duration
}

// NewExecutor creates a new command executor.
func NewExecutor() *Executor {
	return &Executor{
		timeout: DefaultTimeout,
	}
}

// WithTimeout sets a custom timeout.
func (e *Executor) WithTimeout(d time.Duration) *Executor {
	copy := *e
	copy.timeout = d
	return &copy
}

// Run executes a command and returns the result.
func (e *Executor) Run(ctx context.Context, dir string, name string, args ...string) *ExecResult {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ExecResult{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			// Many linters return non-zero on findings, which is not an error
			result.Err = nil
		} else if ctx.Err() == context.DeadlineExceeded {
			result.Err = fmt.Errorf("command timed out after %v", e.timeout)
		} else {
			result.Err = err
		}
	}

	return result
}

// CommandAvailable checks if a command is available in PATH.
func (e *Executor) CommandAvailable(ctx context.Context, name string) bool {
	resultCh := make(chan error, 1)

	go func() {
		_, err := exec.LookPath(name)
		resultCh <- err
	}()

	select {
	case <-ctx.Done():
		return false
	case err := <-resultCh:
		return err == nil
	}
}

// GetVersion runs a command with --version and extracts the version string.
func (e *Executor) GetVersion(ctx context.Context, name string, args ...string) (string, error) {
	if len(args) == 0 {
		args = []string{"--version"}
	}

	result := e.Run(ctx, "", name, args...)
	if result.Err != nil {
		return "", result.Err
	}

	output := string(result.Stdout)
	if output == "" {
		output = string(result.Stderr)
	}

	// Extract first line and clean up
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "unknown", nil
}
