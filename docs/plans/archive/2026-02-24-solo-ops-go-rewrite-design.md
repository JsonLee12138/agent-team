# agent-team Go Rewrite Design

## Motivation

将 agent-team 从 Python 脚本重写为 Go，编译成单一二进制文件，通过 Homebrew 分发，消除用户对 Python 环境的依赖。

## Current State

- `skills/solo-ops/scripts/solo_ops.py` — ~620 行 Python，零第三方依赖
- 8 个子命令：create, delete, open, open-all, assign, reply, status, merge
- 终端后端：WezTerm / tmux（通过 `AGENT_TEAM_BACKEND` 环境变量切换）
- 配置：每个 role 一个 `config.yaml`（6 个字段），手写行解析

## Decision

- **语言**：Go（交叉编译简单、goreleaser + Homebrew tap 生态成熟、项目复杂度匹配）
- **配置**：struct + `yaml.v3`，不引入 framework 依赖
- **CLI 框架**：cobra

## Project Structure

```
agent-team/
├── main.go
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go              # cobra root command, version, help
│   ├── create.go
│   ├── delete.go
│   ├── open.go              # open + open-all
│   ├── assign.go
│   ├── reply.go
│   ├── status.go
│   └── merge.go
├── internal/
│   ├── config.go            # RoleConfig struct, YAML read/write
│   ├── session.go           # SessionBackend interface + implementations
│   ├── git.go               # GitClient
│   └── role.go              # role discovery, path resolution, CLAUDE.md generation
├── skills/                  # retained for Claude Code skill integration
│   └── solo-ops/
│       ├── SKILL.md
│       └── references/
├── .goreleaser.yaml
└── Makefile
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI subcommand framework |
| `gopkg.in/yaml.v3` | YAML serialization |

No other third-party dependencies.

## Design Patterns

### Strategy: SessionBackend

Current problem: wezterm/tmux branching logic (`if backend == "tmux"`) is scattered across `pane_alive`, `pane_send`, `cmd_open`, `cmd_delete` — classic Shotgun Surgery.

Solution: extract a `SessionBackend` interface, one implementation per terminal multiplexer.

```go
// internal/session.go

type SessionBackend interface {
    PaneAlive(paneID string) bool
    PaneSend(paneID string, text string) error
    SpawnPane(cwd string) (paneID string, err error)
    KillPane(paneID string) error
    SetTitle(paneID string, title string) error
    ActivatePane(paneID string) error
}

type WeztermBackend struct{}
type TmuxBackend struct{}
```

Adding a new backend (e.g., kitty, zellij) means adding one struct — no existing code changes.

### Facade: GitClient

Wraps all git subprocess calls behind a typed API.

```go
// internal/git.go

type GitClient struct {
    root string
}

func NewGitClient() (*GitClient, error)        // calls git rev-parse --show-toplevel
func (g *GitClient) Root() string
func (g *GitClient) WorktreeAdd(path, branch string) error
func (g *GitClient) WorktreeRemove(path string) error
func (g *GitClient) Merge(branch, message string) error
func (g *GitClient) DeleteBranch(branch string) error
func (g *GitClient) CurrentBranch() (string, error)
```

### Dependency Injection: App struct

Commands receive dependencies via an `App` struct rather than reaching into globals.

```go
// cmd/root.go

type App struct {
    Git     *internal.GitClient
    Session internal.SessionBackend
    WtBase  string // ".worktrees" or "worktrees"
}
```

Each subcommand is a method on `App`:

```go
// cmd/create.go
func (a *App) RunCreate(cmd *cobra.Command, args []string) error

// cmd/open.go
func (a *App) RunOpen(cmd *cobra.Command, args []string) error
```

Benefits: testable without subprocess mocking, clear dependency graph.

## Config

### RoleConfig (per-role, read + write)

```go
// internal/config.go

type RoleConfig struct {
    Name            string `yaml:"name"`
    Description     string `yaml:"description"`
    DefaultProvider string `yaml:"default_provider"`
    DefaultModel    string `yaml:"default_model"`
    CreatedAt       string `yaml:"created_at"`
    PaneID          string `yaml:"pane_id"`
}

func LoadRoleConfig(path string) (*RoleConfig, error)
func (c *RoleConfig) Save(path string) error
```

### Constants (hardcoded)

```go
var SupportedProviders = map[string]bool{
    "claude":   true,
    "codex":    true,
    "opencode": true,
}

var LaunchCommands = map[string]string{
    "claude":   "claude --dangerously-skip-permissions",
    "codex":    "codex --dangerously-bypass-approvals-and-sandbox",
    "opencode": "opencode",
}
```

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `AGENT_TEAM_BACKEND` | Terminal backend selection | `wezterm` |
| `WEZTERM_PANE` | Current pane ID (for focus return) | empty |

## Commands

All commands mirror the Python version 1:1.

| Command | Positional Args | Flags |
|---------|-----------------|-------|
| `create <name>` | name | — |
| `delete <name>` | name | — |
| `open <name> [provider]` | name, provider (optional) | `--model` |
| `open-all [provider]` | provider (optional) | `--model` |
| `assign <name> <task> [provider]` | name, task, provider (optional) | `--model` |
| `reply <name> <answer>` | name, answer | — |
| `status` | — | — |
| `merge <name>` | name | — |

## Role Directory Layout

Unchanged from Python version:

```
.worktrees/<name>/
  CLAUDE.md                          ← auto-generated from prompt.md on open
  agents/teams/<name>/
    config.yaml                      ← RoleConfig
    prompt.md                        ← role system prompt
    tasks/
      pending/<timestamp>-<slug>.md
      done/<timestamp>-<slug>.md
```

## Distribution

### goreleaser

`.goreleaser.yaml` targets:
- `darwin/amd64`, `darwin/arm64`
- `linux/amd64`, `linux/arm64`

### Homebrew

Homebrew tap repository (e.g., `leeforge/homebrew-tap`) with formula auto-generated by goreleaser.

```
brew tap leeforge/tap
brew install agent-team
```

## Testing Strategy

- `internal/` packages: unit tests with interface mocks (no subprocess needed)
- `cmd/`: integration tests that exercise the full command flow
- CI: `go test ./...` on push

## Migration

1. Build and validate Go version against existing Python test cases
2. Update `skills/solo-ops/SKILL.md` to reference `agent-team` binary instead of `python3 ... solo_ops.py`
3. Python scripts remain in repo for reference, removed in a follow-up cleanup
