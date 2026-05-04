package repository

import (
	"context"
	"errors"
	"time"

	emailverification "github.com/Mirnda/mirandaclin/internal/domain/email_verification"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type emailVerificationRepository struct{}

func NewEmailVerificationRepository() emailverification.Repository {
	return &emailVerificationRepository{}
}

func (r *emailVerificationRepository) Create(ctx context.Context, db *gorm.DB, ev *emailverification.EmailVerification) error {
	return db.WithContext(ctx).Create(ev).Error
}

func (r *emailVerificationRepository) FindByToken(ctx context.Context, db *gorm.DB, token string) (*emailverification.EmailVerification, error) {
	var ev emailverification.EmailVerification
	err := db.WithContext(ctx).Where("token = ?", token).First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ev, err
}

func (r *emailVerificationRepository) MarkUsed(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	now := time.Now()
	return db.WithContext(ctx).
		Model(&emailverification.EmailVerification{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now).Error
}
