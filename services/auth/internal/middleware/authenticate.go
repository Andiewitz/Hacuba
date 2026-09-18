package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hacuba/auth/internal/tokens"
	"github.com/hacuba/auth/internal/users"
)

// Authenticate verifies the Bearer access JWT and loads the caller.
// On success the verified UUID (+ user row + CSRF binding) is stored in
// the request context. On any failure it returns 401 — never partial data.
func Authenticate(secret []byte, store users.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, csrfHash, ok := verifyRequest(r, secret, store)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r.WithContext(withAuthContext(r.Context(), user, csrfHash)))
	})
}

// verifyRequest parses `Authorization: Bearer <jwt>`, validates it, and
// reloads the user row so deleted users lose access immediately.
func verifyRequest(r *http.Request, secret []byte, store users.Store) (*users.User, string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
		return nil, "", false
	}
	userID, csrfHash, err := tokens.VerifyAccess(secret, parts[1])
	if err != nil {
		return nil, "", false
	}
	user, err := store.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil, "", false
		}
		return nil, "", false
	}
	return user, csrfHash, true
}
