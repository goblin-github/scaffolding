# Scaffolding

基于 Go 1.26 + Gin 的 Web 后端脚手架，提供开箱即用的项目骨架、分层架构和基础设施集成。

## 目录

- [快速开始](#快速开始)
- [项目架构](#项目架构)
- [目录结构](#目录结构)
- [分层设计](#分层设计)
- [组件装配](#组件装配)
- [配置管理](#配置管理)
- [中间件](#中间件)
- [错误处理 & i18n](#错误处理--i18n)
- [认证体系](#认证体系)
- [开发规范](#开发规范)
- [依赖](#依赖)

---

## 快速开始

```bash
# 1. 修改配置
vim configs/config.yml

# 2. 准备 i18n 翻译文件（可选，已有示例）
# configs/i18n/locales/en.json
# configs/i18n/locales/zh.json

# 3. 运行
go run cmd/app-server/main.go
```

### 配置要点

- `app.mode`: `debug`（开发） / `release`（生产），生产模式 Gin 关闭堆栈追踪
- `jwt.secret`: 留空则跳过 JWT 初始化，认证中间件自动关闭
- 所有配置项均可通过环境变量覆盖，前缀 `APP_`，分隔符用 `_`，例如 `APP_DATABASE_DSN`

---

## 项目架构

```
┌──────────────────────────────────────────────────────────────┐
│                      cmd/app-server                          │
│  main() — 编排入口：加载配置 → 初始化 → 装配 → 启动 → 优雅关闭  │
└──────────────────────────┬───────────────────────────────────┘
                           │
              ┌────────────▼────────────┐
              │   internal/registry.go  │
              │  Registry 聚合基础设施    │
              │  DB / Cache / JWT / i18n │
              └────────────┬────────────┘
                           │
              ┌────────────▼────────────┐
              │   internal/server.go    │
              │  装配调用链 → HTTP Server │
              └────────────┬────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    ┌────▼────┐    ┌───────▼──────┐    ┌─────▼─────┐
    │ Router  │    │  Middleware  │    │  Handler  │
    │ 路由注册  │    │  CORS/Auth/  │    │  请求处理   │
    │          │    │  Log/ReqID  │    │           │
    └─────────┘    └──────────────┘    └─────┬─────┘
                                             │
                                    ┌────────▼────────┐
                                    │    Service       │
                                    │  业务逻辑，纯 Go   │
                                    └────────┬────────┘
                                             │
                                    ┌────────▼────────┐
                                    │  Repository      │
                                    │  唯一写 GORM 的地方 │
                                    └──────────────────┘
```

**核心设计原则：依赖注入（手动装配，无框架）**。每一层的依赖通过构造函数传入，在 `internal/server.go` 的 `NewServer()` 中显式组装。

---

## 目录结构

```
.
├── cmd
│   └── app-server
│       └── main.go              # 入口：配置加载 → 初始化 → 启动 → 优雅关闭
├── configs
│   ├── config.yaml              # 默认配置文件
│   └── i18n
│       └── locales
│           ├── en.json          # 英文翻译
│           └── zh.json          # 中文翻译
├── internal                     # 私有包，不对外暴露
│   ├── config
│   │   ├── config.go            # 配置结构体 + viper 加载
│   │   └── i18n.go              # i18n Bundle：加载翻译、Accept-Language 解析
│   ├── errcode
│   │   └── errcode.go           # 业务错误码（AppError）
│   ├── handler
│   │   └── article.go           # HTTP handler：参数绑定 → 调用 service → 响应
│   ├── middleware
│   │   ├── access_log.go        # 访问日志 + X-Process-Time
│   │   ├── cors.go              # CORS 配置
│   │   ├── jwt_auth.go          # JWT 三层校验中间件
│   │   └── request_id.go        # 请求 ID 注入（优先复用客户端传入的）
│   ├── model
│   │   └── article.go           # GORM 模型（纯数据结构，无方法）
│   ├── registry.go              # Registry 聚合基础设施，InitRegistry 工厂
│   ├── repository
│   │   └── article.go           # 数据访问层（唯一写 GORM 代码的地方）
│   ├── router
│   │   └── router.go            # 路由注册 + Gin Engine 创建
│   ├── server.go                # Server 封装：装配调用链 → http.Server
│   └── service
│       └── article.go           # 业务逻辑层（接收 context.Context）
└── pkg                          # 可复用公共包（可被其他项目引用）
    ├── auth
    │   └── jwt.go               # JWT 签发 + 三层校验（签名→黑名单→版本号）
    ├── cache
    │   └── redis.go             # Redis 客户端封装
    ├── database
    │   └── mysql.go             # MySQL（GORM）连接初始化
    ├── logger
    │   └── logger.go            # 基于 slog 的结构化日志（支持轮转+控制台双输出）
    ├── response
    │   └── response.go          # 统一响应格式 {code, msg, data}
    └── validator
        └── validator.go         # 参数校验错误的多语言翻译
```

### 为什么分 `internal` 和 `pkg`？

| 目录          | 作用                        | 可被外部 import？ |
|-------------|---------------------------|--------------|
| `internal/` | 当前项目的业务代码、配置、路由           | ❌ Go 编译器禁止   |
| `pkg/`      | 通用基础设施封装（DB、Cache、Logger） | ✅ 可被其他项目引用   |

---

## 分层设计

### 调用链

```
Handler → Service → Repository → GORM/DB
```

每层职责明确：

| 层              | 入参                       | 出参                | 规则                   |
|----------------|--------------------------|-------------------|----------------------|
| **Handler**    | `*gin.Context`           | JSON 响应           | 只做参数绑定和响应，不放业务逻辑     |
| **Service**    | `context.Context` + 基本类型 | `(*model, error)` | 纯业务逻辑，不依赖 gin        |
| **Repository** | `context.Context` + 参数   | `(*model, error)` | **唯一允许放 GORM 代码的地方** |
| **Model**      | —                        | —                 | 纯数据结构，无任何方法          |

### Service 为什么接收 `context.Context` 而非 `*gin.Context`？

同一个 service 可以被 HTTP handler 和未来的 gRPC handler、消息队列 worker、定时任务共同调用。不与任何框架绑定。

### Repository 为什么是唯一放 GORM 代码的地方？

如果 GORM 调用散落在 service 层各处，换 ORM 或优化查询时需要改动大量文件。集中到 repository 后，变更只影响这一层。

---

## 组件装配

项目采用**手动依赖注入**，不引入 wire/dig 等框架。所有装配逻辑集中在两处：

### 1. Registry（基础设施）

`internal/registry.go` 的 `InitRegistry()` 负责初始化基础设施并聚合：

```go
type Registry struct {
Config *config.Config
DB     *gorm.DB
Cache  *cache.Cache
JWT    *auth.JWTManager
I18n   *config.Bundle
}
```

初始化顺序：数据库 → Redis → i18n → JWT。cleanup 函数按逆序释放资源。

JWT 和认证中间件可选：`jwt.secret` 为空时跳过 JWT 初始化，`registry.JWT == nil` 时路由不挂载认证中间件。

### 2. Server（调用链）

`internal/server.go` 的 `NewServer()` 显式装配 handler → service → repository：

```go
articleRepo := repository.NewArticleRepository(reg.DB)
articleSvc := service.NewArticleService(articleRepo)
articleHandler := handler.NewArticleHandler(articleSvc)
```

### 3. 路由装配

`internal/router/router.go` 通过 `Dependencies` 结构体接收所有 handler 和中间件，每个模块一个私有函数注册路由：

```go
type Dependencies struct {
Article *handler.ArticleHandler
JWTAuth gin.HandlerFunc
}
```

新增业务模块时：创建 model → repository → service → handler，然后在 `Dependencies` 加字段，在 `NewServer` 中装配，在 router
中加一个 `registerXxxRoutes`。

---

## 配置管理

使用 [viper](https://github.com/spf13/viper) 加载 YAML 配置，兼容环境变量覆盖：

```yaml
# configs/config.yml
app:
  port: ":8080"
  mode: "release"          # debug | release
  read_timeout: 30         # 读请求超时（秒）
  write_timeout: 30        # 写响应超时（秒）
  idle_timeout: 60         # Keep-Alive 空闲超时（秒）
  shutdown_timeout: 10     # 优雅关机最大等待（秒）

logger:
  filename: "logs/app.log"
  level: "info"            # debug | info | warn | error
  max_size: 10             # 单文件最大 MB
  max_backups: 5           # 保留旧文件数
  max_age: 30              # 保留天数
  compress: true           # 旧文件是否 gzip
  console: true            # 是否同时输出到控制台

database:
  dsn: "root:123456@tcp(127.0.0.1:3306)/my_db?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600  # 秒

cache:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  pool_size: 50

jwt:
  secret: ""               # 留空跳过 JWT 初始化
  expires_after: 7200      # 秒
```

环境变量覆盖规则：前缀 `APP_`，分隔符 `.` → `_`。例如 `APP_DATABASE_DSN=root:xxx@tcp(...)` 会覆盖配置文件中的
`database.dsn`。

---

## 中间件

| 中间件       | 文件                         | 作用                                          |
|-----------|----------------------------|---------------------------------------------|
| RequestID | `middleware/request_id.go` | 注入 `X-Request-ID` 到 context 和响应头，优先复用客户端传入值 |
| CORS      | `middleware/cors.go`       | 跨域配置，预检请求直接返回 204                           |
| AccessLog | `middleware/access_log.go` | 记录请求方法、路径、状态码、耗时、客户端 IP                     |
| JWTAuth   | `middleware/jwt_auth.go`   | JWT 鉴权（非全局，按路由模块挂载）                         |
| Recovery  | `gin.Recovery()`           | Gin 内置 panic 恢复                             |

### 全局 vs 路由级中间件

```go
// router.go — 全局中间件（所有路由生效）
engine.Use(middleware.RequestID())
engine.Use(middleware.CORS())
engine.Use(middleware.AccessLog())
engine.Use(gin.Recovery())

// 路由级中间件 — 仅需要认证的接口
if deps.JWTAuth != nil {
articles.Use(deps.JWTAuth)
}
```

---

## 错误处理 & i18n

### 错误码体系

`internal/errcode/errcode.go` 定义 `AppError`：

```go
type AppError struct {
Code    int // 前端 switch-case
Reason  string // i18n 翻译 key + 日志检索关键字
Message string // 默认英文消息（i18n 查不到时的 fallback）
}
```

通用错误码：`400 INVALID_PARAMS` / `401 UNAUTHORIZED` / `403 FORBIDDEN` / `404 NOT_FOUND` / `500 INTERNAL_ERROR` /
`503 SYSTEM_BUSY`。

业务错误码以 5xxx 开头，示例：`5001 TOO_MANY_TAGS`。

### 统一响应格式

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    ...
  }
}
```

- `Success(c, data)` — code=0
- `Error(c, err)` — 从 err 提取 AppError，自动 i18n 翻译 msg
- `ParamError(c, validationErrors)` — 收集所有校验错误并用分号拼接

### i18n 翻译流程

```
request Accept-Language: zh-CN,en;q=0.9
       │
       ▼
config.Translate(lang, appErr.Reason, appErr.Message)
       │
       ├── zh.json 命中 → 返回中文翻译
       ├── 未命中 → fallback en.json
       └── 也未命中 → 返回 appErr.Message（英文原文）
```

翻译文件放在 `configs/i18n/locales/`，key 对应 `AppError.Reason`。

---

## 认证体系

JWT 三层校验（`pkg/auth/jwt.go`）：

```
                    ┌──────────────┐
                    │  Token 传入   │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
              Layer │ 签名 + 过期校验 │  ← jwt 库完成
                    └──────┬───────┘
                           │ 通过
                    ┌──────▼───────┐
              Layer │ JTI 黑名单检查 │  ← Redis SetNX（可选，cache 不为 nil 时生效）
                    └──────┬───────┘
                           │ 通过
                    ┌──────▼───────┐
              Layer │ Token 版本校验 │  ← 改密码后批量踢人（可选，tvStore 不为 nil 时生效）
                    └──────────────┘
```

特性：

- JTI（JWT ID）用 UUID，通过 Redis 黑名单防止 token 被盗后的重放攻击
- `TokenVersionStore` 接口允许在用户改密后递增版本号，使旧 token 全部失效
- `jwt.secret` 为空时 JWT 不初始化，认证中间件自动跳过——适合开发阶段

---

## 开发规范

### 新增业务模块步骤

以添加 `User` 模块为例：

1. **Model** — `internal/model/user.go`

```go
type User struct {
ID   uint   `gorm:"primarykey" json:"id"`
Name string `gorm:"size:100;not null" json:"name"`
}
```

2. **Repository** — `internal/repository/user.go`

```go
type UserRepository struct { db *gorm.DB }
func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db} }
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) { ... }
```

3. **Service** — `internal/service/user.go`

```go
type UserService struct { repo *repository.UserRepository }
func NewUserService(repo *repository.UserRepository) *UserService { return &UserService{repo} }
func (s *UserService) Get(ctx context.Context, id uint) (*model.User, error) { ... }
```

4. **Handler** — `internal/handler/user.go`

```go
type UserHandler struct { svc *service.UserService }
func (h *UserHandler) Get(c *gin.Context) {
// 绑定参数 → 调用 svc → response.Success / response.Error
}
```

5. **注册** — 在 `Dependencies` 加字段，在 `NewServer` 装配，在 `router.go` 增加 `registerUserRoutes`

### 错误处理规范

- Service 层遇到业务错误直接 `return nil, errcode.ErrXxx`
- Handler 层一律通过 `response.Error(c, err)` 响应
- 不要在每个 handler 里写重复的 `c.JSON(500, ...)`

### 日志规范

```go
logger.Info(ctx, "message", "key1", val1, "key2", val2)
logger.Error(ctx, "message", "err", err)
```

- `ctx` 传入 request context，自动携带 `rid`（Request ID）
- 格式为 JSON（slog JSONHandler），方便日志采集系统解析

### 避免的做法

- ❌ Handler 里写 SQL / 业务逻辑
- ❌ Service 里 import `gin` 或直接操作 `*gin.Context`
- ❌ Repository 以外的地方写 GORM 调用
- ❌ 在包之间产生循环依赖（Go 编译器直接拒绝）

---

## 依赖

| 库                                                                     | 用途                        |
|-----------------------------------------------------------------------|---------------------------|
| [gin-gonic/gin](https://github.com/gin-gonic/gin)                     | HTTP 框架                   |
| [spf13/viper](https://github.com/spf13/viper)                         | 配置管理                      |
| [gorm.io/gorm](https://gorm.io/) + MySQL driver                       | ORM                       |
| [redis/go-redis](https://github.com/redis/go-redis)                   | Redis 客户端                 |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt)                   | JWT 签发与校验                 |
| [go-playground/validator](https://github.com/go-playground/validator) | 参数校验                      |
| [google/uuid](https://github.com/google/uuid)                         | UUID 生成（Request ID / JTI） |
| [natefinch/lumberjack](https://github.com/natefinch/lumberjack)       | 日志文件轮转                    |
| [golang.org/x/text](https://pkg.go.dev/golang.org/x/text)             | Accept-Language 解析        |
