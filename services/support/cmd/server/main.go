package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/hacuba/support/internal/config"
	"github.com/hacuba/support/internal/metrics"
	"github.com/hacuba/support/internal/reports"
	"github.com/hacuba/support/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("support: invalid config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var store reports.Store
	if cfg.DatabaseURL != "" {
		postgres, err := reports.NewPostgresStore(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("support: store: %v", err)
		}
		defer postgres.Close()
		store = postgres
	} else {
		sqlite, err := reports.NewSQLiteStore(ctx, cfg.DevSQLitePath)
		if err != nil {
			log.Fatalf("support: development SQLite store: %v", err)
		}
		defer sqlite.Close()
		store = sqlite
		log.Printf("support: DEV_SQLITE_PATH is set — using persistent development SQLite")
	}
	reportMetrics := metrics.New(store)
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: server.NewMux(cfg, store, reportMetrics), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("support: listening on :%s", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
