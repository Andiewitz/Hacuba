CREATE TABLE support_reports (
    id UUID PRIMARY KEY,
    reporter_id UUID,
    category TEXT NOT NULL CHECK (category IN ('listing', 'safety', 'bug', 'account')),
    listing_reference TEXT CHECK (listing_reference IS NULL OR char_length(listing_reference) <= 500),
    contact_email TEXT CHECK (contact_email IS NULL OR char_length(contact_email) <= 254),
    description TEXT NOT NULL CHECK (char_length(description) BETWEEN 20 AND 2000),
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'triaged', 'resolved')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_support_reports_status_created ON support_reports (status, created_at DESC);
CREATE INDEX idx_support_reports_category_created ON support_reports (category, created_at DESC);
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
