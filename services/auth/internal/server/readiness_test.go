package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hacuba/auth/internal/users"
)

type unavailableAuthStore struct{ users.Store }

func (unavailableAuthStore) Ping(context.Context) error { return errors.New("database unavailable") }

func TestRouterReadyzReflectsStoreAvailability(t *testing.T) {
	for name, store := range map[string]users.Store{
		"ready":       users.NewMemoryStore(),
		"unavailable": unavailableAuthStore{users.NewMemoryStore()},
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewMux(testConfig(), store).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
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
