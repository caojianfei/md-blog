# 博客前端重构计划

## Summary

- 目标：按照 `promt.md` 将现有 Go SSR 博客前台重构为「High-end Minimalist Tech Aesthetics」风格，覆盖全局壳层、首页、列表页、文章详情页、归档页、关于页，并补齐深浅色切换、文章 TOC、代码块增强等关键体验。
- 范围：仅重构前台 SSR 页面与其所需的最小后端渲染支持；不改后台管理 UI；不新增需要后台维护的新配置项；不引入假数据。
- 核心决策：
  - 首页 Hero 右侧采用 ASCII 技术面板，而不是动态提交网格。
  - `Current Stack`、`GitHub Stats`、`Project Mirror` 不新增伪造内容或后台配置，关于页改为“Dashboard 化展示现有真实数据”。
  - 深浅色模式采用「手动切换 + localStorage 持久化 + system 回退」。

## Current State Analysis

### 当前前台结构

- 模板入口集中在 `web/templates/layouts/base.html`、`web/templates/pages/*.html`、`web/templates/partials/*.html`。
- 全站样式目前只有一个文件：`web/assets/style.css`。
- 模板由 `internal/view/renderer.go` 从嵌入式文件系统加载，不需要额外前端构建流程。
- 资源嵌入定义在 `embedded.go`，`web/assets/*` 下新增文件会自动打包进二进制。

### 当前实现与设计要求的差距

- 全局视觉仍是通用“圆角卡片博客”风格，圆角较大、阴影较重，与 `promt.md` 要求的硬朗、边框分层、编辑器气质不一致。
- `base.html` 没有主题切换控件、状态栏式信息区、技术感导航结构，也没有加载 `Inter` / `JetBrains Mono`。
- 首页 `home.html` 只有简单 Hero + 卡片列表，没有左对齐极简 Hero、ASCII 面板、流式列表、IDE 行高亮交互。
- 列表卡片 `partials/article_card.html` 仍是传统卡片，不符合“左日期 / 中标题 / 右标签”的列表信息密度。
- 详情页 `pages/article.html` 还没有窄栏阅读区、右侧半透明 TOC、代码块语言标签 / 复制 / 行号、Markdown 元素精修。
- 关于页 `pages/about.html` 只是单一文章容器，没有 Dashboard 化编排。
- `internal/handler/web/handler.go` 当前 `PageData` 不包含文章 TOC、页面统计摘要等前台增强数据。

### 已有能力可复用

- `internal/service/markdown/service.go` 已启用 `parser.WithAutoHeadingID()`，文章标题天然可生成锚点 ID。
- 同一 Markdown 服务的 `RenderResult` 已能返回 `Headings`，可直接为文章 TOC 提供数据来源。
- `goldmark-highlighting` 与 Chroma 依赖已存在，模块源码确认支持 `WithLineNumbers`、`WithClasses`、`WithWrapperRenderer`，可用于输出 GitHub 风格代码块骨架，而不是完全走前端二次解析。

## Proposed Changes

### 1. 重做全局页面壳层与设计系统

#### 文件

- `web/templates/layouts/base.html`
- `web/templates/partials/seo_head.html`
- `web/assets/style.css`
- 新增 `web/assets/frontend.js`

#### 变更内容

- 在 `base.html` 中建立新的技术博客壳层：
  - `html` / `body` 支持 `data-theme`。
  - 顶部导航改成更克制的技术风头部，加入主题切换按钮。
  - 页脚改成状态栏式信息区，展示站点页脚文案、GitHub 链接或 ICP（存在则显示）。
  - 加载 `Inter` 和 `JetBrains Mono` 字体资源。
  - 引入 `frontend.js` 负责主题切换、代码复制、TOC 激活、移动端交互。
- 在 `seo_head.html` 保持现有 SEO 输出，不新增会改变语义的结构；仅确保主题与字体接入不污染 SEO 部分。
- 在 `style.css` 中重构为新的前台设计系统：
  - 颜色令牌改为接近 `promt.md` 的白 / GitHub 深色语义。
  - 弱化投影、强化 1px 边框，圆角统一压到 `4px-6px`。
  - 标题 / 正文 / 元数据 / 代码字体分层明确。
  - 全局 hover / focus / transition 统一 `0.2s ease-in-out`。
  - 为主题、页头、网格背景、列表行 hover、按钮、表单、分页、空状态等建立统一组件样式。
- 新增 `frontend.js`：
  - 首屏读取 `localStorage` 的前台主题设置，没有时回退到系统或 `Site.ThemeDefault`。
  - 绑定主题切换按钮。
  - 为代码块复制按钮、TOC 高亮、可能的移动端目录折叠提供最小交互。

#### 原因

- 当前视觉体系和 `promt.md` 的目标风格不是“微调”关系，而是需要一个新的基础设计系统承托所有页面。

### 2. 首页与所有列表页统一为“技术流式信息架构”

#### 文件

- `web/templates/pages/home.html`
- `web/templates/partials/article_card.html`
- `web/templates/pages/search.html`
- `web/templates/pages/category.html`
- `web/templates/pages/tag.html`
- `web/templates/pages/archives.html`

#### 变更内容

- 首页 `home.html`：
  - Hero 改成左右双栏。
  - 左侧保留站点 Hero 标题、描述、搜索入口和快速导航。
  - 右侧改为 ASCII 技术面板，视觉上模拟终端 / 编辑器欢迎区。
  - 下方文章列表改成流式列表，不再用大卡片。
  - 侧栏保留但重排为更轻的技术信息块，优先放分类、标签、归档入口。
- `article_card.html`：
  - 变为单行 / 双行混合的信息条布局。
  - 左列显示等宽日期，中列显示标题与摘要，右列显示标签与分类。
  - hover 时做 IDE 行选中式背景高亮，而不是升起卡片。
- 搜索、分类、标签页：
  - 统一列表页头部结构，包含标题、说明、搜索框或当前筛选条件。
  - 沿用流式列表与统一分页。
  - 空状态文字和容器风格统一为设计系统中的“终端空状态”表达。
- 归档页：
  - 从纯 chip 列表改成更适合阅读的归档时间轴 / 月份列表。
  - 保持仅使用现有 `Archives` 数据，不新增月内文章明细查询。

#### 原因

- `promt.md` 的首页核心要求不是“卡片更好看”，而是将博客入口做成技术系统首页；同时其他列表页必须跟随统一视觉语言，否则首页和内页会割裂。

### 3. 详情页补齐阅读体验、TOC 与 GitHub 风格代码块

#### 文件

- `internal/handler/web/handler.go`
- `internal/service/markdown/service.go`
- `web/templates/pages/article.html`
- 可新增 `web/templates/partials/article_toc.html`
- `web/assets/style.css`
- `web/assets/frontend.js`

#### 变更内容

- 扩展 `PageData`，增加文章 TOC 所需的标题集合字段；命名以当前 `markdown.Heading` 为基础，避免重复建模。
- 在 `Handler.Article` 中继续使用已存储的 `Article.HTMLContent` 作为正文渲染内容，同时对 `Article.Content` 调用 Markdown 服务提取 `Headings`，供详情页 TOC 使用。
- 如关于页复用目录逻辑，可在 `Handler.About` 中同样保留 headings 能力，但不作为首要验收项。
- 调整 `markdown.Service` 的高亮配置：
  - 使用 Chroma classes 输出，便于 `style.css` 统一适配深浅色主题。
  - 打开代码行号能力。
  - 使用 wrapper renderer 或等效方式给代码块补上可识别的语言标签挂点，供模板 / JS 渲染复制按钮与标题条。
- 重写 `pages/article.html`：
  - 主阅读栏最大宽度固定在 `760px` 左右。
  - 桌面端布局为“内容栏 + 右侧悬浮 TOC”，移动端目录折叠到正文上方。
  - 顶部元信息、封面、标签、上一篇 / 下一篇导航统一进新视觉体系。
- 在 `style.css` 中精修 Markdown 输出：
  - 响应式表格横向滚动。
  - Blockquote 左侧 4px 品牌色竖线。
  - 图片改为细边框 + 微圆角。
  - 代码块做 GitHub Light / Dark 风格，包含标题栏、语言标签、复制按钮、行号区、可滚动内容区。
- 在 `frontend.js` 中补齐：
  - 复制代码按钮。
  - 基于标题锚点的滚动监听，驱动 TOC 当前项高亮。

#### 原因

- 详情页是 `promt.md` 中要求最具体的页面；现有模板和样式不具备通过简单 CSS 覆盖达成目标的条件，必须同步做数据上下文和 Markdown 输出增强。

### 4. 将关于页改造成“基于真实站点数据的 Dashboard”

#### 文件

- `internal/handler/web/handler.go`
- `web/templates/pages/about.html`
- `web/assets/style.css`

#### 变更内容

- 在不新增后台配置的前提下，把关于页从单栏 Markdown 页面改造成 Dashboard 编排：
  - 主区域仍展示 `AboutHTML`，保留用户可在后台维护的正文能力。
  - 侧边 / 辅助区域增加基于现有真实数据的统计和索引块，例如：
    - 已发布文章总数。
    - 分类总数、标签总数。
    - 归档月份数。
    - 站点主题默认值、存储驱动、GitHub 链接（存在时）。
  - 视觉表达对齐 `promt.md` 的 Tech Radar / Dashboard 气质，但不伪造 GitHub Stats 或项目 Star 数据。
- `Handler.About` 增加这些聚合数据的装配逻辑，尽量直接复用当前仓库可得数据：
  - 文章总量可通过现有列表服务的 total 获取。
  - 分类 / 标签总数直接来自 `baseData()` 已返回的列表长度。
  - 归档月份数复用 `Article.Archives()`。

#### 原因

- 用户明确要求首版不引入新配置或假数据，因此关于页应转向“真实数据的可视化排版”，而不是做空壳模块。

### 5. 统一细节与无障碍 / 响应式收口

#### 文件

- `web/assets/style.css`
- `web/assets/frontend.js`
- 涉及所有更新后的模板

#### 变更内容

- 保证所有可点击区域具备明确 hover / focus 状态和 `cursor` 反馈。
- 处理窄屏断点：
  - 顶部导航折行或压缩。
  - 首页 Hero 双栏回落为单栏。
  - 文章页 TOC 改为折叠块。
  - 列表行在移动端重排为日期在上、标题在中、标签在下。
- 控制所有边框、分隔线、间距和空状态风格统一，避免“首页像新站、内页像旧站”。
- 若有必要，为 `prefers-reduced-motion` 降低滚动与过渡效果。

#### 原因

- 这次重构的风险不在单页，而在跨页面一致性；必须把响应式和细节态度写进计划，而不是最后再补。

## Assumptions & Decisions

- 前台仍维持 Go SSR 模板架构，不迁移到新的前端框架。
- 不修改后台管理系统的配置模型来支持新的首页 / 关于页模块内容。
- 不引入伪造的 GitHub 统计、仓库 Star 或开发环境版本号；只展示现有真实数据与站点信息。
- 文章详情页 TOC 以 Markdown 标题为准，不额外做全文索引或阅读进度服务端持久化。
- 代码块增强优先走现有 Goldmark + Chroma 输出能力，避免在前端重写语法高亮。
- 如执行阶段发现个别页面依赖现有 inline style，统一收敛回 `style.css`，不保留模板内联布局样式。

## Verification Steps

- 模板与后端：
  - 执行 `go test ./...`，确保后端改动未破坏编译与已有逻辑。
  - 如新增 `PageData` 字段或 Markdown 配置，确认 `internal/view/renderer.go` 下模板渲染正常。
- 视觉与交互：
  - 首页检查 Hero、ASCII 面板、流式列表、分页、侧栏在浅色 / 深色下都可读。
  - 搜索、分类、标签、归档页检查统一视觉是否成立，空状态是否一致。
  - 文章页检查 TOC 锚点跳转、滚动高亮、代码复制、表格滚动、引用块、图片、上一篇 / 下一篇是否正常。
  - 关于页检查 Dashboard 化布局是否只使用真实数据，且 Markdown 正文仍可正常渲染。
- 响应式：
  - 至少检查 `375px`、`768px`、`1024px`、`1440px` 四档布局。
  - 确认无横向滚动，TOC / 列表 / 头部在移动端可用。
- 主题：
  - 检查首次加载、手动切换、刷新后持久化、跟随系统回退逻辑。
- 质量收口：
  - 对修改过的 Go / 模板文件运行诊断，修复新增语法或模板错误。
