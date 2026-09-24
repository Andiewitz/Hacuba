# Authentication

How sign-in works in Hacuba: a standalone Go auth service owns
identities and sessions, and the Next.js client surfaces it through a
stepped login dialog backed by Redux.

## Architecture

```text
Browser (Next.js client)
  -> AuthDialog (client/src/components/auth-dialog.tsx)
  -> Redux auth slice (client/src/lib/slices/auth.ts)
  -> Next.js route handler (proxy, preserves cookies)
  -> Go auth service (services/auth, :8080)
  -> Private Postgres auth-db (internal Docker network only)
```

Key ownership rule: `auth-db` is reachable only from the auth container
(`auth-db:5432` on an internal network). No other service gets credentials
or routes to it. Without `DATABASE_URL` the server boots with an in-memory
store — dev/tests only. Production Compose always sets it.

## Identity and password rules

- **ID:** UUIDv7 generated in Go (`internal/handlers/registration.go`),
  used as Postgres PK, JWT `sub`, and future `owner_id` FKs.
- **Email:** normalized to lowercase (`users.NormalizeEmail`), max 254
  chars, must parse and contain a dotted domain. Stored with a
  `email = lower(email)` CHECK plus a unique index.
- **Password:** 8–72 chars (bcrypt limit), at least one letter and one
  number, no whitespace (`users.ValidatePassword`). Hashed with bcrypt
  cost 12, never returned by the API.

## Sessions and tokens

| Token | Lifetime | Storage | Transport |
| ----- | -------- | ------- | --------- |
| Access JWT (HS256, `sub` + `csrf` binding) | 15m (`ACCESS_TTL`) | In-memory only (never `localStorage`) | `Authorization: Bearer <jwt>` |
| Refresh (32-byte opaque) | 7d (`REFRESH_TTL=168h`) | SHA-256 hash in `refresh_sessions`, raw value only in `HttpOnly; Secure; SameSite=Lax` cookie | Cookie, auto-sent |
| CSRF (32-byte opaque) | Session-bound | SHA-256 hash in `refresh_sessions` | Readable cookie + `X-CSRF-Token` header (double-submit), hash-bound to JWT/session |

Refresh rotates on every use — replay fails closed. All mutating routes
require the CSRF pair; cross-session token replay is killed by the
JWT/session hash binding.

## Endpoints

| Method | Path | Auth | Notes |
| ------ | ---- | ---- | ----- |
| POST | `/auth/register` | no | Validate -> bcrypt -> UUIDv7 -> INSERT -> auto-login session. `201` + session, `409` on duplicate. |
| POST | `/auth/login` | no | Generic `401` on failure (no account enumeration). Rate-limited: max 5 failures/email/15m, then `429`. Success resets the counter. |
| POST | `/auth/refresh` | refresh cookie + CSRF | Rotates session, single-use. |
| POST | `/auth/logout` | refresh cookie | `204`, clears cookies, deletes session. |
| GET | `/auth/me` | Bearer JWT | Returns own `{id,email,created_at}` only — takes no ID parameter, so cross-user reads are impossible by construction. |
| GET | `/healthz` | no | Liveness. |

## Per-user isolation

1. `Authenticate` middleware verifies the JWT and puts the caller UUID in
   request context.
2. Handlers read identity **only** from context.
3. Resource routes wrap with `RequireOwner` (mismatch -> `404`, not `403`,
   to hide existence) **and** scope SQL with `WHERE owner_id = $1` using
   the JWT `sub`.

## Frontend flow

- Entry: profile menu in `navbar.tsx` -> "Log in / Sign up" opens
  `AuthDialog`.
- Step 1 (email): client-side regex check, "Enter a valid email address
  to continue." on failure.
- Step 2 (password): show/hide toggle, `Edit` returns to step 1, and the
  account mode switches between login and registration. Enter submits.
- The Redux slice calls same-origin Next route handlers for register, login,
  refresh, logout, and becoming a seller. The handlers proxy only the auth
  request headers and preserve each `Set-Cookie` header for the browser.
- The access JWT stays in Redux memory. On a browser reload, the readable CSRF
  cookie permits a refresh request that rotates the HttpOnly refresh session.
- States: Redux tracks idle/loading/authenticated/failed. Escape or backdrop
  click closes and clears errors. The signed-in menu shows the API email and
  buyer or seller role; logging out clears the server session and browser state.
- Google button is present but disabled ("Coming soon").

## Configuration

See `services/auth/.env.example`:

```bash
cp .env.example .env   # set DB_PASSWORD + JWT_SECRET (openssl rand -hex 32)
docker compose up --build
# API on :8080, DB reachable ONLY at auth-db:5432 from the auth container
```

Production requires `DATABASE_URL` (private DSN) and `JWT_SECRET` (min 32
bytes / 64 hex chars). For persistent local development without Docker, set
`DEV_SQLITE_PATH` instead of `DATABASE_URL`, use a local `JWT_SECRET`, and set
`COOKIE_SECURE=false` for HTTP localhost. The SQLite schema is separate from
the production migrations and is never selected when `DATABASE_URL` is set.
Optional settings include `PORT` (default 8080), `ACCESS_TTL=15m`,
`REFRESH_TTL=168h`, and `COOKIE_DOMAIN`.

## Verification

```bash
# Fast unit suite (memory store, no Docker needed)
go vet ./...
go test ./...
go build ./...

# With real Postgres (migrations, constraints, sessions, cascade deletes)
TEST_DATABASE_URL=postgres://auth_service:pw@localhost:5432/auth?sslmode=disable \
  go test -race -count=1 ./...

# Live HTTP contract against the isolated compose stack
docker compose up -d --build
./tests/compose.smoke.sh http://localhost:8080
docker compose down -v
```

Test layers: `internal/users/postgres_test.go` (real DB + migration
contract), `internal/server/router_test.go` (status codes, cookie flags,
rate limit, isolation through the production mux),
`internal/handlers/refresh_flow_test.go` (rotation single-use, expiry,
cross-user CSRF kill, logout), `internal/middleware/*_test.go`
(401/403/404 matrix). CI (`.github/workflows/auth-tests.yml`) runs unit +
postgres + compose smoke on every push touching `services/auth/**`.

## Related files

- Service: `services/auth/cmd/server/main.go`, `internal/{config,users,
  tokens,handlers,middleware}/`, `migrations/001_init.sql`,
  `docker-compose.yml`, `Dockerfile`, `tests/compose.smoke.sh`
- Client: `client/src/components/auth-dialog.tsx`,
  `client/src/components/navbar.tsx`, `client/src/lib/slices/auth.ts`
