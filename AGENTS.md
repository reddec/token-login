# Repository Guidelines

## Project Overview

Token-login is a **forward-auth server** for token-based authorization. It provides an authorization flow based on API keys — reverse proxies (Caddy, Nginx, Traefik, Kubernetes ingress) delegate auth decisions to token-login's auth endpoint. Includes a Svelte admin UI for managing tokens, supports SQLite/Postgres/MySQL storage, and offers three login methods (Basic, OIDC, Proxy).

## Architecture & Data Flow

Two independent HTTP servers, wired in `cmd/token-login/main.go`:

| Server  | Default port | Purpose |
|---------|-------------|---------|
| Admin   | `:8080`     | Admin UI (embedded SPA via `//go:embed`) + REST API (`/api/v1/`) |
| Auth    | `:8081`     | Forward-auth endpoint (`/health` for readiness) |

**Auth flow**: Reverse proxy sends request metadata (host, path, token) to auth server → `web.AuthHandler` parses the key, looks up cache by `KeyID`, validates via `types.AccessKey.Valid()` (globs + SHA3-384 constant-time compare) → returns 204 with `X-User`, `X-Token-Hint`, and custom headers on success, 401 on failure.

**Data flow**: `ent.Client` (ORM) → `cache.Cache` (in-memory, `map[types.KeyID]*cache.Token`, `sync.RWMutex`, polled every `cache.ttl`) → `web.AuthHandler`. Stats flow: `web.Hit` channel (buffer `stats.buffer`) → `plumbing.SyncStats` (aggregates hits + last-access time per interval) → DB transaction.

**Wiring** (`cmd/token-login/main.go:run()`):
1. `ent.New()` opens DB + auto-migrates schema
2. `cache.New(store)` + initial `SyncKeys()`
3. `server.New(store)` → `api.NewServer(srv)` (ogen-generated)
4. `srv.OnRemove(keysCache.Drop)`, `srv.OnUpdate(keysCache.SyncKey)` — cache invalidation hooks
5. Goroutines: cache polling (`PollKeys`), stats sync (`plumbing.SyncStats`)
6. Two `chi` routers started via `Server.Run()` (graceful shutdown with TLS/mTLS support)
7. `multierror.Group` collects goroutine errors; `signal.NotifyContext` for graceful shutdown

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/token-login/` | Single-file entrypoint (`main.go`): CLI config parsing, startup wiring, all login middlewares |
| `internal/ent/` | Ent ORM layer — schema in `schema/token.go`, codegen output + hand-written `init.go` for DB open/auto-migrate |
| `internal/cache/` | In-memory token cache keyed by `types.KeyID`, polled from DB, supports patch/drop/find |
| `internal/plumbing/` | Async stats persistence (buffered channel → periodic DB transactions) |
| `internal/server/` | API handler implementing the ogen-generated `api.Handler` interface — CRUD tokens, scoped to user from context |
| `internal/types/` | Cryptographic types: `Key` (40 bytes: 8 KeyID + 32 secret), `AccessKey` (globs + hash validation), `Headers` |
| `internal/utils/` | Context user helpers (`WithUser`/`GetUser`), flash-cookie helpers (base64, HttpOnly, 10s TTL) |
| `api/` | OpenAPI-generated code (ogen v1.1.0) — handler interface, client, JSON codec, validators; re-generate via `go generate` |
| `web/` | `public.go`: `//go:embed admin-ui/dist` → `fs.FS`. `private.go`: `AuthHandler` — the forward-auth HTTP handler |
| `web/admin-ui/` | Svelte 4 SPA with Vite 5, Bulma CSS, `svelte-spa-router`; API client codegen via `openapi-typescript-codegen` |
| `examples/` | Docker Compose and Kubernetes deployment examples for Caddy, Nginx, Traefik |
| `docs/` | Logo SVG |
| `.github/workflows/` | PR: lint + test + migration test; Release: push tag → GoReleaser |
| `openapi.yaml` | OpenAPI 3.0.3 spec — source of truth for both Go and TypeScript API codegen |

## Development Commands

```bash
make test        # go test -v ./...        (all packages)
make lint        # golangci-lint run       (config: .golangci.yml)
make snapshot    # goreleaser --snapshot + docker tag as :1 and :snapshot
make local       # goreleaser -f .goreleaser.local.yaml --clean  (local build)
make gen         # go generate ./...       (ogen + ent codegen)
cd web/admin-ui && npm run build   # build Svelte SPA to dist/
cd web/admin-ui && npm run dev     # Vite dev server at :5050
cd web/admin-ui && npm run gen     # regenerate TS API client from openapi.yaml
cd web/admin-ui && npm run check   # svelte-check type checking
```

Code generation triggers:
- `api/generate.go`: `//go:generate … ogen … openapi.yaml` → `api/oas_*_gen.go`
- `internal/ent/generate.go`: `//go:generate … ent generate ./schema` → `internal/ent/token*.go` etc.

## Code Conventions & Common Patterns

### Generating code
- **OpenAPI spec** (`openapi.yaml`) is the single source of truth for the REST API. Two codegen pipelines consume it:
  - **Go server**: `ogen` (v1.1.0) generates handler interface, client, JSON codec, validators → `api/`. Config: `.ogen.yaml` (convenient_errors off, server + client paths only).
  - **TypeScript client**: `openapi-typescript-codegen` → `web/admin-ui/src/api/`. Run via `npm run gen`.
- **Database ORM**: `ent` (entgo.io v0.13.1) generates from `internal/ent/schema/token.go`. `ent.New()` in `init.go` auto-migrates on startup.

### Server implementation
- `internal/server.Server` implements `api.Handler` (ogen-generated interface) and wraps `*ent.Client`.
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
- Svelte SPA built into `web/admin-ui/dist/`, embedded via `//go:embed admin-ui/dist` in `web/public.go`.
- Served by `http.FileServerFS(web.Assets())` on the admin router, mounted after the API prefix.
- `vite.config.ts` uses `base: './'` for relative asset paths to work under any mount point.

### Security patterns
- Token keys: 40 bytes = 8-byte public `KeyID` (stored in DB) + 32-byte private payload (NOT stored in DB — only its SHA3-384 hash).
- Validation uses `crypto/subtle.ConstantTimeCompare` for hash comparison.
- Password auth uses bcrypt (`golang.org/x/crypto/bcrypt`).
- OWASP security headers added to admin server in production mode (`X-Frame-Options`, `X-XSS-Protection`, `X-Content-Type-Options`, `Referrer-Policy`).
- Debug mode (`--debug.enable`) enables CORS and request logging.

### Naming conventions
- Package names are short, lowercase, single-word: `cache`, `ent`, `plumbing`, `utils`, `types`.
- Test files use `_test` package suffix (e.g., `package server_test`, `package types_test`) for black-box testing.
- Generated files are named `oas_*_gen.go` (ogen) and `token*.go`/`mutation.go`/`tx.go` etc. (ent).

### Linting
`golangci.yml` enables 60 linters. Tests are excluded (`run.tests: false`). After any code change, run both:
- `go fix ./...` — applies stdlib modernization fixes (see https://go.dev/blog/gofix)
- `golangci-lint run` — enforces all enabled linters

## Important Files

| File | Role |
|------|------|
| `cmd/token-login/main.go` | Single-file entrypoint: config, CLI, startup wiring, all auth middlewares (495 lines) |
| `openapi.yaml` | OpenAPI 3.0.3 spec — 6 endpoints, 4 schemas (Config, Credential, Token, NameValue). Base: `/api/v1` |
| `internal/ent/schema/token.go` | Ent schema definition — the authoritative DB model |
| `internal/ent/init.go` | `ent.New()` — DB open + auto-migrate, supports SQLite and Postgres via URL scheme |
| `internal/cache/cache.go` | In-memory token cache; `FindByKey()` is on the auth hot path |
| `internal/types/access_key.go` | `AccessKey.Valid()` — core auth validation (glob matching + constant-time hash compare) |
| `web/private.go` | `AuthHandler` — the forward-auth HTTP endpoint |
| `web/public.go` | `//go:embed` for the Svelte SPA |
| `internal/server/server.go` | implements `api.Handler` (ogen) — CRUD tokens with per-user scoping |
| `api/generate.go` | `//go:generate` directive for ogen; exports `Prefix = "/api/v1"` |
| `.ogen.yaml` | ogen codegen config |
| `.goreleaser.yaml` | Multi-platform release + multi-arch Docker manifests |
| `Makefile` | Dev commands: lint, test, snapshot, gen |
| `migration-test.sh` | Forward-compatibility test script (DB migration from v1.1.0) |

## Runtime/Tooling Preferences

- **Go 1.26+** (see `go.mod`). Build with `CGO_ENABLED=0`, `-trimpath`.
- **Node 20**, **npm** — frontend build (`web/admin-ui`). Svelte 4, Vite 5, TypeScript 5.
- **Docker**: FROM scratch, binary ADDed to `/`. Exposes 8080, 8081. Volume `/data`.
- **Package manager**: Go modules (`go.mod`), npm for frontend.
- **Codegen**: ogen v1.1.0 (Go server/client from OpenAPI), ent v0.13.1 (ORM from schema), openapi-typescript-codegen (TS client).
- **No external build system**: plain `Makefile` + `go generate`.

## Testing & QA

- **Test framework**: `testing` stdlib + `github.com/stretchr/testify` (assert/require).
- **Run**: `make test` or `go test -v ./...`.
- **Test DB**: In-memory SQLite (`file::memory:?cache=shared`) for unit tests — see `internal/server/server_test.go`.
- **Black-box tests**: Test packages use `_test` suffix (`server_test`, `types_test`).
- **Migration test**: `migration-test.sh` (run in CI) — starts Postgres, launches v1.1.0 container, creates tokens, replaces with snapshot build, verifies data integrity with `jq` against the API. Tests both SQLite and Postgres backends.
- **CI**: PR workflow runs lint (`golangci-lint v1.58`), tests (`make test`), and migration test. Release workflow (on `v*` tag) runs GoReleaser.
