package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/tokens"
	"github.com/hacuba/auth/internal/users"
)

// SessionResponse is what the client receives after register/login/refresh.
// The access token lives in memory on the frontend (never localStorage).
// Refresh + CSRF also arrive as cookies (see setSessionCookies).
type SessionResponse struct {
	AccessToken string    `json:"access_token"`
	CSRFToken   string    `json:"csrf_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	User        UserPublic `json:"user"`
}

// UserPublic is the safe subset of users.User sent to clients.
type UserPublic struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func publicUser(u *users.User) UserPublic {
	return UserPublic{ID: u.ID, Email: u.Email, CreatedAt: u.CreatedAt}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// issueSession creates a refresh session + CSRF pair, persists only hashes,
// sets cookies, and returns the client-facing session payload.
func issueSession(ctx context.Context, w http.ResponseWriter, cfg config.Config, store users.Store, user *users.User) (SessionResponse, error) {
	refreshRaw, refreshHash, err := tokens.NewOpaqueToken()
	if err != nil {
		return SessionResponse{}, err
	}
	csrfRaw, csrfHash, err := tokens.NewOpaqueToken()
	if err != nil {
		return SessionResponse{}, err
	}

	sessID, err := uuid.NewV7()
	if err != nil {
		return SessionResponse{}, err
	}
	sess := users.RefreshSession{
		ID:        sessID,
		UserID:    user.ID,
		TokenHash: refreshHash,
		CSRFHash:  csrfHash,
		ExpiresAt: time.Now().Add(cfg.RefreshTTL),
	}
	// Pass request-scoped ctx in callers; background here keeps helper simple
	// for register/login which already validated input.
	if err := store.CreateRefreshSession(ctx, sess); err != nil {
		return SessionResponse{}, err
	}

	access, err := tokens.IssueAccess(cfg.JWTSecret, user.ID, csrfHash, cfg.AccessTTL)
	if err != nil {
		return SessionResponse{}, err
	}

	setSessionCookies(w, cfg, refreshRaw, csrfRaw, cfg.RefreshTTL)

	return SessionResponse{
		AccessToken: access,
		CSRFToken:   csrfRaw,
		TokenType:   "Bearer",
		ExpiresIn:   int(cfg.AccessTTL.Seconds()),
		User:        publicUser(user),
	}, nil
}
