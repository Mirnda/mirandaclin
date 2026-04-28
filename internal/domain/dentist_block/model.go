package dentistblock

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DentistBlock registra um bloqueio pontual de agenda do dentista.
// ClinicID nil → bloqueia em todas as clínicas. StartTime/EndTime nil → bloqueia o dia inteiro.
type DentistBlock struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"     json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	DentistID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"dentist_id"`
	ClinicID    *uuid.UUID `gorm:"type:uuid"                json:"clinic_id,omitempty"`
	BlockedDate time.Time  `gorm:"type:date;not null"       json:"blocked_date"`
	StartTime   *string    `json:"start_time,omitempty"` // "09:00" — nil = dia inteiro
	EndTime     *string    `json:"end_time,omitempty"`
	Reason      string     `json:"reason"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (d *DentistBlock) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
