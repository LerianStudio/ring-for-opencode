---
title: "Troubleshooting | OpenCode"
source: "https://opencode.ai/docs/troubleshooting/"
crawled: "2026-01-05T00:46:23Z"
---

[Skip to content](https://opencode.ai/docs/troubleshooting/#_top)

# Troubleshooting

Common issues and how to resolve them.

To debug any issues with OpenCode, you can check the logs or the session data
that it stores locally.

* * *

### [Logs](https://opencode.ai/docs/troubleshooting/\#logs)

Log files are written to:

- **macOS/Linux**: `~/.local/share/opencode/log/`
- **Windows**: `%USERPROFILE%\.local\share\opencode\log\`

Log files are named with timestamps (e.g., `2025-01-09T123456.log`) and the most recent 10 log files are kept.

You can set the log level with the `--log-level` command-line option to get more detailed debug information. For example, `opencode --log-level DEBUG`.

* * *

### [Storage](https://opencode.ai/docs/troubleshooting/\#storage)

opencode stores session data and other application data on disk at:

- **macOS/Linux**: `~/.local/share/opencode/`
- **Windows**: `%USERPROFILE%\.local\share\opencode`

This directory contains:

- `auth.json` \- Authentication data like API keys, OAuth tokens
- `log/` \- Application logs
- `project/`\- Project-specific data like session and message data

  - If the project is within a Git repo, it is stored in `./<project-slug>/storage/`
  - If it is not a Git repo, it is stored in `./global/storage/`

* * *

## [Getting help](https://opencode.ai/docs/troubleshooting/\#getting-help)

If you’re experiencing issues with OpenCode:

1. **Report issues on GitHub**

The best way to report bugs or request features is through our GitHub repository:

[**github.com/anomalyco/opencode/issues**](https://github.com/anomalyco/opencode/issues)

Before creating a new issue, search existing issues to see if your problem has already been reported.

2. **Join our Discord**

For real-time help and community discussion, join our Discord server:

[**opencode.ai/discord**](https://opencode.ai/discord)


* * *

## [Common issues](https://opencode.ai/docs/troubleshooting/\#common-issues)

Here are some common issues and how to resolve them.

* * *

### [OpenCode won’t start](https://opencode.ai/docs/troubleshooting/\#opencode-wont-start)

1. Check the logs for error messages
2. Try running with `--print-logs` to see output in the terminal
3. Ensure you have the latest version with `opencode upgrade`

* * *

### [Authentication issues](https://opencode.ai/docs/troubleshooting/\#authentication-issues)

1. Try re-authenticating with the `/connect` command in the TUI
2. Check that your API keys are valid
3. Ensure your network allows connections to the provider’s API

* * *

### [Model not available](https://opencode.ai/docs/troubleshooting/\#model-not-available)

1. Check that you’ve authenticated with the provider
2. Verify the model name in your config is correct
3. Some models may require specific access or subscriptions

If you encounter `ProviderModelNotFoundError` you are most likely incorrectly
referencing a model somewhere.
Models should be referenced like so: `<providerId>/<modelId>`

Examples:

- `openai/gpt-4.1`
- `openrouter/google/gemini-2.5-flash`
- `opencode/kimi-k2`

To figure out what models you have access to, run `opencode models`

* * *

### [ProviderInitError](https://opencode.ai/docs/troubleshooting/\#provideriniterror)

If you encounter a ProviderInitError, you likely have an invalid or corrupted configuration.

To resolve this:

1. First, verify your provider is set up correctly by following the [providers guide](https://opencode.ai/docs/providers)

2. If the issue persists, try clearing your stored configuration:



```
rm -rf ~/.local/share/opencode
```

3. Re-authenticate with your provider using the `/connect` command in the TUI.


* * *

### [AI\_APICallError and provider package issues](https://opencode.ai/docs/troubleshooting/\#ai_apicallerror-and-provider-package-issues)

If you encounter API call errors, this may be due to outdated provider packages. opencode dynamically installs provider packages (OpenAI, Anthropic, Google, etc.) as needed and caches them locally.

To resolve provider package issues:

1. Clear the provider package cache:



```
rm -rf ~/.cache/opencode
```

2. Restart opencode to reinstall the latest provider packages


This will force opencode to download the most recent versions of provider packages, which often resolves compatibility issues with model parameters and API changes.

* * *

### [Copy/paste not working on Linux](https://opencode.ai/docs/troubleshooting/\#copypaste-not-working-on-linux)

Linux users need to have one of the following clipboard utilities installed for copy/paste functionality to work:

**For X11 systems:**

```
apt install -y xclip

# or

apt install -y xsel
```

**For Wayland systems:**

```
apt install -y wl-clipboard
```

**For headless environments:**

```
apt install -y xvfb

# and run:

Xvfb :99 -screen 0 1024x768x24 > /dev/null 2>&1 &

export DISPLAY=:99.0
```

opencode will detect if you’re using Wayland and prefer `wl-clipboard`, otherwise it will try to find clipboard tools in order of: `xclip` and `xsel`.
