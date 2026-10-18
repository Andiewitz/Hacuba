package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/listings/internal/config"
	"github.com/hacuba/listings/internal/handlers"
	"github.com/hacuba/listings/internal/images"
	"github.com/hacuba/listings/internal/listings"
	"github.com/hacuba/listings/internal/metrics"
	"github.com/hacuba/listings/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMux(cfg config.Config, store listings.Store) http.Handler {
	return NewMuxWithObjects(cfg, store, nil)
}

// NewMuxWithObjects injects S3 only in deployments that configure it; tests
// and local API work can still exercise every non-image route without AWS.
func NewMuxWithObjects(cfg config.Config, store listings.Store, objects images.ObjectStore) http.Handler {
	observability := metrics.New()
	h := handlers.Handler{Store: store, Objects: objects, DiscoveryProxySecret: cfg.DiscoveryProxySecret}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"listings"}`))
	})
	mux.Handle("GET /metrics", promhttp.HandlerFor(observability.Registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json")
		if err := store.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","service":"listings"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok","service":"listings"}`))
	})
	mux.HandleFunc("GET /listings", h.ListPublic)
	mux.HandleFunc("GET /discover", h.Discover)
	mux.HandleFunc("GET /listings/{id}", h.GetPublic)
	mux.HandleFunc("GET /listings/{id}/related", h.Related)
	mux.HandleFunc("POST /discovery/events", h.RecordDiscoveryEvents)
	if local, ok := objects.(interface {
		ServeHTTP(http.ResponseWriter, *http.Request)
	}); ok {
		mux.Handle("GET /dev-uploads/{path...}", http.HandlerFunc(local.ServeHTTP))
		mux.Handle("PUT /dev-uploads/{path...}", http.HandlerFunc(local.ServeHTTP))
	}

	authenticated := func(next http.Handler) http.Handler {
		return middleware.Authenticate(cfg.JWTSecret, middleware.RequireSeller(next))
	}
	writes := func(next http.Handler) http.Handler { return authenticated(middleware.RequireCSRF(next)) }
	staffWrites := func(next http.Handler) http.Handler {
		return middleware.Authenticate(cfg.JWTSecret, middleware.RequireStaff(middleware.RequireCSRF(next)))
	}
	mux.Handle("GET /me/listings", authenticated(http.HandlerFunc(h.ListMine)))
	mux.Handle("POST /listings", writes(http.HandlerFunc(h.Create)))
	mux.Handle("PATCH /listings/{id}", writes(http.HandlerFunc(h.Patch)))
	mux.Handle("POST /listings/{id}/publish", writes(http.HandlerFunc(h.Publish)))
	mux.Handle("POST /listings/{id}/unpublish", writes(http.HandlerFunc(h.Unpublish)))
	mux.Handle("POST /listings/{id}/close", writes(http.HandlerFunc(h.Close)))
	mux.Handle("DELETE /listings/{id}", writes(http.HandlerFunc(h.Archive)))
	mux.Handle("POST /listings/{id}/images/presign", writes(http.HandlerFunc(h.PresignImage)))
	mux.Handle("POST /listings/{id}/images", writes(http.HandlerFunc(h.RegisterImage)))
	mux.Handle("DELETE /listings/{id}/images/{imageId}", writes(http.HandlerFunc(h.DeleteImage)))
	mux.Handle("POST /staff/listings/{id}/moderate", staffWrites(http.HandlerFunc(h.Moderate)))
	return withRequestLog("listings", mux, observability)
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
