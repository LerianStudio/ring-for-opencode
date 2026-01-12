#!/usr/bin/env bun
import { Command } from "commander"
import type { DoctorOptions } from "./doctor"
import { doctor } from "./doctor"
import { help } from "./help"
import type { HelpOptions } from "./help"
import { install } from "./install"
import type { InstallArgs } from "./types"
import { version as versionCmd } from "./version"

const packageJson = await import("../../package.json")
const VERSION = packageJson.version ?? "0.0.0"

const program = new Command()

program
  .name("ring")
  .description("Ring - CLI tools for OpenCode configuration and health checks")
  .version(VERSION, "-v, --version", "Show version number")

program
  .command("install")
  .description("Install and configure Ring with schema validation")
  .option("--no-tui", "Run in non-interactive mode")
  .option("--skip-validation", "Skip config validation after install")
  .addHelpText(
    "after",
    `
Examples:
  $ ring install
  $ ring install --no-tui
  $ ring install --no-tui --skip-validation

This command:
  - Adds $schema to Ring config for IDE autocomplete
  - Validates configuration against Ring schema
  - Creates config if it doesn't exist (.opencode/ring.jsonc or ~/.config/opencode/ring/config.jsonc)
`,
  )
  .action(async (options) => {
    const args: InstallArgs = {
      tui: options.tui !== false,
      skipValidation: options.skipValidation ?? false,
    }
    const exitCode = await install(args)
    process.exit(exitCode)
  })

program
  .command("doctor")
  .description("Check Ring installation health and diagnose issues")
  .option("--verbose", "Show detailed diagnostic information")
  .option("--json", "Output results in JSON format")
  .option(
    "--category <category>",
    "Run only specific category (installation, configuration, plugins, dependencies)",
  )
  .addHelpText(
    "after",
    `
Examples:
  $ ring doctor
  $ ring doctor --verbose
  $ ring doctor --json
  $ ring doctor --category configuration

Categories:
  installation     Check OpenCode installation
  configuration    Validate configuration files
  plugins          Check plugin and skill directories
  dependencies     Check runtime dependencies (bun, git)
`,
  )
  .action(async (options) => {
    const doctorOptions: DoctorOptions = {
      verbose: options.verbose ?? false,
      json: options.json ?? false,
      category: options.category,
    }
    const exitCode = await doctor(doctorOptions)
    process.exit(exitCode)
  })

program
  .command("version")
  .description("Show detailed version information")
  .option("--json", "Output in JSON format")
  .action(async (options) => {
    const exitCode = await versionCmd({ json: options.json ?? false })
    process.exit(exitCode)
  })

program
  .command("help [item]")
  .description("Show available Ring skills, commands, and agents")
  .option("--skills", "Show only skills")
  .option("--commands", "Show only commands")
  .option("--agents", "Show only agents")
  .option("--json", "Output in JSON format")
  .addHelpText(
    "after",
    `
Examples:
  $ ring help              # Show all categories
  $ ring help --skills     # Show only skills
  $ ring help --commands   # Show only commands
  $ ring help --agents     # Show only agents
  $ ring help brainstorm   # Show details for 'brainstorm' skill/command
  $ ring help --json       # Output in JSON format

This command lists all available Ring capabilities that can be used in OpenCode.
`,
  )
  .action(async (item, options) => {
    const helpOptions: HelpOptions = {
      skills: options.skills ?? false,
      commands: options.commands ?? false,
      agents: options.agents ?? false,
      json: options.json ?? false,
      item: item,
    }
    const exitCode = await help(helpOptions)
    process.exit(exitCode)
  })

program.parse()
