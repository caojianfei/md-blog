# 后台 Markdown 预览改造计划

## Summary

- 目标：解决后台文章编辑页中 Markdown 标题、分割线及其他块级元素渲染不稳定的问题，并将预览区切换为成熟的 Vue 预览组件方案。
- 决策：采用“成熟预览组件 + GitHub 风格主题”路线，不再继续扩充 `ArticleEdit.vue` 内的手写 `.prose` 样式。
- 视觉要求：浅色模式对应 GitHub 浅色风格，深色模式对应 GitHub 深色风格。
- 范围：仅改造后台文章编辑页的 Markdown 预览区域，不改动当前编辑输入区和工具栏插入逻辑。

## Current State Analysis

### 当前实现

- `web/admin/src/views/ArticleEdit.vue`
  - 当前直接在页面内 `import MarkdownIt from "markdown-it"` 并通过 `computed(() => md.render(editor.content || ""))` 生成 HTML。
  - 预览区使用 `v-html="previewHtml"` 输出 HTML。
  - 代码高亮依赖 `highlight.js` 和 `github-dark.css`。
  - 预览样式全部写在同文件 `<style>` 中，当前仅覆盖 `pre`、`code`、`img`、`a`、`blockquote`、`table` 等少量元素。
- `web/admin/src/components/Layout.vue`
  - 后台主题通过 `document.documentElement` 的 `data-theme` 控制，值来自 `localStorage` 的 `admin-theme`，支持 `light` / `dark` / `system`。
- `web/admin/package.json`
  - 当前未引入专门的 Markdown 预览组件，也未引入成熟 Markdown 样式体系。

### 已确认问题

- 当前后台预览是“解析器 + `v-html` + 零散手写样式”模式，样式覆盖面不完整，导致标题、分割线、列表等块级元素容易继续出现显示异常。
- 当前深色代码主题固定为 `highlight.js/styles/github-dark.css`，并未与后台实际浅色/深色主题状态统一切换。
- 前台站点的 `web/assets/style.css` 中对 `.article-content` 的样式也较少，不能直接作为后台稳定预览方案复用。
- 服务端正式渲染使用 `internal/service/markdown/service.go` 中的 `goldmark`，而后台预览使用 `markdown-it`，语义本身存在一定差异；本次重点先解决“预览稳定性与成熟样式”问题，不在本次范围内统一前后端 Markdown 引擎。

## Proposed Changes

### 1. 引入成熟 Markdown 预览组件

- 文件：`web/admin/package.json`
- 变更：
  - 新增 Vue 3 兼容的 Markdown 预览组件依赖，计划使用 `md-editor-v3`。
- 原因：
  - 该方案可直接替换当前 `v-html + 手写样式` 的预览模式，减少对标题、分割线、表格、列表、引用等块级元素的逐项补样式工作。
  - 支持预览模式与主题配置，适合当前仅改造预览区、不替换输入区的需求。
- 实现方式：
  - 安装 `md-editor-v3`。
  - 使用其预览组件而非完整编辑器组件，保持现有 textarea 编辑体验不变。

### 2. 新建独立的后台 Markdown 预览组件

- 文件：`web/admin/src/components/MarkdownPreview.vue`（新增）
- 变更：
  - 封装统一的预览组件，接收 `content` 作为输入。
  - 在组件内引入 `md-editor-v3` 的 GitHub 预览主题及代码高亮主题资源。
  - 在组件内根据后台当前主题状态决定预览主题：
    - `data-theme="light"` 时使用浅色 GitHub 风格；
    - `data-theme="dark"` 时使用深色 GitHub 风格；
    - `system` 时根据 `prefers-color-scheme: dark` 自动跟随系统。
- 原因：
  - 将预览逻辑从 `ArticleEdit.vue` 中拆出，避免页面继续承载 Markdown 解析、主题判断和样式资源导入。
  - 让后续预览相关问题集中在单一组件修复。
- 实现方式：
  - 组件内部统一处理：
    - Markdown 文本输入；
    - 主题监听；
    - 预览区容器尺寸和滚动行为；
    - GitHub 样式切换。
  - 主题状态通过读取 `document.documentElement.dataset.theme` 并监听系统主题变化实现，不要求本次抽离全局主题状态。

### 3. 用新组件替换文章编辑页中的手写预览区

- 文件：`web/admin/src/views/ArticleEdit.vue`
- 变更：
  - 移除页面内的 `markdown-it`、`highlight.js`、`previewHtml` 和现有大段 `.prose` 预览样式。
  - 保留现有编辑器、光标同步、工具栏插入、图片上传、分屏/预览模式切换逻辑。
  - 将预览区域替换为 `MarkdownPreview` 组件，传入 `editor.content`。
- 原因：
  - 让 `ArticleEdit.vue` 回归“编辑页容器”职责，不再负责 Markdown 解析和样式细节。
  - 降低后续继续出现标题、分割线、表格、列表等块级元素样式漏配的风险。
- 实现方式：
  - 保持现有布局结构和 `previewMode` 行为不变，只替换预览区内部实现。
  - 保证分屏模式下边框、滚动区域、高度占满等布局行为与当前一致。

### 4. 清理旧的预览样式和单点依赖

- 文件：`web/admin/src/views/ArticleEdit.vue`
- 变更：
  - 删除当前内联 `.prose` 规则，避免与成熟预览组件的内置样式冲突。
  - 删除仅服务于旧预览方案的 `highlight.js` 相关引入。
- 原因：
  - 避免双份样式叠加，出现“GitHub 风格 + 旧 prose 风格”互相覆盖的问题。
  - 降低页面耦合度。
- 实现方式：
  - 只保留页面级布局样式；Markdown 内容样式全部收敛到新组件或组件依赖提供的主题样式中。

## Assumptions & Decisions

- 已锁定方案：使用成熟预览组件，不继续扩充手写 `.prose`。
- 已锁定视觉：后台预览采用 GitHub 风格，浅色/深色分别与后台主题同步。
- 本次不替换当前 textarea 编辑器，不引入所见即所得编辑。
- 本次不处理服务端 `goldmark` 与后台预览解析器的完全一致性问题；若后续出现语义差异，再单独规划“预览与正式渲染统一”改造。
- 本次不改前台站点 `web/assets/style.css` 的文章页视觉，仅处理后台编辑器预览稳定性。

## Verification Steps

### 功能验证

- 在 `web/admin/src/views/ArticleEdit.vue` 中输入并预览以下内容，确认渲染正确：
  - `#`、`##`、`###` 标题层级；
  - 紧邻段落上下的 `---` 分割线；
  - 表格；
  - 无序列表、有序列表、任务列表；
  - 引用块；
  - 代码块与行内代码；
  - 图片与链接。

### 主题验证

- 后台主题切换到 `light` 时，预览呈现 GitHub 浅色风格。
- 后台主题切换到 `dark` 时，预览呈现 GitHub 深色风格。
- 后台主题切换到 `system` 时，预览跟随系统主题变化。

### 回归验证

- 分屏、仅编辑、仅预览三种模式切换正常。
- 工具栏插入标题、表格、分割线等 Markdown 语法后，预览立即更新。
- 页面构建通过，且新组件未引入明显的布局溢出、滚动异常或暗色样式错乱。
