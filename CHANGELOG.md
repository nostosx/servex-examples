# Changelog

Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## Unreleased

- Add the unreleased servex-contrib/transport/client module required by the
  matching HTTP/gRPC backends. Refresh its actual version and checksum during
  the coordinated release.

- Migrate HTTP imports and dependencies to servex-contrib/transport/httpx.
  This source revision requires matching unreleased core and contrib backends;
  public-proxy consumer verification remains a coordinated-release gate.

## v1.2.0

Release date: 2026-10-28.

- Select published servex v1.3.0 and transport/grpcx v1.2.0.

- Consolidate public Task commands to check/test/build/lint/vulncheck/fmt/fix/tidy;
  race and coverage use `task test RACE=true COVER=true`, with checks kept internal.
  Contract operations share `task api ACTION=generate|lint|check|breaking`; check remains the default.

- Explicitly select HTTP and gRPC in the CLI scaffold recipe. Native HTTP and
  gRPC adapters share the user query; HTTP does not require a gateway binding.

- Adopt one actual CLI-generated application per repository: domain/application,
  shared infrastructure, service app.go files, private adapters and Proto
  contracts with generated Go. See CLI-GENERATION.md for generation and extensions.
- Share the user query between typed HTTP and generated GetUser gRPC adapters.
- **Breaking (example startup):** cmd/http-server and cmd/grpc-server are replaced by cmd/server and
  services/user/cmd/user; both start HTTP and gRPC. Defaults are HTTP 8080 / gRPC
  9090; gRPC requires API_TOKEN and both listeners require literal loopback IPs.

- Use config.Parse in commands and service.Run(ctx, cfg) in app.go for shared
  assembly. Register only mounts handlers; server.New only constructs transports.

- Development tooling: replace Make entry points with Task v3.53.1 in local
  checks, CI and contribution/release instructions. Use `task check`, `task lint`
  and `task vulncheck`; check coverage, build tags and failure gates are preserved.

## v1.1.0

Release date: 2026-10-27.

- Require published servex v1.2.0 and servex-contrib transport/grpcx v1.1.0.
- http-server registers its handler with `httpx.RouteFunc` instead of a type
  assertion, and both servers use the `With*` option names of servex v1.2.0
  and grpcx v1.1.0.
- Clarify terminal usage, custom gRPC target URLs and the examples' plaintext,
  unauthenticated connection boundary.

## v1.0.0

Release date: 2026-10-23.

### Added

- HTTP and gRPC server/client pairs demonstrating application lifecycle,
  readiness, transport errors, request deadlines and graceful shutdown.
- Local regression tests and Makefile checks for Go 1.27.3, with independent
  cross-platform CI, lint and vulnerability jobs.

### Dependencies

- Use the published kitx v1.0.0 through servex.
- Require published servex v1.0.0 and the following contrib modules at v1.0.0:
  `github.com/nostosx/servex-contrib/transport/grpcx`.
- Resolve dependencies through the Go module proxy without sibling checkouts
  or local replacements.
