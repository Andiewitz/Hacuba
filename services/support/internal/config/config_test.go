package config

import "testing"

const testJWTSecret = "support-test-secret-must-be-at-least-32-bytes"

func TestLoadRejectsDevelopmentSQLiteInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DEV_SQLITE_PATH", ".dev/support.db")
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("SUPPORT_PROXY_SECRET", "proxy-secret")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with DEV_SQLITE_PATH in production")
	}
}

func TestLoadProductionRequiresProxySecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DEV_SQLITE_PATH", "")
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("SUPPORT_PROXY_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded without SUPPORT_PROXY_SECRET in production")
	}
}

func TestLoadRejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DEV_SQLITE_PATH", ".dev/support.db")
	t.Setenv("SHUTDOWN_TIMEOUT", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with an invalid SHUTDOWN_TIMEOUT")
	}
}
