package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/hacuba/auth/internal/tokens"
)

// RequireCSRF enforces double-submit CSRF protection on state-changing
// requests: the X-CSRF-Token header must match the csrf_token cookie.
// When the request is also authenticated, the presented token must hash to
// the CSRF binding inside the access JWT — so a token stolen for one
// session can't be replayed in another.
//
// Safe methods (GET/HEAD/OPTIONS) skip validation.
func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		header := r.Header.Get("X-CSRF-Token")
		cookie, err := r.Cookie("csrf_token")
		if header == "" || err != nil || cookie.Value == "" ||
			subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid CSRF token"})
			return
		}

		// If authenticated, bind to the JWT claim as well.
		if bound, ok := CSRFHashFromContext(r.Context()); ok && bound != "" {
			if !tokens.EqualHash(header, bound) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid CSRF token"})
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
