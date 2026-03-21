# Brainstorming: role-creator 懒安装与 skills 结构重构

**Role**: general strategist
**Date**: 2026-03-21

## 1) Problem Statement & Goals

当前 `role-creator` 仅记录 skill 名称，且 worker 创建阶段倾向预装，不符合“按需懒安装”目标。
需要达成：

1. 创建角色时记录 skill 的简短描述（可读、可审阅）。
2. worker 不预装全部 skills，运行时缺失再处理。
3. 安装策略默认项目级，且 worker 不允许全局安装。
4. 旧格式 role 不做迁移脚本，通过 validate 报错并引导用 create 重做。

## 2) Constraints & Assumptions

- 规则体系需符合 `init/sync` 架构（`.agent-team/rules/core` + index 引用）。
- `role.yaml` 作为角色静态定义，不承载运行时安装策略细节。
- 不做兼容老 `skills: []string` 的读取逻辑。
- 不做自动迁移；只做校验与重建提示。

## 3) Candidate Approaches

### A. `skills` + `skill_metadata` 并行（弃用）
- 优点：改动小
- 缺点：重复数据，易漂移

### B. `skills` 单数组对象 + 核心规则懒安装（推荐）
- 优点：单一事实源；数据与行为边界清晰
- 缺点：需要更新解析与校验路径

### C. 只保留字符串列表、描述写文档（弃用）
- 优点：实现快
- 缺点：机器不可读，无法稳定支持后续自动化

**推荐：B**

## 4) Recommended Design

### 4.1 Role 数据结构（`references/role.yaml`）

仅保留对象数组：

```yaml
skills:
  - name: "ui-ux-pro-max"
    description: "UI/UX 设计与实现增强能力"
  - name: "better-icons"
    description: "图标检索与 SVG 获取"
```

> 不再使用 `skills: []string`，不再新增 `skill_metadata`。

### 4.2 运行时行为规则（放 core rules）

新增规则文件：`/.agent-team/rules/core/skill-resolution.md`

规则要点：

1. 任务需要 skill 且本地缺失时，先执行 `find-skills`。
2. 安装仅允许 **项目级**。
3. 项目级安装失败时，输出 warning（含失败原因与后续建议），**不中断当前任务**。
4. 禁止 worker 触发全局安装路径。
5. 任务结束可建议维护者回写 role（通过 `role create` 重建）。

同时在 `/.agent-team/rules/index.md` 增加该规则入口。

### 4.3 Validate 策略（只报错，不迁移）

`validate` 检查 role.yaml 的 `skills` 是否符合对象数组规范：

- `skills` 必须为 list
- 每项必须含 `name`（非空）
- `description` 建议必填（可按后续约束收紧为必填）

发现旧结构（如 `skills: ["vitest"]`）时：

- 直接报错
- 错误文案明确提示：请使用 `agent-team role create ...` 重新生成该 role
- 不做自动迁移，不做 in-place 修复

## 5) Components Impact

- `skills/role-creator/SKILL.md`：更新“选择技能后的写入格式”和“运行时行为指引”
- `internal/templates/role.yaml.tmpl`：`skills` 输出改为对象数组
- role 读取/安装逻辑（当前从 `internal/skills.go` 读取）改为读取 `skills[].name`
- 规则生成源（`internal/rules_v2.go`）增加 `core/skill-resolution.md` 默认内容，并更新 index 默认内容
- validate 逻辑补充 skills 结构检查与错误提示

## 6) Data Flow

1. `role-creator` 使用 `find-skills` 选出技能。
2. 生成 role.yaml：写入 `skills[{name, description}]`。
3. worker 执行任务时，如发现缺失技能：
   - `find-skills`
   - 项目级安装尝试
   - 失败则 warning，继续任务
4. `validate` 扫描 role 结构，不合规即报错并要求重建。

## 7) Error Handling

- 缺失技能安装失败：warning + 继续执行（不阻断）
- 规则文件缺失：由 `init/sync` 重建 core/index
- role.yaml 旧格式：validate 报错并给出 `role create` 重建提示
- 不允许全局安装：发现全局安装路径时直接阻断并提示规则约束

## 8) Testing / Validation Strategy

1. role create 后检查 role.yaml 中 `skills` 为对象数组。
2. 模拟 worker 缺失 skill：验证只走项目级安装。
3. 模拟项目级安装失败：验证 warning 输出与任务不中断。
4. validate 输入旧格式 role：验证报错内容包含“请用 create 重建”。
5. `init` 与 `rules sync` 后检查：
   - `core/skill-resolution.md` 存在
   - `index.md` 含对应入口

## 9) Risks & Mitigations

- 风险：旧 role 全部会被 validate 判错
  - 缓解：文案明确“预期行为”，提供标准重建命令模板
- 风险：AI 执行时误走全局安装
  - 缓解：core rule 明确禁止 + validate/rule review 守护
- 风险：description 质量参差
  - 缓解：role-creator 生成时加入“简短、可判别”的描述约束

## 10) Open Questions

无（关键策略已确认）。
