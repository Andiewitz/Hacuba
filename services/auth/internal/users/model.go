package users

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// User is the canonical identity row. ID is a UUIDv7 generated in Go
// (see registration.go) — never sequential, never client-supplied.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never serialized
	CreatedAt    time.Time `json:"created_at"`
}

// NormalizeEmail lowercases and trims so uniqueness is case-insensitive.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
