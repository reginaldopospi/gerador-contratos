package auth

import "context"

type Repository interface {
	CreateTenant(ctx context.Context, tenant Tenant) error
	UpdateTenant(ctx context.Context, tenantID string, tenantName string, tenantCNPJ string) error
	DeleteTenant(ctx context.Context, tenantID string) error
	GetTenantByID(ctx context.Context, tenantID string) (*Tenant, error)
	CreateUser(ctx context.Context, user User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetPrimaryTenantAdmin(ctx context.Context, tenantID string) (*User, error)
	ListTenants(ctx context.Context) ([]TenantSummary, error)
	ListPendingTenantAdmins(ctx context.Context) ([]PendingRegistration, error)
	UpdateUserPassword(ctx context.Context, userID string, passwordHash string) error
	SetUserActive(ctx context.Context, userID string, isActive bool) error

	CreateSession(ctx context.Context, session Session) error
	GetSessionByHash(ctx context.Context, refreshTokenHash string) (*Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	StorePasswordResetToken(ctx context.Context, token PasswordResetToken) error
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenID string) error
}
