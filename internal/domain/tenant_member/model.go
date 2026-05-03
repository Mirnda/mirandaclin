package tenant_member

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantMember struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"                                    json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:udx_member,priority:1"   json:"user_id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:udx_member,priority:2"   json:"tenant_id"`
	Role      string    `gorm:"not null"                                                json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *TenantMember) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
