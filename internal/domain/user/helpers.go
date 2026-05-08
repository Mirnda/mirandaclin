package user

import (
	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/google/uuid"
)

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func mergeUserProfile(u *User, p *profile.Profile) *UserWithProfile {
	return &UserWithProfile{
		ID:                    u.ID,
		TenantID:              p.TenantID,
		Email:                 u.Email,
		EmailVerifiedAt:       u.EmailVerifiedAt,
		Role:                  p.Role,
		FullName:              p.FullName,
		Document:              p.Document,
		BirthDate:             p.BirthDate,
		Phone:                 p.Phone,
		HasWhatsapp:           p.HasWhatsapp,
		EmergencyContactName:  p.EmergencyContactName,
		EmergencyContactPhone: p.EmergencyContactPhone,
		Address:               p.Address,
		CreatedAt:             u.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}
