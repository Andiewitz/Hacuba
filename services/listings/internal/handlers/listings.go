package handlers

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/listings/internal/images"
	"github.com/hacuba/listings/internal/listings"
	"github.com/hacuba/listings/internal/middleware"
)

type Handler struct {
	Store                listings.Store
	Objects              images.ObjectStore
	DiscoveryProxySecret []byte
}

type discoverySection struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Reason   string `json:"reason"`
	Listings []any  `json:"listings"`
}

const (
	discoveryProxyHeader  = "X-Hacuba-Discovery-Proxy"
	discoveryViewerHeader = "X-Hacuba-Viewer-ID"
)

type discoveryRateWindow struct {
	started  time.Time
	requests int
}

var discoveryRateLimiter = struct {
	sync.Mutex
	windows map[string]discoveryRateWindow
}{windows: map[string]discoveryRateWindow{}}

func allowDiscoveryRequest(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	now := time.Now()
	discoveryRateLimiter.Lock()
	defer discoveryRateLimiter.Unlock()
	window := discoveryRateLimiter.windows[host]
	if window.started.IsZero() || now.Sub(window.started) >= time.Minute {
		window = discoveryRateWindow{started: now}
	}
	if window.requests >= 120 {
		discoveryRateLimiter.windows[host] = window
		return false
	}
	window.requests++
	discoveryRateLimiter.windows[host] = window
	return true
}

// RecordDiscoveryEvents accepts a bounded client batch. Viewer IDs are
// rotating UUIDs; the service intentionally does not persist IP addresses.
func (h Handler) RecordDiscoveryEvents(w http.ResponseWriter, r *http.Request) {
	if !allowDiscoveryRequest(r.RemoteAddr) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many discovery events"})
		return
	}
	var req struct {
		ViewerID uuid.UUID `json:"viewer_id"`
		Events   []struct {
			ListingID uuid.UUID `json:"listing_id"`
			EventType string    `json:"event_type"`
			Query     string    `json:"query"`
		} `json:"events"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.ViewerID == uuid.Nil || len(req.Events) == 0 || len(req.Events) > 25 {
		badRequest(w, map[string]string{"events": "viewer_id and 1 to 25 events are required"})
		return
	}
	events := make([]listings.DiscoveryEvent, 0, len(req.Events))
	for _, input := range req.Events {
		query := strings.TrimSpace(input.Query)
		if input.ListingID == uuid.Nil || len(query) > 120 || !discoveryEventType(input.EventType) {
			badRequest(w, map[string]string{"events": "invalid listing_id, event_type, or query"})
			return
		}
		id, err := uuid.NewV7()
		if err != nil {
			internalError(w)
			return
		}
		events = append(events, listings.DiscoveryEvent{ID: id, ViewerID: req.ViewerID, ListingID: input.ListingID, EventType: input.EventType, Query: query, CreatedAt: time.Now().UTC()})
	}
	accepted, err := h.Store.RecordDiscoveryEvents(r.Context(), events)
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]int{"accepted": accepted})
}

func discoveryEventType(value string) bool {
	switch value {
	case listings.DiscoveryImpression, listings.DiscoveryCardClick, listings.DiscoveryDetailView, listings.DiscoveryFavorite, listings.DiscoveryContactReveal:
		return true
	default:
		return false
	}
}

func (h Handler) PresignImage(w http.ResponseWriter, r *http.Request) {
	l, ok := h.ownerListing(w, r)
	if !ok {
		return
	}
	if h.Objects == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "image storage unavailable"})
		return
	}
	var req struct {
		ContentType string `json:"content_type"`
		ByteSize    int64  `json:"byte_size"`
	}
	if !decode(w, r, &req) {
		return
	}
	if err := images.Validate(req.ContentType, req.ByteSize); err != nil {
		badRequest(w, map[string]string{"image": err.Error()})
		return
	}
	existing, err := h.Store.ListImages(r.Context(), l.ID)
	if err != nil {
		internalError(w)
		return
	}
	if len(existing) >= 20 {
		badRequest(w, map[string]string{"images": "maximum 20 images"})
		return
	}
	key, err := images.NewKey(l.ID, req.ContentType)
	if err != nil {
		internalError(w)
		return
	}
	target, err := h.Objects.PresignPut(r.Context(), key, req.ContentType, req.ByteSize)
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"object_key": key, "upload_url": target.URL, "upload_headers": target.Headers})
}

func (h Handler) RegisterImage(w http.ResponseWriter, r *http.Request) {
	l, ok := h.ownerListing(w, r)
	if !ok {
		return
	}
	if h.Objects == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "image storage unavailable"})
		return
	}
	var req struct {
		ObjectKey   string  `json:"object_key"`
		ContentType string  `json:"content_type"`
		ByteSize    int64   `json:"byte_size"`
		Position    int16   `json:"position"`
		Alt         *string `json:"alt"`
	}
	if !decode(w, r, &req) {
		return
	}
	if err := images.Validate(req.ContentType, req.ByteSize); err != nil {
		badRequest(w, map[string]string{"image": err.Error()})
		return
	}
	if !strings.HasPrefix(req.ObjectKey, "listings/"+l.ID.String()+"/") {
		badRequest(w, map[string]string{"object_key": "invalid listing object key"})
		return
	}
	if err := h.Objects.Head(r.Context(), req.ObjectKey); err != nil {
		badRequest(w, map[string]string{"object_key": "uploaded object not found"})
		return
	}
	// Promote the object before persisting it. If this S3 control-plane call
	// fails, the client can retry safely and the lifecycle rule will still
	// remove the unregistered upload. A later database validation failure may
	// leave a private registered orphan, but it can never expire a real image.
	if err := h.Objects.MarkRegistered(r.Context(), req.ObjectKey); err != nil {
		internalError(w)
		return
	}
	id, err := uuid.NewV7()
	if err != nil {
		internalError(w)
		return
	}
	image, err := h.Store.AddImage(r.Context(), listings.Image{ID: id, ListingID: l.ID, ObjectKey: req.ObjectKey, ContentType: req.ContentType, ByteSize: req.ByteSize, Position: req.Position, Alt: req.Alt}, l.OwnerID)
	if errors.Is(err, listings.ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		badRequest(w, map[string]string{"image": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, image)
}

func (h Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	l, ok := h.ownerListing(w, r)
	if !ok {
		return
	}
	imageID, err := uuid.Parse(r.PathValue("imageId"))
	if err != nil {
		badRequest(w, map[string]string{"image_id": "invalid UUID"})
		return
	}
	err = h.Store.DeleteImage(r.Context(), l.ID, imageID, l.OwnerID)
	if errors.Is(err, listings.ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	owner, ok := middleware.UserID(r.Context())
	if !ok {
		unauthorized(w)
		return
	}
	var input listings.Input
	if !decode(w, r, &input) {
		return
	}
	id, err := uuid.NewV7()
	if err != nil {
		internalError(w)
		return
	}
	l := listings.Listing{ID: id, OwnerID: owner, Currency: "PHP", Status: listings.StatusDraft}
	l.Apply(input)
	if fields := listings.ValidateDraft(l); len(fields) > 0 {
		badRequest(w, fields)
		return
	}
	created, err := h.Store.Create(r.Context(), l)
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h Handler) Patch(w http.ResponseWriter, r *http.Request) {
	l, ok := h.ownerListing(w, r)
	if !ok {
		return
	}
	var input listings.Input
	if !decode(w, r, &input) {
		return
	}
	candidate := *l
	candidate.Apply(input)
	if fields := listings.ValidateDraft(candidate); len(fields) > 0 {
		badRequest(w, fields)
		return
	}
	if candidate.Status == listings.StatusPublished {
		images, err := h.Store.ListImages(r.Context(), candidate.ID)
		if err != nil {
			internalError(w)
			return
		}
		if fields := listings.MissingPublishFields(candidate, len(images)); len(fields) > 0 {
			badRequest(w, fields)
			return
		}
	}
	saved, err := h.Store.SaveOwner(r.Context(), candidate, candidate.OwnerID)
	if errors.Is(err, listings.ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h Handler) Publish(w http.ResponseWriter, r *http.Request) {
	l, ok := h.ownerListing(w, r)
	if !ok {
		return
	}
	if l.Status != listings.StatusDraft {
		badRequest(w, map[string]string{"status": "only drafts can be published"})
		return
	}
	images, err := h.Store.ListImages(r.Context(), l.ID)
	if err != nil {
		internalError(w)
		return
	}
	if fields := listings.MissingPublishFields(*l, len(images)); len(fields) > 0 {
		badRequest(w, fields)
		return
	}
	now := time.Now().UTC()
	l.Status, l.PublishedAt = listings.StatusPublished, &now
	saved, err := h.Store.SaveOwner(r.Context(), *l, l.OwnerID)
	if errors.Is(err, listings.ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (h Handler) Unpublish(w http.ResponseWriter, r *http.Request) {
	l, ok := h.ownerListing(w, r)
	if !ok {
		return
	}
	if l.Status != listings.StatusPublished {
		badRequest(w, map[string]string{"status": "only published listings can be unpublished"})
		return
	}
	l.Status, l.PublishedAt = listings.StatusDraft, nil
	h.save(w, r, *l)
}

func (h Handler) Close(w http.ResponseWriter, r *http.Request) {
	l, ok := h.ownerListing(w, r)
	if !ok {
		return
	}
	if l.Status != listings.StatusPublished {
		badRequest(w, map[string]string{"status": "only published listings can be closed"})
		return
	}
	l.Status = listings.StatusClosed
	h.save(w, r, *l)
}

func (h Handler) Archive(w http.ResponseWriter, r *http.Request) {
	l, ok := h.ownerListing(w, r)
	if !ok {
		return
	}
	l.Status = listings.StatusArchived
	h.save(w, r, *l)
}

func (h Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	l, err := h.Store.GetPublished(r.Context(), id)
	if errors.Is(err, listings.ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w)
		return
	}
	h.writeDetail(w, r, *l)
}

func (h Handler) ListPublic(w http.ResponseWriter, r *http.Request) {
	f, fields := parseFilter(r)
	if len(fields) > 0 {
		badRequest(w, fields)
		return
	}
	items, err := h.Store.ListPublished(r.Context(), f)
	if err != nil {
		internalError(w)
		return
	}
	response := make([]any, 0, len(items))
	for _, item := range items {
		images, err := h.Store.ListImages(r.Context(), item.ID)
		if err != nil {
			internalError(w)
			return
		}
		// Contact details are displayed only after a buyer opens one listing.
		item.SellerName, item.ContactPhone, item.ContactEmail = "", "", ""
		response = append(response, publicDetail{Listing: item, Images: images})
	}
	next := ""
	if len(items) == f.Limit && len(items) > 0 {
		next = encodeCursor(items[len(items)-1], f.Sort)
	}
	writeJSON(w, http.StatusOK, map[string]any{"listings": response, "next_cursor": next})
}

func (h Handler) Discover(w http.ResponseWriter, r *http.Request) {
	f, fields := parseFilter(r)
	if len(fields) > 0 {
		badRequest(w, fields)
		return
	}
	f.Cursor, f.Limit = nil, 6
	sections := []discoverySection{}
	add := func(id, title, reason string, filter listings.ListFilter) bool {
		items, err := h.Store.ListPublished(r.Context(), filter)
		if err != nil {
			internalError(w)
			return false
		}
		cards := make([]any, 0, len(items))
		for _, item := range items {
			images, err := h.Store.ListImages(r.Context(), item.ID)
			if err != nil {
				internalError(w)
				return false
			}
			item.SellerName, item.ContactPhone, item.ContactEmail = "", "", ""
			cards = append(cards, publicDetail{Listing: item, Images: images})
		}
		if len(cards) > 0 {
			sections = append(sections, discoverySection{ID: id, Title: title, Reason: reason, Listings: cards})
		}
		return true
	}
	if !add("new", "Newly listed", "Fresh properties added to Hacuba.", listings.ListFilter{Sort: "newest", Limit: 6}) {
		return
	}
	if viewerID, ok := h.trustedDiscoveryViewer(r); ok {
		profile, err := h.Store.GetDiscoveryProfile(r.Context(), viewerID)
		if err != nil && !errors.Is(err, listings.ErrNotFound) {
			internalError(w)
			return
		}
		if err == nil {
			viewed, err := h.Store.ListViewedListingIDs(r.Context(), viewerID)
			if err != nil {
				internalError(w)
				return
			}
			if !h.addPersonalizedDiscovery(w, r, &sections, profile, viewed) {
				return
			}
		}
	}
	if f.City != "" && !add("city", "In "+f.City, "Published properties in your selected city.", f) {
		return
	}
	if f.Type != "" && !add("type", "More "+f.Type+" listings", "Matches your property type.", f) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sections": sections})
}

func (h Handler) trustedDiscoveryViewer(r *http.Request) (uuid.UUID, bool) {
	if len(h.DiscoveryProxySecret) == 0 || subtle.ConstantTimeCompare([]byte(r.Header.Get(discoveryProxyHeader)), h.DiscoveryProxySecret) != 1 {
		return uuid.Nil, false
	}
	viewerID, err := uuid.Parse(r.Header.Get(discoveryViewerHeader))
	if err != nil || viewerID == uuid.Nil {
		return uuid.Nil, false
	}
	return viewerID, true
}

func (h Handler) addPersonalizedDiscovery(w http.ResponseWriter, r *http.Request, sections *[]discoverySection, profile listings.DiscoveryProfile, viewed []uuid.UUID) bool {
	filter := listings.ListFilter{City: profile.City, Type: profile.PropertyType, Sort: "newest", Limit: 24}
	if profile.PriceCentavos != nil {
		min, max := *profile.PriceCentavos*70/100, *profile.PriceCentavos*130/100
		filter.MinPrice, filter.MaxPrice = &min, &max
	}
	candidates, err := h.Store.ListPublished(r.Context(), filter)
	if err != nil {
		internalError(w)
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(viewed))
	for _, id := range viewed {
		seen[id] = struct{}{}
	}
	filtered := make([]listings.Listing, 0, len(candidates))
	for _, candidate := range candidates {
		if _, ok := seen[candidate.ID]; !ok {
			filtered = append(filtered, candidate)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		left, right := personalizedScore(filtered[i], profile), personalizedScore(filtered[j], profile)
		if left == right {
			return filtered[i].ID.String() < filtered[j].ID.String()
		}
		return left > right
	})
	cards := make([]any, 0, 6)
	for _, item := range filtered {
		images, err := h.Store.ListImages(r.Context(), item.ID)
		if err != nil {
			internalError(w)
			return false
		}
		item.SellerName, item.ContactPhone, item.ContactEmail = "", "", ""
		cards = append(cards, publicDetail{Listing: item, Images: images})
		if len(cards) == 6 {
			break
		}
	}
	if len(cards) > 0 {
		*sections = append(*sections, discoverySection{ID: "for-you", Title: "For you", Reason: discoveryProfileReason(profile), Listings: cards})
	}
	return true
}

func personalizedScore(candidate listings.Listing, profile listings.DiscoveryProfile) int64 {
	score := int64(0)
	if candidate.City == profile.City {
		score += 100
	}
	if candidate.PropertyType == profile.PropertyType {
		score += 80
	}
	if candidate.PriceCentavos != nil && profile.PriceCentavos != nil {
		delta := *candidate.PriceCentavos - *profile.PriceCentavos
		if delta < 0 {
			delta = -delta
		}
		score -= delta / 1_000_000
	}
	if candidate.Bedrooms != nil && profile.Bedrooms != nil && *candidate.Bedrooms == *profile.Bedrooms {
		score += 20
	}
	if candidate.PublishedAt != nil && candidate.PublishedAt.After(time.Now().UTC().AddDate(0, 0, -14)) {
		score += 10
	}
	return score
}

func discoveryProfileReason(profile listings.DiscoveryProfile) string {
	if profile.City != "" && profile.PropertyType != "" {
		return "Based on properties you explored in " + profile.City + "."
	}
	if profile.City != "" {
		return "Based on properties you explored in " + profile.City + "."
	}
	return "Based on properties you explored on Hacuba."
}

func (h Handler) Related(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	source, err := h.Store.GetPublished(r.Context(), id)
	if errors.Is(err, listings.ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w)
		return
	}
	candidates, err := h.Store.ListPublished(r.Context(), listings.ListFilter{Type: source.PropertyType, City: source.City, Sort: "newest", Limit: 24})
	if err != nil {
		internalError(w)
		return
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return relatedScore(candidates[i], *source) > relatedScore(candidates[j], *source)
	})
	response := make([]any, 0, 6)
	for _, item := range candidates {
		if item.ID == source.ID {
			continue
		}
		images, err := h.Store.ListImages(r.Context(), item.ID)
		if err != nil {
			internalError(w)
			return
		}
		item.SellerName, item.ContactPhone, item.ContactEmail = "", "", ""
		response = append(response, publicDetail{Listing: item, Images: images})
		if len(response) == 6 {
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"reason": "Similar " + source.PropertyType + " listings in " + source.City, "listings": response})
}

func relatedScore(candidate, source listings.Listing) int64 {
	score := int64(0)
	if candidate.City == source.City {
		score += 100
	}
	if candidate.PropertyType == source.PropertyType {
		score += 80
	}
	if candidate.PriceCentavos != nil && source.PriceCentavos != nil {
		delta := *candidate.PriceCentavos - *source.PriceCentavos
		if delta < 0 {
			delta = -delta
		}
		score -= delta / 1_000_000
	}
	if candidate.Bedrooms != nil && source.Bedrooms != nil && *candidate.Bedrooms == *source.Bedrooms {
		score += 20
	}
	return score
}

func (h Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	owner, ok := middleware.UserID(r.Context())
	if !ok {
		unauthorized(w)
		return
	}
	items, err := h.Store.ListOwner(r.Context(), owner)
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"listings": items})
}

func (h Handler) ownerListing(w http.ResponseWriter, r *http.Request) (*listings.Listing, bool) {
	owner, ok := middleware.UserID(r.Context())
	if !ok {
		unauthorized(w)
		return nil, false
	}
	id, ok := pathID(w, r)
	if !ok {
		return nil, false
	}
	l, err := h.Store.GetOwner(r.Context(), id, owner)
	if errors.Is(err, listings.ErrNotFound) {
		notFound(w)
		return nil, false
	}
	if err != nil {
		internalError(w)
		return nil, false
	}
	return l, true
}

func (h Handler) save(w http.ResponseWriter, r *http.Request, l listings.Listing) {
	saved, err := h.Store.SaveOwner(r.Context(), l, l.OwnerID)
	if errors.Is(err, listings.ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

type publicDetail struct {
	listings.Listing
	Images []listings.Image `json:"images"`
}

func (h Handler) writeDetail(w http.ResponseWriter, r *http.Request, l listings.Listing) {
	images, err := h.Store.ListImages(r.Context(), l.ID)
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, publicDetail{Listing: l, Images: images})
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		badRequest(w, map[string]string{"body": "invalid JSON"})
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		badRequest(w, map[string]string{"body": "invalid JSON"})
		return false
	}
	return true
}

func parseFilter(r *http.Request) (listings.ListFilter, map[string]string) {
	q, fields := r.URL.Query(), map[string]string{}
	f := listings.ListFilter{Mode: q.Get("mode"), Type: q.Get("type"), City: q.Get("city"), Barangay: q.Get("barangay"), Query: q.Get("q"), Sort: q.Get("sort"), Limit: 20}
	if f.Mode != "" && f.Mode != listings.ModeSale && f.Mode != listings.ModeRent {
		fields["mode"] = "must be for_sale or for_rent"
	}
	if f.Type != "" && f.Type != listings.TypeHouse && f.Type != listings.TypeApartment && f.Type != listings.TypeCondo && f.Type != listings.TypeLot && f.Type != listings.TypeCommercial {
		fields["type"] = "invalid property type"
	}
	if f.Sort == "" {
		f.Sort = "newest"
	}
	if f.Sort != "newest" && f.Sort != "price_asc" && f.Sort != "price_desc" {
		fields["sort"] = "must be newest, price_asc, or price_desc"
	}
	parseInt64 := func(name string) *int64 {
		raw := q.Get(name)
		if raw == "" {
			return nil
		}
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 0 {
			fields[name] = "must be a non-negative integer"
			return nil
		}
		return &value
	}
	f.MinPrice, f.MaxPrice = parseInt64("min_price"), parseInt64("max_price")
	if f.MinPrice != nil && f.MaxPrice != nil && *f.MinPrice > *f.MaxPrice {
		fields["price"] = "min_price must not exceed max_price"
	}
	if raw := q.Get("beds"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 16)
		if err != nil || value < 0 {
			fields["beds"] = "must be a non-negative integer"
		} else {
			beds := int16(value)
			f.Beds = &beds
		}
	}
	if raw := q.Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 50 {
			fields["limit"] = "must be between 1 and 50"
		} else {
			f.Limit = value
		}
	}
	if raw := q.Get("cursor"); raw != "" {
		cursor, err := decodeCursor(raw, f.Sort)
		if err != nil {
			fields["cursor"] = "invalid cursor"
		} else {
			f.Cursor = cursor
		}
	}
	return f, fields
}

type browseCursor struct {
	ID            uuid.UUID `json:"id"`
	Sort          string    `json:"sort"`
	PublishedAt   time.Time `json:"published_at"`
	PriceCentavos int64     `json:"price_centavos"`
}

func encodeCursor(l listings.Listing, sort string) string {
	cursor := browseCursor{ID: l.ID, Sort: sort}
	if l.PublishedAt != nil {
		cursor.PublishedAt = *l.PublishedAt
	}
	if l.PriceCentavos != nil {
		cursor.PriceCentavos = *l.PriceCentavos
	}
	raw, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeCursor(raw, sort string) (*listings.ListCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	var cursor browseCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil || cursor.ID == uuid.Nil || cursor.Sort != sort {
		return nil, errors.New("invalid cursor")
	}
	if sort == "newest" && cursor.PublishedAt.IsZero() {
		return nil, errors.New("invalid cursor")
	}
	if (sort == "price_asc" || sort == "price_desc") && cursor.PriceCentavos < 1 {
		return nil, errors.New("invalid cursor")
	}
	return &listings.ListCursor{ID: cursor.ID, PublishedAt: cursor.PublishedAt, PriceCentavos: cursor.PriceCentavos}, nil
}

func pathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		badRequest(w, map[string]string{"id": "invalid UUID"})
		return uuid.Nil, false
	}
	return id, true
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func badRequest(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bad request", "fields": fields})
}
func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}
func unauthorized(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}
func internalError(w http.ResponseWriter) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
