package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hacuba/listings/internal/config"
	"github.com/hacuba/listings/internal/listings"
)

type unavailableListingsStore struct{ listings.Store }

func (unavailableListingsStore) Ping(context.Context) error {
	return errors.New("database unavailable")
}

func TestRouterReadyzReflectsStoreAvailability(t *testing.T) {
	for name, store := range map[string]listings.Store{
		"ready":       listings.NewMemoryStore(),
		"unavailable": unavailableListingsStore{listings.NewMemoryStore()},
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewMux(config.Config{JWTSecret: []byte("router-test-secret-32-bytes-minimum!")}, store).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
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
