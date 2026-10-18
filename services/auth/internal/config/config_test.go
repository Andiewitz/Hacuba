package config

import "testing"

func TestLoadRejectsDevelopmentSQLiteInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DEV_SQLITE_PATH", ".dev/auth.db")
	t.Setenv("JWT_SECRET", "auth-test-secret-must-be-at-least-32-bytes")
	t.Setenv("COOKIE_SECURE", "true")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with DEV_SQLITE_PATH in production")
	}
}

func TestLoadProductionRequiresExplicitJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DEV_SQLITE_PATH", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("COOKIE_SECURE", "true")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded without a production JWT_SECRET")
	}
}
