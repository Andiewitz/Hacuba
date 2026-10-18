package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json")
		if err := store.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","service":"auth"}`))
			return
		}
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

	return withSecurityHeaders(withRequestLog("auth", mux))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

func withRequestLog(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		w.Header().Set("X-Request-ID", requestID)
		recorder := &statusRecorder{ResponseWriter: w}
		started := time.Now()
		next.ServeHTTP(recorder, r)
		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		slog.Info("http request", "service", service, "request_id", requestID, "method", r.Method, "route", route, "status", recorder.status, "latency_ms", time.Since(started).Milliseconds())
	})
}

// withSecurityHeaders sets baseline response headers on every route,
// including error paths (headers are set before handlers run).
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
