package middleware

import (
	"net/http"

	"github.com/Mirnda/mirandaclin/pkg/response"
)

func APIKey(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Api-Key") != secret {
				response.Error(w, http.StatusUnauthorized, "api key inválida")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
