# md-blog

一个基于 Go SSR + Vue Admin 的个人博客系统，支持文章管理、分类标签、媒体上传、SEO 设置和前台渲染。

## 环境要求

- Go 1.23+
- Node.js 18+（推荐 LTS）
- npm 9+

## 目录说明

- `cmd/server`：服务启动入口
- `internal`：后端核心业务代码
- `web/templates`：前台页面模板
- `web/admin`：后台管理前端（Vue + Vite）
- `data`：默认数据目录（SQLite、上传文件）
- `build/build.sh`：构建脚本（前端打包 + Go 二进制）

## 如何配置

项目采用“双层配置”：

- 启动配置通过 YAML 配置文件读取（`internal/config/config.go`），并允许环境变量覆盖
- 站点运行配置、预览密钥、上传限制、存储配置通过数据库中的后台设置维护

推荐从示例文件开始：

```bash
cp config.yaml.example config.yaml
```

### 最小可用配置（本地开发）

```yaml
app:
  addr: ":8080"
  env: "development"
  data_dir: "./data"
  session_secret: "please-change-me"

database:
  driver: "sqlite"
  sqlite_path: "./data/blog.db"
  auto_migrate: true

bootstrap_admin:
  username: "admin"
  password: "admin123456"
```

### 常用配置项

| YAML 键名 | 默认值 | 说明 |
| --- | --- | --- |
| `app.addr` | `:8080` | 服务监听地址 |
| `app.session_secret` | `change-me-session-secret` | Session 签名密钥，生产环境必须修改 |
| `app.data_dir` | `./data` | 数据目录 |
| `database.driver` | `sqlite` | 数据库驱动：`sqlite` 或 `mysql` |
| `database.sqlite_path` | `./data/blog.db` | SQLite 文件路径 |
| `database.mysql_dsn` | `root:root@tcp(127.0.0.1:3306)/md_blog?...` | MySQL 连接串（当 `database.driver=mysql`） |
| `database.auto_migrate` | `true` | 启动时自动迁移并初始化数据 |
| `bootstrap_admin.username` | `admin` | 初始管理员用户名 |
| `bootstrap_admin.password` | `admin123456` | 初始管理员密码（仅首次种子生效） |

### 环境变量覆盖

配置文件是主配置来源，但以下环境变量仍然会覆盖同名项，便于容器部署和密钥注入：

```bash
export APP_ADDR=:9090
export APP_SESSION_SECRET=please-change-me
export DB_DRIVER=sqlite
export DB_SQLITE_PATH=./data/blog.db
export DB_AUTO_MIGRATE=true
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin123456
```

以下配置会在首次启动时回填到数据库后台设置，后续以后台保存值为准；这一层仍沿用现有环境变量逻辑，本次未纳入 `config.yaml`：

- `APP_NAME`
- `APP_BASE_URL`
- `APP_PREVIEW_SECRET`
- `STORAGE_*`

## 如何启动

### 1) 安装后台前端依赖并打包

首次启动建议先构建后台静态资源，否则访问 `/admin` 可能返回 `admin not built`。

```bash
cd web/admin
npm install
npm run build
cd ../..
```

### 2) 启动后端服务

```bash
go run ./cmd/server --config ./config.yaml
```

也支持短参数：

```bash
go run ./cmd/server -c ./config.yaml
```

如果项目根目录下存在 `./config.yaml`，则不传参数也会自动加载；若文件不存在，则回退到“代码默认值 + 环境变量覆盖”模式。

启动后访问：

- 前台首页：`http://localhost:8080/`
- 管理后台：`http://localhost:8080/admin`
- 健康检查：`http://localhost:8080/healthz`

## 如何开发

### 后端开发

```bash
go run ./cmd/server --config ./config.yaml
```

如果你只依赖默认路径，也可以直接执行：

```bash
go run ./cmd/server
```

后端默认提供 API 前缀：`/api/admin/*`。

### 前端开发（热更新）

在另一个终端运行：

```bash
cd web/admin
npm install
npm run dev
```

Vite 默认地址：`http://localhost:5173`，并已代理：

- `/api` -> `http://127.0.0.1:8080`
- `/uploads` -> `http://127.0.0.1:8080`

这意味着你需要同时启动后端（8080）和前端 dev server（5173）进行联调。

## 构建发布

执行项目构建脚本：

```bash
./build/build.sh
```

会完成以下步骤：

- 构建后台前端产物到 `web/admin/dist`
- 构建后端二进制到 `dist/md-blog`

## 默认账号说明

当数据库为空且 `DB_AUTO_MIGRATE=true` 时，会自动创建管理员账号：

- 用户名：`ADMIN_USERNAME`（默认 `admin`）
- 密码：`ADMIN_PASSWORD`（默认 `admin123456`）

请在生产环境务必修改默认账号与 `APP_SESSION_SECRET`，并在后台设置中更新站点地址、预览密钥与存储配置。
