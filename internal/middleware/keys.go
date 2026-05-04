package middleware

import "github.com/google/uuid"

type contextKey string

const (
	keyTenantID  contextKey = "tenant_id"
	keyUserID    contextKey = "user_id"
	keyRole      contextKey = "role"
	keyScope     contextKey = "scope"
	keySessionID contextKey = "session_id"
)

// TenantFromContext extrai o tenant_id injetado pelo middleware de auth.
func TenantFromContext(ctx interface{ Value(any) any }) uuid.UUID {
	if v, ok := ctx.Value(keyTenantID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// UserIDFromContext extrai o user_id injetado pelo middleware de auth.
func UserIDFromContext(ctx interface{ Value(any) any }) uuid.UUID {
	if v, ok := ctx.Value(keyUserID).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}

// RoleFromContext extrai o role do usuário autenticado.
func RoleFromContext(ctx interface{ Value(any) any }) string {
	v, _ := ctx.Value(keyRole).(string)
	return v
}

// SessionIDFromContext extrai o session_id (jti do JWT) injetado pelo middleware de auth.
func SessionIDFromContext(ctx interface{ Value(any) any }) string {
	v, _ := ctx.Value(keySessionID).(string)
	return v
}
