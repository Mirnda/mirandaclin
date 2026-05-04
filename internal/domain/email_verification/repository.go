package email_verification

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, ev *EmailVerification) error
	FindByToken(ctx context.Context, db *gorm.DB, token string) (*EmailVerification, error)
	MarkUsed(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}
