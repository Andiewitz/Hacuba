-- 001_init.sql — listings-db schema.
--
-- OWNERSHIP: this database is private to the listings service.
-- Only the listings service role connects (see docker-compose.yml).
-- No other service gets credentials or network routes to listings-db.
--
-- Users live in auth-db (a different database), so listings reference
-- them by UUID only: owner_id has NO foreign key. Identity always comes
-- from the JWT sub, and every write is scoped with WHERE owner_id = $1.
--
-- UUIDs are generated in Go (uuid v7) and sent as primary keys —
-- Postgres never mints identities, it only enforces uniqueness.
--
-- Money is integer centavos (price_centavos bigint CHECK > 0). Never float.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS listings (
    id UUID PRIMARY KEY,
    -- JWT sub of the seller. No FK: users live in auth-db.
    owner_id UUID NOT NULL,
    listing_mode TEXT NOT NULL CHECK (listing_mode IN ('for_sale', 'for_rent')),
    property_type TEXT NOT NULL CHECK (property_type IN ('house', 'apartment', 'condo', 'lot', 'commercial')),
    title TEXT NOT NULL CHECK (char_length(title) BETWEEN 5 AND 120),
    description TEXT NOT NULL CHECK (char_length(description) <= 5000),
    price_centavos BIGINT NOT NULL CHECK (price_centavos > 0),
    currency CHAR(3) NOT NULL DEFAULT 'PHP',
    -- 'monthly' for rent, NULL for sale — tied to listing_mode.
    price_period TEXT CHECK (
        (listing_mode = 'for_rent' AND price_period = 'monthly') OR
        (listing_mode = 'for_sale' AND price_period IS NULL)
    ),
    city TEXT NOT NULL,
    barangay TEXT,
    -- Private until the publish-address policy is decided; never in card DTOs.
    address_line TEXT,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    bedrooms SMALLINT CHECK (bedrooms IS NULL OR bedrooms >= 0),
    bathrooms SMALLINT CHECK (bathrooms IS NULL OR bathrooms >= 0),
    floor_area_sqm NUMERIC CHECK (floor_area_sqm IS NULL OR floor_area_sqm > 0),
    lot_area_sqm NUMERIC CHECK (lot_area_sqm IS NULL OR lot_area_sqm > 0),
    -- Unfiltered long tail only (zoning, parking, amenities). Never filtered on.
    details JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'closed', 'archived')),
    -- Drives "New listings". Set on publish, cleared on unpublish.
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_listings_owner_id ON listings (owner_id);
CREATE INDEX IF NOT EXISTS idx_listings_status_published_at ON listings (status, published_at DESC);
-- Public browse path: status + type + city + price ordering.
CREATE INDEX IF NOT EXISTS idx_listings_browse
    ON listings (status, property_type, city, price_centavos);

CREATE TABLE IF NOT EXISTS listing_images (
    id UUID PRIMARY KEY,
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    -- S3 object key only (e.g. listings/{listing_id}/{uuid}.webp).
    -- Never a full URL, so the CDN domain can change.
    object_key TEXT NOT NULL UNIQUE,
    position SMALLINT NOT NULL DEFAULT 0 CHECK (position >= 0),
    alt TEXT,
    content_type TEXT NOT NULL,
    byte_size BIGINT NOT NULL CHECK (byte_size > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_listing_images_listing_id ON listing_images (listing_id, position);

-- Lock down: revoke default public access. The migration runner / listings
-- role is the only writer. Adjust role name to match POSTGRES_USER.
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
