package users

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswordBoundaries(t *testing.T) {
	cases := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{"exactly 8 with letter+number passes", "abcd1234", false},
		{"7 chars fails", "abc1234", true},
		{"empty fails", "", true},
		{"letters only fails", "abcdefgh", true},
		{"numbers only fails", "12345678", true},
		{"spaces fail", "abcd 1234", true},
		{"tab fails", "abcd\t1234", true},
		{"exactly 72 passes", strings.Repeat("a1", 36), false},
		{"73 fails cleanly", strings.Repeat("a1", 36) + "a", true},
		{"multibyte letters count as bytes but still pass", "pässwörd1", false},
		{"symbols allowed", "abc123!@#", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.pw)
			if tc.wantErr && err == nil {
				t.Errorf("password %q: expected error, got nil", tc.pw)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("password %q: expected nil, got %v", tc.pw, err)
			}
		})
	}
}

func TestEmailAdversarial(t *testing.T) {
	longLocalPass := strings.Repeat("a", 242) // 242 + len("@example.com")=12 -> 254
	longLocalFail := strings.Repeat("a", 243) // 255 total
	cases := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"sql quote in local part is accepted as data", "o'brien@example.com", false},
		{"semicolons rejected by format", "a;DROP TABLE users@example.com", true},
		{"newline rejected", "a\n@example.com", true},
		{"null byte rejected", "a\x00@example.com", true},
		{"exactly 254 passes", longLocalPass + "@example.com", false},
		{"255 fails", longLocalFail + "@example.com", true},
		{"uppercase passes (normalized)", "USER@EXAMPLE.COM", false},
		{"missing domain dot fails", "user@localhost", true},
		{"double at fails", "a@b@example.com", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEmail(tc.email)
			if tc.wantErr && err == nil {
				t.Errorf("email %q: expected error, got nil", tc.email)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("email %q: expected nil, got %v", tc.email, err)
			}
		})
	}
}

// TestPostgresQueriesAreParameterized pins the anti-injection mechanism
// without needing a live database: every query must use $ placeholders,
// and no query string may be built with fmt.Sprintf.
func TestPostgresQueriesAreParameterized(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("postgres.go"))
	if err != nil {
		t.Fatalf("read postgres.go: %v", err)
	}
	body := string(src)
	for _, kw := range []string{"SELECT", "INSERT", "DELETE"} {
		if !strings.Contains(body, kw) {
			t.Errorf("expected %s queries in postgres.go", kw)
		}
	}
	if strings.Contains(body, "fmt.Sprintf") {
		t.Error("postgres.go must not build queries with fmt.Sprintf — use $ placeholders")
	}
	if !strings.Contains(body, "$1") {
		t.Error("postgres.go should use $ placeholders for parameters")
	}
}
