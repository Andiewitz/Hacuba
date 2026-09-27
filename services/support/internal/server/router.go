package server

import (
	"net/http"

	"github.com/hacuba/support/internal/config"
	"github.com/hacuba/support/internal/handlers"
	"github.com/hacuba/support/internal/metrics"
	"github.com/hacuba/support/internal/reports"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMux(cfg config.Config, store reports.Store, reportMetrics *metrics.Metrics) http.Handler {
	h := handlers.New(store, reportMetrics, cfg.JWTSecret, cfg.ProxySecret)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"support"}`))
	})
	mux.Handle("GET /metrics", promhttp.HandlerFor(reportMetrics.Registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("POST /reports", h.CreateReport)
	mux.HandleFunc("GET /staff/reports", h.ListReports)
	mux.HandleFunc("PATCH /staff/reports/{id}", h.TransitionReport)
	return withSecurityHeaders(mux)
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
