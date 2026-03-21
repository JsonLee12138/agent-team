# P1 Task Lifecycle & Verification Gate Brainstorming

日期：2026-03-21
角色：general strategist

---

## 1. Problem Statement

当前 `agent-team` 已经为 task 标准化引入了 `verification.md`，但 P1 仍存在两个缺口：

1. task 生命周期还没有真正消费 verification 结果。
2. 状态视图还不能直观看到 task 的验证状态与是否可归档。
3. 当前只有 `archive`，缺少一个“任务终止但非成功闭环”的终态分支。
4. `done` 与 `archive` 的语义重叠，生命周期表达不够清晰。

本轮需要把 `verification.md` 从“标准工件”提升为“状态流转可感知工件”，同时补足 `deprecated` 能力。

---

## 2. Goals

### 本轮目标

- 将 task 生命周期从 `done` 语义调整为 `verifying` 语义。
- 让 `task archive` 感知 `verification.md` 的 `Result`。
- 在默认模式下允许 `pass` / `partial` 归档，阻止 `pending` / `fail` 归档。
- 设计 strict 模式：仅 `pass` 可归档。
- 新增 `deprecated` 终态与独立目录沉淀能力。
- 在 `task list` 中展示 verification 状态与归档可行性。

### 非目标

- 本轮不引入完整可配置 gate 系统。
- 本轮不重写 `verification.md` 模板结构。
- 本轮不实现复杂 markdown AST 解析器。
- 本轮不扩展 roadmap / milestone / phase 层。

---

## 3. Constraints And Assumptions

### Constraints

- 必须保持当前 task-first 模型。
- 必须继续沿用 `.agent-team` 作为统一目录根。
- 必须保持实现尽量轻量，避免把 P1 做成通用规则引擎。
- 必须保留文件工件可审计性，不能通过删除目录来表达废弃。

### Assumptions

- `verification.md` 将继续作为标准 task 工件存在。
- `Result` 段落维持稳定、可解析的文本结构。
- 用户接受保留 `task done` 命令名，但内部语义切换为 `verifying`。
- `deprecated` 是终态，且需要与 `archive` 一样落到独立目录中。

---

## 4. Candidate Approaches

### Approach A — 最小感知型

仅做文件存在性检查：
- `task done/archive` 检查 `verification.md` 是否存在
- 状态视图只显示是否存在 verification

#### 优点
- 改动最小
- 风险最低

#### 缺点
- 无法真正利用 `Result`
- 不能表达 `pass/partial/fail` 差异
- 无法满足用户提出的 `fail` 阻断需求

---

### Approach B — 结果感知型（推荐）

轻量解析 `verification.md` 的 `Result`，将其纳入归档判定与状态展示。
同时调整生命周期为 `verifying`，补上 `deprecated` 终态。

#### 优点
- 直接满足 P1 目标
- 生命周期语义更清晰
- 改动仍然可控
- 为后续 strict gate 保留扩展点

#### 缺点
- 需要更新状态枚举、命令语义、测试和展示逻辑
- 需要定义 `partial` 在默认模式和 strict 模式下的差异

---

### Approach C — 预埋完整 gate 引擎

本轮直接引入更多 gate 条件、模式开关、策略配置。

#### 优点
- 未来扩展性最强

#### 缺点
- 远超 P1 范围
- 复杂度显著增加
- 会把当前需求变成配置系统设计问题

---

## 5. Recommended Design

选择 **Approach B — 结果感知型**。

原因：
- 能直接满足当前用户需求；
- 能消除 `done` / `archive` 的语义重叠；
- 能把 `verification.md` 从“附属工件”升级为“真实闭环工件”；
- 仍然保持实现轻量，不把 P1 扩大成完整 gate 引擎。

---

## 6. Lifecycle Design

### 新状态枚举

- `draft`
- `assigned`
- `verifying`
- `archived`
- `deprecated`

### 状态流转

- `draft -> assigned`
- `assigned -> verifying`
- `verifying -> archived`
- `draft|assigned|verifying -> deprecated`

### 状态语义

- `draft`：任务已创建但未分配
- `assigned`：任务已分配并执行中
- `verifying`：开发交付完成，等待验证结论
- `archived`：验证通过后归档闭环
- `deprecated`：任务不再继续，但保留完整工件

### 关键决策

- 不再保留 `done` 作为长期状态。
- `task done` 命令先保留，但内部行为变为“进入 `verifying`”。
- 真正闭环由 `archived` / `deprecated` 表达。

---

## 7. Directory Model

### 活跃任务

- `.agent-team/task/<task-id>/`

### 成功归档

- `.agent-team/archive/task/<task-id>/`

### 废弃任务

- `.agent-team/deprecated/task/<task-id>/`

### 设计理由

- `deprecated` 与 `archive` 平级，更符合“两个不同终态”的语义。
- 避免把 `deprecated` 混入 active task 列表。
- 保留审计能力与后续回溯能力。

---

## 8. Verification Gate Design

### P1 解析范围

只解析 `verification.md` 中 `## Result` 段落的首个列表值。

支持值：
- `pending`
- `pass`
- `partial`
- `fail`

额外降级规则：
- 文件不存在 -> `missing`
- 格式不合法 / 无法识别 -> `pending`

### 默认归档规则

- `pass` -> 可归档
- `partial` -> 可归档
- `pending` -> 不可归档
- `fail` -> 不可归档
- `missing` -> 不可归档

### strict 模式规则

- 仅 `pass` -> 可归档
- `partial / pending / fail / missing` -> 不可归档

### 设计理由

- 默认模式允许 `partial`，适合“已知残留但可收口”的场景。
- strict 模式为后续更严格治理提供落点。
- `fail` 必须立即形成硬阻断。

---

## 9. Command Behavior Design

### `task done`

#### 行为
- 保留命令入口，减少 CLI 破坏性变更。
- 实际语义改为：将 task 状态设为 `verifying`。
- 输出文案也改为 `verifying` 语义。

#### 理由
- 避免一次性改动命令名造成兼容问题。
- 先完成状态语义迁移，再考虑未来是否新增更精确命令名。

---

### `task archive`

#### 行为
- 归档前读取 verification result。
- 默认模式：
  - `pass` / `partial` 允许归档
  - `pending` / `fail` / `missing` 拒绝归档
- strict 模式：
  - 仅 `pass` 允许归档

#### 失败提示
- `missing`：缺少 verification 工件
- `pending`：验证尚未完成
- `fail`：验证失败，不允许归档
- strict 下 `partial`：strict 模式仅允许 `pass`

---

### `task deprecated`

#### 新增命令
- `agent-team task deprecated <task-id>`

#### 行为
- 将任务移动到 `.agent-team/deprecated/task/<task-id>/`
- 保留 `task.yaml` / `context.md` / `verification.md`
- 更新状态为 `deprecated`
- 记录 `deprecated_at`
- 不要求 `merged_sha`

#### 理由
- 补足“非成功闭环”的终态能力
- 与 `archive` 一样保持文件可追溯性

---

### `task list`

维持简表，但新增两列：
- `Verification`：`missing|pending|pass|partial|fail`
- `Archive Ready`：`yes|no|strict-no`

#### 列语义
- `yes`：默认模式可归档，strict 模式也可归档
- `strict-no`：默认模式可归档，但 strict 模式不可归档（典型为 `partial`）
- `no`：默认模式不可归档

---

### `task show`

继续输出完整 task package，同时补充验证摘要：
- verification 文件是否存在
- 当前 `Result`
- 默认模式下是否允许 archive
- strict 模式下是否允许 archive

---

## 10. Data Model Changes

### `task.yaml` 保留字段

- `task_id`
- `title`
- `role`
- `worker_id`
- `task_path`
- `created_at`
- `assigned_at`
- `archived_at`
- `merged_sha`

### 调整字段

- `done_at` -> `verifying_at`

### 新增字段

- `deprecated_at`

### 状态枚举调整

从：
- `draft`
- `assigned`
- `done`
- `archived`

调整为：
- `draft`
- `assigned`
- `verifying`
- `archived`
- `deprecated`

---

## 11. Error Handling

### `task done`
- 已 `archived` / 已 `deprecated` 的任务，拒绝进入 `verifying`
- 非法状态流转直接报错

### `task archive`
- verification 缺失 -> 报错
- `pending` -> 报错
- `fail` -> 报错
- strict 下 `partial` -> 报错
- 已 `deprecated` / 已 `archived` 的任务不可重复归档

### `task deprecated`
- 已 `archived` / 已 `deprecated` 的任务不可再次 deprecated
- deprecated 不依赖 merge SHA

---

## 12. Validation And Test Strategy

### 生命周期测试

- `assigned -> verifying` 成功
- `verifying -> archived` 在 `pass` 下成功
- `verifying -> archived` 在 `partial` 下默认成功
- `verifying -> archived` 在 strict 下因 `partial` 失败
- `verifying -> archived` 在 `pending` 下失败
- `verifying -> archived` 在 `fail` 下失败
- `draft|assigned|verifying -> deprecated` 成功

### 工件与目录测试

- archived 后保留 `verification.md`
- deprecated 后保留 `verification.md`
- deprecated 后目录进入 `.agent-team/deprecated/task/...`

### 展示测试

- `task list` 正确显示 `Verification` 列
- `task list` 正确显示 `Archive Ready` 列
- `task show` 正确展示 verification 摘要

### 解析测试

- `Result = pass`
- `Result = partial`
- `Result = pending`
- `Result = fail`
- 缺失文件 -> `missing`
- 非法格式 -> `pending`

---

## 13. Risks And Mitigations

### 风险 1：旧语义残留

`done` 语义已经出现在现有代码、测试和文档中，迁移时容易遗漏。

**Mitigation**
- 集中更新状态常量与时间字段命名
- 更新 README / README.zh 与相关测试
- 保留 `task done` 命令名，只迁移其内部语义

### 风险 2：verification 解析过度复杂

如果解析逻辑太“聪明”，容易导致脆弱和难维护。

**Mitigation**
- 仅解析 `## Result` 段落首个列表项
- 无法识别统一降级为 `pending`

### 风险 3：strict 模式语义与默认模式混淆

如果状态视图不清楚，用户可能误解 `partial` 是否可归档。

**Mitigation**
- 在 `task list` 中引入 `strict-no`
- 在 `task show` 中明确分别展示默认 / strict 下的归档结论

### 风险 4：deprecated 与 archive 混淆

如果目录结构或命令语义不清晰，可能让人误以为二者只是别名。

**Mitigation**
- 采用平级目录：`archive/` vs `deprecated/`
- 在文案中明确“成功闭环”与“终止闭环”的区别

---

## 14. Open Questions

当前已收敛完成，无阻塞性开放问题。

后续可在下一阶段继续讨论：
- strict 模式的配置入口放在哪里
- 是否新增更准确的命令名（例如 `task verify` / `task submit`）
- `deprecated` 是否需要理由字段或附加说明工件

---

## 15. Final Recommendation

P1 应该聚焦于一件事：

**把 task 生命周期从“状态记录”升级为“带验证感知的交付闭环”。**

最合适的落地方式是：
- 用 `verifying` 替代 `done`
- 用 `verification.Result` 驱动 `archive`
- 允许 `partial` 在默认模式下归档
- 在 strict 模式下收紧为 `pass`-only
- 新增 `deprecated` 作为独立终态和独立目录沉淀能力
- 在 `task list` 中直接展示 verification 与 archive readiness

这样既满足当前需求，也为后续 gate 体系预留了明确扩展点。
