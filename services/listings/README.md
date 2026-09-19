# Listings service (Go)

Property marketplace for Hacuba: sale + rental listings in one `listings`
table (`listing_mode` = `for_sale` | `for_rent`). Backed by a **private
Postgres** (`listings-db`) on an internal Docker network — same isolation
rule as `auth-db`. Users live in `auth-db`, so listings reference sellers
by UUID only (`owner_id`, no FK); identity always comes from the JWT `sub`.

Status: domain + migration only (build-order step 1). Service skeleton,
auth `role` migration, and shared JWT package land in steps 2–4.

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

## Search (planned)

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
