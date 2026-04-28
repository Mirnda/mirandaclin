package middleware

import "github.com/google/uuid"

type contextKey string

const (
	keyRequestID contextKey = "request_id"
	keyTenantID  contextKey = "tenant_id"
	keyUserID    contextKey = "user_id"
	keyRole      contextKey = "role"
	keyScope     contextKey = "scope"
)

// TenantFromContext extrai o tenant_id injetado pelo middleware de tenant.
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

// RequestIDFromContext extrai o request_id gerado pelo middleware.
func RequestIDFromContext(ctx interface{ Value(any) any }) string {
	v, _ := ctx.Value(keyRequestID).(string)
	return v
}
