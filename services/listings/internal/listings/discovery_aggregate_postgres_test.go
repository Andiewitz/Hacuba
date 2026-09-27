package listings

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestPostgresDiscoveryAggregation is intentionally integration-only. It
// exercises the PostgreSQL locking, upsert, and JSONB profile contract rather
// than merely checking that the worker compiles.
func TestPostgresDiscoveryAggregation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL unset — skipping PostgreSQL discovery integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for _, migration := range []string{"001_init.sql", "002_domain_fixes.sql", "003_public_contact_details.sql", "004_discovery_search.sql", "005_discovery_events.sql", "006_discovery_aggregation.sql"} {
		body, err := os.ReadFile(filepath.Join("..", "..", "migrations", migration))
		if err != nil {
			t.Fatalf("read %s: %v", migration, err)
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			t.Fatalf("apply %s: %v", migration, err)
		}
	}
	if _, err := pool.Exec(ctx, "TRUNCATE discovery_events, listing_metrics_daily, discovery_profiles, listing_images, listings CASCADE"); err != nil {
		t.Fatal(err)
	}

	store, err := NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ownerID, viewerID, listingID := uuid.New(), uuid.New(), uuid.New()
	price, bedrooms, publishedAt := int64(750_000_000), int16(3), time.Now().UTC()
	_, err = store.Create(ctx, Listing{ID: listingID, OwnerID: ownerID, ListingMode: ModeSale, PropertyType: TypeHouse, Title: "Cebu discovery integration home", Description: "Published home used to verify PostgreSQL discovery aggregation.", PriceCentavos: &price, Currency: "PHP", City: "Cebu City", Bedrooms: &bedrooms, Status: StatusPublished, PublishedAt: &publishedAt})
	if err != nil {
		t.Fatal(err)
	}
	eventTime := time.Now().UTC()
	accepted, err := store.RecordDiscoveryEvents(ctx, []DiscoveryEvent{
		{ID: uuid.New(), ViewerID: viewerID, ListingID: listingID, EventType: DiscoveryImpression, CreatedAt: eventTime},
		{ID: uuid.New(), ViewerID: viewerID, ListingID: listingID, EventType: DiscoveryDetailView, CreatedAt: eventTime},
		{ID: uuid.New(), ViewerID: viewerID, ListingID: listingID, EventType: DiscoveryContactReveal, CreatedAt: eventTime},
		{ID: uuid.New(), ViewerID: viewerID, ListingID: listingID, EventType: DiscoveryCardClick, CreatedAt: eventTime.AddDate(0, 0, -1)},
	})
	if err != nil || accepted != 4 {
		t.Fatalf("record events = %d, %v", accepted, err)
	}
	processed, err := store.AggregateDiscoveryEvents(ctx, 10)
	if err != nil || processed != 4 {
		t.Fatalf("aggregate events = %d, %v", processed, err)
	}
	processed, err = store.AggregateDiscoveryEvents(ctx, 10)
	if err != nil || processed != 0 {
		t.Fatalf("repeat aggregation = %d, %v", processed, err)
	}

	var impressions, detailViews, contactReveals int
	if err := pool.QueryRow(ctx, `SELECT impressions, detail_views, contact_reveals FROM listing_metrics_daily WHERE listing_id=$1 AND day=$2::date`, listingID, eventTime).Scan(&impressions, &detailViews, &contactReveals); err != nil {
		t.Fatal(err)
	}
	if impressions != 1 || detailViews != 1 || contactReveals != 1 {
		t.Fatalf("daily metrics = impressions:%d detail_views:%d contact_reveals:%d", impressions, detailViews, contactReveals)
	}
	var priorCardClicks int
	if err := pool.QueryRow(ctx, `SELECT card_clicks FROM listing_metrics_daily WHERE listing_id=$1 AND day=$2::date`, listingID, eventTime.AddDate(0, 0, -1)).Scan(&priorCardClicks); err != nil {
		t.Fatal(err)
	}
	if priorCardClicks != 1 {
		t.Fatalf("prior-day card_clicks = %d", priorCardClicks)
	}
	var city, propertyType string
	if err := pool.QueryRow(ctx, `SELECT preferences->>'city', preferences->>'property_type' FROM discovery_profiles WHERE viewer_id=$1`, viewerID).Scan(&city, &propertyType); err != nil {
		t.Fatal(err)
	}
	if city != "Cebu City" || propertyType != TypeHouse {
		t.Fatalf("profile = city:%q property_type:%q", city, propertyType)
	}
	profile, err := store.GetDiscoveryProfile(ctx, viewerID)
	if err != nil || profile.City != "Cebu City" || profile.PropertyType != TypeHouse || profile.PriceCentavos == nil || *profile.PriceCentavos != price || profile.Bedrooms == nil || *profile.Bedrooms != bedrooms {
		t.Fatalf("load profile = %#v, %v", profile, err)
	}
	viewed, err := store.ListViewedListingIDs(ctx, viewerID)
	if err != nil || len(viewed) != 1 || viewed[0] != listingID {
		t.Fatalf("viewed listing IDs = %#v, %v", viewed, err)
	}
}
