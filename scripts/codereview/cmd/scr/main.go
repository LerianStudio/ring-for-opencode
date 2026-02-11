// Package main provides the unified scr (static-code-reviewer) CLI binary.
package main

import (
	"os"

	"github.com/lerianstudio/ring/scripts/codereview/internal/recovery"
)

func main() {
	os.Exit(recovery.WrapMain(realMain))
}

func realMain() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
