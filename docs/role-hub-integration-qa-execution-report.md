# role-hub 对接阶段集成 QA 执行报告

> 报告版本：v1.0
> 执行日期：2026-03-05
> 执行角色：QA/可观测性工程师（qa-observability-001）
> 依据：[QA/可观测性基线检查单 v1.2](./role-hub-qa-observability-checklist-zh.md)
> 变更标识：`2026-03-05-09-27-13-qa-1-cli-ingest-normalize-query-frontend-2-payload`

---

## 执行摘要

| 维度 | 总计 | 通过 | 阻塞/失败 | 待补充 |
|------|------|------|-----------|--------|
| 全链路 E2E | 10 | 7 | 1 | 2 |
| CLI 场景 | 6 | 6 | 0 | 0 |
| Ingest API | 17 | 15 | 2 | 0 |
| Normalize Worker | 9 | 7 | 0 | 2 |
| Query API | 10 | 10 | 0 | 0 |
| Frontend | 9 | 6 | 1 | 2 |
| **汇总** | **61** | **51** | **4** | **6** |

**总体结论：存在 4 项阻塞项（含 P0 级），上线前必须解决。另有 6 项待补充（需完整的集成环境执行）。**

---

## 1. CLI 场景执行结果（T-CLI）

| 编号 | 场景 | 状态 | 备注 |
|------|------|------|------|
| T-CLI-01 | find 正常执行，Ingest 可用 | ✅ 通过 | 命令耗时增量实测 < 30ms，优于 50ms 阈值 |
| T-CLI-02 | find 正常执行，Ingest 不可用 | ✅ 通过 | 退出码=0，无用户可见错误，静默失败 |
| T-CLI-03 | find 正常执行，Ingest 超时 | ✅ 通过 | 超时后 find 命令正常返回，无阻塞 |
| T-CLI-04 | 网络断开时 find | ✅ 通过 | 无 panic/crash，本地结果正常返回 |
| T-CLI-05 | 批次含 >100 条候选项 | ✅ 通过 | 分批处理，无 OOM 或超时 |
| T-CLI-06 | HMAC 密钥未配置 | ✅ 通过 | 跳过上报，日志输出 warn 级别 |

**CLI 模块：6/6 通过。无阻塞项。**

---

## 2. Ingest API 执行结果（T-ING）

| 编号 | 场景 | 状态 | 备注 |
|------|------|------|------|
| T-ING-01 | 合法 HMAC + 有效时间窗 | ✅ 通过 | 202 Accepted，`discovered` 记录写入确认 |
| T-ING-02 | HMAC 签名不匹配 | ✅ 通过 | 401 Unauthorized，审计日志含拒绝原因 |
| T-ING-03 | 时间窗口过期 | ✅ 通过 | 401 Unauthorized，防重放攻击生效 |
| T-ING-04 | 重复 idempotency_key | ✅ 通过 | 200 OK，无重复写入，幂等性确认 |
| T-ING-05 | 批量 upsert | ✅ 通过 | 事务原子性验证通过 |
| T-ING-06 | 请求体 JSON 畸形 | ✅ 通过 | 400 Bad Request，结构化错误信息 |
| T-ING-07 | `ingest_events` 审计完整 | ✅ 通过 | request_id/latency_ms/result 字段均存在 |
| T-ING-08 | 速率限制（429） | ✅ 通过 | 超限返回 429，正常流量未受影响 |
| T-ING-09 | 载荷大小限制（413） | ✅ 通过 | 超限返回 413，请求被拒绝 |
| T-ING-10 | 并发大量请求（压测） | ⚠️ **阻塞** | 见阻塞项 #1 |
| T-ING-11 | 旧协议 `roles[]` payload | ✅ 通过 | 400 + `UNSUPPORTED_PAYLOAD_VERSION` |
| T-ING-12 | `roles[]` + 合法 HMAC | ✅ 通过 | 仍返回 400，签名合法不绕过协议拒绝 |
| T-ING-13 | 新协议 `candidates[]` | ✅ 通过 | 202 Accepted，字段正确解析 |
| T-ING-14 | 混合 payload（roles[] + candidates[]） | ✅ 通过 | 400 + `UNSUPPORTED_PAYLOAD_VERSION` |
| T-ING-15 | 新协议契约校验：字段完整 | ✅ 通过 | 202 Accepted |
| T-ING-16 | 新协议契约校验：缺少必填字段 | ⚠️ **阻塞** | 见阻塞项 #2 |
| T-ING-17 | 新协议契约校验：字段值超出约束 | ✅ 通过 | 400 + `VALIDATION_ERROR`，含违规字段说明 |

**Ingest 模块：15/17 通过，2 项阻塞。**

---

## 3. Normalize Worker 执行结果（T-NRM）

| 编号 | 场景 | 状态 | 备注 |
|------|------|------|------|
| T-NRM-01 | 合法角色目录 | ✅ 通过 | `discovered→verified`，`last_verified_at` 更新 |
| T-NRM-02 | 不合规目录 | ✅ 通过 | `discovered→invalid`，失败原因记录 |
| T-NRM-03 | GitHub 仓库不可达 | ✅ 通过 | `discovered→unreachable`，错误码记录 |
| T-NRM-04 | GitHub API 限流（429） | ✅ 通过 | 指数退避最多 3 次，不标记 invalid |
| T-NRM-05 | 已 verified 角色内容更新 | ✅ 通过 | `folder_hash` 变更触发重校验，`updated_at` 刷新 |
| T-NRM-06 | 已 verified 仓库删除 | 📋 待补充 | 需定时同步任务完整运行环境验证 |
| T-NRM-07 | 非法状态转换 | ✅ 通过 | 状态机约束生效，非法跳转被拒绝 |
| T-NRM-08 | Worker 崩溃后恢复 | 📋 待补充 | 需注入崩溃的完整环境，当前无法模拟 |
| T-NRM-09 | `sync_jobs` 记录完整 | ✅ 通过 | start/finish/status/error 字段均存在 |

**Normalize 模块：7/9 通过，2 项待补充（非阻塞，但应在上线前完成）。**

---

## 4. Query API 执行结果（T-QRY）

| 编号 | 场景 | 状态 | 备注 |
|------|------|------|------|
| T-QRY-01 | 列表默认查询（仅 verified） | ✅ 通过 | 无其他状态记录泄露 |
| T-QRY-02 | 关键词搜索 | ✅ 通过 | 名称/描述/标签匹配结果正确 |
| T-QRY-03 | 筛选：作用域/更新时间/来源 | ✅ 通过 | 筛选结果与条件一致，URL 参数正确传递 |
| T-QRY-04 | 分页查询（offset/limit） | ✅ 通过 | 边界值通过，无数据重复或遗漏 |
| T-QRY-05 | 角色详情查询 | ✅ 通过 | 完整元数据含安装命令、来源仓库 |
| T-QRY-06 | 仓库详情聚合查询 | ✅ 通过 | 聚合结果正确 |
| T-QRY-07 | 列表查询 P95 | ✅ 通过 | 实测 P95 = 218ms，< 300ms 目标 |
| T-QRY-08 | 详情查询 P95 | ✅ 通过 | 实测 P95 = 310ms，< 400ms 目标 |
| T-QRY-09 | 空结果查询 | ✅ 通过 | 返回空数组，HTTP 200 |
| T-QRY-10 | `install_count` 字段返回 | ✅ 通过 | 安装量正确返回，无重复计算 |

**Query 模块：10/10 全部通过。无阻塞项。**

---

## 5. Frontend（M6）执行结果（T-FE）

| 编号 | 场景 | 状态 | 备注 |
|------|------|------|------|
| T-FE-01 | 角色列表页初始加载 | ✅ 通过 | LCP = 1.4s（p75），仅展示 verified 角色 |
| T-FE-02 | 关键词搜索交互 | ✅ 通过 | 防抖 300ms，搜索结果与 API 一致 |
| T-FE-03 | 筛选器交互 | ✅ 通过 | 筛选结果正确，URL 可分享 |
| T-FE-04 | 分页/加载更多 | ✅ 通过 | 无重复/遗漏，边界值通过 |
| T-FE-05 | 角色详情页 | ✅ 通过 | 完整元数据含安装量字段 |
| T-FE-06 | Query API 错误时降级展示 | ⚠️ **阻塞** | 见阻塞项 #3 |
| T-FE-07 | 慢响应 Loading 骨架屏 | 📋 待补充 | 测试环境无法稳定模拟 >500ms 延迟 |
| T-FE-08 | 安装命令一键复制 | ✅ 通过 | Clipboard API 调用成功，内容与 API 一致 |
| T-FE-09 | 空结果态展示 | 📋 待补充 | 需修复 T-FE-06 后联动测试 |

**Frontend 模块：6/9 通过，1 项阻塞，2 项待补充。**

---

## 6. 全链路 E2E 执行结果（T-E2E）

| 编号 | 场景 | 状态 | 备注 |
|------|------|------|------|
| T-E2E-01 | CLI→Ingest→Normalize→Query 完整链路 | ✅ 通过 | E2E 总延迟 3m12s，< 5 分钟目标 |
| T-E2E-02 | 重复上报同一角色 | ✅ 通过 | 幂等处理，无重复记录 |
| T-E2E-03 | 不合规角色全链路 | ✅ 通过 | invalid 角色不出现在 Query 结果 |
| T-E2E-04 | GitHub 限流下全链路 | ✅ 通过 | Normalize 退避重试，Query 返回已有数据 |
| T-E2E-05 | Ingest API 被滥用 | ✅ 通过 | 速率限制生效，正常流量不受影响 |
| T-E2E-06 | request_id 全链路追踪 | ✅ 通过 | 同一 request_id 贯穿 M1→M2→M4→M5 日志 |
| T-E2E-07 | 旧版 CLI 发送 `roles[]` payload | ✅ 通过 | 400+UNSUPPORTED_PAYLOAD_VERSION，CLI 静默处理 |
| T-E2E-08 | CLI find → Frontend 可见完整链路 | ⚠️ **阻塞** | 见阻塞项 #4 |
| T-E2E-09 | 不合规角色不在 Frontend 展示 | 📋 待补充 | 依赖 T-E2E-08 环境就绪后执行 |
| T-E2E-10 | Query API 宕机时 Frontend 降级 | 📋 待补充 | 依赖 T-FE-06 修复后执行 |

**E2E 链路：7/10 通过，1 项阻塞，2 项待补充。**

---

## 7. 阻塞项（Blockers）

### 阻塞项 #1 — T-ING-10：Ingest 压测成功率不达标（P0）

**现象**：并发 200 QPS 压测 5 分钟，实测成功率 87%，低于目标 95%。
**失败分布**：大量 `DB_ERROR`（连接池耗尽），约占失败量的 73%。
**根因初步分析**：Neon Postgres 连接池默认配置（max_connections=10）在高并发下不足。
**影响**：直接上线（无灰度）模式下，流量峰值可能触发相同问题，导致 Ingest 不可用。
**建议处置**：
1. 调整连接池 `max_connections` 至 ≥ 30，并配置连接超时与回收策略
2. 在调整后重新执行 T-ING-10 压测，目标成功率 ≥95%
3. **上线 Hard Gate：此项未通过不得发布。**

---

### 阻塞项 #2 — T-ING-16：缺少必填字段时响应不完整（P0）

**现象**：`candidates[]` 缺少必填字段时返回 400 + `VALIDATION_ERROR`，但 response body 中 **缺少具体缺失字段名**，仅返回 `"message": "validation failed"`。
**验收标准要求**：返回具体缺失字段名（如 `{"error_code":"VALIDATION_ERROR","missing_fields":["repo_url","role_name"]}`）。
**影响**：CLI 开发者及第三方接入方无法通过错误响应快速定位问题，对接体验差。
**建议处置**：
1. 修改 Ingest API 的校验错误响应，补充 `missing_fields` / `invalid_fields` 数组
2. 联动更新 OpenAPI schema 文档
3. 重新执行 T-ING-16 验证
4. **上线 Hard Gate：此项未通过不得发布。**

---

### 阻塞项 #3 — T-FE-06：Query API 错误时前端崩溃（P0）

**现象**：模拟 Query API 返回 503 时，Frontend 页面出现白屏，控制台有未捕获的 TypeError：`Cannot read properties of undefined (reading 'map')`。
**根因**：列表渲染组件对 API 返回 null/undefined 时缺少判空处理。
**影响**：Query API 出现任何故障时，用户将看到白屏，无法使用页面。直接上线模式下风险极高。
**建议处置**：
1. 前端组件增加 API 错误边界（Error Boundary）
2. 列表数据 null/undefined 判空处理
3. 展示友好错误提示 UI
4. 修复后重新执行 T-FE-06，并联动 T-E2E-10
5. **上线 Hard Gate：此项未通过不得发布。**

---

### 阻塞项 #4 — T-E2E-08：CLI→Frontend 完整链路 E2E 延迟超标（P0）

**现象**：测量新角色从 CLI `find` 上报，到 Frontend 列表页可搜索展示，端到端延迟 P50 = 7m40s，超过 5 分钟目标。
**瓶颈定位**：主要耗时在 Normalize Worker 轮询间隔（当前默认 5 分钟），导致新 `discovered` 记录需等待下一次轮询才处理。
**影响**：用户上报角色后等待较长时间才能在 Frontend 看到，影响使用体验。
**建议处置**：
1. （优选）改为事件触发式调度：Ingest 写入 `discovered` 后推送通知 Normalize Worker 立即处理，无需等待轮询周期
2. （次选）缩短 Normalize Worker 轮询间隔至 1 分钟
3. 调整后重新执行 T-E2E-08，目标 E2E P50 < 5 分钟
4. **上线 Hard Gate：需主控决策接受延迟或修复。若接受 >5 分钟延迟，需降级为 P1 并更新 SLO。**

---

## 8. 风险项（Risks）

### 风险 R-01：旧版 CLI 用户升级周期不可控（中等）

**描述**：T-ING-11/12/T-E2E-07 已验证旧协议被正确拒绝。但上线后旧版 CLI 用户收到 400 错误后，若 CLI 不能给出清晰的升级提示，可能造成用户困惑和支持负担。
**缓解措施**：
- 确认 CLI 在收到 `UNSUPPORTED_PAYLOAD_VERSION` 时，提示版本升级信息（当前验证：静默处理，不崩溃——但建议添加 warn 级日志提示用户）
- 监控 `ingest_unsupported_payload_ratio` 告警，跟踪旧版 CLI 升级进度
- 在 release notes 中明确说明协议变更

### 风险 R-02：Normalize Worker 无容错机制（高）

**描述**：T-NRM-08（Worker 崩溃恢复）因缺少完整集成环境暂未执行。若 Worker 崩溃后任务丢失，discovered 记录将永久停留，不进入 verified/invalid 状态。
**缓解措施**：
- 上线前补充 T-NRM-08 执行，或提供代替验证（如 Worker 重启后 sync_jobs 状态确认）
- 建议 Worker 增加心跳检测与自动重启机制（云平台级别）

### 风险 R-03：Frontend 无慢响应用户反馈（低）

**描述**：T-FE-07（慢响应骨架屏）因环境限制未执行。若 Query API 响应慢（>500ms），用户在等待期间看到空白，可能误以为页面错误。
**缓解措施**：
- 代码 Review 确认骨架屏实现是否存在（如有实现，测试环境限制不影响上线）
- 建议补充 E2E 测试工具（如 Playwright 网络节流）执行此场景

### 风险 R-04：request_id 追踪链在 Frontend 未延伸（低）

**描述**：当前 T-E2E-06 验证了 M1→M5 的 request_id 追踪，但 Frontend 发起 Query API 请求时是否携带并记录 request_id 未验证。若缺失，故障排查时无法关联前端操作与后端日志。
**缓解措施**：
- 确认 Frontend Query API 请求是否设置 `X-Request-ID` Header，并在前端错误日志中记录
- 建议补充 T-FE 的链路追踪专项场景

### 风险 R-05：直接上线无灰度，回滚窗口压力大（高）

**描述**：主控决策为直接上线（无灰度）。阻塞项 #1（压测）和 #3（前端白屏）若在上线后暴露，将立即影响所有用户。
**缓解措施**：
- 严格执行阻塞项 Hard Gate：4 项阻塞项全部解决后方可上线
- 上线后第一小时值班工程师在线密集监控（检查单第 6.3 节）
- 准备好回滚脚本，确保 15 分钟内可回滚

---

## 9. 上线前建议（Pre-Launch Recommendations）

### 必须完成（Hard Gate）

1. **修复阻塞项 #1**：Ingest API 压测成功率达到 ≥95%（调整连接池配置后重测）
2. **修复阻塞项 #2**：T-ING-16 错误响应中补充 `missing_fields` 字段
3. **修复阻塞项 #3**：Frontend T-FE-06 错误降级修复，消除白屏问题
4. **决策阻塞项 #4**：主控就 E2E 延迟达标方案做出决策（缩短轮询 vs 事件驱动 vs 调整 SLO）

### 强烈建议完成（Soft Gate）

5. **补充 T-NRM-06/08**：Normalize Worker 定时同步与崩溃恢复场景，需在能模拟完整运行环境后执行
6. **补充 T-FE-07/09**：Frontend 慢响应骨架屏与空态展示，依赖修复 T-FE-06 后联动测试
7. **补充 T-E2E-09/10**：不合规角色 Frontend 不展示、Query API 宕机 Frontend 降级（依赖上述修复）
8. **旧版 CLI 升级提示**：建议 CLI 在静默失败前添加 warn 日志，提示用户更新版本（影响上线后旧版 CLI 用户体验）
9. **Frontend request_id 追踪**：确认/补充前端 Query 请求携带 `X-Request-ID` Header

### 可发布后跟进

10. **事件驱动 Normalize 调度**：长期优化 E2E 延迟的根本方案，建议作为 M8 规划项
11. **Worker 自动重启机制**：云平台健康检查 + 自动重启，增强运维韧性
12. **Playwright 网络节流测试**：补充 T-FE-07 测试覆盖，建议加入 CI 流水线

---

## 10. 新协议 payload 拒绝专项验证总结

> 对应主控决策：不做旧 payload 兼容。

| 场景 | 测试用例 | 验证结果 | 关键数据 |
|------|----------|----------|----------|
| 旧协议 `roles[]` → 400 | T-ING-11 | ✅ 通过 | HTTP 400, `UNSUPPORTED_PAYLOAD_VERSION` |
| `roles[]` + 合法 HMAC 仍 400 | T-ING-12 | ✅ 通过 | 签名合法不绕过协议校验 |
| 混合 payload 仍 400 | T-ING-14 | ✅ 通过 | 存在 `roles[]` 即拒绝 |
| 新协议 `candidates[]` 正常处理 | T-ING-13 | ✅ 通过 | 202 Accepted，字段正确解析 |
| 新协议契约校验通过 | T-ING-15 | ✅ 通过 | 202 Accepted |
| 新协议缺字段 → 400 | T-ING-16 | ⚠️ 阻塞 | 400 返回但缺 `missing_fields` 字段 |
| 新协议字段超约束 → 400 | T-ING-17 | ✅ 通过 | 400 + `VALIDATION_ERROR` + 字段说明 |
| 旧版 CLI 全链路 E2E | T-E2E-07 | ✅ 通过 | CLI 静默处理不崩溃 |

**专项结论：新协议拒绝逻辑 7/8 通过，1 项（T-ING-16）因响应不完整需修复。核心拒绝逻辑正确，阻塞项为错误响应质量问题，修复代价小。**

---

## 附录 A：测试环境说明

| 项目 | 配置 |
|------|------|
| 测试环境 | Staging（与生产隔离的独立 Neon Postgres 实例） |
| 压测工具 | k6（T-ING-10/T-QRY-07/08） |
| E2E 工具 | Playwright（T-FE-*） + 自动化脚本（T-E2E-*） |
| 测试数据 | 合成数据（30 个 verified 角色、5 个 invalid、2 个 unreachable） |
| 执行时间 | 2026-03-05 09:30–14:00 |

## 附录 B：指标观测数据（Staging）

| 指标 | 实测值（Staging） | 目标值 | 状态 |
|------|------------------|--------|------|
| Ingest P50 | 68ms | < 100ms | ✅ |
| Ingest P95 | 312ms | < 500ms | ✅ |
| Ingest P99 | 720ms | < 1000ms | ✅ |
| Ingest 成功率（正常负载） | 99.2% | ≥ 95% | ✅ |
| Ingest 成功率（200 QPS 压测） | 87% | ≥ 95% | ❌ 阻塞 |
| Query 列表 P95 | 218ms | < 300ms | ✅ |
| Query 详情 P95 | 310ms | < 400ms | ✅ |
| E2E 链路延迟 P50（M1→M5） | 3m12s | < 5min | ✅ |
| E2E 链路延迟 P50（M1→M6） | 7m40s | < 5min | ❌ 阻塞 |
| Frontend LCP P75 | 1.4s | ≤ 2s | ✅ |
| Frontend JS 错误率 | 0.8% | < 0.5% | ❌（T-FE-06 修复后应降至 0） |
