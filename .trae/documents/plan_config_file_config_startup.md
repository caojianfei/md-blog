# 启动配置改为配置文件方案

## Summary

- 目标：将当前 `internal/config/config.go` 基于环境变量的启动配置，改造为“YAML 配置文件为主、环境变量可覆盖、启动时可通过 `--config` / `-c` 指定配置文件”的模式。
- 范围：仅覆盖启动配置层，即 `internal/config` 及其调用入口；数据库中的后台设置和 `internal/migration` 中用于首次初始化站点设置的环境变量读取暂不调整。
- 核心实现：引入成熟配置库 `github.com/spf13/viper` 负责 YAML 文件解析、默认值注入与环境变量覆盖；在 `cmd/server/main.go` 增加配置文件路径参数解析。

## Current State Analysis

- `cmd/server/main.go` 当前直接调用 `config.Load()`，没有任何命令行参数解析逻辑。
- `internal/config/config.go` 当前通过 `os.Getenv` + `getEnv/getIntEnv/getBoolEnv` 读取以下启动配置：
  - `App`: `APP_ADDR`、`APP_ENV`、`APP_DATA_DIR`、`APP_SESSION_SECRET`、`APP_READ_TIMEOUT`、`APP_WRITE_TIMEOUT`
  - `Database`: `DB_DRIVER`、`DB_SQLITE_PATH`、`DB_MYSQL_DSN`、`DB_AUTO_MIGRATE`、`DB_MAX_OPEN_CONNS`、`DB_MAX_IDLE_CONNS`、`DB_CONN_MAX_LIFETIME`
  - `BootstrapAdmin`: `ADMIN_USERNAME`、`ADMIN_PASSWORD`
- 当前配置校验仅在 `Config.Validate()` 中完成，校验项较少，但足以保证基础启动。
- 仓库中暂无现成 `config.yaml`，也没有后端 CLI 参数解析代码。
- `README.md` 当前文档以 shell 环境变量导出为主，`.env` 也是环境变量示例文件。

## Proposed Changes

### 1. `go.mod`

- 新增依赖 `github.com/spf13/viper`。
- 理由：`viper` 在 Go 生态足够成熟，原生支持 YAML、默认值、环境变量覆盖和显式配置文件加载，能满足本次需求且避免继续维护手写 env 解析逻辑。

### 2. `cmd/server/main.go`

- 增加启动参数解析，支持：
  - `--config`
  - `-c`
- 参数值统一映射为“配置文件路径”输入给 `config.Load(...)`。
- 实现方式：
  - 使用标准库 `flag`，避免为仅一个启动参数引入完整 CLI 框架。
  - `config.Load()` 改为接受配置路径参数，例如 `config.Load(configPath)`。
- 行为约定：
  - 显式传入 `--config/-c` 时，若文件不存在或解析失败，启动直接报错并退出。
  - 未显式传参时，尝试读取项目根目录默认配置文件 `./config.yaml`；若不存在，则回退到“代码默认值 + 环境变量覆盖”模式。

### 3. `internal/config/config.go`

- 保留现有 `Config`、`AppConfig`、`DatabaseConfig`、`AdminConfig` 结构体，避免影响下游调用。
- 移除当前手写的 `getEnv/getIntEnv/getBoolEnv` 加载流程，改为基于 `viper` 的加载器。
- 新的加载顺序固定为：
  - 代码默认值
  - YAML 配置文件
  - 环境变量覆盖
- 具体实现细节：
  - 为 `viper` 设置默认值，保持当前默认行为不变，例如：
    - `app.addr = :8080`
    - `app.env = development`
    - `app.data_dir = ./data`
    - `database.driver = sqlite`
    - `bootstrap_admin.username = admin`
  - 使用 `SetConfigFile` 加载指定路径或默认路径。
  - 使用 `SetConfigType("yaml")` 兼容标准 YAML 配置文件。
  - 通过 `SetEnvKeyReplacer(strings.NewReplacer(".", "_"))` 和统一前缀映射，使环境变量可继续覆盖结构化字段。
  - 环境变量名继续兼容当前命名，不改动现有部署变量：
    - `APP_ADDR` 对应 `app.addr`
    - `DB_DRIVER` 对应 `database.driver`
    - `ADMIN_USERNAME` 对应 `bootstrap_admin.username`
    - 其余字段保持一一映射
  - `DB_SQLITE_PATH` 的默认值继续依赖 `App.DataDir`，即默认落在 `${app.data_dir}/blog.db`。
- 保留并适度增强 `Validate()`：
  - 继续校验 `SessionSecret` 非空
  - 继续限制 `database.driver` 为 `sqlite` 或 `mysql`
  - 新增按驱动校验必要字段：
    - `sqlite` 时要求 `sqlite_path` 非空
    - `mysql` 时要求 `mysql_dsn` 非空
- 返回方式建议从“内部 `panic`”改为“返回错误给调用方”，但如果为了尽量减少改动，也可保留 `Load()` 内部失败即 `panic` 的现状。
- 本次计划采用更清晰的方式：
  - `config.Load(path string) (Config, error)`
  - 由 `main.go` 负责记录日志并退出
  - 这样错误来源会更明确，也更利于后续测试

### 4. `internal/config/config_test.go`

- 新增配置加载单测，覆盖最关键行为：
  - 读取 YAML 配置文件成功
  - 环境变量优先级高于 YAML
  - 未提供配置文件时，默认值仍可正常生成配置
  - 指定不存在的配置文件时报错
  - `sqlite/mysql` 驱动校验生效
- 测试策略：
  - 使用临时目录创建测试 YAML 文件
  - 用 `t.Setenv` 设置环境变量，避免污染全局环境
  - 不引入端到端测试，保持测试聚焦且成本可控

### 5. `config.yaml.example`

- 在仓库根目录新增示例配置文件 `config.yaml.example`。
- 内容与 `Config` 结构一一对应，示例结构如下：
  - `app`
  - `database`
  - `bootstrap_admin`
- 目的：
  - 给本地开发和部署提供可复制模板
  - 让 YAML 键名成为稳定文档，而不是让用户猜测字段名
- 示例文件不放敏感真实值，只保留安全示例和注释式说明。

### 6. `README.md`

- 更新“如何配置”和“如何启动”部分，改为“配置文件为主，环境变量可覆盖”。
- 增加 YAML 示例与启动示例：
  - `go run ./cmd/server --config ./config.yaml`
  - `go run ./cmd/server -c ./config.yaml`
- 补充默认行为说明：
  - 若存在 `./config.yaml` 则自动加载
  - 环境变量可覆盖同名配置项，适合容器部署和密钥注入
- 保留说明：`internal/migration` 中首次启动写入数据库的配置仍沿用现有环境变量逻辑，本次不纳入统一配置文件。

### 7. `.env`

- 是否修改 `.env` 不作为必须项。
- 建议处理：
  - 保留现有 `.env` 作为环境变量覆盖示例，减少对现有本地习惯的破坏。
  - 如果需要，可在文件顶部补充注释，说明其角色已变为“可选覆盖层”，而不是主配置来源。
- 若执行阶段考虑最小改动，可不修改此文件。

## Assumptions & Decisions

- 配置文件格式确定为 YAML。
- 仅改造启动配置，不重构数据库中的后台设置体系，也不调整 `internal/migration` 的环境变量初始化逻辑。
- 配置依赖选型确定为 `viper`，CLI 参数解析使用标准库 `flag`。
- 启动参数同时支持 `--config` 和 `-c`。
- 默认路径确定为项目根目录 `./config.yaml`。
- 优先级确定为：默认值 < YAML 文件 < 环境变量。
- 为了提升可维护性，`config.Load` 将改为返回错误而不是直接 `panic`。

## Verification Steps

- 执行 `go test ./internal/config`，确认新增配置加载测试通过。
- 执行 `go test ./...`，确认配置改造未破坏现有后端逻辑。
- 手动验证以下场景：
  - 仅使用默认值启动
  - 使用 `./config.yaml` 自动加载启动
  - 使用 `--config` / `-c` 指定配置文件启动
  - 使用环境变量覆盖 YAML 中的单个字段，例如覆盖 `APP_ADDR` 或 `DB_DRIVER`
  - 指定不存在的配置文件时返回清晰错误
- 检查文档与示例文件：
  - `README.md` 的命令示例与实际行为一致
  - `config.yaml.example` 的键名与 `Config` 结构和 `viper` 映射一致
