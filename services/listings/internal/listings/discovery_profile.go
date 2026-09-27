package listings

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
)

type discoveryProfileSample struct {
	Listing   Listing
	EventType string
	CreatedAt time.Time
}

func discoveryProfileFromSamples(viewerID uuid.UUID, samples []discoveryProfileSample) (DiscoveryProfile, error) {
	cityCounts, typeCounts := map[string]int{}, map[string]int{}
	prices := []int64{}
	beds := []int16{}
	cutoff := time.Now().UTC().AddDate(0, 0, -30)
	for _, sample := range samples {
		if !meaningfulDiscoveryEvent(sample.EventType) || sample.CreatedAt.Before(cutoff) {
			continue
		}
		if sample.Listing.City != "" {
			cityCounts[sample.Listing.City]++
		}
		if sample.Listing.PropertyType != "" {
			typeCounts[sample.Listing.PropertyType]++
		}
		if sample.Listing.PriceCentavos != nil {
			prices = append(prices, *sample.Listing.PriceCentavos)
		}
		if sample.Listing.Bedrooms != nil {
			beds = append(beds, *sample.Listing.Bedrooms)
		}
	}
	if len(cityCounts) == 0 && len(typeCounts) == 0 && len(prices) == 0 && len(beds) == 0 {
		return DiscoveryProfile{}, ErrNotFound
	}
	profile := DiscoveryProfile{ViewerID: viewerID, City: mostCommon(cityCounts), PropertyType: mostCommon(typeCounts)}
	if len(prices) > 0 {
		sort.Slice(prices, func(i, j int) bool { return prices[i] < prices[j] })
		price := prices[len(prices)/2]
		profile.PriceCentavos = &price
	}
	if len(beds) > 0 {
		sort.Slice(beds, func(i, j int) bool { return beds[i] < beds[j] })
		bedrooms := beds[len(beds)/2]
		profile.Bedrooms = &bedrooms
	}
	return profile, nil
}

func mostCommon(values map[string]int) string {
	best, count := "", -1
	for value, candidateCount := range values {
		if candidateCount > count || (candidateCount == count && value < best) {
			best, count = value, candidateCount
		}
	}
	return best
}

func meaningfulDiscoveryEvent(eventType string) bool {
	switch eventType {
	case DiscoveryCardClick, DiscoveryDetailView, DiscoveryFavorite, DiscoveryContactReveal:
		return true
	default:
		return false
	}
}

func (m *MemoryStore) GetDiscoveryProfile(_ context.Context, viewerID uuid.UUID) (DiscoveryProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	samples := make([]discoveryProfileSample, 0)
	for _, event := range m.events {
		listing, ok := m.listings[event.ListingID]
		if event.ViewerID == viewerID && ok && listing.Status == StatusPublished {
			samples = append(samples, discoveryProfileSample{Listing: listing, EventType: event.EventType, CreatedAt: event.CreatedAt})
		}
	}
	return discoveryProfileFromSamples(viewerID, samples)
}

func (m *MemoryStore) ListViewedListingIDs(_ context.Context, viewerID uuid.UUID) ([]uuid.UUID, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cutoff := time.Now().UTC().AddDate(0, 0, -30)
	seen := map[uuid.UUID]struct{}{}
	for _, event := range m.events {
		if event.ViewerID == viewerID && meaningfulDiscoveryEvent(event.EventType) && !event.CreatedAt.Before(cutoff) {
			seen[event.ListingID] = struct{}{}
		}
	}
	ids := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	return ids, nil
}
