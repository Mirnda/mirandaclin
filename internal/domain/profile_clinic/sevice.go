package profileclinic

import "github.com/google/uuid"

type CreateProfileClinicRequest struct {
	ProfileID           uuid.UUID
	ClinicID            uuid.UUID
	WorkingDays         []ShiftPerDay
	SlotDurationMinutes int
}
