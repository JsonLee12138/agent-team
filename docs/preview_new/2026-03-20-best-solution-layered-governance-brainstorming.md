# AI 团队流程最佳方案（强治理分层 + 结构化任务包 + 模块自动验收 + 全局人工验收）

- 日期：2026-03-20
- 角色：general strategist
- 类型：流程设计脑暴（最佳方案定版稿）
- 来源：基于主文档《2026-03-20-context-accuracy-document-memory-brainstorming.md》与外部参考项目调研综合收敛
- 主题：以规则分层、索引优先、声明引用校验和双层验收闭环，降低多智能体执行中的思维偏差与上下文漂移

---

## 1) 方案结论

本轮最佳方案定义为：

> **强治理文档主线 + 分层规则加载 + 结构化任务包 + 模块自动验收 + 全局人工验收**

一句话说明：

> **以 Index Registry 为单一事实入口，以 Public Rule > Module Rule > Task Rule 为规则装载顺序，以 Task Packet 作为唯一执行载体，以 declared-reference existence 作为唯一硬引用校验，以模块标准化验收 + 全局人工验收形成闭环。**

---

## 2) 为什么这是最佳方案

### 当前最需要解决的问题
当前风险不在于模型能力不足，而在于：
1. 执行者在多轮交接中逐步偏离原始规则与目标。
2. 上下文依赖会话记忆，难以审计与回放。
3. 模块化推进时，低层规则可能反向污染公共基线。
4. 自动化验收可以覆盖局部标准，却无法替代全局业务判断。

### 为什么不直接照搬单个参考项目
- **gstack** 强在工件化流程，但引用与审批约束不够强。
- **hermes-agent** 强在分层上下文、压缩与回滚，但缺少显式规则宪法。
- **agency-agents** 强在 state machine 与 handoff 模板，但 human signoff 不够强。
- **everything-claude-code** 强在 hooks、索引与 project-scoped memory，但更偏平台层。
- **superpowers** 强在执行纪律与 fresh subagent，但索引和持久治理偏轻。

因此最佳方案不是复刻单一项目，而是吸收其互补长处，收敛为适合当前阶段的轻量强治理模型。

---

## 3) 核心设计原则

### 3.1 规则原则
- **Public-First**：公共规则只读，优先级最高。
- **Module-Inherit-Only**：模块规则只能继承与补充，不能覆盖公共规则。
- **Task-Consume-Only**：任务只消费已声明规则，不重新解释规则。

### 3.2 上下文原则
- **Index-First**：所有执行对象先入索引，再进入流转。
- **No-Doc-No-Work**：没有 DocID / TaskID / RuleRef 的事项不得启动。
- **Fresh-Context-Execution**：执行者只接收当前任务包及其声明引用，不灌入无关历史上下文。

### 3.3 校验原则
- **Declared-Ref-Strict**：仅当“已声明引用缺失实体”时才硬阻断。
- **Rule-Conflict-Blocking**：任何低层规则覆盖高层规则，直接阻断。
- **Evidence-over-Claims**：未绑定证据的完成结论默认无效。

### 3.4 验收原则
- **Module Acceptance Standardized**：模块内验收尽可能标准化。
- **Global Acceptance Human-Owned**：全局业务闭环由人工负责最终签署。

---

## 4) 目标架构

### 4.1 分层治理架构
系统由五个核心层组成：

1. **Index Layer**
   - 维护所有 workflow / module / task / rule / evidence 的索引与状态。

2. **Rule Layer**
   - 管理 Public Rule、Module Rule、Task Rule 的装载顺序与冲突检测。

3. **Execution Layer**
   - 基于结构化 Task Packet 派发任务，执行者只读取声明内容。

4. **Gate Layer**
   - 对引用有效性、规则冲突、模块验收进行标准化判断。

5. **Review Layer**
   - 模块负责人与全局 Reviewer 分层决策，保留全局人工签署。

### 4.2 角色分工
- **Worker**：执行任务包，产出结果与证据。
- **Module Lead**：审查模块规则、任务边界与模块验收结果。
- **Global Reviewer**：基于全局业务闭环作最终人工判断。
- **Rule Owner**：维护公共规则与变更审批流。

---

## 5) 四个必须存在的治理对象

### 5.1 Index Registry
单一事实入口，负责记录所有对象及其关系。

#### 最小字段建议
- `ID`
- `Type`（workflow / module / task / rule / evidence）
- `ParentRef`
- `PublicRuleID`
- `ModuleRuleID`
- `AcceptanceRef`
- `EvidenceRefs`
- `Owner`
- `Status`

#### 作用
- 所有对象必须先注册，再流转。
- 所有引用必须可回查。
- 所有状态变更必须可追踪。

---

### 5.2 Rule Loader
显式规则装载器，负责按固定顺序解析规则。

#### 固定加载顺序
1. `Public Rule`
2. `Module Rule`
3. `Task Rule`

#### 约束
- Public Rule 为只读基线。
- Module Rule 必须声明继承的 `PublicRuleID`。
- Task Rule 仅允许补充执行信息，不允许覆盖上层规则。

#### 阻断场景
- 模块规则覆盖公共规则：`rule_override_conflict`
- 任务规则覆盖模块或公共规则：`rule_override_conflict`

---

### 5.3 Task Packet
唯一合法的任务执行载体。

#### 最小字段建议
- `TaskID`
- `Goal`
- `Scope`
- `OutOfScope`
- `PublicRuleID`
- `ModuleRuleID`
- `AcceptanceRef`
- `Dependencies`
- `EvidenceRequired`
- `Owner`
- `CurrentStatus`

#### 设计意图
- 任务不再依赖长篇自然语言交接。
- 执行者只接收“当前必须知道的最小完整集”。
- 任务目标、边界、规则与证据要求同时被固定下来。

---

### 5.4 Gate Engine
统一门禁引擎，只负责有限而稳定的硬阻断。

#### 硬阻断 A：`declared_reference_not_found`
已声明引用，但实体不存在。

适用例：
- 写了 `PublicRuleID` 但规则不存在。
- 写了 `ModuleRuleID` 但模块规则不存在。
- 写了 `EvidenceRef` 但对应证据不存在。
- 写了 `AcceptanceRef` 但验收标准不存在。

#### 硬阻断 B：`rule_override_conflict`
低层规则试图覆盖高层规则。

适用例：
- Module Rule 改写 Public Rule。
- Task Rule 绕过 Module Rule / Public Rule。

#### 非阻断事项
- 某部分说明缺少
- 未声明字段为空
- 说明不完整但未构成引用缺失

以上默认视为“不适用”或“待补充”，而非硬阻断。

---

## 6) 标准执行流程

### Phase 0：建立 Root / Workflow Doc
- 创建全局入口文档。
- 绑定 active `PublicRuleID`。
- 定义全局 Acceptance 标准。
- 注册 Workflow Owner 与全局阶段状态。

### Phase 1：模块拆分
- 为每个模块创建 `Module Rule Doc`。
- 绑定对应 `PublicRuleID`。
- 定义模块验收标准。
- 分配 Module Owner。

### Phase 2：任务下发
- 为每个任务生成 `Task Packet`。
- 强制声明：`PublicRuleID + ModuleRuleID + AcceptanceRef`。
- 缺任一引用，不允许进入执行阶段。

### Phase 3：Worker 执行
- 基于 fresh context 执行。
- 只加载当前 Task Packet 与其声明引用。
- 增量绑定 `EvidenceRef`。
- 不依赖隐式会话记忆。

### Phase 4：模块自动验收
统一做五类判断：
1. 引用是否可解析。
2. 是否遵守 Public Rule。
3. 是否遵守 Module Rule。
4. 声明证据是否存在。
5. Acceptance 是否满足。

输出标准状态：
- `PASS`
- `FAIL`
- `RETRY`
- `ESCALATE`

### Phase 5：集成冒烟
- 仅验证跨模块关键路径。
- 不在这一层重做全量业务验收。

### Phase 6：全局人工验收
人工 Reviewer 只关注四项：
1. 业务闭环是否成立。
2. 跨模块关键路径是否成立。
3. 风险是否可接受。
4. 是否允许发布 / 合并 / 进入下一阶段。

---

## 7) 方案吸收的外部参考优势

### 来自 gstack
- 工件化流程链
- 共享规则块统一注入
- review readiness / dashboard 思路

### 来自 hermes-agent
- 分层上下文装配
- 冻结式记忆快照
- context compression / rollback / lineage
- 子代理隔离

### 来自 agency-agents
- Workflow Registry 四视图
- handoff packet 模板
- QA PASS/FAIL / retry / escalation
- evidence-over-claims

### 来自 everything-claude-code
- hooks / profile-based gate
- project-scoped memory
- codemap / command-agent map / search-first
- 索引优先的能力发现方式

### 来自 superpowers
- fresh subagent per task
- plan-first
- dual review
- verification-before-completion

---

## 8) 为什么这套方案适合当前阶段

### 8.1 准确性优先，而不是自动化优先
当前阶段最大目标是降低偏差，不是追求全自动流转。因此：
- 模块内标准化，便于自动化
- 全局保留人工判断，避免业务误判

### 8.2 强治理但不过重
该方案先落最小治理闭环：
- Index Registry
- Rule Loader
- Task Packet
- Gate Engine

而不是一开始建设重平台、全量 memory、复杂 orchestration runtime。

### 8.3 对执行者友好
执行者只需要：
- 接收任务包
- 读取声明引用
- 产出证据
- 等待 gate 结果

无需自己重建大上下文，也不需要自己解释规则层级。

---

## 9) 风险与缓解

### 风险 1：索引维护成本增加
- **缓解**：只维护最小字段集；避免把索引变成长文档。

### 风险 2：模块负责人变成瓶颈
- **缓解**：负责人只审边界、冲突、证据与验收状态，不审执行细节。

### 风险 3：任务包模板变成形式主义
- **缓解**：控制字段数量，只保留最小必需集。

### 风险 4：模块验收覆盖不了业务闭环
- **缓解**：明确模块验收与全局人工验收职责边界，不混用。

### 风险 5：公共规则变更过慢
- **缓解**：单独维护 Public Rule 变更队列，不让模块执行流阻塞在规则设计上。

---

## 10) 推荐落地顺序

### 第一步：最小治理闭环
只落四项：
1. `Index Registry`
2. `Public / Module / Task` 三层规则模型
3. `Task Packet` 模板
4. `declared_reference_not_found` + `rule_override_conflict` 两类门禁

### 第二步：执行纪律增强
补充：
- fresh worker / subagent
- 模块 PASS / FAIL / RETRY / ESCALATE
- owner signoff
- 集成冒烟
- 全局人工验收清单

### 第三步：增强治理能力
最后再补：
- compact / handoff summary
- rollback / checkpoint
- deviation ledger
- dashboard
- project-scoped memory

---

## 11) 最终定版结论

1. **最佳方案不是单一项目复刻，而是组合式收敛方案。**
2. **治理核心不是增强 agent 自由度，而是限制其跑偏空间。**
3. **规则必须显式分层，且优先级固定为 Public > Module > Task。**
4. **执行必须结构化，Task Packet 应成为唯一合法任务载体。**
5. **硬阻断只保留两类：声明引用缺失、规则越权覆盖。**
6. **验收必须分层：模块尽量自动化，全局必须人工签署。**
7. **当前阶段应优先建设最小治理闭环，而非重型平台。**

---

## 12) 与主文档关系

本文件是对主文档《2026-03-20-context-accuracy-document-memory-brainstorming.md》的“最佳方案定版补充”。

主文档回答“为什么要这样治理”；
本文件回答“最佳收敛方案具体长什么样”。
