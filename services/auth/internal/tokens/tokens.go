package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/authjwt"
)

// Claims is shared with services that verify Hacuba access tokens.
type Claims = authjwt.Claims

// IssueAccess mints a short-lived access token containing the caller's role.
func IssueAccess(secret []byte, userID uuid.UUID, role, csrfHash string, ttl time.Duration) (string, error) {
	return authjwt.IssueAccess(secret, userID, role, csrfHash, ttl)
}

// VerifyAccess checks signature, expiry, issuer/audience and returns the
// caller UUID + bound CSRF hash. Any error means unauthenticated.
func VerifyAccess(secret []byte, raw string) (uuid.UUID, string, error) {
	claims, err := authjwt.VerifyAccess(secret, raw)
	if err != nil {
		return uuid.Nil, "", err
	}
	userID, err := claims.UserID()
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid sub claim: %w", err)
	}
	return userID, claims.CSRFHash, nil
}

// NewOpaqueToken generates a 32-byte random token for refresh/CSRF use.
// It returns (raw, hash): send raw to the client, store only hash.
func NewOpaqueToken() (raw, hash string, err error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b[:])
	return raw, HashToken(raw), nil
}

// HashToken SHA-256 hex-hashes an opaque token for storage/comparison.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

// EqualHash compares a raw token against a stored hash in constant time.
func EqualHash(raw, hash string) bool {
	return subtle.ConstantTimeCompare([]byte(HashToken(raw)), []byte(hash)) == 1
}
