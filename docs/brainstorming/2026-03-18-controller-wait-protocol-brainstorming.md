# Brainstorming: Controller 派发后等待协议修复

日期：2026-03-18
角色：general strategist

## 问题陈述

当前 Controller 侧关于任务派发后的调度协议不够清晰，容易让上层流程把 `agent-team worker status` 当成常规轮询工具。用户希望把 Controller 的默认行为收口为：

1. 派发任务后，不需要轮询查看员工状态
2. 默认等待 worker 自己通过 `reply-main` 主动回报
3. `worker status` 仅在异常排查、超时、或用户明确要求时使用

本次修复文档只讨论 Controller 侧 skills 的协议收口，不改 CLI 实现，不改 worker 侧规则，也不处理 role / alias / worker-id 映射问题。

## 目标

- 明确 Controller 在 `assign` 之后的默认行为是“等待回报”，不是“主动轮询”
- 让 `skills/agent-team/SKILL.md` 和 `skills/workflow/SKILL.md` 的调度叙述与这一协议一致
- 降低 controller 侧流程对 `worker status` 的依赖
- 保持现有 CLI 能力和 workflow 抽象层不变

## 约束与边界

- 不修改 `agent-team` CLI 源码
- 不修改 worker 侧 rules 或 worker workflow 文档
- 不要求 `workflow` 收口到显式 `worker-id`
- 不要求 `agent-team` skill 收口到“`worker assign` 只接受精确 worker-id”
- 不处理 role / alias / worker-id 的映射歧义
- 仅修复 Controller 侧文档协议和默认操作指引

## 候选方案

### 方案 A：最小收口，仅修 Controller 侧等待协议

修改：

- `skills/agent-team/SKILL.md`
- `skills/workflow/SKILL.md`

核心规则：

- `assign` 后默认进入 `waiting for worker reply`
- 默认观察信号是 worker 的 `reply-main`
- `worker status` 不再作为常规轮询步骤
- `worker status` 仅作为异常排查工具

优点：

- 变更面最小
- 不触碰实现
- 能直接解决“Controller 习惯性轮询状态”的协议误导

缺点：

- 其他 references 和 README 仍可能保留旧叙述
- 只解决这次目标，不解决更广的文档一致性问题

### 方案 B：中等收口，连同 controller references 一起修

在方案 A 基础上，再同步所有 Controller 相关 references。

优点：

- Controller 侧文档更一致

缺点：

- 范围扩大
- 容易把本次目标和其他历史漂移混在一起

### 方案 C：全面收口，统一所有相关文档

同步 skill、references、README、workflow references。

优点：

- 一致性最好

缺点：

- 超出本次目标
- 接近一次完整文档重构

## 推荐方案

推荐方案 A。

原因：

1. 这次需求非常具体，只针对 Controller 派发后等待协议
2. 用户明确表示不需要动 worker-id 相关边界，也不需要扩展到 worker 侧规则
3. 最小收口可以先把默认行为固定下来，避免继续把 `worker status` 当成常规调度动作

## 推荐设计

### 1. `skills/agent-team/SKILL.md` 的收口方向

把 Controller 视角下的“分配任务”描述为两个连续阶段：

1. `dispatch`
   - 执行 `agent-team worker assign ...`
2. `wait`
   - Controller 停止主动轮询
   - 等待 worker 通过 `agent-team reply-main "<message>"` 主动回报

需要强调：

- `worker status` 不是派发后的默认下一步
- Controller 不需要在派发后进入轮询循环
- 正常路径下，下一次动作应由 worker 回报触发

### 2. `skills/workflow/SKILL.md` 的收口方向

保留 workflow 的抽象层，不要求它显式写死 `worker-id`。

只调整节点行为语义：

- `assign_role_task`
  - 完成派发后，进入等待 worker 回报阶段
- `wait_for_completion`
  - 默认等待的是 worker 主动发送到 controller 的回报消息
  - 默认输入不是 `worker status` 的表格轮询结果

需要强调：

- `wait_for_completion` 的默认语义是“等待回报”
- 不是“循环执行 `worker status` 直到看到变化”

### 3. Controller 默认数据流

推荐将 Controller 数据流描述为：

1. Controller 运行 `agent-team worker assign ...`
2. Controller 记录当前节点进入 waiting 状态
3. Worker 在自身执行过程中，使用 `agent-team reply-main` 主动回报
4. Controller 根据回报决定：
   - 继续下一节点
   - 回复 worker
   - 向用户升级问题

默认不包括：

- 周期性 `worker status`
- 定时轮询
- “无消息即失败”的推断

### 4. 异常处理

`worker status` 保留，但只作为例外路径：

1. `assign` 命令失败
   - 可以立刻查看 `worker status` 或其它诊断信息
2. 已成功 `assign`，但长时间未收到回报
   - 可以把 `worker status` 当作超时后的排查手段
3. 用户明确要求查看当前 worker 状态
   - 可以主动查询并汇报

非异常路径下：

- 不主动查状态
- 不把状态表当成主驱动信号

## 涉及文件

本轮修复文档建议修改：

- `skills/agent-team/SKILL.md`
- `skills/workflow/SKILL.md`

本轮明确不动：

- `skills/agent-team/references/*`
- worker 侧规则文档
- CLI 源码
- README

## 风险与缓解

### 风险 1：局部修复后，其他文档仍保留旧习惯

影响：

- 仍可能有人从其他文档继承轮询思路

缓解：

- 本次先把 Controller 主入口 skill 收口
- 后续若仍反复出现，再做第二轮文档清理

### 风险 2：有人把“等待回报”理解成“永远不排查”

影响：

- 遇到超时或 assign 失败时可能不敢使用 `worker status`

缓解：

- 在 skill 中明确：`worker status` 是异常排查工具，不是禁用命令

### 风险 3：范围扩大到其他已知问题

影响：

- 文档修复目标失焦

缓解：

- 明确写入边界：本次不处理 worker-id、alias、role 映射问题

## 验证策略

文档修改完成后，按以下标准验证：

1. `skills/agent-team/SKILL.md`
   - 是否明确 `assign` 后默认等待 worker 主动回报
   - 是否不再暗示派发后需要持续查看状态
2. `skills/workflow/SKILL.md`
   - 是否把 `wait_for_completion` 默认语义改为“等待 worker 回报”
   - 是否把 `worker status` 降级为异常排查工具
3. 范围控制
   - 是否没有引入 CLI 行为变更要求
   - 是否没有把本次修复扩展到 worker 侧文档

## 实施建议

建议执行顺序：

1. 先修改 `skills/agent-team/SKILL.md`
2. 再修改 `skills/workflow/SKILL.md`
3. 逐段检查 Controller 调度叙述是否统一为“assign -> wait for reply”

## 开放问题

本轮无新增开放问题。

## 最终结论

本次最合适的修复方式，是只在 Controller 侧主技能文档中收口“派发后等待协议”：

- 正常路径：`assign -> wait for worker reply`
- 默认信号：`reply-main`
- `worker status`：仅用于异常排查，不用于常规轮询

这样可以在不改 CLI、不中断现有 workflow 抽象、也不触碰 worker 侧规则的前提下，先把 Controller 的默认调度习惯修正过来。
