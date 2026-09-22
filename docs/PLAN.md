# Listings implementation plan

This plan turns Hacuba from a static property catalogue into a sale and rental
marketplace. Work is released in phases so each layer has a working contract
before the next depends on it.

## Working rules

- Listings support `for_sale` and `for_rent`; money is integer PHP centavos.
- A seller owns every listing operation. Cross-owner reads and writes return
  `404`; a buyer receives `403` on seller-only routes.
- Drafts may be incomplete. Publishing returns `400` with every missing or
  invalid field and leaves the listing as a draft.
- Public endpoints expose published listings only. They never reveal owner ID,
  private address, drafts, or archived rows.
- Auth and listings share JWT verification only. Listings never gets auth-db
  credentials. A role or account change can take up to the access-token TTL to
  reach listings.

## Phase 0 — Domain and authentication foundation

**Status: implemented; keep as the regression baseline.**

- Keep the additive listings migration fixes: draft-nullable fields, strict
  sale/rent price-period rule, PHP-only currency, coordinate constraints,
  published-at rules, and public/owner indexes.
- Keep auth buyer/seller roles, `POST /auth/become-seller`, the shared
  `authjwt` module, bounded auth requests, atomic refresh rotation, and
  graceful shutdown.
- Keep the listings HTTP domain: public browse/detail, owner-scoped drafts and
  lifecycle transitions, strict publishing, CSRF-protected writes, and the
  memory/Postgres store boundary.

**Tests**

- Run `go vet ./...` and `go test ./...` in `shared/authjwt`, `services/auth`,
  and `services/listings`.
- Run listings migration tests against real Postgres. Verify each constraint
  directly, including a rent listing without `price_period`, mismatched
  coordinates, a draft, and an incomplete published row.
- Exercise the isolation matrix through the production router: seller A cannot
  read, edit, publish, close, archive, or manage seller B's listing; all are
  `404`. A buyer gets `403` on seller routes.

## Phase 1 — Deployable listings service and database

- Add the listings Dockerfile, `.env.example`, private `listings-db` Compose
  network, healthcheck, and migration mount. Publish only the application
  port, never Postgres or Redis.
- Make the Postgres store the Compose default; retain memory storage only for
  focused unit tests and local development without `DATABASE_URL`.
- Add a listings Compose smoke script and CI workflow triggered by
  `services/listings/**` and `shared/authjwt/**`.
- Add `pg_trgm` in a later migration and prove the intended public browse,
  seller dashboard, and text-search indexes with `EXPLAIN` on a representative
  data set.

**Tests**

- Start Compose, wait for `/healthz`, run migrations, then run the smoke script
  against the published application port.
- In the smoke script, create a seller token through auth, create a draft,
  verify its invisibility, publish a valid listing, filter it publicly, close
  it, and verify public `404`.
- Assert no database container has a host port and no listing environment
  variable includes the auth database DSN.
- Integration tests use a disposable Postgres database and apply migrations in
  filename order from an empty schema.

## Phase 2 — Image upload pipeline

- Add S3 configuration and an object-store interface. Implement short-lived
  presigned `PUT` URLs for keys shaped as `listings/{listing_id}/{uuid}.{ext}`.
- `POST /listings/{id}/images/presign` validates seller ownership, JPEG/PNG/
  WebP type, a 10 MB cap, and the per-listing image cap before issuing a URL.
- `POST /listings/{id}/images` performs S3 `HEAD` verification before adding
  the database row. It stores only the object key, content type, byte size,
  position, and optional alt text.
- Add owner-scoped image deletion and gallery reordering. Every image mutation
  invalidates the public-detail cache introduced in Phase 3.

**Tests**

- Unit-test key generation, content-type allow-list, byte-size and image-count
  limits, and position uniqueness.
- Use a fake object store to prove that a missing object returns `400` and does
  not create an image row.
- Run the isolation matrix for presign, registration, deletion, and reorder.
- In integration tests, run against MinIO or a dedicated S3 test bucket and
  verify the signed request binds the expected content type and length.

## Phase 3 — Public-detail Redis cache

- Add a cache interface and Redis implementation with a 50 ms timeout. Redis
  failure always falls back to Postgres.
- Cache only `GET /listings/{id}` under `listing:detail:v1:{listing_id}` for
  five minutes. Cache public DTO bytes only; do not cache owner ID, address,
  drafts, seller views, browse results, or search results.
- Cache a public miss for 30 seconds so unknown-id probing does not repeatedly
  query Postgres. A draft and a nonexistent listing remain indistinguishable.
- Use a short lock or single-flight mechanism so one request rebuilds an
  expired hot key. Invalidate the key after edit, publish, unpublish, close,
  archive, or image mutation.

**Tests**

- Unit-test hit, miss, expiry, negative-cache, invalidation, and Redis-down
  paths using a fake cache and deterministic clock.
- Integration-test Redis with parallel detail requests and assert one database
  rebuild per cold key.
- Serialize the response before caching and assert that private fields are
  absent in both cached and uncached responses.

## Phase 4 — Infrastructure

- Add Terraform for a private S3 bucket, public-access block, CloudFront Origin
  Access Control, client-origin-only bucket CORS, scoped service IAM policy,
  and orphan-upload lifecycle expiry.
- Keep Redis private in Compose and production: password required, memory cap,
  `allkeys-lru`, and no public port.
- Define environment separation for development and production, then choose the
  Terraform region and deployment domain before apply.

**Tests**

- Run `terraform fmt -check` and `terraform validate` for every environment.
- Use static policy checks to assert public access is blocked and service IAM
  access is restricted to `listings/*` object operations.
- In a non-production account, upload through a presigned URL, retrieve only
  through CloudFront, and confirm a direct public S3 URL is denied.

## Phase 5 — Client API and property presentation

- Replace `client/src/data/listings.ts` mock rental data with shared TypeScript
  listing DTOs and a server-side listings API client.
- Update cards to use stable listing IDs; display property type, mode badge,
  PHP price, `/month` for rent, and available beds, baths, and area. Remove
  nights, rating, guest-favorite, fake saving, and “Available next month.”
- Make category pages read the same filtered API, not slices of unrelated seed
  data. Empty results receive a deliberate empty state.
- Keep all new UI within `DESIGN.md`: semantic variables, Figtree headings,
  Inter UI text, 8px spacing, hierarchical radius, and forest-tinted shadows.

**Tests**

- Run `npm run typecheck`, `npm run lint`, `npm run lint:ox`, and `npm run
  build` from the client project.
- Component-test sale, rent, partial-spec, missing-image, and empty-result
  cards. Assert no retired rental copy appears in rendered output.
- Use a mocked listings API to test loading, API error, and empty states.

## Phase 6 — Search and seller experience

- Make the search controls write validated `city`, `min_price`, `max_price`,
  `type`, and `mode` query parameters with `next/navigation`; pages read those
  parameters server-side and call the public browse API.
- Add Next route handlers for listings writes. They forward the browser cookie,
  CSRF header, and in-memory access token to the listings service without
  exposing service secrets to the browser.
- Add “Become a seller,” “My listings,” create/edit draft, strict publish
  feedback, image upload, close, unpublish, and archive flows.
- Treat access tokens as memory-only client state. Route handlers use Next
  16's async `cookies()` API for cookie forwarding.

**Tests**

- Browser-level tests prove a copied search URL recreates the same filters.
- Route-handler tests assert CSRF/cookie forwarding, backend error pass-through,
  and no auth secret in response bodies.
- End-to-end test: register, become seller, create an incomplete draft, see all
  publish errors, upload/register an image, publish, find it publicly, and
  verify another seller cannot edit it.

## Phase 7 — Favorites and launch hardening

- Add `favorites(user_id, listing_id)` to listings-db with owner-safe bookmark
  APIs and a persistent client state. Favorites are a separate final feature,
  not a client-only toggle.
- Decide which scam controls launch now: seller verification, price/title
  sanity checks, house image minimums, phone/link detection, and new-seller
  publishing limits. Seller verification needs an admin workflow before it is
  a real control.
- Add monitoring, structured errors, rate limits, and an operational runbook.

**Tests**

- Verify favorites are per user and remain after refresh; cross-user favorite
  access remains hidden.
- Add tests for every selected anti-scam rule, including boundary values and
  bypass attempts.
- Run the full CI matrix: Go unit/integration/router tests, Compose smoke,
  Terraform validation, client checks, and end-to-end browser coverage.

## Release gate

Do not publish a listing service release until every phase before the intended
release scope is green, migrations are tested on an empty database and a copy
of existing data, cache invalidation tests pass, the cross-owner isolation
matrix is green, and the browser end-to-end seller flow succeeds.
