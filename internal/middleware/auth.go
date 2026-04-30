package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/response"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims representa o payload do JWT local emitido pela aplicação.
type Claims struct {
	jwt.RegisteredClaims
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	Scope    string `json:"scope"`
}

// Auth valida o token JWT do header Authorization e injeta claims no contexto.
func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ctx = r.Context()
			var log = logger.FromContext(ctx)
			ip := remoteIP(r)

			tokenStr := bearerToken(r)
			if tokenStr == "" {
				log.Info("token ausente", logger.String("ip", ip))
				response.Error(w, http.StatusUnauthorized, "token ausente")
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				log.Info("token inválido ou expirado", logger.String("ip", ip), logger.Err(err))
				response.Error(w, http.StatusUnauthorized, "token inválido ou expirado")
				return
			}

			tenantID, err := uuid.Parse(claims.TenantID)
			if err != nil {
				log.Info("token malformado: tenant_id inválido", logger.String("ip", ip))
				response.Error(w, http.StatusUnauthorized, "token malformado")
				return
			}
			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				log.Info("token malformado: user_id inválido", logger.String("ip", ip))
				response.Error(w, http.StatusUnauthorized, "token malformado")
				return
			}

			EnrichRequestLog(ctx, tenantID.String(), userID.String())

			ctx = context.WithValue(ctx, keyTenantID, tenantID)
			ctx = context.WithValue(ctx, keyUserID, userID)
			ctx = context.WithValue(ctx, keyRole, claims.Role)
			ctx = context.WithValue(ctx, keyScope, claims.Scope)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
