package shared

import "slices"

const (
	RoleAdmin     = "admin"
	RoleDentist   = "dentist"
	RoleSecretary = "secretary"
	RolePatient   = "patient"
)

func IsValidRole(role string) bool {
	permitedRoles := []string{RoleAdmin, RoleDentist, RoleSecretary, RolePatient}
	return slices.Contains(permitedRoles, role)
}

// except role patient
func IsValidCollaboratorRole(role string) bool {
	permitedRoles := []string{RoleAdmin, RoleDentist, RoleSecretary}

	return slices.Contains(permitedRoles, role)
}
