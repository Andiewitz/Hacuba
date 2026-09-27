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
