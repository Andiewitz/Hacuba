// Command server is the auth service entrypoint.
//
// Routes:
//
//	POST /auth/register  public, CSRF-exempt (no session yet)
//	POST /auth/login     public, rate-limited, CSRF-exempt (no session yet)
//	POST /auth/refresh   refresh cookie + CSRF double-submit, rotates session
//	POST /auth/logout    clears session + cookies
//	GET  /auth/me        Bearer JWT -> own user only (isolation by construction)
//	GET  /healthz        unauthenticated liveness probe
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/server"
	"github.com/hacuba/auth/internal/users"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("auth: invalid config: %v", err)
	}
	// Production requires Postgres. Persistent local development must opt in
	// to the separate SQLite store so a production launch cannot fall back.
	if cfg.DevSQLitePath == "" && cfg.DatabaseURL == "" {
		log.Fatal("auth: not production-ready: DATABASE_URL is required; set DEV_SQLITE_PATH only for local development")
	}

	store, err := newStore(cfg)
	if err != nil {
		log.Fatalf("auth: store: %v", err)
	}
	if closer, ok := store.(interface{ Close() }); ok {
		defer closer.Close()
	}

	mux := server.NewMux(cfg, store)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("auth: listening on :%s", cfg.Port)
	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.ListenAndServe() }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("auth: server error: %v", err)
		}
	case <-quit:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("auth: graceful shutdown: %v", err)
		}
	}
}

// newStore connects to private Postgres in production or the separate,
// persistent SQLite schema only when DEV_SQLITE_PATH is explicitly set.
func newStore(cfg config.Config) (users.Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if cfg.DevSQLitePath != "" {
		store, err := users.NewSQLiteStore(ctx, cfg.DevSQLitePath)
		if err == nil {
			log.Print("auth: DEV_SQLITE_PATH is set — using persistent development SQLite")
		}
		return store, err
	}
	return users.NewPostgresStore(ctx, cfg.DatabaseURL)
}
