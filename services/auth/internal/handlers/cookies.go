package handlers

import (
	"net/http"
	"time"

	"github.com/hacuba/auth/internal/config"
)

const (
	refreshCookieName = "refresh_token"
	csrfCookieName    = "csrf_token"
	csrfHeaderName    = "X-CSRF-Token"
)

// setSessionCookies stores the refresh token as HttpOnly (JS can't read it)
// and the CSRF token as readable (JS must echo it back in X-CSRF-Token).
func setSessionCookies(w http.ResponseWriter, cfg config.Config, refreshRaw, csrfRaw string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshRaw,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrfRaw,
		Path:     "/",
		Domain:   cfg.CookieDomain,
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: false,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookies(w http.ResponseWriter, cfg config.Config) {
	for _, name := range []string{refreshCookieName, csrfCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			Domain:   cfg.CookieDomain,
			MaxAge:   -1,
			HttpOnly: name == refreshCookieName,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
		})
	}
}
