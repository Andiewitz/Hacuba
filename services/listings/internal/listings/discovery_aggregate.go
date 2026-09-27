package listings

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type discoveryEventRow struct {
	ID        uuid.UUID
	ViewerID  uuid.UUID
	ListingID uuid.UUID
	EventType string
	CreatedAt time.Time
}

var discoveryMetricColumn = map[string]string{
	DiscoveryImpression:    "impressions",
	DiscoveryCardClick:     "card_clicks",
	DiscoveryDetailView:    "detail_views",
	DiscoveryFavorite:      "favorites",
	DiscoveryContactReveal: "contact_reveals",
}

// AggregateDiscoveryEvents atomically moves a bounded batch of raw events to
// daily listing counters and refreshes broad, anonymous viewer preferences.
// SKIP LOCKED permits more than one worker without double-counting rows.
func (s *PostgresStore) AggregateDiscoveryEvents(ctx context.Context, limit int) (int, error) {
	if limit < 1 || limit > 1000 {
		return 0, fmt.Errorf("discovery aggregation batch size must be 1 through 1000")
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `SELECT id, viewer_id, listing_id, event_type, created_at
		FROM discovery_events
		WHERE aggregated_at IS NULL
		ORDER BY created_at ASC, id ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	events := make([]discoveryEventRow, 0, limit)
	for rows.Next() {
		var event discoveryEventRow
		if err := rows.Scan(&event.ID, &event.ViewerID, &event.ListingID, &event.EventType, &event.CreatedAt); err != nil {
			return 0, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(events) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}
		return 0, nil
	}

	viewerIDs := make(map[uuid.UUID]struct{})
	eventIDs := make([]uuid.UUID, 0, len(events))
	for _, event := range events {
		column, ok := discoveryMetricColumn[event.EventType]
		if !ok {
			return 0, fmt.Errorf("unsupported discovery event type %q", event.EventType)
		}
		query := `INSERT INTO listing_metrics_daily (listing_id, day, ` + column + `)
			VALUES ($1, $2::date, 1)
			ON CONFLICT (listing_id, day) DO UPDATE SET ` + column + ` = listing_metrics_daily.` + column + ` + 1`
		if _, err := tx.Exec(ctx, query, event.ListingID, event.CreatedAt.UTC()); err != nil {
			return 0, err
		}
		viewerIDs[event.ViewerID] = struct{}{}
		eventIDs = append(eventIDs, event.ID)
	}

	if _, err := tx.Exec(ctx, `UPDATE discovery_events SET aggregated_at=now() WHERE id = ANY($1)`, eventIDs); err != nil {
		return 0, err
	}
	for viewerID := range viewerIDs {
		if err := refreshDiscoveryProfile(ctx, tx, viewerID); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(events), nil
}

func refreshDiscoveryProfile(ctx context.Context, tx pgx.Tx, viewerID uuid.UUID) error {
	_, err := tx.Exec(ctx, `WITH interactions AS (
		SELECT l.city, l.property_type, l.price_centavos, l.bedrooms
		FROM discovery_events AS event
		JOIN listings AS l ON l.id = event.listing_id
		WHERE event.viewer_id = $1
			AND event.created_at >= now() - interval '30 days'
			AND event.event_type IN ('card_click', 'detail_view', 'favorite', 'contact_reveal')
	), preferences AS (
		SELECT jsonb_strip_nulls(jsonb_build_object(
			'city', (SELECT city FROM interactions WHERE city IS NOT NULL GROUP BY city ORDER BY count(*) DESC, city ASC LIMIT 1),
			'property_type', (SELECT property_type FROM interactions WHERE property_type IS NOT NULL GROUP BY property_type ORDER BY count(*) DESC, property_type ASC LIMIT 1),
			'price_centavos', (SELECT percentile_disc(0.5) WITHIN GROUP (ORDER BY price_centavos) FROM interactions WHERE price_centavos IS NOT NULL),
			'bedrooms', (SELECT percentile_disc(0.5) WITHIN GROUP (ORDER BY bedrooms) FROM interactions WHERE bedrooms IS NOT NULL)
		)) AS value
	)
	INSERT INTO discovery_profiles (viewer_id, preferences, updated_at)
	SELECT $1, value, now() FROM preferences
	ON CONFLICT (viewer_id) DO UPDATE
	SET preferences=EXCLUDED.preferences, updated_at=EXCLUDED.updated_at`, viewerID)
	return err
}
