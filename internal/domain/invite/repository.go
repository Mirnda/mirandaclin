package invite

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, db *gorm.DB, inv *Invite) error
	FindByToken(ctx context.Context, db *gorm.DB, token string) (*Invite, error)
	MarkUsed(ctx context.Context, db *gorm.DB, id uuid.UUID) error
}
