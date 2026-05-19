# 系统配置入库与后台管理重构方案

## Summary

- 目标：将当前 `internal/config/config.go` 中“非基础设施、非必须启动依赖”的配置迁移到数据库，并通过现有后台 `Settings` 页面统一管理，显著减少本地和生产启动时需要显式设置的环境变量。
- 核心策略：把配置拆成两层。
  - `BootConfig`：仅保留服务启动不可缺少的配置，继续来自环境变量。
  - `RuntimeSettings`：站点展示、SEO、预览、上传限制、存储驱动与 S3 参数等改为数据库单行配置，由后台管理。
- 用户已确认的实现方向：
  - 存储配置支持在后台切换，且修改后按新的数据库配置生效。
  - 初始管理员仍保留 `ADMIN_USERNAME` / `ADMIN_PASSWORD` 的 env 兜底，仅用于首次种子。

## Current State Analysis

### 1. 当前环境变量配置过重

- `internal/config/config.go` 当前把以下内容都放进 `Config`：
  - 应用启动：`APP_ADDR`、`APP_ENV`、`APP_DATA_DIR`、`APP_SESSION_SECRET`、`APP_READ_TIMEOUT`、`APP_WRITE_TIMEOUT`
  - 站点运行：`APP_NAME`、`APP_BASE_URL`、`APP_PREVIEW_SECRET`
  - 数据库：`DB_*`
  - 存储：`STORAGE_*`
  - 管理员初始化：`ADMIN_USERNAME`、`ADMIN_PASSWORD`
- 其中 `APP_PREVIEW_SECRET` 还被 `Validate()` 强制要求非空，导致项目本地运行必须设置该值。

### 2. 数据库里已经有一套“站点设置”，但覆盖范围有限

- `internal/model/models.go` 已存在 `SiteSetting` 单表单行模型，当前主要覆盖：
  - 站点名称、副标题、描述、关键词、About 内容
  - Hero 文案、主题、页脚、GitHub、ICP、Logo、默认 OG 图
  - `StorageDriver`、`StoragePublicURL`、`SearchPlaceholder`
- `internal/repository/setting_repository.go` 只有 `Get()` / `Save()`。
- `internal/handler/api/handler.go` 已提供 `/api/admin/settings` 读写接口。
- `web/admin/src/views/Settings.vue` 已有后台设置页，但只编辑展示型字段，没有系统运行配置、存储完整配置、安全设置。

### 3. 启动期与运行期配置耦合较深

- `internal/bootstrap/app.go`
  - 用 `cfg.App.SessionSecret` 初始化 session store。
  - 用 `cfg.Database.*` 打开数据库。
  - 用 `cfg.Storage.LocalDir` 创建上传目录。
  - 用 `cfg.App.BaseURL` 初始化 `seo.Service`。
  - 用 `cfg.App.Addr` / `ReadTimeout` / `WriteTimeout` 初始化 HTTP Server。
- `internal/service/media/service.go`
  - 持有完整 `config.Config`，存储驱动和 S3 客户端在服务创建时冻结，无法后台切换。
- `internal/router/router.go`
  - `/uploads/*` 使用固定的 `http.FileServer(http.Dir(c.Config.Storage.LocalDir))`，路径在启动时固定，无法按数据库配置动态变化。
- `internal/handler/web/handler.go`
  - 文章预览校验依赖 `h.c.Config.App.PreviewSecret`。
  - RSS / Sitemap / Robots 依赖 `h.c.Config.App.BaseURL`。
  - RSS 标题仍用 `h.c.Config.App.Name`，与数据库中的 `SiteSetting.SiteName` 存在双源。
- `internal/handler/api/handler.go`
  - 上传大小限制依赖 `h.c.Config.Storage.MaxUploadSize`。
  - 预览配置接口直接返回 `h.c.Config.App.PreviewSecret`。

### 4. 管理员账号当前只有“首次种子”，没有后台维护能力

- `internal/migration/migrate.go`
  - `seedAdminUser()` 仅在数据库不存在该用户名时，用 env 中的 `ADMIN_USERNAME` / `ADMIN_PASSWORD` 创建初始管理员。
- 当前仓库没有管理员账号资料或改密 API / 页面。
- 登录页 `web/admin/src/views/Login.vue` 仍预填 `admin / admin123456`。

## Proposed Changes

### 1. 将 `internal/config` 重构为最小启动配置

目标文件：

- `internal/config/config.go`
- `cmd/server/main.go`
- `internal/container/container.go`

改动内容：

- 将现有 `Config` 重构为“启动配置”语义，建议保持文件路径不变，但收敛字段范围，仅保留：
  - `App`: `Addr`、`Env`、`DataDir`、`SessionSecret`、`ReadTimeout`、`WriteTimeout`
  - `Database`: `Driver`、`SQLitePath`、`MySQLDSN`、`AutoMigrate`、连接池参数
  - `BootstrapAdmin`: `Username`、`Password`
- 从启动配置中移除并停止强制环境变量读取：
  - `APP_NAME`
  - `APP_BASE_URL`
  - `APP_PREVIEW_SECRET`
  - 全部 `STORAGE_*`
- `Validate()` 仅校验真正的启动必需项：
  - `APP_SESSION_SECRET` 非空
  - 数据库驱动合法
  - 连接参数合法
- 保留 `APP_DATA_DIR` 作为运行时文件根目录；后续本地上传目录改为基于该目录和数据库设置动态计算。

为什么这样做：

- 数据库连接、session secret、监听地址是启动前必须知道的，无法从数据库反向读取。
- 站点名、BaseURL、预览密钥、上传限制、存储驱动等都属于“应用运行配置”，适合入库并通过后台维护。

### 2. 扩展 `SiteSetting`，把运行配置正式入库

目标文件：

- `internal/model/models.go`
- `internal/migration/migrate.go`
- `internal/repository/setting_repository.go`

改动内容：

- 继续沿用现有 `SiteSetting` 单行模型，不新增第二张配置表，避免后台接口和前端设置页出现双模型。
- 在 `SiteSetting` 中补齐运行配置字段，建议新增：
  - `BaseURL`
  - `PreviewSecret`
  - `MaxUploadSize`
  - `StorageLocalPath`
  - `StorageLocalBaseURL`
  - `StorageS3Endpoint`
  - `StorageS3Region`
  - `StorageS3Bucket`
  - `StorageS3AccessKey`
  - `StorageS3SecretKey`
  - `StorageS3UseSSL`
  - `StorageS3PublicURL`
- 保留已有字段：
  - `SiteName`
  - `StorageDriver`
  - `StoragePublicURL` 可作为兼容字段处理；执行时统一收敛为 `StorageS3PublicURL`，避免和本地 URL 语义混淆。
- `seedSiteSetting()` 改为：
  - 首次启动时创建完整默认配置。
  - 默认值优先来自现有环境变量兼容值，以便老部署平滑迁移。
  - 对已存在记录只做“缺省回填”，不覆盖用户已保存的数据库值。
- 为仓库补充更稳的读取能力，建议增加：
  - `GetOrCreate(defaults)` 或 `Ensure()`：保证永远能取到配置单行。

为什么这样做：

- 当前已有 `/settings` 和单行配置模式，直接扩展比额外新增 `system_settings` 表更贴合现有代码结构。
- 首次迁移从旧 env 回填，可避免升级后丢失旧站点行为。

### 3. 新增设置服务，统一“数据库配置 + 启动配置”的解析与校验

目标文件：

- 新增 `internal/service/setting/service.go`
- `internal/bootstrap/app.go`
- `internal/container/container.go`

改动内容：

- 新增 `setting.Service`，职责包括：
  - 读取 `SiteSetting`
  - 对字段做默认值填充与格式规范化
  - 校验保存时的业务规则
  - 生成可供运行时直接使用的“解析后配置”
- 服务输出建议分层：
  - `SiteSetting`：数据库原始模型
  - `ResolvedSettings`：运行时解析结果，例如：
    - `BaseURL`
    - `PreviewSecret`
    - `MaxUploadSize`
    - `Storage.Driver`
    - `Storage.LocalDirAbs`
    - `Storage.LocalBaseURL`
    - `Storage.S3...`
- `ResolvedSettings.Storage.LocalDirAbs` 通过 `BootConfig.App.DataDir + SiteSetting.StorageLocalPath` 拼出，确保本地文件根目录仍受启动期 `APP_DATA_DIR` 控制，但无需再额外设置 `STORAGE_LOCAL_DIR`。
- `SaveSettings` 时由 `setting.Service` 统一做合法性校验：
  - `BaseURL` 必须是合法 URL，保存时去尾部 `/`
  - `PreviewSecret` 非空
  - `StorageDriver` 仅允许 `local` / `s3`
  - `MaxUploadSize` 必须大于 0
  - 选择 `s3` 时强校验 endpoint / bucket / access key / secret key
  - 选择 `local` 时强校验 `StorageLocalPath` / `StorageLocalBaseURL`

为什么这样做：

- 目前配置读取散落在 handler、router、service、bootstrap 中，没有统一的默认值和约束。
- 后续要支持后台切换存储，必须先有统一的解析入口，否则每个模块都会自己拼字段。

### 4. 让依赖运行配置的模块改为从数据库动态读取

目标文件：

- `internal/bootstrap/app.go`
- `internal/service/seo/service.go`
- `internal/service/media/service.go`
- `internal/handler/api/handler.go`
- `internal/handler/web/handler.go`
- `internal/router/router.go`
- `internal/container/container.go`

改动内容：

- `seo.Service`
  - 不再在构造时冻结 `baseURL`。
  - 改为依赖 `setting.Service` 或其 `ResolvedSettings` 读取器，在生成 Meta 时动态获取 `BaseURL`。
- `web.Handler`
  - 文章预览改为从数据库读取 `PreviewSecret`。
  - RSS / Sitemap / Robots 统一改用数据库中的 `BaseURL`。
  - RSS 标题改用 `SiteSetting.SiteName`，消除与旧 `cfg.App.Name` 的双源。
- `api.Handler`
  - 上传大小限制改用数据库 `MaxUploadSize`。
  - `/api/admin/preview-config` 返回数据库中的预览密钥。
- `media.Service`
  - 不再持有冻结的完整 `config.Config`。
  - 改为依赖 `BootConfig` + `setting.Service`：
    - 每次上传时读取当前解析后的存储配置。
    - 选择 `s3` 时按当前数据库配置创建或复用客户端。
    - 选择 `local` 时按 `ResolvedSettings.Storage.LocalDirAbs` 保存文件。
  - 为避免 S3 配置改动后仍使用旧客户端，服务内部应按配置签名做懒加载刷新，而不是只在启动时创建一次。
- `router.New()`
  - 取消固定 `http.FileServer(http.Dir(c.Config.Storage.LocalDir))` 的静态绑定。
  - 改成一个自定义上传文件 handler：
    - 请求命中 `/uploads/*` 时，通过 `setting.Service` 获取当前本地存储目录和 base URL。
    - 仅当当前驱动为 `local` 且 URL 前缀匹配本地 base URL 时从本地目录读取文件。
    - 当前驱动为 `s3` 时，本地 `/uploads/*` 可返回 `404`，避免错误暴露旧目录。

为什么这样做：

- 只有把运行期配置读取改成动态来源，后台修改才真正生效。
- 这一步是“后台切换存储配置”成立的关键。

### 5. 扩展后台设置页，覆盖系统运行配置

目标文件：

- `web/admin/src/views/Settings.vue`
- 如需拆分表单，可新增 `web/admin/src/components/*`

改动内容：

- 在现有“站点设置”页中新增三个配置分区，继续复用已有 `/api/admin/settings`：
  - 站点运行
    - 站点地址 `BaseURL`
    - 预览密钥 `PreviewSecret`
    - 上传大小限制 `MaxUploadSize`
  - 存储配置
    - 驱动切换：`local` / `s3`
    - 本地模式：`StorageLocalPath`、`StorageLocalBaseURL`
    - S3 模式：endpoint / region / bucket / access key / secret key / use ssl / public url
  - 兼容现有展示字段
    - 继续保留站点名称、SEO、Hero、About、Logo 等已有字段
- 表单交互要求：
  - 驱动切换时展示对应字段组
  - Secret 字段支持回显和编辑
  - 前端基本校验与后端校验保持一致
  - 保存成功后刷新本地状态，避免页面仍展示旧值

为什么这样做：

- 用户目标是“可以在管理后台进行配置”，现有设置页是最直接的落点。
- 沿用同一路由与页面，用户心智成本最低。

### 6. 补齐管理员账号后台维护，但保留 env 首次种子

目标文件：

- `internal/repository/admin_repository.go`
- `internal/service/auth/service.go`
- `internal/handler/api/handler.go`
- `internal/router/router.go`
- `web/admin/src/views/Settings.vue`

改动内容：

- 保留 `internal/migration/migrate.go` 里的 env 种子逻辑，但语义调整为“仅首次兜底初始化”：
  - 当数据库无管理员记录时，使用 `ADMIN_USERNAME` / `ADMIN_PASSWORD`
  - 若已有管理员，不再受 env 影响
- 补充后台管理员资料维护能力，建议最小范围包括：
  - 修改用户名
  - 修改密码
- 新增受保护 API，例如：
  - `POST /api/admin/account`
  - 或拆成 `POST /api/admin/account/profile` 与 `POST /api/admin/account/password`
- 在 `Settings.vue` 增加“安全设置”分区：
  - 当前用户名
  - 新密码 / 确认密码
  - 保存后提示重新登录或刷新当前会话
- 登录页 `web/admin/src/views/Login.vue` 去掉默认预填密码，降低默认口令误导。

为什么这样做：

- 用户明确希望减少启动时配置项，管理员账号不能永远靠 env 管理。
- 但完全移除 env 初始化会引入首启引导流程，超出本次最小重构范围，因此保留首次种子兜底最稳妥。

### 7. 文档与兼容迁移

目标文件：

- `README.md`
- `.env`（如仓库保留示例）

改动内容：

- 更新文档中的“最小可用配置”，将推荐显式设置项收敛为：
  - `APP_SESSION_SECRET`
  - 数据库连接相关 `DB_*`
  - 可选 `APP_ADDR`、`APP_DATA_DIR`、`APP_ENV`
  - 首次初始化时可选 `ADMIN_USERNAME` / `ADMIN_PASSWORD`
- 从文档中移除“必须设置 `APP_PREVIEW_SECRET` / `STORAGE_*` / `APP_BASE_URL`”的说法。
- 新增迁移说明：
  - 首次升级启动会把旧 env 中可迁移的值回填到数据库设置
  - 之后以后台保存的数据库值为准
  - `APP_SESSION_SECRET` 与数据库连接仍必须保留在环境变量

## Assumptions & Decisions

- 决定继续沿用 `SiteSetting` 单表，而不是新建 `SystemSetting` 表。
- 决定将“可运行时调整”的配置入库；“必须先知道才能启动”的配置保留 env。
- 决定 `APP_PREVIEW_SECRET` 从必填 env 改为数据库字段，并在首次迁移时生成或回填默认值。
- 决定本地存储目录以 `APP_DATA_DIR` 为根，数据库仅保存相对路径或逻辑路径，避免再次引入大量文件系统 env 配置。
- 决定后台存储切换按“保存后新上传立即生效”实现，不处理历史媒体自动迁移；历史媒体 URL 继续按保存时的 `Media.URL` 访问。
- 决定管理员账号采用“env 首次种子 + 后台后续维护”的折中方案，不新增独立首启安装向导。
- 决定本次不把以下配置迁入数据库：
  - `APP_SESSION_SECRET`
  - 数据库连接与连接池参数
  - HTTP Server 监听地址与读写超时

## Verification Steps

### 后端验证

- 启动时仅设置最小 env：
  - `APP_SESSION_SECRET`
  - `DB_*`
  - 可选 `APP_ADDR` / `APP_DATA_DIR`
  - 可选 `ADMIN_USERNAME` / `ADMIN_PASSWORD`
- 确认未设置 `APP_PREVIEW_SECRET`、`APP_BASE_URL`、`STORAGE_*` 时服务仍能启动。
- 首次启动后检查数据库中的 `site_settings` 已自动回填完整默认配置。
- 验证 `/api/admin/settings` 能读写新增字段，且校验错误信息合理。
- 验证 `/api/admin/preview-config` 返回数据库中的预览密钥。
- 验证修改 `BaseURL` 后 `rss.xml`、`sitemap.xml`、`robots.txt` 输出同步变化。

### 存储验证

- 驱动为 `local` 时：
  - 上传文件成功
  - 文件落盘到 `APP_DATA_DIR + StorageLocalPath`
  - `StorageLocalBaseURL` 可正常访问本地上传文件
- 在后台切换为 `s3` 后：
  - 新上传文件写入 S3
  - 新生成媒体 URL 使用数据库中的 S3 Public URL 或推导 URL
  - 本地 `/uploads/*` 对新驱动下请求不再错误读取旧目录
- 再切回 `local` 后：
  - 新上传文件重新落本地，无需重启服务

### 管理员验证

- 空库首次启动时，`ADMIN_USERNAME` / `ADMIN_PASSWORD` 能创建初始管理员。
- 已存在管理员后，再修改 env 不会覆盖已有账号。
- 后台可以修改用户名和密码。
- 修改密码后原密码失效，新密码可登录。

### 前端验证

- `web/admin/src/views/Settings.vue` 能完整加载、展示、保存新增系统配置。
- 存储驱动切换时，表单字段联动正确。
- 敏感字段、数字字段、URL 字段在前端有基础校验，非法输入会得到明确反馈。
