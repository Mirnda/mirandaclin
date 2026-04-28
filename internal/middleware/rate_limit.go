package middleware

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/Mirnda/mirandaclin/pkg/response"
)

// RateLimit implementa sliding window por IP usando Redis (ou Noop).
// limit: máximo de requisições; window: duração da janela.
func RateLimit(c cache.Cache, limit int64, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			key := fmt.Sprintf("rl:ip:%s:%s", ip, r.URL.Path)

			count, err := c.Incr(r.Context(), key)
			if err == nil && count == 1 {
				_ = c.Expire(r.Context(), key, window)
			}
			if count > limit {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				response.Error(w, http.StatusTooManyRequests, "muitas requisições, tente novamente em instantes")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
