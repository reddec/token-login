# Repository Guidelines

## Project Overview

Token-login is a **forward-auth server** for token-based authorization. It provides an authorization flow based on API keys — reverse proxies (Caddy, Nginx, Traefik, Kubernetes ingress) delegate auth decisions to token-login's auth endpoint. Includes a Vue 3 admin UI for managing tokens, supports SQLite/Postgres storage, and offers three login methods (Basic, OIDC, Proxy).

## Architecture & Data Flow

One HTTP server, wired in `cmd/token-login/main.go`:

|Server|Default port|Purpose|
|---|---|---|
|HTTP|`:8080`|Admin UI + REST API (`/api/v1/`) + forward-auth (`/auth`) + health (`/health`)|

**Auth flow**: Reverse proxy sends request metadata (host, path, token) to the server → `web.AuthHandler` parses the key, looks up cache by `KeyID`, validates via `types.AccessKey.Valid()` (globs + SHA3-384 constant-time compare) → returns 204 with `X-User`, `X-Token-Hint`, and custom headers on success, 401 on failure.

**Data flow**: `dbo.Store` (sqlc) → `cache.Cache` (in-memory, `map[types.KeyID]*cache.Token`, `sync.RWMutex`, polled every `cache.ttl`) → `web.AuthHandler`. Stats flow: `web.Hit` channel (buffer `stats.buffer`) → `plumbing.SyncStats` (aggregates hits + last-access time per interval) → DB transaction via `Store.UpdateStats`.

**Wiring** (`cmd/token-login/main.go:run()`):
1. `open.Open()` opens DB + runs versioned migrations + data migration
2. `cache.New(store)` + initial `SyncKeys()`
3. `server.New(store)` → `api.NewServer(srv)` (ogen-generated)
4. `srv.OnRemove(keysCache.Drop)`, `srv.OnUpdate(keysCache.SyncKey)` — cache invalidation hooks
5. Goroutines: cache polling (`PollKeys`), stats sync (`plumbing.SyncStats`)
6. One `chi` router started via `Server.Run()` (graceful shutdown with TLS/mTLS support)
7. `multierror.Group` collects goroutine errors; `signal.NotifyContext` for graceful shutdown

## Key Directories

|Directory|Purpose|
|---|---|
|`cmd/token-login/`|Single-file entrypoint (`main.go`): CLI config parsing, startup wiring, all login middlewares|
|`internal/dbo/`|Database access layer — `store.go` (Store interface + domain types), `sqlite/` and `postgres/` engine packages (sqlc-generated code + adapters), `open/` (factory + migration runner)|
|`internal/cache/`|In-memory token cache keyed by `types.KeyID`, polled from DB, supports patch/drop/find|
|`internal/plumbing/`|Async stats persistence (buffered channel → periodic DB transactions via `Store.UpdateStats`)|
|`internal/server/`|API handler implementing the ogen-generated `api.Handler` interface — CRUD tokens, scoped to user from context|
|`internal/types/`|Cryptographic types: `Key` (40 bytes: 8 KeyID + 32 secret), `AccessKey` (globs + hash validation), `Headers`|
|`internal/utils/`|Context user helpers (`WithUser`/`GetUser`), flash-cookie helpers (base64, HttpOnly, 10s TTL)|
|`internal/redisstore/`|Redis-backed session store for OIDC login (wraps redigo pool)|
|`api/`|OpenAPI-generated code (ogen) — handler interface, client, JSON codec, validators; re-generate via `go generate`|
|`web/`|`public.go`: `//go:embed admin-ui/dist` → `fs.FS`. `private.go`: `AuthHandler` — the forward-auth HTTP handler|
|`web/admin-ui/`|Vue 3 SPA with Vite, shadcn-vue, Tailwind CSS 4; API client codegen via `@hey-api/openapi-ts`|
|`examples/`|Docker Compose and Kubernetes deployment examples for Caddy, Nginx, Traefik|
|`docs/`|Logo SVG|
|`.github/workflows/`|PR: lint + test + migration test; Release: push tag → GoReleaser|
|`openapi.yaml`|OpenAPI 3.0.3 spec — source of truth for both Go and TypeScript API codegen|

## Development Commands

```bash
make test        # go test -v ./...        (all packages)
make lint        # golangci-lint run       (config: .golangci.yml)
make snapshot    # goreleaser --snapshot + docker tag as :1 and :snapshot
make local       # goreleaser -f .goreleaser.local.yaml --clean  (local build)
make gen         # go generate ./...       (ogen + sqlc codegen)
cd web/admin-ui && npm run dev     # Vite dev server at :5050
cd web/admin-ui && npm run build   # type-check + build Vue SPA to dist/
cd web/admin-ui && npm run openapi # regenerate TS API client from openapi.yaml
cd web/admin-ui && npm run type-check  # vue-tsc type checking
cd web/admin-ui && npm run lint    # oxlint + eslint
cd web/admin-ui && npm run format  # prettier
```

Code generation triggers:
- `api/generate.go`: `//go:generate go tool ogen … openapi.yaml` → `api/oas_*_gen.go`
- `internal/dbo/store.go`: `//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc generate` → `internal/dbo/sqlite/*.sql.go`, `internal/dbo/postgres/*.sql.go`

## Code Conventions & Common Patterns

### Generating code
- **OpenAPI spec** (`openapi.yaml`) is the single source of truth for the REST API. Two codegen pipelines consume it:
  - **Go server**: `ogen` generates handler interface, client, JSON codec, validators → `api/`. Config: `.ogen.yaml` (convenient_errors off, server + client paths only).
  - **TypeScript client**: `@hey-api/openapi-ts` with `@hey-api/client-fetch` adapter → `web/admin-ui/src/api/`. Run via `npm run openapi`.
- **Database**: `sqlc` (type-safe SQL codegen) generates from SQL queries in `internal/dbo/sqlite/queries/` and `internal/dbo/postgres/queries/`. `open.Open()` runs versioned migrations on startup via `sql-migrate`. Sqlc config: `internal/dbo/sqlc.yaml` (v2, two engines, JSON tags, custom type overrides for `types.KeyID`/`types.Headers`).

### Server implementation
- `internal/server.Server` implements `api.Handler` (ogen-generated interface) and wraps `dbo.Store`.
- All CRUD methods scope queries to the current user via `utils.GetUser(ctx)`. This enforces multi-user isolation.
- Mutation hooks (`OnUpdate`/`OnRemove` callback lists) notify the cache layer of changes.

### Middleware pattern
Login methods implement `func(chi.Router) func(http.Handler) http.Handler`:
```go
func (cfg *Basic) createMiddleware(router chi.Router) func(http.Handler) http.Handler
func (cfg *OIDC) createMiddleware(ctx context.Context, router chi.Router) func(http.Handler) http.Handler
func (pa *ProxyAuth) createMiddleware(router chi.Router) func(http.Handler) http.Handler
```
Each middleware extracts/validates user identity, then calls `utils.WithUser(ctx, user)` and invokes the wrapped handler. The admin router mounts login-specific endpoints (e.g., `/oauth/logout` for OIDC logout redirect).

### Context-based user injection
```go
ctx = utils.WithUser(ctx, "alice")    // store user in context
user := utils.GetUser(ctx)             // retrieve (defaults to "anonymous")
```
This is the sole mechanism for passing user identity through the API layer. Server handlers pull the user from context; no separate auth header parsing in the API layer.

### Error handling
- Uses `github.com/go-faster/errors` and `go.uber.org/multierr` for error wrapping/joining.
- `hashicorp/go-multierror` for goroutine error collection in `run()`.
- ogen generates "convenient errors" disabled — errors are plain `error` returns from the handler interface.
- `defer` of functions that return errors (e.g. `defer conn.Close()`, `defer store.Close()`) is normal and expected — no need to wrap or explicitly ignore the returned error in these cases.

### CLI config
- `github.com/jessevdk/go-flags` struct-based configuration with env-var and CLI overrides.
- Version/commit/date baked via `-ldflags` at build time.
- Config struct in `main.go` uses `group`/`namespace`/`env-namespace` tags for hierarchical option groups.

### Web embedding
- Vue 3 SPA built into `web/admin-ui/dist/`, embedded via `//go:embed admin-ui/dist` in `web/public.go`.
- Served by `http.FileServerFS(web.Assets())` on the admin router, mounted after the API prefix.
- `vite.config.ts` uses `base: './'` for relative asset paths to work under any mount point.

### Frontend patterns
- **State**: Pinia 3 stores (setup-function / Composition API style). Two stores: `credential` (transient token pass-through) and `preferences` (dark mode + sidebar, persisted to localStorage). One composable: `useNotifications()` for reactive toast queue.
- **Routing**: vue-router 5 with hash history (`createWebHashHistory`). Lazy-loaded views via dynamic `import()`.
- **API client**: Auto-generated by `@hey-api/openapi-ts` with `@hey-api/client-fetch`. Client configured with `throwOnError: true` in `main.ts`; SDK calls wrapped in try/catch. Errors decoded via `getErrorMessage()` helper.
- **UI**: shadcn-vue pattern — Reka UI headless primitives + CVA variants + `cn()` from tailwind-merge/clsx. All components use `<script setup lang="ts">`.

### Security patterns
- Token keys: 40 bytes = 8-byte public `KeyID` (stored in DB) + 32-byte private payload (NOT stored in DB — only its SHA3-384 hash).
- Validation uses `crypto/subtle.ConstantTimeCompare` for hash comparison.
- Password auth uses bcrypt (`golang.org/x/crypto/bcrypt`).
- OWASP security headers added to the HTTP server in production mode (`X-Frame-Options`, `X-XSS-Protection`, `X-Content-Type-Options`, `Referrer-Policy`).
- Debug mode (`--debug.enable`) enables CORS and request logging.

### Naming conventions
- Package names are short, lowercase, single-word: `cache`, `dbo`, `plumbing`, `utils`, `types`.
- Test files use `_test` package suffix (e.g., `package server_test`, `package types_test`) for black-box testing.
- Generated files are named `oas_*_gen.go` (ogen) and `*.sql.go`/`db.go`/`models.go` (sqlc).

### Linting
`.golangci.yml` (v2 config) enables all linters minus ~20 exclusions. Tests are excluded (`run.tests: false`). Four formatters enabled: `gci`, `gofmt`, `gofumpt`, `goimports`. After any code change, run both:
- `go fix ./...` — applies stdlib modernization fixes (see https://go.dev/blog/gofix)
- `golangci-lint run` — enforces all enabled linters

## Important Files

|File|Role|
|---|---|
|`cmd/token-login/main.go`|Single-file entrypoint: config, CLI, startup wiring, all auth middlewares|
|`openapi.yaml`|OpenAPI 3.0.3 spec — 9 endpoints, 9 schemas (Project, Token, Credential, etc.). Base: `/api/v1`|
|`internal/dbo/store.go`|Store interface + domain types — the authoritative DB access contract|
|`internal/dbo/open/open.go`|`open.Open()` — DB open + migrations, supports SQLite and Postgres via URL scheme|
|`internal/cache/cache.go`|In-memory token cache; `FindByKey()` is on the auth hot path|
|`internal/types/access_key.go`|`AccessKey.Valid()` — core auth validation (glob matching + constant-time hash compare)|
|`web/private.go`|`AuthHandler` — the forward-auth HTTP endpoint|
|`web/public.go`|`//go:embed` for the Vue 3 SPA|
|`internal/server/server.go`|implements `api.Handler` (ogen) — CRUD tokens with per-user scoping|
|`api/generate.go`|`//go:generate` directive for ogen; exports `Prefix = "/api/v1"`|
|`.ogen.yaml`|ogen codegen config|
|`.goreleaser.yaml`|Multi-platform release (3 OS × 2 arch) + multi-arch Docker manifests|
|`Makefile`|Dev commands: lint, test, snapshot, gen|
|`migration-test.sh`|Forward-compatibility test script (DB migration from v1.1.0)|

## Runtime/Tooling Preferences

- **Go 1.26+** (see `go.mod`). Build with `CGO_ENABLED=0`, `-trimpath`.
- **Node 20**, **npm** — frontend build (`web/admin-ui`). Vue 3.5, Vite 8, TypeScript 6, Tailwind CSS 4.
- **Docker**: FROM scratch, binary ADDed to `/`. Exposes 8080. Volume `/data`. Default env: `DB_URL=sqlite:///data/token-login.sqlite?cache=shared`.
- **Package manager**: Go modules (`go.mod`), npm for frontend.
- **Codegen**: ogen (Go server/client from OpenAPI), sqlc (type-safe SQL from queries), `@hey-api/openapi-ts` (TypeScript client).
- **No external build system**: plain `Makefile` + `go generate`.

## Testing & QA

- **Test framework**: `testing` stdlib + `github.com/stretchr/testify` (assert/require).
- **Run**: `make test` or `go test -v ./...`.
- **Test DB**: In-memory SQLite (`file::memory:?cache=shared`) for unit tests — see `internal/server/server_test.go`.
- **Black-box tests**: Test packages use `_test` suffix (`server_test`, `types_test`).
- **HTTP tests**: `httptest.Server` for testing auth handler behavior.
- **Migration test**: Two layers:
  - Go test (`internal/dbo/open/migration_test.go`): uses `testcontainers-go` to launch old server containers (v1.0.0, v1.1.0, v1.2.0), populate via API, stop, then open with current code and verify data integrity. Tests both SQLite and Postgres backends.
  - Shell script (`migration-test.sh`): standalone E2E using docker+curl+jq, same validation flow.
- **CI**: PR workflow runs lint (`golangci-lint`), tests (`make test`), and migration test. Release workflow (on `v*` tag) runs GoReleaser — no test step.
