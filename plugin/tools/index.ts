/**
 * Ring Tools
 *
 * Custom tools registered by Ring plugin.
 * Uses the tool() helper from @opencode-ai/plugin.
 */

import { tool } from "@opencode-ai/plugin"

/**
 * Ring Doubt Resolver Tool
 *
 * Provides a structured "pick from options" prompt when the agent has doubts.
 * This is NOT a native OpenCode UI picker - it prints a numbered list
 * and asks the user to reply normally.
 */
export const ringDoubtTool = tool({
  description:
    "Ask the user to choose from options (single/multi + free-text alternative) when the agent has doubts.",
  args: {
    question: tool.schema.string().min(1).describe("What decision is needed"),
    options: tool.schema
      .array(tool.schema.string().min(1))
      .min(1)
      .describe("List of selectable options"),
    multi: tool.schema.boolean().optional().describe("Allow multiple selections"),
    allowCustom: tool.schema
      .boolean()
      .optional()
      .describe("Allow user to type something else instead of picking"),
    context: tool.schema.string().optional().describe("Optional extra context for the decision"),
  },
  execute: async (args) => {
    const multi = Boolean(args.multi)
    const allowCustom = args.allowCustom !== false

    const header =
      "<MANDATORY-USER-MESSAGE>\nDOUBT CHECKPOINT - USER CHOICE REQUIRED\n</MANDATORY-USER-MESSAGE>"

    const lines: string[] = []
    lines.push(header)
    lines.push("")
    lines.push(args.question.trim())

    if (args.context) {
      lines.push("")
      lines.push("Context:")
      lines.push(args.context.trim())
    }

    lines.push("")
    lines.push("Options:")
    args.options.forEach((opt, i) => {
      lines.push(`${i + 1}) ${opt}`)
    })

    lines.push("")

    if (multi) {
      lines.push(`Reply with one or more numbers (example: "1" or "1,3").`)
    } else {
      lines.push(`Reply with one number (example: "2").`)
    }

    if (allowCustom) {
      lines.push('Or type your own answer (example: "Other: ...").')
    }

    return lines.join("\n")
  },
})

/**
 * All Ring tools to register with OpenCode.
 */
export const ringTools = {
  ring_doubt: ringDoubtTool,
}
