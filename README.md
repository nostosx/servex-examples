# servex-examples

This development snapshot combines published servex v1.4.0 with unreleased
contrib composition APIs. Validate and run it against a matching servex-contrib
source checkout through the temporary source helper below.

English | [简体中文](README.zh-CN.md)

A single Go application generated with nostosx-cli, then extended with a user
query demonstrated over typed servex HTTP and generated gRPC. The module is
`github.com/nostosx/servex-examples`; use Go 1.27.3 and Task 3.53.1.
The historical v1.4.0 tag uses only published dependencies. The current
development tree requires matching contrib source until those APIs are released;
canonical module files contain no local replacements or go.work file.

## Layout and generation

```text
app.go                       monolith Run; reuses the single service assembly
config/config.toml           monolith process settings
services/user/config/config.toml  standalone service settings
cmd/                         thin command entries; cmd/server calls application.Run
domain/user/                 domain declarations and the user value
application/                 user/query/GetUser, notes use cases and shared identity
infrastructure/              configuration, server and RPC boundaries
services/user/               app.go, private HTTP/RPC adapters and cmd/user
api/proto/user/v1/            source contracts
api/gen/go/user/v1/           generated Go and gRPC code
```
`cmd/server` parses configuration and calls root `app.go`'s `Run(ctx, cfg)`,
which reuses `user.Run`. The standalone `services/user/cmd/user` entry calls
`user.Run` directly. The service app.go explicitly registers handlers with the
shared transport constructors. Both expose HTTP
and gRPC; there is no separate HTTP-only or gRPC-only server implementation.
HTTP and RPC adapters call the same application query. Domain/application code
does not import generated contracts or transports.

The repository was built using actual CLI `new`, `add context`, `add service`
and `add api` commands. See [CLI-GENERATION.md](CLI-GENERATION.md) for the exact
recipe and the application-owned extensions. Generated output is committed;
ordinary builds do not invoke the CLI or require Buf.

Configuration is loaded by contrib Viper through infrastructure/config.Load:
defaults < TOML file < environment < explicit flags. The default file is optional;
--config selects a required file. APP_ADDRESS, APP_LOG_LEVEL and APP_TOKEN override
file settings; APP_TOKEN takes precedence over API_TOKEN. The servex example also
supports APP_GRPC_ADDRESS. --log-level selects the Zap level (default info).
Monoliths load `config/config.toml`; the standalone service loads
`services/user/config/config.toml`. Run from the repository root because
paths are relative to the working directory. These files are independent; neither
entry merges the other process's configuration. Use --config to select a deployment path.
Both entries keep the same default ports; configure distinct addresses when running them together.

Composition roots create contrib Zap console text and optional JSON file logging, inject the core Logger into
the application and transport boundaries, and flush after shutdown.

Console records render modules and fields as [Module] message [key:value],
with [操作:…], [耗时:…], [状态:…] and [错误:…] summaries. Newlines and tabs
in messages/fields are escaped so one event occupies one physical line.
An explicit component selects the module; otherwise it uses the caller package.
Console records use local timestamps in YYYY-MM-DD HH:mm:ss format, uppercase levels,
short caller locations and readable durations. File records use one JSON object
per line, the same second-precision local timestamps and numeric durations in seconds. Business fields
remain separate in both outputs. Set log_file in the process TOML file, APP_LOG_FILE or
--log-file to enable JSON file output; an explicit empty --log-file disables it.
The default is console only. Enabled files rotate at 100 MiB, retain at most five
backups for up to seven days, and compress backups with lumberjack v2.2.1.
Give each service process its own file path; close the logger after its users stop.

## Run

The helper reads `SERVEX_CONTRIB_SOURCE` (default `../servex-contrib`) and
uses a temporary modfile to select that checkout. It never writes local
replacements into go.mod/go.sum. Keep the checkout versions aligned.


```sh
git clone https://github.com/nostosx/servex-examples.git
cd servex-examples
export GOWORK=off
# Use the matching contrib development revision from the same change set.
git clone https://github.com/nostosx/servex-contrib.git ../servex-contrib
export SERVEX_CONTRIB_SOURCE="$(cd ../servex-contrib && pwd)"
export API_TOKEN="$(openssl rand -hex 32)"
bash .github/scripts/check-source.sh -- go run ./cmd/server
```
`API_TOKEN` must contain at least 32 printable ASCII characters without whitespace.
The default listeners are `127.0.0.1:8080` and `127.0.0.1:9090`.
HTTP health and the user route are public local demonstrations; gRPC requires
`authorization: Bearer <API_TOKEN>`. Both listeners require literal loopback IPs.

In a second terminal, set the same token and run:

```sh
export GOWORK=off
# Set API_TOKEN to the same value as the server.
bash .github/scripts/check-source.sh -- go run ./cmd/http-client
bash .github/scripts/check-source.sh -- go run ./cmd/grpc-client
curl http://127.0.0.1:8080/v1/users/7
curl http://127.0.0.1:8080/readyz
```

The HTTP client prints `alice`; the gRPC health client prints `SERVING`.
`GET /v1/users/7` returns `{"id":7,"name":"alice"}` and nonpositive IDs
return a structured InvalidArgument error. The generated `user.v1.UserService/GetUser`
RPC returns the same data and rejects unauthenticated requests.

To run the independently deployable service instead of the combined host:

```sh
bash .github/scripts/check-source.sh -- go run ./services/user/cmd/user -addr 127.0.0.1:8100 -grpc-addr 127.0.0.1:9100
bash .github/scripts/check-source.sh -- go run ./cmd/http-client -target http://127.0.0.1:8100
bash .github/scripts/check-source.sh -- go run ./cmd/grpc-client -target grpc://127.0.0.1:9100
```

Stop with Ctrl+C. Readiness follows the application lifecycle. HTTP retains the
64 KiB request limit, five-second header timeout, ten-second read/write timeouts
and one-minute idle timeout. Both clients have a ten-second deadline.
Connections are plaintext; add TLS and application authorization before external use.

## Checks and contracts

```sh
bash .github/scripts/check-source.sh check
bash .github/scripts/check-source.sh lint
bash .github/scripts/check-source.sh vulncheck
task api
```

The eight common tasks are `check`, `test`, `build`, `lint`, `vulncheck`, `fmt`,
`fix` and `tidy`; bare `task` runs `check`. Use `task --list` to view them,
`task test RACE=true` for race detection, and `task test RACE=true COVER=true`
for race coverage in `coverage.out`. Check helpers are internal.
This repository also exposes `task api ACTION=generate|lint|check|breaking`;
`task api` defaults to check (including lint), and breaking requires `API_BASE`.

`check` includes normal and integration builds and shuffled race tests with local
HTTP/gRPC fixtures. It exercises both user adapters, gRPC authentication and shutdown.
Default checks need no external services. The explicitly enabled composition
integration test requires disposable backends as described below.
Contract development uses Buf 1.72.0 and pinned nostosx-cli v1.4.0; use `NOSTOSX`
to select a CLI binary built from a checkout. Run `task api ACTION=generate`, review the
generated diff, then `task api` after editing Proto sources.

The v1.1.0 tag remains a historical source snapshot; development uses the layout
described here. See [CONTRIBUTING.md](CONTRIBUTING.md), [RELEASING.md](RELEASING.md),
[CHANGELOG.md](CHANGELOG.md) and [SECURITY.md](SECURITY.md). Licensed under [MIT](LICENSE).

## Composed application

The unreleased composition is assembled in `services/user/app.go`. It keeps the
original public `/v1/users/{id}` and gRPC bearer token behavior, and adds a small
notes application behind browser login. Backend addresses are empty by default:
notes, browser sessions, pending OAuth states and cache stay in bounded local
memory. Login has no default username, password or trusted identity.

Use the source validation helper described above for these unreleased contrib
APIs. Generate your own password hash from stdin, then explicitly enable local
HTTP cookies and a demonstration account:

```sh
export API_TOKEN="$(openssl rand -hex 32)"
# Set DEMO_PASSWORD privately in your shell; printf does not add a newline.
export APP_COMPOSITION__DEMO_USERNAME="demo"
export APP_COMPOSITION__DEMO_PASSWORD_HASH="$(printf %s "$DEMO_PASSWORD" | bash .github/scripts/check-source.sh -- go run ./cmd/password-hash)"
export APP_COMPOSITION__INSECURE_COOKIES=true
bash .github/scripts/check-source.sh -- go run ./cmd/server
```

Keep `insecure_cookies=false` when a TLS terminator serves the application to a
browser. Session cookies are host-only, HttpOnly and SameSite=Lax. Password
sign-in rotates the opaque token; logout deletes server-side state. OIDC and
password login restore the same typed application Principal before routing.
All notes and object operations use that Principal to isolate users.

For a local command-line roundtrip, send the configured password as JSON to
`POST /auth/password` and save the response cookie, then use it with `POST
/v1/notes` (`{"text":"hello"}`), `GET /v1/notes` and `POST /auth/logout`.
A request without a valid session receives 401 on the protected routes.

| Operation | Route | Enabled when |
| --- | --- | --- |
| Password sign-in / sign-out | `POST /auth/password`, `POST /auth/logout` | Explicit demo account / any session |
| OIDC redirect / callback | `GET /auth/oidc`, `GET /auth/callback` | Complete OIDC configuration |
| Current identity | `GET /v1/me` | Valid browser session |
| Create / list notes | `POST /v1/notes`, `GET /v1/notes` | Valid browser session |
| Update / delete note | `PUT /v1/notes/{id}`, `DELETE /v1/notes/{id}` | Owner of the note |
| Background cache refresh | `POST /v1/jobs/refresh` | Redis and browser session |
| Last completed refresh | `GET /v1/jobs/last-run` | Browser session |
| Upload / list objects | `POST /v1/objects`, `GET /v1/objects` | S3 bucket and browser session |
| Download / delete listed key | `GET /v1/objects/content?key=...`, `DELETE /v1/objects/content?key=...` | S3 and owner of the key |
| Acquire / release lock probe | `POST /v1/coordination/probe` | Etcd endpoints and browser session |
| Process metrics | `GET /metrics` | Always, on the existing loopback listener |

`config/config.toml` and `services/user/config/config.toml` carry matching
`[composition]`, `[composition.oidc]` and `[composition.s3]` groups. The existing
file/env/flag precedence remains. For example, `APP_COMPOSITION__POSTGRES_DSN`,
`APP_COMPOSITION__REDIS_ADDRESS` and `APP_COMPOSITION__S3__BUCKET` override their
nested TOML values. Supply credentials through deployment environment or a
private configuration file; the checked-in files contain no credentials.

Setting `postgres_dsn` selects the GORM notes repository. The GORM Database owns
one connection pool: repositories borrow `.GORM()` and the explicit Goose
migration entry borrows `.SQL()`. Serving never migrates automatically. Run the
migration command before starting a configured PostgreSQL instance:

```sh
export APP_COMPOSITION__POSTGRES_DSN="postgres://..."
bash .github/scripts/check-source.sh -- go run ./cmd/migrate
bash .github/scripts/check-source.sh -- go run ./cmd/server
```

Run one migration process per deployment. The migration command creates the
notes table in the configured database, and does not create a database itself.
Setting `redis_address` switches Session, OAuth one-time state and cache to
borrow one Redis client; asynq borrows that same client. Its finite I/O timeouts
and disabled transport retries preserve uncertain session mutation outcomes.
The resource scope closes the client after its consumers stop. The `namespace`
separates this application's Redis keys and queue name from other applications.

`job_schedule` optionally configures asynq periodic enqueueing, for example
`"@every 1m"`. The worker records an idempotent timestamp in cache. This refresh
is explicitly requested or scheduled, and is independent of note transactions.
Deploy one scheduler per schedule set: multiple schedulers can enqueue duplicate
jobs. An in-process `scheduler/cron` hook is an alternative for process-local
callbacks; this example uses the Redis-backed scheduler and does not also run
a second periodic scheduler for the same task.

OIDC requires issuer, client ID, authorization/token URLs and redirect URL; add
a client secret when required by the provider. The callback verifies the ID token
and nonce, then derives the local user ID from the verified issuer and subject.
Display claims grant no additional permissions. Use Redis for Session and state
when serving multiple replicas. HTTP loopback providers are permitted only
with the explicit local development cookie option.

S3 uses the official AWS SDK v2. Configure bucket, region and credentials, plus
an optional endpoint for an S3-compatible service. The application does not
create buckets on startup. Uploads are limited to 64 KiB. User IDs are hashed
into key prefixes; List returns keys relative to the application namespace so
the exact returned key can be passed to Get or Delete. Clients URL-encode that
key in the `key` query parameter. The download path closes the returned body.
The etcd probe returns an acquisition revision and releases the lock without an
external write. Real protected writes must enforce the revision as a fencing
token in the protected resource.

The root registers ordinary `notes.NewService` and HTTP handler constructors
through `di/do.ProvideConstructors`, resolves the entry once, then passes concrete
dependencies to handlers. The container does not appear in application logic or
requests. SDK connections have one owner in the lifecycle scope. `Build` rolls
back constructor failures; `Runtime.Close` also works before `App.Start`.

Native HTTP order is request lifetime tracking → shared profile observability
and recovery → optional CORS → standard CrossOriginProtection → login/Session →
routing. Trusted browser origins are explicit in `allowed_origins`; they also
permit CSRF-checked writes from those origins. gRPC unary/stream handlers and
workers reuse the same profile and Prometheus meter. Stream chains never add
message validation, retry or per-message timeout implicitly.

Close rejects new HTTP work, joins consumers and in-flight native handlers,
then releases shared resources in reverse order. A caller's close deadline only
bounds its wait; each native transport receives its own shutdown budget. Cleanup
continues after the caller returns. A handler that never cooperates can keep
cleanup waiting; dependencies remain open until that handler exits.

The ordinary HTTP suite covers login, CRUD, token rotation, cross-origin
rejection, logout revocation and cleanup ordering without external services.
`TestIntegrationComposition` additionally uses two instances against disposable
PostgreSQL, Redis, S3 and etcd. Set `SERVEX_EXAMPLES_IT_POSTGRES_DSN`,
`SERVEX_EXAMPLES_IT_REDIS_ADDR`, `SERVEX_EXAMPLES_IT_S3_ENDPOINT`,
`SERVEX_EXAMPLES_IT_S3_BUCKET`, `SERVEX_EXAMPLES_IT_S3_ACCESS_KEY`,
`SERVEX_EXAMPLES_IT_S3_SECRET_KEY` and `SERVEX_EXAMPLES_IT_ETCD_ENDPOINTS` (comma
separated), then run the source helper with `test INTEGRATION=true`. Supply a
fresh bucket name: the test creates that bucket, deletes its own objects and
removes it. The test never drops a database or table.
