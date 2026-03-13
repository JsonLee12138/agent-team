# Installation & Upgrade Guide

Per-provider instructions for installing and upgrading agent-team. AI agents can use these steps to self-install or self-upgrade when needed.

> **IMPORTANT — Installation Strategy**
> - **Plugin platforms** (Claude Code, Gemini CLI, OpenCode): Install and update **exclusively** through each platform's native plugin/extension system. Do **NOT** use `npx skills add` on these platforms — the plugin already bundles all skill files.
> - **Codex**: No plugin support. Use `npx skills add` to install skills **globally only** (`-a codex -g`).
> - **Other platforms** (no plugin support): Use `npx skills add` with the specific platform flag. See [Agent Skill (Non-Plugin Platforms)](#agent-skill-non-plugin-platforms).
> - **NEVER** use `-a '*'` (all-platform wildcard) — it pollutes the project with config directories for providers the user does not use.

## Quick Check

```bash
# Verify agent-team binary is available
which agent-team && agent-team --version
```

If `agent-team` is not found, follow the installation steps for your provider below.

---

## Claude Code (Plugin Marketplace)

> **Plugin handles everything.** Do NOT use `npx skills add` for Claude Code — the plugin bundles all skills.

### Install

```bash
# Option 1: Via Claude Code slash commands
/plugin marketplace add JsonLee12138/agent-team
/plugin install agent-team@agent-team

# Option 2: Via CLI
claude plugin marketplace add JsonLee12138/agent-team
claude plugin install agent-team@agent-team
```

### Upgrade

```bash
# Via slash command
/plugin marketplace update agent-team

# Via CLI
claude plugin marketplace update agent-team
```

After upgrading the plugin, run `agent-team init` to sync bundled roles to global.

---

## Gemini CLI (Extension)

> **Extension handles everything.** Do NOT use `npx skills add` for Gemini CLI — the extension bundles all skills.

### Install

```bash
gemini extensions install https://github.com/JsonLee12138/agent-team
```

This installs the `gemini-extension.json` manifest. The `GEMINI.md` context file is loaded automatically in worktrees.

Requires `agent-team` binary in PATH (see [Homebrew](#homebrew-macos) or [From Source](#from-source) below).

### Upgrade

```bash
gemini extensions update agent-team
```

Also upgrade the CLI binary separately (Homebrew or Source).

---

## OpenCode (npm Plugin)

> **Plugin handles everything.** Do NOT use `npx skills add` for OpenCode — the plugin bundles all skills.

### Install

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

Requires `agent-team` binary in PATH (see [Homebrew](#homebrew-macos) or [From Source](#from-source) below).

### Upgrade

```bash
npm update opencode-agent-team
```

Also upgrade the CLI binary separately (Homebrew or Source).

---

## Codex (Agent Skill — Global Only)

Codex has no plugin system. Behaviors like brainstorming gate and quality checks are enforced via prompt conventions in the role's `system.md`.

> **Global install only.** Codex skills are installed globally so they're available across all projects.

### Install

```bash
npx skills add JsonLee12138/agent-team -a codex -g -y
```

Requires `agent-team` binary in PATH (see [Homebrew](#homebrew-macos) or [From Source](#from-source) below).

### Upgrade

```bash
npx skills add JsonLee12138/agent-team -a codex -g -y
```

Re-running the same command overwrites with the latest version.

---

## Agent Skill (Non-Plugin Platforms)

For platforms that do **not** have native plugin/extension support (e.g. Cursor, Windsurf, Cline, Amp, etc.), use `npx skills add` to install skill files directly.

> **Do NOT use this method for Claude Code, Gemini CLI, or OpenCode** — those platforms have dedicated plugin/extension systems (see above).
> **Do NOT use `-a '*'`** — each platform creates its own config directory. Install only the platforms you use.

### Install

```bash
# Project-level — specify platform explicitly
npx skills add JsonLee12138/agent-team -a <platform> -y

# Global
npx skills add JsonLee12138/agent-team -a <platform> -g -y
```

### Upgrade

Same command as install — re-running overwrites with the latest:

```bash
npx skills add JsonLee12138/agent-team -a <platform> -y
```

---

## Homebrew (macOS)

### Install

```bash
brew tap JsonLee12138/agent-team
brew install agent-team
```

### Upgrade

```bash
brew update && brew upgrade agent-team
```

---

## From Source

Requires Go 1.24+.

### Install

```bash
go install github.com/JsonLee12138/agent-team@latest
```

### Upgrade

```bash
go install github.com/JsonLee12138/agent-team@latest
```

---

## From GitHub Releases

Download a binary from [Releases](https://github.com/JsonLee12138/agent-team/releases), extract it, and add it to your `PATH`.

### Upgrade

Download the latest release and replace the old binary.

---

## Post-Install: Initialize

After installing (or upgrading), run:

```bash
agent-team init
```

Before running `init`, ask the user which AI providers they use. This will:
1. Detect installed AI providers (claude, gemini, opencode, codex)
2. Create `.agents/teams/` project structure (if in a git repo)
3. Install/update bundled roles to `~/.agents/roles/`

Use `--global-only` to skip project setup, or `--skip-detect` to skip provider detection.
