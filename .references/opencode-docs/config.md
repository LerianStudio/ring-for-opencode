---
title: "Config | OpenCode"
source: "https://opencode.ai/docs/config"
crawled: "2026-01-05T00:46:23Z"
---

[Skip to content](https://opencode.ai/docs/config#_top)

# Config

Using the OpenCode JSON config.

You can configure OpenCode using a JSON config file.

* * *

## [Format](https://opencode.ai/docs/config\#format)

OpenCode supports both **JSON** and **JSONC** (JSON with Comments) formats.

```
{

  "$schema": "https://opencode.ai/config.json",

  // Theme configuration

  "theme": "opencode",

  "model": "anthropic/claude-sonnet-4-5",

  "autoupdate": true,

}
```

* * *

## [Locations](https://opencode.ai/docs/config\#locations)

You can place your config in a couple of different locations and they have a
different order of precedence.

Configuration files are merged together, not replaced. Settings from the following config locations are combined. Where later configs override earlier ones only for conflicting keys. Non-conflicting settings from all configs are preserved.

For example, if your global config sets `theme: "opencode"` and `autoupdate: true`, and your project config sets `model: "anthropic/claude-sonnet-4-5"`, the final configuration will include all three settings.

* * *

### [Global](https://opencode.ai/docs/config\#global)

Place your global OpenCode config in `~/.config/opencode/opencode.json`. You’ll want to use the global config for things like themes, providers, or keybinds.

* * *

### [Per project](https://opencode.ai/docs/config\#per-project)

You can also add a `opencode.json` in your project. Settings from this config are merged with and can override the global config. This is useful for configuring providers or modes specific to your project.

When OpenCode starts up, it looks for a config file in the current directory or traverse up to the nearest Git directory.

This is also safe to be checked into Git and uses the same schema as the global one.

* * *

### [Custom path](https://opencode.ai/docs/config\#custom-path)

You can also specify a custom config file path using the `OPENCODE_CONFIG` environment variable.

```
export OPENCODE_CONFIG=/path/to/my/custom-config.json

opencode run "Hello world"
```

Settings from this config are merged with and **can override** the global and project configs.

* * *

### [Custom directory](https://opencode.ai/docs/config\#custom-directory)

You can specify a custom config directory using the `OPENCODE_CONFIG_DIR`
environment variable. This directory will be searched for agents, commands,
modes, and plugins just like the standard `.opencode` directory, and should
follow the same structure.

```
export OPENCODE_CONFIG_DIR=/path/to/my/config-directory

opencode run "Hello world"
```

The custom directory is loaded after the global config and `.opencode` directories, so it **can override** their settings.

* * *

## [Schema](https://opencode.ai/docs/config\#schema)

The config file has a schema that’s defined in [**`opencode.ai/config.json`**](https://opencode.ai/config.json).

Your editor should be able to validate and autocomplete based on the schema.

* * *

### [TUI](https://opencode.ai/docs/config\#tui)

You can configure TUI-specific settings through the `tui` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tui": {

    "scroll_speed": 3,

    "scroll_acceleration": {

      "enabled": true

    },

    "diff_style": "auto"

  }

}
```

Available options:

- `scroll_acceleration.enabled` \- Enable macOS-style scroll acceleration. **Takes precedence over `scroll_speed`.**
- `scroll_speed` \- Custom scroll speed multiplier (default: `1`, minimum: `1`). Ignored if `scroll_acceleration.enabled` is `true`.
- `diff_style` \- Control diff rendering. `"auto"` adapts to terminal width, `"stacked"` always shows single column.

[Learn more about using the TUI here](https://opencode.ai/docs/tui).

* * *

### [Server](https://opencode.ai/docs/config\#server)

You can configure server settings for the `opencode serve` and `opencode web` commands through the `server` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "server": {

    "port": 4096,

    "hostname": "0.0.0.0",

    "mdns": true,

    "cors": ["http://localhost:5173"]

  }

}
```

Available options:

- `port` \- Port to listen on.
- `hostname` \- Hostname to listen on. When `mdns` is enabled and no hostname is set, defaults to `0.0.0.0`.
- `mdns` \- Enable mDNS service discovery. This allows other devices on the network to discover your OpenCode server.
- `cors` \- Additional origins to allow for CORS when using the HTTP server from a browser-based client. Values must be full origins (scheme + host + optional port), eg `https://app.example.com`.

[Learn more about the server here](https://opencode.ai/docs/server).

* * *

### [Tools](https://opencode.ai/docs/config\#tools)

You can manage the tools an LLM can use through the `tools` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "tools": {

    "write": false,

    "bash": false

  }

}
```

[Learn more about tools here](https://opencode.ai/docs/tools).

* * *

### [Models](https://opencode.ai/docs/config\#models)

You can configure the providers and models you want to use in your OpenCode config through the `provider`, `model` and `small_model` options.

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {},

  "model": "anthropic/claude-sonnet-4-5",

  "small_model": "anthropic/claude-haiku-4-5"

}
```

The `small_model` option configures a separate model for lightweight tasks like title generation. By default, OpenCode tries to use a cheaper model if one is available from your provider, otherwise it falls back to your main model.

Provider options can include `timeout` and `setCacheKey`:

```
{

  "$schema": "https://opencode.ai/config.json",

  "provider": {

    "anthropic": {

      "options": {

        "timeout": 600000,

        "setCacheKey": true

      }

    }

  }

}
```

- `timeout` \- Request timeout in milliseconds (default: 300000). Set to `false` to disable.
- `setCacheKey` \- Ensure a cache key is always set for designated provider.

You can also configure [local models](https://opencode.ai/docs/models#local). [Learn more](https://opencode.ai/docs/models).

* * *

### [Themes](https://opencode.ai/docs/config\#themes)

You can configure the theme you want to use in your OpenCode config through the `theme` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "theme": ""

}
```

[Learn more here](https://opencode.ai/docs/themes).

* * *

### [Agents](https://opencode.ai/docs/config\#agents)

You can configure specialized agents for specific tasks through the `agent` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "agent": {

    "code-reviewer": {

      "description": "Reviews code for best practices and potential issues",

      "model": "anthropic/claude-sonnet-4-5",

      "prompt": "You are a code reviewer. Focus on security, performance, and maintainability.",

      "tools": {

        // Disable file modification tools for review-only agent

        "write": false,

        "edit": false,

      },

    },

  },

}
```

You can also define agents using markdown files in `~/.config/opencode/agent/` or `.opencode/agent/`. [Learn more here](https://opencode.ai/docs/agents).

* * *

### [Default agent](https://opencode.ai/docs/config\#default-agent)

You can set the default agent using the `default_agent` option. This determines which agent is used when none is explicitly specified.

```
{

  "$schema": "https://opencode.ai/config.json",

  "default_agent": "plan"

}
```

The default agent must be a primary agent (not a subagent). This can be a built-in agent like `"build"` or `"plan"`, or a [custom agent](https://opencode.ai/docs/agents) you’ve defined. If the specified agent doesn’t exist or is a subagent, OpenCode will fall back to `"build"` with a warning.

This setting applies across all interfaces: TUI, CLI (`opencode run`), desktop app, and GitHub Action.

* * *

### [Sharing](https://opencode.ai/docs/config\#sharing)

You can configure the [share](https://opencode.ai/docs/share) feature through the `share` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "share": "manual"

}
```

This takes:

- `"manual"` \- Allow manual sharing via commands (default)
- `"auto"` \- Automatically share new conversations
- `"disabled"` \- Disable sharing entirely

By default, sharing is set to manual mode where you need to explicitly share conversations using the `/share` command.

* * *

### [Commands](https://opencode.ai/docs/config\#commands)

You can configure custom commands for repetitive tasks through the `command` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "command": {

    "test": {

      "template": "Run the full test suite with coverage report and show any failures.\nFocus on the failing tests and suggest fixes.",

      "description": "Run tests with coverage",

      "agent": "build",

      "model": "anthropic/claude-haiku-4-5",

    },

    "component": {

      "template": "Create a new React component named $ARGUMENTS with TypeScript support.\nInclude proper typing and basic structure.",

      "description": "Create a new component",

    },

  },

}
```

You can also define commands using markdown files in `~/.config/opencode/command/` or `.opencode/command/`. [Learn more here](https://opencode.ai/docs/commands).

* * *

### [Keybinds](https://opencode.ai/docs/config\#keybinds)

You can customize your keybinds through the `keybinds` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "keybinds": {}

}
```

[Learn more here](https://opencode.ai/docs/keybinds).

* * *

### [Autoupdate](https://opencode.ai/docs/config\#autoupdate)

OpenCode will automatically download any new updates when it starts up. You can disable this with the `autoupdate` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "autoupdate": false

}
```

If you don’t want updates but want to be notified when a new version is available, set `autoupdate` to `"notify"`.

* * *

### [Formatters](https://opencode.ai/docs/config\#formatters)

You can configure code formatters through the `formatter` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "formatter": {

    "prettier": {

      "disabled": true

    },

    "custom-prettier": {

      "command": ["npx", "prettier", "--write", "$FILE"],

      "environment": {

        "NODE_ENV": "development"

      },

      "extensions": [".js", ".ts", ".jsx", ".tsx"]

    }

  }

}
```

[Learn more about formatters here](https://opencode.ai/docs/formatters).

* * *

### [Permissions](https://opencode.ai/docs/config\#permissions)

By default, opencode **allows all operations** without requiring explicit approval. You can change this using the `permission` option.

For example, to ensure that the `edit` and `bash` tools require user approval:

```
{

  "$schema": "https://opencode.ai/config.json",

  "permission": {

    "edit": "ask",

    "bash": "ask"

  }

}
```

[Learn more about permissions here](https://opencode.ai/docs/permissions).

* * *

### [Compaction](https://opencode.ai/docs/config\#compaction)

You can control context compaction behavior through the `compaction` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "compaction": {

    "auto": true,

    "prune": true

  }

}
```

- `auto` \- Automatically compact the session when context is full (default: `true`).
- `prune` \- Remove old tool outputs to save tokens (default: `true`).

* * *

### [Watcher](https://opencode.ai/docs/config\#watcher)

You can configure file watcher ignore patterns through the `watcher` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "watcher": {

    "ignore": ["node_modules/**", "dist/**", ".git/**"]

  }

}
```

Patterns follow glob syntax. Use this to exclude noisy directories from file watching.

* * *

### [MCP servers](https://opencode.ai/docs/config\#mcp-servers)

You can configure MCP servers you want to use through the `mcp` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "mcp": {}

}
```

[Learn more here](https://opencode.ai/docs/mcp-servers).

* * *

### [Plugins](https://opencode.ai/docs/config\#plugins)

[Plugins](https://opencode.ai/docs/plugins) extend OpenCode with custom tools, hooks, and integrations.

Place plugin files in `.opencode/plugin/` or `~/.config/opencode/plugin/`. You can also load plugins from npm through the `plugin` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "plugin": ["opencode-helicone-session", "@my-org/custom-plugin"]

}
```

[Learn more here](https://opencode.ai/docs/plugins).

* * *

### [Instructions](https://opencode.ai/docs/config\#instructions)

You can configure the instructions for the model you’re using through the `instructions` option.

```
{

  "$schema": "https://opencode.ai/config.json",

  "instructions": ["CONTRIBUTING.md", "docs/guidelines.md", ".cursor/rules/*.md"]

}
```

This takes an array of paths and glob patterns to instruction files. [Learn more\\
about rules here](https://opencode.ai/docs/rules).

* * *

### [Disabled providers](https://opencode.ai/docs/config\#disabled-providers)

You can disable providers that are loaded automatically through the `disabled_providers` option. This is useful when you want to prevent certain providers from being loaded even if their credentials are available.

```
{

  "$schema": "https://opencode.ai/config.json",

  "disabled_providers": ["openai", "gemini"]

}
```

The `disabled_providers` option accepts an array of provider IDs. When a provider is disabled:

- It won’t be loaded even if environment variables are set.
- It won’t be loaded even if API keys are configured through the `/connect` command.
- The provider’s models won’t appear in the model selection list.

* * *

### [Enabled providers](https://opencode.ai/docs/config\#enabled-providers)

You can specify an allowlist of providers through the `enabled_providers` option. When set, only the specified providers will be enabled and all others will be ignored.

```
{

  "$schema": "https://opencode.ai/config.json",

  "enabled_providers": ["anthropic", "openai"]

}
```

This is useful when you want to restrict OpenCode to only use specific providers rather than disabling them one by one.

If a provider appears in both `enabled_providers` and `disabled_providers`, the `disabled_providers` takes priority for backwards compatibility.

* * *

### [Experimental](https://opencode.ai/docs/config\#experimental)

The `experimental` key contains options that are under active development.

```
{

  "$schema": "https://opencode.ai/config.json",

  "experimental": {}

}
```

* * *

## [Variables](https://opencode.ai/docs/config\#variables)

You can use variable substitution in your config files to reference environment variables and file contents.

* * *

### [Env vars](https://opencode.ai/docs/config\#env-vars)

Use `{env:VARIABLE_NAME}` to substitute environment variables:

```
{

  "$schema": "https://opencode.ai/config.json",

  "model": "{env:OPENCODE_MODEL}",

  "provider": {

    "anthropic": {

      "models": {},

      "options": {

        "apiKey": "{env:ANTHROPIC_API_KEY}"

      }

    }

  }

}
```

If the environment variable is not set, it will be replaced with an empty string.

* * *

### [Files](https://opencode.ai/docs/config\#files)

Use `{file:path/to/file}` to substitute the contents of a file:

```
{

  "$schema": "https://opencode.ai/config.json",

  "instructions": ["./custom-instructions.md"],

  "provider": {

    "openai": {

      "options": {

        "apiKey": "{file:~/.secrets/openai-key}"

      }

    }

  }

}
```

File paths can be:

- Relative to the config file directory
- Or absolute paths starting with `/` or `~`

These are useful for:

- Keeping sensitive data like API keys in separate files.
- Including large instruction files without cluttering your config.
- Sharing common configuration snippets across multiple config files.
