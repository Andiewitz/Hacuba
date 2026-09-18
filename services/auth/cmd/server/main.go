// Command server is the auth service entrypoint.
//
// Routes:
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
	}

	log.Printf("auth: listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("auth: server error: %v", err)
	}
}

// newStore connects to the private Postgres when DATABASE_URL is set.
// Without it (unit tests, `go run` without docker) it falls back to an
// in-memory store so the service still boots — production Compose always
// sets DATABASE_URL to auth-db, the only network route to that database.
func newStore(cfg config.Config) (users.Store, error) {
	if cfg.DatabaseURL == "" {
		log.Print("auth: DATABASE_URL unset — using in-memory store (dev only)")
		return users.NewMemoryStore(), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return users.NewPostgresStore(ctx, cfg.DatabaseURL)
}
