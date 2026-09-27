ALTER TABLE listings DROP CONSTRAINT IF EXISTS listings_status_check;
ALTER TABLE listings ADD CONSTRAINT listings_status_check CHECK (status IN ('draft', 'published', 'closed', 'archived', 'moderation_hidden'));
ALTER TABLE listings ADD COLUMN IF NOT EXISTS moderation_reason TEXT;
ALTER TABLE listings ADD COLUMN IF NOT EXISTS moderation_report_id UUID;
ALTER TABLE listings ADD COLUMN IF NOT EXISTS moderated_at TIMESTAMPTZ;
CREATE TABLE IF NOT EXISTS listing_moderation_actions (
    id UUID PRIMARY KEY,
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    report_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('hide', 'restore')),
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_listing_moderation_actions_listing_created ON listing_moderation_actions (listing_id, created_at DESC);
