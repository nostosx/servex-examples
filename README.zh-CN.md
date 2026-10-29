# servex-examples

本快照使用公开发布的 servex v1.4.0，以及独立的 contrib HTTP、配置和日志模块，
无需同级源码或本地 replace。

[English](README.md) | 简体中文

使用 nostosx-cli 实际生成的单模块 Go 工程，再补入用户查询示例，展示 servex 类型化
HTTP 路由与生成的 gRPC 服务。模块为 `github.com/nostosx/servex-examples`，
使用 Go 1.27.3、Task 3.53.1；依赖公开发布的 servex v1.4.0 和 grpcx v1.3.0，
无需同级仓库、本地 replace 或 go.work。

## 目录与生成

```text
app.go                       单体 Run，复用单服务装配
cmd/                         精简命令入口；cmd/server 调用 application.Run
domain/user/                 用户领域值
application/user/query/      HTTP 与 RPC 共用的 GetUser 查询
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

```sh
git clone https://github.com/nostosx/servex-examples.git
cd servex-examples
export GOWORK=off
go mod download
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
检查查询、认证和退出。没有 smoke job、Python 脚本或外部服务依赖。
契约开发使用 Buf 1.72.0 和固定的 nostosx-cli v1.4.0；可通过 `NOSTOSX` 指向
从源码构建的 CLI。修改 Proto 后运行 `task api ACTION=generate`、审查生成差异并运行 `task api`。

v1.1.0 标签仍标识历史源码快照，当前开发版本使用上述结构。
参见 [贡献指南](CONTRIBUTING.md)、[发布指南](RELEASING.md)、
[变更记录](CHANGELOG.md) 和 [安全策略](SECURITY.md)。许可证为 [MIT](LICENSE)。
