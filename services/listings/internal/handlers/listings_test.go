package handlers_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/authjwt"
	"github.com/hacuba/listings/internal/config"
	"github.com/hacuba/listings/internal/listings"
	"github.com/hacuba/listings/internal/server"
)

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
