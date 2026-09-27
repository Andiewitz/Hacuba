# Discovery service plan

Discovery is implemented inside the Listings service until measured catalogue
size or traffic justifies a separate read system. Listings remains the only
source of truth for publication state, price, location, and gallery data.

## Phase 1 — Search foundation

**Status: implemented.**

- `GET /listings` uses opaque, sort-bound keyset cursors for newest and price
  ordering. A page can no longer repeat when listings are added between
  requests.
- PostgreSQL indexes support seller dashboard ordering, published newest/price
  ordering, and case-insensitive partial search of title, city, and barangay.
- Only `published` listings are indexed as public discovery candidates.

## Phase 2 — Interaction capture and aggregation

Add a rate-limited, batched `POST /discovery/events` endpoint. Record only a
rotating anonymous viewer ID or an authenticated account ID, listing ID, event
type, search context, and timestamp. Event types are impression, card click,
detail view, favorite, and contact reveal. Never store raw IP addresses as a
recommendation feature.

`discovery_events`, `listing_metrics_daily`, and `discovery_profiles` are
implemented. `cmd/discovery-worker` claims events with `FOR UPDATE SKIP
LOCKED`, updates daily listing counters, marks raw rows aggregated, and
derives a 30-day anonymous preference profile (city, property type, median
price, and median bedrooms). It needs a PostgreSQL `DATABASE_URL`; the
SQLite development store deliberately does not pretend to support the
production aggregation worker.

## Phase 3 — Candidate generation and ranking

Add `GET /discover` and `GET /listings/{id}/related`. Candidate generation
starts with published listings only, then filters by location, property type,
price, and bedroom proximity. Ranking combines textual relevance, filter
match, freshness, engagement, and observed preferences while removing already
viewed listings and near-duplicate cards.

Every response includes a short reason suitable for the UI, such as “Similar
homes in Cebu City.” Ranking snapshots and feature flags make score changes
auditable and reversible.

## Phase 4 — Product surfaces

Use discovery sections for newly listed properties, matching city/type views,
budget matches, related listings, and a signed-in “For you” feed. Seller
analytics show impressions, card clicks, detail views, and contact reveals;
they do not imply a payment, contract, or completed transaction.

## Phase 5 — Scale and quality gates

Track zero-result rate, result click-through rate, contact-reveal rate,
duplicate-card rate, page latency, and ranking coverage. Test cursor stability,
inactive-listing exclusion, cross-user privacy, event batching, deterministic
ranking, and new-user fallback. Move search reads to OpenSearch only when
PostgreSQL query plans and measured traffic demonstrate a need.
