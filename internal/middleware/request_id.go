package middleware

import (
	"net/http"

	"github.com/Mirnda/mirandaclin/pkg/logger"
)

// RequestID gera um UUID v4 por requisição, injeta no contexto e devolve no header X-Request-ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// id := uuid.New().String()  || TODO REMOVER
		// ctx := context.WithValue(r.Context(), keyRequestID, id)

		ctx := logger.WithRequestID(r.Context())

		w.Header().Set("X-Request-ID", logger.GetRequestID(ctx))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
