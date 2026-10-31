# servex-examples

[PostgreSQL 事务索引示例](POSTGRESQL.md)：组合业务写入、PGMQ、向量和全文索引，明确事务边界，并验收失败回滚与重复投递。



v1.6.0 快照组合已发布的 servex 与独立版本的 contrib 模块，提供可运行的
HTTP/gRPC、类型化 CQRS、浏览器登录和笔记应用。

[English](README.md) | 简体中文

使用 nostosx-cli 实际生成的单模块 Go 工程，再补入用户查询示例，展示 servex 类型化
HTTP 路由与生成的 gRPC 服务。模块为 `github.com/nostosx/servex-examples`，
使用 Go 1.27.3、Task 3.53.1。依赖通过公开 Go proxy 解析，
无需相邻仓库、本地 replace 或 go.work。

应用通过注入的 `cqrs.Query` 让 HTTP/gRPC 共享 GetUser 用例及 chain；服务 Query 只装配一次 `user.get_user`，并选择 CQRS 日志中间件。笔记 CRUD 也通过注入的 command/query 调用，使用稳定的 notes 操作名；认证与授权保留在入口。

应用结构采用服务内 `internal/application/command.go`、`query.go` 中的 `Command`、`Query`，统一装配稳定 operation 与类型化 CQRS chain；`app.go` 通过 `CommandDeps`/`QueryDeps` 构造依赖并注入协议适配器。业务 context 与可部署 service 独立，协议端口只做转换。

## 目录与生成

```text
app.go                       单体 Run，复用单服务装配
cmd/                         精简命令入口；cmd/server 调用 application.Run
domain/user/                 用户领域值
application/                 GetUser 查询、notes 用例和共享身份
infrastructure/              配置、传输构造与 RPC 边界
services/user/               app.go、私有 HTTP/RPC 适配器、cmd/user
api/proto/user/v1/           契约源文件
api/gen/go/user/v1/          生成的 Go/gRPC 代码
```

`cmd/server` 解析配置后调用根目录 `app.go` 的 `Run(ctx, cfg)`，后者复用
`user.Run`；独立入口 `services/user/cmd/user` 直接调用 `user.Run`。
服务的 `app.go` 使用共享传输构造函数并显式注册处理器。
两种入口均暴露 HTTP 和 gRPC，
不再分别维护 HTTP-only、gRPC-only 服务端。两个适配器调用同一个应用查询；
领域及应用层不依赖生成代码或传输层。

本仓库实际通过 CLI 的 `new`、`add context`、`add service`、`add api`
创建，再补入示例实现。完整步骤与定制范围见 [CLI-GENERATION.md](CLI-GENERATION.md)。
生成结果随源码保存，普通构建不执行 CLI，也不需要 Buf。

配置由 infrastructure/config.Load 使用 contrib Viper 读取，优先级为：
默认值 < TOML 配置文件 < 环境变量 < 显式参数。默认配置文件可缺省；
--config 指定的文件必须存在。APP_ADDRESS、APP_LOG_LEVEL、APP_TOKEN 覆盖文件配置，
APP_TOKEN 优先于原有的 API_TOKEN；servex 示例还支持 APP_GRPC_ADDRESS。
--log-level 选择 Zap 日志级别，默认 info。装配入口创建 contrib Zap 控制台文本和可选 JSON 文件 Logger，
通过核心 Logger 接口注入应用和传输边界，并在关闭后 flush。

单体入口默认读取 `config/config.toml`，独立服务入口默认读取
`services/user/config/config.toml`。路径相对于启动工作目录，请从仓库根目录运行。
每个进程只读取自己的文件，不自动合并另一个入口的配置；部署时可用
`--config "/etc/user/config.toml"` 指定挂载路径。
两种入口保留相同的默认监听端口；同时运行时，需分别修改配置或使用地址参数。

控制台以 [模块] 消息 [字段:值] 展示事件，常用字段显示为 [操作:…]、
[耗时:…]、[状态:…]、[错误:…]。消息及字段中的换行、制表符转义为文本，
每个事件只占一行；模块优先取 component，缺省使用调用者包名。
控制台使用本地时间（格式为 2026-10-29 17:13:50）、大写级别、简短调用位置和易读耗时；
文件每行写一个 JSON 对象，使用相同的秒级本地时间及以秒为单位的数值耗时。
两种输出都保留独立业务字段。通过进程 TOML 配置的 log_file、APP_LOG_FILE
或 --log-file 开启文件输出，显式传入空的 --log-file 可关闭文件输出。
默认只输出控制台；启用文件后，通过 lumberjack v2.2.1 按 100 MiB 轮转，
最多保留五个备份、七天并压缩。每个服务进程使用独立文件路径，所有使用者停止后关闭 Logger。

## 运行

直接使用公开依赖运行：


```sh
git clone https://github.com/nostosx/servex-examples.git
cd servex-examples
export GOWORK=off
export API_TOKEN="$(openssl rand -hex 32)"
go run ./cmd/server
```

`API_TOKEN` 至少包含 32 个不含空白的可打印 ASCII 字符。默认监听
`127.0.0.1:8080`、`127.0.0.1:9090`。HTTP 健康检查和用户路由是本地公开演示；
gRPC 要求 `authorization: Bearer <API_TOKEN>`。监听地址必须使用字面量 loopback IP。

另一个终端设置相同 token 后运行：

```sh
export GOWORK=off
# 将 API_TOKEN 设置为与服务端相同的值。
go run ./cmd/http-client
go run ./cmd/grpc-client
curl http://127.0.0.1:8080/v1/users/7
curl http://127.0.0.1:8080/readyz
```

HTTP 客户端输出 `alice`，gRPC 健康客户端输出 `SERVING`。
`GET /v1/users/7` 返回 `{"id":7,"name":"alice"}`；非正 ID 返回结构化
InvalidArgument 错误。生成的 `user.v1.UserService/GetUser` RPC 返回相同用户数据，
并拒绝未认证请求。

需要独立服务入口时，替代单体宿主运行：

```sh
go run ./services/user/cmd/user -addr 127.0.0.1:8100 -grpc-addr 127.0.0.1:9100
go run ./cmd/http-client -target http://127.0.0.1:8100
go run ./cmd/grpc-client -target grpc://127.0.0.1:9100
```

Ctrl+C 退出，就绪状态跟随应用生命周期。保留 64 KiB 请求体限制、五秒请求头超时、
十秒读写超时、一分钟空闲超时，以及客户端十秒截止时间。
连接为明文；对外部署前补充 TLS 和应用授权。

## 检查与契约

```sh
task check
task lint
task vulncheck
task api
```

统一公开八个入口：`check`、`test`、`build`、`lint`、`vulncheck`、`fmt`、`fix`、`tidy`；
直接运行 `task` 等同于 `task check`。`task --list` 查看入口，`task test RACE=true`
启用竞态检测，`task test RACE=true COVER=true` 同时生成 `coverage.out` 并打印覆盖率。
格式、现代化、依赖和 vet 检查作为内部步骤保留。
本仓库另有 `task api`，通过 `ACTION=generate|lint|check|breaking` 选择契约操作；
默认 check 同时执行 lint，breaking 必须提供 `API_BASE`。

`check` 包含默认及 integration 构建、随机顺序 race 测试，通过本地 HTTP/gRPC fixture
检查查询、认证和退出。默认检查不依赖外部服务；显式启用的完整组合集成测试
需要下文列出的隔离后端。
契约开发使用 Buf 1.72.0 和固定的 nostosx-cli v1.5.0；可通过 `NOSTOSX` 指向
从源码构建的 CLI。修改 Proto 后运行 `task api ACTION=generate`、审查生成差异并运行 `task api`。

v1.1.0 标签仍标识历史源码快照，当前开发版本使用上述结构。
参见 [贡献指南](CONTRIBUTING.md)、[发布指南](RELEASING.md)、
[变更记录](CHANGELOG.md) 和 [安全策略](SECURITY.md)。许可证为 [MIT](LICENSE)。

## 完整应用组合

完整组合示例统一装配在 `services/user/app.go`，保留公开的
`/v1/users/{id}` 和原有 gRPC bearer token 行为，并新增浏览器登录后的笔记应用。
默认所有后端地址为空：笔记、Session、待完成 OAuth state 和缓存使用有容量上限的
进程内存。没有默认用户名、密码或生产身份。

以下新增 contrib API 使用上文的源码验证入口运行。先从标准输入生成自己的密码
PHC hash，再显式启用本地 HTTP Cookie 和演示账号：

```sh
export API_TOKEN="$(openssl rand -hex 32)"
# 在自己的 shell 中私下设置 DEMO_PASSWORD，printf 不添加换行。
export APP_COMPOSITION__DEMO_USERNAME="demo"
export APP_COMPOSITION__DEMO_PASSWORD_HASH="$(printf %s "$DEMO_PASSWORD" | go run ./cmd/password-hash)"
export APP_COMPOSITION__INSECURE_COOKIES=true
go run ./cmd/server
```

浏览器通过生产 TLS 入口访问时保持 `insecure_cookies=false`。Session Cookie
具有 host-only、HttpOnly、SameSite=Lax 属性；密码登录轮换不透明随机 token，
退出时删除服务端状态。密码与 OIDC 登录恢复同一种业务 Principal，笔记及对象
操作均使用这个身份做用户隔离。

本地命令行验收可以将配置好的用户名、密码作为 JSON 发给
`POST /auth/password` 并保存响应 Cookie，再携带 Cookie 请求
`POST /v1/notes`（`{"text":"hello"}`）、`GET /v1/notes`、
`POST /auth/logout`。没有有效 Session 的受保护接口返回 401。

| 能力 | 路由 | 启用条件 |
| --- | --- | --- |
| 密码登录 / 退出 | `POST /auth/password`、`POST /auth/logout` | 显式配置演示账号 / 任意 Session |
| OIDC 跳转 / 回调 | `GET /auth/oidc`、`GET /auth/callback` | 完整 OIDC 配置 |
| 当前身份 | `GET /v1/me` | 有效浏览器 Session |
| 新建 / 列出笔记 | `POST /v1/notes`、`GET /v1/notes` | 有效浏览器 Session |
| 修改 / 删除笔记 | `PUT /v1/notes/{id}`、`DELETE /v1/notes/{id}` | 笔记所属用户 |
| 后台缓存刷新 | `POST /v1/jobs/refresh` | Redis 和浏览器 Session |
| 最近一次刷新 | `GET /v1/jobs/last-run` | 浏览器 Session |
| 上传 / 列出对象 | `POST /v1/objects`、`GET /v1/objects` | S3 bucket 和浏览器 Session |
| 下载 / 删除列出的 key | `GET /v1/objects/content?key=...`、`DELETE /v1/objects/content?key=...` | S3 和对象所属用户 |
| 获取 / 释放锁探测 | `POST /v1/coordination/probe` | Etcd endpoints 和浏览器 Session |
| 进程指标 | `GET /metrics` | 始终启用，使用现有 loopback 监听 |

`config/config.toml` 和 `services/user/config/config.toml` 的新增分组一致：
`[composition]`、`[composition.oidc]`、`[composition.s3]`。
文件、环境变量和显式参数的优先级保持一致。例如
`APP_COMPOSITION__POSTGRES_DSN`、`APP_COMPOSITION__REDIS_ADDRESS`、
`APP_COMPOSITION__S3__BUCKET` 分别覆盖对应的嵌套 TOML 值。
凭证通过部署环境变量或私有配置文件注入，仓库配置不包含凭证。

配置 `postgres_dsn` 后选择 GORM 笔记仓储。GORM Database 独占一个 SQL 连接池，
仓储借用 `.GORM()`，显式 Goose 迁移借用 `.SQL()`。服务启动不会自动迁移。
启动 PostgreSQL 版本之前先单独执行：

```sh
export APP_COMPOSITION__POSTGRES_DSN="postgres://..."
go run ./cmd/migrate
go run ./cmd/server
```

每次部署运行一个迁移进程。迁移在已配置的数据库内创建笔记表，不负责创建数据库。
配置 `redis_address` 后，Session、OAuth 一次性 state、缓存及 asynq 共用一个
Redis 客户端，均借用连接；资源作用域在消费者退出后关闭客户端。客户端使用有限
I/O 超时并关闭传输层自动重试，保留 Session 写入结果不确定时的真实错误语义。
`namespace` 隔离本应用的 Redis key 和队列名称。

`job_schedule` 可选启用 asynq 周期入队，例如 `"@every 1m"`。worker 幂等更新缓存
时间戳；该刷新由显式请求或调度触发，与笔记事务独立。每组任务部署一个 scheduler，
多实例 scheduler 可能重复入队。`scheduler/cron` 适合进程本地回调，是另一种选项；
本示例采用 Redis 周期队列，不会为同一任务重复启动两套周期调度。

OIDC 需要 issuer、client ID、授权和 token URL、redirect URL，提供商要求时再配置
client secret。回调验证 ID token 与 nonce，使用已验证 issuer 和 subject 派生
本地用户 ID；显示名等 claim 不授予额外权限。多副本必须使用共享的 Redis Session
与 state。HTTP loopback 提供商仅在显式启用本地开发 Cookie 选项时允许。

S3 使用官方 AWS SDK v2，配置 bucket、region、凭证；兼容服务可以另外配置 endpoint。
应用启动不创建 bucket。上传限制 64 KiB，用户 ID 的哈希形成对象 key 前缀；List
返回相对应用命名空间的 key，可以将返回值原样交给 Get 或 Delete。客户端将该 key
做 URL 编码后放入 `key` 查询参数；下载处理关闭 SDK 返回的 Body。
etcd 探测获取并释放锁，只返回 acquisition revision，不执行外部写入。实际受锁
保护的写入必须在目标资源强制校验该 revision，作为 fencing token。

装配根通过 `di/do.ProvideConstructors` 注册普通的 `notes.NewService` 和 HTTP
处理器构造函数，只在启动时解析一次入口，再直接传递依赖。业务和请求代码不使用
容器。SDK 资源只有一个生命周期所有者；构造失败会回滚，`Runtime.Close` 在
`App.Start` 之前调用同样能清理资源。

HTTP 顺序是请求生命周期跟踪 → 共享 profile 可观测与恢复 → 可选 CORS → 标准库
CrossOriginProtection → login/Session → 路由。`allowed_origins` 显式列出可信浏览器
来源，同时允许这些来源通过 CSRF 检查执行写入。gRPC unary、stream 和 worker
使用同一 profile 与 Prometheus meter；stream 链不会隐式添加消息校验、重试或
逐消息超时。

关闭时先拒绝新 HTTP 工作，等待消费者及所有原生 HTTP handler 退出，再按相反顺序
释放共享资源。调用方的关闭 deadline 只限制等待；每个原生传输有独立的 shutdown
预算。调用方超时后后台继续清理。永不退出的 handler 仍可能阻塞最终清理，共享依赖
会保留到该 handler 退出。

普通 HTTP 测试无需外部服务，覆盖登录、CRUD、token 轮换、跨源拒绝、退出撤销和
清理顺序。`TestIntegrationComposition` 另外用两个实例连接一次性 PostgreSQL、
Redis、S3 和 etcd。设置 `SERVEX_EXAMPLES_IT_POSTGRES_DSN`、
`SERVEX_EXAMPLES_IT_REDIS_ADDR`、`SERVEX_EXAMPLES_IT_S3_ENDPOINT`、
`SERVEX_EXAMPLES_IT_S3_BUCKET`、`SERVEX_EXAMPLES_IT_S3_ACCESS_KEY`、
`SERVEX_EXAMPLES_IT_S3_SECRET_KEY`、`SERVEX_EXAMPLES_IT_ETCD_ENDPOINTS`
（逗号分隔）后，运行 `task test INTEGRATION=true`。
bucket 必须使用全新名称：测试创建 bucket、删除自身对象并移除 bucket，不删除数据库或表。
