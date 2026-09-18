package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/hacuba/auth/internal/tokens"
	"github.com/hacuba/auth/internal/users"
)

func TestRequireOwnerAllowsSelfDeniesOther(t *testing.T) {
	caller := uuid.Must(uuid.NewV7())
	other := uuid.Must(uuid.NewV7())

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := RequireOwner(func(r *http.Request) (uuid.UUID, bool) {
		return OwnerIDFromPath(r, "id")
	}, next)

	mux := http.NewServeMux()
	mux.Handle("/users/{id}", wrapped)

	withCaller := func(target uuid.UUID) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/users/"+target.String(), nil)
		user := &users.User{ID: caller, Email: "a@example.com"}
		req = req.WithContext(withAuthContext(req.Context(), user, ""))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	if rec := withCaller(caller); rec.Code != http.StatusOK {
		t.Errorf("self access: got %d, want 200", rec.Code)
	}
	// Cross-user access must be 404 (not 403) to hide existence.
	if rec := withCaller(other); rec.Code != http.StatusNotFound {
		t.Errorf("cross-user access: got %d, want 404", rec.Code)
	}
}

func TestAuthenticateRejectsBadTokens(t *testing.T) {
	secret := []byte("test-secret-must-be-32-bytes-minimum!!")
	store := users.NewMemoryStore()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	protected := Authenticate(secret, store, next)

	cases := map[string]string{
		"missing header": "",
		"malformed":      "not-a-bearer-token",
		"tampered jwt":   "Bearer tampered.token.value",
	}
	for name, header := range cases {
		req := httptest.NewRequest("GET", "/auth/me", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: got %d, want 401", name, rec.Code)
		}
	}
}

func TestRequireCSRFDoubleSubmit(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := RequireCSRF(next)

	// Valid double-submit passes.
	raw, _, _ := tokens.NewOpaqueToken()
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: raw})
	req.Header.Set("X-CSRF-Token", raw)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("valid CSRF: got %d, want 200", rec.Code)
	}

	// Mismatched header/cookie fails closed.
	req2 := httptest.NewRequest("POST", "/auth/refresh", nil)
	req2.AddCookie(&http.Cookie{Name: "csrf_token", Value: raw})
	req2.Header.Set("X-CSRF-Token", "different-value")
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("mismatched CSRF: got %d, want 403", rec2.Code)
	}

	// Safe methods skip validation.
	req3 := httptest.NewRequest("GET", "/auth/me", nil)
	rec3 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Errorf("GET passthrough: got %d, want 200", rec3.Code)
	}
}
