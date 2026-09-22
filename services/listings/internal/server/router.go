package server

import (
	"net/http"

	"github.com/hacuba/listings/internal/config"
	"github.com/hacuba/listings/internal/handlers"
	"github.com/hacuba/listings/internal/listings"
	"github.com/hacuba/listings/internal/middleware"
)

func NewMux(cfg config.Config, store listings.Store) http.Handler {
	h := handlers.Handler{Store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"listings"}`))
	})
	mux.HandleFunc("GET /listings", h.ListPublic)
	mux.HandleFunc("GET /listings/{id}", h.GetPublic)

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
	return mux
}
