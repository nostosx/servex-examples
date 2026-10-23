# Changelog

No snapshot has been released yet. Snapshot versions identify example source
and dependencies; they do not establish a stable library API.

## Unreleased

## v1.0.0

Prepared first source snapshot; not yet published.

### Added

- HTTP and gRPC server/client pairs demonstrating application lifecycle,
  readiness, transport errors, request deadlines and graceful shutdown.
- Local regression tests and Makefile checks for Go 1.27.3, with independent
  cross-platform CI, lint and vulnerability jobs.

### Dependencies

- Use the published kitx v1.0.0 through servex.
- Keep local replacements for the unreleased servex and grpcx modules.
  Independently buildable snapshots require published versions; see
  [RELEASING.md](RELEASING.md).
