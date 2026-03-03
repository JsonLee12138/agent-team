# opencode-agent-team

[Agent Team](https://github.com/JsonLee12138/agent-team) plugin for [OpenCode](https://opencode.ai) — role injection, quality checks, and task lifecycle hooks.

## How It Works

This plugin delegates all hook logic to the `agent-team` Go binary. It registers OpenCode plugin hooks (`tool.execute.before` / `tool.execute.after`) and forwards events to the CLI via shell commands.

## Prerequisites

- `agent-team` binary in PATH ([Installation](https://github.com/JsonLee12138/agent-team#installation))

## Install

```bash
npm install opencode-agent-team
```

Then add to your `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["opencode-agent-team"]
}
```

## Upgrade

```bash
npm update opencode-agent-team
```

Also upgrade the `agent-team` CLI binary separately (Homebrew or from source).

## Hooks

| Hook | Description |
|------|-------------|
| `tool.execute.before` | Runs `agent-team hook pre-tool-use` before each tool call |
| `tool.execute.after` | Runs `agent-team hook post-tool-use` after each tool call |

## License

MIT
