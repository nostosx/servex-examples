# Publishing servex-examples

This repository contains runnable examples, not a library release series. There
is no automatic release or publishing workflow. kitx uses the published v1.0.0;
local replacements remain for unreleased servex and servex-contrib modules.
Go 1.27.3 is the toolchain requirement, not a project version. The first source
snapshot is planned as v1.0.0; its notes are prepared in `CHANGELOG.md` and
it has not been tagged or published.

## Reproducible source snapshots

For a source snapshot that retains local replacements, record the exact commit
of this repository and every required sibling repository. Use the README's
checkout layout and run `make check`, `make lint` and `make vulncheck` with
`GOWORK=off`. CI currently checks out the siblings' default branches for joint
development; their revisions are not a published compatibility guarantee.
A successful snapshot must record the actual sibling revisions tested.

## Adopt released dependencies

The initial snapshot targets v1.0.0 of kitx, the core and each imported contrib
module. kitx is already published; wait for the remaining releases before
removing replacements. Select versions independently for subsequent snapshots:

| Module | Purpose |
|---|---|
| `github.com/nostosx/kitx` | Service-discovery retry delays through servex |
| `github.com/nostosx/servex` | Core library |
| `github.com/nostosx/servex-contrib/transport/grpcx` | gRPC transport |

1. Keep kitx v1.0.0, or independently verify a newer published release. Publish
   the upstream core first, followed by any contrib modules
   against that core version. Confirm the selected versions can be downloaded;
   a module path or placeholder version is not proof of availability.
2. Replace each placeholder requirement with its selected version and remove
   only its corresponding local replacement. Dependency modules' replacements
   are not inherited by consumers. Run `make tidy` and inspect go.mod/go.sum.
3. Update CI to resolve these module versions without sibling checkouts, keeping
   optional development-against-siblings checks separate. Validate a fresh
   checkout without the sibling directories or a workspace file:

   ```sh
   make check
   make lint
   make vulncheck
   ```

4. Run each HTTP and gRPC server/client pair locally, verify the documented
   responses, and stop the servers with SIGINT or SIGTERM. Use unused loopback
   ports; no database, broker or shared infrastructure is required.
5. Review the prepared `v1.0.0` notes and add the release date when publishing.
   Update both READMEs and [CHANGELOG.md](CHANGELOG.md) with any changed setup
   or behavior. In the published source
   snapshot or release notes, record the examples revision, exact module
   versions, toolchain, check results and known limitations. Require the
   corresponding hosted CI jobs to pass before describing that snapshot as
   validated.

Tags or GitHub source releases, if maintainers choose to create them, identify
example snapshots. They do not establish a stable library API or publish
prebuilt binaries. Creating a tag or publishing a release is a separate
maintainer action.
