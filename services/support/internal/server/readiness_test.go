package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hacuba/support/internal/config"
	"github.com/hacuba/support/internal/metrics"
	"github.com/hacuba/support/internal/reports"
)

type unavailableReportsStore struct{ reports.Store }

func (unavailableReportsStore) Ping(context.Context) error { return errors.New("database unavailable") }

func TestRouterReadyzReflectsStoreAvailability(t *testing.T) {
	for name, store := range map[string]reports.Store{
		"ready":       reports.NewMemoryStore(),
		"unavailable": unavailableReportsStore{reports.NewMemoryStore()},
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewMux(config.Config{JWTSecret: []byte("router-test-secret-32-bytes-minimum!")}, store, metrics.New(store)).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			want := http.StatusOK
			if name == "unavailable" {
				want = http.StatusServiceUnavailable
			}
			if rec.Code != want {
				t.Fatalf("readyz = %d, want %d", rec.Code, want)
			}
		})
	}
}
