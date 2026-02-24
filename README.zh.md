# agent-team

[English](./README.md) | 中文

多智能体开发工作流的 AI 团队角色管理器。借助 WezTerm 或 tmux，在隔离的 Git worktree 中协调多个 AI 编程智能体。

每个角色拥有独立的 Git 分支、worktree、终端会话和任务收件箱 —— 全部通过自然语言提示词或 CLI 命令管理。

## 目录

- [工作原理](#工作原理)
- [环境要求](#环境要求)
- [安装](#安装)
  - [Agent Skill（推荐）](#agent-skill推荐)
  - [Homebrew (macOS)](#homebrew-macos)
  - [从源码编译](#从源码编译)
  - [从 GitHub Releases 下载](#从-github-releases-下载)
- [升级](#升级)
- [作为 Skill 使用](#作为-skill-使用)
  - [创建角色](#创建角色)
  - [打开角色会话](#打开角色会话)
  - [分配任务](#分配任务)
  - [回复角色](#回复角色)
  - [查看状态](#查看状态)
  - [合并与清理](#合并与清理)
- [CLI 命令参考](#cli-命令参考)
- [目录结构](#目录结构)
- [支持的 Provider](#支持的-provider)
- [环境变量](#环境变量)
- [License](#license)

## 工作原理

```
主分支
    │
    ├── .worktrees/frontend/     ← team/frontend 分支 + Claude 会话
    ├── .worktrees/backend/      ← team/backend 分支 + Codex 会话
    └── .worktrees/qa/           ← team/qa 分支 + OpenCode 会话
```

每个角色在独立环境中运行。你通过主智能体分配任务，各角色独立工作，完成后合并回主分支。

## 环境要求

- Git
- [WezTerm](https://wezfurlong.org/wezterm/) 或 [tmux](https://github.com/tmux/tmux)
- 至少一个 AI provider CLI：[claude](https://github.com/anthropics/claude-code)、[codex](https://github.com/openai/codex) 或 [opencode](https://opencode.ai)

## 安装

### Agent Skill（推荐）

安装为 Agent Skill，让 Claude Code（或其他兼容的 AI 智能体）通过自然语言管理你的团队：

```bash
npx skills add JsonLee12138/agent-team
```

### Homebrew (macOS)

```bash
brew tap JsonLee12138/agent-team
brew install agent-team
```

### 从源码编译

需要 Go 1.24+。

```bash
go install github.com/JsonLee12138/agent-team@latest
```

### 从 GitHub Releases 下载

从 [Releases](https://github.com/JsonLee12138/agent-team/releases) 下载对应平台的二进制文件，解压后加入 `PATH` 即可。

## 升级

### Agent Skill

```bash
npx skills add JsonLee12138/agent-team
```

### Homebrew

```bash
brew update && brew upgrade agent-team
```

### 从源码编译

```bash
go install github.com/JsonLee12138/agent-team@latest
```

## 作为 Skill 使用

安装完成后，直接在 AI 智能体会话中用自然语言描述你想做的事，无需记忆命令语法。

### 创建角色

> 帮我创建一个叫 "frontend" 的团队角色，负责开发 React 组件。

> 创建一个 backend 角色，专门处理 Node.js API 开发。

智能体会自动搭建 worktree，并引导你在 `prompt.md` 中定义角色的专业领域和行为。

---

### 打开角色会话

> 用 Claude 打开 frontend 角色的会话。

> 用 Codex 打开所有角色的会话。

> 用 claude-opus-4-5 模型打开 backend 角色。

将在新的终端 tab（或 tmux 窗口）中启动 AI provider，运行在该角色的 worktree 目录下。

---

### 分配任务

> 给 frontend 分配一个任务：实现带移动端汉堡菜单的响应式导航栏。

> 告诉 backend 角色添加一个 JWT 鉴权中间件。

任务会写入角色的待处理收件箱，并通知运行中的会话。如果会话未启动，会自动开启。

---

### 回复角色

> 回复 frontend：外层布局用 CSS Grid，子元素用 Flexbox。

> 告诉 backend 角色我们用的是 PostgreSQL，不是 MySQL。

消息会以 `[Main Controller Reply]` 为前缀发送到运行中的角色会话。

---

### 查看状态

> 查看团队状态。

> 哪些角色正在运行？

显示所有角色、会话运行状态和待处理任务数量。

---

### 合并与清理

> 合并 frontend 分支。

> 合并后删除 backend 角色。

将 `team/<name>` 以 `--no-ff` 方式合并到当前分支，之后可选择移除 worktree 和分支。

---

## CLI 命令参考

所有命令需在 Git 仓库目录下运行。

| 命令 | 说明 |
|------|------|
| `agent-team create <name>` | 创建新角色（分支 + worktree + 目录结构） |
| `agent-team open <name> [provider] [--model <m>]` | 在新终端 tab 中打开角色会话 |
| `agent-team open-all [provider] [--model <m>]` | 打开所有角色的会话 |
| `agent-team assign <name> "<task>" [provider]` | 分配任务并通知会话 |
| `agent-team reply <name> "<message>"` | 向运行中的角色会话发送回复 |
| `agent-team status` | 查看所有角色状态和待处理任务 |
| `agent-team merge <name>` | 将角色分支合并到当前分支 |
| `agent-team delete <name>` | 关闭会话并删除 worktree 和分支 |

切换到 tmux 后端：在命令前加 `AGENT_TEAM_BACKEND=tmux`。

## 目录结构

```
项目根目录/
└── .worktrees/
    └── <name>/
        ├── CLAUDE.md                              ← 由 prompt.md 自动生成
        └── agents/teams/<name>/
            ├── config.yaml                        ← provider、model、pane_id
            ├── prompt.md                          ← 角色定义（手动编辑）
            └── tasks/
                ├── pending/<timestamp>-<slug>.md  ← 待处理任务
                └── done/<timestamp>-<slug>.md     ← 已完成任务
```

## 支持的 Provider

| Provider | 值 |
|----------|----|
| Claude Code | `claude`（默认） |
| OpenAI Codex | `codex` |
| OpenCode | `opencode` |

可在 `config.yaml` 的 `default_provider` 字段中设置每个角色的默认 provider，也可在 `open` / `assign` 命令中作为位置参数传入。

## 环境变量

| 变量 | 说明 |
|------|------|
| `AGENT_TEAM_BACKEND` | 终端后端：`wezterm`（默认）或 `tmux` |

## License

MIT
