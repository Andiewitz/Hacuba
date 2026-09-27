package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/authjwt"
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

func TestCreateReportKeepsLinkedListingIDPrivate(t *testing.T) {
	store := reports.NewMemoryStore()
	handler := New(store, metrics.New(store), []byte("support-test-secret-must-be-at-least-32-bytes"), nil)
	listingID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/reports", strings.NewReader(`{"category":"listing","listing_id":"`+listingID.String()+`","description":"The photos and property details do not appear to describe the same home."}`))
	rec := httptest.NewRecorder()
	handler.CreateReport(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create linked report = %d: %s", rec.Code, rec.Body.String())
	}
	items, err := store.List(t.Context(), 1)
	if err != nil || len(items) != 1 || items[0].ListingID == nil || *items[0].ListingID != listingID {
		t.Fatalf("linked report = %#v, %v", items, err)
	}
}

func TestStaffCanTriageReportsButBuyersCannot(t *testing.T) {
	store := reports.NewMemoryStore()
	secret := []byte("support-test-secret-must-be-at-least-32-bytes")
	handler := New(store, metrics.New(store), secret, nil)
	created, err := store.Create(t.Context(), reports.Report{ID: uuid.New(), Category: reports.CategoryListing, Description: "This listing repeats misleading information about the property location.", Status: reports.StatusOpen})
	if err != nil {
		t.Fatal(err)
	}
	issue := func(role string) string {
		token, err := authjwt.IssueAccess(secret, uuid.New(), role, "csrf", time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	buyerList := httptest.NewRecorder()
	buyerRequest := httptest.NewRequest(http.MethodGet, "/staff/reports", nil)
	buyerRequest.Header.Set("Authorization", "Bearer "+issue("buyer"))
	handler.ListReports(buyerList, buyerRequest)
	if buyerList.Code != http.StatusForbidden {
		t.Fatalf("buyer queue = %d", buyerList.Code)
	}
	staffList := httptest.NewRecorder()
	staffRequest := httptest.NewRequest(http.MethodGet, "/staff/reports", nil)
	staffRequest.Header.Set("Authorization", "Bearer "+issue("staff"))
	handler.ListReports(staffList, staffRequest)
	if staffList.Code != http.StatusOK || !strings.Contains(staffList.Body.String(), created.Description) {
		t.Fatalf("staff queue = %d: %s", staffList.Code, staffList.Body.String())
	}
	staffID := uuid.New()
	staffToken, err := authjwt.IssueAccess(secret, staffID, "staff", "csrf", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	transition := httptest.NewRequest(http.MethodPatch, "/staff/reports/"+created.ID.String(), strings.NewReader(`{"status":"triaged"}`))
	transition.SetPathValue("id", created.ID.String())
	transition.Header.Set("Authorization", "Bearer "+staffToken)
	transition.Header.Set("Content-Type", "application/json")
	transitionRec := httptest.NewRecorder()
	handler.TransitionReport(transitionRec, transition)
	if transitionRec.Code != http.StatusOK || !strings.Contains(transitionRec.Body.String(), `"status":"triaged"`) {
		t.Fatalf("staff transition = %d: %s", transitionRec.Code, transitionRec.Body.String())
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
