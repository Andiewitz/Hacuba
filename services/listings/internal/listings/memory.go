package listings

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryStore is deterministic test storage. Production is expected to use
// Postgres with the migration constraints in migrations/.
type MemoryStore struct {
	mu       sync.RWMutex
	listings map[uuid.UUID]Listing
	images   map[uuid.UUID]map[uuid.UUID]Image
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{listings: map[uuid.UUID]Listing{}, images: map[uuid.UUID]map[uuid.UUID]Image{}}
}

func (m *MemoryStore) Create(_ context.Context, l Listing) (*Listing, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	l.CreatedAt = time.Now().UTC()
	l.UpdatedAt = l.CreatedAt
	m.listings[l.ID] = l
	copy := l
	return &copy, nil
}

func (m *MemoryStore) GetPublished(_ context.Context, id uuid.UUID) (*Listing, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.listings[id]
	if !ok || l.Status != StatusPublished {
		return nil, ErrNotFound
	}
	return copyListing(l), nil
}

func (m *MemoryStore) GetOwner(_ context.Context, id, owner uuid.UUID) (*Listing, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.listings[id]
	if !ok || l.OwnerID != owner {
		return nil, ErrNotFound
	}
	return copyListing(l), nil
}

func (m *MemoryStore) SaveOwner(_ context.Context, l Listing, owner uuid.UUID) (*Listing, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored, ok := m.listings[l.ID]
	if !ok || stored.OwnerID != owner {
		return nil, ErrNotFound
	}
	l.OwnerID = stored.OwnerID
	l.CreatedAt = stored.CreatedAt
	l.UpdatedAt = time.Now().UTC()
	m.listings[l.ID] = l
	return copyListing(l), nil
}

func (m *MemoryStore) ListPublished(_ context.Context, f ListFilter) ([]Listing, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Listing, 0)
	for _, l := range m.listings {
		if l.Status != StatusPublished || !matches(l, f) {
			continue
		}
		out = append(out, *copyListing(l))
	}
	sort.Slice(out, func(i, j int) bool {
		if f.Sort == "price_asc" {
			return *out[i].PriceCentavos < *out[j].PriceCentavos
		}
		if f.Sort == "price_desc" {
			return *out[i].PriceCentavos > *out[j].PriceCentavos
		}
		return out[i].PublishedAt.After(*out[j].PublishedAt)
	})
	if f.Cursor != "" {
		for i, l := range out {
			if l.ID.String() == f.Cursor {
				out = out[i+1:]
				break
			}
		}
	}
	if f.Limit > 0 && len(out) > f.Limit {
		out = out[:f.Limit]
	}
	return out, nil
}

func (m *MemoryStore) ListOwner(_ context.Context, owner uuid.UUID) ([]Listing, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Listing, 0)
	for _, l := range m.listings {
		if l.OwnerID == owner {
			out = append(out, *copyListing(l))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (m *MemoryStore) ListImages(_ context.Context, id uuid.UUID) ([]Image, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	images := m.images[id]
	out := make([]Image, 0, len(images))
	for _, image := range images {
		out = append(out, image)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

func (m *MemoryStore) AddImage(_ context.Context, image Image, owner uuid.UUID) (*Image, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.listings[image.ListingID]
	if !ok || l.OwnerID != owner {
		return nil, ErrNotFound
	}
	if m.images[image.ListingID] == nil {
		m.images[image.ListingID] = map[uuid.UUID]Image{}
	}
	for _, other := range m.images[image.ListingID] {
		if other.Position == image.Position {
			return nil, errors.New("image position already exists")
		}
	}
	image.CreatedAt = time.Now().UTC()
	m.images[image.ListingID][image.ID] = image
	return &image, nil
}

func (m *MemoryStore) DeleteImage(_ context.Context, listingID, imageID, owner uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.listings[listingID]
	if !ok || l.OwnerID != owner {
		return ErrNotFound
	}
	if _, ok := m.images[listingID][imageID]; !ok {
		return ErrNotFound
	}
	delete(m.images[listingID], imageID)
	return nil
}

func copyListing(l Listing) *Listing { copied := l; return &copied }

func matches(l Listing, f ListFilter) bool {
	if f.Mode != "" && l.ListingMode != f.Mode {
		return false
	}
	if f.Type != "" && l.PropertyType != f.Type {
		return false
	}
	if f.City != "" && !strings.EqualFold(l.City, f.City) {
		return false
	}
	if f.Barangay != "" && (l.Barangay == nil || !strings.EqualFold(*l.Barangay, f.Barangay)) {
		return false
	}
	if f.MinPrice != nil && (l.PriceCentavos == nil || *l.PriceCentavos < *f.MinPrice) {
		return false
	}
	if f.MaxPrice != nil && (l.PriceCentavos == nil || *l.PriceCentavos > *f.MaxPrice) {
		return false
	}
	if f.Beds != nil && (l.Bedrooms == nil || *l.Bedrooms < *f.Beds) {
		return false
	}
	q := strings.ToLower(strings.TrimSpace(f.Query))
	return q == "" || strings.Contains(strings.ToLower(l.Title), q) || strings.Contains(strings.ToLower(l.City), q)
}
