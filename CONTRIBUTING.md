# Contributing to servex-examples

Install [Task](https://taskfile.dev/docs/installation/) v3.53.1 outside the module,
matching CI, and put its binary on `PATH`:

```sh
go install github.com/go-task/task/v3/cmd/task@v3.53.1
```

Run `task --list` to discover tasks. Tool overrides accept environment variables or
Task arguments, for example `task check GO="/path with spaces/go"`.
Checks force `GOWORK=off`; `check`, `lint` and `vulncheck` do not apply source fixes.

Keep each example focused on one composition of public APIs. Thin process entries belong in `cmd/` and `services/<name>/cmd/`; example
behavior belongs in `application/`, configuration in `infrastructure/`, and
private adapters in `services/<name>/internal/port/`.
Libraries keep their own godoc Example tests. Comments are English; keep both
READMEs aligned with behavior.

Clone this repository on its own and use Go 1.27.3 and Task v3.53.1 with a POSIX
shell (Git Bash on Windows). `Taskfile.yml` sets `GOWORK=off`; core, contrib and
kitx v1.0.0 resolve through the Go module proxy. No sibling checkouts or local
replacements are needed. Do not commit a go.work file.

```sh
task check
task lint
task vulncheck
task api
```

`task check` rejects nonempty gofmt, modernizer and tidy diffs, runs vet, builds
all commands and runs normal/integration race tests with randomized order and disabled test caching.
The public tasks are `check`, `test`, `build`, `lint`, `vulncheck`, `fmt`,
`fix` and `tidy`. Both `task --list` and `task --list-all` show this public
surface; formatting, modernization, dependency and vet helpers are internal.
Bare `task` runs `check`. Run `task test RACE=true` for race detection or
`task test RACE=true COVER=true` to also write `coverage.out` and print function
coverage. Tests always shuffle and disable result caching. `check` reuses the
same test task with race detection enabled.
The additional `api` task manages this repository's Proto contracts.
`task fmt`, `task fix` and `task tidy` explicitly modify files.
CI jobs are independent. All checks use Go 1.27.3, with golangci-lint v2.15.0
built using that version. Install the check tools outside the module:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
go install golang.org/x/vuln/cmd/govulncheck@v1.9.0
```

Put the Go bin directory on PATH. Do not add these tools to go.mod. Task variables
`GO`, `GOFMT`, `GOLANGCI_LINT` and `GOVULNCHECK` support explicit tool paths.
Select Go through process-local GOROOT/PATH and GOTOOLCHAIN; do not edit global
Go settings.

Test observable behavior, error paths and cleanup without relying on sleeps or
internal call counts. Join goroutines before returning from a test; report
worker errors to the owning test. Use temporary directories and local listeners.

Automated tests use isolated local HTTP and gRPC fixtures and do not need shared
services. Keep listeners on loopback for demos. gRPC uses the generated API_TOKEN boundary; HTTP user routes remain public local
demonstrations. Add HTTP authorization and TLS before external deployment.
Never commit credentials or personal configuration.

Describe behavior changes and checks in pull requests. Record user-visible
changes in [CHANGELOG.md](CHANGELOG.md). Dependency updates also
follow [RELEASING.md](RELEASING.md). Select published versions and verify a
standalone checkout before updating dependencies. For private security reports,
see [SECURITY.md](SECURITY.md). Contributions use the [MIT license](LICENSE).

Use CLI-GENERATION.md for the scaffold recipe. Install Buf 1.72.0 before contract
work. Run task api ACTION=generate, review the generated diff and run task api. Both
actions use published CLI v1.4.0 by default and accept NOSTOSX as an absolute executable override. Keep transport/generated
imports out of domain and application; assemble dependencies through app.go.

Use `task api ACTION=lint` for lint alone; `task api` defaults to `ACTION=check`
and preserves both lint and generated-output checks. `task api ACTION=breaking API_BASE=<revision>` requires an explicit compatibility baseline. Invalid actions
and a missing breaking baseline fail before invoking the CLI.

The current snapshot uses published core/contrib dependencies. Validate standalone
consumers without temporary module overrides or sibling replacements.
