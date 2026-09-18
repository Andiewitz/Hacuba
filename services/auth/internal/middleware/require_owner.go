package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// RequireOwner is the per-user isolation gate for future resource handlers.
//
// Pattern: extract the resource owner's UUID (from URL, e.g. /users/{id}/...,
// or from the loaded DB row's owner_id) and compare it to the JWT-verified
// caller UUID. Mismatch returns 404 — not 403 — so attackers can't probe
// whether another user's resource IDs exist.
//
// SQL-level enforcement (defense in depth) must mirror this check:
//
//	SELECT ... FROM listings WHERE id = $1 AND owner_id = $2
//	-- $2 is the JWT sub from UserIDFromContext, never client input.
//
// Auth-service endpoints like GET /auth/me don't need this wrapper because
// they never take a user ID parameter at all — they read the caller UUID
// straight from context, which is isolation by construction.
func RequireOwner(getOwnerID func(r *http.Request) (uuid.UUID, bool), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callerID, ok := UserIDFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		ownerID, ok := getOwnerID(r)
		if !ok || ownerID != callerID {
			// Deliberate 404: hides existence of other users' resources.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// OwnerIDFromPath is a helper for routes like /users/{id}/... using the
// stdlib ServeMux path value. Returns false when the segment is missing or
// not a UUID, which RequireOwner maps to 404.
func OwnerIDFromPath(r *http.Request, param string) (uuid.UUID, bool) {
	raw := r.PathValue(param)
	if raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
