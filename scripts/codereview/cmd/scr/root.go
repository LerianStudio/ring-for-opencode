package main

import (
	"os"
	"strings"

	"github.com/lerianstudio/ring/scripts/codereview/internal/cli"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	"github.com/spf13/cobra"
)

var (
	// Version info set via ldflags
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"

	// Global config shared across commands
	globalCfg = cli.DefaultConfig()
)

// rootCmd is the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "scr",
	Short: "Static Code Reviewer - Multi-phase code analysis tool",
	Long: `scr (Static Code Reviewer) is a multi-phase static code analysis tool
that prepares comprehensive context for code reviews.

It executes 6 analysis phases:
  1. scope    - Detect changed files and project language
  2. static   - Run linters and static analyzers
  3. ast      - Extract semantic AST diffs
  4. callgraph - Analyze call relationships
  5. dataflow - Security-focused data flow analysis
  6. context  - Compile reviewer-specific context files

Output artifacts are written to .scr/ (configurable via --output-dir).`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if globalCfg.Verbose {
			logger.SetDefault(logger.NewLogger(logger.WithLevel(logger.LevelDebug)))
		}
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().StringVarP(&globalCfg.OutputDir, "output-dir", "o", ".scr", "Output directory for all artifacts")
	rootCmd.PersistentFlags().BoolVarP(&globalCfg.Verbose, "verbose", "v", false, "Enable verbose output")

	// Git ref flags (used by run and scope)
	rootCmd.PersistentFlags().StringVar(&globalCfg.BaseRef, "base", "", "Base git reference (commit/branch)")
	rootCmd.PersistentFlags().StringVar(&globalCfg.HeadRef, "head", "", "Head git reference (commit/branch)")
	rootCmd.PersistentFlags().BoolVar(&globalCfg.Unstaged, "unstaged", false, "Analyze only unstaged and untracked files")

	// File pattern flags
	rootCmd.PersistentFlags().StringVar(&globalCfg.Files, "files", "", "Comma-separated file patterns to analyze")
	rootCmd.PersistentFlags().StringVar(&globalCfg.FilesFrom, "files-from", "", "Path to file containing file patterns (one per line)")

	// Register subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(phaseCmd)
	rootCmd.AddCommand(versionCmd)
}

// parseSkipList parses a comma-separated skip list.
func parseSkipList(skip string) []string {
	if skip == "" {
		return nil
	}
	var result []string
	for _, s := range strings.Split(skip, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// getWorkDir returns the working directory.
func getWorkDir() string {
	if globalCfg.WorkDir != "" {
		return globalCfg.WorkDir
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
