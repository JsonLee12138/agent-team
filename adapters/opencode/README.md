# opencode-agent-team

[Agent Team](https://github.com/JsonLee12138/agent-team) plugin for [OpenCode](https://opencode.ai) — role injection and quality checks.

## How It Works

This plugin delegates logic to the `agent-team` Go binary. It registers OpenCode plugin events and forwards them to the CLI via shell commands.

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

## License

MIT
