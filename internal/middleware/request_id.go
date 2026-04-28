package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestID gera um UUID v4 por requisição, injeta no contexto e devolve no header X-Request-ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), keyRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
