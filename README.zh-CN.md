# servex-examples

[English](README.md) | 简体中文

组合 servex 与 servex-contrib 的 HTTP、gRPC 可运行示例。
独立模块为 `github.com/nostosx/servex-examples`，要求 Go 1.27.3。
gRPC 示例使用标准 health API，无需生成 protobuf 代码。

首个源码快照计划为 v1.0.0，尚未发布。快照版本标识示例源码及依赖，
不代表稳定的库 API。

## 本地运行

将尚未发布的项目放在同一级，由 `go.mod` 中的相对 `replace` 解析依赖：

```text
nostosx/
├── servex/
├── servex-contrib/
└── servex-examples/
```

本模块替换尚未发布的 core 和 contrib transport 模块；kitx 使用发布的 v1.0.0，
由 Go 模块代理下载，无需本地检出。
从本目录分别在两个终端启动 HTTP 服务端和客户端：

```sh
GOWORK=off go run ./cmd/http-server
GOWORK=off go run ./cmd/http-client
```

默认监听 127.0.0.1:8000，客户端输出 `alice`。
`GET /v1/users/7` 返回 `{"id":7,"name":"alice"}`；
ID 为零时返回结构化的 InvalidArgument 错误。
`/readyz` 跟随应用生命周期更新就绪状态。
请求体上限为 64 KiB；HTTP 服务配置五秒请求头超时、十秒读写超时和一分钟空闲超时。

gRPC 同样分别启动：

```sh
GOWORK=off go run ./cmd/grpc-server
GOWORK=off go run ./cmd/grpc-client
```

默认监听 127.0.0.1:9000；客户端检查标准 health 服务并输出 `SERVING`。
状态不是 `SERVING` 时，客户端返回错误。
两个客户端均设置十秒超时；服务端通过 `app.SignalContext` 响应 SIGINT/SIGTERM。
服务端用 `-addr` 指定监听地址，客户端用 `-target` 指定目标：

```sh
GOWORK=off go run ./cmd/http-server -addr 127.0.0.1:8100
GOWORK=off go run ./cmd/http-client -target http://127.0.0.1:8100
```

## 检查

```sh
make check
make lint
make vulncheck
```

CI 将 core、contrib 放在同级目录，编译四个程序并执行检查，不依赖共享服务。
本模块沿用 grpcx 的依赖版本，库不反向依赖示例。
core 和 contrib 发布后再选择实际版本、移除对应替换、tidy 并重新检查。
库旁的 godoc `Example` 测试继续保留。本项目没有发布工作流。
工具安装与检查说明见 [CONTRIBUTING.md](CONTRIBUTING.md)，可重现快照及依赖更新见
[RELEASING.md](RELEASING.md)，私密漏洞报告见 [SECURITY.md](SECURITY.md)，
变更记录见 [CHANGELOG.md](CHANGELOG.md)。采用 [MIT 许可证](LICENSE)。
