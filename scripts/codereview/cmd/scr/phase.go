package main

import (
	"github.com/spf13/cobra"
)

var phaseCmd = &cobra.Command{
	Use:   "phase",
	Short: "Run individual analysis phases",
	Long: `Run individual analysis phases.

Available phases:
  scope    - Phase 0: Detect changed files and project language
  static   - Phase 1: Run linters and static analyzers
  ast      - Phase 2: Extract semantic AST diffs
  callgraph - Phase 3: Analyze call relationships
  dataflow - Phase 4: Security-focused data flow analysis
  context  - Phase 5: Compile reviewer-specific context files

Use "scr phase [phase] --help" for more information about a specific phase.`,
}

func init() {
	// Add phase subcommands
	phaseCmd.AddCommand(phaseScopeCmd)
	phaseCmd.AddCommand(phaseStaticCmd)
	phaseCmd.AddCommand(phaseASTCmd)
	phaseCmd.AddCommand(phaseCallgraphCmd)
	phaseCmd.AddCommand(phaseDataflowCmd)
	phaseCmd.AddCommand(phaseContextCmd)
}
