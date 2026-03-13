# Brainstorming: Remove Hooks System

**Date:** 2026-03-13
**Role:** General Strategist
**Status:** Approved

## Problem Statement

agent-team 项目当前同时使用 hooks（自动化事件触发）和 skills+CLI（手动调用）两套机制。经分析，大部分 hooks 功能已被 skills/CLI 覆盖，且 hooks 仅在 worktree 环境下生效，非 worktree 环境全部静默跳过。项目需要简化架构，移除 hooks 依赖。

## Goals

1. 完全移除所有 hooks 相关代码和配置
2. Plugin 变为纯 skills+CLI 项目
3. 不影响现有 CLI 和 skills 功能
4. 清理关联文档

## Constraints & Assumptions

- Subagent 功能暂不需要，可直接移除
- `ListActiveChanges()`、`TasksDir()`、`ChangeDirPath()` 可能被 task 命令共用，需确认后保留
- `InjectRolePromptWithPath()` 被 worker create/open 使用，必须保留
- 版本号需要 bump（移除功能属于 breaking change）

## Candidate Approaches

### Approach A: One-Shot Clean Removal (Recommended)

一次性删除所有 hooks 相关文件和代码。

**Pros:**
- 代码库干净，无死代码
- 维护成本降至零

**Cons:**
- 变更量较大
- 需仔细确认共用函数

### Approach B: Gradual Removal

先设置 `"enabled": false` 禁用，后续逐步清理。

**Pros:** 风险低，可随时回退
**Cons:** 留下大量死代码

### Approach C: Remove + Migrate Guards to Skills

移除 hooks，但将 brainstorming gate 和 quality check 迁移到 skill prompt 约束。

**Pros:** 不丢失守卫能力
**Cons:** 需要额外 skill 改造

## Recommended Design: Approach A

### Hook 功能 vs Skills/CLI 替代分析

| Hook | 功能 | Skills/CLI 替代方案 | 替代程度 |
|------|------|-------------------|---------|
| `session-start` (role-inject) | 自动注入 role prompt | `worker open` 已经做了注入 | 完全替代（hook 是冗余的二次注入） |
| `stop` (stop-guard) | 未归档 change 警告 | `worker status` / `task list` | 可替代（需主动查看） |
| `pre-tool-use` (brainstorming-gate) | design.md 存在性检查 | `brainstorming` skill 通过 prompt 引导 | 部分替代（软约束 vs 硬约束） |
| `post-tool-use` (post-edit-check) | 自动运行 quality checks | 无直接替代，需手动执行 | 不能替代 |
| `task-completed` (task-archive) | 自动归档 change | `task archive` CLI | 可替代（需手动执行） |
| `subagent-start` (subagent-context) | 子 agent 上下文注入 | 无替代（暂不需要此功能） | N/A |
| `teammate-idle` | 通知 main worker 空闲 | 无直接替代 | N/A（暂不需要） |

### 删除清单

#### 1. 文件删除（直接删除）

| 文件 | 说明 |
|------|------|
| `hooks/hooks.json` | Hook 配置文件 |
| `cmd/hooks.go` | Hook 命令注册（`newHookCmd()`） |
| `cmd/hook_session_start.go` | SessionStart hook 实现 |
| `cmd/hook_stop.go` | Stop hook 实现 |
| `cmd/hook_pre_tool_use.go` | PreToolUse hook 实现 |
| `cmd/hook_post_tool_use.go` | PostToolUse hook 实现 |
| `cmd/hook_task_completed.go` | TaskCompleted hook 实现 |
| `cmd/hook_subagent_start.go` | SubagentStart hook 实现 |
| `cmd/hook_teammate_idle.go` | TeammateIdle hook 实现 |

删除后清理空的 `hooks/` 目录。

#### 2. internal/ Hook 专属代码（14 项）

以下函数/类型仅被 hooks 使用，安全删除：

- `ParseHookInput()` — 解析 hook stdin JSON
- `HookInput` struct — hook 输入数据结构
- `ResolveWorktree()` — worktree 检测
- `WorktreeInfo` struct — worktree 信息
- `Provider` type + `ProviderClaude`, `ProviderGemini`, `ProviderOpenCode`, `ProviderUnknown` 常量
- `ParseProvider()` — 解析 provider 字符串
- `DetectProvider()` — 从环境变量检测 provider
- `LoadWorkerFromWorktree()` — 从 worktree 加载 worker config
- `ReadRoleQualityChecks()` — 读取 role.yaml 的 quality_checks
- `ExtractAgentTeamSection()` — 提取 AGENT_TEAM section

需确认后决定：
- `ListActiveChanges()` — 检查是否被 task 命令使用
- `TasksDir()` — 检查是否被 task 命令使用（很可能共用，保留）
- `ChangeDirPath()` — 检查是否被 task 命令使用（很可能共用，保留）

#### 3. internal/ 共用代码（必须保留）

- `ApplyChangeTransition()` — task archive/verify 使用
- `SaveChange()` — task create/done/verify/archive 使用
- `InjectRolePromptWithPath()` — worker create/open 使用
- `Change` struct, `ChangeStatus*` 常量 — task 命令使用
- `WorkerConfig` struct, `LoadWorkerConfig()` — worker 命令使用

#### 4. settings.json 清理

移除 hook 专属字段：
```json
{
  "agent_team": {
    "brainstorming_gate": true,      // 删除 - hook 专属
    "post_edit_check": true,         // 删除 - hook 专属
    "remote_skill_download": true,   // 保留 - skills.go 使用
    "deprecated_path_warning": true  // 保留 - role.go 使用
  }
}
```

#### 5. cmd/root.go 修改

- 移除 `rootCmd.AddCommand(newHookCmd())` 调用
- 移除 `hasAncestorNamed(cmd, "hook")` 跳过 git 检查逻辑（若无其他使用者）

#### 6. 文档更新

| 文件 | 更新内容 |
|------|---------|
| `README.md` | 移除 hooks 支持表格、架构图中的 hooks 部分 |
| `README.zh.md` | 同上（中文版） |
| `GEMINI.md` | 移除 `agent-team hook` 命令文档 |
| `skills/agent-team/references/install-upgrade.md` | 移除 hook 安装说明 |
| `skills/agent-team/SKILL.md` | 移除 hook 相关引用 |
| `adapters/opencode/README.md` | 移除 OpenCode hooks 集成文档 |
| `docs/opencode-codex-skills-extension.md` | 更新 Codex hook 支持说明 |

## Risks & Mitigations

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 误删共用函数导致编译失败 | 高 | 删除前用 `go build` 验证；对 `ListActiveChanges` 等可疑函数逐个确认引用 |
| 丢失 brainstorming gate 硬约束 | 中 | `brainstorming` skill 的 prompt 引导仍然有效（软约束） |
| 丢失自动 quality checks | 中 | 后续可考虑在 skill prompt 中提示用户手动运行检查 |
| 文档更新遗漏 | 低 | 用 grep 搜索所有 "hook" 引用确保覆盖 |

## Validation / Test Strategy

1. **编译验证：** `go build` 确保无编译错误
2. **测试验证：** `go test ./... -v` 确保现有测试通过
3. **功能验证：** 运行 `agent-team worker create/open/assign/status` 确认 worker 流程正常
4. **Grep 检查：** `grep -r "hook" --include="*.go"` 确认无残留引用

## Open Questions

1. `ListActiveChanges()` 是否被 task 命令（如 `task list`）使用？需要在实施时确认
2. 是否需要在此 PR 中 bump 版本号？（移除功能建议 minor bump + changelog）
3. hook 相关的测试文件（如果有）是否也需要删除？
