package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hacuba/auth/internal/middleware"
	"github.com/hacuba/auth/internal/users"
	"github.com/hacuba/authjwt"
)

func TestBecomeSellerIssuesSellerClaimAndIsIdempotent(t *testing.T) {
	cfg, store := testConfig(), users.NewMemoryStore()
	registered := doJSON(t, Register(cfg, store), http.MethodPost, "/auth/register", map[string]string{"email": uniqueTestEmail("seller"), "password": "validpassword123"})
	if registered.Code != http.StatusCreated {
		t.Fatalf("register = %d", registered.Code)
	}
	var initial SessionResponse
	_ = json.NewDecoder(registered.Body).Decode(&initial)

	upgrade := func(token, csrf string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/auth/become-seller", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-CSRF-Token", csrf)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrf})
		rec := httptest.NewRecorder()
		middleware.Authenticate(cfg.JWTSecret, store, middleware.RequireCSRF(BecomeSeller(cfg, store))).ServeHTTP(rec, req)
		return rec
	}
	first := upgrade(initial.AccessToken, initial.CSRFToken)
	if first.Code != http.StatusOK {
		t.Fatalf("upgrade = %d: %s", first.Code, first.Body.String())
	}
	var upgraded SessionResponse
	_ = json.NewDecoder(first.Body).Decode(&upgraded)
	if upgraded.User.Role != users.RoleSeller {
		t.Fatalf("role = %q, want seller", upgraded.User.Role)
	}
	claims, err := authjwt.VerifyAccess(cfg.JWTSecret, upgraded.AccessToken)
	if err != nil || claims.Role != users.RoleSeller {
		t.Fatalf("seller claim missing: claims=%+v err=%v", claims, err)
	}
	if second := upgrade(upgraded.AccessToken, upgraded.CSRFToken); second.Code != http.StatusOK {
		t.Fatalf("idempotent upgrade = %d", second.Code)
	}
}
