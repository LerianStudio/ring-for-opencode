/**
 * Notification Hook Factory
 *
 * Provides platform-specific notifications for session events.
 * Supports macOS, Linux, and Windows notification systems.
 */

import type { Hook, HookContext, HookFactory, HookOutput, HookResult } from "../types.js"

async function execFileAsync(cmd: string, args: string[]): Promise<void> {
  const { execFile } = await import("node:child_process")

  await new Promise<void>((resolve, reject) => {
    execFile(cmd, args, (err) => {
      if (err) {
        reject(err)
        return
      }
      resolve()
    })
  })
}

export type NotificationDispatch = (
  title: string,
  message: string,
  sound: boolean,
) => Promise<boolean>

/**
 * Configuration for notification hook.
 */
export interface NotificationConfig {
  /** Enable notifications */
  enabled?: boolean
  /** Notification sound (macOS only) */
  sound?: boolean
  /** Custom notification title */
  title?: string
  /** Events to notify on */
  notifyOn?: {
    sessionIdle?: boolean
    taskComplete?: boolean
    error?: boolean
  }

  /**
   * Optional notification transport override.
   * Primarily used for testing.
   */
  dispatch?: NotificationDispatch
}

type ResolvedNotificationConfig = Required<Omit<NotificationConfig, "dispatch">>

/** Default configuration */
const DEFAULT_CONFIG: ResolvedNotificationConfig = {
  enabled: true,
  sound: true,
  title: "Ring OpenCode",
  notifyOn: {
    sessionIdle: true,
    taskComplete: true,
    error: true,
  },
}

/**
 * Platform type for notification dispatch.
 */
type Platform = "darwin" | "linux" | "win32" | "unknown"

/**
 * Get current platform.
 */
function getPlatform(): Platform {
  const platform = process.platform
  if (platform === "darwin" || platform === "linux" || platform === "win32") {
    return platform
  }
  return "unknown"
}

/**
 * Escape string for AppleScript.
 */
function escapeAppleScript(str: string): string {
  return str.replace(/\\/g, "\\\\").replace(/"/g, '\\"')
}

/**
 * Sanitize string for Windows notification (remove special characters).
 */
function sanitizeWindowsString(str: string): string {
  return str.replace(/[^\w\s.,!?-]/g, "")
}

/**
 * Build notify-send args (Linux).
 *
 * NOTE: This is exported for unit testing.
 */
export function buildNotifySendArgs(title: string, message: string): string[] {
  // `--` ensures notify-send treats user strings as positional args
  // even if they start with '-' (option injection hardening).
  return ["--", title, message]
}

/**
 * Build osascript args (macOS).
 *
 * NOTE: This is exported for unit testing.
 */
export function buildOsaScriptArgs(title: string, message: string, sound: boolean): string[] {
  // IMPORTANT: Never interpolate user-provided content into the AppleScript itself.
  // Instead, pass it as argv after `--` and reference argv inside the script.
  const scriptBase = `on run argv
  set theTitle to item 1 of argv
  set theMessage to item 2 of argv
  display notification theMessage with title theTitle
end run`

  const scriptWithSound = `on run argv
  set theTitle to item 1 of argv
  set theMessage to item 2 of argv
  display notification theMessage with title theTitle sound name "Glass"
end run`

  const script = sound ? scriptWithSound : scriptBase

  return ["-e", script, "--", title, message]
}

/**
 * Send notification on macOS using osascript.
 */
async function notifyDarwin(title: string, message: string, sound: boolean): Promise<boolean> {
  try {
    // Use execFile with array arguments to prevent command injection
    await execFileAsync("osascript", buildOsaScriptArgs(title, message, sound))
    return true
  } catch {
    return false
  }
}

/**
 * Send notification on Linux using notify-send.
 */
async function notifyLinux(title: string, message: string): Promise<boolean> {
  try {
    // Use execFile with array arguments to prevent command injection
    await execFileAsync("notify-send", buildNotifySendArgs(title, message))
    return true
  } catch {
    // Try alternative methods
    try {
      // Try using zenity as fallback with array arguments
      await execFileAsync("zenity", ["--notification", `--text=${title}: ${message}`])
      return true
    } catch {
      return false
    }
  }
}

/**
 * Send notification on Windows using PowerShell.
 */
async function notifyWindows(title: string, message: string): Promise<boolean> {
  try {
    // Sanitize inputs to prevent PowerShell injection
    const safeTitle = sanitizeWindowsString(title)
    const safeMessage = sanitizeWindowsString(message)

    // Use Windows Forms NotifyIcon for basic toast notification
    const script = `
Add-Type -AssemblyName System.Windows.Forms
$balloon = New-Object System.Windows.Forms.NotifyIcon
$balloon.Icon = [System.Drawing.SystemIcons]::Information
$balloon.BalloonTipTitle = '${safeTitle}'
$balloon.BalloonTipText = '${safeMessage}'
$balloon.Visible = $true
$balloon.ShowBalloonTip(5000)
Start-Sleep -Milliseconds 5100
$balloon.Dispose()
    `.trim()

    // Use execFile with array arguments to prevent command injection
    await execFileAsync("powershell", ["-NoProfile", "-NonInteractive", "-Command", script])
    return true
  } catch {
    // No fallback - msg * is a security risk
    return false
  }
}

/**
 * Send a cross-platform notification.
 */
async function sendNotification(title: string, message: string, sound: boolean): Promise<boolean> {
  const platform = getPlatform()

  switch (platform) {
    case "darwin":
      return notifyDarwin(title, message, sound)
    case "linux":
      return notifyLinux(title, message)
    case "win32":
      return notifyWindows(title, message)
    default:
      return false
  }
}

/**
 * Get notification message based on event type.
 */
function getNotificationMessage(eventType: string, properties?: Record<string, unknown>): string {
  switch (eventType) {
    case "session.idle":
      return "Session is idle. Waiting for your input."
    case "task.complete":
      return "Task completed successfully!"
    case "session.error": {
      const errorMsg = properties?.message ?? properties?.error ?? "An error occurred"
      return `Session error: ${String(errorMsg)}`
    }
    case "error": {
      const errorMsg = properties?.message ?? "An error occurred"
      return `Error: ${String(errorMsg)}`
    }
    default:
      return `Event: ${eventType}`
  }
}

/**
 * Create a notification hook.
 */
export const createNotificationHook: HookFactory<NotificationConfig> = (
  config?: NotificationConfig,
): Hook => {
  const { dispatch, ...restConfig } = config ?? {}

  // Deep merge to preserve nested config properties
  const cfg: ResolvedNotificationConfig = {
    ...DEFAULT_CONFIG,
    ...restConfig,
    notifyOn: { ...DEFAULT_CONFIG.notifyOn, ...restConfig.notifyOn },
  }

  const notify: NotificationDispatch = dispatch ?? sendNotification

  return {
    name: "notification",
    lifecycles: ["session.idle", "session.error", "todo.updated", "event"],
    priority: 200, // Run late, after main processing
    enabled: cfg.enabled,

    async execute(ctx: HookContext, _output: HookOutput): Promise<HookResult> {
      if (!cfg.enabled) {
        return { success: true, data: { skipped: true, reason: "disabled" } }
      }

      try {
        let eventType = ctx.event?.type ?? ctx.lifecycle
        const properties = ctx.event?.properties

        // Check if we should notify for this lifecycle/event
        let shouldNotify = false

        // Task-completion notifications are emitted via chainData on todo.updated.
        if (ctx.lifecycle === "todo.updated") {
          const isTaskCompletion =
            ctx.chainData?.triggerNotification === true && ctx.chainData?.allTasksComplete === true

          if (!isTaskCompletion) {
            return { success: true, data: { skipped: true, reason: "not a completion transition" } }
          }

          eventType = "task.complete"
          shouldNotify = cfg.notifyOn.taskComplete === true
        } else if (ctx.lifecycle === "session.idle") {
          shouldNotify = cfg.notifyOn.sessionIdle === true
        } else if (ctx.lifecycle === "session.error") {
          shouldNotify = cfg.notifyOn.error === true
        } else if (eventType === "task.complete") {
          shouldNotify = cfg.notifyOn.taskComplete === true
        } else if (eventType === "error" || eventType === "session.error") {
          shouldNotify = cfg.notifyOn.error === true
        }

        if (!shouldNotify) {
          return {
            success: true,
            data: { skipped: true, reason: "event type not configured for notification" },
          }
        }

        const message = getNotificationMessage(eventType, properties)
        const sent = await notify(cfg.title, message, cfg.sound)

        return {
          success: true,
          data: {
            notificationSent: sent,
            platform: getPlatform(),
            eventType,
          },
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : String(error)
        return {
          success: false,
          error: `Notification hook failed: ${errorMessage}`,
        }
      }
    },
  }
}

/**
 * Hook registry entry for notification.
 */
export const notificationEntry = {
  name: "notification" as const,
  factory: createNotificationHook,
  defaultEnabled: true,
  description: "Sends platform-specific notifications for session events",
}
