---
title: "ACP Support | OpenCode"
source: "https://opencode.ai/docs/acp"
crawled: "2026-01-05T00:46:23Z"
---

[Skip to content](https://opencode.ai/docs/acp#_top)

# ACP Support

Use OpenCode in any ACP-compatible editor.

OpenCode supports the [Agent Client Protocol](https://agentclientprotocol.com/) or (ACP), allowing you to use it directly in compatible editors and IDEs.

ACP is an open protocol that standardizes communication between code editors and AI coding agents.

* * *

## [Configure](https://opencode.ai/docs/acp\#configure)

To use OpenCode via ACP, configure your editor to run the `opencode acp` command.

The command starts OpenCode as an ACP-compatible subprocess that communicates with your editor over JSON-RPC via stdio.

Below are examples for popular editors that support ACP.

* * *

### [Zed](https://opencode.ai/docs/acp\#zed)

Add to your [Zed](https://zed.dev/) configuration (`~/.config/zed/settings.json`):

```
{

  "agent_servers": {

    "OpenCode": {

      "command": "opencode",

      "args": ["acp"]

    }

  }

}
```

To open it, use the `agent: new thread` action in the **Command Palette**.

You can also bind a keyboard shortcut by editing your `keymap.json`:

```
[\
\
  {\
\
    "bindings": {\
\
      "cmd-alt-o": [\
\
        "agent::NewExternalAgentThread",\
\
        {\
\
          "agent": {\
\
            "custom": {\
\
              "name": "OpenCode",\
\
              "command": {\
\
                "command": "opencode",\
\
                "args": ["acp"]\
\
              }\
\
            }\
\
          }\
\
        }\
\
      ]\
\
    }\
\
  }\
\
]
```

* * *

### [JetBrains IDEs](https://opencode.ai/docs/acp\#jetbrains-ides)

Add to your [JetBrains IDE](https://www.jetbrains.com/) acp.json according to the [documentation](https://www.jetbrains.com/help/ai-assistant/acp.html):

```
{

  "agent_servers": {

    "OpenCode": {

      "command": "/absolute/path/bin/opencode",

      "args": ["acp"]

    }

  }

}
```

To open it, use the new ‘OpenCode’ agent in the AI Chat agent selector.

* * *

### [Avante.nvim](https://opencode.ai/docs/acp\#avantenvim)

Add to your [Avante.nvim](https://github.com/yetone/avante.nvim) configuration:

```
{

  acp_providers = {

    ["opencode"] = {

      command = "opencode",

      args = { "acp" }

    }

  }

}
```

If you need to pass environment variables:

```
{

  acp_providers = {

    ["opencode"] = {

      command = "opencode",

      args = { "acp" },

      env = {

        OPENCODE_API_KEY = os.getenv("OPENCODE_API_KEY")

      }

    }

  }

}
```

* * *

### [CodeCompanion.nvim](https://opencode.ai/docs/acp\#codecompanionnvim)

To use OpenCode as an ACP agent in [CodeCompanion.nvim](https://github.com/olimorris/codecompanion.nvim), add the following to your Neovim config:

```
require("codecompanion").setup({

  strategies = {

    chat = {

      adapter = {

        name = "opencode",

        model = "claude-sonnet-4",

      },

    },

  },

})
```

This config sets up CodeCompanion to use OpenCode as the ACP agent for chat.

If you need to pass environment variables (like `OPENCODE_API_KEY`), refer to [Configuring Adapters: Environment Variables](https://codecompanion.olimorris.dev/configuration/adapters#environment-variables-setting-an-api-key) in the CodeCompanion.nvim documentation for full details.

## [Support](https://opencode.ai/docs/acp\#support)

OpenCode works the same via ACP as it does in the terminal. All features are supported:

- Built-in tools (file operations, terminal commands, etc.)
- Custom tools and slash commands
- MCP servers configured in your OpenCode config
- Project-specific rules from `AGENTS.md`
- Custom formatters and linters
- Agents and permissions system
