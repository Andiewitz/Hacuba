package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPanicRecoveryReturnsGenericServerError(t *testing.T) {
	rec := httptest.NewRecorder()
	withPanicRecovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("do not expose internal failure")
	})).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("panic response = %d, want 500", rec.Code)
	}
	if rec.Body.String() != "Internal Server Error\n" {
		t.Fatalf("panic response body = %q", rec.Body.String())
	}
}
