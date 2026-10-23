# Contributing to servex-examples

Keep each example focused on one composition of public APIs. Programs belong in
`cmd/`; shared helpers belong in `internal/` only when multiple commands use them.
Libraries keep their own godoc Example tests. Comments are English; keep both
READMEs aligned with behavior.

Clone this repository on its own and use Go 1.27.3 and GNU Make with a POSIX
shell (Git Bash on Windows). The Makefile sets `GOWORK=off`; core, contrib and
kitx v1.0.0 resolve through the Go module proxy. No sibling checkouts or local
replacements are needed. Do not commit a go.work file.

```sh
make check
make lint
make vulncheck
```

`make check` rejects nonempty gofmt, modernizer and tidy diffs, runs vet, builds
all commands and runs race tests with randomized order and disabled test caching.
`make test` runs ordinary tests; `make race` runs the race suite separately.
`make fmt`, `make fix` and `make tidy` explicitly modify files.
CI jobs are independent. All checks use Go 1.27.3, with golangci-lint v2.15.0
built using that version. Install the check tools outside the module:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
go install golang.org/x/vuln/cmd/govulncheck@v1.9.0
```

Put the Go bin directory on PATH. Do not add these tools to go.mod. Make variables
`GO`, `GOFMT`, `GOLANGCI_LINT` and `GOVULNCHECK` support explicit tool paths.
Select Go through process-local GOROOT/PATH and GOTOOLCHAIN; do not edit global
Go settings.

Test observable behavior, error paths and cleanup without relying on sleeps or
internal call counts. Join goroutines before returning from a test; report
worker errors to the owning test. Use temporary directories and local listeners.

Automated tests use isolated local HTTP and gRPC fixtures and do not need shared
services. Keep listeners on loopback for demos. The programs do not implement
application authentication or TLS; add them before exposing an adapted service.
Never commit credentials or personal configuration.

Describe behavior changes and checks in pull requests. Record user-visible
changes in [CHANGELOG.md](CHANGELOG.md). Dependency updates also
follow [RELEASING.md](RELEASING.md). Select published versions and verify a
standalone checkout before updating dependencies. For private security reports,
see [SECURITY.md](SECURITY.md). Contributions use the [MIT license](LICENSE).
