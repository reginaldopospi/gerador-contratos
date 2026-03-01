package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"gerador-contratos/backend/internal/modules/common"
)

type fakeRepo struct {
	tenants  map[string]Tenant
	users    map[string]User
	emailIx  map[string]string
	sessions map[string]Session
	resets   map[string]PasswordResetToken
}

type fakePasswordResetNotifier struct {
	err            error
	callCount      int
	notification   PasswordResetNotification
	lastContextErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		tenants:  map[string]Tenant{},
		users:    map[string]User{},
		emailIx:  map[string]string{},
		sessions: map[string]Session{},
		resets:   map[string]PasswordResetToken{},
	}
}

// SendPasswordReset registra a notificacao para validacao dos testes.
func (f *fakePasswordResetNotifier) SendPasswordReset(ctx context.Context, notification PasswordResetNotification) error {
	f.callCount++
	f.notification = notification
	f.lastContextErr = ctx.Err()
	return f.err
}

func (f *fakeRepo) CreateTenant(ctx context.Context, tenant Tenant) error {
	f.tenants[tenant.ID] = tenant
	return nil
}

func (f *fakeRepo) CreateUser(ctx context.Context, user User) error {
	if _, ok := f.emailIx[user.Email]; ok {
		return common.NewConflict("email_exists", "email ja cadastrado")
	}
	f.users[user.ID] = user
	f.emailIx[user.Email] = user.ID
	return nil
}

func (f *fakeRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	id, ok := f.emailIx[email]
	if !ok {
		return nil, common.NewNotFound("user_not_found", "not found")
	}
	u := f.users[id]
	return &u, nil
}

func (f *fakeRepo) GetUserByID(ctx context.Context, id string) (*User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, common.NewNotFound("user_not_found", "not found")
	}
	return &u, nil
}

func (f *fakeRepo) UpdateUserPassword(ctx context.Context, userID string, passwordHash string) error {
	u := f.users[userID]
	u.PasswordHash = passwordHash
	f.users[userID] = u
	return nil
}

func (f *fakeRepo) CreateSession(ctx context.Context, session Session) error {
	f.sessions[session.ID] = session
	return nil
}

func (f *fakeRepo) GetSessionByHash(ctx context.Context, refreshTokenHash string) (*Session, error) {
	for _, s := range f.sessions {
		if s.RefreshTokenHash == refreshTokenHash {
			copy := s
			return &copy, nil
		}
	}
	return nil, common.NewUnauthorized("invalid_refresh_token", "invalid")
}

func (f *fakeRepo) RevokeSession(ctx context.Context, sessionID string) error {
	s := f.sessions[sessionID]
	now := time.Now().UTC()
	s.RevokedAt = &now
	f.sessions[sessionID] = s
	return nil
}

func (f *fakeRepo) RevokeAllUserSessions(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	for k, s := range f.sessions {
		if s.UserID == userID {
			s.RevokedAt = &now
			f.sessions[k] = s
		}
	}
	return nil
}

func (f *fakeRepo) StorePasswordResetToken(ctx context.Context, token PasswordResetToken) error {
	f.resets[token.ID] = token
	return nil
}

func (f *fakeRepo) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	for _, t := range f.resets {
		if t.TokenHash == tokenHash {
			copy := t
			return &copy, nil
		}
	}
	return nil, common.NewNotFound("reset_token_not_found", "not found")
}

func (f *fakeRepo) MarkPasswordResetTokenUsed(ctx context.Context, tokenID string) error {
	t := f.resets[tokenID]
	now := time.Now().UTC()
	t.UsedAt = &now
	f.resets[tokenID] = t
	return nil
}

func TestRegisterAndLoginFlow(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
		PasswordResetTTL: 30 * time.Minute,
		JWTSecret:        "test-secret",
		AppEnv:           "dev",
	})

	result, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Teste",
		Name:       "Admin",
		Email:      "admin@teste.com",
		Password:   "senhaForte123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if result.User.Role != RoleAdmin {
		t.Fatalf("expected admin role, got %s", result.User.Role)
	}
	if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
		t.Fatalf("expected tokens")
	}

	loginResult, err := svc.Login(context.Background(), LoginInput{Email: "admin@teste.com", Password: "senhaForte123"}, ClientMetadata{})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResult.User.Email != "admin@teste.com" {
		t.Fatalf("unexpected user email: %s", loginResult.User.Email)
	}
}

func TestForgotAndResetPasswordFlow(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
		PasswordResetTTL: 30 * time.Minute,
		JWTSecret:        "test-secret",
		AppEnv:           "dev",
	})

	_, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Teste",
		Name:       "Admin",
		Email:      "admin@teste.com",
		Password:   "senhaForte123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	devToken, err := svc.ForgotPassword(context.Background(), ForgotPasswordInput{Email: "admin@teste.com"})
	if err != nil {
		t.Fatalf("forgot password failed: %v", err)
	}
	if devToken == "" {
		t.Fatalf("expected reset token in dev mode")
	}

	if err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:       devToken,
		NewPassword: "novaSenhaForte123",
	}); err != nil {
		t.Fatalf("reset password failed: %v", err)
	}

	_, err = svc.Login(context.Background(), LoginInput{Email: "admin@teste.com", Password: "novaSenhaForte123"}, ClientMetadata{})
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
}

func TestLoginWithPasswordContainingSpaces(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
		PasswordResetTTL: 30 * time.Minute,
		JWTSecret:        "test-secret",
		AppEnv:           "dev",
	})

	// This password intentionally keeps leading/trailing spaces.
	passwordWithSpaces := "  senhaComEspacos123  "
	_, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Teste",
		Name:       "Admin",
		Email:      "admin-space@teste.com",
		Password:   passwordWithSpaces,
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err = svc.Login(context.Background(), LoginInput{
		Email:    "admin-space@teste.com",
		Password: passwordWithSpaces,
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("login failed for password with spaces: %v", err)
	}
}

func TestForgotPasswordSendsNotificationInProd(t *testing.T) {
	repo := newFakeRepo()
	notifier := &fakePasswordResetNotifier{}
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       7 * 24 * time.Hour,
		PasswordResetTTL:      30 * time.Minute,
		JWTSecret:             "test-secret",
		AppEnv:                "prod",
		PasswordResetNotifier: notifier,
	})

	_, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Teste",
		Name:       "Admin",
		Email:      "admin@teste.com",
		Password:   "senhaForte123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	devToken, err := svc.ForgotPassword(context.Background(), ForgotPasswordInput{Email: "admin@teste.com"})
	if err != nil {
		t.Fatalf("forgot password failed: %v", err)
	}
	if devToken != "" {
		t.Fatalf("expected empty dev token in prod mode")
	}
	if notifier.callCount != 1 {
		t.Fatalf("expected notifier to be called once, got %d", notifier.callCount)
	}
	if notifier.notification.ToEmail != "admin@teste.com" {
		t.Fatalf("unexpected notification email: %s", notifier.notification.ToEmail)
	}
	if notifier.notification.Token == "" {
		t.Fatalf("expected notification token")
	}
}

func TestForgotPasswordReturnsErrorWhenNotificationFails(t *testing.T) {
	repo := newFakeRepo()
	notifier := &fakePasswordResetNotifier{err: errors.New("smtp down")}
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       7 * 24 * time.Hour,
		PasswordResetTTL:      30 * time.Minute,
		JWTSecret:             "test-secret",
		AppEnv:                "prod",
		PasswordResetNotifier: notifier,
	})

	_, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Teste",
		Name:       "Admin",
		Email:      "admin@teste.com",
		Password:   "senhaForte123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err = svc.ForgotPassword(context.Background(), ForgotPasswordInput{Email: "admin@teste.com"})
	if err == nil {
		t.Fatalf("expected forgot password error")
	}

	appErr, ok := err.(*common.AppError)
	if !ok {
		t.Fatalf("expected app error, got %T", err)
	}
	if appErr.Code != "password_reset_notification_failed" {
		t.Fatalf("unexpected error code: %s", appErr.Code)
	}
	if notifier.callCount != 1 {
		t.Fatalf("expected notifier call, got %d", notifier.callCount)
	}
}
