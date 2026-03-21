# 多用户群聊智能体系统最佳实践（Claude API / Agent SDK 主架构）

日期：2026-03-21

## 结论

如果目标是构建一个：

- 多用户
- 多房间
- 支持群聊协作
- 支持人工审批 / A-B 选择 / 远程决策
- 后续可扩展为多智能体协作

的系统，推荐把 **Claude API / Agent SDK** 作为主架构，把 **Claude Code** 视为可选执行节点，而不是把多个 Claude Code 窗口当作系统主干。

一句话原则：

> **群聊系统的主状态、主流程、主权限控制都放在你自己的服务里；模型只是参与者，不是系统控制器。**

---

## 为什么推荐 Claude API / Agent SDK，而不是多开 Claude 窗口

### 推荐方案：Claude API / Agent SDK

优点：

- 服务端状态可控，便于持久化与恢复
- 更容易支持多用户、多房间、多角色
- 更容易做消息审计、权限控制、重试与幂等
- 更容易接入人工审批、外部表单、远程按钮选择
- 更容易扩容、部署、灰度和监控
- 不依赖终端窗口、TUI 状态或本地交互生命周期

### 不推荐作为主架构：多开 Claude Code / 多开窗口

问题：

- 会话生命周期难控
- 上下文清理、恢复、重启复杂
- 自动化依赖窗口/终端层，脆弱
- 难以做真正服务端产品化
- 难以统一做权限、审计、隔离与容量管理

适用边界：

- 内部实验
- 本地工程多 worker 编排
- 需要强依赖 Claude Code 本地 repo/tool/worktree 能力的任务

---

## 推荐架构

```text
Web / Mobile / IM Client
        |
        v
Group Chat Service
  - room state
  - message routing
  - identity / auth
  - approval workflow
  - audit log
        |
        v
Agent Orchestrator
  - participant registry
  - turn policy
  - tool policy
  - retry / timeout / idempotency
        |
        +-------------------+
        |                   |
        v                   v
Claude API / Agent SDK   Optional Execution Workers
                         - Claude Code worker
                         - MCP tools
                         - repo task runner
                         - browser / shell executor
```

### 核心原则

1. **你的服务持有真状态**
2. **模型输出只是候选动作，不直接等于系统动作**
3. **所有外部副作用都经过审批或策略层**
4. **把 Claude Code 限定为专用执行器，不当主会话总线**

---

## 推荐的领域模型

### 1. Room

表示一个群聊空间。

建议字段：

- `room_id`
- `title`
- `members`
- `shared_context_version`
- `status`（active / paused / archived）

### 2. Participant

表示群聊参与者，可以是人，也可以是智能体。

建议类型：

- `human`
- `agent`
- `system`

建议字段：

- `participant_id`
- `room_id`
- `type`
- `display_name`
- `role_profile`
- `tool_policy_id`
- `memory_scope`

### 3. Message

所有输入输出统一落为消息对象。

建议字段：

- `message_id`
- `room_id`
- `sender_id`
- `message_type`（user / agent / tool / system / approval_request / approval_result）
- `content`
- `reply_to`
- `created_at`
- `trace_id`

### 4. Decision / Approval

把“选 A / 选 B / 继续执行”设计成一等对象，不要把它埋进自然语言里。

建议字段：

- `decision_id`
- `room_id`
- `requested_by`
- `options`
- `selected_option`
- `status`（pending / approved / rejected / expired）
- `expires_at`
- `source`（web / mobile / api / operator）

### 5. Task / Action

将“说了什么”和“系统要做什么”分离。

建议字段：

- `task_id`
- `origin_message_id`
- `owner_agent_id`
- `status`
- `input_snapshot`
- `result_snapshot`
- `side_effects`

---

## 最佳实践一：让服务做主控，不让模型直接驱动系统副作用

### 推荐做法

模型只输出结构化建议，例如：

```json
{
  "intent": "request_approval",
  "question": "请选择部署环境",
  "options": ["staging", "production"],
  "reason": "production 会影响真实用户"
}
```

然后由你的服务决定：

- 是否展示为审批卡片
- 由谁审批
- 审批后触发哪个后续任务

### 不推荐做法

直接信任模型自然语言：

- “我建议直接发生产”
- “看起来可以继续”

这种文本不能直接驱动真实副作用。

---

## 最佳实践二：群聊上下文拆成三层，不要把所有内容直接塞给模型

建议分层：

### A. Room Shared Context

房间共享背景，例如：

- 当前项目目标
- 共享术语
- 房间规则
- 当前公共任务板

### B. Participant Private Context

参与者私有上下文，例如：

- 某个 agent 的角色说明
- 某个用户的私有偏好
- 某个执行器的本地状态

### C. Task Working Context

本轮任务临时上下文，例如：

- 本次需要回答的问题
- 当前选择分支
- 本轮工具调用结果

### 原则

> 模型每次只拿“当前任务真正需要的上下文子集”，不要直接吃整个群聊历史。

---

## 最佳实践三：审批流要产品化，不要只靠 hook 注入或隐式提示

如果你要支持“远程选 A / B”，推荐流程是：

```text
Agent 输出结构化审批请求
-> Service 创建 decision record
-> 推送到前端 / IM / Mobile
-> 用户点击 A/B
-> Service 写入 decision result
-> Orchestrator 继续任务
```

### 为什么不要把 hook 当审批主通道

hook 更适合：

- 出站通知
- 轻量提醒
- 附加上下文
- 本地自动化

hook 不适合当：

- 主状态机
- 主消息总线
- 跨会话控制平面

---

## 最佳实践四：所有 agent 都要有显式角色边界

每个 agent 至少要有：

- `role`
- `goal`
- `allowed_tools`
- `forbidden_actions`
- `escalation_policy`
- `approval_required_actions`

示例：

### 产品经理 Agent
- 负责需求拆解、优先级建议、验收标准草拟
- 不允许直接执行 shell / git / deploy

### 工程 Agent
- 负责实现建议、代码变更、测试建议
- 可调用 repo 内受控工具
- 遇到部署或数据变更必须升级审批

### 运维 Agent
- 可处理部署建议、环境检查
- 生产环境动作必须经人批准

---

## 最佳实践五：把 Claude Code 放在“执行层”，不要放在“控制层”

Claude Code 适合承担：

- 本地仓库读写
- 工具执行
- worktree 隔离执行
- 面向代码库的 agent 任务

Claude Code 不适合承担：

- 多用户群聊主状态
- 长生命周期审批总线
- 跨房间消息路由中心
- 服务端主编排器

### 推荐姿势

把 Claude Code worker 设计成：

- `execution worker`
- `repo operator`
- `specialized local agent`

由你的 Orchestrator 调用它，而不是让它反过来控制系统。

---

## 最佳实践六：结构化输出优先

不管是 Claude API 还是 Agent SDK，都建议优先让 agent 输出结构化对象，而不是只输出自然语言。

推荐输出类型：

- `next_action`
- `approval_request`
- `tool_plan`
- `task_status`
- `handoff_message`
- `final_answer`

这样你的服务才能：

- 校验
- 路由
- 审计
- 重试
- 幂等处理

---

## 最佳实践七：把幂等、重试、超时设计进系统，而不是事后补救

### 每个关键对象都要有稳定 ID

例如：

- `message_id`
- `task_id`
- `decision_id`
- `trace_id`
- `request_id`

### 重试原则

- 模型调用失败可重试
- 外部副作用调用必须幂等
- 审批提交必须防重复点击
- worker 执行结果必须可回放

### 超时原则

审批和 agent 任务都要可超时：

- `pending too long -> escalate`
- `worker timeout -> mark failed`
- `decision expired -> request refresh`

---

## 最佳实践八：把人工接管设计成系统能力，不要把它当异常路径

多智能体系统里，人类接管不是例外，而是核心能力。

推荐内建：

- 暂停任务
- 恢复任务
- 改选项
- 驳回 agent 决策
- 指派给其他 agent
- 强制结束本轮执行

### 人工接管常见场景

- agent 对风险操作拿不准
- 用户想选 A/B
- 需要业务负责人拍板
- 工具结果冲突
- 长任务中途改需求

---

## 最佳实践九：可观测性必须一开始就做

至少记录：

- 谁发起了什么请求
- 哪个 agent 在何时拿到了什么上下文
- 输出了什么结构化结果
- 是否触发了审批
- 审批是谁做的
- 最终执行了什么副作用
- 哪一步失败

建议维度：

- `trace_id`
- `room_id`
- `participant_id`
- `task_id`
- `decision_id`
- `tool_call_id`

### 推荐日志视角

1. 房间视角
2. 用户视角
3. agent 视角
4. 任务视角
5. 外部副作用视角

---

## 最佳实践十：安全边界要前置

### 强制分层

- 聊天层：用户与 agent 互动
- 编排层：路由、审批、状态机
- 执行层：工具、shell、repo、部署

### 高风险动作必须升级审批

例如：

- 部署生产
- 改数据库
- 删除资源
- 发送外部消息
- 修改共享仓库主分支

### 不要做的事

- 不要让一个普通聊天 agent 直接持有高权限执行能力
- 不要把生产动作藏进自然语言里直接执行
- 不要把“用户模糊同意”当作正式授权

---

## 什么时候用 Agent SDK，什么时候只用 Claude API

## 优先只用 Claude API 的场景

适合：

- 你已经有自己的编排层
- 你自己定义消息模型、状态机、任务模型
- 你只需要模型生成、工具选择、结构化输出

优点：

- 更简单
- 控制权更高
- 更容易与你现有系统整合

## 用 Agent SDK 的场景

适合：

- 你希望更快搭建 agent runtime
- 希望复用 agent 运行时模式
- 需要更标准化的 agent 执行框架

原则：

> 如果你的产品已经有明确的后端状态机，优先让 Agent SDK 服从你的编排层，而不是让 SDK 取代你的系统状态机。

---

## MVP 路线建议

### Phase 1：单房间 + 单 agent + 人工审批

目标：

- 跑通消息、agent 回复、审批按钮、结果继续执行

建议能力：

- room/message/decision 三张核心表
- 一个 agent
- 一个审批面板
- 一个 trace 视图

### Phase 2：多 participant + 多角色

增加：

- 产品 agent
- 工程 agent
- reviewer agent
- turn policy
- participant registry

### Phase 3：接入专用执行器

增加：

- Claude Code worker
- MCP 工具
- repo task runner
- browser executor

### Phase 4：可靠性与运维强化

增加：

- 重试队列
- 审计报表
- 成本控制
- 回放与恢复
- 灰度发布

---

## 一个推荐的最小消息流

```text
1. 用户在 room 发消息
2. Orchestrator 判断由哪个 agent 响应
3. Agent 输出结构化结果：answer / approval_request / tool_plan
4. 如果是普通回答，直接回 room
5. 如果是 approval_request，创建 decision record
6. 用户在前端点选 A / B
7. Service 写入 decision result
8. Orchestrator 根据结果继续推进 task
9. 如果需要代码执行，再调用 Claude Code worker 或其他执行器
10. 所有过程记录 trace 与 audit log
```

---

## 反模式

### 反模式 1：把整个系统做成“多个 Claude 窗口互相聊天”

问题：

- 难控
- 难恢复
- 难审计
- 难部署

### 反模式 2：把 hook 当主消息总线

问题：

- hook 适合局部自动化，不适合主状态机

### 反模式 3：用自然语言隐式表达关键控制命令

问题：

- 无法稳定解析
- 无法审计
- 无法保证安全

### 反模式 4：让所有 agent 共享一份巨大的上下文

问题：

- 成本高
- 容易漂移
- 推理质量下降

---

## 最终建议

如果你的目标是“多用户 + 群聊 + 多 agent + 人工审批 + 可扩展”，最佳实践是：

1. **用你自己的服务做系统主控**
2. **用 Claude API / Agent SDK 做 agent runtime**
3. **把审批、A/B 选择、恢复执行做成显式业务对象**
4. **把 Claude Code 限定为可选执行器，而不是主架构**
5. **从结构化输出、审计、幂等、人工接管开始设计**

一句话总结：

> **做产品系统时，Claude 是能力组件；你的服务才是系统本体。**
