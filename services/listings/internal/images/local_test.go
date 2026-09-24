package images

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStoreUploadsAndServesExpectedObject(t *testing.T) {
	store, err := NewLocalStore(filepath.Join(t.TempDir(), "uploads"), "http://localhost:3000/api/dev-uploads")
	if err != nil {
		t.Fatal(err)
	}
	key := "listings/listing-id/image.jpg"
	url, err := store.PresignPut(context.Background(), key, "image/jpeg", 3)
	if err != nil || !strings.HasSuffix(url, "/"+key) {
		t.Fatalf("presign = %q, %v", url, err)
	}

	put := httptest.NewRequest(http.MethodPut, "/dev-uploads/"+key, strings.NewReader("jpg"))
	put.Header.Set("Content-Type", "image/jpeg")
	put.ContentLength = 3
	putResult := httptest.NewRecorder()
	store.ServeHTTP(putResult, put)
	if putResult.Code != http.StatusNoContent {
		t.Fatalf("put = %d: %s", putResult.Code, putResult.Body.String())
	}
	if err := store.Head(context.Background(), key); err != nil {
		t.Fatalf("head: %v", err)
	}

	get := httptest.NewRequest(http.MethodGet, "/dev-uploads/"+key, nil)
	getResult := httptest.NewRecorder()
	store.ServeHTTP(getResult, get)
	if getResult.Code != http.StatusOK || getResult.Body.String() != "jpg" {
		t.Fatalf("get = %d, %q", getResult.Code, getResult.Body.String())
	}
}
