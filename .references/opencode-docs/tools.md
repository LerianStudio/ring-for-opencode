---
title: "Tools | OpenCode"
source: "https://opencode.ai/docs/tools/"
crawled: "2026-01-05T00:46:23Z"
---

[Skip to content](https://opencode.ai/docs/tools/#_top)

# Tools

Manage the tools an LLM can use.

Tools allow the LLM to perform actions in your codebase. OpenCode comes with a set of built-in tools, but you can extend it with [custom tools](https://opencode.ai/docs/custom-tools) or [MCP servers](https://opencode.ai/docs/mcp-servers).

By default, all tools are **enabled** and don’t need permission to run. But you can configure this and control the [permissions](https://opencode.ai/docs/permissions) through your config.

* * *

## [Configure](https://opencode.ai/docs/tools/\#configure)

You can configure tools globally or per agent. Agent-specific configs override global settings.

By default, all tools are set to `true`. To disable a tool, set it to `false`.

* * *

### [Global](https://opencode.ai/docs/tools/\#global)

Disable or enable tools globally using the `tools` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "write": false,

    "bash": false,

    "webfetch": true

  }

}
```

You can also use wildcards to control multiple tools at once. For example, to disable all tools from an MCP server:

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "mymcp_*": false

  }

}
```

* * *

### [Per agent](https://opencode.ai/docs/tools/\#per-agent)

Override global tool settings for specific agents using the `tools` config in the agent definition.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "write": true,

    "bash": true

  },

  "agent": {

    "plan": {

      "tools": {

        "write": false,

        "bash": false

      }

    }

  }

}
```

For example, here the `plan` agent overrides the global config to disable `write` and `bash` tools.

You can also configure tools for agents in Markdown.

```
---

description: Read-only analysis agent

mode: subagent

tools:

  write: false

  edit: false

  bash: false

---

Analyze code without making any modifications.
```

[Learn more](https://opencode.ai/docs/agents#tools) about configuring tools per agent.

* * *

## [Built-in](https://opencode.ai/docs/tools/\#built-in)

Here are all the built-in tools available in OpenCode.

* * *

### [bash](https://opencode.ai/docs/tools/\#bash)

Execute shell commands in your project environment.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "bash": true

  }

}
```

This tool allows the LLM to run terminal commands like `npm install`, `git status`, or any other shell command.

* * *

### [edit](https://opencode.ai/docs/tools/\#edit)

Modify existing files using exact string replacements.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "edit": true

  }

}
```

This tool performs precise edits to files by replacing exact text matches. It’s the primary way the LLM modifies code.

* * *

### [write](https://opencode.ai/docs/tools/\#write)

Create new files or overwrite existing ones.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "write": true

  }

}
```

Use this to allow the LLM to create new files. It will overwrite existing files if they already exist.

* * *

### [read](https://opencode.ai/docs/tools/\#read)

Read file contents from your codebase.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "read": true

  }

}
```

This tool reads files and returns their contents. It supports reading specific line ranges for large files.

* * *

### [grep](https://opencode.ai/docs/tools/\#grep)

Search file contents using regular expressions.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "grep": true

  }

}
```

Fast content search across your codebase. Supports full regex syntax and file pattern filtering.

* * *

### [glob](https://opencode.ai/docs/tools/\#glob)

Find files by pattern matching.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "glob": true

  }

}
```

Search for files using glob patterns like `**/*.js` or `src/**/*.ts`. Returns matching file paths sorted by modification time.

* * *

### [list](https://opencode.ai/docs/tools/\#list)

List files and directories in a given path.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "list": true

  }

}
```

This tool lists directory contents. It accepts glob patterns to filter results.

* * *

### [lsp (experimental)](https://opencode.ai/docs/tools/\#lsp-experimental)

Interact with your configured LSP servers to get code intelligence features like definitions, references, hover info, and call hierarchy.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "lsp": true

  }

}
```

Supported operations include `goToDefinition`, `findReferences`, `hover`, `documentSymbol`, `workspaceSymbol`, `goToImplementation`, `prepareCallHierarchy`, `incomingCalls`, and `outgoingCalls`.

To configure which LSP servers are available for your project, see [LSP Servers](https://opencode.ai/docs/lsp).

* * *

### [patch](https://opencode.ai/docs/tools/\#patch)

Apply patches to files.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "patch": true

  }

}
```

This tool applies patch files to your codebase. Useful for applying diffs and patches from various sources.

* * *

### [skill](https://opencode.ai/docs/tools/\#skill)

Load a [skill](https://opencode.ai/docs/skills) (a `SKILL.md` file) and return its content in the conversation.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "skill": true

  }

}
```

You can control approval prompts for loading skills via [permissions](https://opencode.ai/docs/permissions) using `permission.skill`.

* * *

### [todowrite](https://opencode.ai/docs/tools/\#todowrite)

Manage todo lists during coding sessions.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "todowrite": true

  }

}
```

Creates and updates task lists to track progress during complex operations. The LLM uses this to organize multi-step tasks.

* * *

### [todoread](https://opencode.ai/docs/tools/\#todoread)

Read existing todo lists.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "todoread": true

  }

}
```

Reads the current todo list state. Used by the LLM to track what tasks are pending or completed.

* * *

### [webfetch](https://opencode.ai/docs/tools/\#webfetch)

Fetch web content.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "webfetch": true

  }

}
```

Allows the LLM to fetch and read web pages. Useful for looking up documentation or researching online resources.

* * *

## [Custom tools](https://opencode.ai/docs/tools/\#custom-tools)

Custom tools let you define your own functions that the LLM can call. These are defined in your config file and can execute arbitrary code.

[Learn more](https://opencode.ai/docs/custom-tools) about creating custom tools.

* * *

## [MCP servers](https://opencode.ai/docs/tools/\#mcp-servers)

MCP (Model Context Protocol) servers allow you to integrate external tools and services. This includes database access, API integrations, and third-party services.

[Learn more](https://opencode.ai/docs/mcp-servers) about configuring MCP servers.

* * *

## [Internals](https://opencode.ai/docs/tools/\#internals)

Internally, tools like `grep`, `glob`, and `list` use [ripgrep](https://github.com/BurntSushi/ripgrep) under the hood. By default, ripgrep respects `.gitignore` patterns, which means files and directories listed in your `.gitignore` will be excluded from searches and listings.

* * *

### [Ignore patterns](https://opencode.ai/docs/tools/\#ignore-patterns)

To include files that would normally be ignored, create a `.ignore` file in your project root. This file can explicitly allow certain paths.

```
!node_modules/

!dist/

!build/
```

For example, this `.ignore` file allows ripgrep to search within `node_modules/`, `dist/`, and `build/` directories even if they’re listed in `.gitignore`.
