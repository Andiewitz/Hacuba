package handlers

import (
	"net/http"

	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/tokens"
	"github.com/hacuba/auth/internal/users"
)

// Logout handles POST /auth/logout. Deletes the refresh session (by token
// hash — the raw token never hits the DB) and clears both cookies.
// The short-lived access JWT simply expires; no blocklist needed.
func Logout(cfg config.Config, store users.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(refreshCookieName); err == nil && c.Value != "" {
			_ = store.DeleteRefreshSessionByHash(r.Context(), tokens.HashToken(c.Value))
		}
		clearSessionCookies(w, cfg)
		w.WriteHeader(http.StatusNoContent)
	}
}
