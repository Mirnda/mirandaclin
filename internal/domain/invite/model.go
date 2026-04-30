package invite

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Invite struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"      json:"id"`
	TenantID     uuid.UUID  `gorm:"type:uuid;not null;index"  json:"tenant_id"`
	Token        string     `gorm:"not null;uniqueIndex"      json:"token"`
	Email        string     `gorm:"not null"                  json:"email"`
	Role         string     `gorm:"not null"                  json:"role"`
	PasswordHash string     `gorm:"not null"                  json:"-"`
	Salt         string     `gorm:"not null"                  json:"-"`
	EventId      string     `                                 json:"event_id,omitempty"`
	UsedAt       *time.Time `                                 json:"used_at,omitempty"`
	ExpiresAt    time.Time  `gorm:"not null"                  json:"expires_at"`
	CreatedAt    time.Time  `                                 json:"created_at"`
}

func (i *Invite) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
