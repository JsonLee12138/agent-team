# Role Hub UI 方案文档

## 1. 概述与设计定位

本方案旨在为 `role-hub` 提供明确的 UI/UX 指导。
- **目标受众**：开发者、AI 编排工程师、Agent 平台使用者。
- **UI 风格参考**：[Loot Drop](https://www.loot-drop.io/) —— 现代、极简、高对比度的“独立开发者 / Tech”美学，强调清晰的网格、微妙的玻璃态（Glassmorphism）或实色描边（Brutalism/Neo-Brutalism）边框，信息密度高但不显拥挤。
- **功能结构参考**：[Skills.sh](https://skills.sh/) —— 面向工具/包管理的目录结构，侧重于快速检索、分类过滤以及“一键复制”安装命令的无缝体验。

## 2. 视觉方向 (Visual Direction)

### 2.1 色彩系统 (Color Palette)
采用暗色模式为主（Dark Mode First）或极简高对比度亮色，营造 Tech & Developer 氛围。
- **主背景 (Background)**：深灰/极暗色（如 `#09090B`）或 纯白底色带细线网格。
- **表面/卡片 (Surface/Card)**：微弱的色差衬托（如 `#18181B`）辅以 `1px` 实色边框（如 `#27272A`）和微妙悬浮阴影。
- **主品牌色 (Primary/Accent)**：高饱和度的点缀色（如 霓虹绿 `#A8D8C8`、赛博紫 `#B8A9C9` 或 亮黄色 `#FFF3C4`），用于主要 CTA 按钮、Hover 状态和高亮标签。
- **文字颜色 (Text)**：主文本 `#FAFAFA` (Dark) / `#111827` (Light)，次文本/描述 `#A1A1AA` (Dark) / `#6B7280` (Light)。
- **语义色 (Semantic)**：
  - Verified (已验证)：青绿色 `#10B981` (图标与徽章)
  - Invalid (无效)：玫瑰红 `#F43F5E`
  - Warning/Syncing：琥珀黄 `#F59E0B`

### 2.2 字体排版 (Typography)
- **字体家族**：
  - 标题/UI：Inter, system-ui, sans-serif（冷峻、现代、清晰）。
  - 代码/命令：JetBrains Mono, Fira Code（开发者原生感）。
- **字号阶梯**：
  - Hero 标题：`48px` / `56px` (加粗 ExtraBold, 紧凑字间距)
  - 模块标题：`24px` / `32px` (SemiBold)
  - 正文/卡片内容：`14px` / `16px` (Regular/Medium，行高 `1.6`)
  - 辅助说明/标签：`12px` (Medium)

### 2.3 组件语义 (Component Semantics)
- **卡片 (Cards)**：Bento Grid (便当盒) 布局。直角或小圆角（`8px - 12px`），带 `1px` 边框，Hover 时边框颜色变为品牌主色或产生位移阴影。
- **按钮 (Buttons)**：
  - 主 CTA：主色背景，深色文字，无边框。
  - 次 CTA：透明背景，浅色边框，Hover 时背景反色。
  - 复制按钮 (Copy/Install)：固定在卡片或代码块右上角/右侧，带交互反馈（点击后变成 Check 图标）。
- **标签 (Tags/Badges)**：胶囊状或方块状，带轻微背景色透明度，用于展示 Role 的领域、框架、状态。

### 2.4 动效 (Motion)
- **核心原则**：克制、快速、反馈清晰（Snappy & Responsive）。
- **悬浮 (Hover)**：列表卡片的细微上浮（`-2px`）或边框发光（Glow effect，`200ms` ease-out）。
- **复制反馈**：点击安装命令时，图标变化伴随轻微缩放 (`scale(0.95) -> scale(1)`) 和短暂的 Toast 提示。
- **加载状态**：Skeleton 骨架屏优先，避免大面积菊花图（Spinner）。

---

## 3. 页面级结构与信息层级 (IA & Layout)

功能结构全面对标 `skills.sh` 的心智模型，核心分为：侧边栏（导航/过滤）+ 主内容区（列表/详情）。

### 3.1 首页 / 目录页 (`/roles`)
作为主入口，兼具落地页与目录的双重职责。
- **Hero 区 (顶部)**：
  - 极简居中标题（如 "Find the perfect role for your Agent"）。
  - 巨大的全局搜索框（支持 `Cmd+K` 快捷键聚焦）。
  - 安装说明的快速示例（"How to install?" snippet）。
- **主体布局 (Left Sidebar + Right Grid)**：
  - **左侧边栏 (Sidebar - Filter)**：
    - 分类目录（All Roles, Trending, Recently Added）。
    - Faceted Filters (复选框)：根据 Framework, Scope, Status (Verified) 过滤。
    - 作者/仓库维度筛选。
  - **右侧内容区 (Role Grid)**：
    - 顶部工具栏：Sort by (Relevance, Installs, Newest)、View Toggle (List / Grid)。
    - **角色卡片 (Role Card)**：
      - 第一层级：角色名称（`role_name`）、Verified 徽章。
      - 第二层级：一句话简介（`description`）。
      - 第三层级：标签（`tags`）、更新时间（`updated_at`）、来源仓库（Owner/Repo）。
      - **快速操作**：卡片底部或悬浮展示“一键复制安装命令”（`agent-team role install xxx`）。

### 3.2 角色详情页 (`/roles/:id`)
面向开发者评估和安装的沉浸式页面。
- **顶部面包屑**：Home > Roles > `role_name`
- **头部 (Header)**：
  - 大标题 `display_name` + `@owner/repo` 链接。
  - 安装命令行区块（大且醒目，右侧带 Copy 按钮）。
  - 状态标记（Verified）、版本/更新时间。
- **内容双栏布局**：
  - **左侧主干 (Main Content)**：
    - 角色 Readme 或详细描述（Markdown 渲染）。
    - Configuration (环境变量、所需权限)。
    - Example Usage (使用示例代码)。
  - **右侧侧边栏 (Meta Sidebar)**：
    - Author (仓库作者头像与链接)。
    - Source Repo (GitHub 链接、Stars 数、License)。
    - Tags (关键词标签)。
    - Install count (若有)。

### 3.3 仓库详情页 (`/repos/:owner/:repo`)
以组织/仓库视角聚合展现。
- **Header**：仓库名称、描述、GitHub 跳转链接、全仓库整体安装/导入说明。
- **内容区**：
  - 列出该仓库下包含的所有 Roles（复用目录页的卡片组件）。
  - 标注整个仓库最后一次 Sync (同步) 的状态与时间。

---

## 4. 关键交互与响应式策略 (Interactions & RWD)

### 4.1 关键交互
- **即时搜索 (Instant Search)**：输入内容时列表实时过滤（Debounce `300ms`），无整页刷新。
- **无限滚动 / 加载更多 (Pagination)**：采用底部点击 "Load More" 或滚动到底部触发，保持页面流畅。
- **骨架屏加载 (Skeleton Loading)**：在发起查询与图片加载前使用对应的几何占位符。
- **一键复制 (One-Click Copy)**：整个平台的原子级体验核心，任何 Role 的展示都伴随 Install Code 的复制功能。

### 4.2 响应式适配 (Responsive Strategy)
- **Desktop (`>1024px`)**：左侧固定 Filter 栏（`~280px`），右侧卡片 `3 列` 网格。
- **Tablet (`768px - 1024px`)**：侧边栏可折叠或变成顶部横向筛选（Pills），卡片 `2 列`。
- **Mobile (`<768px`)**：
  - 搜索框常驻顶部。
  - 筛选栏收入 Bottom Sheet (底部抽屉) 按钮中（"Filters (3)"）。
  - 角色卡片全宽 `1 列`。
  - 详情页侧边栏元数据（Meta）移动到主干内容之上或最底部。

---

## 5. 可交付清单 (Deliverable Checklist for Design Phase)

在下一阶段（Pencil.pen 或 Figma 绘制），需要产出以下高保真设计/组件：

1. **Design System & Tokens**
   - [ ] Color Palette (CSS Variables，匹配 UnoCSS)
   - [ ] Typography Scale
   - [ ] Component Kit (Button, Input, Tag, Badge, Card, Checkbox)
2. **Page Designs**
   - [ ] `/roles` - 首页/目录（Desktop + Mobile）
   - [ ] `/roles/:id` - 角色详情页（Desktop + Mobile）
   - [ ] `/repos/:owner/:repo` - 仓库聚合页
   - [ ] Empty States / Not Found (搜索无结果态)
3. **Micro-interactions (状态展示)**
   - [ ] Card Hover & Copy Success 态
   - [ ] Skeleton Loading 骨架屏
