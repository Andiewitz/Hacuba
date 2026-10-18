package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hacuba/auth/internal/users"
)

func TestRouterMetricsUseBoundedRouteLabels(t *testing.T) {
	mux := NewMux(testConfig(), users.NewMemoryStore())
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz?email=private@example.com", nil))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics = %d, want 200", rec.Code)
	}
	output := rec.Body.String()
	if !strings.Contains(output, `hacuba_http_requests_total{method="GET",route="GET /healthz",service="auth",status="200"} 1`) {
		t.Fatalf("health request metric missing: %s", output)
	}
	if strings.Contains(output, "private@example.com") {
		t.Fatalf("query value leaked into metrics: %s", output)
	}
}
