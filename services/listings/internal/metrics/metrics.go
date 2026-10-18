package metrics

import "github.com/prometheus/client_golang/prometheus"

// Metrics is an isolated registry for the listings service. Its labels are all
// bounded protocol values so requests cannot create attacker-controlled series.
type Metrics struct {
	HTTPRequests *prometheus.CounterVec
	HTTPLatency  *prometheus.HistogramVec
	Registry     *prometheus.Registry
}

func New() *Metrics {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hacuba_http_requests_total",
		Help: "HTTP requests by service, route, method, and status.",
	}, []string{"service", "route", "method", "status"})
	latency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "hacuba_http_request_duration_seconds",
		Help:    "HTTP request latency by service and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "route", "method"})
	registry.MustRegister(requests, latency)
	return &Metrics{HTTPRequests: requests, HTTPLatency: latency, Registry: registry}
}
