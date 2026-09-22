package listings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SQLiteStore is a persistent local-development store. It deliberately uses
// a separate schema from Postgres so production behavior stays owned by the
// Postgres migrations and does not accidentally inherit SQLite semantics.
type SQLiteStore struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSQLiteStore(ctx context.Context, path string) (*SQLiteStore, error) {
	if path == "" {
		return nil, fmt.Errorf("SQLite path is required")
	}
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open development SQLite: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping development SQLite: %w", err)
	}
	store := &SQLiteStore{db: db}
	if err := store.initialize(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) initialize(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS dev_listings (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			payload TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS dev_listing_images (
			id TEXT PRIMARY KEY,
			listing_id TEXT NOT NULL,
			position INTEGER NOT NULL,
			payload TEXT NOT NULL,
			UNIQUE(listing_id, position)
		);
	`)
	if err != nil {
		return fmt.Errorf("initialize development SQLite: %w", err)
	}
	return nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) Create(ctx context.Context, l Listing) (*Listing, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	l.CreatedAt, l.UpdatedAt = now, now
	if err := s.writeListing(ctx, l); err != nil {
		return nil, err
	}
	return copyListing(l), nil
}

func (s *SQLiteStore) GetPublished(ctx context.Context, id uuid.UUID) (*Listing, error) {
	l, err := s.readListing(ctx, id)
	if err != nil || l.Status != StatusPublished {
		if err == nil {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return l, nil
}

func (s *SQLiteStore) GetOwner(ctx context.Context, id, owner uuid.UUID) (*Listing, error) {
	l, err := s.readListing(ctx, id)
	if err != nil {
		return nil, err
	}
	if l.OwnerID != owner {
		return nil, ErrNotFound
	}
	return l, nil
}

func (s *SQLiteStore) SaveOwner(ctx context.Context, l Listing, owner uuid.UUID) (*Listing, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, err := s.readListing(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	if stored.OwnerID != owner {
		return nil, ErrNotFound
	}
	l.OwnerID, l.CreatedAt, l.UpdatedAt = stored.OwnerID, stored.CreatedAt, time.Now().UTC()
	if err := s.writeListing(ctx, l); err != nil {
		return nil, err
	}
	return copyListing(l), nil
}

func (s *SQLiteStore) ListPublished(ctx context.Context, filter ListFilter) ([]Listing, error) {
	all, err := s.allListings(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Listing, 0)
	for _, listing := range all {
		if listing.Status == StatusPublished && matches(listing, filter) {
			out = append(out, listing)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if filter.Sort == "price_asc" {
			return *out[i].PriceCentavos < *out[j].PriceCentavos
		}
		if filter.Sort == "price_desc" {
			return *out[i].PriceCentavos > *out[j].PriceCentavos
		}
		return out[i].PublishedAt.After(*out[j].PublishedAt)
	})
	if filter.Cursor != "" {
		for i, listing := range out {
			if listing.ID.String() == filter.Cursor {
				out = out[i+1:]
				break
			}
		}
	}
	if filter.Limit > 0 && len(out) > filter.Limit {
		out = out[:filter.Limit]
	}
	return out, nil
}

func (s *SQLiteStore) ListOwner(ctx context.Context, owner uuid.UUID) ([]Listing, error) {
	all, err := s.allListings(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Listing, 0)
	for _, listing := range all {
		if listing.OwnerID == owner {
			out = append(out, listing)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *SQLiteStore) ListImages(ctx context.Context, listingID uuid.UUID) ([]Image, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT payload FROM dev_listing_images WHERE listing_id=? ORDER BY position`, listingID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Image
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var image Image
		if err := json.Unmarshal([]byte(raw), &image); err != nil {
			return nil, fmt.Errorf("decode development SQLite image: %w", err)
		}
		out = append(out, image)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) AddImage(ctx context.Context, image Image, owner uuid.UUID) (*Image, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	listing, err := s.readListing(ctx, image.ListingID)
	if err != nil {
		return nil, err
	}
	if listing.OwnerID != owner {
		return nil, ErrNotFound
	}
	image.CreatedAt = time.Now().UTC()
	raw, err := json.Marshal(image)
	if err != nil {
		return nil, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO dev_listing_images (id, listing_id, position, payload) VALUES (?, ?, ?, ?)`, image.ID.String(), image.ListingID.String(), image.Position, string(raw))
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (s *SQLiteStore) DeleteImage(ctx context.Context, listingID, imageID, owner uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	listing, err := s.readListing(ctx, listingID)
	if err != nil {
		return err
	}
	if listing.OwnerID != owner {
		return ErrNotFound
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM dev_listing_images WHERE id=? AND listing_id=?`, imageID.String(), listingID.String())
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLiteStore) writeListing(ctx context.Context, listing Listing) error {
	raw, err := json.Marshal(listing)
	if err != nil {
		return fmt.Errorf("encode development SQLite listing: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO dev_listings (id, owner_id, payload) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET owner_id=excluded.owner_id, payload=excluded.payload`, listing.ID.String(), listing.OwnerID.String(), string(raw))
	return err
}

func (s *SQLiteStore) readListing(ctx context.Context, id uuid.UUID) (*Listing, error) {
	var ownerID, raw string
	err := s.db.QueryRowContext(ctx, `SELECT owner_id, payload FROM dev_listings WHERE id=?`, id.String()).Scan(&ownerID, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var listing Listing
	if err := json.Unmarshal([]byte(raw), &listing); err != nil {
		return nil, fmt.Errorf("decode development SQLite listing: %w", err)
	}
	listing.OwnerID, err = uuid.Parse(ownerID)
	if err != nil {
		return nil, fmt.Errorf("decode development SQLite owner ID: %w", err)
	}
	return &listing, nil
}

func (s *SQLiteStore) allListings(ctx context.Context) ([]Listing, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT owner_id, payload FROM dev_listings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Listing
	for rows.Next() {
		var ownerID, raw string
		if err := rows.Scan(&ownerID, &raw); err != nil {
			return nil, err
		}
		var listing Listing
		if err := json.Unmarshal([]byte(raw), &listing); err != nil {
			return nil, fmt.Errorf("decode development SQLite listing: %w", err)
		}
		listing.OwnerID, err = uuid.Parse(ownerID)
		if err != nil {
			return nil, fmt.Errorf("decode development SQLite owner ID: %w", err)
		}
		out = append(out, listing)
	}
	return out, rows.Err()
}
