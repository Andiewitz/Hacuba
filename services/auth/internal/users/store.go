package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when no row matches. Handlers map it to 404
// (never 403 for cross-user access) so callers can't probe existence.
var ErrNotFound = errors.New("not found")

// ErrEmailTaken is returned on duplicate registration. Handlers map it to 409.
var ErrEmailTaken = errors.New("email already registered")

// RefreshSession is a server-side record for one login session.
// Only a SHA-256 hash of the opaque refresh token is stored —
// a DB leak alone cannot mint new access tokens.
type RefreshSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	CSRFHash  string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Store is the persistence boundary. Postgres is the only implementation
// and only the auth service talks to it (see docker-compose.yml —
// auth-db lives on an internal network with no published ports).
//
// Ownership rule: every method takes the caller's UUID from the JWT
// (never from client input). Future resource stores must follow the same
// pattern: WHERE owner_id = $1 with the authenticated UUID.
type Store interface {
	CreateUser(ctx context.Context, user User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	// GetUserByID returns only the user matching id — the caller passes
	// the JWT sub, so a user can only ever load their own row.
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)

	CreateRefreshSession(ctx context.Context, s RefreshSession) error
	GetRefreshSessionByHash(ctx context.Context, tokenHash string) (*RefreshSession, error)
	DeleteRefreshSessionByHash(ctx context.Context, tokenHash string) error
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
}
