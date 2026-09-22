package server

import (
	"net/http"

	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/handlers"
	"github.com/hacuba/auth/internal/middleware"
	"github.com/hacuba/auth/internal/users"
)

// NewMux builds the full auth HTTP router. cmd/server and all
// router-level tests share this constructor so contract tests can never
// drift from production wiring.
func NewMux(cfg config.Config, store users.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"auth"}`))
	})

	// Public: no session exists yet, so CSRF binding is impossible.
	mux.Handle("POST /auth/register", handlers.Register(cfg, store))
	mux.Handle("POST /auth/login", handlers.Login(cfg, store))

	// Session-bound: refresh validates the CSRF double-submit inside its
	// handler (must run before Authenticate, which needs the short-lived
	// access JWT that may already have expired).
	mux.Handle("POST /auth/refresh",
		middleware.RequireCSRF(handlers.Refresh(cfg, store)))
	mux.Handle("POST /auth/logout", handlers.Logout(cfg, store))

	// Authenticated: Bearer JWT -> context UUID -> own row only.
	me := middleware.Authenticate(cfg.JWTSecret, store, handlers.Me(store))
	mux.Handle("GET /auth/me", me)
	becomeSeller := middleware.Authenticate(cfg.JWTSecret, store,
		middleware.RequireCSRF(handlers.BecomeSeller(cfg, store)))
	mux.Handle("POST /auth/become-seller", becomeSeller)

	return withSecurityHeaders(mux)
}

// withSecurityHeaders sets baseline response headers on every route,
// including error paths (headers are set before handlers run).
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
