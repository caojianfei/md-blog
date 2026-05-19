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

- 启动配置通过环境变量读取（`internal/config/config.go`）
- 站点运行配置、预览密钥、上传限制、存储配置通过数据库中的后台设置维护

没有强制 `.env` 文件。你可以直接在 shell 里导出变量。

### 最小可用配置（本地开发）

```bash
export APP_ENV=development
export APP_ADDR=:8080
export APP_SESSION_SECRET=please-change-me

export DB_DRIVER=sqlite
export DB_SQLITE_PATH=./data/blog.db
export DB_AUTO_MIGRATE=true

export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=admin123456
```

### 常用配置项

| 变量名 | 默认值 | 说明 |
| --- | --- | --- |
| `APP_ADDR` | `:8080` | 服务监听地址 |
| `APP_SESSION_SECRET` | `change-me-session-secret` | Session 签名密钥，生产环境必须修改 |
| `APP_DATA_DIR` | `./data` | 数据目录 |
| `DB_DRIVER` | `sqlite` | 数据库驱动：`sqlite` 或 `mysql` |
| `DB_SQLITE_PATH` | `./data/blog.db` | SQLite 文件路径 |
| `DB_MYSQL_DSN` | `root:root@tcp(127.0.0.1:3306)/md_blog?...` | MySQL 连接串（当 `DB_DRIVER=mysql`） |
| `DB_AUTO_MIGRATE` | `true` | 启动时自动迁移并初始化数据 |
| `ADMIN_USERNAME` | `admin` | 初始管理员用户名 |
| `ADMIN_PASSWORD` | `admin123456` | 初始管理员密码（仅首次种子生效） |

以下配置会在首次启动时回填到数据库后台设置，后续以后台保存值为准：

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
go run ./cmd/server
```

启动后访问：

- 前台首页：`http://localhost:8080/`
- 管理后台：`http://localhost:8080/admin`
- 健康检查：`http://localhost:8080/healthz`

## 如何开发

### 后端开发

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
