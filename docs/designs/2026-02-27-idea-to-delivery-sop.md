# Idea-to-Delivery 标准作业程序（SOP）

## 1. 文档目的

定义从“业务想法”到“产品上线并验证价值”的统一执行标准，确保跨角色协作可复制、可审计、可复盘。

## 2. 适用范围

- 适用于新功能、新产品、重大优化、跨模块改造
- 不适用于纯 typo、纯配置微调、纯修复且无行为变化的小改动

## 3. 角色与职责（RACI）

- Sponsor（业务发起人）
  - A: 业务目标、优先级、最终业务验收
- PM（产品经理）
  - A/R: 需求澄清、PRD、范围冻结、迭代节奏
- Strategy/Analyst（商业分析）
  - R: 业务分析、假设与验证路径
- Tech Lead/Architect（技术负责人）
  - A/R: 技术方案、任务拆解、工程风险控制
- FE Architect（前端架构）
  - R: 前端框架搭建、工程规范、架构验收
- BE Architect（后端架构）
  - R: 后端框架搭建、服务边界、架构验收
- FE/BE Engineers（研发）
  - R: 实现、单测、联调修复
- QA（测试）
  - A/R: 测试策略、质量门禁、发布前验收
- DevOps/SRE（发布运维）
  - A/R: 发布执行、监控、回滚与稳定性保障

## 4. 端到端流程

```text
P0 Idea Intake
 -> P1 Discovery（脑暴澄清）
 -> P2 Business Framing（商业分析）
 -> P3 Product Planning（产品规划）
 -> P3A Product Design Archive（归档产品设计内容）
 -> P4A FE Architecture Baseline（前端架构搭建）
 -> P4B BE Architecture Baseline（后端架构搭建）
 -> P4C Contract Freeze & Dev Handoff（契约冻结与研发交接）
 -> P5A UI/UX Design（设计产出）
 -> P5B Backend Development（后端研发）
 -> P5B Backend Baseline（后端架构验收）
 -> P5C Frontend Development（前端研发）
 -> P5C Frontend Baseline（前端架构验收）
 -> P5D Integration（前后端联调）
 -> P5E CTO Acceptance（CTO验收）
 -> P6 QA & Release Readiness（测试与发布准备）
 -> P7 Launch & Hypercare（上线与观察）
 -> P8 Retrospective（复盘迭代）
```

## 5. 分阶段 SOP（输入、动作、产物、门禁、责任人）

## P0 Idea Intake（想法输入）

- Owner: Sponsor
- 输入:
  - 初始想法、业务背景、目标窗口
- 标准动作:
  - 填写 Idea Brief（问题、目标用户、业务目标、时间预算）
- 产物:
  - `idea-brief.md`
- 进入下一阶段门禁:
  - 问题陈述可验证（不是“直接指定方案”）

## P1 Discovery（需求脑暴澄清）

- Owner: PM
- 核心 Skill:
  - `brainstorming`
- 输入:
  - `idea-brief.md`
- 标准动作:
  - 一次一个问题澄清目标/约束/成功标准
  - 提供 2-3 个可选方向并推荐
  - 获取显式批准
- 产物:
  - `design-draft.md`
- 门禁（硬性）:
  - 未获批准不得进入实现相关动作

## P2 Business Framing（商业分析）

- Owner: Strategy/PM
- 核心 Skill:
  - `mckinsey-consultant`（至少 STEP 1-5）
- 输入:
  - `design-draft.md`
- 标准动作:
  - 形成 Issue Tree
  - 形成可验证 Hypotheses
  - 输出 Dummy 论证结构
- 产物:
  - `issue-tree.md`
  - `hypotheses.md`
  - `dummy-pages.md`
- 门禁:
  - 每条核心假设都有验证方法和数据来源

## P3 Product Planning（产品规划与范围冻结）

- Owner: PM
- 输入:
  - `issue-tree.md`
  - `hypotheses.md`
  - `dummy-pages.md`
- 标准动作:
  - 产出 PRD（用户、场景、FR/NFR、非目标）
  - 设定范围优先级（MVP/v1/v2）
  - 定义验收标准
- 产物:
  - `prd.md`
  - `release-scope.md`
  - `backlog.md`
- 门禁:
  - 范围冻结并通过跨职能评审（PM/Tech Lead/QA/Sponsor）

## P3A Product Design Archive（归档产品设计内容）

- Owner: PM（UI/UX + Tech Lead 协作）
- 输入:
  - `prd.md`
  - `release-scope.md`
  - `backlog.md`
- 标准动作:
  - 对产品设计产物进行版本化归档（页面流、交互、约束）
  - 标注“冻结版本”和“变更窗口”
  - 归档后禁止直接改动，变更需走评审
- 产物:
  - `product-design-archive.md`
- 门禁:
  - 归档版本已发布并被 FE/BE/QA 确认

## P4A FE Architecture Baseline（前端架构搭建）

- Owner: FE Architect
- 核心 Skill:
  - `writing-plans`
  - `using-git-worktrees`
- 输入:
  - `prd.md`
  - `release-scope.md`
  - `ui-spec.md`（若已完成 P5A，可回填）
- 标准动作:
  - 搭建前端基础框架（目录、路由、状态管理、网络层、错误边界）
  - 建立组件规范、代码规范、测试基线
  - 输出前端脚手架示例页面与开发约定
- 产物:
  - `fe-architecture-baseline.md`
  - `frontend-bootstrap-report.md`
- 门禁:
  - FE Architect 确认框架可扩展
  - FE Lead 确认研发可接入

## P4B BE Architecture Baseline（后端架构搭建）

- Owner: BE Architect
- 核心 Skill:
  - `writing-plans`
  - `using-git-worktrees`
- 输入:
  - `prd.md`
  - `release-scope.md`
- 标准动作:
  - 搭建后端基础框架（分层结构、配置管理、日志、鉴权、中间件、健康检查）
  - 建立数据访问规范与错误处理规范
  - 输出核心接口骨架与服务模板
- 产物:
  - `be-architecture-baseline.md`
  - `backend-bootstrap-report.md`
  - `technical-design.md`
- 门禁:
  - BE Architect 确认服务边界清晰
  - BE Lead 确认研发可接入

## P4C Contract Freeze & Dev Handoff（契约冻结与研发交接）

- Owner: Tech Lead（FE Architect + BE Architect 共同参与）
- 输入:
  - `fe-architecture-baseline.md`
  - `be-architecture-baseline.md`
  - `prd.md`
- 标准动作:
  - 固化 API 契约与错误码
  - 固化前后端联调约定（鉴权、分页、时区、版本）
  - 定义 TDD 先行测试清单（后端/前端/契约/关键链路）
  - 拆分 FE/BE 研发任务并明确验收标准
- 产物:
  - `api-contract.md`
  - `architecture-handoff.md`
  - `implementation-plan.md`
  - `tdd-plan.md`
  - `task-breakdown.md`
- 门禁:
  - API 契约冻结（变更需走评审）
  - TDD 清单已评审通过（Tech Lead + QA）
  - 研发任务可独立执行

## P5A UI/UX Design（UI 设计产出）

- Owner: UI/UX Designer（无专职时由 FE 代理）
- 输入:
  - `prd.md`
  - `release-scope.md`
  - 品牌规范/设计系统（如有）
- 标准动作:
  - 产出关键页面线框和高保真稿
  - 输出组件清单、交互状态、响应式断点
  - 与 PM/FE/QA 评审可实现性
- 产物:
  - `ui-spec.md`
  - `screen-flow.md`
  - `component-spec.md`
- 门禁:
  - PM 确认业务一致性
  - FE 确认可实现
  - QA 确认可测试

## P5B Backend Development（后端研发）

- Owner: Backend Engineer
- 核心 Skill:
  - `subagent-driven-development` 或 `executing-plans`
  - `test-driven-development`
  - `systematic-debugging`
- 输入:
  - `technical-design.md`
  - `task-breakdown.md`
  - `api-contract.md`（来自方案阶段）
- 标准动作:
  - 逐任务执行 TDD（RED -> GREEN -> REFACTOR）
  - 实现接口、领域逻辑、数据模型迁移
  - 编写单元/集成测试
  - 输出接口变更说明与示例
- 产物:
  - 后端代码与测试
  - `backend-tdd-evidence.md`
  - `backend-acceptance.md`
  - `api-changelog.md`
- 门禁:
  - 接口契约通过（字段/状态码/错误码一致）
  - 单元与集成测试通过
  - 无高危安全问题

## P5B Backend Baseline（后端架构验收）

- Owner: BE Architect + Tech Lead
- 输入:
  - `backend-acceptance.md`
  - `backend-tdd-evidence.md`
  - `technical-design.md`
  - `api-contract.md`
- 标准动作:
  - 对照后端架构基线检查目录分层、依赖方向、异常与日志规范
  - 检查是否存在绕过架构约束的临时实现
- 产物:
  - `backend-baseline-acceptance.md`
- 门禁:
  - 后端架构验收通过，才可进入联调阶段

## P5C Frontend Development（前端研发）

- Owner: Frontend Engineer
- 核心 Skill:
  - `subagent-driven-development` 或 `executing-plans`
  - `test-driven-development`
  - `systematic-debugging`
- 输入:
  - `ui-spec.md`
  - `component-spec.md`
  - `api-contract.md`
- 标准动作:
  - 逐任务执行 TDD（RED -> GREEN -> REFACTOR）
  - 实现页面、组件、状态管理、接口调用
  - 覆盖关键交互（加载/空态/异常态）
  - 补齐前端测试（单测/组件测试）
- 产物:
  - 前端代码与测试
  - `frontend-tdd-evidence.md`
  - `frontend-acceptance.md`
- 门禁:
  - UI 与设计稿关键项一致
  - 关键路径可用且错误态可恢复
  - 前端测试和构建通过

## P5C Frontend Baseline（前端架构验收）

- Owner: FE Architect + Tech Lead
- 输入:
  - `frontend-acceptance.md`
  - `frontend-tdd-evidence.md`
  - `fe-architecture-baseline.md`
  - `api-contract.md`
- 标准动作:
  - 对照前端架构基线检查目录结构、状态管理、边界分层、错误处理
  - 检查 UI 工程实现是否偏离约定
- 产物:
  - `frontend-baseline-acceptance.md`
- 门禁:
  - 前端架构验收通过，才可进入联调阶段

## P5D Integration（前后端联调）

- Owner: FE + BE（Tech Lead 负责协调）
- 输入:
  - `frontend-acceptance.md`
  - `backend-acceptance.md`
  - `frontend-baseline-acceptance.md`
  - `backend-baseline-acceptance.md`
- 标准动作:
  - 对齐环境变量、鉴权、跨域、时区、分页等契约细节
  - 执行端到端关键链路联调
  - 修复阻塞性联调问题
- 产物:
  - `integration-report.md`
- 门禁:
  - 关键业务链路全通
  - 无 P0/P1 联调缺陷

## P5E CTO Acceptance（CTO验收）

- Owner: CTO（Tech Lead/FE Architect/BE Architect 参与）
- 输入:
  - `backend-baseline-acceptance.md`
  - `frontend-baseline-acceptance.md`
  - `integration-report.md`
  - `architecture-handoff.md`
- 标准动作:
  - 对照业务目标、架构约束、关键链路结果进行最终技术验收
  - 评估上线风险与技术债处理策略
  - 输出“通过/不通过”结论及强制整改项
- 产物:
  - `cto-acceptance.md`
- 门禁:
  - CTO 验收通过后才可进入 P6 QA
  - 不通过则退回 P5B/P5C/P5D 修复

## P6 QA & Release Readiness（测试与发布准备）

- Owner: QA / Release Owner
- 核心 Skill:
  - `requesting-code-review`
  - `receiving-code-review`
  - `verification-before-completion`
  - `playwright`（如需 E2E）
- 输入:
  - 构建产物与变更列表
- 标准动作:
  - 功能回归、E2E、关键路径校验
  - 发布前检查（监控、告警、回滚、配置）
- 产物:
  - `test-report.md`
  - `release-checklist.md`
- 门禁:
  - Go/No-Go 评审通过

## P7 Launch & Hypercare（上线与观察）

- Owner: DevOps/SRE
- 核心 Skill:
  - `deployment-config-validate`
  - `deployment-execute`
  - `deployment-post-checks`
  - `deployment-record-archive`
  - `finishing-a-development-branch`
- 输入:
  - `release-checklist.md`
  - 发布配置
- 标准动作:
  - 分批发布/灰度
  - 监控关键指标，异常触发回滚
- 产物:
  - `deployment-record.md`
  - `hypercare-summary.md`
- 门禁:
  - 观察窗口内稳定，无 P0/P1 事故

## P8 Retrospective（复盘与下一轮）

- Owner: PM + Tech Lead
- 输入:
  - 上线结果、指标看板、事故记录
- 标准动作:
  - 对比目标与结果
  - 识别偏差根因
  - 形成下一轮 backlog
- 产物:
  - `retrospective.md`
  - `next-iteration-backlog.md`
- 完成定义:
  - 本轮目标、质量、业务结果均闭环

## 5.1 TDD 集成机制（必选）

- 适用阶段:
  - P5B Backend Development
  - P5C Frontend Development
  - P5D Integration（关键链路自动化回归）
- 执行规则:
  - 每个任务按 `RED -> GREEN -> REFACTOR` 执行
  - 先写失败测试（RED），再写最小实现（GREEN），再重构（REFACTOR）
  - 禁止“先写完功能再补测试”作为常规路径
- 阶段绑定:
  - 在 P4C 产出 `tdd-plan.md`（用例清单、优先级、Owner、完成标准）
  - 在 P5B 输出 `backend-tdd-evidence.md`
  - 在 P5C 输出 `frontend-tdd-evidence.md`
  - 在 P5D 至少保留关键链路自动化结果
- 验收要求:
  - 未提交 TDD 证据文档，不得通过 P5B/P5C 阶段验收
  - 关键链路自动化失败，不得进入 P5E/P6

## 6. 关键门禁清单（DoD）

1. Discovery DoD
   - 目标、约束、成功标准已确认
2. Business DoD
   - 假设可验证，数据来源可追溯
3. Product DoD
   - PRD 完整，范围冻结，验收标准明确
4. Product Archive DoD
   - 设计归档版本已冻结并完成跨角色确认
5. FE Architecture DoD
   - 前端框架和工程规范可复用、可扩展
6. BE Architecture DoD
   - 后端服务框架与边界清晰、可扩展
7. Contract Freeze DoD
   - API 契约冻结并完成研发交接
8. TDD Plan DoD
   - `tdd-plan.md` 完整并通过评审
9. UI Design DoD
   - 关键页面/组件/交互状态定义完整
10. Backend Development DoD
   - API 契约一致，后端测试通过，TDD 证据完整
11. Backend Baseline DoD
   - 后端架构验收通过
12. Frontend Development DoD
   - UI 一致性和关键交互通过，TDD 证据完整
13. Frontend Baseline DoD
   - 前端架构验收通过
14. Integration DoD
   - 关键链路联调通过，无阻塞缺陷
15. CTO Acceptance DoD
   - CTO 验收通过并明确上线风险结论
16. QA DoD
   - 回归/E2E 通过，无阻塞缺陷
17. Release DoD
   - 发布检查单完成，监控和回滚链路可用
18. Outcome DoD
   - KPI 达标或偏差有明确行动项

## 7. 阶段验收文档与签字矩阵

| 阶段 | 必验收内容 | 必备验收文档 | 验收人（签字） |
|---|---|---|---|
| P1 Discovery | 目标、约束、成功标准清晰 | `design-draft.md` | PM, Sponsor |
| P2 Business | 假设可验证、来源可追溯 | `issue-tree.md`, `hypotheses.md`, `dummy-pages.md` | Strategy, PM |
| P3 Product | 范围冻结、验收标准明确 | `prd.md`, `release-scope.md`, `backlog.md` | PM, Tech Lead, QA |
| P3A 归档 | 设计版本冻结、可追溯 | `product-design-archive.md` | PM, FE Lead, BE Lead, QA |
| P4A FE 架构 | 前端框架可复用、可扩展 | `fe-architecture-baseline.md`, `frontend-bootstrap-report.md` | FE Architect, FE Lead |
| P4B BE 架构 | 后端框架边界清晰、可扩展 | `be-architecture-baseline.md`, `backend-bootstrap-report.md`, `technical-design.md` | BE Architect, BE Lead |
| P4C 交接 | 契约冻结、研发可接手 | `api-contract.md`, `architecture-handoff.md`, `implementation-plan.md`, `tdd-plan.md`, `task-breakdown.md` | Tech Lead, FE Architect, BE Architect, QA |
| P5A UI | 设计可实现、可测试 | `ui-spec.md`, `screen-flow.md`, `component-spec.md` | PM, FE Lead, QA |
| P5B BE 研发 | 接口与逻辑达标 | `backend-acceptance.md`, `backend-tdd-evidence.md`, `api-changelog.md` | BE Lead, Tech Lead |
| P5B BE 架构验收 | 后端实现符合架构基线 | `backend-baseline-acceptance.md` | BE Architect, Tech Lead |
| P5C FE 研发 | UI与交互达标 | `frontend-acceptance.md`, `frontend-tdd-evidence.md` | FE Lead, PM |
| P5C FE 架构验收 | 前端实现符合架构基线 | `frontend-baseline-acceptance.md` | FE Architect, Tech Lead |
| P5D 联调 | 关键链路打通 | `integration-report.md` | Tech Lead, FE Lead, BE Lead |
| P5E CTO 验收 | 关键链路与架构目标通过最终技术验收 | `cto-acceptance.md` | CTO, Tech Lead, PM |
| P6 QA | 回归/E2E/风险评估通过 | `test-report.md`, `release-checklist.md` | QA Lead, PM |
| P7 发布 | 灰度稳定、可回滚 | `deployment-record.md`, `hypercare-summary.md` | DevOps/SRE, QA |
| P8 复盘 | KPI 对比、行动项明确 | `retrospective.md`, `next-iteration-backlog.md` | PM, Tech Lead, Sponsor |

## 8. 会议与节奏（默认）

- 每日 15 分钟站会（PM 主持）
- 每周一次范围与风险评审（PM + Tech Lead + QA + Sponsor）
- 发布前 Go/No-Go（QA/DevOps 主持）
- 上线后 24h/72h 两次健康检查
- 迭代结束复盘会（PM + Tech Lead）

## 9. 时效 SLA（建议）

- P1 Discovery: 0.5-2 天
- P2 Business Framing: 1-3 天
- P3 Product Planning: 1-2 天
- P3A Product Design Archive: 0.5 天
- P4A FE Architecture Baseline: 1-3 天
- P4B BE Architecture Baseline: 1-3 天
- P4C Contract Freeze & Handoff: 0.5-1 天
- P5A UI/UX Design: 1-3 天
- P5B Backend Development: 2-7 天
- P5B Backend Baseline: 0.5-1 天
- P5C Frontend Development: 2-7 天
- P5C Frontend Baseline: 0.5-1 天
- P5D Integration: 1-3 天
- P5E CTO Acceptance: 0.5-1 天
- P6 QA: 1-2 天
- P7 Hypercare: 1-3 天

## 10. 升级与决策机制

- 范围冲突: PM 仲裁，Sponsor 最终决策
- 技术分歧: Tech Lead 仲裁，必要时架构评审
- 架构验收不通过: 必须退回对应研发阶段整改，不允许带病进入 QA
- CTO 验收不通过: 必须回退 P5B/P5C/P5D 定位并修复
- 质量风险: QA 可一票否决上线
- 线上事故: DevOps/SRE 启动回滚流程并同步全员

## 11. agent-team 执行指令（示例）

1. 创建 worker
`agent-team worker create <role-name>`

2. 打开 worker 会话
`agent-team worker open <worker-id> codex`

3. 分配任务（带 design/proposal）
`agent-team worker assign <worker-id> "<desc>" --design <design.md> --proposal <proposal.md>`

4. 主控与 worker 双向沟通
`agent-team reply <worker-id> "<message>"`
`agent-team reply-main "<message>"`

5. 合并与收尾
`agent-team worker merge <worker-id>`
`agent-team worker delete <worker-id>`

## 12. 文档归档规范

建议统一放到可跟踪目录：
`openspec/changes/<change-id>/analysis/`

最小归档集：
- `idea-brief.md`
- `design-draft.md`
- `issue-tree.md`
- `hypotheses.md`
- `prd.md`
- `product-design-archive.md`
- `fe-architecture-baseline.md`
- `be-architecture-baseline.md`
- `frontend-bootstrap-report.md`
- `backend-bootstrap-report.md`
- `architecture-handoff.md`
- `ui-spec.md`
- `component-spec.md`
- `api-contract.md`
- `technical-design.md`
- `implementation-plan.md`
- `tdd-plan.md`
- `backend-tdd-evidence.md`
- `frontend-tdd-evidence.md`
- `backend-acceptance.md`
- `backend-baseline-acceptance.md`
- `frontend-acceptance.md`
- `frontend-baseline-acceptance.md`
- `integration-report.md`
- `cto-acceptance.md`
- `test-report.md`
- `deployment-record.md`
- `retrospective.md`

## 13. SOP 启动模板（复制即用）

```text
[项目名称]
[业务目标/KPI]
[时间窗口]
[MVP范围]
[非目标]
[角色Owner]
[当前阶段: P0-P8]
[下一门禁]
```

## 14. 验收文档模板（复制即用）

## 14.1 `backend-acceptance.md`

```markdown
# Backend Acceptance

## Scope
- 本次后端范围:

## Contract Check
- 接口列表:
- 字段一致性: Pass/Fail
- 状态码一致性: Pass/Fail
- 错误码一致性: Pass/Fail

## Test Results
- Unit:
- Integration:
- 覆盖率:

## Risks
- 已知风险:
- 缓解方案:

## Sign-off
- BE Lead:
- Tech Lead:
```

## 14.2 `frontend-acceptance.md`

```markdown
# Frontend Acceptance

## Scope
- 本次前端范围:

## UI/UX Check
- 页面清单:
- 设计一致性: Pass/Fail
- 响应式断点: Pass/Fail
- 空态/错误态: Pass/Fail

## Test Results
- Unit/Component:
- E2E(如有):

## Risks
- 已知风险:
- 缓解方案:

## Sign-off
- FE Lead:
- PM:
```

## 14.3 `integration-report.md`

```markdown
# Integration Report

## Critical User Journeys
- Journey 1:
- Journey 2:

## Integration Defects
- P0:
- P1:
- P2:

## Environment Check
- 鉴权:
- 配置:
- 时区/分页:

## Final Decision
- Ready for QA: Yes/No

## Sign-off
- FE Lead:
- BE Lead:
- Tech Lead:
```

## 14.4 `release-checklist.md`

```markdown
# Release Checklist

## Pre-release
- 变更说明完整
- 回滚脚本可用
- 监控告警已配置
- 数据迁移已演练

## Go/No-Go
- QA: Go/No-Go
- DevOps: Go/No-Go
- PM: Go/No-Go

## Sign-off
- QA Lead:
- DevOps/SRE:
- PM:
```

## 14.5 `architecture-handoff.md`

```markdown
# Architecture Handoff

## Baseline Summary
- FE 架构基线:
- BE 架构基线:

## Frozen Contracts
- API 契约版本:
- 错误码规范:
- 鉴权规范:
- 分页/时区规范:

## Development Tasks
- FE 任务清单:
- BE 任务清单:
- 联调任务清单:

## Constraints
- 禁止跨层调用:
- 禁止临时绕过项:

## Sign-off
- FE Architect:
- BE Architect:
- Tech Lead:
```

## 14.6 `backend-baseline-acceptance.md`

```markdown
# Backend Baseline Acceptance

## Scope
- 验收范围:

## Rule Compliance
- 目录结构符合后端基线: Pass/Fail
- 依赖边界符合约束: Pass/Fail
- API 契约一致: Pass/Fail
- 异常与日志规范一致: Pass/Fail

## Deviations
- 偏差项:
- 是否允许带技术债上线: Yes/No
- 偿还计划:

## Decision
- 结论: Pass/Fail
- Fail 时退回阶段: P5B

## Sign-off
- BE Architect:
- Tech Lead:
```

## 14.7 `tdd-plan.md`

```markdown
# TDD Plan

## Scope
- 覆盖范围:

## Test Inventory
- Backend Unit:
- Backend Integration:
- Frontend Unit/Component:
- Contract Test:
- Critical E2E:

## Priority
- P0:
- P1:
- P2:

## Owners
- BE:
- FE:
- QA:

## Entry/Exit Criteria
- Entry: 契约冻结完成
- Exit: P0/P1 用例全通过

## Sign-off
- Tech Lead:
- QA Lead:
```

## 14.8 `backend-tdd-evidence.md`

```markdown
# Backend TDD Evidence

## RED
- 失败测试用例清单:

## GREEN
- 通过测试用例清单:

## REFACTOR
- 重构项:
- 行为一致性校验:

## Result
- 通过率:
- 未通过项及处理:

## Sign-off
- Backend Engineer:
- BE Lead:
```

## 14.9 `frontend-tdd-evidence.md`

```markdown
# Frontend TDD Evidence

## RED
- 失败测试用例清单:

## GREEN
- 通过测试用例清单:

## REFACTOR
- 重构项:
- UI 行为一致性校验:

## Result
- 通过率:
- 未通过项及处理:

## Sign-off
- Frontend Engineer:
- FE Lead:
```

## 14.10 `cto-acceptance.md`

```markdown
# CTO Acceptance

## Scope
- 验收范围:

## Decision Inputs
- `backend-baseline-acceptance.md`
- `frontend-baseline-acceptance.md`
- `integration-report.md`
- `architecture-handoff.md`

## Risk Assessment
- 上线风险等级:
- 技术债清单与偿还计划:

## Decision
- Pass/Fail
- Fail 回退阶段: P5B / P5C / P5D

## Sign-off
- CTO:
- Tech Lead:
- PM:
```

## 14.11 `frontend-baseline-acceptance.md`

```markdown
# Frontend Baseline Acceptance

## Scope
- 验收范围:

## Rule Compliance
- 目录结构符合前端基线: Pass/Fail
- 状态管理符合规范: Pass/Fail
- API 契约一致: Pass/Fail
- 错误处理与边界符合规范: Pass/Fail

## Deviations
- 偏差项:
- 是否允许带技术债进入联调: Yes/No
- 偿还计划:

## Decision
- 结论: Pass/Fail
- Fail 时退回阶段: P5C

## Sign-off
- FE Architect:
- Tech Lead:
```
