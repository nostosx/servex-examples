# servex-examples development

Parent workspace instructions apply. Always load `golang-how-to` for Go work,
then the relevant Go skills. Read README.md, CONTRIBUTING.md and
.github/workflows/ci.yml before editing.

- Go 1.27.3; runnable commands live in cmd/, shared private helpers in internal/.
- Compose public APIs with explicit parameters. Do not add a DI container or
  move application wiring back into the libraries.
- Core, contrib and kitx dependencies use published versions without
  local replacements. Keep standalone builds reproducible with GOWORK=off;
  do not commit go.work files or sibling checkout requirements.
- Code comments are English. Preserve package-level Example tests in libraries.
- `make check` checks formatting, modernizer/tidy diffs, vet, builds and race
  tests with randomized order. `make lint` uses golangci-lint v2.15.0 with
  Go 1.27.3; `make vulncheck` uses govulncheck. `make test` is ordinary
  testing; `make race` runs the race suite. `make fmt`, `make fix` and
  `make tidy` modify files. See RELEASING.md for dependency/snapshot checks.
- Use isolated local HTTP/gRPC fixtures and loopback listeners; do not add
  dependencies on shared services. Never store credentials or personal config.
