package profileblock

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProfileBlock registra um bloqueio pontual de agenda do profile.
// ClinicID nil → bloqueia em todas as clínicas. StartTime/EndTime nil → bloqueia o dia inteiro.
type ProfileBlock struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"     json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ProfileID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"profile_id"`
	ClinicID    *uuid.UUID `gorm:"type:uuid"                json:"clinic_id,omitempty"`
	BlockedDate time.Time  `gorm:"type:date;not null"       json:"blocked_date"`
	StartTime   *string    `json:"start_time,omitempty"` // "09:00" — nil = dia inteiro
	EndTime     *string    `json:"end_time,omitempty"`
	Reason      string     `json:"reason"`
	CreatedAt   time.Time  `json:"created_at"`
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`
}

func (d *ProfileBlock) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func RemoveDuplicates(blocks []ProfileBlock) []ProfileBlock {
	seen := make(map[string]bool)
	result := []ProfileBlock{}

	for _, pb := range blocks {
		if !seen[pb.ID.String()] {
			seen[pb.ID.String()] = true
			result = append(result, pb)
		}
	}

	return result
}
