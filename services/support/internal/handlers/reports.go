package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hacuba/authjwt"
	"github.com/hacuba/support/internal/metrics"
	"github.com/hacuba/support/internal/reports"
)

type Handler struct {
	Store    reports.Store
	Metrics  *metrics.Metrics
	JWTKey   []byte
	ProxyKey []byte
	limiter  rateLimiter
}

type rateWindow struct {
	started  time.Time
	requests int
}
type rateLimiter struct {
	sync.Mutex
	windows map[string]rateWindow
}

func New(store reports.Store, reportMetrics *metrics.Metrics, jwtKey, proxyKey []byte) *Handler {
	return &Handler{Store: store, Metrics: reportMetrics, JWTKey: jwtKey, ProxyKey: proxyKey, limiter: rateLimiter{windows: map[string]rateWindow{}}}
}

func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	if !h.allow(h.clientAddress(r)) {
		h.Metrics.ReportsRejected.WithLabelValues("rate_limited").Inc()
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many reports; try again in a minute"})
		return
	}
	var input struct {
		Category         string `json:"category"`
		ListingReference string `json:"listing_reference"`
		ContactEmail     string `json:"contact_email"`
		Description      string `json:"description"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		h.Metrics.ReportsRejected.WithLabelValues("invalid_payload").Inc()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid report payload"})
		return
	}
	report := reports.Report{ID: uuid.Must(uuid.NewV7()), Category: strings.TrimSpace(input.Category), ListingReference: strings.TrimSpace(input.ListingReference), ContactEmail: strings.TrimSpace(input.ContactEmail), Description: strings.TrimSpace(input.Description), Status: reports.StatusOpen}
	if userID, valid := optionalReporterID(r, h.JWTKey); valid {
		report.ReporterID = &userID
	} else if r.Header.Get("Authorization") != "" {
		h.Metrics.ReportsRejected.WithLabelValues("invalid_auth").Inc()
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid sign-in session"})
		return
	}
	if fields := reports.Validate(report); len(fields) > 0 {
		h.Metrics.ReportsRejected.WithLabelValues("validation").Inc()
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bad request", "fields": fields})
		return
	}
	created, err := h.Store.Create(r.Context(), report)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not save your report"})
		return
	}
	h.Metrics.ReportsAccepted.WithLabelValues(created.Category).Inc()
	writeJSON(w, http.StatusCreated, map[string]any{"id": created.ID, "status": created.Status, "created_at": created.CreatedAt})
}

func (h *Handler) clientAddress(r *http.Request) string {
	if len(h.ProxyKey) > 0 && subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Hacuba-Support-Proxy")), h.ProxyKey) == 1 {
		if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
			return forwarded
		}
	}
	return r.RemoteAddr
}

func optionalReporterID(r *http.Request, secret []byte) (uuid.UUID, bool) {
	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
		return uuid.Nil, false
	}
	claims, err := authjwt.VerifyAccess(secret, parts[1])
	if err != nil {
		return uuid.Nil, false
	}
	id, err := claims.UserID()
	return id, err == nil
}

func (h *Handler) allow(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	now := time.Now()
	h.limiter.Lock()
	defer h.limiter.Unlock()
	window := h.limiter.windows[host]
	if window.started.IsZero() || now.Sub(window.started) >= time.Minute {
		window = rateWindow{started: now}
	}
	if window.requests >= 10 {
		h.limiter.windows[host] = window
		return false
	}
	window.requests++
	h.limiter.windows[host] = window
	return true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
