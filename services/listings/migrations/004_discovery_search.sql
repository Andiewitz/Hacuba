-- Discovery phase 1: keyset browse, seller dashboard, and PostgreSQL search.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Replaces the owner-only index and avoids sorting each seller dashboard.
DROP INDEX IF EXISTS idx_listings_owner_id;
CREATE INDEX IF NOT EXISTS idx_listings_owner_created
    ON listings (owner_id, created_at DESC, id DESC);

-- These partial indexes contain only public candidates and match the API's
-- deterministic keyset orderings.
CREATE INDEX IF NOT EXISTS idx_listings_public_newest
    ON listings (published_at DESC, id DESC)
    WHERE status = 'published';
CREATE INDEX IF NOT EXISTS idx_listings_public_price_asc
    ON listings (price_centavos ASC, id ASC)
    WHERE status = 'published';
CREATE INDEX IF NOT EXISTS idx_listings_public_price_desc
    ON listings (price_centavos DESC, id DESC)
    WHERE status = 'published';

-- Matches ListPublished's case-insensitive partial search expression.
CREATE INDEX IF NOT EXISTS idx_listings_public_search_trgm
    ON listings USING GIN (
        lower(title || ' ' || city || ' ' || coalesce(barangay, '')) gin_trgm_ops
    )
    WHERE status = 'published';
