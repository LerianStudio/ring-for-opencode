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

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// validateScriptsDir validates the scripts directory path for security.
// It prevents path traversal attacks and verifies the directory exists.
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

func run() error {
	// Determine script directory for TypeScript/Python extractors
	scriptsPath := *scriptDir
	if scriptsPath == "" {
		// Default to relative path from executable
		exe, err := os.Executable()
		if err == nil {
			scriptsPath = filepath.Join(filepath.Dir(exe), "..", "..")
		} else {
			scriptsPath = "."
		}
	}

	// Validate scripts directory before use (ALWAYS, not just when flag provided)
	if err := validateScriptsDir(scriptsPath); err != nil {
		return fmt.Errorf("scripts directory validation failed: %w", err)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Scripts directory: %s\n", scriptsPath)
	}

	workDir, workErr := os.Getwd()
	if workErr != nil {
		return fmt.Errorf("failed to get working directory: %w", workErr)
	}

	validator, validatorErr := ast.NewPathValidator(workDir)
	if validatorErr != nil {
		return fmt.Errorf("failed to initialize path validator: %w", validatorErr)
	}

	if *beforeFile != "" {
		validated, err := validator.ValidatePath(*beforeFile)
		if err != nil {
			return fmt.Errorf("invalid before file path: %w", err)
		}
		*beforeFile = validated
	}
	if *afterFile != "" {
		validated, err := validator.ValidatePath(*afterFile)
		if err != nil {
			return fmt.Errorf("invalid after file path: %w", err)
		}
		*afterFile = validated
	}
	if *batchFile != "" {
		validated, err := validator.ValidatePath(*batchFile)
		if err != nil {
			return fmt.Errorf("invalid batch file path: %w", err)
		}
		*batchFile = validated
	}

	// Create registry with all extractors
	registry := ast.NewRegistry()
	registry.Register(ast.NewGoExtractor())
	registry.Register(ast.NewTypeScriptExtractor(scriptsPath))
	registry.Register(ast.NewPythonExtractor(scriptsPath))

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Handle batch mode
	if *batchFile != "" {
		return processBatch(ctx, registry, *batchFile)
	}

	// Single file mode
	if *beforeFile == "" && *afterFile == "" {
		return fmt.Errorf("either -before, -after, or -batch must be specified")
	}

	// Determine file path for language detection
	filePath := *afterFile
	if filePath == "" {
		filePath = *beforeFile
	}

	// Get extractor
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
		fmt.Fprintf(os.Stderr, "Using extractor: %s\n", extractor.Language())
		fmt.Fprintf(os.Stderr, "Before: %s\n", *beforeFile)
		fmt.Fprintf(os.Stderr, "After: %s\n", *afterFile)
	}

	// Extract diff
	diff, err := extractor.ExtractDiff(ctx, *beforeFile, *afterFile)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Output result
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

func processBatch(ctx context.Context, registry *ast.Registry, batchPath string) error {
	data, err := fileutil.ReadJSONFileWithLimit(batchPath)
	if err != nil {
		return fmt.Errorf("failed to read batch file: %w", err)
	}

	var pairs []ast.FilePair
	if err := json.Unmarshal(data, &pairs); err != nil {
		return fmt.Errorf("failed to parse batch file: %w", err)
	}

	// Ensure pairs is not nil (can be nil for empty JSON array or null)
	if pairs == nil {
		pairs = []ast.FilePair{}
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Processing %d file pairs\n", len(pairs))
	}

	diffs, err := registry.ExtractAll(ctx, pairs)
	if err != nil {
		return fmt.Errorf("batch extraction failed: %w", err)
	}

	// Output all diffs
	if *outputFmt == "markdown" {
		fmt.Print(ast.RenderMultipleMarkdown(diffs))
	} else {
		output, err := json.MarshalIndent(diffs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}
		fmt.Println(string(output))
	}

	return nil
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
