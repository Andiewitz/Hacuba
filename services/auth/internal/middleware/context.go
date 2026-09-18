package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/hacuba/auth/internal/users"
)

type ctxKey string

const (
	ctxUserIDKey   ctxKey = "auth_user_id"
	ctxUserKey     ctxKey = "auth_user"
	ctxCSRFHashKey ctxKey = "auth_csrf_hash"
)

// withAuthContext stores the verified identity for downstream handlers.
// Only Authenticate writes these values — handlers only read them, so a
// user can never smuggle another user's UUID through params or JSON.
func withAuthContext(ctx context.Context, user *users.User, csrfHash string) context.Context {
	ctx = context.WithValue(ctx, ctxUserIDKey, user.ID)
	ctx = context.WithValue(ctx, ctxUserKey, user)
	ctx = context.WithValue(ctx, ctxCSRFHashKey, csrfHash)
	return ctx
}

// UserIDFromContext returns the JWT-verified caller UUID.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxUserIDKey).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

// UserFromContext returns the full verified user row.
func UserFromContext(ctx context.Context) (*users.User, bool) {
	u, ok := ctx.Value(ctxUserKey).(*users.User)
	return u, ok && u != nil
}

// CSRFHashFromContext returns the CSRF hash bound to the access token.
func CSRFHashFromContext(ctx context.Context) (string, bool) {
	h, ok := ctx.Value(ctxCSRFHashKey).(string)
	return h, ok && h != ""
}
