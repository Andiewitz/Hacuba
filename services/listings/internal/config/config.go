package config

import (
	"fmt"
	"os"
	"time"
)

// Config contains only settings owned by the listings service. It never
// receives auth-db credentials; JWT verification uses the shared secret.
type Config struct {
	Port          string
	DatabaseURL   string
	DevSQLitePath string
	JWTSecret     []byte
	CacheTTL      time.Duration
	S3Region      string
	S3Bucket      string
}

func Load() (Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	databaseURL, devSQLitePath := os.Getenv("DATABASE_URL"), os.Getenv("DEV_SQLITE_PATH")
	if databaseURL != "" && devSQLitePath != "" {
		return Config{}, fmt.Errorf("DATABASE_URL and DEV_SQLITE_PATH cannot both be set")
	}
	return Config{Port: port, DatabaseURL: databaseURL, DevSQLitePath: devSQLitePath, JWTSecret: []byte(secret), CacheTTL: 5 * time.Minute, S3Region: os.Getenv("S3_REGION"), S3Bucket: os.Getenv("S3_BUCKET")}, nil
}
