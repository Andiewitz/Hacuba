package server

import (
	"net/http"

	"github.com/hacuba/listings/internal/config"
	"github.com/hacuba/listings/internal/handlers"
	"github.com/hacuba/listings/internal/images"
	"github.com/hacuba/listings/internal/listings"
	"github.com/hacuba/listings/internal/middleware"
)

func NewMux(cfg config.Config, store listings.Store) http.Handler {
	return NewMuxWithObjects(cfg, store, nil)
}

// NewMuxWithObjects injects S3 only in deployments that configure it; tests
// and local API work can still exercise every non-image route without AWS.
func NewMuxWithObjects(cfg config.Config, store listings.Store, objects images.ObjectStore) http.Handler {
	h := handlers.Handler{Store: store, Objects: objects}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"listings"}`))
	})
	mux.HandleFunc("GET /listings", h.ListPublic)
	mux.HandleFunc("GET /listings/{id}", h.GetPublic)
	if local, ok := objects.(interface {
		ServeHTTP(http.ResponseWriter, *http.Request)
	}); ok {
		mux.Handle("GET /dev-uploads/{path...}", http.HandlerFunc(local.ServeHTTP))
		mux.Handle("PUT /dev-uploads/{path...}", http.HandlerFunc(local.ServeHTTP))
	}

	authenticated := func(next http.Handler) http.Handler {
		return middleware.Authenticate(cfg.JWTSecret, middleware.RequireSeller(next))
	}
	writes := func(next http.Handler) http.Handler { return authenticated(middleware.RequireCSRF(next)) }
	mux.Handle("GET /me/listings", authenticated(http.HandlerFunc(h.ListMine)))
	mux.Handle("POST /listings", writes(http.HandlerFunc(h.Create)))
	mux.Handle("PATCH /listings/{id}", writes(http.HandlerFunc(h.Patch)))
	mux.Handle("POST /listings/{id}/publish", writes(http.HandlerFunc(h.Publish)))
	mux.Handle("POST /listings/{id}/unpublish", writes(http.HandlerFunc(h.Unpublish)))
	mux.Handle("POST /listings/{id}/close", writes(http.HandlerFunc(h.Close)))
	mux.Handle("DELETE /listings/{id}", writes(http.HandlerFunc(h.Archive)))
	mux.Handle("POST /listings/{id}/images/presign", writes(http.HandlerFunc(h.PresignImage)))
	mux.Handle("POST /listings/{id}/images", writes(http.HandlerFunc(h.RegisterImage)))
	mux.Handle("DELETE /listings/{id}/images/{imageId}", writes(http.HandlerFunc(h.DeleteImage)))
	return mux
}
