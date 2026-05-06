package user

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Mirnda/mirandaclin/internal/domain/profile"
	"github.com/google/uuid"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

// normalizeForSearch converte uma string para minúsculas e remove acentos,
// permitindo busca por substring insensível a maiúsculas e acentos.
// Ex: "Alê" → "ale", "ALEXANDRE" → "alexandre", "Dalessandro" → "dalessandro"
func normalizeForSearch(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	result, _, _ := transform.String(t, s)
	return strings.ToLower(result)
}

func patientsCacheKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s:patients", tenantID)
}
