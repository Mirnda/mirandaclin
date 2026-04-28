// @title          mirandaclin API
// @version        1.0
// @description    SaaS multi-tenant para clínicas odontológicas
// @host           localhost:8080
// @BasePath       /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/appointment"
	"github.com/Mirnda/mirandaclin/internal/domain/clinic"
	"github.com/Mirnda/mirandaclin/internal/domain/consultation"
	"github.com/Mirnda/mirandaclin/internal/domain/user"
	"github.com/Mirnda/mirandaclin/internal/health"
	infraCache "github.com/Mirnda/mirandaclin/internal/infra/cache"
	infraDB "github.com/Mirnda/mirandaclin/internal/infra/db"
	"github.com/Mirnda/mirandaclin/internal/infra/repository"
	"github.com/Mirnda/mirandaclin/internal/middleware"
	"github.com/Mirnda/mirandaclin/pkg/config"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	// modelos para AutoMigrate
	dentistblock "github.com/Mirnda/mirandaclin/internal/domain/dentist_block"
	dentistclinic "github.com/Mirnda/mirandaclin/internal/domain/dentist_clinic"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
)

func main() {
	_ = godotenv.Load() // silencioso — em ECS as vars vêm do ambiente

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("config inválida: %v", err))
	}

	log := logger.New(cfg.AppEnv)
	defer log.Sync()

	// Banco de dados
	db, err := infraDB.New(cfg.DSN())
	if err != nil {
		log.Error("falha ao conectar banco", zap.Error(err))
		panic(err)
	}
	if err := db.AutoMigrate(
		&user.User{},
		&profile.Profile{},
		&clinic.Clinic{},
		&dentistclinic.DentistClinic{},
		&dentistblock.DentistBlock{},
		&appointment.Appointment{},
		&consultation.Consultation{},
	); err != nil {
		log.Error("falha no AutoMigrate", zap.Error(err))
		panic(err)
	}

	// Cache
	cache, err := infraCache.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Warn("Redis indisponível — usando Noop cache", zap.Error(err))
		cache = infraCache.NewNoop()
	}

	// Repositórios
	userRepo := repository.NewUserRepository()
	profileRepo := repository.NewProfileRepository()
	clinicRepo := repository.NewClinicRepository()
	dcRepo := repository.NewDentistClinicRepository()
	blockRepo := repository.NewDentistBlockRepository()
	apptRepo := repository.NewAppointmentRepository()
	consultRepo := repository.NewConsultationRepository()

	// Services
	userSvc := user.NewService(db, userRepo, profileRepo, cache, cfg.JWTSecret, log)
	clinicSvc := clinic.NewService(db, clinicRepo, log)
	apptSvc := appointment.NewService(db, apptRepo, dcRepo, blockRepo, log)
	consultSvc := consultation.NewService(db, consultRepo, log)

	// Handlers
	userH := user.NewHandler(userSvc)
	clinicH := clinic.NewHandler(clinicSvc)
	apptH := appointment.NewHandler(apptSvc)
	consultH := consultation.NewHandler(consultSvc)
	healthH := health.NewHandler(db, cache)

	// Router
	mux := http.NewServeMux()

	// Observabilidade (sem autenticação — proteger por Security Group na AWS)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /health", healthH.Liveness)
	mux.HandleFunc("GET /health/ready", healthH.Readiness)

	// Auth (rate limit agressivo)
	authRL := middleware.RateLimit(cache, 10, time.Minute)
	mux.Handle("POST /v1/api/auth/login", authRL(http.HandlerFunc(userH.Login)))

	// Rotas autenticadas
	authMw := middleware.Auth(cfg.JWTSecret)
	generalRL := middleware.RateLimit(cache, 120, time.Minute)

	protect := func(h http.Handler) http.Handler {
		return generalRL(authMw(h))
	}

	// Users
	mux.Handle("POST /v1/api/users", protect(http.HandlerFunc(userH.Create)))
	mux.Handle("GET /v1/api/users/{id}", protect(http.HandlerFunc(userH.GetByID)))

	// Clinics
	mux.Handle("POST /v1/api/clinics", protect(http.HandlerFunc(clinicH.Create)))
	mux.Handle("GET /v1/api/clinics", protect(http.HandlerFunc(clinicH.List)))
	mux.Handle("GET /v1/api/clinics/{id}", protect(http.HandlerFunc(clinicH.GetByID)))
	mux.Handle("DELETE /v1/api/clinics/{id}", protect(http.HandlerFunc(clinicH.Delete)))

	// Appointments
	mux.Handle("POST /v1/api/appointments", protect(http.HandlerFunc(apptH.Create)))
	mux.Handle("GET /v1/api/appointments/patient/{patient_id}", protect(http.HandlerFunc(apptH.ListByPatient)))
	mux.Handle("PATCH /v1/api/appointments/{id}/cancel", protect(http.HandlerFunc(apptH.Cancel)))

	// Consultations
	reportRL := middleware.RateLimit(cache, 30, time.Minute)
	mux.Handle("POST /v1/api/consultations", protect(http.HandlerFunc(consultH.Create)))
	mux.Handle("GET /v1/api/consultations/patient/{patient_id}", reportRL(authMw(http.HandlerFunc(consultH.ListByPatient))))
	mux.Handle("GET /v1/api/consultations/dentist/{dentist_id}", reportRL(authMw(http.HandlerFunc(consultH.ListByDentist))))

	// Stack global de middlewares
	handler := middleware.RequestID(
		middleware.SecurityHeaders(cfg.AppEnv)(
			middleware.CORS(cfg.CORSAllowedOrigins)(
				middleware.Metrics(mux),
			),
		),
	)

	addr := ":" + cfg.AppPort
	log.Info("servidor iniciando", zap.String("addr", addr), zap.String("env", cfg.AppEnv))

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("servidor encerrado", zap.Error(err))
	}
}
