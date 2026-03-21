# 治理内核重构脑暴方案（主线定版）

- 日期：2026-03-20
- 角色：general strategist
- 类型：重构方案脑暴
- 基线文档：`docs/preview_new/2026-03-20-best-solution-layered-governance-brainstorming.md`

## 1) 问题陈述与目标

当前项目需要重构，核心诉求是：

1. 将“治理语义”与“能力实现”彻底解耦。
2. 将 `agent-team` 从“大一统逻辑入口”降级为纯编排壳。
3. 避免 AI 依赖内隐会话上下文，开发与决策只认文本上下文（更安全、可审计）。
4. 以统一治理内核收敛 `memory/rules/task/workflow` 的一致性问题。

## 2) 约束与原则（确认版）

- 规则优先级固定：`Public Rule > Module Rule > Task Rule`
- `Declared-Ref-Strict`：仅对“已声明引用缺失实体”做硬阻断
- `Index-First`：所有对象先入索引，再参与流程
- `No-Owner-Signoff-No-Exit`：无 Owner 签署不得进入下一阶段
- `archived` 数据默认不得用于开发与决策输入
- 如需排查历史遗留问题，仅允许 owner 单签一次性只读放行
- 开发与决策只认文本工件上下文，不依赖 AI 会话内隐上下文

## 3) 推荐架构（通过）

### 3.1 分层边界

- `internal/governance/*`：唯一治理内核
  - `index-registry`
  - `rule-loader`
  - `task-packet`
  - `gate-engine`
  - `advisor`
- `internal/modules/{memory,rules,task,workflow}/*`：能力模块（实现层）
- `internal/orchestrator/*`：跨模块流程编排（用例层）
- `cmd/*`：CLI 入口，仅参数解析与调用用例

### 3.2 关键职责

- 治理契约（`Task Packet / Index / Gate Result / Workflow Plan`）只能由 `governance` 定义。
- `modules` 只实现，不定义治理语义，不互相直接调用。
- 跨模块协同必须经 `orchestrator`。
- `orchestrator` 只调用 `governance` 用例，不直连 modules。
- `skills/agent-team` 定位为纯编排壳，不承载治理或业务逻辑。

## 4) 核心对象

### 4.1 Workflow Plan（命名已定）

不使用“草案/正式”双对象，统一为单一对象生命周期：

- `proposed`
- `approved`
- `active`
- `closed`

说明：
- `advisor` 只负责产出 `Workflow Plan`（`proposed`）
- 只有 Owner 可将其推进到 `approved`
- `workflow` 模块仅负责执行状态机推进

### 4.2 存储形态（已定）

- 模块内保存 `Workflow Plan` 正文
- 根级索引保存摘要与指针

### 4.3 Archived Exception（已定）

- 默认：`archived` 禁止进入开发/决策输入
- 例外：owner 单签 `exception ticket`
- 作用域：单 `task_id` 一次性
- 权限：只读排障
- 失效：首次读取即失效
- 存储：模块级 `exceptions` 文件
- 最小字段：
  - `ticket_id`
  - `task_id`
  - `owner`
  - `reason`
  - `created_at`
  - `used_at`
  - `status`

## 5) 数据流设计（通过）

### 5.1 计划生成链路

1. `cmd` 接收请求并调用 `orchestrator` 用例
2. `orchestrator` 调 `governance/advisor.GenerateWorkflowPlan`
3. `advisor` 仅读取文本工件（index/rules/task packet/evidence）
4. 产出 `Workflow Plan(status=proposed, input_refs, reasons)`
5. 模块写正文，根级写摘要索引

### 5.2 审批与激活链路

6. Owner 审批：`proposed -> approved`
7. 执行器推进：`approved -> active`
8. 执行中所有关键步均经 gate 校验

### 5.3 执行与归档链路

9. `rule-loader` 按优先级装载规则
10. `gate-engine` 执行 blocker 校验
11. 通过后推进任务状态并持久化结果
12. 完成后归档：主索引保留摘要与指针
13. archived 默认不可用于开发/决策，仅可按例外票据只读排障

## 6) 错误与门禁模型（通过）

### 6.1 固定执行顺序

`Load Rules -> Validate Refs -> Validate Archive Access -> Validate Transition`

### 6.2 三类硬阻断（唯一）

1. `declared_reference_not_found`
2. `rule_override_conflict`
3. `archived_blocked`

### 6.3 统一返回结构

- `code`
- `level`（固定 `blocker`）
- `message`
- `context`（`task_id/module_id/rule_id/ref`）
- `next_action`

### 6.4 非阻断策略

- 仅允许 `warning`
- `warning` 不得绕过三类 blocker

## 7) 验证与验收策略（通过）

### 7.1 三层验证

1. 治理层单测：`rule-loader / gate-engine / advisor`
2. 用例层集成测：`orchestrator -> governance -> modules`
3. 回归场景测：生成、审批、执行、归档、例外读取全链路

### 7.2 双层验收

- 模块自动验收：标准化验收与 PASS/FAIL 结果
- 全局人工验收：Owner/Reviewer 最终签署

### 7.3 文本上下文安全校验（必须）

- 开发/决策输入必须来自文本工件
- archived 默认阻断，例外票据一次性只读并首读失效
- 会话内隐状态不得直接驱动状态推进

## 8) 推荐重构分期（用于后续实施）

### P0：治理骨架落地

- 建立 `internal/governance/*` 目录与契约定义
- 建立 `Workflow Plan` 对象与状态机
- 建立 `gate-engine` 三类 blocker 骨架

### P1：模块解耦迁移

- `memory/rules/task/workflow` 退化为实现层
- 清理模块内治理判定，统一迁移到 governance
- 建立 `orchestrator` 用例化流程入口

### P2：编排入口收敛

- `cmd/*` 一命令一用例
- `skills/agent-team` 降为纯编排壳
- 完成旧路径兼容与退场

## 9) 风险与缓解

### 风险 1：迁移期双轨逻辑并存导致行为不一致

- 缓解：引入“治理路径优先”开关，新增命令默认走新路径；旧路径只读兼容并限期退场

### 风险 2：团队习惯仍依赖会话语义而非文本工件

- 缓解：将 gate 前置到所有关键状态推进点；无文本引用则拒绝推进

### 风险 3：archived 例外放行被滥用

- 缓解：单任务一次性、只读、首读失效 + 审计记录

## 10) 本次脑暴结论

本次主线重构已经明确：

1. 架构采用“治理内核中心化 + 模块实现层 + 编排层分离”。
2. `Workflow Plan` 作为单对象生命周期承载任务编排。
3. 数据流、门禁模型、验收策略已完成并通过。
4. 以“文本上下文优先”作为安全基线，显式抑制 AI 内隐上下文驱动决策。

---

（本文件为重构脑暴通过稿，可作为后续实施规划输入。）
