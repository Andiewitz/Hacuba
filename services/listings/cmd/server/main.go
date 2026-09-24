package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	var store listings.Store = listings.NewMemoryStore()
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
	} else {
		log.Print("listings: DATABASE_URL unset — using in-memory store (dev only)")
	}
	var objects images.ObjectStore
	if cfg.S3Region != "" || cfg.S3Bucket != "" {
		s3Store, err := images.NewS3Store(context.Background(), cfg.S3Region, cfg.S3Bucket)
		if err != nil {
			log.Fatalf("listings: image store: %v", err)
		}
		objects = s3Store
	} else if cfg.DevSQLitePath != "" {
		root := os.Getenv("DEV_UPLOAD_PATH")
		if root == "" {
			root = filepath.Join(filepath.Dir(cfg.DevSQLitePath), "uploads")
		}
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
	log.Fatal(srv.ListenAndServe())
}
