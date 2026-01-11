import { z } from "zod"

// Permission value types
const PermissionValue = z.enum(["ask", "allow", "deny"])

const BashPermission = z.union([PermissionValue, z.record(z.string(), PermissionValue)])

// Agent permission schema
const AgentPermissionSchema = z.object({
  edit: PermissionValue.optional(),
  bash: BashPermission.optional(),
  webfetch: PermissionValue.optional(),
  doom_loop: PermissionValue.optional(),
  external_directory: PermissionValue.optional(),
})

// Skill permission schema
const SkillPermissionSchema = z.union([PermissionValue, z.record(z.string(), PermissionValue)])

// Global permission schema
const PermissionSchema = z.object({
  skill: SkillPermissionSchema.optional(),
  edit: PermissionValue.optional(),
  bash: BashPermission.optional(),
  webfetch: PermissionValue.optional(),
  doom_loop: PermissionValue.optional(),
  external_directory: PermissionValue.optional(),
})

// Agent mode
const AgentMode = z.enum(["primary", "subagent", "all"])

// Agent configuration schema
const AgentConfigSchema = z.object({
  mode: AgentMode.optional(),
  model: z.string().optional(),
  temperature: z.number().min(0).max(2).optional(),
  top_p: z.number().min(0).max(1).optional(),
  prompt: z.string().optional(),
  prompt_append: z.string().optional(),
  tools: z.record(z.string(), z.boolean()).optional(),
  disable: z.boolean().optional(),
  description: z.string().optional(),
  color: z
    .string()
    .regex(/^#[0-9A-Fa-f]{6}$/)
    .optional(),
  permission: AgentPermissionSchema.optional(),
})

// Agents configuration (keyed by agent name)
const AgentsConfigSchema = z.record(z.string(), AgentConfigSchema)

// Hook names specific to Ring
export const RingHookNameSchema = z.enum([
  "session-start",
  "context-injection",
  "notification",
  "task-completion-check",
  "session-outcome",
  "outcome-inference",
  "doubt-resolver",
])

// Skill source schema
const SkillSourceSchema = z.union([
  z.string(),
  z.object({
    path: z.string(),
    recursive: z.boolean().optional(),
    glob: z.string().optional(),
  }),
])

// Skill definition schema
const SkillDefinitionSchema = z.object({
  description: z.string().optional(),
  template: z.string().optional(),
  from: z.string().optional(),
  model: z.string().optional(),
  agent: z.string().optional(),
  subtask: z.boolean().optional(),
  "argument-hint": z.string().optional(),
  license: z.string().optional(),
  compatibility: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  "allowed-tools": z.array(z.string()).optional(),
  disable: z.boolean().optional(),
})

const SkillEntrySchema = z.union([z.boolean(), SkillDefinitionSchema])

// Skills configuration - special keys separated from skill entries
// This allows: { "my-skill": true, "sources": ["/path"], "enable": ["x"], "disable": ["y"] }
const SkillsConfigSchema = z.union([
  z.array(z.string()),
  z
    .object({
      sources: z.array(SkillSourceSchema).optional(),
      enable: z.array(z.string()).optional(),
      disable: z.array(z.string()).optional(),
    })
    .catchall(SkillEntrySchema),
])

// State directory configuration
const StateConfigSchema = z.object({
  directory: z.string().optional(),
  session_tracking: z.boolean().optional(),
  context_warnings: z.boolean().optional(),
})

// Notification configuration
const NotificationConfigSchema = z.object({
  enabled: z.boolean().optional(),
  sound: z.boolean().optional(),
  on_completion: z.boolean().optional(),
  on_error: z.boolean().optional(),
})

// Main Ring OpenCode configuration schema
export const RingOpenCodeConfigSchema = z.object({
  $schema: z.string().optional(),
  version: z.string().optional(),
  name: z.string().optional(),
  description: z.string().optional(),
  permission: PermissionSchema.optional(),
  agent: AgentsConfigSchema.optional(),
  disabled_hooks: z.array(RingHookNameSchema).optional(),
  skills: SkillsConfigSchema.optional(),
  state: StateConfigSchema.optional(),
  notification: NotificationConfigSchema.optional(),
})

// Type exports
export type RingOpenCodeConfig = z.infer<typeof RingOpenCodeConfigSchema>
export type AgentConfig = z.infer<typeof AgentConfigSchema>
export type AgentPermission = z.infer<typeof AgentPermissionSchema>
export type Permission = z.infer<typeof PermissionSchema>
export type SkillsConfig = z.infer<typeof SkillsConfigSchema>
export type SkillDefinition = z.infer<typeof SkillDefinitionSchema>
export type StateConfig = z.infer<typeof StateConfigSchema>
export type NotificationConfig = z.infer<typeof NotificationConfigSchema>
export type RingHookName = z.infer<typeof RingHookNameSchema>
