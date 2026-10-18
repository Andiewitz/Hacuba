package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json")
		if err := store.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","service":"support"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok","service":"support"}`))
	})
	mux.Handle("GET /metrics", promhttp.HandlerFor(reportMetrics.Registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("POST /reports", h.CreateReport)
	mux.HandleFunc("GET /staff/reports", h.ListReports)
	mux.HandleFunc("PATCH /staff/reports/{id}", h.TransitionReport)
	return withSecurityHeaders(withRequestLog("support", mux, reportMetrics))
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

func withRequestLog(service string, next http.Handler, metrics *metrics.Metrics) http.Handler {
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
		metrics.HTTPRequests.WithLabelValues(service, route, r.Method, strconv.Itoa(recorder.status)).Inc()
		metrics.HTTPLatency.WithLabelValues(service, route, r.Method).Observe(time.Since(started).Seconds())
	})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
