package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hacuba/support/internal/metrics"
	"github.com/hacuba/support/internal/reports"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestCreateReportStoresPrivateDetailsAndEmitsBoundedMetrics(t *testing.T) {
	store := reports.NewMemoryStore()
	reportMetrics := metrics.New(store)
	handler := New(store, reportMetrics, []byte("support-test-secret-must-be-at-least-32-bytes"), nil)
	body := []byte(`{"category":"safety","listing_reference":"https://hacuba.test/listings/example","contact_email":"buyer@example.test","description":"The seller asked me to transfer a deposit before I could view the property."}`)
	req := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewReader(body))
	req.RemoteAddr = "198.51.100.10:4000"
	rec := httptest.NewRecorder()
	handler.CreateReport(rec, req)
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"status":"open"`) {
		t.Fatalf("create report = %d: %s", rec.Code, rec.Body.String())
	}
	metricsRec := httptest.NewRecorder()
	promhttp.HandlerFor(reportMetrics.Registry, promhttp.HandlerOpts{}).ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	output := metricsRec.Body.String()
	if !strings.Contains(output, `hacuba_support_reports_total{category="safety"} 1`) {
		t.Fatalf("accepted report metric missing: %s", output)
	}
	if strings.Contains(output, "buyer@example.test") || strings.Contains(output, "transfer a deposit") || strings.Contains(output, "example") {
		t.Fatalf("private report content leaked to metrics: %s", output)
	}
}

func TestCreateReportValidatesAndRateLimits(t *testing.T) {
	store := reports.NewMemoryStore()
	handler := New(store, metrics.New(store), []byte("support-test-secret-must-be-at-least-32-bytes"), nil)
	invalid := httptest.NewRequest(http.MethodPost, "/reports", strings.NewReader(`{"category":"bug","description":"too short"}`))
	invalid.RemoteAddr = "198.51.100.11:4000"
	invalidRec := httptest.NewRecorder()
	handler.CreateReport(invalidRec, invalid)
	if invalidRec.Code != http.StatusBadRequest || !strings.Contains(invalidRec.Body.String(), "description") {
		t.Fatalf("invalid report = %d: %s", invalidRec.Code, invalidRec.Body.String())
	}
	body := `{"category":"bug","description":"The support form loses its contents after I press the send report button."}`
	for attempt := 0; attempt < 9; attempt++ {
		req := httptest.NewRequest(http.MethodPost, "/reports", strings.NewReader(body))
		req.RemoteAddr = "198.51.100.11:4000"
		rec := httptest.NewRecorder()
		handler.CreateReport(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("report %d = %d", attempt+1, rec.Code)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/reports", strings.NewReader(body))
	req.RemoteAddr = "198.51.100.11:4000"
	rec := httptest.NewRecorder()
	handler.CreateReport(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("rate limited report = %d: %s", rec.Code, rec.Body.String())
	}
}
