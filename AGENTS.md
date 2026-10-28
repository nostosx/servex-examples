# servex-examples development

Parent workspace instructions apply. Always load `golang-how-to` for Go work,
then the relevant Go skills. Read README.md, CONTRIBUTING.md and
.github/workflows/ci.yml before editing.

- Go 1.27.3; one CLI-generated module. cmd/ holds thin entries; application/ owns
  example behavior, infrastructure/ owns configuration/lifecycle, and services/
  holds app.go, private adapters and standalone service entries.
- Compose public APIs with explicit parameters. Do not add a DI container or
  move application wiring back into the libraries.
- Core, contrib and kitx dependencies use published versions without
  local replacements. Keep standalone builds reproducible with GOWORK=off;
  do not commit go.work files or sibling checkout requirements.
- Code comments are English. Preserve package-level Example tests in libraries.
- `task check` checks formatting, modernizer/tidy diffs, vet, builds and race
  tests with randomized order. `task lint` uses golangci-lint v2.15.0 with
  Go 1.27.3; `task vulncheck` uses govulncheck. `task test` is ordinary
  testing; `task test RACE=true` runs the race suite and `COVER=true` adds coverage. `task fmt`, `task fix` and
  `task tidy` modify files. The common check/test/build/lint/vulncheck/fmt/fix/tidy tasks and contract api task are public; bare task runs check.
  See RELEASING.md for dependency/snapshot checks.
- Use isolated local HTTP/gRPC fixtures and loopback listeners; do not add
  dependencies on shared services. Never store credentials or personal config.

- Follow CLI-GENERATION.md for the actual new/context/service generation recipe.
  Keep domain/application free of generated contracts and transport imports.
  api/proto is authoritative; regenerate api/gen/go with task api ACTION=generate and run
  task api with Buf 1.72.0. api ACTION also accepts lint/check/breaking (default check);
  breaking requires API_BASE. Do not add smoke or Python tooling.
