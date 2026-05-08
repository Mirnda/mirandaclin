package profileclinic

import "github.com/google/uuid"

type CreateProfileClinicRequest struct {
	ProfileID           uuid.UUID     `json:"profile_id"`
	ClinicID            uuid.UUID     `json:"clinic_id"`
	WorkingDays         []ShiftPerDay `json:"working_days"`
	SlotDurationMinutes int           `json:"slot_duration_minutes"`
}
