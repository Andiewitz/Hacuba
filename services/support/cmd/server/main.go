package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.ListenAndServe() }()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("support: server error: %v", err)
		}
	case <-quit:
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("support: graceful shutdown: %v", err)
		}
	}
}
