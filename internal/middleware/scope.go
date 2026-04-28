package middleware

import (
	"net/http"
	"strings"

	"github.com/Mirnda/mirandaclin/pkg/response"
)

// RequireScope exige que o token contenha o scope especificado.
// admin:* concede acesso a qualquer scope.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenScope, _ := r.Context().Value(keyScope).(string)
			if !hasScope(tokenScope, scope) {
				response.Error(w, http.StatusForbidden, "permissão insuficiente")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasScope(tokenScope, required string) bool {
	if strings.Contains(tokenScope, "admin:*") {
		return true
	}
	for _, s := range strings.Fields(tokenScope) {
		if s == required {
			return true
		}
	}
	return false
}
