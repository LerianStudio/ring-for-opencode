package callgraph

import (
	"os"
)

// fileExists checks if a file exists and is not a directory.
// This is a shared utility used by both Python and TypeScript analyzers.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
