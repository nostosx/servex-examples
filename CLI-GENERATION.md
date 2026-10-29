# CLI generation and application extensions

This repository is one Go module, generated from the nostosx-cli v1.2.0
with the monolith layout. The CLI implementation, templates and tests live
in their own repository; normal builds never read sibling source.

## Create the foundation

Install CLI v1.2.0 outside this application module. Set NOSTOSX to the
nostosx executable's absolute path. Use a new
output directory: generation and additions reject overwriting existing files.
The commands below were used to establish this repository's foundation:

```sh
"$NOSTOSX" new servex-examples --module github.com/nostosx/servex-examples --layout monolith --transport http,grpc --output ./new-example
"$NOSTOSX" add context user --dir ./new-example
"$NOSTOSX" add service user --dir ./new-example
"$NOSTOSX" add api get-user --transport http --service user --method GET --path '/v1/users/{id}' --dir ./new-example
"$NOSTOSX" add api get-user --transport grpc --service user --dir ./new-example
```

These establish configuration/server/RPC infrastructure, a domain/application context,
service app.go and private adapters, Proto sources, Taskfile and CI. This version uses Task and has no smoke/Python tooling. Generated source belongs to
the application and can be adapted; it is not fetched again on startup.

## Application extensions

- domain/user holds the user value; application/user/query/get_user.go owns
  validation and lookup. Both HTTP and gRPC adapters call this query.
- GetUserRequest declares int32 id; GetUserResponse declares int32 id and string
  name. Run the CLI to generate api/gen/go; never edit generated Go by hand.
- services/user/internal/port/http retains the typed httpx route and original
  JSON/error behavior. Proto declares gRPC only; no unused HTTP gateway binding
  or second implementation of the query is generated.
- Root app.go owns the monolith Run entry and selects user.Run for this
  single-service host. services/user/app.go registers HTTP and gRPC explicitly;
  the standalone entry calls user.Run directly. Both share one service assembly.
- Both clients remain in cmd/. The HTTP server and gRPC server commands are
  replaced by the combined host and standalone service. Defaults are now
  HTTP 8080 / gRPC 9090, and the gRPC client requires API_TOKEN.
- The application customizes the generated HTTP body limit to retain 64 KiB.

## Develop and verify

Install Buf 1.72.0 on PATH for contract changes; generators are pinned by the
CLI. In this repository:

```sh
task api ACTION=generate NOSTOSX="$NOSTOSX"
task tidy
task check lint vulncheck
task api NOSTOSX="$NOSTOSX"
```

The Taskfile defaults contract operations to CLI v1.2.0 through go run, without
adding it to go.mod. Use the NOSTOSX override when validating unreleased CLI code.
No generation shell/Python script, smoke job, extra module or go.work is needed.
CI checks the actual checked-in application on Linux/macOS/Windows and checks
contract drift separately. Tests use temporary loopback listeners and local
fixtures; automated checks do not require paid APIs or shared services.

Current CLI templates keep process entries limited to configuration, signal context
and Run. Monolith assembly lives in root app.go; standalone assembly lives in
services/<name>/app.go. The CLI no longer generates AGENTS.md, CONTRIBUTING.md,
README.md, README.zh-CN.md, CHANGELOG.md, SECURITY.md or LICENSE. This example
repository maintains its own documentation and license.

The current scaffold defaults to contrib Viper and contrib Zap. config.Load
merges defaults, config.yaml, environment and explicit flags; app.go constructs,
injects and flushes the Logger. The example uses the same loading/logging files,
with no hand-written configuration engine or global logging backend.
