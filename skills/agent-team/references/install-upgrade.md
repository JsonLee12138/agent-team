# Installation & Upgrade Guide

Per-provider instructions for installing and upgrading agent-team. AI agents can use these steps to self-install or self-upgrade when needed.

> **IMPORTANT — Platform Selection Rule**
> Before installing or upgrading skills, you **MUST** ask the user which platform(s) they use (e.g. `claude`, `gemini`, `opencode`, `codex`).
> **NEVER** use `-a '*'` (all-platform wildcard) for project-level skill installation — it pollutes the project with config directories for providers the user does not use.
> Global-level (`-g`) also requires explicit platform confirmation.

## Quick Check

```bash
# Verify agent-team binary is available
which agent-team && agent-team --version
```

If `agent-team` is not found, follow the installation steps for your provider below.

---

## Claude Code (Plugin Marketplace)

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

### Install

```bash
gemini extensions install https://github.com/JsonLee12138/agent-team
```

This installs the `gemini-extension.json` manifest and hooks. The `GEMINI.md` context file is loaded automatically in worktrees.

Requires `agent-team` binary in PATH (see [Homebrew](#homebrew-macos) or [From Source](#from-source) below).

### Upgrade

```bash
gemini extensions update agent-team
```

Also upgrade the CLI binary separately (Homebrew or Source).

---

## OpenCode (npm Plugin)

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

## Codex (Agent Skill)

Codex has no plugin or hook system. Hook behaviors (brainstorming gate, quality checks) are enforced via prompt conventions in the role's `system.md`.

### Install

```bash
npx skills add JsonLee12138/agent-team
```

Skills are installed into `.codex/skills/` automatically when creating workers with `--provider codex`.

Requires `agent-team` binary in PATH (see [Homebrew](#homebrew-macos) or [From Source](#from-source) below).

### Upgrade

```bash
npx skills add JsonLee12138/agent-team
```

Re-running the same command overwrites with the latest version.

---

## Agent Skill (all providers)

This installs the skill files (SKILL.md, system.md, etc.) into the provider's skill directory.

### Install

Before running, you **MUST** confirm with the user:
1. **Platform**: which specific agent platform(s) — see [platforms.md](platforms.md)
2. **Scope**: project-level or global

> **Do NOT use `-a '*'` for project-level installs.** Each platform creates its own config directory (`.claude/`, `.codex/`, etc.). Installing all platforms pollutes the project with unused directories and may conflict with existing provider configs.

```bash
# Project-level — specify platform explicitly
npx skills add JsonLee12138/agent-team -a <platform> -y

# Multiple specific platforms (if user confirms)
npx skills add JsonLee12138/agent-team -a claude -a gemini -y

# Global — still requires explicit platform(s)
npx skills add JsonLee12138/agent-team -a <platform> -y -g
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
