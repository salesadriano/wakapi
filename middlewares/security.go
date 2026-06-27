package middlewares

import (
	"net/http"
	"strings"
)

var securityHeaders = map[string]string{
	"Cross-Origin-Opener-Policy": "same-origin",
	"Content-Security-Policy":    "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' https: data:; form-action 'self' *.stripe.com; block-all-mixed-content;",
	"X-Frame-Options":            "DENY",
	"X-Content-Type-Options":     "nosniff",
}

const hstsHeaderValue = "max-age=31536000; includeSubDomains"

// SecurityMiddleware is a handler to add some basic security headers to responses
type SecurityMiddleware struct {
	handler http.Handler
}

func NewSecurityMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return &SecurityMiddleware{h}
	}
}

func (f *SecurityMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range securityHeaders {
		if w.Header().Get(k) == "" {
			w.Header().Set(k, v)
		}
	}

	// Only advertise HSTS when the request actually arrived over HTTPS (directly
	// or through a TLS-terminating reverse proxy). Sending it unconditionally
	// would needlessly pin HSTS on operators who intentionally serve the
	// instance over plain HTTP (e.g. on a private network).
	if isHttps(r) && w.Header().Get("Strict-Transport-Security") == "" {
		w.Header().Set("Strict-Transport-Security", hstsHeaderValue)
	}

	f.handler.ServeHTTP(w, r)
}

func isHttps(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
