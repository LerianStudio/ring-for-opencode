package dataflow

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
)

// capitalizeFirst capitalizes the first letter of a string.
// stdlib-only replacement for deprecated strings.Title.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// escapeMarkdownInline escapes special markdown characters in inline text
// to prevent markdown injection attacks in generated reports.
func escapeMarkdownInline(s string) string {
	replacer := strings.NewReplacer(
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"#", "\\#",
		"|", "\\|",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(s)
}

// escapeMarkdownCodeBlock escapes content for code blocks to prevent
// breaking out of code blocks via embedded triple backticks.
func escapeMarkdownCodeBlock(s string) string {
	return strings.ReplaceAll(s, "```", "` ` `")
}

// GenerateSecuritySummary generates a complete markdown report from security analysis results.
func GenerateSecuritySummary(analyses map[string]*FlowAnalysis) string {
	var sb strings.Builder

	// Aggregate statistics
	var totalStats Stats
	var languages []string
	var allFlows []Flow
	var allNilSources []NilSource

	for lang, analysis := range analyses {
		if analysis == nil {
			continue
		}
		languages = append(languages, lang)
		totalStats.TotalSources += analysis.Statistics.TotalSources
		totalStats.TotalSinks += analysis.Statistics.TotalSinks
		totalStats.TotalFlows += analysis.Statistics.TotalFlows
		totalStats.UnsanitizedFlows += analysis.Statistics.UnsanitizedFlows
		totalStats.CriticalFlows += analysis.Statistics.CriticalFlows
		totalStats.HighRiskFlows += analysis.Statistics.HighRiskFlows
		totalStats.NilRisks += analysis.Statistics.NilRisks
		totalStats.UncheckedNilRisks += analysis.Statistics.UncheckedNilRisks
		allFlows = append(allFlows, analysis.Flows...)
		allNilSources = append(allNilSources, analysis.NilSources...)
	}

	sort.Strings(languages)

	// Sort flows by risk priority
	sort.Slice(allFlows, func(i, j int) bool {
		return riskPriority(allFlows[i].Risk) < riskPriority(allFlows[j].Risk)
	})

	// Header
	sb.WriteString("# Security Data Flow Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))

	// Executive Summary
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Languages Analyzed | %d |\n", len(languages)))
	sb.WriteString(fmt.Sprintf("| Total Sources | %d |\n", totalStats.TotalSources))
	sb.WriteString(fmt.Sprintf("| Total Sinks | %d |\n", totalStats.TotalSinks))
	sb.WriteString(fmt.Sprintf("| Total Flows | %d |\n", totalStats.TotalFlows))
	sb.WriteString(fmt.Sprintf("| Unsanitized Flows | %d |\n", totalStats.UnsanitizedFlows))
	sb.WriteString(fmt.Sprintf("| Critical Risk Flows | %d |\n", totalStats.CriticalFlows))
	sb.WriteString(fmt.Sprintf("| High Risk Flows | %d |\n", totalStats.HighRiskFlows))
	sb.WriteString(fmt.Sprintf("| Nil/Null Risks | %d |\n", totalStats.NilRisks))
	sb.WriteString("\n")

	// Risk Assessment
	sb.WriteString("## Risk Assessment\n\n")

	if totalStats.CriticalFlows > 0 {
		sb.WriteString(fmt.Sprintf("### :rotating_light: CRITICAL (%d issues)\n\n", totalStats.CriticalFlows))
		sb.WriteString("Critical security vulnerabilities detected that require immediate attention.\n\n")
	}

	if totalStats.HighRiskFlows > 0 {
		sb.WriteString(fmt.Sprintf("### :warning: HIGH (%d issues)\n\n", totalStats.HighRiskFlows))
		sb.WriteString("High-risk security issues that should be addressed promptly.\n\n")
	}

	if totalStats.UncheckedNilRisks > 0 {
		sb.WriteString(fmt.Sprintf("### :exclamation: NIL SAFETY (%d issues)\n\n", totalStats.UncheckedNilRisks))
		sb.WriteString("Unchecked nil/null values that may cause runtime panics or crashes.\n\n")
	}

	if totalStats.CriticalFlows == 0 && totalStats.HighRiskFlows == 0 && totalStats.UncheckedNilRisks == 0 {
		sb.WriteString("No critical, high-risk, or unchecked nil safety issues detected.\n\n")
	}

	// Critical & High Risk Flows Detail
	if totalStats.CriticalFlows > 0 || totalStats.HighRiskFlows > 0 {
		sb.WriteString("## Critical & High Risk Flows\n\n")

		flowNum := 0
		for _, flow := range allFlows {
			if flow.Risk != RiskCritical && flow.Risk != RiskHigh {
				continue
			}
			flowNum++

			icon := ":warning:"
			if flow.Risk == RiskCritical {
				icon = ":rotating_light:"
			}

			sb.WriteString(fmt.Sprintf("### %s Flow %d: %s\n\n", icon, flowNum, escapeMarkdownInline(flow.Description)))

			sb.WriteString(fmt.Sprintf("**Risk Level:** %s\n\n", capitalizeFirst(string(flow.Risk))))

			sb.WriteString(fmt.Sprintf("**Source:** `%s:%d` (Type: %s)\n\n", escapeMarkdownInline(flow.Source.File), flow.Source.Line, escapeMarkdownInline(string(flow.Source.Type))))

			sb.WriteString(fmt.Sprintf("**Sink:** `%s:%d` (Function: %s)\n\n", escapeMarkdownInline(flow.Sink.File), flow.Sink.Line, escapeMarkdownInline(flow.Sink.Function)))

			if flow.Sanitized {
				sb.WriteString(fmt.Sprintf("**Sanitized:** Yes (Sanitizers: %s)\n\n", escapeMarkdownInline(strings.Join(flow.Sanitizers, ", "))))
			} else {
				sb.WriteString("**Sanitized:** No\n\n")
			}

			if flow.Source.Context != "" {
				sb.WriteString("**Source Context:**\n```\n")
				sb.WriteString(escapeMarkdownCodeBlock(flow.Source.Context))
				sb.WriteString("\n```\n\n")
			}

			if flow.Sink.Context != "" {
				sb.WriteString("**Sink Context:**\n```\n")
				sb.WriteString(escapeMarkdownCodeBlock(flow.Sink.Context))
				sb.WriteString("\n```\n\n")
			}

			sb.WriteString(fmt.Sprintf("**Recommendation:** %s\n\n", getRecommendation(flow)))
			sb.WriteString("---\n\n")
		}
	}

	// Nil/Null Safety Issues
	if len(allNilSources) > 0 {
		sb.WriteString("## Nil/Null Safety Issues\n\n")
		sb.WriteString("| File | Line | Variable | Origin | Checked | Risk |\n")
		sb.WriteString("|------|------|----------|--------|---------|------|\n")

		for _, ns := range allNilSources {
			checked := "No"
			if ns.IsChecked {
				checked = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s | %s | %s |\n",
				escapeMarkdownInline(ns.File), ns.Line, escapeMarkdownInline(ns.Variable), escapeMarkdownInline(ns.Origin), checked, capitalizeFirst(string(ns.Risk))))
		}
		sb.WriteString("\n")
	}

	// Language Breakdown
	if len(languages) > 0 {
		sb.WriteString("## Language Breakdown\n\n")

		for _, lang := range languages {
			analysis := analyses[lang]
			if analysis == nil {
				continue
			}

			sb.WriteString(fmt.Sprintf("### %s\n\n", capitalizeFirst(lang)))
			sb.WriteString("| Metric | Value |\n")
			sb.WriteString("|--------|-------|\n")
			sb.WriteString(fmt.Sprintf("| Sources | %d |\n", analysis.Statistics.TotalSources))
			sb.WriteString(fmt.Sprintf("| Sinks | %d |\n", analysis.Statistics.TotalSinks))
			sb.WriteString(fmt.Sprintf("| Flows | %d |\n", analysis.Statistics.TotalFlows))
			sb.WriteString(fmt.Sprintf("| Unsanitized | %d |\n", analysis.Statistics.UnsanitizedFlows))
			sb.WriteString(fmt.Sprintf("| Critical | %d |\n", analysis.Statistics.CriticalFlows))
			sb.WriteString(fmt.Sprintf("| High | %d |\n", analysis.Statistics.HighRiskFlows))
			sb.WriteString(fmt.Sprintf("| Nil Risks | %d |\n", analysis.Statistics.NilRisks))
			sb.WriteString("\n")
		}
	}

	// General Recommendations
	sb.WriteString("## General Recommendations\n\n")
	sb.WriteString("1. **Input Validation**: Always validate and sanitize user input at the entry point.\n")
	sb.WriteString("2. **Parameterized Queries**: Use prepared statements or parameterized queries for all database operations.\n")
	sb.WriteString("3. **Output Encoding**: Encode output appropriately for the context (HTML, URL, JavaScript).\n")
	sb.WriteString("4. **Nil Checks**: Always check for nil/null before dereferencing pointers or optional values.\n")
	sb.WriteString("5. **Principle of Least Privilege**: Avoid command execution; if required, use strict allow lists.\n")
	sb.WriteString("6. **Security Testing**: Integrate security scanning into CI/CD pipelines for continuous monitoring.\n")

	return sb.String()
}

// riskPriority returns sort priority for risk levels (lower = higher priority).
func riskPriority(risk RiskLevel) int {
	switch risk {
	case RiskCritical:
		return 0
	case RiskHigh:
		return 1
	case RiskMedium:
		return 2
	case RiskLow:
		return 3
	case RiskInfo:
		return 4
	default:
		return 5
	}
}

// getRecommendation returns a context-specific security recommendation based on sink type.
func getRecommendation(flow Flow) string {
	switch flow.Sink.Type {
	case SinkExec:
		return "Remove command execution or use strict allow list"
	case SinkDatabase:
		return "Use parameterized queries or prepared statements"
	case SinkResponse:
		return "Escape output using html.EscapeString()"
	case SinkTemplate:
		return "Ensure template engine auto-escapes"
	case SinkRedirect:
		return "Validate redirect URLs against allow list"
	case SinkFile:
		return "Validate file paths and use filepath.Clean()"
	case SinkLog:
		return "Sanitize log output to prevent log injection"
	default:
		return "Review data flow and apply appropriate sanitization"
	}
}
