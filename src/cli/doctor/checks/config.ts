import { existsSync } from "node:fs"
import type { CheckResult, CheckDefinition } from "../types"
import { CHECK_IDS, CHECK_NAMES } from "../constants"
import { getConfigPath, validateConfig, detectCurrentConfig } from "../../config-manager"
import { SCHEMA_URL } from "../../constants"

export async function checkConfigExists(): Promise<CheckResult> {
  const configPath = getConfigPath()
  const exists = existsSync(configPath)

  if (!exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.CONFIG_EXISTS],
      status: "fail",
      message: "opencode.json not found",
      details: [
        "Create opencode.json in project root",
        "Run: ring install",
      ],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.CONFIG_EXISTS],
    status: "pass",
    message: "Found",
    details: [`Path: ${configPath}`],
  }
}

export async function checkConfigValidity(): Promise<CheckResult> {
  const configPath = getConfigPath()

  if (!existsSync(configPath)) {
    return {
      name: CHECK_NAMES[CHECK_IDS.CONFIG_VALIDATION],
      status: "skip",
      message: "No config file to validate",
    }
  }

  const validation = validateConfig(configPath)

  if (!validation.valid) {
    return {
      name: CHECK_NAMES[CHECK_IDS.CONFIG_VALIDATION],
      status: "fail",
      message: "Configuration has validation errors",
      details: [
        `Path: ${configPath}`,
        ...validation.errors.map((e) => `Error: ${e}`),
      ],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.CONFIG_VALIDATION],
    status: "pass",
    message: "Valid configuration",
    details: [`Path: ${configPath}`],
  }
}

export async function checkSchemaPresent(): Promise<CheckResult> {
  const detected = detectCurrentConfig()

  if (!detected.isInstalled) {
    return {
      name: CHECK_NAMES[CHECK_IDS.SCHEMA_PRESENT],
      status: "skip",
      message: "No config file",
    }
  }

  if (!detected.hasSchema) {
    return {
      name: CHECK_NAMES[CHECK_IDS.SCHEMA_PRESENT],
      status: "warn",
      message: "No $schema reference for IDE autocomplete",
      details: [
        `Add to opencode.json: "$schema": "${SCHEMA_URL}"`,
        "Or run: ring install",
      ],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.SCHEMA_PRESENT],
    status: "pass",
    message: "Schema reference present",
  }
}

export function getConfigCheckDefinitions(): CheckDefinition[] {
  return [
    {
      id: CHECK_IDS.CONFIG_EXISTS,
      name: CHECK_NAMES[CHECK_IDS.CONFIG_EXISTS],
      category: "configuration",
      check: checkConfigExists,
      critical: true,
    },
    {
      id: CHECK_IDS.CONFIG_VALIDATION,
      name: CHECK_NAMES[CHECK_IDS.CONFIG_VALIDATION],
      category: "configuration",
      check: checkConfigValidity,
      critical: false,
    },
    {
      id: CHECK_IDS.SCHEMA_PRESENT,
      name: CHECK_NAMES[CHECK_IDS.SCHEMA_PRESENT],
      category: "configuration",
      check: checkSchemaPresent,
      critical: false,
    },
  ]
}
