# agent-team

[English](./README.md) | 中文

多智能体开发工作流的 AI 团队角色管理器。采用 **Role + Worker** 双层模型，在隔离 Git worktree 中运行。

- **Role**：角色技能包定义，位于 `.agents/teams/<role-name>/`
- **Worker**：角色运行实例，位于 `.worktrees/<worker-id>/`（例如 `frontend-dev-001`）

## 目录

- [工作原理](#工作原理)
- [环境要求](#环境要求)
- [安装](#安装)
- [升级](#升级)
- [快速开始](#快速开始)
- [自带角色](#自带角色)
- [作为 Skill 使用](#作为-skill-使用)
- [CLI 命令参考](#cli-命令参考)
- [Role Repo 锁文件](#role-repo-锁文件)
- [目录结构](#目录结构)
- [支持的 Provider](#支持的-provider)
- [环境变量](#环境变量)
- [License](#license)

## 工作原理

```
主分支
  ├── .agents/teams/<role-name>/           <- 角色技能包定义
  └── .worktrees/<worker-id>/              <- 隔离运行目录
        └── worker.yaml                    <- worker 配置
```

典型流程：

1. 在 `.agents/teams/` 创建或准备角色
2. 创建并打开 worker：`agent-team worker create <role-name> [provider]`
3. 头脑风暴后分配任务：`agent-team worker assign ...`
4. 合并：`agent-team worker merge <worker-id>`
5. 清理：`agent-team worker delete <worker-id>`

## 环境要求

- Git
- [WezTerm](https://wezfurlong.org/wezterm/) 或 [tmux](https://github.com/tmux/tmux)
- 至少一个 AI provider CLI：[claude](https://github.com/anthropics/claude-code)、[codex](https://github.com/openai/codex) 或 [opencode](https://opencode.ai)
- [Node.js](https://nodejs.org/)（创建 worker 时用于自动安装 OpenSpec 和 `npx skills add`）

## 安装

**注意：** 安装方式因平台而异。Claude Code 有内置插件系统，其他平台使用 Agent Skill 方式。

### Claude Code（插件安装）

```bash
claude plugin add JsonLee12138/agent-team
```

### Agent Skill

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
# 插件
claude plugin update agent-team

# Skill
npx skills add JsonLee12138/agent-team

# Homebrew
brew update && brew upgrade agent-team

# 源码安装
go install github.com/JsonLee12138/agent-team@latest
```

## 快速开始

1. 通过 `role-creator` 在 `.agents/teams/` 生成角色
```bash
python3 skills/role-creator/scripts/create_role_skill.py \
  --repo-root . \
  --role-name frontend-dev \
  --target-dir .agents/teams \
  --description "Frontend role for UI implementation" \
  --system-goal "Ship maintainable frontend features"
```

2. 查看角色列表
```bash
agent-team role list
```

3. 创建并打开 worker（创建 worktree、打开会话、安装技能、启动 AI）
```bash
agent-team worker create frontend-dev claude
```

4. 分配任务
```bash
agent-team worker assign frontend-dev-001 "Implement responsive navbar"
```

5. 合并并删除 worker
```bash
agent-team worker merge frontend-dev-001
agent-team worker delete frontend-dev-001
```

## 自带角色

当前仓库自带 1 个角色：

- `frontend-architect`（路径：`skills/frontend-architect/`）

要在 `agent-team` 中使用它，先复制到 `.agents/teams/`：

```bash
mkdir -p .agents/teams
cp -R skills/frontend-architect .agents/teams/
agent-team role list
```

## 自带 Skills

| Skill | 说明 |
|-------|------|
| `role-creator` | 交互式创建或更新角色技能包 |
| `brainstorming` | 通过逐问对话，将粗略想法转化为经过验证的设计文档，先设计再实现 |

## 作为 Skill 使用

安装 Skill 后，可以直接用自然语言描述操作，例如：

- "创建一个前端架构角色。"
- "给 frontend-architect 创建 worker 并用 codex 打开。"
- "给 frontend-architect-001 分配一个变更任务。"
- "查看 worker 状态。"

主控智能体应在 assign 前完成头脑风暴流程，再创建 OpenSpec change 并通知 worker。

## CLI 命令参考

所有命令需在 Git 仓库内运行。

### 角色命令

| 命令 | 说明 |
|------|------|
| `agent-team role list` | 列出 `.agents/teams/` 中可用角色 |

### 角色仓库命令（role-repo）

| 命令 | 说明 |
|------|------|
| `agent-team role-repo search <query>` | 基于严格角色路径契约搜索 GitHub 角色 |
| `agent-team role-repo add <source> [--role <name>...] [--list] [-g] [-y]` | 从 `owner/repo` 或 GitHub URL 发现并安装角色（多角色时默认交互选择） |
| `agent-team role-repo list [-g]` | 查看所选 scope 下已安装的仓库角色 |
| `agent-team role-repo remove [roles...] [-g] [-y]` | 删除已安装角色并清理锁文件条目（默认交互选择/确认） |
| `agent-team role-repo check [-g]` | 用远端目录哈希检查锁文件条目是否可更新 |
| `agent-team role-repo update [-g] [-y]` | 更新有变更的角色（默认交互选择，`-y` 全量更新） |

接受的远端角色路径契约：

- `skills/<role>/references/role.yaml`
- `.agents/teams/<role>/references/role.yaml`

### Worker 命令

| 命令 | 说明 |
|------|------|
| `agent-team worker create <role-name> [provider] [--model <model>] [--new-window]` | 创建 worker 并打开会话、安装技能、启动 AI |
| `agent-team worker open <worker-id> [provider] [--model <model>] [--new-window]` | 重新打开已有 worker 会话 |
| `agent-team worker assign <worker-id> "<description>" [provider] [--proposal <file>] [--design <file>] [--model <model>] [--new-window]` | 创建 OpenSpec change 并通知 worker |
| `agent-team worker status` | 查看 workers、角色、运行状态、技能数量与活跃变更 |
| `agent-team worker merge <worker-id>` | 将 `team/<worker-id>` 合并到当前分支 |
| `agent-team worker delete <worker-id>` | 删除 worker 的 worktree、分支和配置 |

### 通信命令

| 命令 | 说明 |
|------|------|
| `agent-team reply <worker-id> "<answer>"` | 向 worker 会话发送 `[Main Controller Reply]` |
| `agent-team reply-main "<message>"` | worker 向主控发送 `[Worker: <worker-id>]` |

使用 tmux 后端：命令前添加 `AGENT_TEAM_BACKEND=tmux`。

## Role Repo 锁文件

- 项目锁文件：`roles-lock.json`
- 全局锁文件：`~/.agents/.role-lock.json`
- 项目安装目录：`.agents/teams/<role>/`
- 全局安装目录：`~/.agents/roles/<role>/`

## 目录结构

```
项目根目录/
├── .agents/
│   └── teams/
│       └── <role-name>/
│           ├── SKILL.md
│           ├── system.md
│           └── references/role.yaml
└── .worktrees/
    └── <worker-id>/
        ├── worker.yaml              <- worker 配置
        ├── .claude/skills/
        ├── .codex/skills/
        ├── CLAUDE.md
        ├── AGENTS.md
        └── openspec/
            ├── specs/
            └── changes/
```

## 支持的 Provider

| Provider | 值 |
|----------|----|
| Claude Code | `claude`（默认） |
| OpenAI Codex | `codex` |
| OpenCode | `opencode` |

## 环境变量

| 变量 | 说明 |
|------|------|
| `AGENT_TEAM_BACKEND` | 终端后端：`wezterm`（默认）或 `tmux` |

## License

MIT
