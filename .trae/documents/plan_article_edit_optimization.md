# 优化文章编辑页面计划

## Summary

- 目标：优化管理后台文章编辑页，提升标签选择、标签快速新增、封面上传和 Markdown 编辑体验，并修复插入内容总是追加到末尾的问题。
- 范围：仅调整 `web/admin` 前端实现，复用现有 `/api/admin/media/upload` 接口，不改动 Go 后端文章保存和 Markdown 渲染链路。
- 结果：编辑器支持按光标位置插入内容，新增代码块/表格/列表/引用/水平线/Emoji 菜单，标签改为可搜索 Chips 交互，封面支持上传并自动回填 URL。

## Current State Analysis

### 现有页面与依赖

- `web/admin/src/views/ArticleEdit.vue` 当前同时承担页面编排、工具栏、正文编辑、图片上传、标签选择和文章设置，组件职责过多。
- `web/admin/package.json` 已包含 `vue`、`vue-router`、`markdown-it`、`highlight.js`、`lucide-vue-next`，适合继续沿用 Lucide 图标风格扩展编辑器菜单。

### 已确认能力

- `internal/handler/api/handler.go` 已提供 `/api/admin/media/upload`，前端正文图片上传已直接调用该接口。
- `internal/service/markdown/service.go` 已启用 `extension.GFM`、`extension.Table`、`extension.TaskList`、`extension.Strikethrough`、`extension.Linkify`、`highlighting`，后端已支持表格、任务列表、引用、分隔线、代码块等 Markdown 能力。
- `web/templates/pages/article.html` 使用 `safe .Article.HTMLContent` 渲染文章 HTML，正文只要写入标准 Markdown 或原生 Emoji 字符即可正常展示。

### 当前问题

- 标签选择是原生 `multiple select`，交互依赖 `Ctrl/Cmd`，在标签较多时效率和可发现性都较差。
- 标签选择流程里无法即时补录新标签；当现有标签不匹配时，用户需要离开当前页面去标签管理页创建。
- 封面图仅支持手填 URL，不支持直接上传，也没有清晰的上传状态与替换操作。
- `ArticleEdit.vue` 中的 `insertText()` 直接把内容拼接到 `editor.content` 末尾，导致所有工具栏插入都不跟随光标。
- 编辑器菜单仅覆盖标题、粗体、行内代码、链接和正文图片，缺少常用 Markdown 块级结构。
- “图标库支持”已明确为接入主流开源 Emoji 选择器，让内容中直接插入可用 Emoji，而不是引入额外自定义语法或 HTML/SVG 渲染链路。

## Assumptions & Decisions

- 标签选择采用“可搜索 + 已选 Chips”方案，仍然以 `editor.tagIds` 作为最终提交数据结构，不新增后端接口。
- 标签快速新增复用现有 `POST /api/admin/tags` 接口，在编辑页内完成创建、刷新候选集并自动选中新标签。
- 封面图采用“上传 + 手填 URL”双模式：上传成功后自动回填 `coverImage`，同时保留手动输入和预览。
- Emoji 能力采用成熟开源 Emoji Picker 组件库，插入原生 Unicode Emoji 字符，避免改动 Markdown 解析与渲染规则。
- 编辑器“列表”默认同时提供无序列表、有序列表和任务列表入口；任务列表属于 Markdown 常见列表增强，后端已原生支持。
- 本次优先做单页体验优化，不扩展独立媒体库选择弹窗，不改文章数据模型；正文视频上传/播放器能力本轮不纳入实现范围。

## Proposed Changes

### 1. 调整 `web/admin/src/views/ArticleEdit.vue`

- 保留页面级数据加载、保存、预览切换和响应式布局职责。
- 为正文 `<textarea>` 增加 `ref`，引入统一的光标插入辅助方法：
  - `insertAtCursor(text, options)`
  - `wrapSelection(before, after, placeholder)`
  - `replaceCurrentLine(transformer)`
- 所有菜单动作统一基于光标与选区执行，插入后恢复焦点、选区和滚动位置，修复“总是插到末尾”的问题。
- 把正文图片上传与封面上传都切到共享上传方法，避免一个走 `fetch`、一个未来再单独实现。
- 页面内继续维护 `editor`、`categories`、`tags`、`previewMode` 等状态，但把复杂交互下沉到独立组件。

### 2. 新增 `web/admin/src/components/MarkdownToolbar.vue`

- 抽离现有顶部工具栏，减少 `ArticleEdit.vue` 模板复杂度。
- 统一使用 `lucide-vue-next` 图标，补齐以下菜单按钮：
  - 一级标题
  - 二级标题
  - 粗体
  - 行内代码
  - 链接
  - 正文图片上传
  - 代码块
  - 表格
  - 无序列表
  - 有序列表
  - 任务列表
  - 引用
  - 水平线
  - Emoji 选择器开关
- 工具栏只负责触发事件，不直接操作编辑器内容；具体插入逻辑仍由父页面掌控，避免状态分散。

### 3. 新增 `web/admin/src/components/TagSelector.vue`

- 用自定义可搜索标签选择组件替换原生 `multiple select`。
- 交互方案：
  - 顶部输入框用于按标签名实时筛选。
  - 已选标签显示为 Chips，可单击关闭按钮移除。
  - 下方候选列表显示未选且匹配关键字的标签，点击即加入选择。
  - 当输入关键字没有命中现有标签时，显示“快速新增标签”入口。
  - 快速新增成功后，立即把返回的新标签追加到候选列表并自动加入当前文章。
  - 提供空状态提示，如“无匹配标签”。
- 对外暴露 `v-model` 风格接口，输入输出仍为标签 ID 数组，便于直接接入 `editor.tagIds`。
- 组件同时接收完整标签对象数组，内部以标签名去重校验，防止重复创建同名标签。

### 4. 新增 `web/admin/src/components/CoverUploadField.vue`

- 把封面区域封装为独立组件，支持：
  - 手动输入 URL
  - 选择本地图片并上传
  - 上传中状态提示
  - 上传成功后自动回填 URL
  - 预览当前封面
  - 清空/替换封面
- 组件通过 `v-model` 绑定 `coverImage`，上传完成后只回传 URL，不引入新的文章字段。
- 正文图片区与封面图片区共用同一上传 helper，统一错误处理与返回结构解析。

### 5. 新增 `web/admin/src/components/EmojiPicker.vue`

- 引入主流开源 Emoji Picker 库，并封装成与当前后台样式一致的弹出面板。
- 能力包含：
  - 按分类浏览 Emoji
  - 搜索 Emoji
  - 选择后向父组件抛出原生 Emoji 字符
- 选择 Emoji 后，由 `ArticleEdit.vue` 负责按当前光标位置插入正文，不额外引入自定义语法。

### 6. 新增 `web/admin/src/utils/media.js`

- 从 `ArticleEdit.vue` 中抽出上传逻辑，封装统一的 `uploadMedia(file)` 方法。
- 行为包括：
  - 使用 `FormData` 调用 `/api/admin/media/upload`
  - 统一附带 `credentials: "include"`
  - 解析 `{ code, message, data }` 返回结构
  - 在失败时抛出标准错误，供正文图片和封面上传复用

### 7. 更新 `web/admin/package.json`

- 新增 Emoji Picker 所需依赖。
- 保持现有 `lucide-vue-next` 作为后台界面图标来源；Emoji 库只负责正文可插入内容。

## Detailed Implementation Notes

### 光标插入与格式化策略

- 纯插入型按钮直接调用 `insertAtCursor()`，例如水平线、表格模板、代码块模板。
- 包裹型按钮优先包裹当前选区，无选区时插入占位文本并选中，适用于粗体、行内代码、链接。
- 列表类按钮按“当前行转换”处理，避免简单拼接导致格式错乱：
  - 无序列表：在当前行前补 `- `
  - 有序列表：在当前行前补 `1. `
  - 任务列表：在当前行前补 `- [ ] `
  - 引用：在当前行前补 `> `
- 图片上传成功后，正文插入 `![文件名](url)`；封面上传成功后仅更新 `coverImage` 字段。

### 页面结构策略

- 不重做整页布局，延续当前“左侧编辑/右侧设置”的双栏结构与移动端折叠逻辑。
- 把复杂交互组件化，但保持 `ArticleEdit.vue` 作为唯一页面入口，避免为当前需求引入额外状态管理。

### UI 一致性策略

- 交互风格继续对齐现有 Tailwind 实用类写法与 `lucide-vue-next` 图标集。
- 标签 Chips、Emoji 面板、封面上传按钮的 hover/focus 状态使用与当前后台一致的中性灰 + 蓝色强调色。
- 新增按钮和弹层保持键盘可聚焦，避免只支持鼠标操作。

## Verification Steps

- 运行管理后台构建，确认新增依赖与组件拆分后 `web/admin` 能正常打包。
- 手动验证新建文章场景：
  - 标签可搜索、可添加、可移除；无匹配时可即时创建并自动选中，保存后仍能正确回显。
  - 封面图可上传并自动预览，也可直接手填 URL。
  - 工具栏插入项都会落在当前光标或选区位置，而不是正文末尾。
  - 代码块、表格、列表、引用、水平线、Emoji 都能在编辑区插入并在右侧预览中正确显示。
- 手动验证编辑已有文章场景：
  - 已有标签、封面、正文内容能正确回填。
  - 替换封面、继续插入 Markdown 结构、保存草稿/发布后数据不丢失。
- 对最近修改文件执行诊断检查，确保没有新引入的 Vue/JS 语法或类型问题。
