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
	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/user"
	"github.com/Mirnda/mirandaclin/internal/health"
	infraCache "github.com/Mirnda/mirandaclin/internal/infra/cache"
	infraDB "github.com/Mirnda/mirandaclin/internal/infra/db"
	"github.com/Mirnda/mirandaclin/internal/infra/repository"
	"github.com/Mirnda/mirandaclin/pkg/config"
	"github.com/Mirnda/mirandaclin/pkg/logger"
	"github.com/Mirnda/mirandaclin/pkg/mailer"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // silencioso — em ECS as vars vêm do ambiente

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("config inválida: %v", err))
	}

	log := logger.New(cfg.AppEnv)
	defer func() { _ = log.Sync() }()

	// Banco de dados
	db, err := infraDB.New(cfg.DSN())
	if err != nil {
		log.Error("falha ao conectar banco", logger.Err(err))
		panic(err)
	}
	if err := infraDB.Migrate(db); err != nil {
		log.Error("falha na migration", logger.Err(err))
		panic(err)
	}

	// Cache
	cache, err := infraCache.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Warn("Redis indisponível — usando Noop cache", logger.Err(err))
		cache = infraCache.NewNoop()
	}

	// Repositórios
	userRepo := repository.NewUserRepository()
	inviteRepo := repository.NewInviteRepository()
	tenantRepo := repository.NewTenantRepository()
	memberRepo := repository.NewTenantMemberRepository()
	clinicRepo := repository.NewClinicRepository()
	dcRepo := repository.NewDentistClinicRepository()
	blockRepo := repository.NewDentistBlockRepository()
	apptRepo := repository.NewAppointmentRepository()
	consultRepo := repository.NewConsultationRepository()

	// Mailer
	var ml mailer.Mailer
	if cfg.SMTPHost != "" {
		ml = mailer.NewSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	} else {
		ml = mailer.NewNoop()
	}

	// Services
	userSvc := user.NewService(db, userRepo, inviteRepo, tenantRepo, memberRepo, cache, cfg.JWTSecret)
	inviteSvc := invite.NewService(db, inviteRepo, ml, cfg.AppURL)
	clinicSvc := clinic.NewService(db, clinicRepo)
	apptSvc := appointment.NewService(db, apptRepo, dcRepo, blockRepo)
	consultSvc := consultation.NewService(db, consultRepo)

	// Handlers
	h := handlers{
		user:         user.NewHandler(userSvc),
		invite:       invite.NewHandler(inviteSvc),
		clinic:       clinic.NewHandler(clinicSvc),
		appointment:  appointment.NewHandler(apptSvc),
		consultation: consultation.NewHandler(consultSvc),
		health:       health.NewHandler(db, cache),
	}

	mux := http.NewServeMux()
	handler := registerRoutes(mux, h, cfg, cache, log)

	addr := ":" + cfg.AppPort
	log.Info("servidor iniciando", logger.String("addr", addr), logger.String("env", cfg.AppEnv))

	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("servidor encerrado", logger.Err(err))
	}
}
