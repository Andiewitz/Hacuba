-- 002_domain_fixes.sql keeps the original migration immutable for databases
-- that have already applied it. Drafts may be incomplete; publish-time data
-- is enforced here and again in the service before a listing becomes public.

ALTER TABLE listings
    ALTER COLUMN listing_mode DROP NOT NULL,
    ALTER COLUMN property_type DROP NOT NULL,
    ALTER COLUMN title DROP NOT NULL,
    ALTER COLUMN description DROP NOT NULL,
    ALTER COLUMN price_centavos DROP NOT NULL,
    ALTER COLUMN city DROP NOT NULL;

ALTER TABLE listings
    ADD CONSTRAINT listings_currency_php CHECK (currency = 'PHP'),
    ADD CONSTRAINT listings_coordinates_pair_and_range CHECK (
        (lat IS NULL AND lng IS NULL) OR
        (lat BETWEEN -90 AND 90 AND lng BETWEEN -180 AND 180)
    ),
    ADD CONSTRAINT listings_price_period_by_mode CHECK (
        (listing_mode = 'for_rent' AND price_period IS NOT NULL AND price_period = 'monthly') OR
        (listing_mode = 'for_sale' AND price_period IS NULL) OR
        listing_mode IS NULL
    ),
    ADD CONSTRAINT listings_published_fields CHECK (
        status NOT IN ('published', 'closed') OR (
            listing_mode IS NOT NULL AND
            property_type IS NOT NULL AND
            title IS NOT NULL AND char_length(title) BETWEEN 5 AND 120 AND
            description IS NOT NULL AND char_length(description) BETWEEN 1 AND 5000 AND
            price_centavos IS NOT NULL AND price_centavos > 0 AND
            city IS NOT NULL AND city <> '' AND
            published_at IS NOT NULL
        )
    ),
    ADD CONSTRAINT listings_draft_not_published CHECK (
        status <> 'draft' OR published_at IS NULL
    );

DROP INDEX IF EXISTS idx_listings_status_published_at;
DROP INDEX IF EXISTS idx_listings_browse;
DROP INDEX IF EXISTS idx_listing_images_listing_id;

CREATE INDEX idx_listings_public_newest
    ON listings (listing_mode, property_type, city, published_at DESC, id DESC)
    WHERE status = 'published';
CREATE INDEX idx_listings_public_price
    ON listings (listing_mode, property_type, city, price_centavos, id)
    WHERE status = 'published';
CREATE INDEX idx_listings_owner_created
    ON listings (owner_id, created_at DESC, id DESC);
