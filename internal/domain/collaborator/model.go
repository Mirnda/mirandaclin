package collaborator

import (
	"github.com/Mirnda/mirandaclin/internal/domain/clinic"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	profileblock "github.com/Mirnda/mirandaclin/internal/domain/profile_block"
	profileclinic "github.com/Mirnda/mirandaclin/internal/domain/profile_clinic"
)

type Collaborator struct {
	Profile             profile.Profile             `json:"profile"`
	CollaboratorClinics []CollaboratorClinic        `json:"collaborator_clinics"`
	ProfileBlocks       []profileblock.ProfileBlock `json:"profile_blocks"`
}

type CollaboratorClinic struct {
	Clinic        clinic.Clinic               `json:"clinic"`
	ProfileClinic profileclinic.ProfileClinic `json:"profile_clinic"`
}
