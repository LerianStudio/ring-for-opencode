import * as p from "@clack/prompts"
import color from "picocolors"
import type { InstallArgs, InstallConfig } from "./types"
import {
  addSchemaToConfig,
  isOpenCodeInstalled,
  getOpenCodeVersion,
  detectCurrentConfig,
  validateConfig,
} from "./config-manager"
import { SYMBOLS, SCHEMA_URL } from "./constants"

function printHeader(isUpdate: boolean): void {
  const mode = isUpdate ? "Update" : "Install"
  console.log()
  console.log(color.bgCyan(color.white(` Ring ${mode} `)))
  console.log()
}

function printStep(step: number, total: number, message: string): void {
  const progress = color.dim(`[${step}/${total}]`)
  console.log(`${progress} ${message}`)
}

function printSuccess(message: string): void {
  console.log(`${SYMBOLS.check} ${message}`)
}

function printError(message: string): void {
  console.log(`${SYMBOLS.cross} ${color.red(message)}`)
}

function printInfo(message: string): void {
  console.log(`${SYMBOLS.info} ${message}`)
}

function printWarning(message: string): void {
  console.log(`${SYMBOLS.warn} ${color.yellow(message)}`)
}

async function runTuiInstall(): Promise<number> {
  const detected = detectCurrentConfig()
  const isUpdate = detected.isInstalled

  p.intro(color.bgCyan(color.white(isUpdate ? " Ring Update " : " Ring Install ")))

  if (isUpdate && detected.configPath) {
    p.log.info(`Existing configuration found: ${detected.configPath}`)
  }

  const s = p.spinner()
  s.start("Checking OpenCode installation")

  const installed = await isOpenCodeInstalled()
  if (!installed) {
    s.stop("OpenCode is not installed")
    p.log.error("OpenCode is not installed on this system.")
    p.note("Visit https://opencode.ai/docs for installation instructions", "Installation Guide")
    p.outro(color.red("Please install OpenCode first."))
    return 1
  }

  const version = await getOpenCodeVersion()
  s.stop(`OpenCode ${version ?? "installed"} ${color.green("\u2713")}`)

  // Confirm installation
  const shouldContinue = await p.confirm({
    message: isUpdate
      ? "Update Ring configuration with schema validation?"
      : "Install Ring configuration with schema validation?",
    initialValue: true,
  })

  if (p.isCancel(shouldContinue) || !shouldContinue) {
    p.cancel("Installation cancelled.")
    return 1
  }

  s.start("Adding schema to configuration")
  const schemaResult = addSchemaToConfig()
  if (!schemaResult.success) {
    s.stop(`Failed: ${schemaResult.error}`)
    p.outro(color.red("Installation failed."))
    return 1
  }
  s.stop(`Schema added to ${color.cyan(schemaResult.configPath)}`)

  // Validate the updated config
  s.start("Validating configuration")
  const validation = validateConfig(schemaResult.configPath)
  if (!validation.valid) {
    s.stop("Configuration has validation errors")
    p.log.warn("Validation errors found:")
    for (const err of validation.errors) {
      p.log.message(`  ${SYMBOLS.bullet} ${err}`)
    }
  } else {
    s.stop("Configuration is valid")
  }

  p.note(
    `Your opencode.json now has schema validation.\n` +
      `IDE autocomplete is available via the $schema field.\n\n` +
      `Schema URL: ${color.cyan(SCHEMA_URL)}`,
    isUpdate ? "Configuration Updated" : "Installation Complete"
  )

  p.log.success(color.bold(isUpdate ? "Ring configuration updated!" : "Ring installed!"))
  p.log.message(`Run ${color.cyan("ring doctor")} to check your setup.`)

  p.outro(color.green("Happy coding with Ring!"))

  return 0
}

async function runNonTuiInstall(args: InstallArgs): Promise<number> {
  const detected = detectCurrentConfig()
  const isUpdate = detected.isInstalled

  printHeader(isUpdate)

  const totalSteps = 3
  let step = 1

  printStep(step++, totalSteps, "Checking OpenCode installation...")
  const installed = await isOpenCodeInstalled()
  if (!installed) {
    printError("OpenCode is not installed on this system.")
    printInfo("Visit https://opencode.ai/docs for installation instructions")
    return 1
  }

  const version = await getOpenCodeVersion()
  printSuccess(`OpenCode ${version ?? ""} detected`)

  printStep(step++, totalSteps, "Adding schema to configuration...")
  const schemaResult = addSchemaToConfig()
  if (!schemaResult.success) {
    printError(`Failed: ${schemaResult.error}`)
    return 1
  }
  printSuccess(`Schema added ${SYMBOLS.arrow} ${color.dim(schemaResult.configPath)}`)

  if (!args.skipValidation) {
    printStep(step++, totalSteps, "Validating configuration...")
    const validation = validateConfig(schemaResult.configPath)
    if (!validation.valid) {
      printWarning("Configuration has validation errors:")
      for (const err of validation.errors) {
        console.log(`  ${SYMBOLS.bullet} ${err}`)
      }
    } else {
      printSuccess("Configuration is valid")
    }
  }

  console.log()
  printSuccess(color.bold(isUpdate ? "Ring configuration updated!" : "Ring installed!"))
  console.log(`  Run ${color.cyan("ring doctor")} to check your setup.`)
  console.log()

  return 0
}

export async function install(args: InstallArgs): Promise<number> {
  if (!args.tui) {
    return runNonTuiInstall(args)
  }

  return runTuiInstall()
}
