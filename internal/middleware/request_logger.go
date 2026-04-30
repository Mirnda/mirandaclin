package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Mirnda/mirandaclin/pkg/logger"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

type requestLogFields struct {
	tenantID string
	userID   string
}

type requestLogKey struct{}

// EnrichRequestLog é chamado pelo middleware de auth para adicionar tenant_id e user_id ao log da requisição.
// Funciona porque requestLogFields é um ponteiro compartilhado no contexto — qualquer ctx derivado o enxerga.
func EnrichRequestLog(ctx context.Context, tenantID, userID string) {
	if f, ok := ctx.Value(requestLogKey{}).(*requestLogFields); ok {
		f.tenantID = tenantID
		f.userID = userID
	}
}

// RequestLogger loga method, path, status, duration_ms, request_id, ip, e — quando autenticado — user_id e tenant_id.
func RequestLogger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			fields := &requestLogFields{}
			ctx := context.WithValue(r.Context(), requestLogKey{}, fields)

			requestID := RequestIDFromContext(r.Context())
			tracedLog := log.With(logger.String("request_id", requestID))
			ctx = logger.WithContext(ctx, tracedLog)

			method := r.Method
			path := r.URL.Path
			accept := r.Header.Get("Accept")
			acceptEncoding := r.Header.Get("AcceptEncoding")
			connection := r.Header.Get("Connection")
			contentLength := r.Header.Get("Content-Length")
			contentType := r.Header.Get("Content-Type")
			userAgent := r.Header.Get("User-Agent")

			r = r.WithContext(ctx)
			next.ServeHTTP(sw, r)

			logFields := []logger.Field{
				logger.String("method", method),
				logger.String("path", path),
				logger.Int("status", sw.status),
				logger.Int64("duration_ms", time.Since(start).Milliseconds()),
				logger.String("request_id", RequestIDFromContext(r.Context())),
				logger.String("ip", remoteIP(r)),
				logger.String("accept", accept),
				logger.String("accept_encoding", acceptEncoding),
				logger.String("connection", connection),
				logger.String("content_length", contentLength),
				logger.String("user_agent", userAgent),
				logger.String("content_type", contentType),
			}
			if fields.userID != "" {
				logFields = append(logFields, logger.String("user_id", fields.userID))
			}
			if fields.tenantID != "" {
				logFields = append(logFields, logger.String("tenant_id", fields.tenantID))
			}
			log.Debug("request", logFields...)
		})
	}
}

func remoteIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip, _, _ := net.SplitHostPort(fwd)
		if ip != "" {
			return ip
		}
		return fwd
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
