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
	"github.com/Mirnda/mirandaclin/pkg/config"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
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
	if err := infraDB.Migrate(db); err != nil {
		log.Error("falha na migration", zap.Error(err))
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
	h := handlers{
		user:         user.NewHandler(userSvc),
		clinic:       clinic.NewHandler(clinicSvc),
		appointment:  appointment.NewHandler(apptSvc),
		consultation: consultation.NewHandler(consultSvc),
		health:       health.NewHandler(db, cache),
	}

	mux := http.NewServeMux()
	handler := registerRoutes(mux, h, cfg, cache)

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
