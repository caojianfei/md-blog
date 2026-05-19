# 文章/分类/标签计数、删除联动与首页分类弹窗升级计划

## Summary

* 为 `Category` 和 `Tag` 增加持久化的 `articleCount` 字段，统一表示“已发布文章数量”。

* 前台分类、标签按持久化的已发布文章数倒序排列，并隐藏数量为 0 的项。

* 首页分类区块只展示前 5 个分类；若存在更多分类，提供“查看更多分类”入口并在弹窗中展示全部可见分类。

* 管理后台补齐文章删除能力，并增强分类/标签删除逻辑：删除前按“是否关联任何文章（含草稿）”检查，命中时先二次确认，再在后端事务中清理关联关系。

* 文章保存、状态切换、删除后，需要同步回写受影响分类/标签的持久化 `articleCount`。

## Current State Analysis

### 数据模型与仓储

* `internal/model/models.go`

  * `Article` 通过 `CategoryID` 关联分类，通过 `many2many:article_tags` 关联标签。

  * `Category`、`Tag` 当前没有关联文章数量字段。

* `internal/repository/taxonomy_repository.go`

  * `CategoryRepository.List()` 当前按 `sort ASC, created_at DESC` 排序。

  * `TagRepository.List()` 当前按 `created_at DESC` 排序。

  * 分类、标签删除当前只是直接 `Delete(id)`，没有先检查关联文章，也没有做关系清理。

* `internal/repository/article_repository.go`

  * 已支持文章列表、按分类/标签过滤、标签关联替换、仪表盘统计。

  * 还没有文章删除方法。

  * 没有“按已发布文章重算分类/标签计数”的能力。

### 后端接口

* `internal/router/router.go`

  * 已有 `DELETE /api/admin/categories/{id}` 与 `DELETE /api/admin/tags/{id}`。

  * 还没有 `DELETE /api/admin/articles/{id}`。

* `internal/handler/api/handler.go`

  * 分类/标签删除接口仅直接调用仓储删除，无二次确认支撑信息，无事务处理。

  * 分类/标签列表直接返回仓储结果，没有数量字段。

  * 文章列表、详情、保存、状态切换已存在，删除不存在。

* `internal/handler/web/handler.go`

  * `baseData()` 将 `CategoryRepo.List()` / `TagRepo.List()` 直接给模板。

  * 首页没有“Top 5 分类”和“查看更多分类”所需的独立字段。

### 前台模板与交互

* `web/templates/pages/home.html`

  * 首页侧栏分类当前展示全部分类，未显示文章数量，也没有查看更多入口。

  * 标签列表展示全部标签，未按数量排序。

* `web/assets/frontend.js`

  * 已承载首页终端、主题切换等交互。

  * 目前没有首页分类弹窗逻辑。

* `web/assets/style.css`

  * 需要补充首页“查看更多分类”按钮与弹窗样式。

### 管理后台

* `web/admin/src/views/ArticleList.vue`

  * 已支持编辑、状态切换、预览；缺删除按钮与删除请求。

* `web/admin/src/views/CategoryList.vue`

  * 已支持新建/编辑；缺删除按钮，未展示关联文章数量。

* `web/admin/src/views/TagList.vue`

  * 已支持新建/编辑；缺删除按钮，未展示关联文章数量。

* `web/admin/src/utils/request.js`

  * 可直接复用，配合后端返回的 409 / 400 错误信息和 DELETE 请求。

## Proposed Changes

### 1. 模型、迁移与计数同步能力

* `internal/model/models.go`

  * 为 `Category` 新增持久化字段 `ArticleCount int64`，映射 JSON `articleCount`。

  * 为 `Tag` 新增持久化字段 `ArticleCount int64`，映射 JSON `articleCount`。

  * 字段语义统一为“已发布文章数量”，草稿不参与统计。

* `internal/migration/migrate.go`

  * 继续使用现有 `AutoMigrate` 为 `categories` 与 `tags` 表补充 `article_count` 列。

  * 为兼容历史数据，在迁移后补一次计数回填逻辑，避免已有分类/标签初始值都为 0。

* `internal/repository/article_repository.go`

  * 新增 `Delete(id uint) error`，用于文章删除。

  * 新增“定向重算并回写已发布文章数”的方法，按受影响 ID 刷新，而不是每次全量重算，例如：

    * `RefreshCategoryCounts(ids []uint) error`

    * `RefreshTagCounts(ids []uint) error`

  * 这些方法内部基于数据库聚合 `status = published` 的文章数，再更新 `categories.article_count` / `tags.article_count`。

  * 新增分类/标签任意关联文章数量查询方法，用于后台删除确认，例如：

    * `CountByCategory(categoryID uint) (int64, error)`，统计全部关联文章。

    * `CountByTag(tagID uint) (int64, error)`，统计全部关联文章。

  * 新增关系清理方法，并在事务中复用：

    * `ClearCategory(categoryID uint) error`，将相关文章的 `category_id` 更新为 `NULL`。

    * `DetachTag(tagID uint) error`，删除 `article_tags` 关联记录。

* `internal/repository/taxonomy_repository.go`

  * 保留基础 `Save` / `FindByID` / `FindBySlug` / `Delete`。

  * `List()` 继续返回模型自身持久化的 `articleCount`，后台不强制按计数排序。

* `internal/service/article/service.go`

  * `Save()` 在更新文章前先读取旧文章及旧标签集合，收集“受影响的分类 ID / 标签 ID”。

  * 保存文章、替换标签后，调用 `RefreshCategoryCounts()` / `RefreshTagCounts()` 只重算受影响项。

  * `UpdateStatus()` 在状态变更后同步刷新当前文章所属分类与标签的持久化计数。

  * 新增 `Delete(id uint)` 业务方法：

    * 先读取文章及其分类/标签关系。

    * 删除文章。

    * 重算受影响分类/标签的已发布文章数。

  * 采用“定向重算”而不是手工加减，避免文章编辑时同时变更分类、标签、状态造成计数漂移。

### 2. 后端删除接口与二次确认协议

* `internal/router/router.go`

  * 新增 `DELETE /api/admin/articles/{id}` 路由。

* `internal/handler/api/handler.go`

  * 新增 `DeleteArticle()`：

    * 调用 `article.Service.Delete()` 删除文章。

    * 依赖业务层在删除后同步回写受影响分类/标签的已发布文章数。

    * 返回统一成功响应。

  * 重构 `DeleteCategory()`：

    * 先统计该分类下全部关联文章数。

    * 若数量大于 0 且请求未带确认标识，返回 `409 Conflict`，响应中包含 `articleCount` 与说明文案。

    * 若已确认，则在事务中先把相关文章 `category_id` 置空，再删除分类。

  * 重构 `DeleteTag()`：

    * 先统计该标签关联的全部文章数。

    * 若数量大于 0 且请求未带确认标识，返回 `409 Conflict`。

    * 若已确认，则在事务中先删除 `article_tags` 关联，再删除标签。

  * 删除确认标识统一采用查询参数 `force=1`，前端第一次 DELETE，若收到 409 则弹出二次确认并以 `?force=1` 重试。

  * `DeleteCategory()` / `DeleteTag()` 的二次确认判断基于“任意关联文章数”，不使用持久化 `articleCount`。

  * `ListCategories()` / `ListTags()` 直接返回持久化的 `articleCount`，供后台列表使用。

  * `TerminalCategories()` / `TerminalTags()` 同步使用新的前台排序与过滤结果，避免首页终端数据与页面展示不一致。

### 3. 前台分类/标签排序、计数与首页 Top 5

* `internal/handler/web/handler.go`

  * 在 `PageData` 中新增字段：

    * `TopCategories []model.Category`

    * `HasMoreCategories bool`

  * `Categories` 保持“前台全部可见分类”；`Tags` 保持“前台全部可见标签”。

  * 在 `baseData()` 中新增前台数据整形逻辑：

    * 读取全部分类/标签。

    * 直接使用持久化的 `ArticleCount` 作为“已发布文章数量”。

    * 隐藏 `ArticleCount == 0` 的分类和标签。

    * 对分类和标签按 `ArticleCount DESC` 排序；数量相同可退化为当前的 `Sort ASC, created_at DESC`（分类）和 `created_at DESC`（标签）作为稳定次排序。

    * 生成 `TopCategories = Categories[:min(5, len(Categories))]`。

    * 当 `len(Categories) > 5` 时设 `HasMoreCategories = true`。

  * `Category()` / `Tag()` 页面加载当前分类或标签时，直接使用模型上的持久化 `ArticleCount`。

* `web/templates/pages/home.html`

  * 分类区块改为渲染 `TopCategories`。

  * 每个分类项展示“分类名 + 文章数量”。

  * 当 `HasMoreCategories` 为真时，增加“查看更多分类”按钮。

  * 页面底部增加隐藏的分类弹窗结构，弹窗内渲染全部 `Categories`（已经是前台可见全集）。

  * 标签区块改为使用已排序后的 `Tags`；本次不展示数量。

* `web/templates/pages/category.html`

  * 分类页标题处展示当前分类文章数，例如“X 篇文章”，与首页侧栏口径一致。

### 4. 首页弹窗交互与样式

* `web/assets/frontend.js`

  * 新增首页分类弹窗初始化逻辑：

    * 监听“查看更多分类”按钮。

    * 支持打开、关闭、遮罩点击关闭、Esc 关闭。

    * 不影响现有终端和像素猫逻辑。

  * 交互采用 `data-*` 钩子，避免和现有脚本耦合。

* `web/assets/style.css`

  * 补充“查看更多分类”按钮样式。

  * 补充分类数量文本样式。

  * 补充弹窗遮罩、面板、关闭按钮、分类列表项样式，延续现有深色技术感风格。

### 5. 管理后台删除入口与数量展示

* `web/admin/src/views/ArticleList.vue`

  * 增加删除按钮。

  * 点击删除时先做一次普通确认；确认后调用 `DELETE /api/admin/articles/{id}`。

  * 删除成功后刷新当前列表。

* `web/admin/src/views/CategoryList.vue`

  * 增加“关联文章数”列，显示后端返回的持久化 `articleCount`。

  * 增加删除按钮。

  * 删除流程：

    * 先发起 `DELETE /api/admin/categories/{id}`。

    * 若后端返回 409，则弹出明确文案二次确认：“该分类下有 N 篇文章，删除后这些文章将变为未分类，是否继续？”

    * 用户确认后以 `DELETE /api/admin/categories/{id}?force=1` 重试。

  * 新建/编辑弹窗继续复用现有 `Modal.vue`；删除确认可先用 `window.confirm`，避免引入新组件。

* `web/admin/src/views/TagList.vue`

  * 增加“关联文章数”列，显示后端返回的持久化 `articleCount`。

  * 增加删除按钮。

  * 删除流程与分类一致，但二次确认文案改为：“该标签关联 N 篇文章，删除后会同步移除这些文章上的该标签，是否继续？”

## Assumptions & Decisions

* 计数口径已确认：

  * 分类和标签持久化一个 `articleCount` 字段，统一表示“已发布文章数”。

  * 草稿不参与该字段统计。

  * 前台隐藏 `articleCount == 0` 的分类和标签。

* 删除确认口径已确认：

  * 只要分类/标签关联了任何文章，无论发布还是草稿，都需要二次确认。

  * 因此删除确认必须单独查询关联文章数，不能直接依赖持久化 `articleCount`。

* 需要通过 migration 为历史库补列，并在升级后补一次历史计数回填。

* 分类/标签删除的二次确认由“前端接收 409 后再次确认并带 `force=1` 重试”驱动，避免仅靠前端本地计数做不可靠判断。

* 文章删除入口先放在 `ArticleList.vue`，本次不额外扩展到 `ArticleEdit.vue`，因为用户需求中的“支持删除”已经可通过列表完成。

* 首页“查看更多分类”弹窗只展示前台可见的全部分类，即：仅已发布文章数大于 0 的分类，且按数量倒序排列。

## Verification Steps

* 后端验证

  * 新增/运行针对仓储和 API handler 的测试，覆盖：

    * 历史数据回填后分类/标签 `articleCount` 正确。

    * 文章新建、编辑（改分类/改标签）、状态切换、删除后，受影响分类/标签的持久化 `articleCount` 正确。

    * 删除分类时未确认返回 409、确认后文章转为未分类。

    * 删除标签时未确认返回 409、确认后 `article_tags` 关系被清理。

    * 删除文章后文章消失且标签关联记录被移除，相关 `articleCount` 同步减少。

  * 至少执行 `go test ./...`。

* 前端验证

  * 管理后台手动验证：

    * 文章列表可删除文章，删除后列表刷新。

    * 分类/标签列表能显示关联文章数。

    * 删除空分类/空标签直接成功。

    * 删除有文章的分类/标签时先收到二次确认，再重试删除成功。

  * 博客前台手动验证：

    * 首页分类只显示前 5。

    * 首页分类显示文章数，按数量倒序。

    * 标签按数量倒序，且不显示数量。

    * 0 篇文章的分类/标签不显示。

    * 超过 5 个分类时出现“查看更多分类”，弹窗内能看到完整分类列表并可点击跳转。

    * 分类页标题的文章数与实际列表条数一致。

* 构建与诊断

  * 执行后台前端构建检查（如 `npm run build`，目录为 `web/admin`）。

  * 如前台脚本/CSS有改动，启动本地预览后检查首页交互。

  * 对改动过的 Go / Vue / JS 文件执行诊断检查，确保无新增错误。

