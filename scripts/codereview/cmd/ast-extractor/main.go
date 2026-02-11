package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lerianstudio/ring/scripts/codereview/internal/ast"
	"github.com/lerianstudio/ring/scripts/codereview/internal/fileutil"
	"github.com/lerianstudio/ring/scripts/codereview/internal/logger"
	astphase "github.com/lerianstudio/ring/scripts/codereview/internal/phases/ast"
	"github.com/lerianstudio/ring/scripts/codereview/internal/phases"
	"github.com/lerianstudio/ring/scripts/codereview/internal/recovery"
)

var (
	beforeFile = flag.String("before", "", "Path to the before version of the file")
	afterFile  = flag.String("after", "", "Path to the after version of the file")
	language   = flag.String("lang", "", "Force language (go, typescript, python)")
	outputFmt  = flag.String("output", "json", "Output format: json or markdown")
	scriptDir  = flag.String("scripts", "", "Directory containing language scripts (ts/, py/)")
	timeout    = flag.Duration("timeout", 30*time.Second, "Extraction timeout")
	batchFile  = flag.String("batch", "", "JSON file with batch of file pairs to process")
	verbose    = flag.Bool("v", false, "Enable verbose output")
)

func init() {
	flag.BoolVar(verbose, "verbose", false, "Enable verbose output")
}

func main() {
	os.Exit(recovery.WrapMain(realMain))
}

func realMain() {
	fmt.Fprintln(os.Stderr, "DEPRECATED: ast-extractor is deprecated, use 'scr phase ast' instead")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "AST Extractor - Extract semantic diffs from source files\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Compare two Go files:\n")
		fmt.Fprintf(os.Stderr, "  %s -before old.go -after new.go\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # New file (no before version):\n")
		fmt.Fprintf(os.Stderr, "  %s -after new.go\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Deleted file (no after version):\n")
		fmt.Fprintf(os.Stderr, "  %s -before old.go\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Batch mode with JSON input:\n")
		fmt.Fprintf(os.Stderr, "  %s -batch files.json -output markdown\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Batch file format:\n")
		fmt.Fprintf(os.Stderr, `  [{"before_path": "old.go", "after_path": "new.go"}]`+"\n")
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

	scriptsPath := *scriptDir
	if scriptsPath == "" {
		exe, err := os.Executable()
		if err == nil {
			scriptsPath = filepath.Join(filepath.Dir(exe), "..", "..")
		} else {
			scriptsPath = "."
		}
	}

	// Batch mode: delegate to phase library
	if *batchFile != "" {
		return runBatchMode(workDir, scriptsPath)
	}

	// Single file mode: keep existing logic for backward compatibility
	if *beforeFile == "" && *afterFile == "" {
		return fmt.Errorf("either -before, -after, or -batch must be specified")
	}

	return runSingleFileMode(workDir, scriptsPath)
}

func runBatchMode(workDir, scriptsPath string) error {
	tmpDir, err := os.MkdirTemp("", "ast-extractor-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &phases.Config{
		WorkDir:    workDir,
		OutputDir:  tmpDir,
		Verbose:    *verbose,
		ScriptsDir: scriptsPath,
		Language:   *language,
		BatchFile:  *batchFile,
	}

	phase := astphase.New()

	timeoutDuration := *timeout
	if timeoutDuration == 0 {
		timeoutDuration = phase.Timeout()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	if err := phase.Run(ctx, cfg); err != nil {
		return err
	}

	// Read output from phase and print to stdout
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "-ast.json") {
			data, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to read output file: %w", err)
			}

			if *outputFmt == "markdown" {
				var diffs []ast.SemanticDiff
				if err := json.Unmarshal(data, &diffs); err != nil {
					return fmt.Errorf("failed to parse AST output: %w", err)
				}
				fmt.Print(ast.RenderMultipleMarkdown(diffs))
			} else {
				fmt.Println(string(data))
			}
		}
	}

	return nil
}

func runSingleFileMode(workDir, scriptsPath string) error {
	if err := validateScriptsDir(scriptsPath); err != nil {
		return fmt.Errorf("scripts directory validation failed: %w", err)
	}

	if *verbose {
		logger.Debug("scripts directory", "path", scriptsPath)
	}

	validator, err := ast.NewPathValidator(workDir)
	if err != nil {
		return fmt.Errorf("failed to initialize path validator: %w", err)
	}

	before := *beforeFile
	after := *afterFile

	if before != "" {
		validated, err := validator.ValidatePath(before)
		if err != nil {
			return fmt.Errorf("invalid before file path: %w", err)
		}
		before = validated
	}
	if after != "" {
		validated, err := validator.ValidatePath(after)
		if err != nil {
			return fmt.Errorf("invalid after file path: %w", err)
		}
		after = validated
	}

	registry := ast.NewRegistry()
	registry.Register(ast.NewGoExtractor())
	registry.Register(ast.NewTypeScriptExtractor(scriptsPath))
	registry.Register(ast.NewPythonExtractor(scriptsPath))

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	filePath := after
	if filePath == "" {
		filePath = before
	}

	var extractor ast.Extractor
	var extractErr error

	if *language != "" {
		extractor, extractErr = getExtractorByLanguage(*language, scriptsPath)
	} else {
		extractor, extractErr = registry.GetExtractor(filePath)
	}

	if extractErr != nil {
		return fmt.Errorf("failed to get extractor: %w", extractErr)
	}

	if *verbose {
		logger.Debug("using extractor", "language", extractor.Language(), "before", before, "after", after)
	}

	diff, err := extractor.ExtractDiff(ctx, before, after)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	return outputDiff(diff)
}

func getExtractorByLanguage(lang string, scriptsPath string) (ast.Extractor, error) {
	switch strings.ToLower(lang) {
	case "go", "golang":
		return ast.NewGoExtractor(), nil
	case "ts", "typescript", "javascript", "js":
		return ast.NewTypeScriptExtractor(scriptsPath), nil
	case "py", "python":
		return ast.NewPythonExtractor(scriptsPath), nil
	default:
		return nil, fmt.Errorf("unknown language: %s", lang)
	}
}

func validateScriptsDir(scriptsDir string) error {
	if scriptsDir == "" {
		return nil
	}
	if strings.Contains(scriptsDir, "..") {
		return fmt.Errorf("path traversal detected in scripts directory")
	}
	_, err := fileutil.ValidateDirectory(scriptsDir, "")
	return err
}

func outputDiff(diff *ast.SemanticDiff) error {
	if *outputFmt == "markdown" {
		fmt.Print(ast.RenderMarkdown(diff))
		return nil
	}

	output, err := ast.RenderJSON(diff)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
