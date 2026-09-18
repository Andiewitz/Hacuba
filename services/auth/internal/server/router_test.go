package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/users"
)

func testConfig() config.Config {
	return config.Config{
		JWTSecret:    []byte("router-test-secret-32-bytes-minimum!"),
		AccessTTL:    15 * time.Minute,
		RefreshTTL:   7 * 24 * time.Hour,
		CookieSecure: false,
	}
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@example.com", prefix, uuid.Must(uuid.NewV7()).String())
}

func postJSON(mux http.Handler, method, target string, body any, headers map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, target, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

type sessionPayload struct {
	AccessToken string `json:"access_token"`
	CSRFToken   string `json:"csrf_token"`
	TokenType   string `json:"token_type"`
	User        struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

func cookiesByName(rec *httptest.ResponseRecorder) map[string]*http.Cookie {
	out := map[string]*http.Cookie{}
	for _, c := range rec.Result().Cookies() {
		out[c.Name] = c
	}
	return out
}

func TestRouterHealthz(t *testing.T) {
	mux := NewMux(testConfig(), users.NewMemoryStore())
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("healthz content-type = %q", ct)
	}
}

func TestRouterRegisterContract(t *testing.T) {
	cfg := testConfig()
	mux := NewMux(cfg, users.NewMemoryStore())
	email := uniqueEmail("register")

	rec := postJSON(mux, "POST", "/auth/register",
		map[string]string{"email": email, "password": "validpassword123"}, nil, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register = %d, body = %s", rec.Code, rec.Body.String())
	}

	var sess sessionPayload
	if err := json.NewDecoder(rec.Body).Decode(&sess); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if sess.AccessToken == "" || sess.CSRFToken == "" {
		t.Fatal("expected access + CSRF tokens in body")
	}
	if sess.TokenType != "Bearer" {
		t.Errorf("token_type = %q, want Bearer", sess.TokenType)
	}
	if sess.User.Email != email {
		t.Errorf("user email = %q, want %q", sess.User.Email, email)
	}

	// Cookie contract: refresh is HttpOnly, CSRF is readable, both Lax + Path=/.
	cookies := cookiesByName(rec)
	refresh, ok := cookies["refresh_token"]
	if !ok || refresh.Value == "" {
		t.Fatal("missing refresh_token cookie")
	}
	if !refresh.HttpOnly {
		t.Error("refresh_token must be HttpOnly")
	}
	csrf, ok := cookies["csrf_token"]
	if !ok || csrf.Value == "" {
		t.Fatal("missing csrf_token cookie")
	}
	if csrf.HttpOnly {
		t.Error("csrf_token must be readable by JS (HttpOnly=false)")
	}
	for name, c := range cookies {
		if c.Path != "/" {
			t.Errorf("%s Path = %q, want /", name, c.Path)
		}
		if c.SameSite != http.SameSiteLaxMode {
			t.Errorf("%s SameSite = %v, want Lax", name, c.SameSite)
		}
	}
	if csrf.Value != sess.CSRFToken {
		t.Error("csrf cookie value must match csrf_token in body")
	}

	// Duplicate -> 409.
	dup := postJSON(mux, "POST", "/auth/register",
		map[string]string{"email": email, "password": "validpassword123"}, nil, nil)
	if dup.Code != http.StatusConflict {
		t.Errorf("duplicate register = %d, want 409", dup.Code)
	}

	// Bad inputs -> 400.
	for name, body := range map[string]any{
		"invalid json":     "not-an-object",
		"bad email":        map[string]string{"email": "nope", "password": "validpassword123"},
		"weak password":    map[string]string{"email": uniqueEmail("weak"), "password": "short1"},
		"missing password": map[string]string{"email": uniqueEmail("nopw")},
	} {
		rec := postJSON(mux, "POST", "/auth/register", body, nil, nil)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400 (body=%s)", name, rec.Code, rec.Body.String())
		}
	}
}

func TestRouterLoginContract(t *testing.T) {
	cfg := testConfig()
	mux := NewMux(cfg, users.NewMemoryStore())
	email := uniqueEmail("login")
	const password = "validpassword123"

	if rec := postJSON(mux, "POST", "/auth/register",
		map[string]string{"email": email, "password": password}, nil, nil); rec.Code != http.StatusCreated {
		t.Fatalf("setup register = %d", rec.Code)
	}

	ok := postJSON(mux, "POST", "/auth/login",
		map[string]string{"email": email, "password": password}, nil, nil)
	if ok.Code != http.StatusOK {
		t.Fatalf("login = %d, body = %s", ok.Code, ok.Body.String())
	}

	// Wrong password is a generic 401 — same as unknown email, no enumeration.
	bad := postJSON(mux, "POST", "/auth/login",
		map[string]string{"email": email, "password": "wrongpassword999"}, nil, nil)
	if bad.Code != http.StatusUnauthorized {
		t.Errorf("wrong password = %d, want 401", bad.Code)
	}
	unknown := postJSON(mux, "POST", "/auth/login",
		map[string]string{"email": uniqueEmail("ghost"), "password": "wrongpassword999"}, nil, nil)
	if unknown.Code != http.StatusUnauthorized {
		t.Errorf("unknown email = %d, want 401", unknown.Code)
	}
	if strings.TrimSpace(bad.Body.String()) == "" || !strings.Contains(unknown.Body.String(), "invalid credentials") {
		t.Error("failure bodies should carry the generic invalid-credentials error")
	}

	// Missing fields -> 400.
	empty := postJSON(mux, "POST", "/auth/login",
		map[string]string{"email": email}, nil, nil)
	if empty.Code != http.StatusBadRequest {
		t.Errorf("missing password = %d, want 400", empty.Code)
	}
}

func TestRouterLoginRateLimit(t *testing.T) {
	cfg := testConfig()
	mux := NewMux(cfg, users.NewMemoryStore())
	email := uniqueEmail("ratelimit")
	const password = "validpassword123"

	if rec := postJSON(mux, "POST", "/auth/register",
		map[string]string{"email": email, "password": password}, nil, nil); rec.Code != http.StatusCreated {
		t.Fatalf("setup register = %d", rec.Code)
	}

	for i := 0; i < 5; i++ {
		rec := postJSON(mux, "POST", "/auth/login",
			map[string]string{"email": email, "password": "wrongpassword999"}, nil, nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("failure %d = %d, want 401", i+1, rec.Code)
		}
	}
	// 6th attempt — even with the right password — is throttled.
	throttled := postJSON(mux, "POST", "/auth/login",
		map[string]string{"email": email, "password": password}, nil, nil)
	if throttled.Code != http.StatusTooManyRequests {
		t.Errorf("6th attempt = %d, want 429", throttled.Code)
	}
}

func TestRouterMeIsolation(t *testing.T) {
	cfg := testConfig()
	mux := NewMux(cfg, users.NewMemoryStore())

	register := func(email string) sessionPayload {
		rec := postJSON(mux, "POST", "/auth/register",
			map[string]string{"email": email, "password": "validpassword123"}, nil, nil)
		if rec.Code != http.StatusCreated {
			t.Fatalf("register %s = %d", email, rec.Code)
		}
		var s sessionPayload
		_ = json.NewDecoder(rec.Body).Decode(&s)
		return s
	}
	sessA := register(uniqueEmail("me-a"))
	sessB := register(uniqueEmail("me-b"))
	if sessA.User.ID == sessB.User.ID {
		t.Fatal("each user must have its own UUID")
	}

	getMe := func(token string) (int, sessionPayload) {
		req := httptest.NewRequest("GET", "/auth/me", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		var u struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}
		if rec.Code == http.StatusOK {
			_ = json.NewDecoder(rec.Body).Decode(&u)
		}
		return rec.Code, sessionPayload{User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}(u)}
	}

	if code, got := getMe(sessA.AccessToken); code != http.StatusOK || got.User.Email != sessA.User.Email {
		t.Errorf("A me: code=%d email=%q", code, got.User.Email)
	}
	if code, got := getMe(sessB.AccessToken); code != http.StatusOK || got.User.Email != sessB.User.Email {
		t.Errorf("B me: code=%d email=%q", code, got.User.Email)
	}
	// No parameter exists to request another user's row — tampered and
	// missing tokens fail closed with 401.
	if code, _ := getMe(""); code != http.StatusUnauthorized {
		t.Errorf("anonymous me = %d, want 401", code)
	}
	if code, _ := getMe("tampered.token.value"); code != http.StatusUnauthorized {
		t.Errorf("tampered me = %d, want 401", code)
	}
	// Cross-user replay: B's token never yields A's data.
	if code, got := getMe(sessB.AccessToken); code != http.StatusOK || got.User.ID == sessA.User.ID {
		t.Errorf("cross-user leak check failed: code=%d", code)
	}
}

func TestRouterRefreshRequiresCSRF(t *testing.T) {
	cfg := testConfig()
	mux := NewMux(cfg, users.NewMemoryStore())

	// No cookies at all -> CSRF layer rejects before the handler (403).
	rec := postJSON(mux, "POST", "/auth/refresh", nil, nil, nil)
	if rec.Code != http.StatusForbidden {
		t.Errorf("bare refresh = %d, want 403", rec.Code)
	}

	// Refresh cookie without CSRF pair -> still 403.
	email := uniqueEmail("csrf")
	reg := postJSON(mux, "POST", "/auth/register",
		map[string]string{"email": email, "password": "validpassword123"}, nil, nil)
	cookies := cookiesByName(reg)
	refreshOnly := []*http.Cookie{cookies["refresh_token"]}
	rec2 := postJSON(mux, "POST", "/auth/refresh", nil, nil, refreshOnly)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("refresh without CSRF = %d, want 403", rec2.Code)
	}
}
