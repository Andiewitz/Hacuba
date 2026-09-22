package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/listings/internal/listings"
	"github.com/hacuba/listings/internal/middleware"
)

type Handler struct{ Store listings.Store }

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
		response = append(response, publicDetail{Listing: item, Images: images})
	}
	next := ""
	if len(items) == f.Limit && len(items) > 0 {
		next = base64.RawURLEncoding.EncodeToString([]byte(items[len(items)-1].ID.String()))
	}
	writeJSON(w, http.StatusOK, map[string]any{"listings": response, "next_cursor": next})
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
		decoded, err := base64.RawURLEncoding.DecodeString(raw)
		if err != nil {
			fields["cursor"] = "invalid cursor"
		} else if _, err := uuid.Parse(string(decoded)); err != nil {
			fields["cursor"] = "invalid cursor"
		} else {
			f.Cursor = string(decoded)
		}
	}
	return f, fields
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
