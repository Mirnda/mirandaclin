package appointment

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusScheduled  = "scheduled"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

type Appointment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"     json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	PatientID   uuid.UUID  `gorm:"type:uuid;not null"       json:"patient_id"`
	DentistID   uuid.UUID  `gorm:"type:uuid;not null"       json:"dentist_id"`
	ClinicID    uuid.UUID  `gorm:"type:uuid;not null"       json:"clinic_id"`
	SecretaryID *uuid.UUID `gorm:"type:uuid"                json:"secretary_id,omitempty"`
	ScheduledAt time.Time  `gorm:"not null"                 json:"scheduled_at"`
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`
	Status      string     `gorm:"not null;default:scheduled" json:"status"`
	Notes       string     `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (a *Appointment) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
