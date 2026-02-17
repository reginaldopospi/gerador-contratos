package auth

import "context"

type Repository interface {
	CreateTenant(ctx context.Context, tenant Tenant) error
	CreateUser(ctx context.Context, user User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	UpdateUserPassword(ctx context.Context, userID string, passwordHash string) error

	CreateSession(ctx context.Context, session Session) error
	GetSessionByHash(ctx context.Context, refreshTokenHash string) (*Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error

	StorePasswordResetToken(ctx context.Context, token PasswordResetToken) error
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenID string) error
}
