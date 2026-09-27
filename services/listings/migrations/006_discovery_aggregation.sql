-- Raw events are immutable. The aggregation worker marks a row only after
-- its daily metric and viewer profile updates commit in the same transaction.
ALTER TABLE discovery_events
    ADD COLUMN aggregated_at TIMESTAMPTZ;

CREATE INDEX idx_discovery_events_unaggregated
    ON discovery_events (created_at ASC, id ASC)
    WHERE aggregated_at IS NULL;
