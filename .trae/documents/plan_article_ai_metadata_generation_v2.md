# 文章管理接入 AI 自动生成摘要与 SEO 计划（更新版）

## Summary

* 目标：在后台新增或编辑文章时，如果 `excerpt`、`seoKeywords`、`seoDescription` 任一为空，则由 AI 自动生成缺失字段并回填到保存结果。

* 配置位置：AI 配置放入站点设置，由 `SiteSetting` 持久化，不进入 `config.yaml` 或启动环境变量。

* 开关能力：后台提供 AI 总开关；关闭时完全走现有逻辑，开启后才尝试请求模型。

* Provider 方案：采用“统一业务接口 + 多 Provider 适配层”，首版内建 `openai_compatible`、`anthropic`、`gemini` 三类接入方式。

* 通用性目标：`openai_compatible` 通过 `baseUrl + apiKey + model` 适配 OpenAI、DeepSeek、Qwen 兼容网关、Kimi 兼容网关、SiliconFlow、OpenRouter、Groq、Ollama 等国内外主流 OpenAI 风格服务；`anthropic` 和 `gemini` 原生接官方 SDK。

* 失败策略：AI 失败、AI 配置异常、Provider 响应解析失败都不阻塞文章保存；`excerpt` 继续保留正文截取兜底，SEO 字段允许为空。

## Current State Analysis

### 已完成部分

* `internal/model/models.go`

  * `SiteSetting` 已新增 `AIEnabled`、`AIProvider`、`AIModel`、`AIAPIKey`、`AIBaseURL`、`AITimeoutSeconds` 字段。

* `internal/migration/migrate.go`

  * `defaultSiteSetting()` 和 `mergeSiteSetting()` 已补入 AI 默认值与老库兜底逻辑。

* `internal/service/setting/service.go`

  * 已新增 `ResolvedAI` / `ResolvedSettings.AI`。

  * `normalize()` 已做 AI 字段 trim、默认值与 provider 规范化。

  * `Validate()` 已校验 AI 开关、Provider 枚举、模型、密钥、超时和 `openai_compatible` 的 `baseUrl`。

* `internal/service/ai/provider.go`

  * 已定义统一配置、输入输出结构、`MetadataGenerator` 接口和工厂分发逻辑。

* `internal/service/ai/prompt.go`

  * 已定义统一 Prompt、JSON 解析和长度裁剪逻辑。

* `internal/service/ai/noop.go`

  * 已提供空实现，用于 AI 关闭或配置不足场景。

* `internal/service/ai/openai_compatible.go`

  * 已完成 OpenAI 兼容 Provider 的初版实现。

### 当前缺口

* `internal/service/ai`

  * 目前只有 `provider.go`、`prompt.go`、`noop.go`、`openai_compatible.go` 四个文件。

  * `AnthropicGenerator`、`GeminiGenerator` 尚未落文件，`provider.go` 仍引用了未定义类型，当前代码未闭环。

* `internal/service/article/service.go`

  * `Save()` 仍是 `Save(input SaveInput)`，没有 `context.Context`。

  * 当前只有 `Excerpt` 的非 AI 兜底；`SEODescription`、`SEOKeywords` 为空仍直接保存空值。

  * `Service` 构造函数尚未注入 AI 相关依赖。

* `internal/handler/api/handler.go`

  * `SaveArticle()` 仍调用 `h.c.Article.Save(input)`，没有把 `r.Context()` 透传给文章服务。

* `internal/bootstrap/app.go`

  * 文章服务仍按 `articleSvc.New(articleRepo, categoryRepo, tagRepo, markdownService)` 构造，没有设置服务或 AI 生成器依赖。

* `web/admin/src/views/Settings.vue`

  * AI 字段尚未进入 `settings` reactive 对象。

  * 当前 Tab 仍只有站点、内容、SEO、存储、安全，没有 AI 配置页。

* 测试

  * `internal/service/article/service_test.go` 仍只覆盖分类/标签计数逻辑。

  * 尚无 AI Prompt 解析测试、设置 AI 校验测试、文章保存补空测试。

### 版本与依赖现状

* 仓库当前 `go.mod` 已被依赖解析提升到 `go 1.24`。

* 已加入的 AI 依赖：

  * `github.com/sashabaranov/go-openai`

  * `github.com/anthropics/anthropic-sdk-go`

  * `google.golang.org/genai`

* 依赖最低 Go 版本核对结果：

  * `go-openai`：`go 1.18`

  * `anthropic-sdk-go`：`go 1.23.0`

  * `google.golang.org/genai`：`go 1.24`

* 结论：若首版继续原生支持 Gemini 官方 SDK，则项目 `go.mod` 需要保留 `go 1.24`，不再回退到 `1.23`。

## Assumptions & Decisions

* Provider 固定为三种：

  * `openai_compatible`

  * `anthropic`

  * `gemini`

* 采用成熟官方或社区主流 SDK，而不是引入一个跨厂商超级封装：

  * OpenAI 兼容类：`github.com/sashabaranov/go-openai`

  * Anthropic：`github.com/anthropics/anthropic-sdk-go`

  * Gemini：`google.golang.org/genai`

* AI 配置即时生效，不要求重启服务。

  * 具体做法：文章保存时通过 `setting.Service.Resolve()` 读取当前设置，再动态构造当次请求所需的 generator。

  * 不在 `bootstrap` 启动时缓存一个绑定固定配置的长生命周期 AI 客户端。

* 文章保存只补空字段，不覆盖用户已填写内容。

  * `excerpt`：用户手填 > AI 生成 > `excerptFrom(content)`

  * `seoDescription` / `seoKeywords`：用户手填 > AI 生成 > 空值

* 当 AI 开启但数据库中存在脏配置时，文章保存不报错。

  * 文章服务内部应把“读取/解析 AI 设置失败”视为可降级错误，记录日志后继续保存。

  * 管理后台保存设置接口仍保持严格校验，避免管理员写入新的非法配置。

* 首版不纳入以下能力：

  * 不新增“测试连接”按钮

  * 不新增“手动生成摘要/SEO”按钮

  * 不新增温度、Top P、Max Tokens 等厂商专属高级参数

  * 不记录完整 Prompt / Response 审计日志

* AI Key 在后台设置页使用密码输入框展示，延续现有“管理员可编辑保存敏感字段”的交互模式。

* 项目 Go 版本决策：

  * 保留 `go.mod` 的 `go 1.24`，因为 `google.golang.org/genai` 官方 SDK 需要 `go 1.24`。

## Proposed Changes

### 1. 完成 `internal/service/ai/anthropic.go`

* 做什么：

  * 新增 `AnthropicGenerator`，实现 `GenerateArticleMetadata(ctx, input)`。

* 为什么：

  * `provider.go` 已定义 `ProviderAnthropic`，但当前缺实现，导致抽象无法真正覆盖 Claude。

* 怎么做：

  * 使用 `anthropic-sdk-go` 官方客户端。

  * 调用消息接口发送统一 Prompt。

  * 从响应内容中提取文本结果并复用 `parseResponse()`。

  * 为请求套上 `context.WithTimeout()`，超时时间来自 `Config.TimeoutSeconds`。

  * 返回空 choices / 非文本内容时按空结果处理，不中断文章保存上层流程。

### 2. 完成 `internal/service/ai/gemini.go`

* 做什么：

  * 新增 `GeminiGenerator`，实现 `GenerateArticleMetadata(ctx, input)`。

* 为什么：

  * 用户明确要求支持 Gemini；当前只有工厂占位，没有真正实现。

* 怎么做：

  * 使用 `google.golang.org/genai` 官方 SDK。

  * 通过 `GenerateContent` 发送统一 Prompt。

  * 从返回对象中提取最终文本并复用 `parseResponse()`。

  * 同样套用 `context.WithTimeout()`。

  * 非结构化或空响应返回空结果，由文章服务继续兜底。

### 3. 收敛 `internal/service/ai/provider.go` 与 `prompt.go`

* 做什么：

  * 补全编译闭环，必要时增加辅助函数。

* 为什么：

  * 当前 `provider.go` 已引用未实现类型；`prompt.go` 是三类 Provider 的共用层，需要作为唯一解析入口。

* 怎么做：

  * 保持统一接口不变。

  * 若 SDK 返回格式需要预处理，新增内部私有 helper，例如“从响应块提取纯文本”的函数。

  * 保持字段上限与数据库约束一致：

    * `excerpt` 500

    * `seoDescription` 500

    * `seoKeywords` 255

### 4. 改造 `internal/service/article/service.go`

* 做什么：

  * 为文章保存链路接入 AI 自动补空。

* 为什么：

  * 这是新增/修改文章的核心入口，且当前只有摘要兜底，没有 SEO 自动生成。

* 怎么做：

  * 修改 `Service` 结构，新增 `settings *setting.Service` 依赖。

  * `New()` 改为注入 `setting.Service`。

  * `Save()` 改为 `Save(ctx context.Context, input SaveInput)`。

  * 保存流程调整为：

    * 先校验标题和正文。

    * 先渲染 Markdown，保持现有逻辑。

    * 计算三个目标字段的用户输入值。

    * 若三者至少一项为空，则尝试：

      * `settings.Resolve()`

      * 基于 `ResolvedSettings.AI` 构造 `ai.NewGenerator(...)`

      * 只请求一次，生成三个候选值

      * 只回填原本为空的字段

    * 若 AI 失败或设置解析失败，记录日志并继续。

    * `excerpt` 最终仍为空时，再执行 `excerptFrom(content)`。

  * 其他 slug、标签同步、分类/标签计数刷新逻辑保持不变，减少回归面。

### 5. 更新 `internal/handler/api/handler.go`

* 做什么：

  * 透传请求上下文到文章保存服务。

* 为什么：

  * 让 AI 调用能够响应客户端取消、HTTP 超时和服务端 deadline。

* 怎么做：

  * 将 `SaveArticle()` 中的 `h.c.Article.Save(input)` 改为 `h.c.Article.Save(r.Context(), input)`。

  * 不调整返回结构，前端继续使用现有保存成功回填逻辑。

### 6. 更新 `internal/bootstrap/app.go`

* 做什么：

  * 把 `setting.Service` 注入文章服务。

* 为什么：

  * AI 配置位于站点设置中，文章保存时需要实时读取。

* 怎么做：

  * 继续保留现有 `settingsService := settingSvc.New(cfg, settingRepo)`。

  * 文章服务构造改为注入 `settingsService`，例如 `articleSvc.New(articleRepo, categoryRepo, tagRepo, markdownService, settingsService)`。

  * `bootstrap` 不创建固定 AI 客户端实例，避免配置修改后必须重启。

### 7. 视需要更新 `internal/container/container.go`

* 做什么：

  * 同步文章服务构造后的类型签名变化。

* 为什么：

  * 若 `article.Service` 构造函数签名变化，容器装配需要同步。

* 怎么做：

  * `Container` 结构本身通常无需新增 `AI` 字段，因为首版 AI 只服务文章保存。

  * 若执行时发现后续测试或其他模块确实需要复用 AI 能力，再补容器字段；本次默认不暴露 `AI` 服务到容器顶层。

### 8. 更新 `web/admin/src/views/Settings.vue`

* 做什么：

  * 在设置页补齐 AI 配置表单。

* 为什么：

  * 用户明确要求 AI 配置放到管理后台设置中，并提供开关。

* 怎么做：

  * 在 `settings` reactive 对象中加入：

    * `aiEnabled`

    * `aiProvider`

    * `aiModel`

    * `aiApiKey`

    * `aiBaseUrl`

    * `aiTimeoutSeconds`

  * 在 `tabs` 中新增 `ai` 选项卡。

  * 新增 AI 配置区块，包含：

    * AI 开关复选框

    * Provider 下拉框

    * 模型名称输入框

    * API Key 密码框

    * Base URL 输入框，仅 `openai_compatible` 显示

    * 超时秒数输入框

  * 增加简短文案说明：

    * `openai_compatible` 用于 OpenAI / DeepSeek / Qwen 兼容网关 / Kimi 兼容网关 / SiliconFlow / OpenRouter / Ollama 等

    * `anthropic` 用于 Claude

    * `gemini` 用于 Gemini

  * 保存逻辑继续复用现有 `/api/admin/settings`。

### 9. 更新 `internal/service/article/service_test.go`

* 做什么：

  * 补充文章保存自动生成元数据的核心回归测试。

* 为什么：

  * 这是最容易引入行为回归的区域。

* 怎么做：

  * 引入一个测试用设置服务替代真实 AI 调用，避免单测访问外网。

  * 推荐测试场景：

    * 三个字段都为空时，AI 成功补齐三者

    * 只有某一项为空时，只补该项

    * 已填写字段不会被覆盖

    * AI 返回错误时，保存仍成功，`excerpt` 走本地兜底

    * AI 返回超长文本时，最终落库值被裁剪到字段上限

### 10. 新增或更新 `internal/service/setting/service_test.go`

* 做什么：

  * 为 AI 设置规范化与校验补测试。

* 为什么：

  * `setting.Service` 已经承担 AI 配置合法性约束，需要有稳定回归覆盖。

* 怎么做：

  * 覆盖以下场景：

    * 默认值正确

    * `AIEnabled=false` 时允许字段为空

    * `AIEnabled=true` 时模型、Key、超时必填

    * `openai_compatible` 缺 `baseUrl` 时报错

    * provider 非枚举值时报错

### 11. 新增 `internal/service/ai/prompt_test.go`

* 做什么：

  * 覆盖 Prompt 输出解析与清洗逻辑。

* 为什么：

  * 不同厂商响应最终都会收敛到 `parseResponse()`，这是统一风险点。

* 怎么做：

  * 覆盖以下场景：

    * 正常 JSON 解析

    * 带 Markdown 代码块包裹的 JSON 解析

    * 中文逗号转换为英文逗号

    * 空白压缩与长度裁剪

    * 非法 JSON 返回错误

### 12. 调整 `go.mod` / `go.sum`

* 做什么：

  * 在执行阶段整理依赖归属，使实际使用的 AI SDK 成为直接依赖。

* 为什么：

  * 当前三类 AI SDK 多数还是 indirect，且项目 Go 版本已被提升，需要最终收敛成可解释状态。

* 怎么做：

  * 保留 `go 1.24`。

  * 在真正落地 `anthropic.go`、`gemini.go` 后，通过测试或 tidy 让相应依赖转为直接依赖。

  * 不再尝试回退 `go 1.23`。

## Data Flow

1. 管理员在 `Settings.vue` 配置 AI 开关、Provider、模型、密钥等，并保存到 `/api/admin/settings`。
2. 文章新增或编辑时，请求进入 `POST /api/admin/articles`。
3. `internal/handler/api/handler.go` 将 `r.Context()` 和 `SaveInput` 传入文章服务。
4. `internal/service/article/service.go` 检查 `excerpt` / `seoKeywords` / `seoDescription` 是否存在空值。
5. 若需要 AI 补空：

   * 调用 `setting.Service.Resolve()` 获取当前 AI 配置

   * 构造 `ai.Config`

   * 调用 `ai.NewGenerator(...)`

   * 单次请求生成三个字段

   * 只填充原先为空的项
6. 若 AI 失败：

   * 记录日志

   * 不中断保存

   * `excerpt` 继续走 `excerptFrom(content)` 兜底
7. 文章照常落库并返回，前端使用返回结果自然回填表单。

## Edge Cases & Failure Modes

* AI 关闭：

  * 完全不构造远程 generator，直接 Noop。

* AI 配置不完整：

  * `SaveSettings()` 不允许管理员保存非法启用配置。

  * 若数据库已有脏数据，文章保存端降级而非报错。

* 用户手动填写部分字段：

  * 只补空字段，不覆盖已有值。

* Provider 返回非 JSON 文本：

  * 解析失败，走降级保存。

* Provider 响应字段为空：

  * 空字段继续走后续兜底，摘要可退回 `excerptFrom(content)`。

* 内容过长或模型生成过长：

  * 统一通过 `sanitizeResult()` 裁剪到数据库允许上限。

* 请求被用户取消：

  * `r.Context()` 取消后，AI 调用应尽快结束，不应继续阻塞。

## Verification Steps

* 后端单测：

  * `go test ./internal/service/ai ./internal/service/setting ./internal/service/article`

* 编译验证：

  * `go test ./...`

* 诊断检查：

  * 对以下文件运行诊断并清理新增问题：

    * `internal/service/ai/provider.go`

    * `internal/service/ai/openai_compatible.go`

    * `internal/service/ai/anthropic.go`

    * `internal/service/ai/gemini.go`

    * `internal/service/article/service.go`

    * `internal/service/setting/service.go`

    * `internal/handler/api/handler.go`

    * `internal/bootstrap/app.go`

    * `web/admin/src/views/Settings.vue`

* 手工验收：

  * 后台设置页可保存并回显 AI 配置。

  * 关闭 AI 后，文章保存行为与当前一致。

  * 开启 AI 且三个字段都为空时，保存后自动补齐。

  * 仅某一字段为空时，只补该字段。

  * 切换 Provider 或模型后，无需重启服务即可对后续文章保存生效。

  * AI 服务异常时，文章仍能保存成功，摘要保留本地兜底。

