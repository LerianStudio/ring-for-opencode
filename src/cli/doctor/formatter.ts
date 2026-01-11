import color from "picocolors"
import type { CheckResult, DoctorSummary, CheckCategory, DoctorResult } from "./types"
import { SYMBOLS, STATUS_COLORS, CATEGORY_NAMES } from "./constants"

export function formatStatusSymbol(status: CheckResult["status"]): string {
  switch (status) {
    case "pass":
      return SYMBOLS.check
    case "fail":
      return SYMBOLS.cross
    case "warn":
      return SYMBOLS.warn
    case "skip":
      return SYMBOLS.skip
  }
}

export function formatCheckResult(result: CheckResult, verbose: boolean): string {
  const symbol = formatStatusSymbol(result.status)
  const colorFn = STATUS_COLORS[result.status]
  const name = colorFn(result.name)
  const message = color.dim(result.message)

  let line = `  ${symbol} ${name}`
  if (result.message) {
    line += ` ${SYMBOLS.arrow} ${message}`
  }

  if (verbose && result.details && result.details.length > 0) {
    const detailLines = result.details.map((d) => `      ${SYMBOLS.bullet} ${color.dim(d)}`).join("\n")
    line += "\n" + detailLines
  }

  return line
}

export function formatCategoryHeader(category: CheckCategory): string {
  const name = CATEGORY_NAMES[category] || category
  return `\n${color.bold(color.white(name))}\n${color.dim("\u2500".repeat(40))}`
}

export function formatSummary(summary: DoctorSummary): string {
  const lines: string[] = []

  lines.push(color.bold(color.white("Summary")))
  lines.push(color.dim("\u2500".repeat(40)))
  lines.push("")

  const passText = summary.passed > 0 ? color.green(`${summary.passed} passed`) : color.dim("0 passed")
  const failText = summary.failed > 0 ? color.red(`${summary.failed} failed`) : color.dim("0 failed")
  const warnText = summary.warnings > 0 ? color.yellow(`${summary.warnings} warnings`) : color.dim("0 warnings")
  const skipText = summary.skipped > 0 ? color.dim(`${summary.skipped} skipped`) : ""

  const parts = [passText, failText, warnText]
  if (skipText) parts.push(skipText)

  lines.push(`  ${parts.join(", ")}`)
  lines.push(`  ${color.dim(`Total: ${summary.total} checks in ${summary.duration}ms`)}`)

  return lines.join("\n")
}

export function formatHeader(): string {
  return `\n${color.bgCyan(color.white(" Ring Doctor "))}\n`
}

export function formatFooter(summary: DoctorSummary): string {
  if (summary.failed > 0) {
    return `\n${SYMBOLS.cross} ${color.red("Issues detected. Please review the errors above.")}\n`
  }
  if (summary.warnings > 0) {
    return `\n${SYMBOLS.warn} ${color.yellow("All systems operational with warnings.")}\n`
  }
  return `\n${SYMBOLS.check} ${color.green("All systems operational!")}\n`
}

export function formatJsonOutput(result: DoctorResult): string {
  return JSON.stringify(result, null, 2)
}
