package profile

import (
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Profile struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primaryKey"                                            json:"id"`
	UserID                *uuid.UUID     `gorm:"type:uuid"                                                       json:"user_id,omitempty"`
	TenantID              uuid.UUID      `gorm:"type:uuid;not null;index"                                        json:"tenant_id"`
	Role                  string         `gorm:"not null"                                                        json:"role"`
	FullName              string         `json:"full_name"`
	Document              string         `json:"document"`
	BirthDate             *time.Time     `json:"birth_date,omitempty"`
	Phone                 string         `json:"phone"`
	HasWhatsapp           bool           `gorm:"default:false"                                                   json:"has_whatsapp"`
	EmergencyContactName  string         `json:"emergency_contact_name"`
	EmergencyContactPhone string         `json:"emergency_contact_phone"`
	Address               shared.Address `gorm:"embedded;embeddedPrefix:address_"                                json:"address"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index"                                                           json:"-"`
}

func (p *Profile) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
