# servex-examples

Current unreleased source moves the HTTP backend to servex-contrib/transport/httpx
and its shared outbound client to servex-contrib/transport/client.
It requires matching unpublished core/contrib sources; the declared adapter
versions prepare a coordinated release. Published tags retain their original
dependencies. Refresh versions and checksums before public-proxy acceptance.

English | [简体中文](README.zh-CN.md)

A single Go application generated with nostosx-cli, then extended with a user
query demonstrated over typed servex HTTP and generated gRPC. The module is
`github.com/nostosx/servex-examples`; use Go 1.27.3 and Task 3.53.1.
Snapshot v1.2.0 uses published servex v1.3.0 and grpcx v1.2.0, without sibling
checkouts, local replacements or a go.work file.

## Layout and generation

```text
cmd/                         thin command entries; cmd/server calls user.Run
domain/user/                 domain declarations and the user value
application/                 user/query/GetUser shared by both adapters
infrastructure/              configuration, server and RPC boundaries
services/user/               app.go, private HTTP/RPC adapters and cmd/user
api/proto/user/v1/            source contracts
api/gen/go/user/v1/           generated Go and gRPC code
```
`cmd/server` and `services/user/cmd/user` parse configuration and call
`user.Run(ctx, cfg)`. The service app.go explicitly registers handlers with the
shared transport constructors. Both expose HTTP
and gRPC; there is no separate HTTP-only or gRPC-only server implementation.
HTTP and RPC adapters call the same application query. Domain/application code
does not import generated contracts or transports.

The repository was built using actual CLI `new`, `add context`, `add service`
and `add api` commands. See [CLI-GENERATION.md](CLI-GENERATION.md) for the exact
recipe and the application-owned extensions. Generated output is committed;
ordinary builds do not invoke the CLI or require Buf.

## Run

```sh
git clone https://github.com/nostosx/servex-examples.git
cd servex-examples
export GOWORK=off
go mod download
export API_TOKEN="$(openssl rand -hex 32)"
go run ./cmd/server
```
`API_TOKEN` must contain at least 32 printable ASCII characters without whitespace.
The default listeners are `127.0.0.1:8080` and `127.0.0.1:9090`.
HTTP health and the user route are public local demonstrations; gRPC requires
`authorization: Bearer <API_TOKEN>`. Both listeners require literal loopback IPs.

In a second terminal, set the same token and run:

```sh
export GOWORK=off
# Set API_TOKEN to the same value as the server.
go run ./cmd/http-client
go run ./cmd/grpc-client
curl http://127.0.0.1:8080/v1/users/7
curl http://127.0.0.1:8080/readyz
```

The HTTP client prints `alice`; the gRPC health client prints `SERVING`.
`GET /v1/users/7` returns `{"id":7,"name":"alice"}` and nonpositive IDs
return a structured InvalidArgument error. The generated `user.v1.UserService/GetUser`
RPC returns the same data and rejects unauthenticated requests.

To run the independently deployable service instead of the combined host:

```sh
go run ./services/user/cmd/user -addr 127.0.0.1:8100 -grpc-addr 127.0.0.1:9100
go run ./cmd/http-client -target http://127.0.0.1:8100
go run ./cmd/grpc-client -target grpc://127.0.0.1:9100
```

Stop with Ctrl+C. Readiness follows the application lifecycle. HTTP retains the
64 KiB request limit, five-second header timeout, ten-second read/write timeouts
and one-minute idle timeout. Both clients have a ten-second deadline.
Connections are plaintext; add TLS and application authorization before external use.

## Checks and contracts

```sh
task check
task lint
task vulncheck
task api
```

The eight common tasks are `check`, `test`, `build`, `lint`, `vulncheck`, `fmt`,
`fix` and `tidy`; bare `task` runs `check`. Use `task --list` to view them,
`task test RACE=true` for race detection, and `task test RACE=true COVER=true`
for race coverage in `coverage.out`. Check helpers are internal.
This repository also exposes `task api ACTION=generate|lint|check|breaking`;
`task api` defaults to check (including lint), and breaking requires `API_BASE`.

`check` includes normal and integration builds and shuffled race tests with local
HTTP/gRPC fixtures. It exercises both user adapters, gRPC authentication and shutdown.
There are no smoke jobs, Python scripts or external-service requirements.
Contract development uses Buf 1.72.0 and pinned nostosx-cli v1.2.0; use `NOSTOSX`
to select a CLI binary built from a checkout. Run `task api ACTION=generate`, review the
generated diff, then `task api` after editing Proto sources.

The v1.1.0 tag remains a historical source snapshot; development uses the layout
described here. See [CONTRIBUTING.md](CONTRIBUTING.md), [RELEASING.md](RELEASING.md),
[CHANGELOG.md](CHANGELOG.md) and [SECURITY.md](SECURITY.md). Licensed under [MIT](LICENSE).
