# Auth service (Go)

Standalone authentication service for Hacuba. Owns user identities, password
hashes, and sessions. Backed by a **private Postgres** (`auth-db`) that lives
on an internal Docker network — no other service gets credentials or network
routes to it.

## Layout

```text
services/auth/
├── cmd/server/main.go            # route wiring + server
├── internal/
│   ├── config/config.go          # env config (JWT secret, TTLs, DB URL)
│   ├── users/                    # User model (UUID PK), validation,
│   │                             # Store interface, Postgres + memory impls
│   ├── tokens/tokens.go          # access JWT + opaque refresh/CSRF tokens
│   ├── handlers/
│   │   ├── registration.go       # POST /auth/register (UUIDv7 + bcrypt)
│   │   ├── login.go              # POST /auth/login (rate-limited)
│   │   ├── refresh.go            # POST /auth/refresh (rotation)
│   │   ├── logout.go             # POST /auth/logout
│   │   ├── me.go                 # GET /auth/me (own row only)
│   │   └── session.go, cookies.go
│   └── middleware/               # Authenticate, RequireCSRF, RequireOwner
├── migrations/001_init.sql
├── docker-compose.yml            # auth + isolated auth-db
└── Dockerfile
```

## Auth model

- **Identity:** UUIDv7 per user, generated in Go. Used as Postgres PK,
  JWT `sub`, and `owner_id` FK everywhere else.
- **Passwords:** bcrypt cost 12, never returned by the API.
- **Access JWT (15m):** `sub=user UUID`, `role`, and `csrf` hash binding,
  signed with `HS256`. Listings verifies the same claim with the shared
  `authjwt` module; deleted or demoted accounts can retain access there for
  at most the access-token lifetime.
  Frontend keeps it in memory (not localStorage), sends as
  `Authorization: Bearer <jwt>`.
- **Refresh (7d):** 32-byte opaque token in `HttpOnly; Secure; SameSite=Lax`
  cookie. Only its SHA-256 hash is stored. Rotated on every use —
  replay fails closed.
- **CSRF:** 32-byte token as readable cookie + `X-CSRF-Token` header
  (double-submit). Also hash-bound to the JWT/session so tokens can't be
  replayed across sessions. Required on all mutating routes.

## Per-user isolation

1. `Authenticate` verifies the JWT and puts the caller UUID in context.
2. Handlers read identity **only** from context — `/auth/me` takes no ID
   parameter, so cross-user reads are impossible by construction.
3. Future resource routes wrap with `RequireOwner` (mismatch → `404`, not
   `403`, to hide existence) **and** scope SQL with
   `WHERE owner_id = $1` using the JWT sub.

## Run

```bash
cp .env.example .env   # set DB_PASSWORD + JWT_SECRET (openssl rand -hex 32)
docker compose up --build
# API on :8080, DB reachable ONLY at auth-db:5432 from the auth container
```

Without `DATABASE_URL` the server boots with an in-memory store (dev/tests
only). Production Compose always sets it to the internal `auth-db`.

## Endpoints

| Method | Path | Auth | Notes |
| ------ | ---- | ---- | ----- |
| POST | /auth/register | no | 201 + session, 409 on duplicate |
| POST | /auth/login | no | 401 generic, 429 after 5 fails/15m |
| POST | /auth/refresh | refresh cookie + CSRF | rotates session |
| POST | /auth/logout | refresh cookie | 204, clears cookies |
| POST | /auth/become-seller | Bearer JWT + CSRF | idempotently upgrades role and returns a fresh session |
| GET | /auth/me | Bearer JWT | own `{id,email,role,created_at}` only |
| GET | /healthz | no | liveness |

## Verify

```bash
# Fast unit suite (memory store, no Docker needed)
go vet ./...
go test ./...
go build ./...

# With real Postgres (applies migrations, exercises constraints,
# sessions, cascade deletes — skips cleanly without a DSN)
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
cross-user CSRF kill, logout), `internal/middleware/*_test.go` (401/403/404
matrix). CI (`.github/workflows/auth-tests.yml`) runs unit + postgres +
compose smoke on every push touching `services/auth/**`.
