package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port          string
	Production    bool
	DatabaseURL   string
	DevSQLitePath string
	JWTSecret     []byte
	ProxySecret   []byte
}

func Load() (Config, error) {
	production := strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	databaseURL, devSQLitePath := os.Getenv("DATABASE_URL"), os.Getenv("DEV_SQLITE_PATH")
	if databaseURL != "" && devSQLitePath != "" {
		return Config{}, fmt.Errorf("DATABASE_URL and DEV_SQLITE_PATH cannot both be set")
	}
	if databaseURL == "" && devSQLitePath == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required unless DEV_SQLITE_PATH is explicitly set for development")
	}
	if production {
		if databaseURL == "" {
			return Config{}, fmt.Errorf("DATABASE_URL is required when APP_ENV=production")
		}
		if devSQLitePath != "" {
			return Config{}, fmt.Errorf("DEV_SQLITE_PATH cannot be set when APP_ENV=production")
		}
		if os.Getenv("SUPPORT_PROXY_SECRET") == "" {
			return Config{}, fmt.Errorf("SUPPORT_PROXY_SECRET is required when APP_ENV=production")
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	return Config{Port: port, Production: production, DatabaseURL: databaseURL, DevSQLitePath: devSQLitePath, JWTSecret: []byte(secret), ProxySecret: []byte(os.Getenv("SUPPORT_PROXY_SECRET"))}, nil
}
