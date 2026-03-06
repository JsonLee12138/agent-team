# Role Hub SEO 优化设计（分配给 frontend-architect）

## 背景
`role-hub` 当前仅在 `app/root.tsx` 提供全局 title/description，首页、角色详情页、仓库详情页缺少路由级 SEO 元信息与可抓取辅助页面，影响搜索引擎收录质量与结果展示。

## 目标
在不改变现有业务功能与视觉布局的前提下，完成技术 SEO 的最小可行增强：
- 页面级 metadata（title/description/canonical/open graph/twitter）
- 可抓取入口（`sitemap.xml`、`robots.txt`）
- 结构化数据（JSON-LD，至少覆盖首页与角色详情页）

## 架构与改造范围
1. 路由级元信息
- `app/routes/_index.tsx`：输出首页专用 metadata。
- `app/routes/roles.$name.tsx`：基于 role 数据输出动态 metadata（title、description、canonical、OG）。
- `app/routes/repos.$owner.$repo.tsx`：基于仓库数据输出动态 metadata。

2. 统一 SEO 工具层
- 新增 `app/utils/seo.ts`（或等价文件）封装 metadata 生成、canonical URL 拼接、JSON-LD 输出，避免各路由重复逻辑。

3. 抓取辅助路由
- 新增 `app/routes/sitemap[.]xml.tsx`：输出 XML sitemap（至少包含首页、可索引角色详情页、仓库详情页）。
- 新增 `app/routes/robots[.]txt.tsx`：声明 `Sitemap`、基础抓取策略。

4. 结构化数据
- 首页：`WebSite` / `CollectionPage`。
- 角色详情：`BreadcrumbList` + 面向角色的页面结构化描述（使用通用 schema 类型，避免过度语义承诺）。

## 非目标
- 不改动页面 UI 设计。
- 不引入新的后端接口。
- 不做 SEO 平台接入（如 GSC 自动提交）。

## 验收标准
- 首页、角色详情、仓库详情均有可验证的 route-level metadata。
- 存在可访问的 `/sitemap.xml` 与 `/robots.txt`。
- 页面源码中包含预期 JSON-LD。
- `role-hub` 目录下 `npm run lint && npm run test && npm run build` 通过。
