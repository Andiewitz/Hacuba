package handlers

// Unit-level flow coverage for register/login/me. Router-level contract
// (status codes, cookie flags, rate limits, full mux wiring) lives in
// internal/server/router_test.go through server.NewMux — the same
// constructor cmd/server uses, so the two can never drift.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/middleware"
	"github.com/hacuba/auth/internal/users"
)

func testConfig() config.Config {
	return config.Config{
		JWTSecret:    []byte("test-secret-must-be-32-bytes-minimum!!"),
		AccessTTL:    15 * time.Minute,
		RefreshTTL:   7 * 24 * time.Hour,
		CookieSecure: false,
	}
}

func doJSON(t *testing.T, handler http.HandlerFunc, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, target, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestRegisterLoginMeFlow(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()

	// Register auto-logs-in and returns a session.
	reg := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": "user@example.com", "password": "validpassword123"})
	if reg.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body = %s", reg.Code, reg.Body.String())
	}
	var sess SessionResponse
	if err := json.NewDecoder(reg.Body).Decode(&sess); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if sess.AccessToken == "" || sess.CSRFToken == "" {
		t.Fatal("expected access + CSRF tokens")
	}
	if sess.User.ID.String() == "" || sess.User.Email != "user@example.com" {
		t.Fatalf("unexpected user payload: %+v", sess.User)
	}

	// Duplicate registration conflicts.
	dup := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": "user@example.com", "password": "validpassword123"})
	if dup.Code != http.StatusConflict {
		t.Errorf("duplicate register status = %d, want 409", dup.Code)
	}

	// Login with wrong password is a generic 401 (no enumeration).
	bad := doJSON(t, Login(cfg, store), "POST", "/auth/login",
		map[string]string{"email": "user@example.com", "password": "wrongpassword999"})
	if bad.Code != http.StatusUnauthorized {
		t.Errorf("bad login status = %d, want 401", bad.Code)
	}

	// Login succeeds.
	ok := doJSON(t, Login(cfg, store), "POST", "/auth/login",
		map[string]string{"email": "user@example.com", "password": "validpassword123"})
	if ok.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", ok.Code, ok.Body.String())
	}
}

func TestMeReadsOnlyOwnRow(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()

	// Two users register independently.
	recA := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": "a@example.com", "password": "validpassword123"})
	var sessA SessionResponse
	_ = json.NewDecoder(recA.Body).Decode(&sessA)

	recB := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": "b@example.com", "password": "validpassword123"})
	var sessB SessionResponse
	_ = json.NewDecoder(recB.Body).Decode(&sessB)

	if sessA.User.ID == sessB.User.ID {
		t.Fatal("each user must have its own UUID")
	}

	// GET /auth/me through the real Authenticate middleware: A's token
	// returns A, B's token returns B — there is no parameter to smuggle
	// another user's ID through, which is the isolation guarantee.
	protected := middleware.Authenticate(cfg.JWTSecret, store, Me(store))

	getMe := func(token string) (int, UserPublic) {
		req := httptest.NewRequest("GET", "/auth/me", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		var u UserPublic
		if rec.Code == http.StatusOK {
			_ = json.NewDecoder(rec.Body).Decode(&u)
		}
		return rec.Code, u
	}

	if code, u := getMe(sessA.AccessToken); code != http.StatusOK || u.Email != "a@example.com" {
		t.Errorf("A me: code=%d user=%+v", code, u)
	}
	if code, u := getMe(sessB.AccessToken); code != http.StatusOK || u.Email != "b@example.com" {
		t.Errorf("B me: code=%d user=%+v", code, u)
	}
	if code, _ := getMe(""); code != http.StatusUnauthorized {
		t.Errorf("anonymous me: code=%d, want 401", code)
	}
	if code, _ := getMe("tampered.token.value"); code != http.StatusUnauthorized {
		t.Errorf("tampered me: code=%d, want 401", code)
	}
}
