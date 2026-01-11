/**
 * Ring Configuration Schema
 *
 * Defines the structure for Ring's layered configuration system.
 * Uses Zod for runtime validation.
 */

import { z } from "zod"

/**
 * Hook names that can be disabled.
 */
export const HookNameSchema = z.enum([
  "session-start",
  "context-injection",
  "notification",
  "task-completion",
  "session-outcome",
  "outcome-inference",
  "doubt-resolver",
  "background-notification",
  "compaction-context",
  "rules-injector",
  "agent-reminder",
])

/**
 * Agent names that can be disabled.
 */
export const AgentNameSchema = z.enum([
  "code-reviewer",
  "security-reviewer",
  "business-logic-reviewer",
  "test-reviewer",
  "nil-safety-reviewer",
  "codebase-explorer",
  "write-plan",
  "backend-engineer-golang",
  "backend-engineer-typescript",
  "frontend-engineer",
  "frontend-designer",
  "devops-engineer",
  "sre",
  "qa-analyst",
])

/**
 * Skill names that can be disabled.
 */
export const SkillNameSchema = z.enum([
  "using-ring",
  "test-driven-development",
  "systematic-debugging",
  "requesting-code-review",
  "writing-plans",
  "executing-plans",
  "dispatching-parallel-agents",
  "verification-before-completion",
  "commit",
  "brainstorm",
  "codereview",
  "lint",
  "worktree",
])

/**
 * Command names that can be disabled.
 */
export const CommandNameSchema = z.enum([
  "commit",
  "codereview",
  "brainstorm",
  "execute-plan",
  "worktree",
  "lint",
  "create-handoff",
  "resume-handoff",
  "explore-codebase",
  "interview-me",
  "write-plan",
])

/**
 * Background task configuration.
 */
export const BackgroundTaskConfigSchema = z.object({
  /** Default concurrency for background tasks */
  defaultConcurrency: z.number().min(1).max(10).default(3),
  /** Per-agent concurrency limits */
  agentConcurrency: z.record(z.string(), z.number().min(1).max(10)).optional(),
  /** Task timeout in milliseconds */
  taskTimeoutMs: z.number().min(60000).max(3600000).default(1800000),
})

/**
 * Notification configuration.
 */
export const NotificationConfigSchema = z.object({
  /** Enable desktop notifications */
  enabled: z.boolean().default(true),
  /** Notify on session idle */
  onIdle: z.boolean().default(true),
  /** Notify on session error */
  onError: z.boolean().default(true),
  /** Notify on background task completion */
  onBackgroundComplete: z.boolean().default(true),
})

/**
 * Experimental features configuration.
 */
export const ExperimentalConfigSchema = z.object({
  /** Enable preemptive compaction */
  preemptiveCompaction: z.boolean().default(false),
  /** Compaction threshold (0.5-0.95) */
  compactionThreshold: z.number().min(0.5).max(0.95).default(0.8),
  /** Enable aggressive tool output truncation */
  aggressiveTruncation: z.boolean().default(false),
})

/**
 * Main Ring configuration schema.
 */
export const RingConfigSchema = z.object({
  /** Schema URL for IDE support */
  $schema: z.string().optional(),

  /** Disabled hooks (won't be loaded) */
  disabled_hooks: z.array(HookNameSchema).default([]),

  /** Disabled agents (won't be available) */
  disabled_agents: z.array(AgentNameSchema).default([]),

  /** Disabled skills (won't be loaded) */
  disabled_skills: z.array(SkillNameSchema).default([]),

  /** Disabled commands (won't be registered) */
  disabled_commands: z.array(CommandNameSchema).default([]),

  /** Background task configuration */
  background_tasks: BackgroundTaskConfigSchema.optional().default({
    defaultConcurrency: 3,
    taskTimeoutMs: 1800000,
  }),

  /** Notification configuration */
  notifications: NotificationConfigSchema.optional().default({
    enabled: true,
    onIdle: true,
    onError: true,
    onBackgroundComplete: true,
  }),

  /** Experimental features */
  experimental: ExperimentalConfigSchema.optional().default({
    preemptiveCompaction: false,
    compactionThreshold: 0.8,
    aggressiveTruncation: false,
  }),

  /** Custom hook configurations */
  hooks: z.record(z.string(), z.record(z.string(), z.unknown())).optional(),
})

/**
 * Inferred TypeScript types from schemas.
 */
export type HookName = z.infer<typeof HookNameSchema>
export type AgentName = z.infer<typeof AgentNameSchema>
export type SkillName = z.infer<typeof SkillNameSchema>
export type CommandName = z.infer<typeof CommandNameSchema>
export type BackgroundTaskConfig = z.infer<typeof BackgroundTaskConfigSchema>
export type NotificationConfig = z.infer<typeof NotificationConfigSchema>
export type ExperimentalConfig = z.infer<typeof ExperimentalConfigSchema>
export type RingConfig = z.infer<typeof RingConfigSchema>

/**
 * Default configuration values.
 * M3: Derived from schema to ensure consistency with defaults defined in Zod.
 */
export const DEFAULT_RING_CONFIG: RingConfig = RingConfigSchema.parse({})
