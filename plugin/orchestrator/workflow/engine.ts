/**
 * Workflow Engine
 *
 * Orchestrates multi-step workflows with:
 * - Structured handoff section extraction
 * - Per-section character budgets
 * - Intelligent context compaction
 * - Dependency injection for testing
 *
 * Ported from Orchestra's workflow engine pattern.
 */

import { randomUUID } from "node:crypto"
import { existsSync, lstatSync, realpathSync } from "node:fs"
import { isAbsolute, resolve, sep } from "node:path"
import type {
  WorkflowDefinition,
  WorkflowRunInput,
  WorkflowRunLimits,
  WorkflowRunResult,
  WorkflowStepDefinition,
  WorkflowStepResult,
} from "../types.js"

const workflows = new Map<string, WorkflowDefinition>()

export function resetWorkflows(): void {
  workflows.clear()
}

export function registerWorkflow(def: WorkflowDefinition): void {
  workflows.set(def.id, def)
}

export function listWorkflows(): WorkflowDefinition[] {
  return [...workflows.values()].sort((a, b) => a.id.localeCompare(b.id))
}

export function getWorkflow(id: string): WorkflowDefinition | undefined {
  return workflows.get(id)
}

export type WorkflowRunDependencies = {
  resolveWorker: (workerId: string, autoSpawn: boolean) => Promise<string>
  sendToWorker: (
    workerId: string,
    message: string,
    options: { attachments?: WorkflowRunInput["attachments"]; timeoutMs: number },
  ) => Promise<{ success: boolean; response?: string; warning?: string; error?: string }>
  projectRoot?: string
}

const handoffSections = ["Summary", "Actions", "Artifacts", "Risks", "Next"] as const
const carrySections = ["Summary", "Artifacts", "Risks", "Next"] as const

type HandoffSection = (typeof handoffSections)[number]
type CarrySection = (typeof carrySections)[number]

const handoffSectionMap = new Map(handoffSections.map((s) => [s.toLowerCase(), s]))
const carrySectionCaps: Record<CarrySection, number> = {
  Summary: 900,
  Artifacts: 1600,
  Risks: 900,
  Next: 900,
}

function normalizeSectionName(value: string): HandoffSection | undefined {
  return handoffSectionMap.get(value.trim().toLowerCase())
}

type WorkflowAttachment = NonNullable<WorkflowRunInput["attachments"]>[number]

function validateAttachmentPath(path: string, projectRoot: string): string {
  if (isAbsolute(path)) {
    throw new Error("Attachment path must be relative")
  }

  if (path.includes("..")) {
    throw new Error("Path traversal not allowed in attachment paths")
  }

  const resolved = resolve(projectRoot, path)

  if (!existsSync(resolved)) {
    throw new Error(`Attachment file not found: ${path}`)
  }

  const stat = lstatSync(resolved)
  const realRoot = `${realpathSync(projectRoot)}${sep}`
  const realResolved = realpathSync(resolved)

  if (!realResolved.startsWith(realRoot)) {
    throw new Error("Attachment path must be within project directory")
  }

  if (stat.isSymbolicLink()) {
    throw new Error("Symlink attachments are not allowed")
  }

  return realResolved
}

function sanitizeAttachments(
  attachments: WorkflowRunInput["attachments"] | undefined,
  projectRoot?: string,
): WorkflowRunInput["attachments"] | undefined {
  if (!attachments || attachments.length === 0) return attachments

  const resolvedRoot = projectRoot ?? process.cwd()
  const sanitized: WorkflowAttachment[] = []

  for (const attachment of attachments) {
    if (attachment.path) {
      try {
        const resolvedPath = validateAttachmentPath(attachment.path, resolvedRoot)
        sanitized.push({ ...attachment, path: resolvedPath })
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err)
        throw new Error(`Invalid attachment path: ${message}`)
      }
    } else {
      sanitized.push(attachment)
    }
  }

  return sanitized
}

function extractHandoffSections(text: string): Record<HandoffSection, string> {
  const sections: Record<HandoffSection, string[]> = {
    Summary: [],
    Actions: [],
    Artifacts: [],
    Risks: [],
    Next: [],
  }

  const trimmed = text.trim()
  if (!trimmed) {
    return { Summary: "", Actions: "", Artifacts: "", Risks: "", Next: "" }
  }

  const lines = trimmed.split(/\r?\n/)
  let current: HandoffSection | undefined
  let sawHeading = false
  const headingRegex = /^\s*(#{1,3}\s*)?(Summary|Actions|Artifacts|Risks|Next)\s*:?\s*$/i

  for (const line of lines) {
    const match = line.match(headingRegex)
    if (match) {
      const key = normalizeSectionName(match[2] ?? "")
      if (key) {
        current = key
        sawHeading = true
        continue
      }
    }
    if (current) sections[current].push(line)
  }

  const resolved: Record<HandoffSection, string> = {
    Summary: sections.Summary.join("\n").trim(),
    Actions: sections.Actions.join("\n").trim(),
    Artifacts: sections.Artifacts.join("\n").trim(),
    Risks: sections.Risks.join("\n").trim(),
    Next: sections.Next.join("\n").trim(),
  }

  if (!sawHeading) {
    resolved.Summary = trimmed
  } else if (!resolved.Summary && resolved.Actions) {
    resolved.Summary = resolved.Actions
  }

  return resolved
}

function truncateText(text: string, maxChars: number): string {
  const trimmed = text.trim()
  if (trimmed.length <= maxChars) return trimmed
  const suffix = "\n...(truncated)"
  const sliceEnd = Math.max(0, maxChars - suffix.length)
  return `${trimmed.slice(0, sliceEnd).trimEnd()}${suffix}`
}

function compactCarrySections(
  sections: Record<HandoffSection, string>,
  maxCarryChars: number,
): { sections: Record<CarrySection, string>; truncatedSections: CarrySection[] } {
  const totalCaps = Object.values(carrySectionCaps).reduce((sum, v) => sum + v, 0)
  const scale = Math.min(1, maxCarryChars / (totalCaps + 200))
  const compacted = {} as Record<CarrySection, string>
  const truncatedSections: CarrySection[] = []

  for (const section of carrySections) {
    const baseCap = carrySectionCaps[section]
    const cap = Math.max(60, Math.floor(baseCap * scale))
    const content = sections[section] ?? ""
    if (content.length > cap) truncatedSections.push(section)
    compacted[section] = content ? truncateText(content, cap) : ""
  }

  return { sections: compacted, truncatedSections }
}

function formatCarryBlock(
  stepTitle: string,
  responseText: string,
  maxCarryChars: number,
): { text: string; truncated: boolean; truncatedSections: CarrySection[] } {
  const sections = extractHandoffSections(responseText)
  const compacted = compactCarrySections(sections, maxCarryChars)
  const truncatedSections = [...compacted.truncatedSections]
  let truncated = truncatedSections.length > 0

  const blocks = carrySections
    .map((section) => {
      const content = compacted.sections[section]
      if (!content) return ""
      return `#### ${section}\n${content}`
    })
    .filter(Boolean)

  if (blocks.length === 0) {
    const fallback = truncateText(responseText, Math.max(240, Math.floor(maxCarryChars / 4)))
    if (fallback.length < responseText.trim().length) truncated = true
    blocks.push(`#### Summary\n${fallback || "None"}`)
  }

  const block = `### ${stepTitle}\n${blocks.join("\n\n")}`
  if (block.length <= maxCarryChars) {
    return { text: block, truncated, truncatedSections }
  }

  const reducedCap = Math.max(60, Math.floor(maxCarryChars / (blocks.length * 2)))
  const reducedBlocks = blocks.map((b) => truncateText(b, reducedCap)).join("\n\n")
  const reduced = `### ${stepTitle}\n${reducedBlocks}`
  const finalText = reduced.length <= maxCarryChars ? reduced : truncateText(reduced, maxCarryChars)

  return { text: finalText, truncated: true, truncatedSections }
}

function splitCarryBlocks(carry: string): string[] {
  const trimmed = carry.trim()
  if (!trimmed) return []
  if (!trimmed.includes("### ")) return [trimmed]
  return trimmed
    .split(/\n(?=###\s)/g)
    .map((b) => b.trim())
    .filter(Boolean)
}

function appendCarry(
  existing: string,
  next: string,
  maxChars: number,
): { text: string; droppedBlocks: number; truncated: boolean } {
  const blocks = [...splitCarryBlocks(existing), next].filter(Boolean)
  if (blocks.length === 0) return { text: "", droppedBlocks: 0, truncated: false }

  const originalCount = blocks.length
  while (blocks.join("\n\n").length > maxChars && blocks.length > 1) {
    blocks.shift()
  }

  const combined = blocks.join("\n\n")
  const droppedBlocks = Math.max(0, originalCount - blocks.length)

  if (combined.length <= maxChars) {
    return { text: combined, droppedBlocks, truncated: droppedBlocks > 0 }
  }

  return {
    text: truncateText(combined, maxChars),
    droppedBlocks,
    truncated: true,
  }
}

function applyTemplate(template: string, vars: Record<string, string>): string {
  let out = template
  for (const [key, value] of Object.entries(vars)) {
    out = out.replaceAll(`{${key}}`, value)
  }
  return out
}

function resolveStepTimeout(step: WorkflowStepDefinition, limits: WorkflowRunLimits): number {
  const requested =
    typeof step.timeoutMs === "number" && Number.isFinite(step.timeoutMs) && step.timeoutMs > 0
      ? step.timeoutMs
      : limits.perStepTimeoutMs

  return Math.max(1000, Math.min(requested, limits.perStepTimeoutMs))
}

export function validateWorkflowInput(input: WorkflowRunInput, workflow: WorkflowDefinition): void {
  if (input.task.length > input.limits.maxTaskChars) {
    throw new Error(`Task exceeds maxTaskChars (${input.limits.maxTaskChars}).`)
  }
  if (workflow.steps.length > input.limits.maxSteps) {
    throw new Error(
      `Workflow has ${workflow.steps.length} steps (maxSteps=${input.limits.maxSteps}).`,
    )
  }
}

export async function executeWorkflowStep(
  input: {
    runId: string
    workflow: WorkflowDefinition
    stepIndex: number
    task: string
    carry: string
    autoSpawn: boolean
    limits: WorkflowRunLimits
    attachments?: WorkflowRunInput["attachments"]
  },
  deps: WorkflowRunDependencies,
): Promise<{ step: WorkflowStepResult; response?: string; carry: string }> {
  if (input.stepIndex < 0 || input.stepIndex >= input.workflow.steps.length) {
    throw new Error(
      `Invalid stepIndex ${input.stepIndex} for workflow "${input.workflow.id}" ` +
        `with ${input.workflow.steps.length} steps`,
    )
  }

  const step = input.workflow.steps[input.stepIndex]

  const stepStarted = Date.now()

  const workerId = await deps.resolveWorker(step.workerId, input.autoSpawn)
  const prompt = applyTemplate(step.prompt, { task: input.task, carry: input.carry })
  const attachments =
    input.stepIndex === 0 ? sanitizeAttachments(input.attachments, deps.projectRoot) : undefined

  const res = await deps.sendToWorker(workerId, prompt, {
    attachments,
    timeoutMs: resolveStepTimeout(step, input.limits),
  })

  const stepFinished = Date.now()

  if (!res.success) {
    const result: WorkflowStepResult = {
      id: step.id,
      title: step.title,
      workerId,
      status: "error",
      error: res.error ?? "unknown_error",
      startedAt: stepStarted,
      finishedAt: stepFinished,
      durationMs: stepFinished - stepStarted,
    }
    return { step: result, carry: input.carry }
  }

  const response = res.response ?? ""
  const result: WorkflowStepResult = {
    id: step.id,
    title: step.title,
    workerId,
    status: "success",
    response,
    ...(res.warning ? { warning: res.warning } : {}),
    startedAt: stepStarted,
    finishedAt: stepFinished,
    durationMs: stepFinished - stepStarted,
  }

  if (step.carry) {
    const carryBlock = formatCarryBlock(step.title, response, input.limits.maxCarryChars)
    const appended = appendCarry(input.carry, carryBlock.text, input.limits.maxCarryChars)
    return { step: result, response, carry: appended.text }
  }

  return { step: result, response, carry: input.carry }
}

export async function runWorkflow(
  input: WorkflowRunInput,
  deps: WorkflowRunDependencies,
): Promise<WorkflowRunResult> {
  const workflow = getWorkflow(input.workflowId)
  if (!workflow) {
    throw new Error(`Unknown workflow "${input.workflowId}".`)
  }

  validateWorkflowInput(input, workflow)

  const runId = randomUUID()
  const startedAt = Date.now()
  const steps: WorkflowStepResult[] = []
  let carry = ""
  let status: WorkflowRunResult["status"] = "running"

  for (let i = 0; i < workflow.steps.length; i++) {
    const executed = await executeWorkflowStep(
      {
        runId,
        workflow,
        stepIndex: i,
        task: input.task,
        carry,
        autoSpawn: input.autoSpawn ?? true,
        limits: input.limits,
        attachments: input.attachments,
      },
      deps,
    )
    steps.push(executed.step)

    if (executed.step.status === "error") {
      status = "error"
      break
    }
    carry = executed.carry
  }

  const finishedAt = Date.now()

  return {
    runId,
    workflowId: workflow.id,
    workflowName: workflow.name,
    status: status === "error" ? "error" : "success",
    startedAt,
    finishedAt,
    currentStepIndex: Math.min(steps.length, workflow.steps.length),
    steps,
    lastStepResult: steps.at(-1),
  }
}
