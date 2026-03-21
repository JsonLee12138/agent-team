# P2：roadmap / milestone / phase 规划层脑暴

日期：2026-03-21
角色：general strategist

## 问题陈述

当前 `agent-team` 已明确以 `task-first + single-task worker lifecycle` 作为执行核心。

P2 需要讨论的是：是否以及如何引入 `roadmap / milestone / phase`，让系统在不破坏当前执行模型的前提下，支持更清晰的人类可读规划、阶段可视化、并行判断与历史沉淀。

本次脑暴的重点不是把 `roadmap / milestone / phase` 直接做成新的执行层，而是明确它们是否应该作为 **规划与展示层** 存在，以及它们如何与现有 `task` 和 `verification` 协作。

## 目标

- 明确 `roadmap / milestone / phase` 的边界与职责
- 保持 `task` 作为唯一执行单元
- 支持人和 AI 快速看到当前做到哪里、后续做什么、哪些可并行
- 让高层规划对象可独立存在，也可引用下层对象
- 支持归档与废弃
- 为后续 brainstorming skill 与规划层工具演进提供结构基础

## 约束与前提

- 当前优先级仍然是文档同步与 `task/verification.md` 闭环
- `task` 已是当前最小执行单元，本次不重设计 task 层
- `verification.md` 仍然是 task 级正式验收工件
- `phase` 可以存在验收标准，但不应默认触发主动验收
- 正式验收文件如需生成，应通过 `qa-expert` skill
- 规划层应尽量减少默认成本，避免因为结构存在就触发额外 AI 行为
- 规划层引用缺失可以报错，但不应因内容不完整而默认阻断整体使用

## 候选方案与取舍

### 方案 A：纯文档规划层

仅以文档方式描述 `roadmap / milestone / phase`，不引入独立结构约束。

**优点**
- 成本最低
- 不影响当前 task-first 模型
- 适合早期概念澄清

**缺点**
- 结构化程度不足
- 可扩展性弱
- 不利于后续精准脑暴、索引与历史管理

---

### 方案 B：轻工件规划层（推荐）

为 `roadmap / milestone / phase` 提供独立工件，但定位为规划与展示层，不进入执行状态机。

**优点**
- 人类与 AI 都更容易理解全局结构
- 保持 task 作为唯一执行层
- 便于独立存在、按需引用、历史沉淀与后续演进

**缺点**
- 比纯文档多出一定结构维护成本
- 需要明确好目录、引用与生命周期语义

---

### 方案 C：完整规划执行层

让 `roadmap / milestone / phase` 全部进入 CLI 状态流转甚至执行模型。

**优点**
- 结构最完整
- 长期有较强治理潜力

**缺点**
- 当前阶段过重
- 会弱化 task-first 的清晰性
- 容易与现有 phase 语义冲突
- 默认成本高，不符合当前诉求

## 推荐结论

推荐 **方案 B：轻工件规划层**。

核心结论：

- `roadmap / milestone / phase` 是 **规划与展示层**
- `task` 是 **唯一执行层**
- `verification` 是 **task 验收层**
- 规划层主要面向人类阅读、全局理解、并行判断与历史沉淀
- AI 可以读取这些对象，但不应因其存在而自动进入额外执行或验收行为

## 设计一：架构边界

### 分层职责

- **roadmap**：回答整体路线、Now / Next / Later、优先级与方向
- **milestone**：回答阶段性交付目标与用户可获得的能力
- **phase**：回答 milestone 内部先后顺序、可并行性与阶段边界
- **task**：回答谁执行哪个最小工作项，是唯一执行单元
- **verification**：回答 task 如何验收、验收结果如何

### 硬约束

- worker 只绑定 task
- `roadmap / milestone / phase` 不得升级为执行单元
- `phase` 即使独立成文件，也仍然只是规划节点
- 上层对象可以有展示状态，但不能替代 task 状态机

## 设计二：物理结构与逻辑结构解耦

核心原则：**逻辑层级不等于物理目录层级**。

虽然 `roadmap -> milestone -> phase -> task` 是清晰的语义层级，但文件系统不应强制做成父子嵌套，否则会错误表达“必须隶属”。

推荐采用：**物理平铺，逻辑引用**。

### 推荐目录结构

```text
.agent-team/
  planning/
    roadmaps/
      roadmap-001/
        roadmap.md
    milestones/
      milestone-001/
        milestone.md
        acceptance.md      # 可选
    phases/
      phase-001/
        phase.md
        acceptance.md      # 可选

  archive/
    roadmaps/
      roadmap-001/
        roadmap.md
    milestones/
      milestone-001/
        milestone.md
        acceptance.md      # 可选
    phases/
      phase-001/
        phase.md
        acceptance.md      # 可选

  deprecated/
    roadmaps/
      roadmap-002/
        roadmap.md
    milestones/
      milestone-002/
        milestone.md
    phases/
      phase-002/
        phase.md
```

### 目录语义

- `.agent-team/planning/`：当前有效、可继续参考或推进的对象
- `.agent-team/archive/`：已完成、作为历史沉淀保留的对象
- `.agent-team/deprecated/`：取消、废弃、被替代但仍需保留说明的对象

这种结构允许：

- roadmap 独立存在
- milestone 独立存在
- phase 独立存在
- task 继续独立存在，不强制依赖更高层级

## 设计三：验收模型

### task 层

`verification.md` 仍然是 task 级正式验收工件。

- task 负责执行
- verification 负责证据与结果
- 本次不重新设计 task 结构，只承认其作为执行核心的现状

### phase 层

phase 可以存在 `Acceptance Criteria`，但：

- 不默认触发 AI 主动验收
- 不因为存在 criteria 就自动生成额外验收文件
- 默认不成为强制 gate

如果需要正式留痕，可按需生成 `acceptance.md`。

### acceptance.md 的定位

- 是阶段或里程碑级的**决策留痕**
- 不是对 task verification 的重复抄写
- 只在明确需要时生成
- 应通过 `qa-expert` skill 生成或协助生成

### 推荐原则

- 可以有验收标准
- 不默认触发主动验收
- AI 和人按成本收益自行判断是否执行验收
- 正式验收文件由 `qa-expert` 负责

## 设计四：最小字段设计

### roadmap.md

推荐字段：

- `ID`
- `Title`
- `Status`
- `Owner`（可选）
- `Horizon`（Now / Next / Later）
- `Goal`
- `Success Signals`
- `Related Milestones`
- `Related Phases`（可选）
- `Related Tasks`（可选）
- `Risks / Notes`
- `Lifecycle`（active / archived / deprecated）
- `Archived At` / `Deprecated At`（可选）
- `Deprecated Reason`（可选）
- `Replaced By`（可选）

### milestone.md

推荐字段：

- `ID`
- `Title`
- `Status`
- `Goal`
- `User Value`
- `Done Criteria`
- `Related Roadmaps`
- `Indexed Phases`
- `Indexed Tasks`
- `Dependencies`
- `Risks / Notes`
- `Lifecycle`
- `Archived At` / `Deprecated At`
- `Deprecated Reason`
- `Replaced By`

### phase.md

推荐字段：

- `ID`
- `Title`
- `Status`
- `Goal`
- `Scope`
- `Acceptance Criteria`
- `Related Milestones`
- `Related Roadmaps`（可选）
- `Indexed Tasks`
- `Parallelizable With`
- `Dependencies`
- `Risks / Notes`
- `Lifecycle`
- `Archived At` / `Deprecated At`
- `Deprecated Reason`
- `Replaced By`

### acceptance.md（按需）

推荐字段：

- `Target Type`（phase / milestone）
- `Target ID`
- `Summary`
- `Decision`（accept / accept-with-risk / hold）
- `Basis`
  - task verification refs
  - manual review refs
- `Open Issues`
- `Accepted By`
- `Accepted At`

## 设计五：引用与索引模型

### 基本原则

- 使用稳定 ID 引用，不使用标题作为主引用键
- 高层维护低层索引
- 低层可选反向引用，但不强制
- 物理目录不表达强父子关系
- 逻辑关系通过字段显式表达

### 推荐引用方式

- roadmap 维护 `Related Milestones` / `Indexed Phases` / `Indexed Tasks`
- milestone 维护 `Related Roadmaps` / `Indexed Phases` / `Indexed Tasks`
- phase 维护 `Related Milestones` / `Related Roadmaps` / `Indexed Tasks`

### acceptance 的引用原则

`acceptance.md` 只引用事实依据，不重复创造事实：

- task verification refs
- 人工评审依据
- 决策人与决策时间

## 设计六：生命周期与流转

推荐采用轻量生命周期：

1. 新建对象到 `.agent-team/planning/...`
2. 通过 ID 与字段引用其他对象
3. 如需正式留痕，再按需生成 `acceptance.md`
4. 完成后移动到 `.agent-team/archive/...`
5. 废弃或被替代后移动到 `.agent-team/deprecated/...`

这里的重点是：

- 用目录表达生命周期
- 不引入复杂状态机
- 不自动触发执行动作

## 异常处理与校验原则

推荐保持轻校验策略：

- 当对象声明了引用，但目标不存在时，报错或告警
- 内容不完整、字段不齐全、未生成 acceptance，不默认阻断
- 规划层更偏向结构校验，而非流程强制

这与当前偏好的“声明引用缺失才阻断”保持一致。

## Brainstorming skill 的后续演进建议

当前 brainstorming skill 仅支持较通用的设计收敛与保存流程。后续若围绕规划层演进，建议增强：

### 1. 支持选择脑暴目标对象

在保存位置之前，先让用户选择本次脑暴面向：

- roadmap
- milestone
- phase
- task
- generic topic

### 2. 支持基于已有对象精准脑暴

允许用户指定：

- 基于某个 roadmap 内容脑暴
- 基于某个 milestone 内容脑暴
- 基于某个 phase 内容脑暴
- 基于多个 task / phase refs 聚合脑暴

### 3. 支持更灵活的保存位置

除默认 `docs/brainstorming/` 外，可考虑支持：

- 保存到目标对象目录
- 保存到通用 brainstorming 目录
- 仅输出、不保存

## 风险与缓解

### 风险 1：规划层被误用为执行层

**缓解：** 文档和 skill 中始终强调：task 是唯一执行单元。

### 风险 2：phase 验收变成默认成本

**缓解：** 只保留 Acceptance Criteria，不默认触发验收；正式留痕时再生成 `acceptance.md`。

### 风险 3：目录与引用同时存在后维护成本上升

**缓解：** 首版只要求稳定 ID 与基础引用，索引增强后置。

### 风险 4：物理平铺后关系不直观

**缓解：** 通过高层索引字段和后续 index 文件增强可读性。

## 最终建议

P2 最合适的方向不是把 `roadmap / milestone / phase` 直接做成新的执行状态机，而是把它们设计成：

- **人类优先、AI可读的规划与展示层**
- **物理平铺、逻辑引用的独立工件体系**
- **支持 planning / archive / deprecated 生命周期迁移**
- **支持按需 acceptance 留痕，但不默认主动触发验收**
- **始终以 task 作为唯一执行层**

这个方向既保留了你希望的结构化规划能力，也不会破坏当前已经较清晰的 task-first 模型。

## 后续建议

在 P2 脑暴结论基础上，后续可以拆成两个独立主题继续推进：

1. **规划层工件与索引设计**
   - 是否需要 index 文件
   - 如何校验引用完整性
   - 是否需要最小 CLI 支持

2. **brainstorming skill 演进**
   - 保存位置选择增强
   - 支持按 roadmap / milestone / phase 精准脑暴
   - 支持输出到对象目录或通用目录
