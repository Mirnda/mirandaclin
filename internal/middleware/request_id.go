package middleware

import (
	"net/http"

	"github.com/Mirnda/mirandaclin/pkg/logger"
)

// RequestID gera um UUID v4 por requisição, injeta no contexto e devolve no header X-Request-ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logger.WithRequestID(r.Context())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
