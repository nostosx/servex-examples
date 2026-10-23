# servex-examples

English | [简体中文](README.zh-CN.md)

Runnable HTTP and gRPC applications composing servex and servex-contrib.
Go 1.27.3 module: `github.com/nostosx/servex-examples`. The gRPC programs use
the generated standard health API, so no protobuf generation is needed.

The first source snapshot is planned as v1.0.0 and has not been published.
Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## Local setup

Keep the local, unreleased projects side by side:

```text
nostosx/
├── servex/
├── servex-contrib/
└── servex-examples/
```

The module explicitly replaces the unreleased core and contrib transport module.
kitx v1.0.0 downloads from the Go module proxy; no local checkout is needed.
The module preserves grpcx dependency versions; libraries do not depend on examples.

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

## Checks

```sh
make check
make lint
make vulncheck
```

CI checks out servex and servex-contrib next to this project, builds all four
programs and runs checks without shared external services. After the core and
contrib modules are released, select versions and remove the corresponding
replacements in this module, then tidy and check again. There is no publishing
workflow.
Library godoc Example tests remain with the APIs they demonstrate.
See [CONTRIBUTING.md](CONTRIBUTING.md) for tool installation and check details,
[RELEASING.md](RELEASING.md) for reproducible snapshots and dependency updates,
[CHANGELOG.md](CHANGELOG.md) for changes and [SECURITY.md](SECURITY.md) for
private reports. Licensed under the [MIT License](LICENSE).
