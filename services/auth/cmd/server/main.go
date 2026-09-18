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
	"github.com/hacuba/auth/internal/handlers"
	"github.com/hacuba/auth/internal/middleware"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"auth"}`))
	})

	// Public: no session exists yet, so CSRF binding is impossible.
	// Rate limiting lives inside Login; registration is validated + hashed.
	mux.Handle("POST /auth/register", handlers.Register(cfg, store))
	mux.Handle("POST /auth/login", handlers.Login(cfg, store))

	// Session-bound: refresh/logout validate the CSRF double-submit inside
	// their handlers (they must run before Authenticate, which needs the
	// short-lived access JWT that may already have expired).
	mux.Handle("POST /auth/refresh",
		middleware.RequireCSRF(handlers.Refresh(cfg, store)))
	mux.Handle("POST /auth/logout", handlers.Logout(cfg, store))

	// Authenticated: Bearer JWT -> context UUID -> own row only.
	me := middleware.Authenticate(cfg.JWTSecret, store, handlers.Me(store))
	mux.Handle("GET /auth/me", me)

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
