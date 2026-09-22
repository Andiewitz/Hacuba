package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hacuba/authjwt"
)

type contextKey string

const (
	userIDKey contextKey = "listing_user_id"
	roleKey   contextKey = "listing_role"
	csrfKey   contextKey = "listing_csrf_hash"
)

func Authenticate(secret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
			unauthorized(w)
			return
		}
		claims, err := authjwt.VerifyAccess(secret, parts[1])
		if err != nil {
			unauthorized(w)
			return
		}
		id, err := claims.UserID()
		if err != nil {
			unauthorized(w)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, id)
		ctx = context.WithValue(ctx, roleKey, claims.Role)
		ctx = context.WithValue(ctx, csrfKey, claims.CSRFHash)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireSeller(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if role, _ := r.Context().Value(roleKey).(string); role != "seller" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "seller role required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("X-CSRF-Token")
		cookie, err := r.Cookie("csrf_token")
		bound, _ := r.Context().Value(csrfKey).(string)
		sum := sha256.Sum256([]byte(header))
		actualHash := hex.EncodeToString(sum[:])
		if header == "" || err != nil || cookie.Value == "" || bound == "" ||
			subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) != 1 ||
			subtle.ConstantTimeCompare([]byte(actualHash), []byte(bound)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid CSRF token"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
