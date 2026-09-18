package handlers

import (
	"errors"
	"net/http"

	"github.com/hacuba/auth/internal/middleware"
	"github.com/hacuba/auth/internal/users"
)

// Me handles GET /auth/me.
//
// Isolation by construction: the user ID comes from the JWT-verified
// request context (see middleware.Authenticate), never from URL params,
// query strings, or JSON bodies — so a caller can only ever read their
// own row. Unknown IDs map to 404, never 403, to avoid existence leaks.
func Me(store users.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		user, err := store.GetUserByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, users.ErrNotFound) {
				writeError(w, http.StatusNotFound, "not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to load user")
			return
		}
		writeJSON(w, http.StatusOK, publicUser(user))
	}
}
