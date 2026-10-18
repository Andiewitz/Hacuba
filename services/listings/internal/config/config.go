package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config contains only settings owned by the listings service. It never
// receives auth-db credentials; JWT verification uses the shared secret.
type Config struct {
	Port                 string
	Production           bool
	DatabaseURL          string
	DevSQLitePath        string
	JWTSecret            []byte
	CacheTTL             time.Duration
	ShutdownTimeout      time.Duration
	S3Region             string
	S3Bucket             string
	DiscoveryProxySecret []byte
	SentryDSN            string
}

func Load() (Config, error) {
	production := strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	shutdownTimeout := 10 * time.Second
	if value := os.Getenv("SHUTDOWN_TIMEOUT"); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT must be a positive duration")
		}
		shutdownTimeout = parsed
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
		if os.Getenv("DEV_UPLOAD_PATH") != "" {
			return Config{}, fmt.Errorf("DEV_UPLOAD_PATH cannot be set when APP_ENV=production")
		}
		if os.Getenv("S3_REGION") == "" || os.Getenv("S3_BUCKET") == "" {
			return Config{}, fmt.Errorf("S3_REGION and S3_BUCKET are required when APP_ENV=production")
		}
		if os.Getenv("DISCOVERY_PROXY_SECRET") == "" {
			return Config{}, fmt.Errorf("DISCOVERY_PROXY_SECRET is required when APP_ENV=production")
		}
	}
	return Config{Port: port, Production: production, DatabaseURL: databaseURL, DevSQLitePath: devSQLitePath, JWTSecret: []byte(secret), CacheTTL: 5 * time.Minute, ShutdownTimeout: shutdownTimeout, S3Region: os.Getenv("S3_REGION"), S3Bucket: os.Getenv("S3_BUCKET"), DiscoveryProxySecret: []byte(os.Getenv("DISCOVERY_PROXY_SECRET")), SentryDSN: os.Getenv("SENTRY_DSN")}, nil
}
