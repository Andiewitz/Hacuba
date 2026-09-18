package users

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

const (
	minPasswordLen = 12
	maxEmailLen    = 254
)

// ValidateEmail rejects empty, oversized, or unparseable addresses.
// Requires a dotted domain so single-label hosts like "a@b" are rejected.
func ValidateEmail(email string) error {
	email = NormalizeEmail(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if len(email) > maxEmailLen {
		return fmt.Errorf("email too long")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidatePassword enforces a minimum length plus basic complexity.
// Keeps rules simple and explainable: length first, then character variety.
func ValidatePassword(password string) error {
	if len(password) < minPasswordLen {
		return fmt.Errorf("password must be at least %d characters", minPasswordLen)
	}
	var hasLetter, hasNumber bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsNumber(r) {
			hasNumber = true
		}
	}
	if !hasLetter || !hasNumber {
		return fmt.Errorf("password must contain at least one letter and one number")
	}
	if strings.Contains(password, " ") {
		return fmt.Errorf("password must not contain spaces")
	}
	return nil
}
