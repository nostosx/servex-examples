# servex-examples

English | [简体中文](README.zh-CN.md)

Runnable HTTP and gRPC applications composing servex and servex-contrib.
Go 1.27.3 module: `github.com/nostosx/servex-examples`. The gRPC programs use
the generated standard health API, so no protobuf generation is needed.

The v1.1.0 source snapshot uses published servex v1.2.0 and grpcx v1.1.0.
Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## Local setup

Use Go 1.27.3. Clone this repository on its own; dependencies download from
the Go module proxy, with no sibling checkouts or local replacements:

```sh
git clone https://github.com/nostosx/servex-examples.git
cd servex-examples
GOWORK=off go mod download
```

## HTTP

Run the server and client in separate terminals from this directory:

```sh
GOWORK=off go run ./cmd/http-server
GOWORK=off go run ./cmd/http-client
```

The server listens on 127.0.0.1:8000. The client prints `alice`.
`GET /v1/users/7` returns `{"id":7,"name":"alice"}`; ID zero returns a
structured InvalidArgument error. `/readyz` follows the application lifecycle.
Requests are limited to 64 KiB. The HTTP server sets a five-second header
timeout, ten-second read/write timeouts and a one-minute idle timeout.

## gRPC

Run these in separate terminals from the repository root as well:

```sh
GOWORK=off go run ./cmd/grpc-server
GOWORK=off go run ./cmd/grpc-client
```

The server listens on 127.0.0.1:9000. The client checks the standard health
service and prints `SERVING`. Its status follows application readiness.
The client returns an error for any status other than `SERVING`.
Both clients use a ten-second deadline. Servers stop on SIGINT/SIGTERM through
`app.SignalContext`; server flags use `-addr`, client flags use `-target`:

```sh
GOWORK=off go run ./cmd/http-server -addr 127.0.0.1:8100
GOWORK=off go run ./cmd/http-client -target http://127.0.0.1:8100
```

For a custom gRPC port, use a `grpc://` target:

```sh
GOWORK=off go run ./cmd/grpc-server -addr 127.0.0.1:9100
GOWORK=off go run ./cmd/grpc-client -target grpc://127.0.0.1:9100
```

Stop each server with Ctrl+C. Both examples use plaintext connections without
application authentication. Keep them on loopback; configure authentication,
authorization and TLS before exposing an adapted service.

## Checks

```sh
make check
make lint
make vulncheck
```

CI builds all four programs and runs checks without shared external services.
It resolves servex v1.2.0, grpcx v1.1.0 and kitx v1.0.0 from published modules without sibling
checkouts or local replacements. The module preserves grpcx dependency versions;
libraries do not depend on examples. Dependency updates must pass a standalone
build before a new source snapshot is published. There is no automatic
publishing workflow.
Library godoc Example tests remain with the APIs they demonstrate.
See [CONTRIBUTING.md](CONTRIBUTING.md) for tool installation and check details,
[RELEASING.md](RELEASING.md) for reproducible snapshots and dependency updates,
[CHANGELOG.md](CHANGELOG.md) for changes and [SECURITY.md](SECURITY.md) for
private reports. Licensed under the [MIT License](LICENSE).
