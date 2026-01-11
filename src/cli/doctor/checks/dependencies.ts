import { CHECK_IDS, CHECK_NAMES } from "../constants"
import type { CheckDefinition, CheckResult } from "../types"

async function checkCommandExists(
  command: string,
  args: string[],
): Promise<{ exists: boolean; version: string | null }> {
  try {
    const proc = Bun.spawn([command, ...args], {
      stdout: "pipe",
      stderr: "pipe",
    })
    const output = await new Response(proc.stdout).text()
    await proc.exited

    if (proc.exitCode === 0) {
      return { exists: true, version: output.trim().split("\n")[0] }
    }
    return { exists: false, version: null }
  } catch {
    return { exists: false, version: null }
  }
}

export async function checkBunInstalled(): Promise<CheckResult> {
  const result = await checkCommandExists("bun", ["--version"])

  if (!result.exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.BUN_INSTALLED],
      status: "fail",
      message: "Bun is not installed",
      details: ["Install Bun: curl -fsSL https://bun.sh/install | bash"],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.BUN_INSTALLED],
    status: "pass",
    message: `Version ${result.version}`,
  }
}

export async function checkGitInstalled(): Promise<CheckResult> {
  const result = await checkCommandExists("git", ["--version"])

  if (!result.exists) {
    return {
      name: CHECK_NAMES[CHECK_IDS.GIT_INSTALLED],
      status: "warn",
      message: "Git is not installed",
      details: ["Git is recommended for version control", "Install: https://git-scm.com/downloads"],
    }
  }

  return {
    name: CHECK_NAMES[CHECK_IDS.GIT_INSTALLED],
    status: "pass",
    message: result.version ?? "Installed",
  }
}

export function getDependencyCheckDefinitions(): CheckDefinition[] {
  return [
    {
      id: CHECK_IDS.BUN_INSTALLED,
      name: CHECK_NAMES[CHECK_IDS.BUN_INSTALLED],
      category: "dependencies",
      check: checkBunInstalled,
      critical: true,
    },
    {
      id: CHECK_IDS.GIT_INSTALLED,
      name: CHECK_NAMES[CHECK_IDS.GIT_INSTALLED],
      category: "dependencies",
      check: checkGitInstalled,
      critical: false,
    },
  ]
}
