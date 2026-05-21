# 博客导航重做方案

## Summary
- 按照用户提供的参考图重做博客站点导航，统一前台全站导航结构。
- 桌面端采用顶部横向导航布局：左侧站点标题，中间主导航，右侧搜索与主题切换。
- 移动端采用顶部轻量工具栏 + 底部固定 Tab 导航：顶部保留品牌、搜索、主题切换；底部展示首页、归档、关于、RSS 四个主入口。
- 所有导航元素需兼容亮色、暗色与 system 跟随系统模式，并保持现有主题切换逻辑可用。

## Current State Analysis
- `web/templates/layouts/base.html`
  - 当前前台所有页面共用同一套页头。
  - 现有结构为：品牌区、顶部 `nav.site-nav`、独立 `button.theme-toggle`。
  - 顶部导航内目前包含：首页、归档、关于、RSS、搜索图标。
  - 当前仅有一套导航 DOM，没有区分桌面端与移动端导航形态。
- `web/assets/style.css`
  - 已存在完整主题变量体系：`:root`、`html[data-theme="dark"]`、`html[data-theme="system"]` 与 `prefers-color-scheme` 回退。
  - 现有 `.site-header`、`.site-nav`、`.search-link`、`.theme-toggle` 已具备基础样式，但视觉与参考图差异较大。
  - 现有 `@media (max-width: 820px)` 仅将顶部导航折为多行，并未实现参考图中的底部固定导航。
  - 页面主体 `main.site-main` 和底部 `footer.site-footer` 目前未为移动端底部固定导航预留安全间距。
- `web/assets/frontend.js`
  - 主题切换逻辑已稳定可复用。
  - 当前 JS 只依赖 `[data-theme-toggle]`，不限制按钮出现位置，因此可在不改动核心逻辑的前提下复用。

## Proposed Changes

### 1. 调整基础模板导航结构
- 文件：`web/templates/layouts/base.html`
- 变更内容：
  - 将当前头部拆分为更贴近参考图的结构：
    - 品牌区 `site-brand`
    - 桌面端主导航 `site-nav site-nav--desktop`
    - 右侧操作区 `site-header__actions`，包含搜索按钮与主题切换按钮
  - 新增移动端底部导航 `mobile-tabbar`，放在 `main` 后、`footer` 前或后方，使用全站共享 DOM。
  - 为桌面端与移动端导航项统一补充更明确的 class，如 `site-nav__link`、`mobile-tabbar__link`。
  - 底部导航每个入口使用内联 SVG 图标，避免新增静态资源依赖。
  - 保持现有路由入口不变：
    - `/`
    - `/archives`
    - `/about`
    - `/rss.xml`
    - `/search`
  - 保持现有激活态判断方式（`.Path`）并扩展到移动端导航。
- 设计落点：
  - 桌面端不再把搜索放进中间主导航，而是放到右侧独立图标操作区。
  - 移动端顶部只保留品牌 + 搜索 + 主题切换，主导航转移到底部固定栏。
  - 若 `SiteSubtitle` 过长，不作为参考图主要视觉元素，计划在该轮中弱化或在小屏隐藏。

### 2. 重写导航相关样式与响应式断点
- 文件：`web/assets/style.css`
- 变更内容：
  - 新增导航专用视觉变量复用现有主题变量，不另起一套主题系统。
  - 重新设计以下模块样式：
    - `.site-header`
    - `.site-header__inner`
    - `.site-header__actions`
    - `.site-nav--desktop`
    - `.site-nav__link`
    - `.search-link`
    - `.theme-toggle`
    - `.mobile-tabbar`
    - `.mobile-tabbar__link`
    - `.mobile-tabbar__icon`
    - `.mobile-tabbar__label`
  - 调整头部高度、内边距、边框与背景，使其接近参考图的简洁分区效果。
  - 为桌面端导航加入居中排布、激活态强调、hover 状态和更轻的分隔感。
  - 为移动端底部导航实现：
    - 固定定位
    - 圆角容器
    - 阴影/边框
    - 当前项高亮
    - 安全区适配 `env(safe-area-inset-bottom)`
  - 调整 `site-main` 与 `site-footer` 的底部间距，避免内容被移动端底部导航遮挡。
  - 在小屏断点下隐藏桌面横向导航，显示底部 Tab；在大屏断点下隐藏底部 Tab。
  - 优化搜索与主题按钮尺寸，使其在桌面和移动端都接近参考图中的图标按钮感受。
- 兼容性原则：
  - 继续依赖 `var(--bg)`、`var(--surface)`、`var(--surface-strong)`、`var(--text)`、`var(--text-soft)`、`var(--border)`、`var(--shadow)` 等现有变量。
  - 暗色模式不单独复制整套样式，只通过变量驱动颜色变化，保证亮暗主题统一维护。

### 3. 最小化适配主题切换交互
- 文件：`web/assets/frontend.js`
- 变更内容：
  - 预期无需改动主题切换核心逻辑；沿用现有 `[data-theme-toggle]` 选择器。
  - 仅在模板改造后确认按钮仍唯一、仍可正确设置 `data-theme`、`aria-label` 和 `title`。
  - 若模板中需要移动按钮位置，不改逻辑，只确保 DOM 结构不影响初始化。
- 原因：
  - 当前主题逻辑已经支持 `light` / `dark` / `system` 生效态推导，满足本次导航重构需求。

## Assumptions & Decisions
- 本次范围仅覆盖“导航重做”，不调整首页欢迎卡片、背景网格、正文布局和其他页面内容模块。
- 导航改造作用于所有前台页面，因为它们共用 `web/templates/layouts/base.html`。
- 参考图中的移动端底部栏理解为小屏主导航，而不是仅首页特例。
- 搜索入口在移动端保留为顶部图标按钮，不放入底部 Tab，与参考图保持一致。
- RSS 在桌面端保留为文字主导航项，在移动端作为底部 Tab 一项。
- 主题切换继续采用当前“明/暗二态切换，system 作为初始化来源”的现有行为，不在本次顺带改成交替三态。
- 图标优先使用内联 SVG，避免新增图片资源和额外请求。

## Verification Steps
- 模板检查
  - 确认 `base.html` 中桌面端与移动端导航入口完整，且链接地址未变。
  - 确认当前页在顶部导航与底部导航都能正确显示激活态。
- 主题验证
  - 在 `light`、`dark`、`system + 系统浅色`、`system + 系统深色` 四种情况下检查：
    - 头部背景与边框对比度正常
    - 图标与文字可读性正常
    - 激活态和 hover 态在亮暗模式下都清晰
  - 确认搜索按钮与主题切换按钮在两种主题下无“白底白字”或“黑底黑字”问题。
- 响应式验证
  - 桌面宽屏下显示顶部横向导航，隐藏底部 Tab。
  - 移动端宽度下显示顶部工具栏与底部 Tab，隐藏桌面横向导航。
  - 确认底部 Tab 不遮挡 `main` 内容和 `footer`。
  - 确认 iPhone 类安全区场景下底部导航留白合理。
- 代码质量检查
  - 对变更后的 `base.html`、`style.css`、如有需要的 `frontend.js` 运行诊断检查。
  - 若有样式或模板语法问题，先修复再交付。
