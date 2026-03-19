# Strategic Compact Brainstorming

## Role

general strategist

## Problem Statement and Goals

目标是基于 `everything-claude-code` 中的 strategic-compact 思路，优化当前仓库的上下文管理机制，但重点不是做一个通用的 provider-aware 压缩器，而是做一个适配 `agent-team` 工作流的、**token-first** 的上下文管理方案。

核心目标：
- 以减少 **main 会话** 的长期 token 消耗为第一优先级
- 保持多轮派工、审查、合并流程中的上下文清晰度
- 避免 worker 成为长期上下文负担
- 让 compact 策略和现有 `task` / `workflow` 能稳定协同

非目标：
- 不把 provider 差异放进 skill / rules
- 不默认生成 handoff summary
- 不把 worker 设计成长期保留并频繁 compact 的主体

## Constraints and Assumptions

### Constraints
- provider 相关差异已在脚本层处理，skill 和 rules 不需要再区分 Claude / Codex / 其他 provider
- 新能力需要和现有 `task`、`workflow` 配合，而不是另起一套流程
- 当前仓库已经有 `agent-team compact` 作为执行原语
- rules 应保持精简，适合承载触发条件和强制约束，不适合堆积复杂策略逻辑

### Assumptions
- worker 默认应视为一次性执行单元：完成工作后由 main 审查、合并并清理
- compact 的主要价值点在 **main**，因为 main 是长期保留的决策/调度上下文
- worker 只有在长任务、阻塞等待、多轮修改等例外场景才值得进入 compact 策略
- 目标是降低**总 token 成本**，而不是让每次 compact 调用本身最“聪明”

## Candidate Approaches

### Option A — Rule-first + 单一 compact skill

**做法**
- 新建 `strategic-compact` skill
- `.agents/rules/context-management.md` 只保留 trigger 和“命中后必须调用该 skill”
- `task` / `workflow` 在触发点调用该 skill

**优点**
- 结构简单
- rule / skill 边界清晰
- 对现有系统侵入较低

**缺点**
- skill 仍然偏薄
- 对 `task` / `workflow` 的协同表达不够强
- 容易退化成“换个入口调用 compact”

---

### Option B — Context Orchestrator（推荐）

**做法**
- 新建一个统一入口 skill：`agent-team:strategic-compact`
- 它负责：触发标准化、状态收集、策略选择、compact 输入整理、触发原生 compact
- `task` / `workflow` 只负责在命中 trigger 时调用它，不自己决定 compact 内容
- 设计上以 **main-first / token-first** 为优先级，worker 默认不走 compact 主路径

**优点**
- 最适合当前 `agent-team` 的主控-执行工作流
- compact 策略集中，避免在 rules / task / workflow 中四处分叉
- 更容易吸收 `strategic-compact` 的优点，又不会把 provider 逻辑混进来

**缺点**
- 第一版需要先定义清晰的输入/输出契约
- skill 需要谨慎控制读取成本，避免“为了省 token 反而先多花一轮 token”

---

### Option C — 先抽共享 context contract，再由多个 skill 复用

**做法**
- 先抽象一套统一 contract（goal / phase / constraints / next / risks ...）
- 再让 `task`、`workflow`、compact 相关 skill 都基于这套 contract 协作

**优点**
- 长期扩展性最好
- 对 resume / handoff / debugging 等场景也有帮助

**缺点**
- 当前阶段明显过度设计
- 复杂度高于当前目标

## Recommended Design

选择 **Option B — Context Orchestrator**。

### Why
- 用户明确希望新能力是“策略增强”，而不是简单包装 `/compact`
- 新能力需要与 `task` / `workflow` 协同，而不是孤立存在
- token 优化目标要求把 compact 主要放在 **main**，而不是把所有 worker 都做成长生命周期压缩对象

## Architecture

### Role split
- **Rules (`.agents/rules/context-management.md`)**
  - 只定义触发条件与强制约束
  - 命中 trigger 时必须进入 `strategic-compact`
- **Strategic skill (`agent-team:strategic-compact`)**
  - 统一负责：读取状态 -> 选择策略 -> 生成 compact 输入 -> 执行 compact
- **Task / Workflow**
  - 仅负责在阶段切换或特定场景中判断 trigger
  - 命中后调用 `strategic-compact`
  - 不自行实现 compact 内容逻辑

### Layering
- **Policy layer**: rules 决定何时必须 compact
- **Orchestration layer**: skill 决定怎么 compact
- **Execution layer**: 原生 compact / `agent-team compact` 等执行原语

### Main-first / Token-first principle
- worker 默认短生命周期：完成 -> 审查/合并 -> 清理
- main 才是长期上下文管理对象
- worker compact 是例外，不是默认路径

## Components

### 1. Trigger Contract
把所有入口统一为固定 trigger 类型：
- `manual`
- `phase-transition`
- `context-pressure`
- `pre-large-read`
- `resume-after-pause`

### 2. State Collector
深读取当前仍影响后续行动的状态，但输出为稳定 schema，而不是原始材料堆叠：
- current_goal
- current_phase
- task_status
- workflow_status
- constraints
- recent_decisions
- completed_items
- pending_items
- next_step
- risks_or_blockers
- verification_state（如果相关）

### 3. Compact Strategist
根据 trigger、阶段、复杂度与恢复成本决定策略档位：
- **Light**：快速阶段切换
- **Standard**：常规主控上下文管理
- **Deep**：仅在 main 上下文明显失控或高恢复成本时使用

### 4. Executor（第一版可内嵌）
负责把 compact packet 转成 compact 前输入，并调用原生 compact。

### 5. Recovery Contract（第一版可内嵌）
最小恢复锚点：
- Goal
- Phase
- Constraints
- Done
- Next
- Risks/Blockers（如有）

## Data Flow

### Standard flow
1. `task` / `workflow` 或用户触发 `strategic-compact`
2. trigger 被标准化
3. State Collector 读取最小必要状态
4. Strategist 选择 `light` / `standard` / `deep`
5. 生成 compact packet
6. 调用原生 compact
7. 返回最小结果给调用方

### Main-first flow
- **worker 标准路径**：创建 -> 完成任务 -> main 审查/合并 -> 删除
- **main 标准路径**：长期保留目标、阶段、决策、已完成结果与下一步派工判断
- 当 main 命中 trigger 时，再进入 strategic-compact

### Worker exception flow
只有下列情况，worker 才进入 compact 分支：
- 长任务不会立即结束
- 阻塞后仍需保留实现现场
- 多轮修改需要保留当前工作上下文

## Error Handling

### 1. Trigger 误判
- 命中 trigger 后不直接进入重整理
- 先做低成本再判断，默认走 light/standard

### 2. 状态读取过重
- Collector 使用 budgeted read
- 先读最小状态：goal / phase / task/workflow state / next step
- 只有缺恢复锚点时才升级读取
- 超预算时退回 standard compact

### 3. Main 信息不完整
- 允许标记 `unknown` / `needs refresh`
- 不为了补全信息而做大范围回溯

### 4. Worker 被误纳入主路径
- skill 默认假设 worker 是 disposable
- 若 worker 不满足例外条件，则建议 finish/reply/merge/delete，而不是 compact

### 5. Compact 后恢复信息不足
- 强制最小恢复 contract：Goal / Phase / Constraints / Done / Next / Risks
- 只有 contract 无法形成时，才考虑 handoff summary 兜底

### 6. Task / Workflow 策略漂移
- rules 明确：命中 trigger 时只能走 `strategic-compact`
- `task` / `workflow` 不得自行决定 compact 内容

## Validation / Test Strategy

### 1. Structure validation
检查：
- rules 是否只保留 trigger 和强制入口
- compact 策略是否只存在于 `strategic-compact`
- `task` / `workflow` 是否只负责判断 trigger 并调用 skill
- worker 默认路径是否仍是完成后清理

### 2. Behavior validation
检查：
- main 在阶段切换时会调用 `strategic-compact`
- main 在读大输出前会调用 `strategic-compact`
- worker 普通完成路径不会进入 compact
- worker 只有例外场景才进入 compact

### 3. Cost validation
对比：
- **基线流程**：main 长期累积上下文，不做战略 compact
- **优化流程**：main 命中 trigger 时走 strategic-compact，worker 完成即清理

观察：
- 多轮任务后 main 的可持续性
- 是否减少“重新解释背景”的 token 消耗
- 是否降低长生命周期 worker 的需要

### 4. Recovery validation
compact 后验证是否仍能稳定恢复：
- 当前总目标
- 当前阶段
- 已完成结果
- 下一步派工/审查动作
- 关键约束与风险

## Risks and Mitigations

### Risk 1: strategic-compact 自身成本过高
**Mitigation:** 默认走 light / standard，collector 做预算限制，deep 只作为例外。

### Risk 2: 规则与 skill 职责混淆
**Mitigation:** rules 只定义 trigger 和强制入口，所有策略细节移入 skill references。

### Risk 3: worker 也被做成长生命周期上下文对象
**Mitigation:** 把“worker 默认完成即清理”写成核心假设，并在 skill 中默认拒绝普通 worker compact。

### Risk 4: compact 后恢复锚点不稳定
**Mitigation:** 固定最小恢复 contract，必要时才使用 handoff summary 兜底。

## Open Questions

当前讨论里已经明确的关键决策：
- 不在 skill / rules 中考虑 provider 差异
- 选择 Context Orchestrator 方案
- `task` / `workflow` 仅在命中 trigger 时调用 skill
- skill 同时具备阶段感知与 task/workflow 状态感知
- compact 主要服务 main，worker 默认短生命周期

后续实现前仍可进一步细化的问题：
- `strategic-compact` 是否需要显式区分 `light` / `standard` / `deep` 的公开模式，还是仅内部自动决策
- `task` / `workflow` 具体有哪些阶段切换点应接入 trigger 判断
- 第一版是否直接落在 `agent-team` skill 包内，还是单独拆成新 skill 包

## Final Recommendation

把 strategic compact 设计成 **main-first、token-first 的上下文编排 skill**，并通过 rules 强制它成为唯一的 compact 策略入口。

这样可以：
- 把 token 优化重点放在真正长期存在的 main 会话上
- 避免 worker 变成长期上下文包袱
- 让 `task` / `workflow` 接入方式简单稳定
- 在不引入 provider 复杂度的前提下，吸收 `strategic-compact` 的策略优势
