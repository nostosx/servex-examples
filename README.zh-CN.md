# servex-examples

[English](README.md) | 简体中文

使用 nostosx-cli 实际生成的单模块 Go 工程，再补入用户查询示例，展示 servex 类型化
HTTP 路由与生成的 gRPC 服务。模块为 `github.com/nostosx/servex-examples`，
使用 Go 1.27.3、Task 3.53.1；依赖公开发布的 servex v1.3.0 和 grpcx v1.2.0，
无需同级仓库、本地 replace 或 go.work。

## 目录与生成

```text
cmd/                         精简命令入口；cmd/server 调用 user.Run
domain/user/                 用户领域值
application/user/query/      HTTP 与 RPC 共用的 GetUser 查询
infrastructure/              配置、传输构造与 RPC 边界
services/user/               app.go、私有 HTTP/RPC 适配器、cmd/user
api/proto/user/v1/           契约源文件
api/gen/go/user/v1/          生成的 Go/gRPC 代码
```

`cmd/server` 与独立入口 `services/user/cmd/user` 解析配置后调用
`user.Run(ctx, cfg)`，服务的 `app.go` 使用共享传输构造函数并显式注册处理器。
两种入口均暴露 HTTP 和 gRPC，
不再分别维护 HTTP-only、gRPC-only 服务端。两个适配器调用同一个应用查询；
领域及应用层不依赖生成代码或传输层。

本仓库实际通过 CLI 的 `new`、`add context`、`add service`、`add api`
创建，再补入示例实现。完整步骤与定制范围见 [CLI-GENERATION.md](CLI-GENERATION.md)。
生成结果随源码保存，普通构建不执行 CLI，也不需要 Buf。

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
契约开发使用 Buf 1.72.0 和固定的 nostosx-cli v1.2.0；可通过 `NOSTOSX` 指向
从源码构建的 CLI。修改 Proto 后运行 `task api ACTION=generate`、审查生成差异并运行 `task api`。

v1.1.0 标签仍标识历史源码快照，当前开发版本使用上述结构。
参见 [贡献指南](CONTRIBUTING.md)、[发布指南](RELEASING.md)、
[变更记录](CHANGELOG.md) 和 [安全策略](SECURITY.md)。许可证为 [MIT](LICENSE)。
