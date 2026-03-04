# 角色管理站点功能细化（MVP → P2）

## 0. 文档定位
- 目标阶段：`0→1`（新产品从验证到可用）
- 适用范围：`role-hub`（独立仓库，Vercel + Neon Postgres）
- 上下游约束：
  - GitHub 为角色内容事实源
  - `agent-team role-repo find` 之后异步上报候选角色，不阻塞 CLI 主流程
  - 索引/检索使用 Postgres，不引入 Redis（MVP）

## 1. MVP 范围定义

### 1.1 MVP In Scope（必须交付）
1. 候选角色异步接入（Ingest API）
2. 角色标准化校验与状态流转（`discovered → verified|invalid|unreachable`）
3. 公开目录站点：列表、搜索、筛选、详情页
4. 仓库详情页（按仓库聚合角色）
5. 可观测性：接入成功率、校验耗时、失败分布

### 1.2 MVP Out of Scope（明确不做）
1. 用户登录体系与个性化收藏
2. 社区评价/评分/评论
3. 站内编辑角色内容（仍以 GitHub PR 为准）
4. 复杂推荐算法与个性化排序
5. 多数据源接入（GitLab/自托管 Git）
6. 管理后台（MVP 不建设）

### 1.3 MVP 成功标准
1. 目录可用：用户可在站点完成“搜索→查看详情→复制安装命令”
2. 接入可靠：CLI 异步上报失败不影响 `find` 命令返回
3. 数据可信：站点默认仅展示 `verified` 角色
4. 时效可控：新发现角色在目标时延内进入可见目录（见 NFR）

## 2. 用户流程（E2E）

### 2.1 流程 A：CLI 发现到站点可见（核心）
1. 用户执行 `agent-team role-repo find <query>`
2. CLI 立即展示结果（主流程结束）
3. CLI 后台异步调用 Ingest API 发送候选项批次
4. Ingest API 验签、幂等写入 `discovered`
5. 标准化任务拉取候选项，读取 GitHub 元数据并校验目录契约
6. 校验通过标记 `verified`，失败标记 `invalid` 或 `unreachable`
7. 站点列表与搜索仅读取 `verified` 记录

证据标注：
- Research：来源于当前变更 `design.md` 已确认架构与非阻塞原则
- Assumption：具体时延阈值需上线后按真实流量校准

### 2.2 流程 B：维护者更新角色后同步
1. 维护者更新 GitHub 仓库角色内容
2. 站点通过定时任务（MVP）触发增量同步
3. 检测 `folder_hash` 或提交变更并重跑校验
4. 更新目录字段和 `updated_at`

证据标注：
- Assumption：MVP 先用定时同步，Webhook 作为 P1

### 2.3 流程 C：管理员处理异常记录
（P1 可选，非 MVP）
1. 管理端查看 `invalid/unreachable` 列表
2. 查看失败原因（路径不合规、仓库不可达、速率限制等）
3. 执行“重试同步”或“加入黑名单”（P1）

证据标注：
- Research：来自前置 brainstorming 中“数据质量与审核需求”

## 3. 信息架构（IA）

## 3.1 公共站点（Public）
1. 首页（价值说明 + 快速搜索）
2. 角色目录页（列表/筛选/排序）
3. 角色详情页（元数据、安装方式、来源仓库、最近更新）
4. 仓库详情页（该仓库下角色集合，P0）
5. 文档页（路径契约、接入说明、常见错误）

### 3.2 管理后台（Admin）
1. 非 MVP。若后续启用，优先采用 GitHub 登录 + 单管理员白名单。

### 3.3 系统层（非页面）
1. Ingest API（鉴权 + 幂等写入）
2. Normalize Worker（拉取 + 校验 + 状态转换）
3. Query API（目录读模型）

## 4. 功能清单（P0 / P1 / P2）

| 模块 | 功能 | 优先级 | 说明 | 证据 |
|---|---|---|---|---|
| 数据接入 | CLI 异步上报候选角色 | P0 | `find` 后台 fire-and-forget | Research |
| 数据接入 | HMAC+时间戳验签 | P0 | 防伪造、防重放 | Research |
| 数据接入 | 批量幂等 upsert | P0 | key=`repo+role_path+ref` | Research |
| 校验引擎 | 路径契约校验 | P0 | 仅接受约定目录结构 | Research |
| 校验引擎 | 状态流转机 | P0 | discovered/verified/invalid/unreachable | Research |
| 目录查询 | 角色列表分页 | P0 | 默认只看 verified | Research |
| 目录查询 | 关键词搜索 | P0 | 名称、描述、标签 | Assumption |
| 目录查询 | 基础筛选 | P0 | 作用域、更新时间、来源 | Assumption |
| 详情展示 | 角色详情页 | P0 | 安装命令、依赖、来源 | Assumption |
| 目录体验 | 仓库详情页 | P0 | 按 repo 聚合查看 | Assumption |
| 可观测性 | 指标与日志 | P0 | ingest 成功率、延迟、失败原因 | Research |
| 数据同步 | GitHub webhook 同步 | P1 | 降低数据陈旧窗口 | Assumption |
| 管理后台 | 黑白名单策略 | P1 | 屏蔽低质量来源 | Assumption |
| 管理后台 | 异常列表 + 重试 | P1 | 若启用后台，处理 invalid/unreachable | Assumption |
| 目录体验 | 热门/推荐排序 | P1 | 按安装量与新鲜度 | Assumption |
| 数据质量 | 手工“强制验证” | P1 | 管理员对单条执行重验 | Assumption |
| 生态能力 | 用户收藏/订阅 | P2 | 需账号体系 | Assumption |
| 生态能力 | 评分与评论 | P2 | 社区治理成本高 | Assumption |
| 智能能力 | 个性化推荐 | P2 | 依赖行为数据沉淀 | Assumption |

## 5. 页面结构（按优先级）

### 5.1 P0 页面
1. 首页 `/`
   - 区块：Hero、搜索框、热门标签、接入说明入口
2. 目录页 `/roles`
   - 区块：搜索栏、筛选器、列表卡片、分页
3. 角色详情 `/roles/:id`
   - 区块：基本信息、安装命令、来源仓库、最近变更、状态标记
4. 仓库详情 `/repos/:owner/:repo`
   - 区块：仓库元信息、角色集合、更新时间、快速安装入口

### 5.2 P1 页面
1. 同步任务 `/admin/jobs`（可选后台启用时）
2. 规则管理 `/admin/rules`（可选后台启用时）
3. 异常处理 `/admin/issues`（可选后台启用时）

### 5.3 P2 页面
1. 我的收藏 `/me/favorites`
2. 角色评价 `/roles/:id/reviews`

## 6. 数据字段建议（Neon Postgres）

### 6.1 `role_records`（核心读模型）
| 字段 | 类型 | 说明 | 约束/索引 |
|---|---|---|---|
| id | uuid | 主键 | PK |
| role_name | text | 角色名（kebab-case） | idx(role_name) |
| display_name | text | 展示名 |  |
| description | text | 角色简介 | full-text |
| source_owner | text | GitHub owner | idx(source_owner, source_repo) |
| source_repo | text | GitHub repo | idx(source_owner, source_repo) |
| role_path | text | 角色目录路径 | unique(source_owner, source_repo, role_path) |
| source_ref | text | branch/tag/sha |  |
| status | text | discovered/verified/invalid/unreachable | idx(status) |
| folder_hash | text | 目录哈希 |  |
| install_count | bigint | 安装计数（可选） | idx(install_count) |
| tags | text[] | 标签 | gin(tags) |
| last_verified_at | timestamptz | 最近校验时间 | idx(last_verified_at desc) |
| created_at | timestamptz | 创建时间 |  |
| updated_at | timestamptz | 更新时间 | idx(updated_at desc) |

### 6.2 `ingest_events`（接入审计）
| 字段 | 类型 | 说明 | 约束/索引 |
|---|---|---|---|
| id | uuid | 主键 | PK |
| request_id | text | 请求ID | unique(request_id) |
| idempotency_key | text | 幂等键 | idx(idempotency_key) |
| sender | text | 发送方（cli 版本） |  |
| payload_count | int | 批量条数 |  |
| verify_result | text | pass/fail | idx(verify_result) |
| error_code | text | 错误码 | idx(error_code) |
| latency_ms | int | 接入耗时 | idx(latency_ms) |
| created_at | timestamptz | 时间 | idx(created_at desc) |

### 6.3 `sync_jobs`（任务执行）
| 字段 | 类型 | 说明 | 约束/索引 |
|---|---|---|---|
| id | uuid | 主键 | PK |
| job_type | text | normalize/resync/scheduled | idx(job_type) |
| target_role_id | uuid | 目标角色 | idx(target_role_id) |
| status | text | queued/running/success/failed | idx(status) |
| started_at | timestamptz | 开始时间 |  |
| finished_at | timestamptz | 结束时间 |  |
| retry_count | int | 重试次数 |  |
| error_message | text | 失败原因 |  |

### 6.4 `admin_rules`（P1）
| 字段 | 类型 | 说明 |
|---|---|---|
| id | uuid | 主键 |
| rule_type | text | blacklist/whitelist |
| pattern | text | owner/repo/path 规则 |
| enabled | bool | 是否生效 |
| reason | text | 规则说明 |
| updated_by | text | 操作人 |
| updated_at | timestamptz | 更新时间 |

## 7. 非功能需求（NFR）

### 7.1 性能
1. `role-repo find` 用户可感知延迟不因 ingest 增加（异步后台）
2. 目录查询 API：P95 < 300ms（单筛选条件）
3. 详情 API：P95 < 400ms

### 7.2 可靠性
1. Ingest API 可用性目标：月度 >= 99.9%
2. 异步任务失败自动重试（指数退避，最多 3 次）
3. 幂等保障：重复上报不产生重复记录

### 7.3 安全
1. Ingest API 必须验签 + 时间窗口校验（默认 5 分钟）
2. 若启用管理后台，必须 GitHub 登录鉴权（单管理员白名单）
3. 审计日志保留关键操作（重试、规则变更）

### 7.4 可观测性
1. 指标：ingest 成功率、normalize 成功率、平均处理时延、异常分布
2. 日志：请求链路 request_id 全链路关联
3. 告警：连续 10 分钟 ingest 失败率 > 5% 触发告警

### 7.5 数据一致性与时效
1. 新发现角色进入 `verified` 的目标：P50 < 5 分钟，P95 < 30 分钟
2. 仓库更新后的目录刷新目标：P95 < 60 分钟（MVP 定时）

证据标注：
- Assumption：时延目标需灰度期按真实负载调参
- Research：非阻塞与可观测性为前置设计硬要求

## 8. 里程碑与验收标准

| 里程碑 | 时间建议 | 目标 | 验收标准 | 下一责任人 |
|---|---|---|---|---|
| M0 需求冻结 | Week 1 | 确认范围与接口契约 | PRD、API 契约、数据模型评审通过 | PM |
| M1 接入链路打通 | Week 2-3 | CLI→Ingest→DB 可用 | `find` 不阻塞；接入成功率 >95%；幂等通过 | 后端 |
| M2 目录站点上线（内测） | Week 4-5 | Public 列表/搜索/详情可用 | P0 页面可访问；仅 verified 可见；核心查询性能达标 | 前端 |
| M3 管理能力上线（可选） | Week 6 | 异常处理闭环 | 若启用后台：可查看失败原因并重试；审计日志可追溯 | 全栈 |
| M4 Beta 发布 | Week 7 | 真实流量灰度 | 7 天内无 P0 故障；关键指标稳定 | PM+工程负责人 |

## 9. 验收清单（按 Quality Gate）

### 9.1 Objective Clarity
1. 是否实现“角色发现到可检索目录”的闭环：是（P0）
2. 是否明确不做项防止范围膨胀：是（见 1.2）

### 9.2 Measurable Metric
1. 接入成功率
2. 校验成功率
3. 数据新鲜度时延
4. 查询性能指标

### 9.3 Acceptance Criteria
1. CLI 主流程与 Ingest 失败解耦
2. 目录只展示 verified 数据
3. P0 不依赖管理端页面；异常可通过任务与日志闭环
4. 核心页面与 API 满足 P95 目标

### 9.4 Risk List
1. GitHub API 限流导致时效波动
2. 低质量来源导致目录污染
3. 早期数据量增长导致查询退化

### 9.5 Mitigation Plan
1. 限流：退避重试 + token 优先
2. 质量：状态机 + P1 黑白名单
3. 性能：索引治理 + 读写分离预案（P2）

## 10. 开放问题（需主控决策）
1. 已决策：MVP 不建设管理后台；若后续建设，采用 GitHub 登录 + 单管理员白名单。
2. 已决策：“仓库详情页”前置到 P0。
3. Beta 期是否需要公开“安装量”指标（涉及埋点可信度）？

## 11. Handoff 摘要
- Selected skill(s):
  - Primary: `pm`
  - Supporting: `prd-development`, `roadmap-planning`
- Process followed:
  1. 基于现有设计约束抽取产品目标与边界
  2. 按 PRD 结构输出范围/流程/IA/优先级
  3. 补充数据模型、NFR、里程碑与验收门禁
- Evidence level:
  - Research：来自本任务已有 brainstorming/design 结论
  - Assumption：容量阈值与时延目标需灰度校准
- Next owner and next action:
  - PM：组织 M0 评审并冻结 P0 范围
  - 工程负责人：在 M1 前确认 API 契约与表结构变更计划
