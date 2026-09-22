package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/hacuba/listings/internal/config"
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
	} else {
		log.Print("listings: DATABASE_URL unset — using in-memory store (dev only)")
	}
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: server.NewMux(cfg, store), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("listings: listening on :%s", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
