package profileclinic

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTenantRequired  = errors.New("informe o tenant_id")
	ErrProfileRequired = errors.New("informe o profile_id")
	ErrClinicRequired  = errors.New("informe o clinic_id")
)

// ProfileClinic vincula um profile a uma clínica com expediente customizado.
// Unique constraint: (tenant_id, profile_id, clinic_id).
type ProfileClinic struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey"                                         json:"id"`
	TenantID            uuid.UUID      `gorm:"type:uuid;not null;index"                                     json:"tenant_id"`
	ProfileID           uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:udx_profile_clinic,priority:1" json:"profile_id"`
	ClinicID            uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:udx_profile_clinic,priority:2" json:"clinic_id"`
	WorkingDays         []ShiftPerDay  `json:"working_days"` // "monday"..."sunday"
	SlotDurationMinutes int            `gorm:"default:30"                                                   json:"slot_duration_minutes"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index"                                                        json:"-"`
}

type ShiftPerDay struct {
	WeekDay   string `json:"week_day"`   // "monday"..."sunday"
	StartTime string `json:"start_time"` // "08:00"
	EndTime   string `json:"end_time"`   // "17:00"
}

func (pc *ProfileClinic) BeforeCreate(_ *gorm.DB) error {
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}
	if pc.TenantID == uuid.Nil {
		return ErrTenantRequired
	}
	if pc.ProfileID == uuid.Nil {
		return ErrProfileRequired
	}
	if pc.ClinicID == uuid.Nil {
		return ErrClinicRequired
	}
	return nil
}
