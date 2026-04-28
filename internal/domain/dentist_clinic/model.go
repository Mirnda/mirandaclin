package dentistclinic

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// DentistClinic vincula um dentista a uma clínica com horário de trabalho padrão.
// Unique constraint: (tenant_id, dentist_id, clinic_id).
type DentistClinic struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey"                                  json:"id"`
	TenantID            uuid.UUID      `gorm:"type:uuid;not null;index"                              json:"tenant_id"`
	DentistID           uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:udx_dentist_clinic,priority:1" json:"dentist_id"`
	ClinicID            uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:udx_dentist_clinic,priority:2" json:"clinic_id"`
	WorkingDays         pq.StringArray `gorm:"type:text[]"                                           json:"working_days"` // "monday"..."sunday"
	StartTime           string         `json:"start_time"`                                                                // "08:00"
	EndTime             string         `json:"end_time"`                                                                  // "17:00"
	SlotDurationMinutes int            `gorm:"default:30"                                            json:"slot_duration_minutes"`
	Active              bool           `gorm:"default:true"                                          json:"active"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

func (d *DentistClinic) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
