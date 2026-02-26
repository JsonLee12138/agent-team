# agent-team v2 设计文档

## 概述

重构 agent-team，引入"角色 (Role)"和"员工 (Worker)"双层模型。角色是 skill package 定义，由 AI 通过 role-creator skill 创建；员工是角色的运行实例，由 CLI 管理生命周期。

## 核心概念

### 角色 (Role)

- 一个角色 = 一个 skill package（SKILL.md + system.md + references/role.yaml）
- 存放位置可选：
  - `agents/teams/<role-name>/` — 团队使用
  - `skills/<role-name>/` — 开源发布
- 由 AI 通过 role-creator skill 创建/编辑，CLI 不参与角色创建
- 全局已安装的角色 skill 可在创建时直接复制到 `agents/teams/`

### 员工 (Worker)

- 一个员工 = 角色的一个实例，运行在独立 worktree 中
- ID 格式：`<role-name>-<3位序号>`（如 `frontend-dev-001`）
- 一个员工只能关联一个角色，一个角色可以有多个员工
- worktree 路径：`.worktrees/<worker-id>/`
- 分支名：`team/<worker-id>`

## 目录结构

### 主仓库

```
agents/
  teams/
    frontend-dev/              <- 角色定义
      SKILL.md
      system.md
      references/role.yaml
    backend-dev/
      SKILL.md
      system.md
      references/role.yaml
  workers/
    frontend-dev-001/          <- 员工配置
      config.yaml
    frontend-dev-002/
      config.yaml
```

### 员工 Worktree

```
.worktrees/<worker-id>/
  .gitignore                   <- 排除 .gitignore, .claude/, .codex/, openspec/
  .claude/
    skills/                    <- 从角色定义动态复制的 skills
      <role-skill>/
      <dependency-skill-1>/
  .codex/
    skills/                    <- 与 .claude/skills/ 内容一致
      <role-skill>/
      <dependency-skill-1>/
  CLAUDE.md                    <- 从角色 system.md 生成
  openspec/
    changes/<task-slug>/
      design.md                <- 脑暴结果（从 docs/brainstorming/ 复制）
      proposal.md              <- 工作需求
      specs/
      tasks.md
```

### Worker config.yaml

```yaml
worker_id: frontend-dev-001
role: frontend-dev
default_provider: claude
default_model: ""
pane_id: ""
controller_pane_id: ""
created_at: "2026-02-26T10:00:00Z"
```

### Worktree .gitignore

```
.gitignore
.claude/
.codex/
openspec/
```

## CLI 命令设计

Go CLI 使用 cobra 框架，二级子命令结构，全新不兼容旧版。

### Worker 命令

| 命令 | 功能 |
|------|------|
| `agent-team worker create <role-name>` | 创建员工：查找角色定义，创建 worktree + 分支，生成 .gitignore，记录 config |
| `agent-team worker open <worker-id> [provider] [--model] [--new-window]` | 打开会话：动态复制 skills 到 .claude/skills/ 和 .codex/skills/，生成 CLAUDE.md |
| `agent-team worker assign <worker-id> "desc" [--proposal file] [--design file]` | 分配任务：在 openspec/changes/ 创建 design.md 和 proposal.md |
| `agent-team worker status` | 显示所有 worker 状态 |
| `agent-team worker merge <worker-id>` | 合并 worker 分支到当前分支 |
| `agent-team worker delete <worker-id>` | 关闭会话、删除 worktree 和分支 |

### Role 命令

| 命令 | 功能 |
|------|------|
| `agent-team role list` | 列出 agents/teams/ 下所有可用角色 |

### 通信命令

| 命令 | 功能 |
|------|------|
| `agent-team reply <worker-id> "msg"` | 向 worker 会话发送消息 |
| `agent-team reply-main "msg"` | worker 向主控发送消息 |

## 命令详细流程

### worker create

1. 检查 `agents/teams/<role-name>/` 是否存在
2. 如果不存在，检查全局 skills 目录是否有该角色，提示复制
3. 计算下一个序号（扫描 `agents/workers/<role-name>-*`）
4. 创建 worktree `.worktrees/<worker-id>/` 和分支 `team/<worker-id>`
5. 生成 `.gitignore`
6. 创建 `agents/workers/<worker-id>/config.yaml`

### worker open

1. 从 `agents/workers/<worker-id>/config.yaml` 读取关联角色名
2. 从 `agents/teams/<role-name>/references/role.yaml` 读取 skills 列表
3. 复制角色 skill 和依赖 skills 到 `.claude/skills/` 和 `.codex/skills/`
4. 从角色 `system.md` 生成 CLAUDE.md
5. 打开终端会话（WezTerm/tmux）

### worker assign

1. 生成 task slug（timestamp + 描述摘要）
2. 在 worktree 的 `openspec/changes/<task-slug>/` 创建目录
3. 如果提供 `--design` 文件，复制为 `design.md`
4. 如果提供 `--proposal` 文件，复制为 `proposal.md`
5. 通知运行中的会话 `[New Change Assigned]`

## role-creator 改造

### 改动范围

1. `create_role_skill.py` 新增 `--target-dir` 参数：
   - `skills`（默认，开源发布）
   - `agents/teams`（团队使用）
   - 支持自定义路径
2. SKILL.md 更新生成命令示例，包含目标目录选项
3. 创建流程中新增目标目录询问步骤

## agent-team SKILL.md 改造

新的 SKILL.md 需要包含：

1. **概述**：角色 + 员工双层模型说明
2. **角色管理**（AI 流程）：指引使用 role-creator skill，全局角色检测
3. **员工管理**（CLI 命令）：所有 worker 子命令说明
4. **Brainstorming 流程**：保留 HARD-GATE，脑暴结果双存储
5. **任务完成规范**：
   - 完成后执行 `/openspec archive`
   - 完成后用 `reply-main` 通知主控（除非指定不需要）
6. **通信协议**：reply / reply-main

## 任务分配数据流

```
1. 脑暴 -> docs/brainstorming/YYYY-MM-DD-<topic>.md（主仓库存档）
2. 分配 -> agent-team worker assign <worker-id> "desc" \
            --design docs/brainstorming/xxx.md \
            --proposal /tmp/proposal.md
3. CLI 复制:
   - design.md -> .worktrees/<worker-id>/openspec/changes/<task>/design.md
   - proposal.md -> .worktrees/<worker-id>/openspec/changes/<task>/proposal.md
4. 员工执行任务
5. 完成 -> /openspec archive + reply-main 通知
```

## 完整流程示例

```
1. 用户对 AI："创建一个前端开发角色"
   -> AI 使用 role-creator skill
   -> 用户选择目标目录 agents/teams/
   -> 生成 agents/teams/frontend-dev/

2. agent-team worker create frontend-dev
   -> 检测到 agents/teams/frontend-dev/ 存在
   -> 创建 agents/workers/frontend-dev-001/config.yaml
   -> 创建 worktree .worktrees/frontend-dev-001/
   -> 创建分支 team/frontend-dev-001
   -> 生成 .gitignore

3. 用户进行 brainstorming（AI 辅助）
   -> 脑暴结果存到 docs/brainstorming/2026-02-26-login-page.md

4. agent-team worker assign frontend-dev-001 "实现登录页面" \
     --design docs/brainstorming/2026-02-26-login-page.md \
     --proposal /tmp/proposal.md
   -> 复制 design.md 和 proposal.md 到 worktree

5. agent-team worker open frontend-dev-001 claude
   -> 复制 skills 到 .claude/skills/ 和 .codex/skills/
   -> 生成 CLAUDE.md -> 打开终端会话

6. 员工完成任务
   -> /openspec archive
   -> agent-team reply-main "登录页面已完成"

7. agent-team worker merge frontend-dev-001
   -> 合并 team/frontend-dev-001 到 main
```

## 技术选型

- CLI: Go + cobra 框架
- 构建: 本仓库 cmd/agent-team/
- 分发: Homebrew tap（与现有保持一致）
- 终端: WezTerm（默认）/ tmux（通过 AGENT_TEAM_BACKEND 环境变量）
