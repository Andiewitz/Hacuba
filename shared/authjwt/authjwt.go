// Package authjwt defines Hacuba access-token claims and HS256 verification.
// It is intentionally small so services can share authentication semantics
// without sharing databases or importing another service's internal packages.
package authjwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	issuer   = "hacuba-auth"
	audience = "hacuba-client"
)

// Claims is the verified access-token payload. Role is trusted only after
// VerifyAccess has checked its signature, issuer, audience, and expiry.
type Claims struct {
	jwt.RegisteredClaims
	CSRFHash string `json:"csrf"`
	Role     string `json:"role"`
}

// IssueAccess mints a short-lived HS256 access token.
func IssueAccess(secret []byte, userID uuid.UUID, role, csrfHash string, ttl time.Duration) (string, error) {
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
		Role:     role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// VerifyAccess validates a raw access token and returns its claims.
func VerifyAccess(secret []byte, raw string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	}, jwt.WithIssuer(issuer), jwt.WithAudience(audience), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return Claims{}, fmt.Errorf("invalid sub claim: %w", err)
	}
	if claims.Role != "buyer" && claims.Role != "seller" {
		return Claims{}, fmt.Errorf("invalid role claim")
	}
	return claims, nil
}

// UserID parses the JWT subject as a UUID.
func (c Claims) UserID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}
