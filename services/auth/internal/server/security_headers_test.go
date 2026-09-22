package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/auth/internal/tokens"
	"github.com/hacuba/auth/internal/users"
)

func TestSecurityHeadersEverywhere(t *testing.T) {
	mux := NewMux(testConfig(), users.NewMemoryStore())

	targets := []struct {
		method string
		target string
	}{
		{"GET", "/healthz"},
		{"POST", "/auth/register"},
		{"POST", "/auth/login"},
		{"POST", "/auth/refresh"},
		{"GET", "/auth/me"},
		{"GET", "/no-such-route"},
	}
	for _, tc := range targets {
		req := httptest.NewRequest(tc.method, tc.target, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s %s: X-Content-Type-Options = %q, want nosniff",
				tc.method, tc.target, got)
		}
	}
}

func TestForgedTokensRejectedThroughMux(t *testing.T) {
	cfg := testConfig()
	mux := NewMux(cfg, users.NewMemoryStore())

	// alg=none token: unsigned, must never authenticate.
	noneAlg := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
		base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"`+uuid.Must(uuid.NewV7()).String()+`"}`)) + "."

	// Properly signed but expired.
	expired, err := tokens.IssueAccess(cfg.JWTSecret, uuid.Must(uuid.NewV7()), "buyer", "csrf", -time.Minute)
	if err != nil {
		t.Fatalf("mint expired: %v", err)
	}

	// Properly signed but for the wrong audience.
	wrongAud, err := tokens.IssueAccess([]byte("wrong-secret-must-be-32-bytes-minimum!"),
		uuid.Must(uuid.NewV7()), "buyer", "csrf", time.Minute)
	if err != nil {
		t.Fatalf("mint wrong-aud: %v", err)
	}

	// Valid signature, random garbage appended.
	valid, err := tokens.IssueAccess(cfg.JWTSecret, uuid.Must(uuid.NewV7()), "buyer", "csrf", time.Minute)
	if err != nil {
		t.Fatalf("mint valid: %v", err)
	}
	tampered := valid[:len(valid)-2] + "xx"

	// Random bytes as the token.
	raw := make([]byte, 24)
	_, _ = rand.Read(raw)
	garbage := base64.RawURLEncoding.EncodeToString(raw)

	for name, token := range map[string]string{
		"none algorithm": noneAlg,
		"expired":        expired,
		"wrong secret":   wrongAud,
		"tampered sig":   tampered,
		"garbage":        garbage,
		"empty bearer":   "",
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/auth/me", nil)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("got %d, want 401 (body=%s)", rec.Code, rec.Body.String())
			}
			if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("401 missing nosniff header")
			}
		})
	}
}

func TestUnknownRoutesAre404NotLeaks(t *testing.T) {
	mux := NewMux(testConfig(), users.NewMemoryStore())
	for _, target := range []string{"/auth/admin", "/auth/users", "/.env", "/auth/me/extra"} {
		req := httptest.NewRequest("GET", target, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", target, rec.Code)
		}
		if body := rec.Body.String(); len(body) > 0 && containsStackTrace(body) {
			t.Errorf("GET %s leaks internals: %s", target, body)
		}
	}
}

func containsStackTrace(body string) bool {
	for _, marker := range []string{"goroutine", ".go:", "panic:", "postgres://"} {
		if len(body) >= len(marker) {
			for i := 0; i+len(marker) <= len(body); i++ {
				if body[i:i+len(marker)] == marker {
					return true
				}
			}
		}
	}
	return false
}
