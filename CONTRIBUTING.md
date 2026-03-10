# Contributing to agent-team

感谢你对 agent-team 的关注！本文档帮助你快速上手开发。

## 目录

- [项目概览](#项目概览)
- [开发环境](#开发环境)
- [项目结构](#项目结构)
- [构建与测试](#构建与测试)
- [核心概念](#核心概念)
- [代码规范](#代码规范)
- [Git 工作流](#git-工作流)
- [添加新命令](#添加新命令)
- [添加新角色](#添加新角色)
- [添加新技能](#添加新技能)
- [Hook 机制](#hook-机制)
- [多 Provider 适配](#多-provider-适配)
- [发布流程](#发布流程)

## 项目概览

agent-team 是一个多智能体开发管理器，采用 **Role（角色）+ Worker（实例）** 模型：

- **Role**：可复用的技能包（`.agents/teams/`），定义目标、提示词和工具
- **Worker**：隔离的运行实例（`.worktrees/`），拥有独立的 Git 分支和会话
- **Provider**：支持 Claude Code、Gemini CLI、OpenCode、Codex 等多个 AI 编码工具

## 开发环境

### 前置要求

- Go 1.24.2+
- Git
- macOS 或 Linux（Windows 未测试）
- 终端复用器：WezTerm（推荐）或 tmux

### 快速开始

```bash
# 克隆仓库
git clone https://github.com/JsonLee12138/agent-team.git
cd agent-team

# 构建
make build

# 本地安装（符号链接到 /usr/local/bin）
make install

# 运行测试
make test

# 代码检查
make lint
```

### Makefile 命令一览

| 命令 | 说明 |
|------|------|
| `make build` | 构建二进制到 `output/agent-team` |
| `make test` | 运行所有测试（`go test ./... -v`） |
| `make lint` | 代码静态检查（`go vet`） |
| `make install` | 构建并符号链接到 `/usr/local/bin` |
| `make uninstall` | 移除符号链接，恢复备份 |
| `make clean` | 清理构建产物 |
| `make plugin-pack` | 打包 Claude 插件 |
| `make migrate` | 运行数据迁移 |

> **注意**：`make build` 会自动将 `hooks/hooks.json` 复制到 `internal/hooks.json`（用于 `//go:embed`）。`internal/hooks.json` 已在 `.gitignore` 中排除，不要手动提交。

## 项目结构

```
agent-team/
├── main.go                     # 入口
├── cmd/                        # CLI 命令（Cobra）
│   ├── root.go                 # 根命令、App 初始化
│   ├── worker_create.go        # worker create 命令
│   ├── worker_open.go          # worker open 命令
│   ├── role_create.go          # role create 命令
│   ├── hook_*.go               # 生命周期钩子处理
│   ├── task_*.go               # 任务管理命令
│   └── *_test.go               # 命令测试
├── internal/                   # 核心业务逻辑（不对外暴露）
│   ├── role.go                 # 角色解析、Provider 定义
│   ├── role_create.go          # 角色创建（含 go:embed 模板）
│   ├── git.go                  # Git 操作封装
│   ├── config.go               # WorkerConfig 定义
│   ├── hook.go                 # HookInput 解析、Worktree 检测
│   ├── hooks_install.go        # Hook 安装到 .claude/settings.json
│   ├── task.go                 # Change/Task 类型定义
│   ├── task_store.go           # Change 持久化
│   ├── task_lifecycle.go       # Change 状态机
│   ├── skills.go               # 技能安装
│   ├── templates/              # go:embed 模板
│   │   ├── SKILL.md.tmpl
│   │   ├── system.md.tmpl
│   │   └── role.yaml.tmpl
│   └── *_test.go               # 单元测试
├── hooks/                      # 生命周期钩子定义
│   └── hooks.json              # 所有 Provider 的钩子配置
├── skills/                     # 内置技能包
│   ├── agent-team/             # 主技能（SKILL.md + references/）
│   ├── role-creator/           # 角色创建技能
│   └── brainstorming/          # 头脑风暴技能
├── .agents/teams/              # 内置角色包
│   ├── pm/                     # 产品经理
│   ├── frontend-architect/     # 前端架构师
│   ├── vite-react-dev/         # Vite + React 开发
│   ├── uniapp-dev/             # UniApp 开发
│   ├── pencil-designer/        # Pencil 设计工具专家
│   └── growth-marketer/        # 增长营销
├── adapters/opencode/          # OpenCode 适配器（TypeScript）
├── role-hub/                   # 角色中心 Web 应用（Remix）
├── .github/workflows/          # CI/CD
│   ├── release.yml             # Go 二进制发布（GoReleaser）
│   └── release-opencode.yml    # OpenCode 插件发布（npm）
├── settings.json               # 项目级功能开关
├── Makefile
├── go.mod
└── .goreleaser.yaml
```

### 关键依赖

| 包 | 用途 |
|---|------|
| `github.com/spf13/cobra` | CLI 框架 |
| `gopkg.in/yaml.v3` | YAML 解析 |
| `github.com/AlecAivazis/survey/v2` | 交互式命令行提示 |

## 核心概念

### Role（角色）

角色是可复用的技能包，存放在 `.agents/teams/<role-name>/` 下：

```
.agents/teams/pm/
├── system.md                   # 系统提示词
└── references/
    └── role.yaml               # 角色元数据（名称、技能列表、scope）
```

角色解析优先级：项目级（`.agents/teams/`）→ 全局（`~/.agent-team/roles/`）。

### Worker（实例）

Worker 是角色的运行实例，每个 Worker 拥有：
- 独立的 Git worktree（`.worktrees/<worker-id>/`）
- 独立的 Git 分支（`team/<worker-id>`）
- 独立的终端 pane
- `worker.yaml` 配置文件

Worker ID 格式：`<role-name>-<3位数字>`，如 `pm-001`。

### Change（变更）

Change 是 Worker 中的任务单元，存放在 `.tasks/changes/<name>/change.yaml`。

状态机：`draft` → `assigned` → `implementing` → `verifying` → `done` → `archived`

### Hook（钩子）

生命周期钩子在特定事件触发时执行，定义在 `hooks/hooks.json`。Worker 创建时自动安装到 `.claude/settings.json`。

## 代码规范

### Go 代码

- 使用 `go vet` 做静态检查
- 遵循 Go 标准项目布局：`cmd/` 放命令，`internal/` 放业务逻辑
- 文件命名：`snake_case.go`
- 测试文件：`*_test.go`，使用标准 `testing` 包
- 模板文件通过 `//go:embed` 嵌入二进制
- 错误处理：返回 `fmt.Errorf("context: %w", err)` 包装错误
- Hook 处理函数中错误不应中断会话，返回 `nil` 而非 error

### 命名约定

| 对象 | 格式 | 示例 |
|------|------|------|
| 角色名 | kebab-case | `frontend-architect` |
| Worker ID | `<role>-<NNN>` | `pm-001` |
| Go 文件 | snake_case | `worker_create.go` |
| Git 分支 | `<type>/<desc>` | `fix/worker-completion` |
| Worker 分支 | `team/<worker-id>` | `team/pm-001` |

### 注释语言

代码注释语言与现有代码库保持一致（英文）。

## Git 工作流

### 分支策略

- `main` — 主分支，保持稳定
- `feat/*` — 新功能
- `fix/*` — Bug 修复
- `refactor/*` — 重构
- `docs/*` — 文档
- `chore/*` — 杂务

### Commit Message 格式

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<type>(<scope>): <subject>

[optional body]
```

**type**：`feat`、`fix`、`refactor`、`docs`、`chore`、`test`

**scope**（可选）：`hooks`、`teams`、`marketing`、`role-hub` 等

**示例**：
```
feat(role-hub): add technical SEO — metadata, canonical, sitemap
fix(hooks): install hooks to worktree and auto-complete on session stop
refactor(teams): restructure roles and templatize role injection
docs: enforce explicit platform selection for skill installs
chore: clean up legacy .agents/workers references
```

## 添加新命令

1. 在 `cmd/` 下创建 `<command>.go`，使用 Cobra 定义命令：

```go
// cmd/example.go
package cmd

import "github.com/spf13/cobra"

func newExampleCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "example <arg>",
        Short: "One-line description",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return GetApp(cmd).RunExample(args[0])
        },
    }
}

func (a *App) RunExample(arg string) error {
    // 业务逻辑放在 internal/ 中
    return nil
}
```

2. 在 `cmd/root.go` 的 `RegisterCommands` 中注册命令
3. 业务逻辑放在 `internal/` 中，`cmd/` 只做参数解析和调用
4. 添加对应的 `cmd/example_test.go`

## 添加新角色

1. 在 `.agents/teams/` 下创建角色目录：

```
.agents/teams/<role-name>/
├── system.md                   # 系统提示词
└── references/
    └── role.yaml               # 角色元数据
```

2. 或使用 CLI 创建：

```bash
agent-team role create <role-name>
```

3. `role.yaml` 关键字段：

```yaml
name: my-role
description: "角色描述"
system_prompt_file: system.md
scope:
  in_scope: [...]
  out_of_scope: [...]
skills:
  - "org/repo@skill-name"       # 远程技能
  - "local-skill"               # 本地技能
```

## 添加新技能

技能包存放在 `skills/<skill-name>/`：

```
skills/<skill-name>/
├── SKILL.md                    # 技能定义（触发条件、使用说明）
└── references/                 # 可选：参考文档
    └── *.md
```

`SKILL.md` 是技能的入口文件，定义了触发条件和行为规范。

## Hook 机制

### 架构

```
hooks/hooks.json          → 所有 Provider 的钩子定义（源文件）
  ↓ make build
internal/hooks.json       → go:embed 嵌入到二进制
  ↓ worker create/open
.worktrees/<id>/.claude/settings.json  → 写入 Claude 兼容事件
```

### 支持的事件

| 事件 | Provider | 说明 |
|------|----------|------|
| `SessionStart` | Claude | 会话开始，注入角色提示词 |
| `Stop` | Claude | 会话结束，自动收尾（commit + archive + notify） |
| `PreToolUse` | Claude | 工具调用前，头脑风暴门控 |
| `PostToolUse` | Claude | 工具调用后，质量检查 |
| `SubagentStop` | Claude | 子智能体停止 |
| `SessionEnd` | Gemini | 等同于 Stop |
| `BeforeTool` | Gemini | 等同于 PreToolUse |
| `AfterTool` | Gemini | 等同于 PostToolUse |

### 添加新 Hook

1. 在 `hooks/hooks.json` 中添加事件定义
2. 在 `cmd/` 下创建 `hook_<event>.go` 实现处理逻辑
3. 在 `cmd/root.go` 中注册到 `hook` 子命令
4. 如果是 Claude 兼容事件，更新 `internal/hooks_install.go` 中的 `claudeHookEvents`

### Hook 编写原则

- Hook 不应中断用户会话：错误时返回 `nil`，日志输出到 stderr
- 使用 `internal.ParseHookInput(os.Stdin)` 解析输入
- 使用 `internal.ResolveWorktree(input.CWD)` 检测是否在 agent-team worktree 中
- 非 worktree 环境下静默退出

## 多 Provider 适配

项目支持多个 AI 编码工具 Provider：

| Provider | 启动命令 | Hook 事件格式 |
|----------|----------|---------------|
| `claude` | `claude` | `PreToolUse`、`Stop` 等 |
| `gemini` | `gemini` | `BeforeTool`、`SessionEnd` 等 |
| `opencode` | `opencode` | 通过 adapters/ 适配 |
| `codex` | `codex` | — |

添加新 Provider 时需要：
1. 在 `internal/role.go` 的 `SupportedProviders` 和 `launchCommands` 中注册
2. 在 `hooks/hooks.json` 中添加对应的钩子事件映射
3. 如需特殊适配，在 `adapters/` 下创建适配器

## 发布流程

### Go 二进制发布

1. 打 tag：`git tag v1.x.x`
2. 推送：`git push origin v1.x.x`
3. GitHub Actions 自动触发 GoReleaser，构建多平台二进制并发布到 Homebrew tap

### OpenCode 插件发布

1. 打 tag：`git tag opencode-v1.x.x`
2. 推送后自动发布到 npm

## 许可证

[MIT License](LICENSE) — Copyright (c) 2025 JsonLee12138
