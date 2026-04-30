package main

import (
	"net/http"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/appointment"
	"github.com/Mirnda/mirandaclin/internal/domain/clinic"
	"github.com/Mirnda/mirandaclin/internal/domain/consultation"
	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/user"
	"github.com/Mirnda/mirandaclin/internal/health"
	"github.com/Mirnda/mirandaclin/internal/infra/cache"
	"github.com/Mirnda/mirandaclin/internal/middleware"
	"github.com/Mirnda/mirandaclin/pkg/config"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "github.com/Mirnda/mirandaclin/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type handlers struct {
	user         *user.Handler
	invite       *invite.Handler
	clinic       *clinic.Handler
	appointment  *appointment.Handler
	consultation *consultation.Handler
	health       *health.Handler
}

// registerRoutes registra todas as rotas no mux e retorna o handler com o stack global de middlewares aplicado.
func registerRoutes(mux *http.ServeMux, h handlers, cfg *config.Config, c cache.Cache, log logger.Logger) http.Handler {
	authMw := middleware.Auth(cfg.JWTSecret)
	apiKeyMw := middleware.APIKey(cfg.APIKey)
	inviteRL := middleware.RateLimit(c, 5, time.Minute)
	authRL := middleware.RateLimit(c, 10, time.Minute)
	generalRL := middleware.RateLimit(c, 120, time.Minute)
	reportRL := middleware.RateLimit(c, 30, time.Minute)

	protect := func(handler http.Handler) http.Handler {
		return generalRL(authMw(handler))
	}

	// Swagger — disponível apenas fora de produção
	if cfg.AppEnv != "production" {
		mux.Handle("GET /swagger/", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	// Observabilidade — sem autenticação (proteger por Security Group na AWS)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /health", h.health.Liveness)
	mux.HandleFunc("GET /health/ready", h.health.Readiness)

	// Auth — rotas públicas
	mux.Handle("POST /v1/api/auth/login", authRL(http.HandlerFunc(h.user.Login)))

	// Invites
	mux.Handle("POST /v1/api/invites", inviteRL(apiKeyMw(http.HandlerFunc(h.invite.Create))))
	mux.Handle("POST /v1/api/invites/accept", authRL(http.HandlerFunc(h.user.AcceptInvite)))

	// Users
	mux.Handle("POST /v1/api/users", protect(http.HandlerFunc(h.user.Create)))
	mux.Handle("GET /v1/api/users/{id}", protect(http.HandlerFunc(h.user.GetByID)))

	// Clinics
	mux.Handle("POST /v1/api/clinics", protect(http.HandlerFunc(h.clinic.Create)))
	mux.Handle("GET /v1/api/clinics", protect(http.HandlerFunc(h.clinic.List)))
	mux.Handle("GET /v1/api/clinics/{id}", protect(http.HandlerFunc(h.clinic.GetByID)))
	mux.Handle("DELETE /v1/api/clinics/{id}", protect(http.HandlerFunc(h.clinic.Delete)))

	// Appointments
	mux.Handle("POST /v1/api/appointments", protect(http.HandlerFunc(h.appointment.Create)))
	mux.Handle("GET /v1/api/appointments/patient/{patient_id}", protect(http.HandlerFunc(h.appointment.ListByPatient)))
	mux.Handle("PATCH /v1/api/appointments/{id}/cancel", protect(http.HandlerFunc(h.appointment.Cancel)))

	// Consultations — rate limit reduzido por ser rota de relatório
	mux.Handle("POST /v1/api/consultations", protect(http.HandlerFunc(h.consultation.Create)))
	mux.Handle("GET /v1/api/consultations/patient/{patient_id}", reportRL(authMw(http.HandlerFunc(h.consultation.ListByPatient))))
	mux.Handle("GET /v1/api/consultations/dentist/{dentist_id}", reportRL(authMw(http.HandlerFunc(h.consultation.ListByDentist))))

	// Stack global: RequestID → RequestLogger → SecurityHeaders → CORS → Metrics → rotas
	return middleware.RequestID(
		middleware.RequestLogger(log)(
			middleware.SecurityHeaders(cfg.AppEnv)(
				middleware.CORS(cfg.CORSAllowedOrigins)(
					middleware.Metrics(mux),
				),
			),
		),
	)
}
