package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hacuba/listings/internal/config"
	"github.com/hacuba/listings/internal/listings"
)

func TestRouterMetricsUseBoundedRouteLabels(t *testing.T) {
	mux := NewMux(config.Config{JWTSecret: []byte("router-test-secret-32-bytes-minimum!")}, listings.NewMemoryStore())
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz?city=private-place", nil))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics = %d, want 200", rec.Code)
	}
	output := rec.Body.String()
	if !strings.Contains(output, `hacuba_http_requests_total{method="GET",route="GET /healthz",service="listings",status="200"} 1`) {
		t.Fatalf("health request metric missing: %s", output)
	}
	if strings.Contains(output, "private-place") {
		t.Fatalf("query value leaked into metrics: %s", output)
	}
}
