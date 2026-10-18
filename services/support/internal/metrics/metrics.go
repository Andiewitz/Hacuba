package metrics

import (
	"context"
	"time"

	"github.com/hacuba/support/internal/reports"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	ReportsAccepted *prometheus.CounterVec
	ReportsRejected *prometheus.CounterVec
	HTTPRequests    *prometheus.CounterVec
	HTTPLatency     *prometheus.HistogramVec
	Registry        *prometheus.Registry
}

// New keeps all report data in the private store. Prometheus receives only
// bounded category and rejection labels, never email, text, IDs, or IPs.
func New(store reports.Store) *Metrics {
	registry := prometheus.NewRegistry()
	accepted := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "hacuba_support_reports_total", Help: "Accepted support reports by category."}, []string{"category"})
	rejected := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "hacuba_support_report_rejections_total", Help: "Rejected support report submissions by bounded reason."}, []string{"reason"})
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "hacuba_http_requests_total", Help: "HTTP requests by service, route, method, and status."}, []string{"service", "route", "method", "status"})
	latency := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "hacuba_http_request_duration_seconds", Help: "HTTP request latency by service and route.", Buckets: prometheus.DefBuckets}, []string{"service", "route", "method"})
	open := prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "hacuba_support_reports_open", Help: "Current number of unresolved support reports."}, func() float64 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		count, err := store.OpenCount(ctx)
		if err != nil {
			return 0
		}
		return float64(count)
	})
	registry.MustRegister(accepted, rejected, requests, latency, open)
	return &Metrics{ReportsAccepted: accepted, ReportsRejected: rejected, HTTPRequests: requests, HTTPLatency: latency, Registry: registry}
}
