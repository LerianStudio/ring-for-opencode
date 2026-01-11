import color from "picocolors"

export const PACKAGE_NAME = "ring-opencode"
export const CONFIG_FILE_NAME = "opencode.json"
export const SCHEMA_URL =
  "https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-opencode.schema.json"

export const SYMBOLS = {
  check: color.green("\u2713"),
  cross: color.red("\u2717"),
  warn: color.yellow("\u26A0"),
  info: color.blue("\u2139"),
  arrow: color.cyan("\u2192"),
  bullet: color.dim("\u2022"),
  skip: color.dim("\u25CB"),
} as const

export const STATUS_COLORS = {
  pass: color.green,
  fail: color.red,
  warn: color.yellow,
  skip: color.dim,
} as const

export const EXIT_CODES = {
  SUCCESS: 0,
  FAILURE: 1,
} as const
