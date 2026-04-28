package consultation

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Consultation struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"     json:"id"`
	TenantID      uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	AppointmentID uuid.UUID `gorm:"type:uuid;not null"       json:"appointment_id"`
	PatientID     uuid.UUID `gorm:"type:uuid;not null;index" json:"patient_id"`
	DentistID     uuid.UUID `gorm:"type:uuid;not null;index" json:"dentist_id"`
	Diagnosis     string    `json:"diagnosis"`
	Treatment     string    `json:"treatment"`
	CreatedAt     time.Time `json:"created_at"`
}

func (c *Consultation) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
