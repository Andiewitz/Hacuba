package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the auth service.
// All secrets come from the environment — never hardcode them.
type Config struct {
	Port string

	// DatabaseURL is the Postgres DSN. Only the auth service
	// ever uses this value — the auth-db container lives on an
	// internal Docker network unreachable from other services.
	DatabaseURL string

	// JWTSecret signs access tokens. Must be >= 32 bytes.
	JWTSecret []byte

	AccessTTL  time.Duration
	RefreshTTL time.Duration

	CookieDomain string
	CookieSecure bool
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load reads configuration from the environment with safe defaults.
func Load() (Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Dev-only fallback so `go build` / `go test` work without env.
		// main.go refuses to start without a real secret (see EnsureProdReady).
		secret = "dev-only-secret-change-me-32-bytes-minimum!!"
	}
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes, got %d", len(secret))
	}

	accessTTL := 15 * time.Minute
	if v := os.Getenv("ACCESS_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			accessTTL = d
		}
	}

	refreshTTL := 7 * 24 * time.Hour
	if v := os.Getenv("REFRESH_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			refreshTTL = d
		}
	}

	secure := true
	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			secure = b
		}
	}

	return Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		JWTSecret:    []byte(secret),
		AccessTTL:    accessTTL,
		RefreshTTL:   refreshTTL,
		CookieDomain: os.Getenv("COOKIE_DOMAIN"),
		CookieSecure: secure,
	}, nil
}

// EnsureProdReady fails fast when required prod settings are missing.
func (c Config) EnsureProdReady() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if os.Getenv("JWT_SECRET") == "" {
		return fmt.Errorf("JWT_SECRET is required in production")
	}
	return nil
}
