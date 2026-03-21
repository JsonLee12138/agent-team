# CLI Skills Split Brainstorming

## Role

general strategist

## Problem Statement and Goals

当前需要基于现有 `agent-team` CLI 重新设计一批**独立 skills**。这些 skills 不再继续堆叠在 `skills/agent-team/` 之下，而是拆成独立技能目录，由 AI 在用户自然语言场景下触发，并在 skill 内部绑定明确的 CLI 命令。

目标：

- 按**用户场景**组织 skill，而不是机械按命令树切碎。
- 保留与当前 `agent-team` 类似的使用体验：用户说自然语言，skill 判断并调用对应 CLI。
- 明确哪些 skill 给 controller、worker、human 使用，哪些可共享。
- 单独保留一个类似 `strategic-compact` 的**上下文策略 skill**，但语义改为“清理会话上下文并从文件重新读取”，而不是压缩上下文。
- 上下文恢复必须采用**索引优先**模型：先读索引，再按需读正文。

## Constraints and Assumptions

### Constraints

- 这些能力应作为**独立 skills** 存在，不继续放进 `skills/agent-team/` 内。
- skill 触发方式以**用户场景推断**为主，但 skill 内部仍要明确绑定 CLI 命令。
- 不要求所有 skill 同时面向 controller / worker / human 三者；应按职责分配。
- 当前项目**全部走文件上下文**，不再以会话压缩作为核心策略。
- `/compact` 语义不再适合作为主方案。
- 上下文 skill 清理的是**会话上下文**，不是文件内容。
- 上下文恢复时，必须**先读取索引文件**，再根据索引读取实际文件内容。

### Assumptions

- 现有 CLI 主面包括：`task`、`workflow plan`、`worker`、`reply`、`reply-main`、`init/setup/migrate`、`rules`、`skill`、`role-repo`、`catalog`、`role list`。
- skill 设计会延续当前仓库的规则入口模式，至少需要尊重 `.agents/rules/index.md`。
- worker 场景下，`worker.yaml` 是恢复当前任务锚点的稳定入口。

## Candidate Approaches with Trade-offs

### Approach A — 场景优先拆分，命令绑定在 skill 内部（Recommended)

做法：
- skill 按用户意图与工作流场景拆分。
- 每个 skill 在内部绑定一个或多个 CLI 命令。
- 对 controller、worker、human 做显式受众分层。

优点：
- 最接近当前 `agent-team` 的交互体验。
- 用户不必先理解完整 CLI 树。
- 易于把多条 CLI 命令组合为一个稳定场景入口。

缺点：
- skill 说明文档必须写清楚边界，否则容易职责重叠。
- 某些场景 skill 内部会绑定多个命令，文档结构需更严格。

### Approach B — 严格按一级命令拆 skill

做法：
- `task`、`workflow`、`worker`、`role-repo`、`catalog` 等各自一个 skill。

优点：
- 结构清晰。
- 与 CLI 命令树天然对齐。

缺点：
- 不够贴合用户场景。
- 某些 skill 会过大，某些又过小。
- 对 worker / human / controller 的差异表达不足。

### Approach C — 全部细拆到子命令级 skill

做法：
- 如 `task-create`、`task-assign`、`workflow-plan-generate` 等全部拆开。

优点：
- 粒度最细，边界最硬。

缺点：
- 数量爆炸，维护成本高。
- 用户体验差，不像场景驱动系统。
- 很多 skill 只包一条命令，价值不足。

## Recommended Design

采用 **Approach A：场景优先拆分，命令绑定在 skill 内部**。

设计原则：

1. **按用户场景触发**：用户用自然语言表达意图，skill 负责归类并路由到 CLI。
2. **按使用者分层**：区分 controller-first、worker-first、human-first、shared、strategy 五类。
3. **命令显式绑定**：每个 skill 明确列出自己负责调用的 CLI。
4. **保留上下文策略 skill**：新增一个类似 `strategic-compact` 的 skill，但改为清理会话上下文而非压缩。
5. **索引优先恢复上下文**：上下文恢复必须先读索引入口，再按索引结果读正文。

## Final Skill Inventory

### Controller-first

1. `task-orchestrator`
2. `workflow-orchestrator`
3. `worker-dispatch`

### Worker-first

4. `worker-recovery`
5. `worker-reply-main`

### Human-first

6. `project-bootstrap`
7. `rules-maintenance`
8. `skill-maintenance`
9. `role-repo-manager`
10. `catalog-browser`

### Shared / read-only

11. `task-inspector`
12. `worker-inspector`
13. `role-browser`

### Strategy

14. `context-cleanup`

## Skill Spec Summary

### 1. task-orchestrator
- Audience: controller, human
- Triggers: create task, assign task, 完成任务, archive task, task flow
- CLI: `agent-team task create/list/show/assign/done/archive`
- Required entry: `.agents/rules/index.md`
- Expansion: task 工件与所需规则
- Purpose: 任务生命周期总入口

### 2. workflow-orchestrator
- Audience: controller, human
- Triggers: workflow plan, approve plan, activate plan, close plan
- CLI: `agent-team workflow plan generate/approve/activate/close`
- Required entry: `.agents/rules/index.md`
- Expansion: workflow 相关工件
- Purpose: 治理 workflow 总入口

### 3. worker-dispatch
- Audience: controller, human
- Triggers: open worker, 派发 worker, 给 worker 回复, 查看 worker 状态
- CLI: `agent-team worker open/status`, `agent-team reply`
- Required entry: `.agents/rules/index.md`
- Expansion: worker config、task 引用
- Purpose: controller 到 worker 的调度入口

### 4. worker-recovery
- Audience: worker
- Triggers: 恢复任务, 继续干活, resume work
- CLI: 主要读工件，必要时 `agent-team task show`
- Required entry: `worker.yaml`
- Expansion: `task.yaml`、`context.md`、被任务引用的设计文档
- Purpose: worker 标准恢复入口

### 5. worker-reply-main
- Audience: worker
- Triggers: reply main, 汇报完成, 请求决策, blocked
- CLI: `agent-team reply-main`
- Required entry: `worker.yaml`
- Expansion: 当前任务的最小必要摘要
- Purpose: worker -> main 汇报入口

### 6. project-bootstrap
- Audience: human, controller
- Triggers: init project, setup agent-team, migrate project
- CLI: `agent-team init/setup/migrate`
- Required entry: `.agents/rules/index.md`
- Expansion: 初始化相关配置文件
- Purpose: 仓库接入与迁移

### 7. rules-maintenance
- Audience: human, controller
- Triggers: sync rules, 更新规则, rule drift
- CLI: `agent-team rules sync`
- Required entry: `.agents/rules/index.md`
- Expansion: 命中的规则文件
- Purpose: 规则维护

### 8. skill-maintenance
- Audience: human, controller
- Triggers: check skills, update skills, clean skills
- CLI: `agent-team skill check/update/clean`
- Required entry: `.agents/rules/index.md`
- Expansion: skill cache 与依赖信息
- Purpose: 技能缓存维护

### 9. role-repo-manager
- Audience: human, controller
- Triggers: search role repo, add role repo, update role repo
- CLI: `agent-team role-repo find/add/list/check/update/remove`
- Required entry: `.agents/rules/index.md`
- Expansion: role repo 检索结果与目标 role
- Purpose: role 来源管理

### 10. catalog-browser
- Audience: human, controller
- Triggers: search catalog, browse roles, show catalog role
- CLI: `agent-team catalog search/show/list/repo/stats`
- Required entry: `.agents/rules/index.md`
- Expansion: catalog 命中项
- Purpose: 角色目录浏览

### 11. task-inspector
- Audience: controller, worker, human
- Triggers: 查看任务, inspect task, task status
- CLI: `agent-team task list/show`
- Required entry: controller/human 读 `.agents/rules/index.md`；worker 读 `worker.yaml`
- Expansion: task 工件
- Purpose: 只读任务查看

### 12. worker-inspector
- Audience: controller, human
- Triggers: worker status, inspect worker
- CLI: `agent-team worker status`
- Required entry: `.agents/rules/index.md`
- Expansion: worker 状态配置
- Purpose: 只读 worker 查看

### 13. role-browser
- Audience: controller, worker, human
- Triggers: role list, 查看角色
- CLI: `agent-team role list`
- Required entry: `.agents/rules/index.md`
- Expansion: 本地 role 详情
- Purpose: 本地 role 浏览

### 14. context-cleanup
- Audience: controller, worker
- Triggers: clean context, 会话乱了, 重新锚定, 恢复稳定上下文
- CLI: 这是策略 skill，不等价于 `compact`
- Required entry:
  - controller: `.agents/rules/index.md`
  - worker: `worker.yaml`
- Expansion:
  - controller: 命中的规则、当前 workflow/task 工件
  - worker: `task.yaml`、`context.md`、必要引用
- Purpose: 判断何时应清理会话上下文，并在清理后按索引优先模型重新读取文件上下文

## Context Strategy Design

### Why a dedicated context skill still exists

虽然当前系统以文件工件为上下文来源，但仍需要一个类似 `strategic-compact` 的**策略层 skill**，原因是：

- AI 仍然会在长会话中积累临时推理负担。
- 主问题不是“缺文件”，而是“会话上下文失稳”。
- 因此需要一个 skill 专门判断：**什么时候该清理当前会话上下文，并重新从文件锚点进入。**

### Hard rules for context-cleanup

- 清理的是**会话上下文**，不是文件内容。
- 不能继续使用 `/compact` 或“压缩上下文”语义。
- 清理后必须**先读取索引文件**。
- 之后根据索引命中结果，再读取所需正文文件。
- 不允许直接从一组正文文件中猜测当前入口。
- 不允许默认全量扫所有上下文正文。

### Index-first recovery model

#### Controller
1. 读取 `.agents/rules/index.md`
2. 识别当前需要命中的规则入口
3. 读取匹配到的规则正文
4. 再按当前 workflow/task 目标读取所需工件

#### Worker
1. 读取 `worker.yaml`
2. 从中确认当前任务锚点
3. 读取 `task.yaml`
4. 必要时读取 `context.md` 与被任务引用的补充材料

## Risks and Mitigations

### Risk 1: skill 边界重叠
**Mitigation:** 明确 controller-first / worker-first / human-first / shared / strategy 分层，并在 skill 文档里写清楚谁是主受众。

### Risk 2: 上下文恢复重新退化成会话记忆驱动
**Mitigation:** 将“索引优先”写成硬规则；所有相关 skill 先读入口索引，再展开正文。

### Risk 3: skill 数量变多带来维护成本
**Mitigation:** 保留 14 个固定技能，但每个 skill 只负责一个稳定场景，不再继续按子命令细拆。

### Risk 4: `context-cleanup` 被误用成 `/compact` 替身
**Mitigation:** 在 skill 文档中明确禁止 compact 语义，强调它只是会话清理 + 文件重锚。

## Validation Strategy

### Design validation
- 检查 14 个 skill 是否能完整覆盖当前主要 CLI 用户场景。
- 检查每个 skill 是否都有明确受众、触发词、命令绑定、上下文入口。
- 检查 `context-cleanup` 是否与 `/compact` 语义彻底分离。

### Implementation validation (future)
- 为每个 skill 编写最小 SKILL.md，验证其触发语义是否稳定。
- 抽样测试 controller、worker、human 三类典型 prompts 是否能路由到正确 skill。
- 验证上下文类 skill 是否总能先走索引入口，而不是直接跳正文。

## Open Questions

当前已确认的设计中，没有阻塞性 open question。

后续实现阶段仍可单独再决定：
- `context-cleanup` 的最终目录命名是否保持该名字；
- 14 个 skill 的实际落地顺序是否采用两波实施；
- 是否为部分 shared skill 增加更细的触发词约束。