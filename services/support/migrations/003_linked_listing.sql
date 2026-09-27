ALTER TABLE support_reports ADD COLUMN IF NOT EXISTS listing_id UUID;
CREATE INDEX IF NOT EXISTS idx_support_reports_listing_created ON support_reports (listing_id, created_at DESC) WHERE listing_id IS NOT NULL;
