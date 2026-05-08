package user

import (
	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/Mirnda/mirandaclin/internal/domain/tenant"
	"github.com/Mirnda/mirandaclin/pkg/config"
	"github.com/Mirnda/mirandaclin/pkg/mailer"
	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	userRepo    Repository
	profileRepo profile.Repository
	inviteRepo  invite.Repository
	tenantRepo  tenant.Repository
	cfg         config.App
	mailer      mailer.Mailer
	jwtSecret   string
}

func NewService(
	db *gorm.DB, ur Repository, pr profile.Repository,
	ir invite.Repository, tr tenant.Repository, cfg config.App,
	ml mailer.Mailer, secret string,
) *Service {
	return &Service{
		db:          db,
		userRepo:    ur,
		profileRepo: pr,
		inviteRepo:  ir,
		tenantRepo:  tr,
		cfg:         cfg,
		mailer:      ml,
		jwtSecret:   secret,
	}
}
