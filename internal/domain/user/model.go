package user

import (
	"time"

	"github.com/Mirnda/mirandaclin/internal/domain/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleAdmin     = "admin"
	RoleDentist   = "dentist"
	RoleSecretary = "secretary"
	RolePatient   = "patient"
)

const (
	ScopeAdminAll     = "admin:*"
	ScopeDentistRead  = "dentist:read"
	ScopeDentistWrite = "dentist:write"
	ScopePatientRead  = "patient:read"
)

// User armazena apenas credenciais de autenticação. Dados pessoais ficam em Profile.
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email           string         `gorm:"not null;uniqueIndex" json:"email"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at,omitempty"`
	Salt            string         `json:"-"`
	PasswordHash    string         `json:"-"`
	LastTenantID    uuid.UUID      `json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserWithProfile é o shape de resposta para staff (User + Profile combinados).
type UserWithProfile struct {
	ID                    uuid.UUID      `json:"id"`
	TenantID              uuid.UUID      `json:"tenant_id"`
	Email                 string         `json:"email"`
	EmailVerifiedAt       *time.Time     `json:"email_verified_at,omitempty"`
	Role                  string         `json:"role"`
	FullName              string         `json:"full_name"`
	Document              string         `json:"document"`
	BirthDate             *time.Time     `json:"birth_date,omitempty"`
	Phone                 string         `json:"phone"`
	HasWhatsapp           bool           `json:"has_whatsapp"`
	EmergencyContactName  string         `json:"emergency_contact_name"`
	EmergencyContactPhone string         `json:"emergency_contact_phone"`
	Address               shared.Address `json:"address"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

// ScopeForRole retorna os escopos JWT correspondentes ao role do usuário.
func ScopeForRole(role string) string {
	switch role {
	case RoleAdmin:
		return ScopeAdminAll
	case RoleDentist:
		return ScopeDentistRead + " " + ScopeDentistWrite
	case RolePatient:
		return ScopePatientRead
	default:
		return ""
	}
}
