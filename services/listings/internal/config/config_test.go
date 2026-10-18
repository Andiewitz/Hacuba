package config

import "testing"

const testJWTSecret = "listings-test-secret-must-be-32-bytes!!"

func TestLoadRejectsImplicitInMemoryStore(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DEV_SQLITE_PATH", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded without DATABASE_URL or DEV_SQLITE_PATH")
	}
}

func TestLoadAllowsExplicitDevelopmentSQLite(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DEV_SQLITE_PATH", ".dev/listings.db")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DevSQLitePath != ".dev/listings.db" || cfg.DatabaseURL != "" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadProductionRequiresObjectStoreAndProxySecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DEV_SQLITE_PATH", "")
	t.Setenv("DEV_UPLOAD_PATH", "")
	t.Setenv("S3_REGION", "")
	t.Setenv("S3_BUCKET", "")
	t.Setenv("DISCOVERY_PROXY_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded without production image and proxy configuration")
	}
}

func TestLoadRejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DEV_SQLITE_PATH", ".dev/listings.db")
	t.Setenv("SHUTDOWN_TIMEOUT", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded with an invalid SHUTDOWN_TIMEOUT")
	}
}
