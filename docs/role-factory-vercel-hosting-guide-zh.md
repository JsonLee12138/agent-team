# Role Factory 托管与数据库开通指南（Vercel）

## 1. 目标架构
- 前端：Vercel 托管（Hobby 免费层可起步）
- 后端 API：可先同仓部署到 Vercel（Node/Go Serverless）或单独服务
- 数据库：通过 Vercel Marketplace 连接 Postgres（推荐 Neon）
- 可选缓存/限流：Upstash Redis（同样可通过 Marketplace）

## 2. 账号准备
1. 注册并登录 GitHub。
2. 注册并登录 Vercel（建议直接 GitHub OAuth 登录）。
3. 在 Vercel 中创建 Team（个人项目也可直接用个人空间）。

## 3. 导入项目到 Vercel
1. 打开 Vercel Dashboard，点击 `Add New...` -> `Project`。
2. 选择你的 GitHub 仓库并导入。
3. 在 `Framework Preset` 选择前端框架（role-factory 为 Vite）。
4. 配置 Root Directory（若只部署前端，填 `role-factory`）。
5. 先不改复杂构建参数，直接创建项目。

## 4. 创建数据库（Postgres）
1. 进入项目页面 -> `Storage` -> `Create Database`。
2. 选择 Postgres 供应商（推荐 Neon）。
3. 选择 Region（尽量靠近你的主要用户和 Vercel 部署区域）。
4. 创建后，确认环境变量已注入（典型如 `DATABASE_URL`）。

## 5. 配置环境变量
在 Vercel 项目 `Settings` -> `Environment Variables` 补齐：
- `DATABASE_URL`：数据库连接串
- `ROLE_HUB_API_BASE_URL`：前端请求后端的 base URL
- `INGEST_RATE_LIMIT_*`：如后端用到限流参数可在此配置

本地开发同步：
1. 在项目根或 `role-factory/.env.local` 填写同名变量。
2. 避免把 `.env*` 提交到仓库。

## 6. 数据库初始化
1. 执行迁移（按你项目的 migration 命令）。
2. 初始化最小种子数据（至少一组 repo/role/统计数据）。
3. 用只读查询验证表结构和数据可读。

## 7. 部署顺序（推荐）
1. 先部署后端 API（确保查询接口和 ingest 接口可用）。
2. 再部署前端（指向正式 API 地址）。
3. 运行 E2E 冒烟（首页、列表、详情、空状态、旧 payload 拒绝场景）。

## 8. 域名与 HTTPS
1. 先使用 `*.vercel.app` 做联调。
2. 联调通过后在 `Domains` 绑定自定义域名。
3. 按提示在 DNS 提供商处添加记录。
4. 等待证书自动签发并验证 HTTPS。

## 9. 监控与告警（上线前必须）
1. 打开 Vercel 的部署/函数日志。
2. 对后端接入以下关键指标：
   - ingest 4xx 按错误码分布
   - `UNSUPPORTED_PAYLOAD_VERSION` 比例
   - 查询接口 5xx 比例和 P95 延迟
3. 配置最小告警阈值（错误率和可用性）。

## 10. 免费层与成本控制建议
1. Vercel Hobby 可起步，但要关注函数调用和带宽。
2. 数据库套餐按供应商计费（Neon/Upstash 各自有免费层与配额）。
3. 先做基础限流，防止公开 ingest 被刷导致费用异常。
4. 每周查看一次 usage 面板，提前设置预算告警。

## 11. 你现在可以立刻做的动作
1. 完成 GitHub + Vercel 登录。
2. 在 Vercel 导入仓库并创建项目。
3. 开通 Postgres（Neon）并拿到 `DATABASE_URL`。
4. 把环境变量填入 Vercel。
5. 告诉我你已完成到哪一步，我继续给你下一步对接命令。
