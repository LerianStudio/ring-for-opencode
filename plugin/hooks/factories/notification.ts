/**
 * Notification Hook Factory
 *
 * Provides platform-specific notifications for session events.
 * Supports macOS, Linux, and Windows notification systems.
 */

import { execFile } from "node:child_process"
import { promisify } from "node:util"
import type { Hook, HookContext, HookFactory, HookOutput, HookResult } from "../types.js"

const execFileAsync = promisify(execFile)

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
}

/** Default configuration */
const DEFAULT_CONFIG: Required<NotificationConfig> = {
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
 * Send notification on macOS using osascript.
 */
async function notifyDarwin(title: string, message: string, sound: boolean): Promise<boolean> {
  try {
    const soundPart = sound ? 'sound name "Glass"' : ""
    const script = `display notification "${escapeAppleScript(message)}" with title "${escapeAppleScript(title)}" ${soundPart}`

    // Use execFile with array arguments to prevent command injection
    await execFileAsync("osascript", ["-e", script])
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
    await execFileAsync("notify-send", [title, message])
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
  // Deep merge to preserve nested config properties
  const cfg = {
    ...DEFAULT_CONFIG,
    ...config,
    notifyOn: { ...DEFAULT_CONFIG.notifyOn, ...config?.notifyOn },
  }

  return {
    name: "notification",
    lifecycles: ["session.idle", "event"],
    priority: 200, // Run late, after main processing
    enabled: cfg.enabled,

    async execute(ctx: HookContext, _output: HookOutput): Promise<HookResult> {
      if (!cfg.enabled) {
        return { success: true, data: { skipped: true, reason: "disabled" } }
      }

      try {
        const eventType = ctx.event?.type ?? ctx.lifecycle
        const properties = ctx.event?.properties

        // Check if we should notify for this event type
        let shouldNotify = false

        if (ctx.lifecycle === "session.idle" && cfg.notifyOn.sessionIdle) {
          shouldNotify = true
        } else if (eventType === "task.complete" && cfg.notifyOn.taskComplete) {
          shouldNotify = true
        } else if (eventType === "error" && cfg.notifyOn.error) {
          shouldNotify = true
        }

        if (!shouldNotify) {
          return {
            success: true,
            data: { skipped: true, reason: "event type not configured for notification" },
          }
        }

        const message = getNotificationMessage(eventType, properties)
        const sent = await sendNotification(cfg.title, message, cfg.sound)

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
