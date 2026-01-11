import type { CheckDefinition } from "../types"
import { getInstallationCheckDefinitions, getPluginCheckDefinitions } from "./installation"
import { getConfigCheckDefinitions } from "./config"
import { getDependencyCheckDefinitions } from "./dependencies"

export * from "./installation"
export * from "./config"
export * from "./dependencies"

export function getAllCheckDefinitions(): CheckDefinition[] {
  return [
    ...getInstallationCheckDefinitions(),
    ...getConfigCheckDefinitions(),
    ...getPluginCheckDefinitions(),
    ...getDependencyCheckDefinitions(),
  ]
}
