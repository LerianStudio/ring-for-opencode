import color from "picocolors"
import { PACKAGE_NAME } from "./constants"

interface VersionOptions {
  json?: boolean
}

interface VersionInfo {
  name: string
  version: string
  nodeVersion: string
  bunVersion: string | null
  platform: string
  arch: string
}

async function getBunVersion(): Promise<string | null> {
  try {
    const proc = Bun.spawn(["bun", "--version"], {
      stdout: "pipe",
      stderr: "pipe",
    })
    const output = await new Response(proc.stdout).text()
    await proc.exited
    if (proc.exitCode === 0) {
      return output.trim()
    }
    return null
  } catch {
    return null
  }
}

export async function version(options: VersionOptions = {}): Promise<number> {
  // Read version from package.json
  const packageJson = await import("../../package.json")
  const ver = packageJson.version ?? "unknown"

  const bunVersion = await getBunVersion()

  const info: VersionInfo = {
    name: PACKAGE_NAME,
    version: ver,
    nodeVersion: process.version,
    bunVersion,
    platform: process.platform,
    arch: process.arch,
  }

  if (options.json) {
    console.log(JSON.stringify(info, null, 2))
    return 0
  }

  console.log()
  console.log(`${color.bold(color.cyan(PACKAGE_NAME))} ${color.green(`v${ver}`)}`)
  console.log()
  console.log(`  ${color.dim("Node:")}    ${info.nodeVersion}`)
  if (info.bunVersion) {
    console.log(`  ${color.dim("Bun:")}     ${info.bunVersion}`)
  }
  console.log(`  ${color.dim("Platform:")} ${info.platform} (${info.arch})`)
  console.log()

  return 0
}
