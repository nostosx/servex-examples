# Publishing servex-examples

Tags identify reproducible source snapshots of runnable examples. They do not
establish a stable library API or include prebuilt binaries. Go 1.27.3 is the
toolchain requirement, independent of the snapshot version. There is no
automatic release workflow.

## Published dependencies

The v1.3.0 snapshot uses these released modules, with no local replacements:

| Module | Version |
|---|---|
| `github.com/nostosx/servex` | v1.4.0 |
| `github.com/nostosx/servex-contrib/transport/grpcx` | v1.3.0 |
| `github.com/nostosx/kitx` | v1.0.0 |
| `github.com/nostosx/servex-contrib/config/viper` | v1.0.0 |
| `github.com/nostosx/servex-contrib/logger/zap` | v1.0.0 |
| `github.com/nostosx/servex-contrib/transport/httpx` | v1.0.0 |

For subsequent snapshots, publish and verify upstream core and contrib versions
first. Select each version explicitly, tidy go.mod/go.sum and inspect the diff.
Go does not inherit dependencies’ replace directives. Never use sibling
checkouts or a workspace file as release acceptance.

## Validate a snapshot

1. Start from a clean checkout on `main`. Review both READMEs, the selected
   versions and the dated entry in [CHANGELOG.md](CHANGELOG.md).
2. Run `task check`, `task lint`, `task vulncheck` and `task api` with Go 1.27.3.
   Confirm every hosted CI job passes at the exact commit to be tagged.
3. Copy or clone the project into a temporary directory with no sibling
   repositories. With `GOWORK=off`, download dependencies, run `go mod verify`
   and build all commands using `go build -mod=readonly ./...`.
4. Complete the normal and integration tests. Record the source revision,
   exact dependency versions, toolchain, check results and known limitations
   in the release notes. Do not describe unrun live integrations as validated.

Run the combined host or standalone user service and both clients on unused loopback ports with API_TOKEN. Verify
the documented responses and stop the servers with SIGINT or SIGTERM. These
checks do not require databases, brokers or shared infrastructure.

## Publish

After the checks pass, create a GitHub source release for the reviewed commit,
using the matching changelog section as its notes. The initial snapshot is
`v1.0.0`; later snapshots use new versions. Verify the remote tag resolves to
the checked commit and the source archive can be downloaded. Never move an
already published tag. Dependency updates require a new snapshot.
