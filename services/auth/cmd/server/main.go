// Command server is the auth service entrypoint.
//
// It wires HTTP routes to handlers. Auth logic lives in
// internal/handlers (registration.go, login.go, ...), token logic in
// internal/tokens, and request gating in internal/middleware.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/hacuba/auth/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("auth: invalid config: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"auth"}`))
	})

	// Full auth routes (register/login/refresh/logout/me) are wired in
	// later commits. Health check proves the module builds and serves.
	_ = cfg

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
