# agent-team

[English](./README.md) | 中文

多智能体开发工作流的 AI 团队角色管理器。采用 **Role + Worker** 双层模型，在隔离 Git worktree 中运行。

- **Role**：角色技能包定义，位于 `.agents/teams/<role-name>/`（项目级）或 `~/.agents/roles/<role-name>/`（全局）
- **Worker**：角色运行实例，位于 `.worktrees/<worker-id>/`（例如 `frontend-dev-001`）

## 目录

- [工作原理](#工作原理)
- [环境要求](#环境要求)
- [安装](#安装)
- [升级](#升级)
- [快速开始](#快速开始)
- [自带角色](#自带角色)
- [自带 Skills](#自带-skills)
- [支持的 Provider](#支持的-provider)
- [高级](#高级)
- [License](#license)

## 工作原理

```
主分支
  ├── .agents/teams/<role-name>/           <- 项目级角色技能包定义
  └── .worktrees/<worker-id>/              <- 隔离运行目录
        └── worker.yaml                    <- worker 配置

~/.agents/roles/<role-name>/               <- 全局角色技能包定义
```

典型流程：

1. **定义角色** — "给项目创建一个前端开发角色。"
2. **启动 Worker** — "给 frontend-dev 创建一个 worker，用 claude。"
3. **头脑风暴 & 分配任务** — "让 frontend-dev-001 实现响应式导航栏。"
4. **合并成果** — "合并 frontend-dev-001。"
5. **清理** — "删除 worker frontend-dev-001。"

## 环境要求

- Git
- [WezTerm](https://wezfurlong.org/wezterm/) 或 [tmux](https://github.com/tmux/tmux)
- 至少一个 AI provider CLI：[Claude Code](https://github.com/anthropics/claude-code)、[Gemini CLI](https://github.com/google-gemini/gemini-cli)、[Codex](https://github.com/openai/codex) 或 [OpenCode](https://opencode.ai)

## 安装

**注意：** 安装方式因 provider 而异，请选择你使用的 AI 工具对应的章节。

### AI 自助安装（推荐）

让 AI 自行完成 agent-team 的安装和配置，只需两步：

**第 1 步：安装 skill** — 让 AI 获得安装所需的知识。

```bash
# Claude Code
npx skills add JsonLee12138/agent-team -a claude -y

# Gemini CLI
npx skills add JsonLee12138/agent-team -a gemini -y

# OpenCode
npx skills add JsonLee12138/agent-team -a opencode -y

# Codex
npx skills add JsonLee12138/agent-team -a codex -y

# 一次性安装到所有平台
npx skills add JsonLee12138/agent-team -a '*' -y
```

**第 2 步：让 AI 安装并初始化。**

> "安装 agent-team 并初始化项目。"

AI 会读取 skill 中的说明，自动安装二进制文件并执行 `agent-team init` 完成配置 — 无需手动操作。

---

### Claude Code（插件市场安装）

```bash
# 1. 添加插件市场
/plugin marketplace add JsonLee12138/agent-team

# 2. 安装插件
/plugin install agent-team@agent-team
```

或通过 CLI：

```bash
claude plugin marketplace add JsonLee12138/agent-team
claude plugin install agent-team@agent-team
```

### Gemini CLI（扩展安装）

```bash
gemini extensions install https://github.com/JsonLee12138/agent-team
```

安装后会注册 `gemini-extension.json` 清单和 hooks，`GEMINI.md` 上下文文件在 worktree 中自动加载。

### OpenCode（npm 插件安装）

```bash
npm install opencode-agent-team
```

然后在 `opencode.json` 中添加：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["opencode-agent-team"]
}
```

需要 `agent-team` 二进制在 PATH 中（见下方 [Homebrew](#homebrew-macos) 或[从源码编译](#从源码编译)）。

### Codex

Codex 没有插件/hook 系统，安装二进制后使用 Agent Skill 方式：

```bash
npx skills add JsonLee12138/agent-team
```

创建 worker 时使用 `--provider codex` 会自动将 skills 安装到 `.codex/skills/`。Hook 行为（brainstorming gate、quality checks）通过角色 `system.md` 中的 prompt 约定来保证。

### Agent Skill（通用）

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

从 [Releases](https://github.com/JsonLee12138/agent-team/releases) 下载二进制，解压后加入 `PATH`。

## 升级

```bash
# Claude Code 插件
/plugin marketplace update agent-team
# 或
claude plugin marketplace update agent-team

# Gemini CLI 扩展
gemini extensions update agent-team

# Skill
npx skills add JsonLee12138/agent-team

# Homebrew
brew update && brew upgrade agent-team

# 源码安装
go install github.com/JsonLee12138/agent-team@latest
```

## 快速开始

安装完成后，你可以完全通过自然语言来管理团队。AI 会理解你的意图并自动执行对应的命令。

### 1. 创建角色

> "创建一个前端开发角色，负责 UI 实现。"

AI 会引导你通过头脑风暴确定角色的范围、目标和所需技能，然后生成角色技能包到 `.agents/teams/`。

也可以一次创建多个角色：

> "创建一个团队，包含前端开发、QA 工程师和产品经理角色。"

### 2. 查看角色

> "展示所有可用角色。"

### 3. 创建 Worker

> "给 frontend-dev 创建一个 worker，用 claude。"

这会创建一个隔离的 Git worktree，打开终端会话，安装所需技能，并启动 AI provider。

### 4. 分配任务

> "让 frontend-dev-001 实现响应式导航栏。"

AI 会先进行头脑风暴产出设计文档，然后创建任务并通知 worker 会话。

### 5. 查看状态

> "查看团队状态。"

### 6. 合并与清理

> "合并 frontend-dev-001。"
>
> "删除 worker frontend-dev-001。"

### 从 GitHub 安装角色

> "搜索与 react 开发相关的角色。"
>
> "从 owner/repo 安装角色。"

## 自带角色

本仓库在 `.agents/teams/` 中包含以下自带角色：

| 角色 | 说明 |
|------|------|
| `pm` | 产品经理 |
| `frontend-architect` | 前端架构 |
| `vite-react-dev` | Vite + React 开发 |
| `uniapp-dev` | UniApp 开发 |
| `pencil-designer` | Pencil 设计工具专家 |

> "给 frontend-architect 创建一个 worker，用 claude。"

## 自带 Skills

| Skill | 说明 |
|-------|------|
| `role-creator` | 交互式创建或更新角色技能包 |
| `brainstorming` | 通过逐问对话，将粗略想法转化为经过验证的设计文档，先设计再实现 |

## 支持的 Provider

| Provider | 值 | Hook 支持 | 安装方式 |
|----------|----|----------|---------|
| Claude Code | `claude`（默认） | 完整（插件 hooks） | 插件市场 |
| Gemini CLI | `gemini` | 完整（扩展 hooks） | `gemini extensions install` |
| OpenCode | `opencode` | 完整（npm 插件 hooks） | npm 插件 |
| OpenAI Codex | `codex` | 仅 Prompt 驱动 | Agent Skill |

**Hook 支持等级说明：**
- **完整**：自动角色注入、brainstorming gate、质量检查、任务归档、空闲通知
- **仅 Prompt 驱动**：通过角色 prompt 约定保证 hook 行为（无自动拦截）

## 高级

### 角色解析

创建 worker 或引用角色时，工具按**项目优先**的顺序解析角色：

1. **项目级**：`.agents/teams/<role-name>/`
2. **全局**：`~/.agents/roles/<role-name>/`

全局角色**原地引用**（不复制到项目中）。`worker.yaml` 会记录 `role_scope` 和 `role_path`，后续操作（重新打开、prompt 注入）将继续使用正确的来源。

### CLI 命令参考

所有命令需在 Git 仓库内运行。

#### 角色命令

| 命令 | 说明 |
|------|------|
| `agent-team role list` | 列出 `.agents/teams/` 中可用角色 |
| `agent-team role create <role-name> --description "..." --system-goal "..." [--force]` | 创建或更新角色技能包。`--force` 跳过全局重复检查 |

#### 角色仓库命令（role-repo）

| 命令 | 说明 |
|------|------|
| `agent-team role-repo search <query>` | 基于严格角色路径契约搜索 GitHub 角色 |
| `agent-team role-repo add <source> [--role <name>...] [--list] [-g] [-y]` | 从 `owner/repo` 或 GitHub URL 发现并安装角色 |
| `agent-team role-repo list [-g]` | 查看所选 scope 下已安装的仓库角色 |
| `agent-team role-repo remove [roles...] [-g] [-y]` | 删除已安装角色并清理锁文件条目 |
| `agent-team role-repo check [-g]` | 用远端目录哈希检查锁文件条目是否可更新 |
| `agent-team role-repo update [-g] [-y]` | 更新有变更的角色 |

接受的远端角色路径契约：

- `skills/<role>/references/role.yaml`
- `.agents/teams/<role>/references/role.yaml`

#### Worker 命令

| 命令 | 说明 |
|------|------|
| `agent-team worker create <role-name> [provider] [--model <model>] [--new-window]` | 创建 worker 并打开会话、安装技能、启动 AI |
| `agent-team worker open <worker-id> [provider] [--model <model>] [--new-window]` | 重新打开已有 worker 会话 |
| `agent-team worker assign <worker-id> "<description>" [provider] [--proposal <file>] [--design <file>] [--model <model>] [--new-window]` | 创建任务变更并通知 worker |
| `agent-team worker status` | 查看 workers、角色、运行状态、技能数量与活跃变更 |
| `agent-team worker merge <worker-id>` | 将 `team/<worker-id>` 合并到当前分支 |
| `agent-team worker delete <worker-id>` | 删除 worker 的 worktree、分支和配置 |

#### 通信命令

| 命令 | 说明 |
|------|------|
| `agent-team reply <worker-id> "<answer>"` | 向 worker 会话发送 `[Main Controller Reply]` |
| `agent-team reply-main "<message>"` | worker 向主控发送 `[Worker: <worker-id>]` |

### Role Repo 锁文件

- 项目锁文件：`roles-lock.json`
- 全局锁文件：`~/.agents/.role-lock.json`
- 项目安装目录：`.agents/teams/<role>/`
- 全局安装目录：`~/.agents/roles/<role>/`

### 目录结构

```
项目根目录/
├── .agents/
│   └── teams/
│       └── <role-name>/                 <- 项目级角色
│           ├── SKILL.md
│           ├── system.md
│           └── references/role.yaml
├── gemini-extension.json                <- Gemini CLI 扩展清单
├── GEMINI.md                            <- Gemini CLI 上下文文件
├── hooks/
│   └── hooks.json                       <- Claude + Gemini 共用 hooks
├── adapters/
│   └── opencode/                        <- OpenCode npm 插件
└── .worktrees/
    └── <worker-id>/
        ├── worker.yaml
        ├── .claude/skills/
        ├── .codex/skills/
        ├── .gemini/skills/
        ├── CLAUDE.md
        ├── AGENTS.md
        ├── GEMINI.md
        └── .tasks/
            └── changes/

~/.agents/
└── roles/
    └── <role-name>/                     <- 全局角色（原地引用，不复制）
        ├── SKILL.md
        ├── system.md
        └── references/role.yaml
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `AGENT_TEAM_BACKEND` | 终端后端：`wezterm`（默认）或 `tmux` |

## License

MIT
