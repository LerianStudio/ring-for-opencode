package ast

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RenderMarkdown converts a SemanticDiff to markdown format
func RenderMarkdown(diff *SemanticDiff) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Semantic Changes: %s\n\n", diff.FilePath))
	sb.WriteString(fmt.Sprintf("**Language:** %s\n\n", diff.Language))

	// Summary section
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Category | Added | Removed | Modified |\n")
	sb.WriteString("|----------|-------|---------|----------|\n")
	sb.WriteString(fmt.Sprintf("| Functions | %d | %d | %d |\n",
		diff.Summary.FunctionsAdded,
		diff.Summary.FunctionsRemoved,
		diff.Summary.FunctionsModified))
	sb.WriteString(fmt.Sprintf("| Types | %d | %d | %d |\n",
		diff.Summary.TypesAdded,
		diff.Summary.TypesRemoved,
		diff.Summary.TypesModified))
	sb.WriteString(fmt.Sprintf("| Variables | %d | %d | %d |\n",
		diff.Summary.VariablesAdded,
		diff.Summary.VariablesRemoved,
		diff.Summary.VariablesModified))
	sb.WriteString(fmt.Sprintf("| Imports | %d | %d | - |\n\n",
		diff.Summary.ImportsAdded,
		diff.Summary.ImportsRemoved))

	// Functions section
	if len(diff.Functions) > 0 {
		sb.WriteString("## Functions\n\n")
		for _, fn := range diff.Functions {
			sb.WriteString(renderFunction(fn))
		}
	}

	// Types section
	if len(diff.Types) > 0 {
		sb.WriteString("## Types\n\n")
		for _, t := range diff.Types {
			sb.WriteString(renderType(t))
		}
	}

	// Imports section
	if len(diff.Imports) > 0 {
		sb.WriteString("## Imports\n\n")
		for _, imp := range diff.Imports {
			sb.WriteString(renderImport(imp))
		}
	}

	return sb.String()
}

func renderFunction(fn FunctionDiff) string {
	var sb strings.Builder

	icon := getChangeIcon(fn.ChangeType)
	sb.WriteString(fmt.Sprintf("### %s `%s`\n\n", icon, fn.Name))

	switch fn.ChangeType {
	case ChangeAdded:
		sb.WriteString("**Status:** Added\n\n")
		if fn.After != nil {
			sb.WriteString("```\n")
			sb.WriteString(formatSignature(fn.Name, fn.After))
			sb.WriteString("```\n\n")
		}

	case ChangeRemoved:
		sb.WriteString("**Status:** Removed\n\n")
		if fn.Before != nil {
			sb.WriteString("```\n")
			sb.WriteString(formatSignature(fn.Name, fn.Before))
			sb.WriteString("```\n\n")
		}

	case ChangeModified:
		sb.WriteString("**Status:** Modified\n\n")
		if fn.BodyDiff != "" {
			sb.WriteString(fmt.Sprintf("**Changes:** %s\n\n", fn.BodyDiff))
		}

		if fn.Before != nil && fn.After != nil {
			sb.WriteString("**Before:**\n```\n")
			sb.WriteString(formatSignature(fn.Name, fn.Before))
			sb.WriteString("```\n\n")
			sb.WriteString("**After:**\n```\n")
			sb.WriteString(formatSignature(fn.Name, fn.After))
			sb.WriteString("```\n\n")
		}
	}

	return sb.String()
}

func renderType(t TypeDiff) string {
	var sb strings.Builder

	icon := getChangeIcon(t.ChangeType)
	sb.WriteString(fmt.Sprintf("### %s `%s` (%s)\n\n", icon, t.Name, t.Kind))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", capitalizeFirst(string(t.ChangeType))))
	sb.WriteString(fmt.Sprintf("**Lines:** %d-%d\n\n", t.StartLine, t.EndLine))

	if len(t.Fields) > 0 {
		sb.WriteString("**Field Changes:**\n\n")
		sb.WriteString("| Field | Change | Old Type | New Type |\n")
		sb.WriteString("|-------|--------|----------|----------|\n")
		for _, f := range t.Fields {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				f.Name, f.ChangeType, f.OldType, f.NewType))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderImport(imp ImportDiff) string {
	icon := getChangeIcon(imp.ChangeType)
	alias := ""
	if imp.Alias != "" {
		alias = fmt.Sprintf(" as %s", imp.Alias)
	}
	return fmt.Sprintf("- %s `%s`%s\n", icon, imp.Path, alias)
}

func formatSignature(name string, sig *FuncSig) string {
	var params []string
	for _, p := range sig.Params {
		if p.Type != "" {
			params = append(params, fmt.Sprintf("%s: %s", p.Name, p.Type))
		} else {
			params = append(params, p.Name)
		}
	}

	returns := strings.Join(sig.Returns, ", ")
	if returns == "" {
		returns = "void"
	}

	prefix := ""
	if sig.IsAsync {
		prefix = "async "
	}
	if sig.Receiver != "" {
		prefix += fmt.Sprintf("(%s) ", sig.Receiver)
	}

	return fmt.Sprintf("%sfunc %s(%s) -> %s\n", prefix, name, strings.Join(params, ", "), returns)
}

func getChangeIcon(changeType ChangeType) string {
	switch changeType {
	case ChangeAdded:
		return "+"
	case ChangeRemoved:
		return "-"
	case ChangeModified:
		return "~"
	case ChangeRenamed:
		return ">"
	default:
		return "?"
	}
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// RenderJSON returns the diff as formatted JSON
func RenderJSON(diff *SemanticDiff) ([]byte, error) {
	return json.MarshalIndent(diff, "", "  ")
}

// RenderMultipleMarkdown renders multiple diffs into a single markdown document
func RenderMultipleMarkdown(diffs []SemanticDiff) string {
	var sb strings.Builder

	sb.WriteString("# Semantic Diff Report\n\n")

	// Global summary
	var totalFuncsAdded, totalFuncsRemoved, totalFuncsModified int
	var totalTypesAdded, totalTypesRemoved, totalTypesModified int
	var totalVarsAdded, totalVarsRemoved, totalVarsModified int
	var totalImportsAdded, totalImportsRemoved int

	for _, diff := range diffs {
		totalFuncsAdded += diff.Summary.FunctionsAdded
		totalFuncsRemoved += diff.Summary.FunctionsRemoved
		totalFuncsModified += diff.Summary.FunctionsModified
		totalTypesAdded += diff.Summary.TypesAdded
		totalTypesRemoved += diff.Summary.TypesRemoved
		totalTypesModified += diff.Summary.TypesModified
		totalVarsAdded += diff.Summary.VariablesAdded
		totalVarsRemoved += diff.Summary.VariablesRemoved
		totalVarsModified += diff.Summary.VariablesModified
		totalImportsAdded += diff.Summary.ImportsAdded
		totalImportsRemoved += diff.Summary.ImportsRemoved
	}

	sb.WriteString("## Overall Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Files analyzed:** %d\n\n", len(diffs)))
	sb.WriteString("| Category | Added | Removed | Modified |\n")
	sb.WriteString("|----------|-------|---------|----------|\n")
	sb.WriteString(fmt.Sprintf("| Functions | %d | %d | %d |\n",
		totalFuncsAdded, totalFuncsRemoved, totalFuncsModified))
	sb.WriteString(fmt.Sprintf("| Types | %d | %d | %d |\n",
		totalTypesAdded, totalTypesRemoved, totalTypesModified))
	sb.WriteString(fmt.Sprintf("| Variables | %d | %d | %d |\n",
		totalVarsAdded, totalVarsRemoved, totalVarsModified))
	sb.WriteString(fmt.Sprintf("| Imports | %d | %d | - |\n\n",
		totalImportsAdded, totalImportsRemoved))

	sb.WriteString("---\n\n")

	// Individual file diffs
	for _, diff := range diffs {
		sb.WriteString(RenderMarkdown(&diff))
		sb.WriteString("\n---\n\n")
	}

	return sb.String()
}
