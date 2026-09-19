package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/hacuba/auth/internal/config"
	"github.com/hacuba/auth/internal/users"
	"golang.org/x/crypto/bcrypt"
)

// loginLimiter is a minimal in-memory rate limiter for the login endpoint:
// max 5 failures per email per 15 minutes. Production should move this to
// Redis so limits hold across replicas — the interface stays the same.
type loginLimiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
}

var limiter = &loginLimiter{failures: make(map[string][]time.Time)}

func (l *loginLimiter) blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-15 * time.Minute)
	kept := l.failures[key][:0]
	for _, t := range l.failures[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	l.failures[key] = kept
	return len(kept) >= 5
}

func (l *loginLimiter) recordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures[key] = append(l.failures[key], time.Now())
}

func (l *loginLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

// dummyPasswordHash is a cost-12 bcrypt hash of a random password nobody
// knows. Comparing against it on unknown emails costs the same as a real
// password check, so the generic 401 below does not leak account existence
// through timing.
const dummyPasswordHash = "$2a$12$X2Z2hxRKLbN9xZTVlwdauOB8BfYId1bTUEi.sDDtaijtHeDddWhFO"

// Login handles POST /auth/login.
//
// Flow: rate-limit -> lookup by email -> bcrypt compare (constant-time) ->
// issue access JWT + rotating refresh session + CSRF pair.
// Failures always return generic 401 so callers can't enumerate accounts.
func Login(cfg config.Config, store users.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		email := users.NormalizeEmail(req.Email)
		if email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "email and password are required")
			return
		}
		if limiter.blocked(email) {
			writeError(w, http.StatusTooManyRequests, "too many attempts, try again later")
			return
		}

		user, err := store.GetUserByEmail(r.Context(), email)
		if err != nil {
			if errors.Is(err, users.ErrNotFound) {
				// Burn the same bcrypt cost as a real check so unknown emails
				// do not return measurably faster (see dummyPasswordHash).
				_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(req.Password))
				limiter.recordFailure(email)
				writeError(w, http.StatusUnauthorized, "invalid credentials")
				return
			}
			writeError(w, http.StatusInternalServerError, "login failed")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			limiter.recordFailure(email)
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		limiter.reset(email)

		sess, err := issueSession(r.Context(), w, cfg, store, user)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create session")
			return
		}
		writeJSON(w, http.StatusOK, sess)
	}
}
