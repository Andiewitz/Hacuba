# Listings service (Go)

Property marketplace for Hacuba: sale + rental listings in one `listings`
table (`listing_mode` = `for_sale` | `for_rent`). Backed by a **private
Postgres** (`listings-db`) on an internal Docker network — same isolation
rule as `auth-db`. Users live in `auth-db`, so listings reference sellers
by UUID only (`owner_id`, no FK); identity always comes from the JWT `sub`.

Status: domain, auth boundary, Postgres store, Compose stack, and public/seller
listing routes are implemented. S3 image upload, Redis detail caching,
Terraform, and client integration remain later build steps.

## Domain model

See `migrations/001_init.sql` (source of truth). Summary:

- **Identity:** UUIDv7 generated in Go, Postgres PK. `owner_id` = JWT
  `sub`, never from body or URL; writes always scoped with
  `WHERE id = $1 AND owner_id = $2`.
- **Money:** integer centavos (`price_centavos bigint CHECK > 0`).
  Never float. `currency` defaults to `PHP`; `price_period` is
  `'monthly'` for rent, `NULL` for sale (CHECK-tied to `listing_mode`).
- **Place:** structured `city` (+ `barangay`) drives the "Where" filter;
  `address_line` stays private until the publish-address policy is
  decided; `lat`/`lng` reserved for a future map (no PostGIS yet).
- **Specs:** typed `bedrooms`, `bathrooms`, `floor_area_sqm`,
  `lot_area_sqm`. Only the unfiltered long tail (zoning, parking,
  amenities) goes in `details` jsonb.
- **Images** (`listing_images`): stores the S3 **object key**, never a
  full URL, so the CDN domain can change. Rows cascade on listing delete.

### Status machine

```
draft -> published -> closed (sold/rented) -> archived
  ^          |
  +----------+  (unpublish)
```

Publishing requires: title, description, price, city, property_type, and
at least one image. `DELETE` is a soft delete to `archived`. Drafts are
visible to the owner only; public reads see `published` (else 404, not
403 — same existence-hiding rule as auth's `RequireOwner`).

### Removed rental concepts

The old client `Listing` type (`nights`, `rating`, `isGuestFavorite`,
"for N nights") described short-term rentals, not a property marketplace.
`isNew` becomes a computation over `published_at` (e.g. last 14 days).
"Available next month" has no meaning for sales — still open whether it
becomes "Recently reduced" or is dropped.

## API (implemented against the store interface)

Public routes are `GET /healthz`, `GET /listings`, and
`GET /listings/{id}`. Browse accepts `mode`, `type`, `city`, `barangay`,
`min_price`, `max_price`, `beds`, `q`, `sort`, `cursor`, and `limit`.

Seller routes require a verified `role=seller` JWT; writes also require the
CSRF cookie/header pair: `POST /listings`, `PATCH /listings/{id}`,
`POST /listings/{id}/publish`, `/unpublish`, `/close`, `DELETE /listings/{id}`
and `GET /me/listings`. All cross-owner access is a 404. Publishing returns a
400 with all missing fields and leaves the draft unchanged.

## Search (planned for Postgres)

Plain Postgres: btree on `(status, property_type, city, price_centavos)`
(already in `001_init.sql`), `pg_trgm` for free-text "Where" in a later
migration. No Elasticsearch. Cursor pagination, not offset.

## Layout (planned, mirrors `services/auth/`)

```
services/listings/
├── cmd/server/main.go
├── internal/
│   ├── config/       env config, S3/CloudFront settings
│   ├── listings/     model, validate, Store, memory + postgres impls
│   ├── images/       presign logic, content-type/size policy
│   ├── handlers/     one file per route group
│   ├── middleware/   Authenticate (JWT only), RequireSeller, RequireCSRF
│   └── server/       NewMux shared by main + tests
├── migrations/001_init.sql
├── tests/compose.smoke.sh
├── docker-compose.yml   listings + listings-db (private network)
├── Dockerfile
├── .env.example
└── README.md
```

## Open decisions

- JWT verification (option A chosen): shared verify package + same
  `JWT_SECRET`. Stale role/subject valid up to `ACCESS_TTL` (15 min) —
  accepted at this scale, to be documented in `docs/`.
- Thumbnails (Lambda vs on-the-fly) deferred; serve originals first.
- Buyer→seller contact/inquiry flow is a separate feature, not in this plan.

## Test-plan resolutions (agreed on approval)

- Audit regressions for already-fixed findings (B1 DSN, B2 startup guard,
  B3 comment, M4 dummy hash) are written as passing guards, not
  failing-first. B3 replay-closed and M4 identical-401 behavior were
  already pinned by `TestRefreshRotatesSingleUse` and
  `TestLoginEnumerationParity`; B2 is pinned by a binary-exec test in
  `services/auth/cmd/server`, B1 by a compose-DSN check in CI.
- M1/M2/M3/M5/M6 regression tests **and** fixes ride with build step 2
  (auth role change), per the test plan's implementation order — not here.
- The §5a Go-level cross-service contract test is dropped as
  unimplementable (`internal/` is unimportable across modules). The
  auth→listings handshake is covered by unit tests inside the shared
  `authjwt` package plus the live compose smoke test.
- `listing_images` enforces `UNIQUE (listing_id, position)` so gallery
  order is deterministic and the constraint test in §4.2 has a real
  target.
- Statistical timing-parity test for login stays `-short`-skippable (step 2).
