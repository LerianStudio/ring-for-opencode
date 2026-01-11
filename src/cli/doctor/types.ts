export type CheckStatus = "pass" | "fail" | "warn" | "skip"

export interface CheckResult {
  name: string
  status: CheckStatus
  message: string
  details?: string[]
  duration?: number
}

export type CheckFunction = () => Promise<CheckResult>

export type CheckCategory =
  | "installation"
  | "configuration"
  | "plugins"
  | "dependencies"

export interface CheckDefinition {
  id: string
  name: string
  category: CheckCategory
  check: CheckFunction
  critical?: boolean
}

export interface DoctorOptions {
  verbose?: boolean
  json?: boolean
  category?: CheckCategory
}

export interface DoctorSummary {
  total: number
  passed: number
  failed: number
  warnings: number
  skipped: number
  duration: number
}

export interface DoctorResult {
  results: CheckResult[]
  summary: DoctorSummary
  exitCode: number
}

export interface OpenCodeInfo {
  installed: boolean
  version: string | null
  path: string | null
}

export interface ConfigInfo {
  exists: boolean
  path: string | null
  format: "json" | "jsonc" | null
  valid: boolean
  errors: string[]
  hasSchema: boolean
}

export interface PluginInfo {
  name: string
  loaded: boolean
  path: string | null
  error?: string
}
