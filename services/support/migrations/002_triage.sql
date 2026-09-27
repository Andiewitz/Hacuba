CREATE TABLE support_report_actions (
    id UUID PRIMARY KEY,
    report_id UUID NOT NULL REFERENCES support_reports(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL,
    from_status TEXT NOT NULL CHECK (from_status IN ('open', 'triaged', 'resolved')),
    to_status TEXT NOT NULL CHECK (to_status IN ('triaged', 'resolved')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_support_report_actions_report_created ON support_report_actions (report_id, created_at DESC);
