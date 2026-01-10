---
title: "CLI | OpenCode"
source: "https://opencode.ai/docs/cli/"
crawled: "2026-01-05T00:46:23Z"
---

[Skip to content](https://opencode.ai/docs/cli/#_top)

# CLI

OpenCode CLI options and commands.

The OpenCode CLI by default starts the [TUI](https://opencode.ai/docs/tui) when run without any arguments.

```
opencode
```

But it also accepts commands as documented on this page. This allows you to interact with OpenCode programmatically.

```
opencode run "Explain how closures work in JavaScript"
```

* * *

### [tui](https://opencode.ai/docs/cli/\#tui)

Start the OpenCode terminal user interface.

```
opencode [project]
```

#### [Flags](https://opencode.ai/docs/cli/\#flags)

| Flag | Short | Description |
| --- | --- | --- |
| `--continue` | `-c` | Continue the last session |
| `--session` | `-s` | Session ID to continue |
| `--prompt` |  | Prompt to use |
| `--model` | `-m` | Model to use in the form of provider/model |
| `--agent` |  | Agent to use |
| `--port` |  | Port to listen on |
| `--hostname` |  | Hostname to listen on |

* * *

## [Commands](https://opencode.ai/docs/cli/\#commands)

The OpenCode CLI also has the following commands.

* * *

### [agent](https://opencode.ai/docs/cli/\#agent)

Manage agents for OpenCode.

```
opencode agent [command]
```

* * *

### [attach](https://opencode.ai/docs/cli/\#attach)

Attach a terminal to an already running OpenCode backend server started via `serve` or `web` commands.

```
opencode attach [url]
```

This allows using the TUI with a remote OpenCode backend. For example:

```
# Start the backend server for web/mobile access

opencode web --port 4096 --hostname 0.0.0.0

# In another terminal, attach the TUI to the running backend

opencode attach http://10.20.30.40:4096
```

#### [Flags](https://opencode.ai/docs/cli/\#flags-1)

| Flag | Short | Description |
| --- | --- | --- |
| `--dir` |  | Working directory to start TUI in |
| `--session` | `-s` | Session ID to continue |

* * *

#### [create](https://opencode.ai/docs/cli/\#create)

Create a new agent with custom configuration.

```
opencode agent create
```

This command will guide you through creating a new agent with a custom system prompt and tool configuration.

* * *

#### [list](https://opencode.ai/docs/cli/\#list)

List all available agents.

```
opencode agent list
```

* * *

### [auth](https://opencode.ai/docs/cli/\#auth)

Command to manage credentials and login for providers.

```
opencode auth [command]
```

* * *

#### [login](https://opencode.ai/docs/cli/\#login)

OpenCode is powered by the provider list at [Models.dev](https://models.dev/), so you can use `opencode auth login` to configure API keys for any provider you’d like to use. This is stored in `~/.local/share/opencode/auth.json`.

```
opencode auth login
```

When OpenCode starts up it loads the providers from the credentials file. And if there are any keys defined in your environments or a `.env` file in your project.

* * *

#### [list](https://opencode.ai/docs/cli/\#list-1)

Lists all the authenticated providers as stored in the credentials file.

```
opencode auth list
```

Or the short version.

```
opencode auth ls
```

* * *

#### [logout](https://opencode.ai/docs/cli/\#logout)

Logs you out of a provider by clearing it from the credentials file.

```
opencode auth logout
```

* * *

### [github](https://opencode.ai/docs/cli/\#github)

Manage the GitHub agent for repository automation.

```
opencode github [command]
```

* * *

#### [install](https://opencode.ai/docs/cli/\#install)

Install the GitHub agent in your repository.

```
opencode github install
```

This sets up the necessary GitHub Actions workflow and guides you through the configuration process. [Learn more](https://opencode.ai/docs/github).

* * *

#### [run](https://opencode.ai/docs/cli/\#run)

Run the GitHub agent. This is typically used in GitHub Actions.

```
opencode github run
```

##### [Flags](https://opencode.ai/docs/cli/\#flags-2)

| Flag | Description |
| --- | --- |
| `--event` | GitHub mock event to run the agent for |
| `--token` | GitHub personal access token |

* * *

### [mcp](https://opencode.ai/docs/cli/\#mcp)

Manage Model Context Protocol servers.

```
opencode mcp [command]
```

* * *

#### [add](https://opencode.ai/docs/cli/\#add)

Add an MCP server to your configuration.

```
opencode mcp add
```

This command will guide you through adding either a local or remote MCP server.

* * *

#### [list](https://opencode.ai/docs/cli/\#list-2)

List all configured MCP servers and their connection status.

```
opencode mcp list
```

Or use the short version.

```
opencode mcp ls
```

* * *

#### [auth](https://opencode.ai/docs/cli/\#auth-1)

Authenticate with an OAuth-enabled MCP server.

```
opencode mcp auth [name]
```

If you don’t provide a server name, you’ll be prompted to select from available OAuth-capable servers.

You can also list OAuth-capable servers and their authentication status.

```
opencode mcp auth list
```

Or use the short version.

```
opencode mcp auth ls
```

* * *

#### [logout](https://opencode.ai/docs/cli/\#logout-1)

Remove OAuth credentials for an MCP server.

```
opencode mcp logout [name]
```

* * *

#### [debug](https://opencode.ai/docs/cli/\#debug)

Debug OAuth connection issues for an MCP server.

```
opencode mcp debug <name>
```

* * *

### [models](https://opencode.ai/docs/cli/\#models)

List all available models from configured providers.

```
opencode models [provider]
```

This command displays all models available across your configured providers in the format `provider/model`.

This is useful for figuring out the exact model name to use in [your config](https://opencode.ai/docs/config/).

You can optionally pass a provider ID to filter models by that provider.

```
opencode models anthropic
```

#### [Flags](https://opencode.ai/docs/cli/\#flags-3)

| Flag | Description |
| --- | --- |
| `--refresh` | Refresh the models cache from models.dev |
| `--verbose` | Use more verbose model output (includes metadata like costs) |

Use the `--refresh` flag to update the cached model list. This is useful when new models have been added to a provider and you want to see them in OpenCode.

```
opencode models --refresh
```

* * *

### [run](https://opencode.ai/docs/cli/\#run-1)

Run opencode in non-interactive mode by passing a prompt directly.

```
opencode run [message..]
```

This is useful for scripting, automation, or when you want a quick answer without launching the full TUI. For example.

```
opencode run Explain the use of context in Go
```

You can also attach to a running `opencode serve` instance to avoid MCP server cold boot times on every run:

```
# Start a headless server in one terminal

opencode serve

# In another terminal, run commands that attach to it

opencode run --attach http://localhost:4096 "Explain async/await in JavaScript"
```

#### [Flags](https://opencode.ai/docs/cli/\#flags-4)

| Flag | Short | Description |
| --- | --- | --- |
| `--command` |  | The command to run, use message for args |
| `--continue` | `-c` | Continue the last session |
| `--session` | `-s` | Session ID to continue |
| `--share` |  | Share the session |
| `--model` | `-m` | Model to use in the form of provider/model |
| `--agent` |  | Agent to use |
| `--file` | `-f` | File(s) to attach to message |
| `--format` |  | Format: default (formatted) or json (raw JSON events) |
| `--title` |  | Title for the session (uses truncated prompt if no value provided) |
| `--attach` |  | Attach to a running opencode server (e.g., [http://localhost:4096](http://localhost:4096/)) |
| `--port` |  | Port for the local server (defaults to random port) |

* * *

### [serve](https://opencode.ai/docs/cli/\#serve)

Start a headless OpenCode server for API access. Check out the [server docs](https://opencode.ai/docs/server) for the full HTTP interface.

```
opencode serve
```

This starts an HTTP server that provides API access to opencode functionality without the TUI interface.

#### [Flags](https://opencode.ai/docs/cli/\#flags-5)

| Flag | Description |
| --- | --- |
| `--port` | Port to listen on |
| `--hostname` | Hostname to listen on |
| `--mdns` | Enable mDNS discovery |
| `--cors` | Additional browser origin(s) to allow CORS |

* * *

### [session](https://opencode.ai/docs/cli/\#session)

Manage OpenCode sessions.

```
opencode session [command]
```

* * *

#### [list](https://opencode.ai/docs/cli/\#list-3)

List all OpenCode sessions.

```
opencode session list
```

##### [Flags](https://opencode.ai/docs/cli/\#flags-6)

| Flag | Short | Description |
| --- | --- | --- |
| `--max-count` | `-n` | Limit to N most recent sessions |
| `--format` |  | Output format: table or json (table) |

* * *

### [stats](https://opencode.ai/docs/cli/\#stats)

Show token usage and cost statistics for your OpenCode sessions.

```
opencode stats
```

#### [Flags](https://opencode.ai/docs/cli/\#flags-7)

| Flag | Description |
| --- | --- |
| `--days` | Show stats for the last N days (all time) |
| `--tools` | Number of tools to show (all) |
| `--models` | Show model usage breakdown (hidden by default). Pass a number to show top N |
| `--project` | Filter by project (all projects, empty string: current project) |

* * *

### [export](https://opencode.ai/docs/cli/\#export)

Export session data as JSON.

```
opencode export [sessionID]
```

If you don’t provide a session ID, you’ll be prompted to select from available sessions.

* * *

### [import](https://opencode.ai/docs/cli/\#import)

Import session data from a JSON file or OpenCode share URL.

```
opencode import <file>
```

You can import from a local file or an OpenCode share URL.

```
opencode import session.json

opencode import https://opncd.ai/s/abc123
```

* * *

### [web](https://opencode.ai/docs/cli/\#web)

Start a headless OpenCode server with a web interface.

```
opencode web
```

This starts an HTTP server and opens a web browser to access OpenCode through a web interface.

#### [Flags](https://opencode.ai/docs/cli/\#flags-8)

| Flag | Description |
| --- | --- |
| `--port` | Port to listen on |
| `--hostname` | Hostname to listen on |
| `--mdns` | Enable mDNS discovery |
| `--cors` | Additional browser origin(s) to allow CORS |

* * *

### [acp](https://opencode.ai/docs/cli/\#acp)

Start an ACP (Agent Client Protocol) server.

```
opencode acp
```

This command starts an ACP server that communicates via stdin/stdout using nd-JSON.

#### [Flags](https://opencode.ai/docs/cli/\#flags-9)

| Flag | Description |
| --- | --- |
| `--cwd` | Working directory |
| `--port` | Port to listen on |
| `--hostname` | Hostname to listen on |

* * *

### [uninstall](https://opencode.ai/docs/cli/\#uninstall)

Uninstall OpenCode and remove all related files.

```
opencode uninstall
```

#### [Flags](https://opencode.ai/docs/cli/\#flags-10)

| Flag | Short | Description |
| --- | --- | --- |
| `--keep-config` | `-c` | Keep configuration files |
| `--keep-data` | `-d` | Keep session data and snapshots |
| `--dry-run` |  | Show what would be removed without removing |
| `--force` | `-f` | Skip confirmation prompts |

* * *

### [upgrade](https://opencode.ai/docs/cli/\#upgrade)

Updates opencode to the latest version or a specific version.

```
opencode upgrade [target]
```

To upgrade to the latest version.

```
opencode upgrade
```

To upgrade to a specific version.

```
opencode upgrade v0.1.48
```

#### [Flags](https://opencode.ai/docs/cli/\#flags-11)

| Flag | Short | Description |
| --- | --- | --- |
| `--method` | `-m` | The installation method that was used; curl, npm, pnpm, bun, brew |

* * *

## [Global Flags](https://opencode.ai/docs/cli/\#global-flags)

The opencode CLI takes the following global flags.

| Flag | Short | Description |
| --- | --- | --- |
| `--help` | `-h` | Display help |
| `--version` | `-v` | Print version number |
| `--print-logs` |  | Print logs to stderr |
| `--log-level` |  | Log level (DEBUG, INFO, WARN, ERROR) |

* * *

## [Environment variables](https://opencode.ai/docs/cli/\#environment-variables)

OpenCode can be configured using environment variables.

| Variable | Type | Description |
| --- | --- | --- |
| `OPENCODE_AUTO_SHARE` | boolean | Automatically share sessions |
| `OPENCODE_GIT_BASH_PATH` | string | Path to Git Bash executable on Windows |
| `OPENCODE_CONFIG` | string | Path to config file |
| `OPENCODE_CONFIG_DIR` | string | Path to config directory |
| `OPENCODE_CONFIG_CONTENT` | string | Inline json config content |
| `OPENCODE_DISABLE_AUTOUPDATE` | boolean | Disable automatic update checks |
| `OPENCODE_DISABLE_PRUNE` | boolean | Disable pruning of old data |
| `OPENCODE_DISABLE_TERMINAL_TITLE` | boolean | Disable automatic terminal title updates |
| `OPENCODE_PERMISSION` | string | Inlined json permissions config |
| `OPENCODE_DISABLE_DEFAULT_PLUGINS` | boolean | Disable default plugins |
| `OPENCODE_DISABLE_LSP_DOWNLOAD` | boolean | Disable automatic LSP server downloads |
| `OPENCODE_ENABLE_EXPERIMENTAL_MODELS` | boolean | Enable experimental models |
| `OPENCODE_DISABLE_AUTOCOMPACT` | boolean | Disable automatic context compaction |
| `OPENCODE_CLIENT` | string | Client identifier (defaults to `cli`) |
| `OPENCODE_ENABLE_EXA` | boolean | Enable Exa web search tools |

* * *

### [Experimental](https://opencode.ai/docs/cli/\#experimental)

These environment variables enable experimental features that may change or be removed.

| Variable | Type | Description |
| --- | --- | --- |
| `OPENCODE_EXPERIMENTAL` | boolean | Enable all experimental features |
| `OPENCODE_EXPERIMENTAL_ICON_DISCOVERY` | boolean | Enable icon discovery |
| `OPENCODE_EXPERIMENTAL_DISABLE_COPY_ON_SELECT` | boolean | Disable copy on select in TUI |
| `OPENCODE_EXPERIMENTAL_BASH_MAX_OUTPUT_LENGTH` | number | Max output length for bash commands |
| `OPENCODE_EXPERIMENTAL_BASH_DEFAULT_TIMEOUT_MS` | number | Default timeout for bash commands in ms |
| `OPENCODE_EXPERIMENTAL_OUTPUT_TOKEN_MAX` | number | Max output tokens for LLM responses |
| `OPENCODE_EXPERIMENTAL_FILEWATCHER` | boolean | Enable file watcher for entire dir |
| `OPENCODE_EXPERIMENTAL_OXFMT` | boolean | Enable oxfmt formatter |
| `OPENCODE_EXPERIMENTAL_LSP_TOOL` | boolean | Enable experimental LSP tool |
