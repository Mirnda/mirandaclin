package collaborator

import (
	"github.com/Mirnda/mirandaclin/internal/domain/clinic"
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	profileblock "github.com/Mirnda/mirandaclin/internal/domain/profile_block"
	profileclinic "github.com/Mirnda/mirandaclin/internal/domain/profile_clinic"
)

type Collaborator struct {
	Profile             profile.Profile
	CollaboratorClinics []CollaboratorClinic
	ProfileBlocks       []profileblock.ProfileBlock
}

type CollaboratorClinic struct {
	Clinic        clinic.Clinic
	ProfileClinic profileclinic.ProfileClinic
}
