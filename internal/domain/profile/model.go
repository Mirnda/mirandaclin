package profile

import (
	"time"

	"github.com/google/uuid"
	"github.com/mirandev/mirandaclin/internal/domain/shared"
	"gorm.io/gorm"
)

type Profile struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"     json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	FullName  string         `json:"full_name"`
	Document  string         `json:"document"` // CPF
	BirthDate *time.Time     `json:"birth_date,omitempty"`
	Address   shared.Address `gorm:"embedded;embeddedPrefix:address_" json:"address"`
}

func (p *Profile) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
