# role-hub QA/可观测性基线检查单（M7）

> 版本：v1.0 | 更新日期：2026-03-04
> 角色：QA/可观测性工程师（worker-6）
> 依赖：M1~M6 构建产物
> 主控决策：直接上线（无灰度）、安装量公开、Ingest 为公开接口+防滥用

---

## 目录

1. [测试矩阵](#1-测试矩阵cliingestnormalizequery)
2. [核心指标定义](#2-核心指标定义)
3. [告警规则与阈值](#3-告警规则与阈值)
4. [日志规范与链路追踪](#4-日志规范与链路追踪)
5. [回归测试检查单](#5-回归测试检查单)
6. [发布检查单](#6-发布检查单quality-gate)
7. [监控面板布局](#7-监控面板布局)
8. [事故响应手册](#8-事故响应手册playbook)

---

## 1. 测试矩阵（CLI→Ingest→Normalize→Query）

### 1.1 CLI 异步上报（M1）

| 编号 | 场景 | 预期结果 | 优先级 | 验收标准 |
|------|------|----------|--------|----------|
| T-CLI-01 | `find` 正常执行，Ingest API 可用 | `find` 返回结果，后台静默上报成功 | P0 | 命令耗时无显著回归（<50ms 增量） |
| T-CLI-02 | `find` 正常执行，Ingest API 不可用 | `find` 正常返回，上报静默失败 | P0 | 退出码 = 0；无用户可见错误 |
| T-CLI-03 | `find` 正常执行，Ingest API 超时（>5s） | `find` 不阻塞，上报超时放弃 | P0 | `find` 总耗时不受 Ingest 影响 |
| T-CLI-04 | 网络断开时执行 `find` | `find` 正常返回本地结果 | P0 | 无 panic/crash |
| T-CLI-05 | 上报批次含 >100 条候选项 | 分批上报或单批处理 | P1 | 不因批量大小导致 OOM 或超时 |
| T-CLI-06 | HMAC 密钥未配置 | 跳过上报，不影响 `find` | P0 | 日志记录 warn 级别 |

### 1.2 Ingest API（M2）

| 编号 | 场景 | 预期结果 | 优先级 | 验收标准 |
|------|------|----------|--------|----------|
| T-ING-01 | 合法 HMAC + 有效时间窗（5min）请求 | 202 Accepted，写入 `discovered` | P0 | 数据库记录存在 |
| T-ING-02 | HMAC 签名不匹配 | 401 Unauthorized | P0 | 审计日志记录拒绝原因 |
| T-ING-03 | 时间窗口过期（>5min） | 401 Unauthorized | P0 | 防重放攻击 |
| T-ING-04 | 重复 idempotency_key 请求 | 200 OK，不重复写入 | P0 | `role_records` 不产生重复行 |
| T-ING-05 | 批量 upsert（多条候选项） | 全部写入或全部失败（事务） | P0 | 原子性保证 |
| T-ING-06 | 请求体格式错误（JSON 畸形） | 400 Bad Request | P0 | 返回结构化错误信息 |
| T-ING-07 | `ingest_events` 审计记录完整 | 记录 request_id / latency_ms / result | P0 | 可查询审计日志 |
| T-ING-08 | 公开接口防滥用：速率限制 | 超限返回 429 Too Many Requests | P0 | 按 IP/sender 限流生效 |
| T-ING-09 | 公开接口防滥用：载荷大小限制 | 超限返回 413 Payload Too Large | P1 | 拒绝异常大请求 |
| T-ING-10 | 并发大量请求（负载测试） | 成功率 >95%，无数据丢失 | P0 | 压测报告确认 |

### 1.3 Normalize Worker（M4）

| 编号 | 场景 | 预期结果 | 优先级 | 验收标准 |
|------|------|----------|--------|----------|
| T-NRM-01 | 合法角色目录结构 | `discovered → verified` | P0 | 状态正确、`last_verified_at` 更新 |
| T-NRM-02 | 目录结构不合规 | `discovered → invalid` | P0 | 记录失败原因 |
| T-NRM-03 | GitHub 仓库不可达（404/私有） | `discovered → unreachable` | P0 | 记录错误码 |
| T-NRM-04 | GitHub API 限流（429） | 保持 `discovered`，退避重试 | P0 | 指数退避最多 3 次；不标记 invalid |
| T-NRM-05 | 已 `verified` 角色内容更新 | `folder_hash` 变更后重跑校验 | P0 | `updated_at` 刷新 |
| T-NRM-06 | 已 `verified` 角色仓库删除 | `verified → unreachable` | P1 | 定时同步检测到变更 |
| T-NRM-07 | 非法状态转换尝试（如 `invalid → verified` 直接跳转） | 按状态机规则拒绝或走正规路径 | P0 | 状态机约束生效 |
| T-NRM-08 | Worker 崩溃后恢复 | 未完成任务重新拉取 | P1 | 无任务丢失 |
| T-NRM-09 | `sync_jobs` 记录完整 | 每次执行有 start/finish/status/error | P0 | 可追溯执行历史 |

### 1.4 Query API（M5）

| 编号 | 场景 | 预期结果 | 优先级 | 验收标准 |
|------|------|----------|--------|----------|
| T-QRY-01 | 角色列表默认查询 | 仅返回 `verified` 记录 | P0 | 无 `discovered/invalid/unreachable` 记录泄露 |
| T-QRY-02 | 关键词搜索（名称/描述/标签） | 返回匹配结果 | P0 | 相关性合理 |
| T-QRY-03 | 筛选：按作用域/更新时间/来源 | 正确过滤 | P0 | 筛选结果与条件一致 |
| T-QRY-04 | 分页查询（offset/limit） | 分页正确，无重复/遗漏 | P0 | 边界值测试通过 |
| T-QRY-05 | 角色详情查询 | 返回完整元数据 | P0 | 包含安装命令、来源仓库 |
| T-QRY-06 | 仓库详情查询（聚合） | 按仓库聚合角色列表 | P0 | 聚合正确 |
| T-QRY-07 | 列表查询性能 | P95 < 300ms（单筛选条件） | P0 | 压测报告确认 |
| T-QRY-08 | 详情查询性能 | P95 < 400ms | P0 | 压测报告确认 |
| T-QRY-09 | 空结果查询 | 返回空数组，非 500 | P0 | 正确的空态响应 |
| T-QRY-10 | 安装量字段返回 | `install_count` 正确返回 | P0 | 安装量公开决策，数值准确 |

### 1.5 全链路 E2E 场景

| 编号 | 场景 | 覆盖模块 | 预期结果 | 优先级 |
|------|------|----------|----------|--------|
| T-E2E-01 | CLI find → Ingest → Normalize → Query 完整链路 | M1→M2→M4→M5 | 新角色在目标时延内可被查询到 | P0 |
| T-E2E-02 | 重复上报同一角色 | M1→M2 | 幂等处理，无重复记录 | P0 |
| T-E2E-03 | 不合规角色全链路 | M1→M2→M4→M5 | 写入 discovered → 标记 invalid → 不出现在列表 | P0 |
| T-E2E-04 | GitHub 限流下全链路表现 | M4→M5 | Normalize 退避重试；Query 返回已有数据 | P0 |
| T-E2E-05 | Ingest API 被滥用攻击 | M2 | 速率限制生效，正常流量不受影响 | P0 |
| T-E2E-06 | request_id 全链路追踪 | M1→M2→M4→M5 | 同一 request_id 贯穿所有日志 | P0 |

---

## 2. 核心指标定义

### 2.1 接入层（Ingest）

| 指标名称 | 计算方式 | 目标值 | 采集源 |
|----------|----------|--------|--------|
| `ingest_success_rate` | 成功请求数 / 总请求数 | ≥ 95% | `ingest_events.verify_result` |
| `ingest_latency_p50` | 请求处理耗时 P50 | < 100ms | `ingest_events.latency_ms` |
| `ingest_latency_p95` | 请求处理耗时 P95 | < 500ms | `ingest_events.latency_ms` |
| `ingest_latency_p99` | 请求处理耗时 P99 | < 1000ms | `ingest_events.latency_ms` |
| `ingest_error_distribution` | 按 error_code 分组统计 | 无主导错误码 | `ingest_events.error_code` |
| `ingest_ratelimit_rejections` | 429 响应数/分钟 | 监控趋势 | API 网关日志 |
| `ingest_request_volume` | 每分钟请求量 | 基线参考 | API 网关日志 |

### 2.2 标准化层（Normalize）

| 指标名称 | 计算方式 | 目标值 | 采集源 |
|----------|----------|--------|--------|
| `normalize_success_rate` | verified 数 / 已处理总数 | 监控趋势（无硬性目标） | `role_records.status` |
| `normalize_latency_p50` | 从 discovered 到终态的耗时 P50 | < 5 分钟 | `sync_jobs` 时间差 |
| `normalize_latency_p95` | 从 discovered 到终态的耗时 P95 | < 30 分钟 | `sync_jobs` 时间差 |
| `normalize_queue_depth` | queued 状态的 sync_jobs 数 | < 100 | `sync_jobs.status = 'queued'` |
| `normalize_retry_rate` | retry_count > 0 的比例 | < 10% | `sync_jobs.retry_count` |
| `normalize_failure_distribution` | 按失败原因分组 | 无主导错误类型 | `sync_jobs.error_message` |

### 2.3 查询层（Query）

| 指标名称 | 计算方式 | 目标值 | 采集源 |
|----------|----------|--------|--------|
| `query_list_latency_p95` | 列表查询 P95 | < 300ms | APM / 应用日志 |
| `query_detail_latency_p95` | 详情查询 P95 | < 400ms | APM / 应用日志 |
| `query_error_rate` | 5xx 响应占比 | < 0.1% | API 网关日志 |
| `query_request_volume` | 每分钟查询量 | 基线参考 | API 网关日志 |

### 2.4 数据质量

| 指标名称 | 计算方式 | 目标值 | 采集源 |
|----------|----------|--------|--------|
| `data_freshness_p50` | 新角色从发现到 verified 的 P50 | < 5 分钟 | `role_records` 时间差 |
| `data_freshness_p95` | 新角色从发现到 verified 的 P95 | < 30 分钟 | `role_records` 时间差 |
| `data_staleness_ratio` | 超过 24h 未刷新的 verified 记录占比 | < 20% | `role_records.updated_at` |
| `status_distribution` | 各状态占比 | verified 占主导 | `role_records.status` |
| `install_count_accuracy` | 安装量计数与实际偏差 | 无重复计算 | `role_records.install_count` |

### 2.5 防滥用

| 指标名称 | 计算方式 | 目标值 | 采集源 |
|----------|----------|--------|--------|
| `abuse_ratelimit_triggers` | 速率限制触发次数/小时 | 监控趋势 | API 网关 |
| `abuse_unique_senders` | 独立 sender 标识数/天 | 基线参考 | `ingest_events.sender` |
| `abuse_payload_rejections` | 因载荷过大被拒绝数/天 | 监控趋势 | API 网关 |

---

## 3. 告警规则与阈值

### 3.1 Critical（立即处理，影响核心功能）

| 告警名称 | 条件 | 持续时间 | 通知渠道 | 处置动作 |
|----------|------|----------|----------|----------|
| IngestHighFailureRate | `ingest_success_rate < 90%` | 连续 10 分钟 | Slack #alerts + PagerDuty | 检查 API 可用性、DB 连接 |
| QueryAPIDown | `query_error_rate > 5%` | 连续 5 分钟 | Slack #alerts + PagerDuty | 检查服务状态、数据库连接 |
| DatabaseConnectionFailed | DB 连接池耗尽或连接失败 | 立即 | Slack #alerts + PagerDuty | 检查 Neon Postgres 状态 |
| IngestAbuseSpike | `ingest_ratelimit_triggers > 1000/min` | 连续 5 分钟 | Slack #alerts + PagerDuty | 检查是否遭受攻击，临时加严限流 |

### 3.2 Warning（需关注，可能演化为 Critical）

| 告警名称 | 条件 | 持续时间 | 通知渠道 | 处置动作 |
|----------|------|----------|----------|----------|
| IngestElevatedFailureRate | `ingest_success_rate < 95%` | 连续 10 分钟 | Slack #alerts | 排查失败原因分布 |
| NormalizeQueueBacklog | `normalize_queue_depth > 100` | 连续 15 分钟 | Slack #alerts | 检查 Worker 状态、GitHub API 限流 |
| NormalizeHighRetryRate | `normalize_retry_rate > 20%` | 连续 30 分钟 | Slack #alerts | 检查 GitHub API 可用性 |
| QueryLatencyDegraded | `query_list_latency_p95 > 500ms` | 连续 10 分钟 | Slack #alerts | 检查查询计划、索引命中 |
| DataFreshnessDegraded | `data_freshness_p95 > 60 min` | 连续 30 分钟 | Slack #alerts | 检查 Normalize Worker 处理效率 |

### 3.3 Info（记录观察，无需立即行动）

| 告警名称 | 条件 | 通知渠道 | 说明 |
|----------|------|----------|------|
| IngestVolumeAnomaly | 请求量偏离基线 ±50% | Slack #monitoring | 流量模式变化 |
| NewSenderDetected | 出现新的 sender 标识 | 日志 | 新 CLI 版本或新用户接入 |
| StatusDistributionShift | invalid/unreachable 占比 >30% | Slack #monitoring | 数据质量风险 |

---

## 4. 日志规范与链路追踪

### 4.1 结构化日志字段（必填）

```json
{
  "timestamp": "2026-03-04T06:57:56.123Z",
  "level": "info|warn|error",
  "service": "ingest-api|normalize-worker|query-api",
  "request_id": "req_xxxxxxxx",
  "trace_id": "trace_xxxxxxxx",
  "message": "human readable message",
  "duration_ms": 42,
  "error_code": "HMAC_INVALID",
  "metadata": {}
}
```

### 4.2 链路追踪要求

| 层级 | request_id 来源 | 传播方式 |
|------|----------------|----------|
| CLI → Ingest | CLI 生成 UUID | HTTP Header `X-Request-ID` |
| Ingest → DB | 继承请求 request_id | `ingest_events.request_id` 字段 |
| Ingest → Normalize | 通过任务队列传递 | `sync_jobs` 关联 `ingest_event_id` |
| Normalize → role_records | 继承任务 context | 日志关联 sync_job_id |
| Query API | 独立 request_id | HTTP Header `X-Request-ID` |

### 4.3 日志级别规范

| 级别 | 使用场景 | 示例 |
|------|----------|------|
| ERROR | 影响功能的失败 | DB 写入失败、GitHub API 5xx |
| WARN | 可恢复的异常 | HMAC 校验失败、速率限制触发、重试中 |
| INFO | 正常业务流程 | 请求接收、状态转换完成、查询执行 |
| DEBUG | 调试信息（生产默认关闭） | 请求体详情、SQL 查询计划 |

---

## 5. 回归测试检查单

每次后端变更合入主分支前，需通过以下检查：

### 5.1 功能回归

- [ ] T-CLI-01~06：CLI 异步上报全场景通过
- [ ] T-ING-01~10：Ingest API 全场景通过
- [ ] T-NRM-01~09：Normalize Worker 全场景通过
- [ ] T-QRY-01~10：Query API 全场景通过
- [ ] T-E2E-01~06：全链路 E2E 场景通过

### 5.2 性能回归

- [ ] `find` 命令耗时无显著回归（基线 ±50ms）
- [ ] Ingest API P95 < 500ms
- [ ] Query 列表 P95 < 300ms
- [ ] Query 详情 P95 < 400ms

### 5.3 安全回归

- [ ] HMAC 验签拒绝非法请求
- [ ] 时间窗口校验拒绝过期请求
- [ ] 速率限制对异常流量生效
- [ ] 载荷大小限制生效
- [ ] 无 SQL 注入风险（参数化查询）
- [ ] 无敏感信息泄露（日志脱敏）

### 5.4 数据完整性

- [ ] 幂等键重复请求不产生重复记录
- [ ] 状态流转符合状态机规则
- [ ] 索引覆盖核心查询路径
- [ ] 迁移脚本可执行可回滚
- [ ] 安装量计数无重复累加

---

## 6. 发布检查单（Quality Gate）

> 主控决策：直接上线（无灰度），需在发布前加强验证深度。

### 6.1 发布前（Pre-Release）

#### 代码质量
- [ ] 所有 P0 测试用例通过（T-CLI/T-ING/T-NRM/T-QRY/T-E2E）
- [ ] 代码审查完成，无 P0/P1 缺陷遗留
- [ ] 无已知安全漏洞（依赖扫描通过）

#### 性能验证
- [ ] 压测报告：Ingest API 在目标 QPS 下成功率 >95%
- [ ] 压测报告：Query API P95 达标
- [ ] 数据库查询计划审查（无全表扫描）

#### 安全验证
- [ ] HMAC 验签端到端验证通过
- [ ] 速率限制配置就绪并测试通过
- [ ] 载荷大小限制配置就绪
- [ ] 公开接口无认证绕过路径

#### 可观测性就绪
- [ ] 监控面板部署并可访问
- [ ] 告警规则已配置并测试通过（模拟触发验证）
- [ ] 结构化日志格式验证通过
- [ ] request_id 全链路追踪验证通过

#### 数据准备
- [ ] 数据库迁移脚本在 staging 验证通过
- [ ] 回滚脚本在 staging 验证通过
- [ ] 初始数据（如有）导入验证通过

#### 运维准备
- [ ] 部署文档与操作手册就绪
- [ ] 回滚方案文档化（含数据库回滚）
- [ ] 值班人员确认与升级路径
- [ ] 发布窗口确认（建议工作日白天）

### 6.2 发布中（Release Execution）

- [ ] 数据库迁移执行成功
- [ ] 服务部署完成，健康检查通过
- [ ] 冒烟测试通过（T-E2E-01 核心链路）
- [ ] 监控面板指标开始采集

### 6.3 发布后（Post-Release）— 直接上线专项

> 因无灰度阶段，发布后需加强监控密度。

#### 第一小时（密集监控）
- [ ] 每 5 分钟检查 Ingest 成功率（目标 ≥95%）
- [ ] 每 5 分钟检查 Query API 错误率（目标 <0.1%）
- [ ] 确认无 Critical 告警触发
- [ ] 确认日志无异常错误模式
- [ ] 确认速率限制未误伤正常流量

#### 第一天（持续观察）
- [ ] Ingest 成功率稳定 ≥95%
- [ ] Query API P95 延迟达标
- [ ] Normalize 队列无积压
- [ ] 数据新鲜度 P95 < 30 分钟
- [ ] 安装量计数功能正常
- [ ] 无公开接口滥用迹象

#### 第一周（稳定性确认）
- [ ] 连续 7 天无 Critical 告警
- [ ] 核心指标无退化趋势
- [ ] 错误分布无新增主导错误码
- [ ] 流量模式符合预期基线
- [ ] 防滥用策略有效（无持续攻击穿透）

### 6.4 回滚判定标准

> 直接上线模式下，回滚决策需更迅速。

| 条件 | 触发时间 | 决策 | 执行人 |
|------|----------|------|--------|
| Ingest 成功率 <80% 持续 15 分钟 | 发布后 1h 内 | 立即回滚 | 值班工程师 |
| Query API 不可用（5xx >10%） | 发布后任何时间 | 立即回滚 | 值班工程师 |
| 数据库迁移导致数据不一致 | 发布后任何时间 | 立即回滚 + 数据修复 | 值班工程师 + DBA |
| P0 功能缺陷影响用户 | 发布后 24h 内 | 评估回滚或 hotfix | PM + 工程负责人 |
| 防滥用失效导致数据污染 | 发布后任何时间 | 临时下线 Ingest + 数据清理 | 值班工程师 |

---

## 7. 监控面板布局

### 7.1 总览面板（Overview Dashboard）

```
┌─────────────────────────────────────────────────────────┐
│  role-hub 总览                                          │
├─────────────┬─────────────┬─────────────┬──────────────┤
│ Ingest 成功率│ Normalize   │ Query P95   │ 数据新鲜度   │
│   ≥95%      │ 队列深度    │  <300ms     │  P50 <5min   │
│  (实时大数)  │ (实时大数)  │ (实时大数)   │ (实时大数)   │
├─────────────┴─────────────┴─────────────┴──────────────┤
│ [Ingest 成功率时序图 - 24h]                              │
├────────────────────────────┬────────────────────────────┤
│ [请求量时序图 - 24h]       │ [错误分布饼图]              │
├────────────────────────────┴────────────────────────────┤
│ [状态分布 - verified/invalid/unreachable/discovered]     │
├─────────────────────────────────────────────────────────┤
│ [速率限制触发趋势 - 24h]                                 │
└─────────────────────────────────────────────────────────┘
```

### 7.2 Ingest 详情面板

- Ingest 请求量（按 sender 分组）
- 延迟分布（P50/P95/P99 时序）
- 错误码分布（堆叠柱状图）
- 幂等命中率
- 速率限制触发详情（按 IP/sender）
- 载荷大小分布

### 7.3 Normalize 详情面板

- 任务处理速率（jobs/min）
- 状态转换分布（Sankey 图）
- 重试率与重试次数分布
- GitHub API 调用量与限流状态
- 队列深度趋势

### 7.4 Query 详情面板

- 请求量与延迟（按端点分组）
- 慢查询 Top 10
- 缓存命中率（如有）
- 安装量查询频率

---

## 8. 事故响应手册（Playbook）

### 8.1 Ingest 高失败率

**症状**：`IngestHighFailureRate` 告警触发

**排查步骤**：
1. 检查 `ingest_events` 最近失败记录的 `error_code` 分布
2. 若多为 `DB_ERROR` → 检查 Neon Postgres 状态
3. 若多为 `HMAC_INVALID` → 检查是否存在异常请求来源（防滥用）
4. 若多为 `PAYLOAD_ERROR` → 检查是否有 CLI 版本异常
5. 若多为 `RATE_LIMITED` → 评估限流阈值是否过严

**处置**：
- DB 问题：联系 Neon 支持或检查连接池配置
- 防滥用误伤：临时调整限流策略
- CLI 兼容性：通知 CLI 工程师（worker-2）

### 8.2 Normalize 队列积压

**症状**：`NormalizeQueueBacklog` 告警触发

**排查步骤**：
1. 检查 `sync_jobs` 中 `status = 'running'` 的任务是否卡住
2. 检查 GitHub API 剩余配额（`X-RateLimit-Remaining`）
3. 检查 Worker 进程是否正常运行
4. 检查是否有批量新数据涌入

**处置**：
- GitHub 限流：等待配额恢复，确认退避策略正常
- Worker 异常：重启 Worker 进程
- 流量突增：评估是否需要临时增加 Worker 实例

### 8.3 Query API 延迟退化

**症状**：`QueryLatencyDegraded` 告警触发

**排查步骤**：
1. 检查慢查询日志，识别退化的端点
2. `EXPLAIN ANALYZE` 分析查询计划
3. 检查数据库连接池使用率
4. 检查是否有索引缺失或数据量突增

**处置**：
- 索引缺失：添加缺失索引
- 数据量增长：评估分页约束是否合理
- 连接池：调整连接池大小

### 8.4 公开接口遭受攻击

**症状**：`IngestAbuseSpike` 告警触发

**排查步骤**：
1. 检查速率限制日志，识别攻击来源（IP/sender）
2. 评估攻击流量规模与模式
3. 检查是否有数据污染（大量垃圾记录）

**处置**：
- 轻度：加严单 IP 限流阈值
- 中度：添加 IP 黑名单，清理污染数据
- 重度：临时关闭 Ingest 公开访问，通知主控决策

---

## 附录 A：指标采集技术选型建议

| 组件 | 推荐方案 | 备选 | 说明 |
|------|----------|------|------|
| 指标采集 | Vercel Analytics + 自定义指标 | Prometheus + Grafana | 与 Vercel 部署一致 |
| 日志 | 结构化 JSON → Vercel Logs | Loki | 支持 request_id 检索 |
| 告警 | Vercel 监控 + Slack Webhook | PagerDuty | 按团队规模选择 |
| APM | Vercel Speed Insights | Sentry | 前端性能监控 |
| 压测 | k6 / Artillery | Locust | 脚本化压测 |

## 附录 B：测试用例与指标编号索引

| 前缀 | 模块 | 范围 |
|------|------|------|
| T-CLI | CLI 异步上报 | T-CLI-01 ~ T-CLI-06 |
| T-ING | Ingest API | T-ING-01 ~ T-ING-10 |
| T-NRM | Normalize Worker | T-NRM-01 ~ T-NRM-09 |
| T-QRY | Query API | T-QRY-01 ~ T-QRY-10 |
| T-E2E | 全链路 E2E | T-E2E-01 ~ T-E2E-06 |
