# 文章管理接入 AI 自动生成摘要与 SEO 计划

## Summary

* 目标：在文章新增与修改保存时，如果 `excerpt`、`seoKeywords`、`seoDescription` 为空，则由 AI 自动生成并回填。

* 配置位置：AI 配置放到后台“站点设置”中，由数据库持久化，不放入 `config.yaml` 或环境变量。

* 开关能力：提供 AI 总开关，关闭时完全走现有逻辑；开启时再按配置的 Provider 调用模型。

* Provider 策略：采用“统一业务接口 + 多 Provider 适配层”，而不是把所有厂商强行压到 OpenAI 兼容协议。

* 兼容目标：首版内建三类接入方式：

  * `openai_compatible`：覆盖 OpenAI、DeepSeek、Qwen 兼容网关、Kimi 兼容网关、Groq、OpenRouter、Ollama、SiliconFlow 等国内外主流 OpenAI 风格服务

  * `anthropic`：原生接 Claude

  * `gemini`：原生接 Gemini

* 失败策略：AI 失败不阻塞保存；摘要继续保留现有正文截取兜底，SEO 字段允许为空并记录降级日志。

## Current State Analysis

### 现有文章保存链路

* `internal/router/router.go` 已将后台文章写入统一到 `POST /api/admin/articles`。

* `internal/handler/api/handler.go` 的 `SaveArticle()` 直接解码 `article.SaveInput` 并调用 `h.c.Article.Save(input)`。

* `internal/service/article/service.go` 的 `Save()` 已包含标题/正文校验、Markdown 渲染、摘要默认值、SEO 字段写入、标签关联与分类/标签计数刷新。

* 当前只有 `Excerpt` 具备自动兜底；`SEODescription`、`SEOKeywords` 为空时直接保存为空。

### 现有文章字段约束

* `internal/model/models.go` 中：

  * `Excerpt` 最大 500

  * `SEODescription` 最大 500

  * `SEOKeywords` 最大 255

* 这些字段已贯通后台编辑页、前台文章页 SEO 标签和文章列表卡片，因此不需要新增文章表字段。

### 现有设置链路

* `internal/model/models.go` 的 `SiteSetting` 已承载站点 SEO、预览密钥、上传限制和存储配置。

* `internal/repository/setting_repository.go`、`internal/service/setting/service.go`、`internal/handler/api/handler.go` 已形成完整的“读取/保存站点设置”后端链路。

* `internal/migration/migrate.go` 会 `AutoMigrate(&model.SiteSetting{})`，并通过 `defaultSiteSetting()` + `mergeSiteSetting()` 维护默认值。

* `web/admin/src/views/Settings.vue` 已有完整的设置页 Tab、表单、加载与保存逻辑，适合直接扩展 AI 配置 UI。

### 现有前端回填能力

* `web/admin/src/views/ArticleEdit.vue` 保存成功后，会使用服务端返回值重新回填 `excerpt`、`seoKeywords`、`seoDescription`。

* 因此本次只要服务端在保存时补齐字段，前端无需额外新增“生成结果同步”逻辑。

### 现有依赖边界

* 项目当前 `go.mod` 为 Go 1.23。

* 仓库中还没有任何 AI 客户端依赖，也没有统一日志组件。

* 若直接选单一 OpenAI 兼容 SDK，无法原生覆盖 Gemini / Anthropic；因此更适合拆成“兼容类 Provider + 原生 Provider”两层。

## Assumptions & Decisions

* AI 配置统一落在 `SiteSetting`，沿用现有设置页保存机制，不新增独立 AI 配置表。

* 首版 Provider 枚举固定为：

  * `openai_compatible`

  * `anthropic`

  * `gemini`

* `openai_compatible` 通过 `baseUrl + apiKey + model` 适配不同服务商，覆盖国内外大量主流模型平台。

* `anthropic` 通过官方 Anthropic Go SDK 原生调用。

* `gemini` 通过官方 Google GenAI Go SDK 原生调用。

* 依赖选择：

  * OpenAI 兼容类：`github.com/sashabaranov/go-openai`

  * Anthropic：`github.com/anthropics/anthropic-sdk-go`

  * Gemini：`google.golang.org/genai`

* 不引入统一的第三方“全 Provider 超级 SDK”作为核心依赖；业务层自己定义统一接口，Provider 适配层分别封装成熟 SDK。

* AI 调用合并为单次请求，一次生成摘要、SEO 描述、SEO 关键词，避免串行 3 次请求。

* 字段优先级：

  * `excerpt`：用户手填 > AI 生成 > `excerptFrom(content)`

  * `seoDescription` / `seoKeywords`：用户手填 > AI 生成 > 空值

* AI 功能受站点设置总开关控制；关闭时完全不发起任何模型请求。

* AI Key 等敏感字段沿用当前设置页对 `storageS3SecretKey` 的处理方式，后台可编辑保存；前端输入框建议使用 `type="password"`。

## Proposed Changes

### 1. 更新 `internal/model/models.go`

* 在 `SiteSetting` 中新增 AI 配置字段，建议如下：

  * `AIEnabled bool`

  * `AIProvider string`

  * `AIModel string`

  * `AIAPIKey string`

  * `AIBaseURL string`

  * `AITimeoutSeconds int`

* 字段职责：

  * `AIEnabled`：AI 总开关

  * `AIProvider`：Provider 类型，值为 `openai_compatible` / `anthropic` / `gemini`

  * `AIModel`：模型名

  * `AIAPIKey`：访问密钥

  * `AIBaseURL`：仅 `openai_compatible` 场景必填

  * `AITimeoutSeconds`：请求超时，避免拖慢保存链路

### 2. 更新 `internal/migration/migrate.go`

* 在 `defaultSiteSetting()` 中补充 AI 默认值：

  * `AIEnabled: false`

  * `AIProvider: "openai_compatible"`

  * `AITimeoutSeconds: 15`

* 在 `mergeSiteSetting()` 中补充 AI 字段默认值合并，确保老库升级后能自动补齐基础值。

* 不新增单独 migration 文件，继续复用现有 `AutoMigrate` 机制。

### 3. 更新 `internal/service/setting/service.go`

* 在 `normalize()` 中新增 AI 字段 trim、默认值和 Provider 规范化。

* 在 `Validate()` 中新增 AI 配置校验：

  * `AIEnabled=false` 时，允许 AI 字段为空

  * `AIEnabled=true` 时：

    * `AIProvider` 必须是三种枚举之一

    * `AIModel` 非空

    * `AITimeoutSeconds > 0`

    * `openai_compatible` 需要 `AIAPIKey`、`AIBaseURL`

    * `anthropic` 需要 `AIAPIKey`

    * `gemini` 需要 `AIAPIKey`

* 如有必要，新增 `ResolvedAISettings`，统一输出给 AI 服务使用，避免业务层直接读裸 `SiteSetting`。

### 4. 更新 `internal/handler/api/handler.go`

* `GetSettings()` / `SaveSettings()` 结构保持不变，因为当前已直接传输 `model.SiteSetting`。

* `SaveArticle()` 调用改为 `h.c.Article.Save(r.Context(), input)`，为 AI 调用透传取消与超时上下文。

### 5. 更新 `web/admin/src/views/Settings.vue`

* 在 `settings` 响应式对象中新增 AI 配置字段。

* 在 `tabs` 中新增 `ai` 选项卡，避免把 AI 配置混入现有 SEO/存储区域。

* 新增 AI 设置表单，建议包含：

  * AI 开关

  * Provider 下拉框

  * 模型名称输入框

  * API Key 输入框

  * Base URL 输入框（仅 `openai_compatible` 显示）

  * 超时秒数输入框

* 在 UI 中加入简短说明，明确：

  * `openai_compatible` 可用于 OpenAI、DeepSeek、Ollama、OpenRouter、SiliconFlow、兼容网关版 Qwen/Kimi 等

  * `anthropic` 用于 Claude

  * `gemini` 用于 Gemini

* 保存逻辑继续复用现有 `/api/admin/settings`。

### 6. 新增 `internal/service/ai/provider.go`

* 定义业务层统一接口，例如：

  * `type MetadataGenerator interface { GenerateArticleMetadata(ctx context.Context, input GenerateInput) (*GenerateResult, error) }`

* `GenerateInput` 至少包含：

  * `Title string`

  * `Content string`

* `GenerateResult` 包含：

  * `Excerpt string`

  * `SEOKeywords string`

  * `SEODescription string`

* 新增统一的 `ProviderType`、`Config`、`Factory`，由工厂按后台设置构造具体实现。

### 7. 新增 `internal/service/ai/openai_compatible.go`

* 使用 `go-openai` 封装 OpenAI 兼容 Provider。

* 由后台设置中的 `AIBaseURL + AIAPIKey + AIModel` 构造客户端。

* 适配对象为所有 OpenAI 风格平台，不在代码里硬编码具体厂商列表。

* 统一输出 `GenerateResult`，不把 SDK 类型泄漏到业务层。

### 8. 新增 `internal/service/ai/anthropic.go`

* 使用 `anthropic-sdk-go` 封装 Claude Provider。

* 使用官方消息接口生成 JSON 结构化文本结果。

* 配置来源为后台设置中的 `AIAPIKey + AIModel`。

### 9. 新增 `internal/service/ai/gemini.go`

* 使用 `google.golang.org/genai` 封装 Gemini Provider。

* 通过官方 SDK 生成严格 JSON 文本，并转换为统一 `GenerateResult`。

* 配置来源为后台设置中的 `AIAPIKey + AIModel`。

### 10. 新增 `internal/service/ai/noop.go`

* 当 `AIEnabled=false` 或配置不完整时，返回 Noop 实现。

* Noop 不发起请求，直接返回空结果，让文章服务自然走现有兜底逻辑。

### 11. 新增 `internal/service/ai/prompt.go`

* 抽离统一 Prompt 构造与输出解析逻辑，确保不同 Provider 行为一致。

* Prompt 约束：

  * 仅依据标题和正文生成

  * 输出中文

  * 摘要用于文章列表展示

  * SEO 描述自然，不堆砌关键词

  * SEO 关键词返回简洁短语列表

  * 只输出 JSON，不加解释文字和代码块

* 统一做后处理：

  * trim 空格

  * 压缩换行

  * 长度裁剪到字段上限

  * 兼容中英文逗号

### 12. 更新 `internal/bootstrap/app.go`

* 启动时先解析站点设置，再按 `ResolvedAISettings` 构造 AI 元数据服务。

* 构造顺序建议为：

  * `settingsService := settingSvc.New(...)`

  * `resolvedSettings, err := settingsService.Resolve()`

  * `aiService := aiSvc.NewFromSettings(resolvedSettings)`

  * `articleService := articleSvc.New(..., aiService)`

* 这样 AI 配置变更后，只要服务端读取的是最新 settings，就能在后续保存流程生效；如果首版实现采用启动期装配，则计划中应同步说明需要重启服务才能更新客户端配置。

* 更稳妥的首版建议：不要把具体 SDK 客户端永久缓存为只读配置，而是在文章保存时根据当前 settingsService 解析并创建轻量 provider，确保后台设置保存后立即生效。

### 13. 更新 `internal/container/container.go`

* 如 AI 服务需要被其他模块复用，可新增 `AI` 字段。

* 若 AI 只服务于文章保存，也可以只注入 `article.Service`，不暴露到容器外层。

* 推荐方案：容器中新增 `AISettings` 或 `AI` 字段，便于后续扩展标题建议、标签建议等功能。

### 14. 更新 `internal/service/article/service.go`

* 修改 `Service` 结构，增加 AI 依赖或 AI 工厂依赖。

* 修改 `Save()` 签名为 `Save(ctx context.Context, input SaveInput)`。

* 在保存流程中加入“仅补空字段”的步骤：

  * 当 `excerpt`、`seoKeywords`、`seoDescription` 至少一项为空时，调用 AI

  * 只回填原本为空的字段

  * AI 失败时记录日志并降级

  * 摘要仍为空时再走 `excerptFrom(content)`

* 其他 slug、Markdown 渲染、标签同步、计数刷新逻辑保持原样，减少行为回归。

### 15. 更新测试

* `internal/service/article/service_test.go`

  * 增加 AI 成功补全空字段

  * 已有值不被覆盖

  * AI 失败时保存仍成功

  * 超长返回值被裁剪

* 新增 `internal/service/setting/service_test.go`

  * 校验 AI 设置默认值与 Provider 规则

  * 校验 `AIEnabled=true` 时的必填项

* 视实现复杂度新增 `internal/service/ai/service_test.go`

  * Prompt 输出 JSON 解析

  * 不同 Provider 结果转换一致性

  * 降级路径覆盖

## Detailed Implementation Notes

### Provider 选择逻辑

* `AIEnabled=false`：直接 Noop

* `AIProvider=openai_compatible`：走兼容协议客户端

* `AIProvider=anthropic`：走 Claude 原生客户端

* `AIProvider=gemini`：走 Gemini 原生客户端

### 保存时序

1. 后台编辑页提交文章。
2. `SaveArticle` 解码请求体并传入 `r.Context()`。
3. `article.Service.Save(ctx, input)` 校验标题和正文。
4. 若三个目标字段存在空值，则通过当前设置构造 Provider 并请求 AI。
5. 仅补空字段，不覆盖手填值。
6. AI 失败则降级，不中断保存。
7. 摘要仍为空时走 `excerptFrom(content)`。
8. 按原逻辑保存文章并返回，前端用返回值自然回填。

### 设置即时生效策略

* 因为 AI 配置位于后台设置而不是启动配置，首版应优先保证“保存设置后立即生效”。

* 推荐做法：文章服务内部依赖 `setting.Service` 或其解析结果，而不是只在 `bootstrap` 启动时读一次静态配置。

* 这样用户在后台切换 Provider、模型、Key 或开关后，不需要重启服务。

### 安全与显示策略

* AI API Key 在前端使用密码输入框展示，减少误泄露。

* 后端仍按现有 settings 风格返回完整值给已登录管理员，不额外设计脱敏协议，以保持与现有 `storageS3SecretKey` 一致。

### 不纳入本次范围

* 不新增“测试连接”按钮

* 不新增“手动重新生成摘要/SEO”按钮

* 不对不同厂商做专属高级参数面板（如温度、max tokens、top\_p）

* 不记录完整 Prompt/Response 审计日志

## Verification Steps

* 数据与设置验证：

  * 启动后老库自动补齐 AI 字段默认值

  * 后台设置保存 AI 配置成功并能再次读取回显

  * 关闭 AI 开关时，文章保存行为与当前完全一致

* Provider 验证：

  * `openai_compatible` 场景可通过 `baseUrl + key + model` 工作

  * `anthropic` 场景配置最小字段后可工作

  * `gemini` 场景配置最小字段后可工作

* 文章保存验证：

  * 三个字段都为空时，保存后自动生成

  * 仅一个字段为空时，只补该字段

  * AI 调用失败时，保存仍成功，摘要有兜底，SEO 允许为空

* 测试与诊断：

  * 覆盖 `article`、`setting`、`ai` 相关单测

  * 对 `models.go`、`migrate.go`、`service.go`、`Settings.vue`、`handler.go` 获取诊断，确保没有新增编译错误

