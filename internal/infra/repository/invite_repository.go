package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/invite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type inviteRepository struct{}

func NewInviteRepository() invite.Repository {
	return &inviteRepository{}
}

func (r *inviteRepository) Create(ctx context.Context, db *gorm.DB, inv *invite.Invite) error {
	return db.WithContext(ctx).Create(inv).Error
}

func (r *inviteRepository) FindByToken(ctx context.Context, db *gorm.DB, token string) (*invite.Invite, error) {
	var inv invite.Invite
	err := db.WithContext(ctx).Where("token = ?", token).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &inv, err
}

func (r *inviteRepository) MarkUsed(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	now := time.Now()
	return db.WithContext(ctx).
		Model(&invite.Invite{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now).Error
}
