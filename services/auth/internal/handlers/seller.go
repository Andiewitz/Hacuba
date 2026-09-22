package handlers

import (
	"net/http"

	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/middleware"
	"github.com/hacuba/auth/internal/users"
)

// BecomeSeller upgrades the authenticated account and immediately issues a
// session whose access token carries role=seller. Repeating the request is
// safe and returns the same role.
func BecomeSeller(cfg config.Config, store users.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		user, err := store.BecomeSeller(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		sess, err := issueSession(r.Context(), w, cfg, store, user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create session")
			return
		}
		writeJSON(w, http.StatusOK, sess)
	}
}
