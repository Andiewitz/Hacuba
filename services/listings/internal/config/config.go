package config

import (
	"fmt"
	"os"
	"time"
)

// Config contains only settings owned by the listings service. It never
// receives auth-db credentials; JWT verification uses the shared secret.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   []byte
	CacheTTL    time.Duration
	S3Region    string
	S3Bucket    string
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
	return Config{Port: port, DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: []byte(secret), CacheTTL: 5 * time.Minute, S3Region: os.Getenv("S3_REGION"), S3Bucket: os.Getenv("S3_BUCKET")}, nil
}
