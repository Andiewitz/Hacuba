package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/tokens"
	"github.com/hacuba/auth/internal/users"
)

// Refresh handles POST /auth/refresh.
//
// The client sends the HttpOnly refresh cookie + X-CSRF-Token header.
// Server: validate CSRF double-submit -> lookup session by token hash ->
// check expiry -> ROTATE (delete old session, create new one) -> issue new
// access JWT + refresh/CSRF pair. Rotation means a stolen refresh token is
// only usable once — reuse is detectable and revokes the session chain.
func Refresh(cfg config.Config, store users.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshCookie, err := r.Cookie(refreshCookieName)
		if err != nil || refreshCookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "missing refresh token")
			return
		}
		csrfHeader := r.Header.Get(csrfHeaderName)
		csrfCookie, csrfErr := r.Cookie(csrfCookieName)
		if csrfHeader == "" || csrfErr != nil || csrfCookie.Value == "" {
			writeError(w, http.StatusForbidden, "missing CSRF token")
			return
		}
		// Double-submit check: header must match cookie value.
		if csrfHeader != csrfCookie.Value {
			writeError(w, http.StatusForbidden, "invalid CSRF token")
			return
		}

		refreshHash := tokens.HashToken(refreshCookie.Value)
		sess, err := store.GetRefreshSessionByHash(r.Context(), refreshHash)
		if err != nil {
			if errors.Is(err, users.ErrNotFound) {
				writeError(w, http.StatusUnauthorized, "invalid refresh token")
				return
			}
			writeError(w, http.StatusInternalServerError, "refresh failed")
			return
		}

		// Session-bound CSRF check: presented token must match the session.
		if !tokens.EqualHash(csrfHeader, sess.CSRFHash) {
			// Possible theft — kill the session so it can't be retried.
			_ = store.DeleteRefreshSessionByHash(r.Context(), refreshHash)
			clearSessionCookies(w, cfg)
			writeError(w, http.StatusForbidden, "invalid CSRF token")
			return
		}

		if time.Now().After(sess.ExpiresAt) {
			_ = store.DeleteRefreshSessionByHash(r.Context(), refreshHash)
			clearSessionCookies(w, cfg)
			writeError(w, http.StatusUnauthorized, "session expired")
			return
		}

		user, err := store.GetUserByID(r.Context(), sess.UserID)
		if err != nil {
			if errors.Is(err, users.ErrNotFound) {
				writeError(w, http.StatusUnauthorized, "user not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "refresh failed")
			return
		}

		// Rotate: single-use refresh tokens. Delete before issuing so a
		// concurrent replay of the old token fails closed.
		_ = store.DeleteRefreshSessionByHash(r.Context(), refreshHash)

		rotated, err := issueSession(r.Context(), w, cfg, store, user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to rotate session")
			return
		}
		writeJSON(w, http.StatusOK, rotated)
	}
}
