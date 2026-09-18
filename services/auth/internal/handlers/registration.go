package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/users"
	"golang.org/x/crypto/bcrypt"
)

// Register handles POST /auth/register.
//
// Flow: validate input -> bcrypt hash -> UUIDv7 -> INSERT -> auto-login
// (issue access JWT + refresh session + CSRF pair, same as login).
func Register(cfg config.Config, store users.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		email := users.NormalizeEmail(req.Email)
		if err := users.ValidateEmail(email); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := users.ValidatePassword(req.Password); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to process password")
			return
		}

		// UUIDv7: time-ordered, unguessable, DB-friendly. Generated here in
		// Go so Postgres never mints identities — the UUID is also the JWT
		// sub and the owner_id foreign key everywhere else.
		id, err := uuid.NewV7()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate user id")
			return
		}

		user, err := store.CreateUser(r.Context(), users.User{
			ID:           id,
			Email:        email,
			PasswordHash: string(hash),
		})
		if err != nil {
			if errors.Is(err, users.ErrEmailTaken) {
				writeError(w, http.StatusConflict, "email already registered")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return
		}

		sess, err := issueSession(r.Context(), w, cfg, store, user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create session")
			return
		}
		writeJSON(w, http.StatusCreated, sess)
	}
}
