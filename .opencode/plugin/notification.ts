import type { Plugin } from "@opencode-ai/plugin"

/**
 * Ring Notification Plugin
 * Desktop notification when sessions complete.
 *
 * Provides desktop notifications for session events,
 * useful when running long-running tasks in the background.
 *
 * Platform support:
 * - macOS: Uses osascript for native notifications
 * - Linux: Uses notify-send if available
 * - Windows: Uses PowerShell toast notifications
 *
 * SECURITY: All notification content is sanitized to prevent command injection.
 */
/**
 * Sanitize notification content to prevent command injection.
 * Removes all characters except alphanumeric, spaces, and basic punctuation.
 * Exported for testing.
 */
export function sanitizeNotificationContent(content: string, maxLength: number = 100): string {
  return content
    .replace(/[^a-zA-Z0-9 .,!?:;()\-]/g, "")
    .slice(0, maxLength)
}

export const RingNotification: Plugin = async ({ $ }) => {

  const sendNotification = async (title: string, message: string) => {
    const platform = process.platform

    // SECURITY: Sanitize all user-influenced content before shell execution
    const safeTitle = sanitizeNotificationContent(title, 50)
    const safeMessage = sanitizeNotificationContent(message, 100)

    try {
      if (platform === "darwin") {
        // macOS - Use array syntax for proper escaping
        const appleScript = `display notification "${safeMessage}" with title "${safeTitle}"`
        await $`osascript -e ${appleScript}`
      } else if (platform === "linux") {
        // Linux with notify-send - Use array syntax for proper escaping
        await $`notify-send ${[safeTitle]} ${[safeMessage]}`
      } else if (platform === "win32") {
        // Windows PowerShell toast - sanitized content prevents injection
        const script = `
          [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
          $template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
          $textNodes = $template.GetElementsByTagName("text")
          $textNodes.Item(0).AppendChild($template.CreateTextNode("${safeTitle}")) | Out-Null
          $textNodes.Item(1).AppendChild($template.CreateTextNode("${safeMessage}")) | Out-Null
          $toast = [Windows.UI.Notifications.ToastNotification]::new($template)
          [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Ring").Show($toast)
        `
        await $`powershell -Command ${script}`
      }
    } catch (err) {
      // Notification not critical - log for debugging
      if (process.env.DEBUG) {
        console.debug('[ring] Notification failed:', err)
      }
    }
  }

  return {
    event: async ({ event }) => {
      if (event.type === "session.idle") {
        await sendNotification("Ring", "Session completed!")
      }

      if (event.type === "session.error") {
        await sendNotification("Ring", "Session encountered an error")
      }
    }
  }
}

export default RingNotification
