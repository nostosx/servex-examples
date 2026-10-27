# servex-examples

[English](README.md) | 简体中文

组合 servex 与 servex-contrib 的 HTTP、gRPC 可运行示例。
独立模块为 `github.com/nostosx/servex-examples`，要求 Go 1.27.3。
gRPC 示例使用标准 health API，无需生成 protobuf 代码。

v1.1.0 源码快照使用已发布的 servex v1.2.0 和 grpcx v1.1.0。快照版本标识示例源码及依赖，
不代表稳定的库 API。

## 本地运行

使用 Go 1.27.3，独立克隆本仓库即可。依赖从 Go 模块代理下载，
无需同级仓库或本地替换：

```sh
git clone https://github.com/nostosx/servex-examples.git
cd servex-examples
GOWORK=off go mod download
```

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

指定 gRPC 端口时，客户端目标需带 `grpc://`：

```sh
GOWORK=off go run ./cmd/grpc-server -addr 127.0.0.1:9100
GOWORK=off go run ./cmd/grpc-client -target grpc://127.0.0.1:9100
```

使用 Ctrl+C 停止各服务。两个示例均使用明文连接，没有应用鉴权。
请保持回环监听；对外提供改造后的服务前，应配置认证、授权和 TLS。

## 检查

```sh
make check
make lint
make vulncheck
```

CI 编译四个程序并执行检查，不依赖共享服务。servex、grpcx 和 kitx
分别使用已发布的 v1.2.0、v1.1.0 和 v1.0.0，无需同级源码检出或本地替换。
本模块沿用 grpcx 的依赖版本，库不反向依赖示例。
依赖更新需要通过独立构建，随后发布新的源码快照。
库旁的 godoc `Example` 测试继续保留。本项目没有自动发布工作流。
工具安装与检查说明见 [CONTRIBUTING.md](CONTRIBUTING.md)，可重现快照及依赖更新见
[RELEASING.md](RELEASING.md)，私密漏洞报告见 [SECURITY.md](SECURITY.md)，
变更记录见 [CHANGELOG.md](CHANGELOG.md)。采用 [MIT 许可证](LICENSE)。
