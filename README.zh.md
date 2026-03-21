# agent-team

[English](./README.md) | 中文

**以手术般的精度编排你的 AI 劳动力。** 🚀

`agent-team` 是一个多智能体开发管理器，采用 **Role + Worker** 模型，在隔离的 Git worktree 和专用终端会话中运行 AI 智能体。

- **🎭 Role (角色)**: 可复用的技能包 (`.agents/teams/`)，定义目标、提示词和工具。
- **🛠️ Worker (工作者)**: 隔离的运行实例 (`.worktrees/`)，拥有独立的逻辑分支和会话。

---

## 目录

- [安装](#️-安装)
  - [AI 自助安装（推荐）](#-ai-自助安装推荐)
  - [手动安装](#-手动安装)
- [升级](#-升级)
- [快速开始](#-快速开始)
- [工具箱](#-工具箱)
- [支持的 Provider](#-支持的-provider)
- [高级用法](#️-高级用法)
  - [CLI 命令参考](#cli-命令参考)
  - [目录结构](#目录结构)
  - [环境变量](#环境变量)
- [License](#-license)

---

## 🛠️ 安装

### 🤖 AI 自助安装（推荐）
让你的 AI 智能体自行完成安装，只需两步：

1. **安装 Skill**:
   ```bash
   npx skills add JsonLee12138/agent-team -a <platform> -y
   ```
   将 `<platform>` 替换为你的 Provider：`claude`、`gemini`、`opencode` 或 `codex`。
2. **对智能体说**:
   > "安装 agent-team 并初始化项目。"

---

### 📦 手动安装

| 方式 | 命令 |
| :--- | :--- |
| **Homebrew** | `brew tap JsonLee12138/agent-team && brew install agent-team` |
| **Go Install** | `go install github.com/JsonLee12138/agent-team@latest` |
| **Claude Plugin** | `/plugin marketplace add JsonLee12138/agent-team` |
| **Gemini Ext** | `gemini extensions install https://github.com/JsonLee12138/agent-team` |
| **OpenCode Plugin** | `{ "plugin": ["opencode-agent-team"] }` in `~/.config/opencode/opencode.json` |

---

## 🔄 升级

| 方式 | 命令 |
| :--- | :--- |
| **Claude Plugin** | `/plugin marketplace update agent-team` |
| **Skill** | `npx skills add JsonLee12138/agent-team -a '*' -y` |
| **Homebrew** | `brew update && brew upgrade agent-team` |
| **Go Install** | `go install github.com/JsonLee12138/agent-team@latest` |

---

## 🚀 快速开始

完全通过自然语言管理你的团队。你的 AI 智能体会为你处理底层命令。

### 1. 定义团队
> "创建一个 **frontend-architect** 角色，负责管理我们的 UI 架构。"

### 2. 安装精选角色 (可选)
从本仓库拉取专家预设的角色：
```bash
agent-team role-repo add JsonLee12138/agent-team
```

### 3. 启动 Worker
> "使用 Claude 为 **frontend-architect** 创建一个 worker。"
*这会打开一个新的终端窗口，并进入隔离的 worktree。*

### 4. 分配任务 & 头脑风暴
> "分配 **frontend-architect-001** 负责设计新的登录流程。"
*在改动代码前，智能体会先通过对话产出设计文档。*

### 5. 合并与清理
> "合并 **frontend-architect-001** 并删除该 worker。"

---

## 🧰 工具箱

### 内置角色
位于 `.agents/teams/`:
- `pm`: 产品经理，负责需求梳理。
- `frontend-architect`: 前端架构，负责 UI/UX 顶层设计。
- `vite-react-dev`: 专为 Vite + React 开发优化。
- `pencil-designer`: Pencil UI 设计工具专家。

### 内置 Skills
- `task-orchestrator`: 任务生命周期入口。
- `workflow-orchestrator`: 治理型 workflow plan 入口。
- `worker-dispatch`: controller 侧 worker 打开/回复入口。
- `worker-recovery`: worker 恢复当前任务入口。
- `worker-reply-main`: worker -> main 汇报入口。
- `context-cleanup`: 会话清理 + 索引优先文件重锚。
- `task-inspector`: 只读任务查看。
- `role-repo-manager`: role 来源管理。
- `catalog-browser`: 只读 catalog 浏览。
- `project-bootstrap`: `init` / `migrate` 入口。
- `rules-maintenance`: `rules sync` 入口。
- `skill-maintenance`: skill 缓存维护。
- `worker-inspector`: 只读 worker 状态查看。
- `role-browser`: 只读本地 role 浏览。
- `role-creator`: 交互式构建新的智能体角色。
- `brainstorming`: 在实现前通过对话验证想法，并支持规划对象目标选择（`roadmap` / `milestone` / `phase` / `task` / `generic topic`）以及明确的保存目标选择（`docs/brainstorming/`、对象目录、自定义目录或仅输出不保存）。
- `agent-team`: 已降级为兼容导航壳，负责把旧入口导向场景化 skills。
- `strategic-compact`: 已废弃的兼容壳，统一迁移到 `context-cleanup`。

---

## 🤖 支持的 Provider

| Provider | CLI 值 | 集成方式 |
| :--- | :--- | :--- |
| **Claude Code** | `claude` | Plugin |
| **Gemini CLI** | `gemini` | Extension |
| **OpenCode** | `opencode` | NPM Plugin |
| **OpenAI Codex** | `codex` | 仅 Prompt |

---

## ⚙️ 高级用法

<details>
<summary><b>📖 CLI 命令参考</b></summary>

### 角色管理
- `agent-team role list`: 列出本地角色。
- `agent-team role create <name>`: 创建新的角色包。
- `agent-team role-repo add <owner/repo>`: 从 GitHub 安装角色。

### Worker 操作
- `agent-team worker create <role> [--provider <provider>] [--model <model>]`: 创建新的 worker（不启动会话）。
- `agent-team worker open <worker-id> [--provider <provider>] [--model <model>] [--new-window]`: 启动或重新打开 worker 会话。
- `agent-team worker close <worker-id>`: 关闭 worker 会话（不删除 worker）。
- `agent-team worker status`: 查看活跃的 worker 和任务。
- `agent-team worker assign <id> "<task>"`: 分配工作。
- `agent-team worker merge <id>`: 合并 worker 变更（不关闭会话）。
- `agent-team worker delete <id>`: 删除 worker 及其工作树。

### 通信
- `agent-team reply <id> "<msg>"`: 向 worker 发送消息。
- `agent-team reply-main "<msg>"`: Worker 向主控回传消息。

### 规划工件
- `agent-team planning create --kind <roadmap|milestone|phase> "<title>"`: 创建规划工件。
- `agent-team planning list [--kind <kind>] [--lifecycle <planning|archived|deprecated>]`: 列出规划工件。
- `agent-team planning show <id>`: 查看规划工件及引用检查结果。
- `agent-team planning move <id> --to <planning|archived|deprecated>`: 在不同生命周期目录间迁移规划工件。

### 任务工件
每个活跃 task 包固定包含 `.agent-team/task/<task-id>/` 下的三个标准工件：
- `task.yaml`：生命周期元数据与状态。
- `context.md`：背景、范围、约束与设计输入。
- `verification.md`：验收合同、测试范围边界、实际检查记录与最终验收结果。

执行 `agent-team task create` 时会自动生成 `verification.md`。默认模板保留 `E2E Required: no`、`Verified By: qa` 和 `pending` 结果，便于后续由 QA 或人工补全验收记录。

生命周期摘要：
- `task done` 现在表示把任务从 `assigned` 推进到 `verifying`。
- `task archive` 会读取 `verification.md`，只有 `## Result` 通过 gate 才允许归档。
- `task deprecated` 会把未完成或放弃的任务移动到 `.agent-team/deprecated/task/<task-id>/`，同时保留完整 task 包。

</details>

<details>
<summary><b>📂 目录结构</b></summary>

```
项目根目录/
├── .agent-team/task/                  <- 活跃 Task 包
├── .agent-team/planning/roadmaps/     <- 活跃 roadmap 工件
├── .agent-team/planning/milestones/   <- 活跃 milestone 工件
├── .agent-team/planning/phases/       <- 活跃 phase 工件
├── .agent-team/archive/task/          <- 已归档 Task 包
├── .agent-team/archive/roadmaps/      <- 已归档 roadmap 工件
├── .agent-team/archive/milestones/    <- 已归档 milestone 工件
├── .agent-team/archive/phases/        <- 已归档 phase 工件
├── .agent-team/deprecated/task/       <- 已废弃 Task 包
├── .agent-team/deprecated/roadmaps/   <- 已废弃 roadmap 工件
├── .agent-team/deprecated/milestones/ <- 已废弃 milestone 工件
├── .agent-team/deprecated/phases/     <- 已废弃 phase 工件
├── .agents/teams/                     <- 项目专属角色
├── .worktrees/                        <- 隔离的 worker 工作区
├── roles-lock.json                    <- 远程角色版本锁
└── gemini-extension.json              <- 扩展清单
```

</details>

<details>
<summary><b>🌐 环境变量</b></summary>

| 变量 | 默认值 | 说明 |
| :--- | :--- | :--- |
| `AGENT_TEAM_BACKEND` | `wezterm` | 终端：`wezterm` 或 `tmux`。 |
| `AGENT_TEAM_ROLE_HUB_URL` | `https://...` | 埋点上报地址。 |
| `AGENT_TEAM_ROLE_HUB_DEBUG` | `0` | 若设为 `1` 则等待上报完成。 |

</details>

---

## 📄 License

MIT © [JsonLee](https://github.com/JsonLee12138)
