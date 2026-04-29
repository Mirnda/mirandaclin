package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders aplica headers de segurança em todas as respostas.
// HSTS só é enviado fora de development para não quebrar HTTP local.
func SecurityHeaders(env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "0")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
			if env != "development" {
				w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}
			// Swagger UI precisa de inline scripts/styles — relaxa CSP apenas nessa rota
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
			} else {
				w.Header().Set("Content-Security-Policy", "default-src 'none'")
			}
			next.ServeHTTP(w, r)
		})
	}
}
