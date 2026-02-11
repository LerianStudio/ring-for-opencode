// Package main implements the data-flow binary for security-focused data flow analysis.
// DEPRECATED: Use 'scr phase dataflow' instead.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	dataflowphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/dataflow"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/recovery"
)

var (
	scopePath = flag.String("scope", "scope.json", "Path to scope.json from Phase 0")
	outputDir = flag.String("output", ".", "Output directory for results")
	scriptDir = flag.String("scripts", "", "Path to scripts/codereview directory (auto-detect from executable path if not provided)")
	language  = flag.String("lang", "", "Analyze specific language only (go, python, typescript)")
	jsonOnly  = flag.Bool("json", false, "Output JSON only, no markdown summary")
	verbose   = flag.Bool("v", false, "Verbose output")
)

func main() {
	os.Exit(recovery.WrapMain(realMain))
}

func realMain() {
	fmt.Fprintln(os.Stderr, "DEPRECATED: data-flow is deprecated, use 'scr phase dataflow' instead")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Data Flow Analyzer - Security-focused data flow analysis\n\n")
		fmt.Fprintf(os.Stderr, "DEPRECATED: Use 'scr phase dataflow' instead.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *verbose {
		logger.SetDefault(logger.NewLogger(logger.WithLevel(logger.LevelDebug)))
	}

	if err := run(); err != nil {
		logger.Error("error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	scriptsDir := *scriptDir
	if scriptsDir == "" {
		execPath, err := os.Executable()
		if err == nil {
			scriptsDir = filepath.Dir(filepath.Dir(execPath))
		}
		if scriptsDir == "" {
			scriptsDir = "."
		}
	}

	cfg := &phases.Config{
		WorkDir:    workDir,
		OutputDir:  *outputDir,
		Verbose:    *verbose,
		ScopePath:  *scopePath,
		ScriptsDir: scriptsDir,
		Language:   *language,
	}

	ctx := context.Background()
	phase := dataflowphase.New()

	if err := phase.Run(ctx, cfg); err != nil {
		return err
	}

	// For --json mode, read generated JSON files and output combined JSON to stdout
	if *jsonOnly {
		return outputJSONOnly(*outputDir)
	}

	return nil
}

// outputJSONOnly reads the generated language-specific JSON files and outputs combined JSON to stdout.
func outputJSONOnly(outputDir string) error {
	languages := []string{"go", "python", "typescript"}
	results := make(map[string]json.RawMessage)

	for _, lang := range languages {
		jsonPath := filepath.Join(outputDir, fmt.Sprintf("%s-flow.json", lang))
		data, err := fileutil.ReadJSONFileWithLimit(jsonPath)
		if err != nil {
			// File may not exist if no files for this language
			continue
		}
		results[lang] = data
	}

	output := struct {
		Languages map[string]json.RawMessage `json:"languages"`
	}{
		Languages: results,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}
