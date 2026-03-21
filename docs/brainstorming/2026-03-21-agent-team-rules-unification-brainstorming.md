# Agent-Team Rules Unification Brainstorming

## Role

general strategist

## Problem Statement and Goals

当前代码中的文件上下文系统存在目录分散与兼容路径并存的问题：规则入口主要在 `.agents/rules/`，任务工件在 `.agent-team/task/`，worker 还保留 `.agents/workers/` 兼容路径与 fallback。用户希望统一目录到 `.agent-team/`，删除 `.agents/workers` 兼容逻辑，并重构 `init` 相关方案，使其能够自动生成更完整、可扩展、单一职责的规则文档体系。

本次目标：

1. 把规则、上下文和相关工件统一到 `.agent-team/`。
2. 删除 `.agents/workers` 及其 fallback，采用硬切换。
3. 保留根部 provider 文件自动更新，但统一引用 `.agent-team/rules/index.md`。
4. 让 `index.md` 只做入口和简述，公共规范拆分到其他单一职责文件。
5. 重构 `init`：固定生成规则入口与基础规则，使用 AI 生成项目规范文件集合。
6. 新增 `validate` 命令，对生成结果做结构与体量校验，报错但不阻止写入。

## Constraints and Assumptions

- `.agent-team/rules/index.md` 只能作为入口与简单描述，不能承载完整规则正文。
- 所有规则文件必须遵守单一职责原则。
- `core/` 仅承载跨项目稳定规则；项目特定规范放在 `project/`。
- `init` 自动创建公共规范文档，不需要额外询问用户。
- `project` 规则由 AI 生成，但提示词只规定覆盖方向，不应把命令和拆分结构完全写死。
- 对确实需要硬性保证的内容，可通过 `validate` 做检测。
- `validate` 要报错并告诉 AI 需要调整，但不回滚或阻止文件写入。
- 本次范围包含：目录统一、fallback 删除、`init` 重构、规则生成重构、`validate` 新增。

## Candidate Approaches

### Approach A — Recommended

统一到 `.agent-team/`，采用“固定模板 + AI 生成 + validate 校验”的组合方案。

- 固定模板生成：
  - `.agent-team/rules/index.md`
  - `.agent-team/rules/core/*.md`
- AI 生成：
  - `.agent-team/rules/project/*.md`
- 校验：
  - `validate` 检查结构与体量

**Why recommended:**

- 最符合“统一目录”和“单一职责”的目标。
- 基础规则稳定，项目规范灵活。
- 通过 `validate` 把“允许生成，但要求整改”的流程闭环建立起来。

### Approach B

统一目录，但把大部分项目规范逻辑仍然集中到单个 project 规则文件中，再通过 `validate` 控制体量。

**Trade-offs:**

- 实现较简单。
- 但容易违反单一职责，后续继续拆分会反复重构。

### Approach C

全部规则都由 AI 动态生成，包括 index 与 core。

**Trade-offs:**

- 灵活度高。
- 但基础规则容易漂移，稳定性差，不适合作为系统骨架。

## Recommended Design

### 1. Directory Layout

统一目录根为 `.agent-team/`：

- `.agent-team/rules/index.md`
- `.agent-team/rules/core/*.md`
- `.agent-team/rules/project/*.md`
- `.agent-team/task/<task-id>/task.yaml`
- `.agent-team/task/<task-id>/context.md`
- `.agent-team/archive/task/<task-id>/...`
- `.agent-team/main-session.yaml`
- `.agent-team/teams/...`

删除并停止支持：

- `.agents/workers/`
- 所有相关 fallback 读取/写入逻辑

采用硬切换，不做旧 `.agents/*` 自动迁移。

### 2. Rules Structure

#### `index.md`

职责：

- 唯一入口
- 简短说明
- 加载顺序
- 规则映射

不承载完整正文。

#### `core/`

用于跨项目稳定规则，例如：

- `context-management.md`
- `worktree.md`
- `merge-workflow.md`
- `agent-team-commands.md`

#### `project/`

用于当前项目特定规范。

特点：

- 由 AI 在 `init` 时生成多个文件
- 文件数和拆分方式不在代码中写死
- 由提示词引导 AI 按单一职责拆分

内容方向至少应覆盖：

- 当前项目的执行命令与入口
- 工作目录要求
- 失败后的检查方式
- 不允许猜命令
- 项目特有约束与注意事项

### 3. Provider File Behavior

保留根部 provider 文件自动更新：

- `CLAUDE.md`
- `AGENTS.md`
- `GEMINI.md`

但它们统一引用：

- `.agent-team/rules/index.md`

### 4. Init Behavior

`agent-team init` 的职责：

1. 创建 `.agent-team/teams/`
2. 创建 `.agent-team/rules/index.md`
3. 创建 `.agent-team/rules/core/*.md`
4. 通过 AI 生成 `.agent-team/rules/project/*.md`
5. 更新根部 provider 文件，使其引用 `.agent-team/rules/index.md`
6. 不再创建或维护 `.agents/rules/*`
7. 不再创建或维护 `.agents/workers/*`

### 5. AI Prompt Strategy for Project Rules

`init` 的 AI 提示词应规定：

- 需要覆盖哪些方向
- 生成结果必须是多个单一职责文件
- 不要把所有内容塞进一个文件
- 不要猜测项目不存在的命令
- 如果信息不足，应写出如何确认，而不是编造

提示词不应把以下内容完全写死：

- 固定文件数
- 固定文件名
- 固定章节模板
- 固定命令清单

也就是说，提示词约束“思考方向”和“输出原则”，而不是硬编码完整答案。

### 6. Validate Command

新增 `validate` 命令，第一版重点做两类校验。

#### Structure Validation

- `.agent-team/rules/index.md` 是否存在
- `index.md` 中引用的规则文件是否存在
- `core/` 是否包含必需基础规则
- `project/` 是否存在且生成了规则文件
- 目录结构是否符合统一后的约定

#### Size / Scope Validation

- 单文件是否过大
- 标题层级或段落规模是否异常
- 是否疑似违反单一职责

#### Failure Mode

- `validate` 需要返回错误
- 需要明确告诉 AI 哪些文件需要调整
- 但不能阻止已生成文件写入
- 不做自动回滚

这意味着工作流应是：

1. 先生成并写入规则文件
2. 再运行 `validate`
3. 若失败，提示 AI 修正

## Risks and Mitigations

### Risk 1: Existing references still point to `.agents/rules/*`

**Mitigation:**

- 系统性替换源码、测试、文档中的规则路径引用。
- 统一改为 `.agent-team/rules/*`。

### Risk 2: AI-generated project rules may collapse into one oversized file

**Mitigation:**

- 在提示词中明确单一职责要求。
- 通过 `validate` 对体量和结构进行检测。

### Risk 3: Hard cut may break local repos relying on old layout

**Mitigation:**

- 明确本次为硬切换。
- 在 `init`、`rules sync` 或错误信息中明确新目录要求。
- 不再维持兼容层。

### Risk 4: `validate` thresholds may be too strict or too loose

**Mitigation:**

- 第一版只做结构和体量的基础校验。
- 阈值保持保守，后续根据实际使用继续调整。

## Validation and Test Strategy

至少覆盖以下测试：

1. `init` 后 `.agent-team/rules/index.md` 正确生成。
2. `init` 后 `.agent-team/rules/core/*.md` 正确生成。
3. `init` 后 `.agent-team/rules/project/` 下生成项目规范文件。
4. provider 文件正确引用 `.agent-team/rules/index.md`。
5. worker/task/context 路径全部切换到 `.agent-team/`。
6. `.agents/workers` fallback 删除后，相关读取和写入测试通过。
7. `validate` 能检测结构错误。
8. `validate` 能检测体量过大的规则文件。
9. `validate` 失败时返回错误，但不删除已写入文件。

## Open Questions

当前脑暴中已确认关键方向，暂无线程内未决问题；后续实现时需要补充具体 `validate` 阈值与 `init` 提示词细节。