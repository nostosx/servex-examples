# servex-examples development

Parent workspace instructions apply. Always load `golang-how-to` for Go work,
then the relevant Go skills. Read README.md, CONTRIBUTING.md and
.github/workflows/ci.yml before editing.

- Go 1.27.3; one CLI-generated module. cmd/ holds thin entries; application/ owns
  example behavior, infrastructure/ owns configuration/lifecycle, and services/
  holds app.go, private adapters and standalone service entries.
- Root app.go selects the monolith composition. Commands parse configuration,
  create a signal context and call Run; keep assembly in root or service app.go.
- Compose APIs with explicit parameters. The approved composition registers
  ordinary constructors through di/do once in service app.go. Keep containers
  out of business constructors and request paths; wiring stays in the application.
- Released snapshots use public dependencies. The unreleased composition uses
  .github/scripts/check-source.sh and matching contrib source through an external
  temporary modfile. Keep GOWORK=off and never commit go.work, source replacements
  or invented peer checksums; verify public installation after publication.
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

- Default process configuration uses contrib config/viper, including config/config.toml,
  APP_ overrides and explicit flags. app.go constructs contrib logger/zap console/file
  logging and injects the core Logger; keep globals and concrete backends out of domain/application.

- TOML process files are independent: config/config.toml for cmd/server and
  services/user/config/config.toml for the standalone service. Entry points pass
  their default path to config.Load; preserve file/env/flag precedence.

- Default contract acceptance (2026-10-29): task api passed in an independent
  copy with an empty GOMODCACHE, NOSTOSX/NOSTOSX_VERSION unset, Buf 1.72.0,
  GOPROXY=https://proxy.golang.org,direct and GOSUMDB=sum.golang.org. Public CLI
  v1.3.0 resolves to a7e193283bb9fff46fc6d17c946f1d0483b1fd28. A stale proxy
  .info 404 required the direct fallback; warm-cache or local-CLI checks alone
  do not establish cold-start availability. See RELEASING.md.
