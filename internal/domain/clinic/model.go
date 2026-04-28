package clinic

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mirandev/mirandaclin/internal/domain/shared"
	"gorm.io/gorm"
)

type Clinic struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"     json:"id"`
	TenantID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name          string         `gorm:"not null"                 json:"name"`
	Phone         string         `json:"phone"`
	Address       shared.Address `gorm:"embedded;embeddedPrefix:address_" json:"address"`
	OperatingDays pq.StringArray `gorm:"type:text[]"              json:"operating_days"` // "monday"..."sunday"
	OpenTime      string         `json:"open_time"`                                       // "08:00"
	CloseTime     string         `json:"close_time"`                                      // "18:00"
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                    json:"-"`
}

func (c *Clinic) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
