package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	DatabaseURL   string
	DevSQLitePath string
	JWTSecret     []byte
	ProxySecret   []byte
}

func Load() (Config, error) {
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	return Config{Port: port, DatabaseURL: databaseURL, DevSQLitePath: devSQLitePath, JWTSecret: []byte(secret), ProxySecret: []byte(os.Getenv("SUPPORT_PROXY_SECRET"))}, nil
}
