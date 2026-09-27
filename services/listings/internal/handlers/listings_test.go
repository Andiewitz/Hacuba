package handlers_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/authjwt"
	"github.com/hacuba/listings/internal/config"
	"github.com/hacuba/listings/internal/images"
	"github.com/hacuba/listings/internal/listings"
	"github.com/hacuba/listings/internal/server"
)

type testObjectStore struct {
	lastKey          string
	markedRegistered bool
}

func (s *testObjectStore) PresignPut(_ context.Context, key, _ string, _ int64) (images.UploadTarget, error) {
	s.lastKey = key
	return images.UploadTarget{URL: "https://uploads.example.test/" + key, Headers: map[string]string{"x-amz-tagging": "state=unregistered"}}, nil
}

func (s *testObjectStore) Head(_ context.Context, key string) error {
	if key != s.lastKey {
		return errors.New("unknown object")
	}
	return nil
}

func (s *testObjectStore) MarkRegistered(_ context.Context, key string) error {
	if key != s.lastKey {
		return errors.New("unknown object")
	}
	s.markedRegistered = true
	return nil
}

func TestDraftPublishValidationAndOwnerIsolation(t *testing.T) {
	secret := []byte("listings-test-secret-must-be-32-bytes!!")
	store := listings.NewMemoryStore()
	mux := server.NewMux(config.Config{JWTSecret: secret}, store)
	owner, other := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	ownerToken, ownerCSRF := sellerToken(t, secret, owner)
	otherToken, otherCSRF := sellerToken(t, secret, other)

	post := func(method, target, token, csrf, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-CSRF-Token", csrf)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrf})
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	created := post(http.MethodPost, "/listings", ownerToken, ownerCSRF, `{}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create draft = %d: %s", created.Code, created.Body.String())
	}
	var draft listings.Listing
	_ = json.NewDecoder(created.Body).Decode(&draft)
	failed := post(http.MethodPost, "/listings/"+draft.ID.String()+"/publish", ownerToken, ownerCSRF, "")
	if failed.Code != http.StatusBadRequest || !strings.Contains(failed.Body.String(), "images") || !strings.Contains(failed.Body.String(), "title") {
		t.Fatalf("incomplete publish = %d: %s", failed.Code, failed.Body.String())
	}
	if forbidden := post(http.MethodPatch, "/listings/"+draft.ID.String(), otherToken, otherCSRF, `{ "title": "stolen listing" }`); forbidden.Code != http.StatusNotFound {
		t.Fatalf("cross-owner patch = %d, want 404", forbidden.Code)
	}
}

func TestBuyerCannotCreateListing(t *testing.T) {
	secret := []byte("listings-test-secret-must-be-32-bytes!!")
	mux := server.NewMux(config.Config{JWTSecret: secret}, listings.NewMemoryStore())
	csrf := "csrf-value"
	sum := sha256.Sum256([]byte(csrf))
	token, err := authjwt.IssueAccess(secret, uuid.Must(uuid.NewV7()), "buyer", hex.EncodeToString(sum[:]), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/listings", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrf})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("buyer create = %d, want 403", rec.Code)
	}
}

func TestImageUploadPassesSignedHeadersAndRegistersObject(t *testing.T) {
	secret := []byte("listings-test-secret-must-be-32-bytes!!")
	store, objects := listings.NewMemoryStore(), &testObjectStore{}
	mux := server.NewMuxWithObjects(config.Config{JWTSecret: secret}, store, objects)
	owner := uuid.Must(uuid.NewV7())
	token, csrf := sellerToken(t, secret, owner)
	request := func(method, target, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-CSRF-Token", csrf)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrf})
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	created := request(http.MethodPost, "/listings", `{}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create draft = %d: %s", created.Code, created.Body.String())
	}
	var draft listings.Listing
	if err := json.NewDecoder(created.Body).Decode(&draft); err != nil {
		t.Fatal(err)
	}
	presigned := request(http.MethodPost, "/listings/"+draft.ID.String()+"/images/presign", `{"content_type":"image/jpeg","byte_size":3}`)
	if presigned.Code != http.StatusOK || !strings.Contains(presigned.Body.String(), "x-amz-tagging") {
		t.Fatalf("presign = %d: %s", presigned.Code, presigned.Body.String())
	}
	var target struct {
		ObjectKey string `json:"object_key"`
	}
	if err := json.NewDecoder(presigned.Body).Decode(&target); err != nil {
		t.Fatal(err)
	}
	registered := request(http.MethodPost, "/listings/"+draft.ID.String()+"/images", `{"object_key":"`+target.ObjectKey+`","content_type":"image/jpeg","byte_size":3,"position":0}`)
	if registered.Code != http.StatusCreated || !objects.markedRegistered {
		t.Fatalf("register = %d, marked=%t: %s", registered.Code, objects.markedRegistered, registered.Body.String())
	}
}

func TestPublicDetailIncludesContactButBrowseDoesNot(t *testing.T) {
	store := listings.NewMemoryStore()
	listingID, ownerID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	price := int64(750_000_000)
	now := time.Now().UTC()
	_, err := store.Create(t.Context(), listings.Listing{
		ID: listingID, OwnerID: ownerID, ListingMode: listings.ModeSale, PropertyType: listings.TypeHouse,
		Title: "Family home near Cebu IT Park", Description: "A public detail test listing.", PriceCentavos: &price,
		Currency: "PHP", City: "Cebu City", SellerName: "Mika Santos", ContactPhone: "+63 917 555 0142",
		ContactEmail: "mika.santos@example.test", Status: listings.StatusPublished, PublishedAt: &now,
	})
	if err != nil {
		t.Fatal(err)
	}
	mux := server.NewMux(config.Config{JWTSecret: []byte("listings-test-secret-must-be-32-bytes!!")}, store)

	detailRequest := httptest.NewRequest(http.MethodGet, "/listings/"+listingID.String(), nil)
	detail := httptest.NewRecorder()
	mux.ServeHTTP(detail, detailRequest)
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), "mika.santos@example.test") {
		t.Fatalf("public detail = %d: %s", detail.Code, detail.Body.String())
	}

	browseRequest := httptest.NewRequest(http.MethodGet, "/listings", nil)
	browse := httptest.NewRecorder()
	mux.ServeHTTP(browse, browseRequest)
	if browse.Code != http.StatusOK || strings.Contains(browse.Body.String(), "mika.santos@example.test") {
		t.Fatalf("public browse = %d: %s", browse.Code, browse.Body.String())
	}
}

func TestPublicBrowseUsesOpaqueKeysetCursor(t *testing.T) {
	store := listings.NewMemoryStore()
	now := time.Now().UTC().Truncate(time.Second)
	price := int64(100_000_000)
	owner := uuid.Must(uuid.NewV7())
	ids := []uuid.UUID{uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())}
	for i, id := range ids {
		publishedAt := now.Add(time.Duration(i) * time.Minute)
		_, err := store.Create(t.Context(), listings.Listing{
			ID: id, OwnerID: owner, ListingMode: listings.ModeSale, PropertyType: listings.TypeHouse,
			Title: "Cursor test home", Description: "A listing used to verify keyset pagination.", PriceCentavos: &price,
			Currency: "PHP", City: "Cebu City", Status: listings.StatusPublished, PublishedAt: &publishedAt,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	mux := server.NewMux(config.Config{JWTSecret: []byte("listings-test-secret-must-be-32-bytes!!")}, store)
	type page struct {
		Listings []struct {
			ID uuid.UUID `json:"id"`
		} `json:"listings"`
		NextCursor string `json:"next_cursor"`
	}
	get := func(target string) page {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("browse = %d: %s", rec.Code, rec.Body.String())
		}
		var result page
		if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}

	first := get("/listings?limit=2")
	if len(first.Listings) != 2 || first.NextCursor == "" || first.Listings[0].ID != ids[2] || first.Listings[1].ID != ids[1] {
		t.Fatalf("first page = %#v", first)
	}
	second := get("/listings?limit=2&cursor=" + first.NextCursor)
	if len(second.Listings) != 1 || second.Listings[0].ID != ids[0] || second.Listings[0].ID == first.Listings[0].ID || second.Listings[0].ID == first.Listings[1].ID {
		t.Fatalf("second page repeated or skipped rows: %#v", second)
	}

	bad := httptest.NewRecorder()
	mux.ServeHTTP(bad, httptest.NewRequest(http.MethodGet, "/listings?sort=price_asc&cursor="+first.NextCursor, nil))
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("cursor used with another sort = %d: %s", bad.Code, bad.Body.String())
	}
}

func TestDiscoveryEventsAcceptOnlyPublishedListings(t *testing.T) {
	store := listings.NewMemoryStore()
	listingID, ownerID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	price, publishedAt := int64(100_000_000), time.Now().UTC()
	_, err := store.Create(t.Context(), listings.Listing{ID: listingID, OwnerID: ownerID, ListingMode: listings.ModeSale, PropertyType: listings.TypeHouse, Title: "Discovery event home", Description: "Published listing used for discovery event validation.", PriceCentavos: &price, Currency: "PHP", City: "Cebu City", Status: listings.StatusPublished, PublishedAt: &publishedAt})
	if err != nil {
		t.Fatal(err)
	}
	mux := server.NewMux(config.Config{JWTSecret: []byte("listings-test-secret-must-be-32-bytes!!")}, store)
	body := `{"viewer_id":"` + uuid.Must(uuid.NewV7()).String() + `","events":[{"listing_id":"` + listingID.String() + `","event_type":"detail_view","query":"cebu"}]}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/discovery/events", strings.NewReader(body)))
	if rec.Code != http.StatusAccepted || !strings.Contains(rec.Body.String(), `"accepted":1`) {
		t.Fatalf("discovery events = %d: %s", rec.Code, rec.Body.String())
	}
}

func sellerToken(t *testing.T, secret []byte, id uuid.UUID) (string, string) {
	t.Helper()
	csrf := "csrf-value-" + id.String()
	sum := sha256.Sum256([]byte(csrf))
	token, err := authjwt.IssueAccess(secret, id, "seller", hex.EncodeToString(sum[:]), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return token, csrf
}
