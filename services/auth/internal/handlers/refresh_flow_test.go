package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/middleware"
	"github.com/hacuba/auth/internal/tokens"
	"github.com/hacuba/auth/internal/users"
)

func uniqueTestEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@example.com", prefix, uuid.Must(uuid.NewV7()).String())
}

// registerSession registers a user and returns the session plus raw cookies.
func registerSession(t *testing.T, cfg config.Config, store users.Store, email string) (SessionResponse, []*http.Cookie) {
	t.Helper()
	rec := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": email, "password": "validpassword123"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s = %d (%s)", email, rec.Code, rec.Body.String())
	}
	var sess SessionResponse
	if err := json.NewDecoder(rec.Body).Decode(&sess); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	return sess, rec.Result().Cookies()
}

func cookieValue(cookies []*http.Cookie, name string) string {
	for _, c := range cookies {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// doRefresh mirrors production wiring: RequireCSRF -> Refresh.
func doRefresh(cfg config.Config, store users.Store, refreshRaw, csrfRaw string) *httptest.ResponseRecorder {
	h := middleware.RequireCSRF(Refresh(cfg, store))
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	if refreshRaw != "" {
		req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: refreshRaw})
	}
	if csrfRaw != "" {
		req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrfRaw})
		req.Header.Set(csrfHeaderName, csrfRaw)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRefreshRotatesSingleUse(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()
	_, cookies := registerSession(t, cfg, store, uniqueTestEmail("rotate"))

	refreshRaw := cookieValue(cookies, refreshCookieName)
	csrfRaw := cookieValue(cookies, csrfCookieName)
	if refreshRaw == "" || csrfRaw == "" {
		t.Fatal("expected refresh + csrf cookies")
	}

	first := doRefresh(cfg, store, refreshRaw, csrfRaw)
	if first.Code != http.StatusOK {
		t.Fatalf("first refresh = %d (%s)", first.Code, first.Body.String())
	}
	var rotated SessionResponse
	if err := json.NewDecoder(first.Body).Decode(&rotated); err != nil {
		t.Fatalf("decode rotated: %v", err)
	}
	if rotated.AccessToken == "" {
		t.Fatal("rotated session missing access token")
	}
	newCookies := first.Result().Cookies()
	if cookieValue(newCookies, refreshCookieName) == refreshRaw {
		t.Error("refresh token must rotate — got the same raw value back")
	}

	// Old pair is single-use: replay fails closed.
	replay := doRefresh(cfg, store, refreshRaw, csrfRaw)
	if replay.Code != http.StatusUnauthorized {
		t.Errorf("replay of old refresh = %d, want 401", replay.Code)
	}

	// New pair works for an authenticated read.
	me := middleware.Authenticate(cfg.JWTSecret, store, Me(store))
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+rotated.AccessToken)
	rec := httptest.NewRecorder()
	me.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("me with rotated access = %d, want 200", rec.Code)
	}
}

func TestRefreshWithExpiredAccessStillWorks(t *testing.T) {
	cfg := testConfig()
	cfg.AccessTTL = 20 * time.Millisecond
	store := users.NewMemoryStore()
	sess, cookies := registerSession(t, cfg, store, uniqueTestEmail("short-access"))

	time.Sleep(60 * time.Millisecond)

	// Old access token is dead...
	me := middleware.Authenticate(cfg.JWTSecret, store, Me(store))
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+sess.AccessToken)
	rec := httptest.NewRecorder()
	me.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expired access me = %d, want 401", rec.Code)
	}

	// ...but the refresh cookie + CSRF pair still rotates a fresh session.
	res := doRefresh(cfg, store,
		cookieValue(cookies, refreshCookieName),
		cookieValue(cookies, csrfCookieName))
	if res.Code != http.StatusOK {
		t.Fatalf("refresh after access expiry = %d (%s)", res.Code, res.Body.String())
	}
}

func TestRefreshExpiredSessionFailsAndClearsCookies(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()
	sess, cookies := registerSession(t, cfg, store, uniqueTestEmail("expired-sess"))

	refreshRaw := cookieValue(cookies, refreshCookieName)
	csrfRaw := cookieValue(cookies, csrfCookieName)
	refreshHash := tokens.HashToken(refreshRaw)

	// Swap the live session for an expired one with identical hashes.
	ctx := t.Context()
	old, err := store.GetRefreshSessionByHash(ctx, refreshHash)
	if err != nil {
		t.Fatalf("lookup session: %v", err)
	}
	_ = sess
	_ = store.DeleteRefreshSessionByHash(ctx, refreshHash)
	expiredID, _ := uuid.NewV7()
	if err := store.CreateRefreshSession(ctx, users.RefreshSession{
		ID:        expiredID,
		UserID:    old.UserID,
		TokenHash: refreshHash,
		CSRFHash:  old.CSRFHash,
		ExpiresAt: time.Now().Add(-time.Hour),
	}); err != nil {
		t.Fatalf("seed expired session: %v", err)
	}

	res := doRefresh(cfg, store, refreshRaw, csrfRaw)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expired refresh = %d, want 401", res.Code)
	}
	// Cookies must be cleared so the client drops the dead session.
	cleared := map[string]bool{}
	for _, c := range res.Result().Cookies() {
		if c.Value == "" && c.MaxAge < 0 {
			cleared[c.Name] = true
		}
	}
	if !cleared[refreshCookieName] || !cleared[csrfCookieName] {
		t.Errorf("expired session must clear both cookies, got %v", res.Result().Cookies())
	}
}

func TestRefreshCrossUserCSRFKillsSession(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()
	_, cookiesA := registerSession(t, cfg, store, uniqueTestEmail("csrf-a"))
	_, cookiesB := registerSession(t, cfg, store, uniqueTestEmail("csrf-b"))

	refreshA := cookieValue(cookiesA, refreshCookieName)
	csrfB := cookieValue(cookiesB, csrfCookieName)

	// A's refresh cookie with B's CSRF token: session-bound check fails,
	// server kills A's session so it can't be retried.
	res := doRefresh(cfg, store, refreshA, csrfB)
	// RequireCSRF fires first on header/cookie mismatch when routed through
	// the mux; through the handler chain here the session binding rejects.
	// Either way it must fail closed with 4xx, never 200.
	if res.Code == http.StatusOK {
		t.Fatalf("cross-user CSRF refresh = 200, want 4xx")
	}

	// A's original pair must now be dead (session killed on mismatch).
	csrfA := cookieValue(cookiesA, csrfCookieName)
	retry := doRefresh(cfg, store, refreshA, csrfA)
	if retry.Code == http.StatusOK {
		t.Error("A's session should be dead after CSRF mismatch")
	}
}

func TestRefreshCSRFFailures(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()
	_, cookies := registerSession(t, cfg, store, uniqueTestEmail("csrf-matrix"))
	refreshRaw := cookieValue(cookies, refreshCookieName)
	csrfRaw := cookieValue(cookies, csrfCookieName)

	h := middleware.RequireCSRF(Refresh(cfg, store))
	call := func(refresh, header, cookie string) int {
		req := httptest.NewRequest("POST", "/auth/refresh", nil)
		if refresh != "" {
			req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: refresh})
		}
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: cookie})
		}
		if header != "" {
			req.Header.Set(csrfHeaderName, header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := call("", "", ""); code != http.StatusForbidden {
		t.Errorf("bare refresh = %d, want 403", code)
	}
	if code := call(refreshRaw, "", ""); code != http.StatusForbidden {
		t.Errorf("missing CSRF = %d, want 403", code)
	}
	if code := call(refreshRaw, csrfRaw, "different-cookie"); code != http.StatusForbidden {
		t.Errorf("header/cookie mismatch = %d, want 403", code)
	}
}

func TestLogoutKillsSessionAndClearsCookies(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()
	_, cookies := registerSession(t, cfg, store, uniqueTestEmail("logout"))
	refreshRaw := cookieValue(cookies, refreshCookieName)
	csrfRaw := cookieValue(cookies, csrfCookieName)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: refreshRaw})
	rec := httptest.NewRecorder()
	Logout(cfg, store)(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout = %d, want 204", rec.Code)
	}
	cleared := 0
	for _, c := range rec.Result().Cookies() {
		if c.Value == "" && c.MaxAge < 0 {
			cleared++
		}
	}
	if cleared < 2 {
		t.Errorf("logout must clear both cookies, got %v", rec.Result().Cookies())
	}

	// Refresh after logout fails — the session hash is gone.
	if res := doRefresh(cfg, store, refreshRaw, csrfRaw); res.Code != http.StatusUnauthorized {
		t.Errorf("refresh after logout = %d, want 401", res.Code)
	}
}
