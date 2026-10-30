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

Use Go 1.27.3 and Task v3.53.1 with a POSIX shell (Git Bash on Windows).
`Taskfile.yml` sets `GOWORK=off`. The released example resolves public modules
without a sibling checkout. This unreleased composition uses matching
servex-contrib source explicitly: clone it beside this repository, or set
`SERVEX_CONTRIB_SOURCE` to its absolute path. The helper creates an external
temporary modfile; canonical go.mod/go.sum contain no local replacements or
invented peer checksums. Do not commit a go.work file.

```sh
bash .github/scripts/check-source.sh check
bash .github/scripts/check-source.sh lint
bash .github/scripts/check-source.sh vulncheck
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
`bash .github/scripts/check-source.sh test INTEGRATION=true` runs the real
application composition with race detection and rejects missing configuration,
absent/skipped/failed or unfinished integration tests. Supply all seven variables:
`SERVEX_EXAMPLES_IT_POSTGRES_DSN`, `SERVEX_EXAMPLES_IT_REDIS_ADDR`,
`SERVEX_EXAMPLES_IT_S3_ENDPOINT`, `SERVEX_EXAMPLES_IT_S3_BUCKET`,
`SERVEX_EXAMPLES_IT_S3_ACCESS_KEY`, `SERVEX_EXAMPLES_IT_S3_SECRET_KEY`, and
`SERVEX_EXAMPLES_IT_ETCD_ENDPOINTS`. Use owned disposable backends and a unique
bucket; the test runs migrations and creates/deletes its fixture bucket. It
exercises two Runtime instances sharing SQL/Redis, jobs, S3 and etcd. Ordinary
unit checks do not require these services. CI provisions isolated containers
and always removes them. Integration-gate regression fixtures reject false passes.
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
services. Keep listeners on loopback for demos. gRPC retains the API_TOKEN
boundary; the original HTTP user route remains a public local demonstration.
Workspace routes require the typed Session principal and CSRF checks. Password
login uses an explicitly configured demonstration account; replace that account
lookup with application policy for a deployed service.
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

Source checks for this snapshot do not establish public module installation.
After coordinated publication, update the changed module requirements and verify
standalone consumers without temporary overrides or sibling replacements.
