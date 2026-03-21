# verification.md Brainstorming

日期：2026-03-21
角色：general strategist

## 问题与目标

当前 `agent-team` 的 task 工件只有 `task.yaml` 与 `context.md`，缺少独立、标准化的验收工件。

本轮 P0 目标：把 `verification.md` 纳入 task 标准工件体系，先补齐“验收合同 + 验收结果”闭环，但不在这一轮承担执行流程控制。

## 约束与边界

- 本轮只考虑 `verification.md`
- 任务执行方式、TDD、`qa-expert` 调用、E2E 默认策略等执行规则，不放在 `verification.md` 内部承担
- 上述执行约束应在现有 skills 中修改
- `verification.md` 只负责：
  - 验收标准
  - 测试范围边界
  - 实际检查记录
  - 最终验收结果

## 候选方案

### 方案 A：独立 `verification.md`，task 创建时自动生成（最终采用）

特点：

- 与 `task.yaml`、`context.md` 并列
- 从 task 创建时就存在
- 作为 task 的标准工件

优点：

- 工件边界清晰
- 最符合“task 闭环”
- 后续 P1 可基于该文件做状态感知

缺点：

- 早期会有多一个文件的维护成本

### 方案 B：把 verification 信息并入 `task.yaml`

未采用。

原因：

- 会把“验收工件”退化成元数据字段
- 人工编辑体验差
- 不符合当前要强调的独立工件语义

### 方案 C：在 `task done` 时再生成 `verification.md`

未采用。

原因：

- 不利于 task 创建即闭环
- 容易让验收继续散落在聊天中
- 不符合“标准工件”定位

## 推荐设计

### 工件位置

每个 task 目录固定包含：

```text
.agent-team/task/<task-id>/
  ├── task.yaml
  ├── context.md
  └── verification.md
```

### 工件职责分工

- `task.yaml`：生命周期状态与结构化元数据
- `context.md`：背景、范围、设计输入、任务定义
- `verification.md`：验收合同、测试范围、检查记录、结果

### `verification.md` 模板

```md
# Verification

## Acceptance Criteria
- TODO

## Test Scope
- Unit Test Coverage Required: yes
- E2E Required: no

## Checks Performed
- Not run yet.

## Result
- pending

## Issues
- None.

## Verified By
- qa

## Verified At
- TODO
```

### 字段说明

- `Acceptance Criteria`
  - 记录交付必须满足的验收条件
- `Test Scope`
  - 只记录验收到哪一层
  - 当前确认保留 `E2E Required: no` 默认值
- `Checks Performed`
  - 记录实际执行过的验证动作
- `Result`
  - 建议枚举值：`pending / pass / partial / fail`
- `Issues`
  - 记录遗留问题、风险、偏差
- `Verified By`
  - 默认 `qa`
  - 也允许人工验收时写具体人或 `human`
- `Verified At`
  - 记录验收完成时间

### 明确不保留的字段

- `Contract Owner`
  - 不保留
  - 因为工件语义已经默认偏向 QA / 人工验收视角，重复声明价值不高

## 风险与缓解

### 风险 1：`Acceptance Criteria` 与 `context.md` 中的 Acceptance 有重叠

缓解：

- 接受轻微重复
- 语义上区分：
  - `context.md` 是任务定义
  - `verification.md` 是验收合同

### 风险 2：后续 P1 需要解析 markdown

缓解：

- 现在先把 `Result` 与 `Test Scope` 写成稳定、简单格式
- 后续只做最小解析，不做复杂 markdown 解释器

### 风险 3：执行规则可能被误塞进 verification

缓解：

- 明确规定：
  - `verification.md` 只管合同与结果
  - 执行规则去改现有 skills

## 验证/测试策略

这一轮设计验证标准是：

- task 创建时可以稳定生成 `verification.md`
- 模板字段足够支撑 QA/人工验收
- 不把执行流程逻辑混入验收工件
- 能为后续 P1 的 `done/archive/status` 感知留出稳定接口

## 已确认结论

- 先只做 P0
- 术语定义放进 README
- `verification.md` 是 task 标准工件
- 任务创建时自动生成 `verification.md`
- `E2E Required: no` 作为默认合同字段保留
- `Contract Owner` 不保留
- `Verified By` 保留，默认 `qa`，也可人工填写
- `Verified At` 保留
- 执行流程相关规则在现有 skills 中修改，不塞进 `verification.md`
