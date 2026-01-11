import { existsSync } from "node:fs"
import { join } from "node:path"
import { getOpenCodeVersion, isOpenCodeInstalled } from "../../config-manager"
import { CHECK_IDS, CHECK_NAMES } from "../constants"
import type { CheckDefinition, CheckResult } from "../types"

export async function checkOpenCodeInstallation(): Promise<CheckResult> {
  const installed = await isOpenCodeInstalled()

  if (!installed) {
    return {
      name: CHECK_NAMES[CHECK_IDS.OPENCODE_INSTALLATION],
      status: "fail",
      message: "OpenCode is not installed",
      details: [
        "Install OpenCode: https://opencode.ai/docs",
        "Run: curl -fsSL https://opencode.ai/install | bash",
      ],
    }
  }

  const version = await getOpenCodeVersion()
  return {
    name: CHECK_NAMES[CHECK_IDS.OPENCODE_INSTALLATION],
    status: "pass",
    message: version ? `Version ${version}` : "Installed",
  }
}

export async function checkPluginDirectory(): Promise<CheckResult> {
  const pluginDir = join(process.cwd(), "plugin")
  const exists = existsSync(pluginDir)

  if (!exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.PLUGIN_DIRECTORY],
      status: "warn",
      message: "Plugin directory not found",
      details: ["Expected: ./plugin/"],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.PLUGIN_DIRECTORY],
    status: "pass",
    message: "Found",
    details: [`Path: ${pluginDir}`],
  }
}

export async function checkSkillDirectory(): Promise<CheckResult> {
  const skillDir = join(process.cwd(), "skill")
  const exists = existsSync(skillDir)

  if (!exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.SKILL_DIRECTORY],
      status: "warn",
      message: "Skill directory not found",
      details: ["Expected: ./skill/"],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.SKILL_DIRECTORY],
    status: "pass",
    message: "Found",
    details: [`Path: ${skillDir}`],
  }
}

export async function checkStateDirectory(): Promise<CheckResult> {
  const stateDir = join(process.cwd(), ".opencode", "state")
  const exists = existsSync(stateDir)

  if (!exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.STATE_DIRECTORY],
      status: "skip",
      message: "State directory will be created on first run",
      details: ["Expected: ./.opencode/state/"],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.STATE_DIRECTORY],
    status: "pass",
    message: "Found",
    details: [`Path: ${stateDir}`],
  }
}

export function getInstallationCheckDefinitions(): CheckDefinition[] {
  return [
    {
      id: CHECK_IDS.OPENCODE_INSTALLATION,
      name: CHECK_NAMES[CHECK_IDS.OPENCODE_INSTALLATION],
      category: "installation",
      check: checkOpenCodeInstallation,
      critical: true,
    },
  ]
}

export function getPluginCheckDefinitions(): CheckDefinition[] {
  return [
    {
      id: CHECK_IDS.PLUGIN_DIRECTORY,
      name: CHECK_NAMES[CHECK_IDS.PLUGIN_DIRECTORY],
      category: "plugins",
      check: checkPluginDirectory,
    },
    {
      id: CHECK_IDS.SKILL_DIRECTORY,
      name: CHECK_NAMES[CHECK_IDS.SKILL_DIRECTORY],
      category: "plugins",
      check: checkSkillDirectory,
    },
    {
      id: CHECK_IDS.STATE_DIRECTORY,
      name: CHECK_NAMES[CHECK_IDS.STATE_DIRECTORY],
      category: "configuration",
      check: checkStateDirectory,
    },
  ]
}
