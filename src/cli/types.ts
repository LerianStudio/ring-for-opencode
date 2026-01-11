export interface InstallArgs {
  tui: boolean
  skipValidation?: boolean
}

export interface InstallConfig {
  configPath: string
  isNewInstall: boolean
}

export interface ConfigMergeResult {
  success: boolean
  configPath: string
  error?: string
}

export interface DetectedConfig {
  isInstalled: boolean
  configPath: string | null
  hasSchema: boolean
  version: string | null
}
