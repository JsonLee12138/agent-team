# Worker Execution Mode — Brainstorming

**Date:** 2026-03-22
**Role:** General Strategist

## Problem Statement

Worker 在接到任务后有时会主动进入 plan 模式（如 Claude 的 EnterPlanMode），导致规划工作重复甚至越权扩大任务范围。需要在 worker 提示词中明确约束执行行为，使 worker 保持在"执行者"角色而非"规划者"角色。

## Goals

- Worker 接到任务后直接执行，不进入任何规划阶段
- 约束适用于所有 AI provider（Claude / Codex / Gemini / OpenCode）
- 约束与角色定义解耦，不污染角色 skill 内容
- 不清楚任务要求时通过 `agent-team reply-main` 问 controller，而不是自行规划

## Constraints & Assumptions

- 任务由 controller 侧规划完毕后才分配给 worker，worker 不需要重新规划
- 现有注入架构：`InjectRolePrompt` → `InjectSection` 写入 AGENT_TEAM tagged section
- 两个模板路径：`slimRoleSectionTmpl`（有 rules 目录时）和 `legacyRoleSectionTmpl`（fallback）
- PaneSend 消息只做读文件引导，不适合承载行为约束

## Candidate Approaches

### 方案 A：AGENT_TEAM 模板扩展（推荐）

在 `slimRoleSectionTmpl` 和 `legacyRoleSectionTmpl` 的 `**Workflow:**` 列表中，在现有步骤 1 之前插入 **Step 0**，表达执行约束。

**优点：** 改动最小；约束随 role prompt 注入，AI 在会话开始就读到；两个模板路径都覆盖。
**缺点：** 和角色 section 在同一 block 内，不完全独立。

### 方案 B：独立 section tag（WORKER_EXEC）

新增 `WORKER_EXEC` tag，用独立的 `InjectSection` 调用写入，与 AGENT_TEAM section 完全解耦。

**优点：** 解耦好，可单独更新执行约束。
**缺点：** 增加一次 InjectSection 调用和一个新 tag，系统复杂度稍高。

### 方案 C：Rules 文件

放到 `.agent-team/rules/core/worker-execution.md`，通过 rules index 引用。

**优点：** 最解耦，与规则体系一致。
**缺点：** 依赖 AI 主动读取，bootstrap 阶段不保证被加载。

### 方案 D：PaneSend 消息追加

在 `task assign` 的 PaneSend 消息中追加执行约束。

**优点：** 最即时。
**缺点：** 消息可能被 AI 忽略或遗忘；每次 assign 才触发，`worker open` 路径无法覆盖。

## Recommended Design

**方案 A：AGENT_TEAM 模板扩展**

### 改动位置

`internal/role.go` 中的两个模板：`slimRoleSectionTmpl` 和 `legacyRoleSectionTmpl`。

### 具体内容

在 `**Workflow:**` 列表前插入 Step 0，原有步骤编号 +1：

```
**Workflow:**

0. **Execute, do not plan** — Your task has already been planned and assigned by the main controller. Execute it directly.
   - **Never** enter plan mode or any equivalent planning/design phase before working.
   - **Never** expand the task scope beyond what is described in the assigned task.
   - If requirements are unclear, use `agent-team reply-main` to ask the controller — do not start a planning phase yourself.
1. **Match skills first** — Check which of your available skills are relevant to the task before doing any direct work.
2. **Invoke matched skills** — For each relevant skill, invoke it via `/skill-name` or the Skill tool.
3. **Combine skill outputs** — If a task spans multiple skills, invoke them in logical order and integrate their outputs.
4. **Direct work only as fallback** — Only work directly when no available skill covers the requirement.
5. **Dynamic skill discovery** — If no current skill matches, invoke `find-skills` to search for one.
```

### 覆盖范围

| 文件 | 改动 |
|---|---|
| `internal/role.go` | `slimRoleSectionTmpl`：Workflow 列表加 Step 0，原步骤编号 +1 |
| `internal/role.go` | `legacyRoleSectionTmpl`：同上 |

### 不改动

- `cmd/task_assign.go` PaneSend 消息（职责不同，保持不变）
- `.agent-team/rules/` 下任何规则文件（执行约束属于 worker bootstrap，不属于项目规则）
- 角色 skill 文件（`SKILL.md`、`system.md`）

## Risks & Mitigations

| 风险 | 缓解 |
|---|---|
| AI 仍然忽略 Step 0 | Step 0 放在最前，使用 **Never** 加粗强调 |
| 步骤编号 +1 破坏 rules 文件中对步骤号的引用 | rules 文件当前没有引用步骤号，无影响 |
| legacy 模板 fallback 路径遗漏 | 两个模板同步修改，有集成测试覆盖 |

## Validation Strategy

- 构建通过（`go build ./...`）
- 现有 `TestInjectSection`、`TestBuildLegacyRoleSection`、`TestBuildSlimRoleSection` 等集成测试通过
- 手动检查注入后的 CLAUDE.md 包含 Step 0 内容

## Open Questions

无。
