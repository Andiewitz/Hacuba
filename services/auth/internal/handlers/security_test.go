package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hacuba/auth/internal/users"
)

// postRaw sends a raw body so malformed-JSON and wrong-type payloads can
// be exercised exactly as an attacker would send them.
func postRaw(t *testing.T, handler http.HandlerFunc, target, raw string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", target, strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func TestEightCharPasswordRoundTrip(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()
	email := uniqueTestEmail("eight")

	reg := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": email, "password": "abcd1234"})
	if reg.Code != http.StatusCreated {
		t.Fatalf("register 8-char = %d (%s)", reg.Code, reg.Body.String())
	}
	ok := doJSON(t, Login(cfg, store), "POST", "/auth/login",
		map[string]string{"email": email, "password": "abcd1234"})
	if ok.Code != http.StatusOK {
		t.Fatalf("login 8-char = %d (%s)", ok.Code, ok.Body.String())
	}
}

func TestOverlongPasswordIs400Not500(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()
	// 100 chars would blow past bcrypt's 72-byte limit — validation must
	// reject it with a clean 400, never a hashing 500.
	rec := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": uniqueTestEmail("longpw"), "password": strings.Repeat("a1", 50)})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("100-char password = %d, want 400", rec.Code)
	}
}

func TestSQLPayloadsAreInert(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()

	// Quote in the local part is a valid address — stored as data, and the
	// parameterized lookup finds it again on login.
	quoted := "o'brien-" + strings.ReplaceAll(uniqueTestEmail("q"), "@", "+q@")
	reg := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": quoted, "password": "validpassword123"})
	if reg.Code != http.StatusCreated {
		t.Fatalf("quoted email register = %d (%s)", reg.Code, reg.Body.String())
	}
	ok := doJSON(t, Login(cfg, store), "POST", "/auth/login",
		map[string]string{"email": quoted, "password": "validpassword123"})
	if ok.Code != http.StatusOK {
		t.Fatalf("quoted email login = %d (%s)", ok.Code, ok.Body.String())
	}

	// Spaceless injection string as password: hashes and round-trips as an
	// opaque secret. A wrong password still fails, and the store is intact.
	evil := "abc123';DROPTABLE(users);--"
	evilEmail := uniqueTestEmail("evil")
	if rec := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": evilEmail, "password": evil}); rec.Code != http.StatusCreated {
		t.Fatalf("evil password register = %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, Login(cfg, store), "POST", "/auth/login",
		map[string]string{"email": evilEmail, "password": evil}); rec.Code != http.StatusOK {
		t.Fatalf("evil password login = %d (%s)", rec.Code, rec.Body.String())
	}
	if rec := doJSON(t, Login(cfg, store), "POST", "/auth/login",
		map[string]string{"email": evilEmail, "password": "abc123"}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("truncated evil password = %d, want 401", rec.Code)
	}
	// Store demonstrably intact: a fresh user registers fine afterwards.
	if rec := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": uniqueTestEmail("after"), "password": "validpassword123"}); rec.Code != http.StatusCreated {
		t.Fatalf("post-attack register = %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestLoginEnumerationParity(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()
	email := uniqueTestEmail("enum")
	if rec := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": email, "password": "validpassword123"}); rec.Code != http.StatusCreated {
		t.Fatalf("setup register = %d", rec.Code)
	}

	unknown := doJSON(t, Login(cfg, store), "POST", "/auth/login",
		map[string]string{"email": uniqueTestEmail("ghost"), "password": "validpassword123"})
	wrongPw := doJSON(t, Login(cfg, store), "POST", "/auth/login",
		map[string]string{"email": email, "password": "validpassword999"})

	if unknown.Code != http.StatusUnauthorized || wrongPw.Code != http.StatusUnauthorized {
		t.Fatalf("both must be 401, got %d and %d", unknown.Code, wrongPw.Code)
	}
	if unknown.Body.String() != wrongPw.Body.String() {
		t.Errorf("responses differ — account enumeration possible:\nunknown: %s\nwrongPw: %s",
			unknown.Body.String(), wrongPw.Body.String())
	}
}

func TestMalformedBodiesAre400Never500(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()

	bodies := map[string]string{
		"empty":         "",
		"not json":      "hello",
		"array":         `["a","b"]`,
		"number":        `42`,
		"wrong types":   `{"email":123,"password":true}`,
		"null":          `null`,
		"missing all":   `{}`,
		"huge email":    `{"email":"` + strings.Repeat("a", 500) + `@example.com","password":"validpassword123"}`,
		"huge pw":       `{"email":"` + uniqueTestEmail("huge") + `","password":"` + strings.Repeat("x", 5000) + `1"}`,
		"trailing junk": `{"email":"a@example.com"} garbage`,
	}
	for name, body := range bodies {
		t.Run("register/"+name, func(t *testing.T) {
			rec := postRaw(t, Register(cfg, store), "/auth/register", body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("got %d, want 400 (body=%s)", rec.Code, truncate(rec.Body.String(), 120))
			}
		})
		t.Run("login/"+name, func(t *testing.T) {
			rec := postRaw(t, Login(cfg, store), "/auth/login", body)
			if rec.Code < 400 || rec.Code >= 500 {
				t.Errorf("got %d, want 4xx", rec.Code)
			}
		})
	}
}

func TestErrorBodiesNeverEchoInput(t *testing.T) {
	cfg := testConfig()
	store := users.NewMemoryStore()

	xss := `<script>alert(1)</script>@example.com`
	rec := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": xss, "password": "validpassword123"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("xss email = %d, want 400", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<script>") {
		t.Errorf("error echoes input — reflected XSS risk: %s", rec.Body.String())
	}

	longInput := strings.Repeat("B", 600) + "@example.com"
	rec2 := doJSON(t, Register(cfg, store), "POST", "/auth/register",
		map[string]string{"email": longInput, "password": "validpassword123"})
	if strings.Contains(rec2.Body.String(), strings.Repeat("B", 100)) {
		t.Error("error echoes oversized input back to the client")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
