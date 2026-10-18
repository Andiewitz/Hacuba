package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hacuba/listings/internal/config"
	"github.com/hacuba/listings/internal/images"
	"github.com/hacuba/listings/internal/listings"
	"github.com/hacuba/listings/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("listings: invalid config: %v", err)
	}
	if cfg.Production {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
	var store listings.Store
	if cfg.DatabaseURL != "" {
		postgres, err := listings.NewPostgresStore(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("listings: store: %v", err)
		}
		defer postgres.Close()
		store = postgres
	} else if cfg.DevSQLitePath != "" {
		sqlite, err := listings.NewSQLiteStore(context.Background(), cfg.DevSQLitePath)
		if err != nil {
			log.Fatalf("listings: development SQLite store: %v", err)
		}
		defer sqlite.Close()
		store = sqlite
		log.Printf("listings: DEV_SQLITE_PATH is set — using persistent development SQLite")
	}
	var objects images.ObjectStore
	if cfg.S3Region != "" || cfg.S3Bucket != "" {
		s3Store, err := images.NewS3Store(context.Background(), cfg.S3Region, cfg.S3Bucket)
		if err != nil {
			log.Fatalf("listings: image store: %v", err)
		}
		objects = s3Store
	} else if root := developmentUploadRoot(cfg); root != "" {
		publicURL := os.Getenv("DEV_UPLOAD_URL")
		if publicURL == "" {
			publicURL = fmt.Sprintf("http://localhost:%s/dev-uploads", cfg.Port)
		}
		localStore, err := images.NewLocalStore(root, publicURL)
		if err != nil {
			log.Fatalf("listings: development image store: %v", err)
		}
		objects = localStore
		log.Printf("listings: using development local image uploads at %s", root)
	}
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: server.NewMuxWithObjects(cfg, store, objects), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("listings: listening on :%s", cfg.Port)
	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.ListenAndServe() }()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listings: server error: %v", err)
		}
	case <-quit:
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("listings: graceful shutdown: %v", err)
		}
	}
}

// developmentUploadRoot keeps the disk-backed object store behind explicit
// development settings. E2E runs use it with disposable Postgres databases so
// the browser can exercise the same presign, upload, and registration contract
// without production credentials. Production deployments leave DEV_UPLOAD_PATH
// unset and must configure S3 instead.
func developmentUploadRoot(cfg config.Config) string {
	if root := os.Getenv("DEV_UPLOAD_PATH"); root != "" {
		return root
	}
	if cfg.DevSQLitePath != "" {
		return filepath.Join(filepath.Dir(cfg.DevSQLitePath), "uploads")
	}
	return ""
}
