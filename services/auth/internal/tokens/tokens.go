package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	issuer   = "hacuba-auth"
	audience = "hacuba-client"
)

// Claims is the access-token payload. sub is the user's UUID — the only
// identity the rest of the system trusts. csrf binds the token to one
// CSRF secret (stored as a hash, never the raw value).
type Claims struct {
	jwt.RegisteredClaims
	CSRFHash string `json:"csrf"`
}

// IssueAccess mints a short-lived JWT for userID, bound to csrfHash.
func IssueAccess(secret []byte, userID uuid.UUID, csrfHash string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  jwt.ClaimStrings{audience},
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		CSRFHash: csrfHash,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// VerifyAccess checks signature, expiry, issuer/audience and returns the
// caller UUID + bound CSRF hash. Any error means unauthenticated.
func VerifyAccess(secret []byte, raw string) (uuid.UUID, string, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	},
		jwt.WithIssuer(issuer),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return uuid.Nil, "", fmt.Errorf("invalid token")
	}
	userID, err := uuid.Parse(claims.Subject)
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
