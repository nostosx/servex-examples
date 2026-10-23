# Changelog

Snapshot versions identify example source and dependencies; they do not
establish a stable library API.

## Unreleased

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
