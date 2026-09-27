CREATE TABLE discovery_events (
    id UUID PRIMARY KEY,
    viewer_id UUID NOT NULL,
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('impression', 'card_click', 'detail_view', 'favorite', 'contact_reveal')),
    query TEXT CHECK (query IS NULL OR char_length(query) <= 120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_discovery_events_viewer_created ON discovery_events (viewer_id, created_at DESC);
CREATE INDEX idx_discovery_events_listing_type_created ON discovery_events (listing_id, event_type, created_at DESC);

CREATE TABLE listing_metrics_daily (
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    day DATE NOT NULL,
    impressions INTEGER NOT NULL DEFAULT 0 CHECK (impressions >= 0),
    card_clicks INTEGER NOT NULL DEFAULT 0 CHECK (card_clicks >= 0),
    detail_views INTEGER NOT NULL DEFAULT 0 CHECK (detail_views >= 0),
    favorites INTEGER NOT NULL DEFAULT 0 CHECK (favorites >= 0),
    contact_reveals INTEGER NOT NULL DEFAULT 0 CHECK (contact_reveals >= 0),
    PRIMARY KEY (listing_id, day)
);

CREATE TABLE discovery_profiles (
    viewer_id UUID PRIMARY KEY,
    preferences JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
