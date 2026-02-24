package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gerador-contratos/backend/internal/modules/common"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) CreateTenant(ctx context.Context, tenant Tenant) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenants (id, nome_fantasia, cnpj)
		VALUES (?, ?, ?)
	`, tenant.ID, tenant.NomeFantasia, nullIfEmpty(tenant.CNPJ))
	if err != nil {
		return fmt.Errorf("create tenant: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) CreateUser(ctx context.Context, user User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, name, email, password_hash, role, is_active)
		VALUES (?, ?, ?, LOWER(?), ?, ?, ?)
	`, user.ID, user.TenantID, user.Name, user.Email, user.PasswordHash, string(user.Role), boolToInt(user.IsActive))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return common.NewConflict("email_exists", "email ja cadastrado")
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		-- Case-insensitive lookup keeps login/forgot-password compatible with legacy records.
		WHERE LOWER(email) = LOWER(?)
	`, email)

	var user User
	var role string
	var isActive int
	if err := row.Scan(
		&user.ID,
		&user.TenantID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&role,
		&isActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewNotFound("user_not_found", "usuario nao encontrado")
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	user.Role = Role(role)
	user.IsActive = isActive == 1
	return &user, nil
}

func (r *SQLiteRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id)

	var user User
	var role string
	var isActive int
	if err := row.Scan(
		&user.ID,
		&user.TenantID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&role,
		&isActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewNotFound("user_not_found", "usuario nao encontrado")
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	user.Role = Role(role)
	user.IsActive = isActive == 1
	return &user, nil
}

func (r *SQLiteRepository) UpdateUserPassword(ctx context.Context, userID string, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) CreateSession(ctx context.Context, session Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, refresh_token_hash, ip, user_agent, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, session.ID, session.UserID, session.RefreshTokenHash, nullIfEmpty(session.IP), nullIfEmpty(session.UserAgent), session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetSessionByHash(ctx context.Context, refreshTokenHash string) (*Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, refresh_token_hash, COALESCE(ip, ''), COALESCE(user_agent, ''), expires_at, revoked_at, created_at
		FROM sessions
		WHERE refresh_token_hash = ?
		LIMIT 1
	`, refreshTokenHash)

	var s Session
	var revokedAt sql.NullTime
	if err := row.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.IP, &s.UserAgent, &s.ExpiresAt, &revokedAt, &s.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewUnauthorized("invalid_refresh_token", "refresh token invalido")
		}
		return nil, fmt.Errorf("get session by hash: %w", err)
	}
	if revokedAt.Valid {
		t := revokedAt.Time
		s.RevokedAt = &t
	}
	return &s, nil
}

func (r *SQLiteRepository) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = COALESCE(revoked_at, CURRENT_TIMESTAMP)
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return fmt.Errorf("revoke all sessions: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) StorePasswordResetToken(ctx context.Context, token PasswordResetToken) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, ?)
	`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("store password reset token: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = ?
		LIMIT 1
	`, tokenHash)

	var t PasswordResetToken
	var usedAt sql.NullTime
	if err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &usedAt, &t.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewNotFound("reset_token_not_found", "token nao encontrado")
		}
		return nil, fmt.Errorf("get password reset token: %w", err)
	}

	if usedAt.Valid {
		tm := usedAt.Time
		t.UsedAt = &tm
	}

	return &t, nil
}

func (r *SQLiteRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE password_reset_tokens
		SET used_at = COALESCE(used_at, CURRENT_TIMESTAMP)
		WHERE id = ?
	`, tokenID)
	if err != nil {
		return fmt.Errorf("mark reset token used: %w", err)
	}
	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}
