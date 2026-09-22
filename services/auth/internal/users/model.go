package users

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	RoleBuyer  = "buyer"
	RoleSeller = "seller"
)

// User is the canonical identity row. ID is a UUIDv7 generated in Go
// (see registration.go) — never sequential, never client-supplied.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never serialized
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// NormalizeRole gives newly-created in-memory users the same default as the
// database migration. Callers cannot create arbitrary roles.
func NormalizeRole(role string) string {
	if role == RoleSeller {
		return RoleSeller
	}
	return RoleBuyer
}

// NormalizeEmail lowercases and trims so uniqueness is case-insensitive.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
