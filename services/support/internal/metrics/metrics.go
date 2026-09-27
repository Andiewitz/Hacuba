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
	Registry        *prometheus.Registry
}

// New keeps all report data in the private store. Prometheus receives only
// bounded category and rejection labels, never email, text, IDs, or IPs.
func New(store reports.Store) *Metrics {
	registry := prometheus.NewRegistry()
	accepted := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "hacuba_support_reports_total", Help: "Accepted support reports by category."}, []string{"category"})
	rejected := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "hacuba_support_report_rejections_total", Help: "Rejected support report submissions by bounded reason."}, []string{"reason"})
	open := prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "hacuba_support_reports_open", Help: "Current number of unresolved support reports."}, func() float64 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		count, err := store.OpenCount(ctx)
		if err != nil {
			return 0
		}
		return float64(count)
	})
	registry.MustRegister(accepted, rejected, open)
	return &Metrics{ReportsAccepted: accepted, ReportsRejected: rejected, Registry: registry}
}
