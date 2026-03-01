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

func (f *fakeRepo) UpdateTenant(ctx context.Context, tenantID string, tenantName string, tenantCNPJ string) error {
	tenant, ok := f.tenants[tenantID]
	if !ok {
		return common.NewNotFound("tenant_not_found", "not found")
	}
	tenant.NomeFantasia = tenantName
	tenant.CNPJ = tenantCNPJ
	f.tenants[tenantID] = tenant
	return nil
}

func (f *fakeRepo) DeleteTenant(ctx context.Context, tenantID string) error {
	if _, ok := f.tenants[tenantID]; !ok {
		return common.NewNotFound("tenant_not_found", "not found")
	}
	delete(f.tenants, tenantID)

	// Simula o comportamento de cascata do SQLite para users/sessions/tokens.
	for userID, user := range f.users {
		if user.TenantID != tenantID {
			continue
		}
		delete(f.emailIx, user.Email)
		delete(f.users, userID)
	}
	return nil
}

func (f *fakeRepo) GetTenantByID(ctx context.Context, tenantID string) (*Tenant, error) {
	tenant, ok := f.tenants[tenantID]
	if !ok {
		return nil, common.NewNotFound("tenant_not_found", "not found")
	}
	copy := tenant
	return &copy, nil
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

func (f *fakeRepo) GetPrimaryTenantAdmin(ctx context.Context, tenantID string) (*User, error) {
	for _, user := range f.users {
		if user.TenantID == tenantID && user.Role == RoleAdmin {
			copy := user
			return &copy, nil
		}
	}
	return nil, common.NewNotFound("tenant_admin_not_found", "not found")
}

func (f *fakeRepo) ListTenants(ctx context.Context) ([]TenantSummary, error) {
	items := make([]TenantSummary, 0)
	for _, tenant := range f.tenants {
		summary := TenantSummary{
			TenantID:   tenant.ID,
			TenantName: tenant.NomeFantasia,
			TenantCNPJ: tenant.CNPJ,
			CreatedAt:  tenant.CreatedAt,
		}
		for _, user := range f.users {
			if user.TenantID != tenant.ID {
				continue
			}
			summary.TotalUsers++
			if user.IsActive {
				summary.ActiveUsers++
			}
			if summary.AdminEmail == "" && user.Role == RoleAdmin {
				summary.AdminEmail = user.Email
			}
		}
		items = append(items, summary)
	}
	return items, nil
}

func (f *fakeRepo) ListPendingTenantAdmins(ctx context.Context) ([]PendingRegistration, error) {
	items := make([]PendingRegistration, 0)
	for _, user := range f.users {
		if user.Role != RoleAdmin || user.IsActive {
			continue
		}
		tenant := f.tenants[user.TenantID]
		items = append(items, PendingRegistration{
			UserID:     user.ID,
			TenantID:   user.TenantID,
			TenantName: tenant.NomeFantasia,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
			IsActive:   user.IsActive,
			CreatedAt:  user.CreatedAt,
		})
	}
	return items, nil
}

func (f *fakeRepo) UpdateUserPassword(ctx context.Context, userID string, passwordHash string) error {
	u := f.users[userID]
	u.PasswordHash = passwordHash
	f.users[userID] = u
	return nil
}

func (f *fakeRepo) SetUserActive(ctx context.Context, userID string, isActive bool) error {
	u := f.users[userID]
	u.IsActive = isActive
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
	if result.Tokens == nil {
		t.Fatalf("expected tokens")
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

func TestRegisterTenantAdminPendingApprovalFlow(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: true,
	})

	result, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Pendente",
		Name:       "Novo Admin",
		Email:      "novo@teste.com",
		Password:   "senhaForte123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if !result.PendingApproval {
		t.Fatalf("expected pending approval")
	}
	if result.Tokens != nil {
		t.Fatalf("did not expect tokens when approval is required")
	}

	_, err = svc.Login(context.Background(), LoginInput{Email: "novo@teste.com", Password: "senhaForte123"}, ClientMetadata{})
	if err == nil {
		t.Fatalf("expected inactive user to fail login")
	}
}

func TestPlatformAdminApprovesPendingRegistrationWithNewPassword(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: true,
	})

	if err := svc.BootstrapPlatformAdmin(context.Background(), PlatformAdminBootstrapInput{
		TenantName: "Plataforma",
		Name:       "Administrador da Plataforma",
		Email:      "admin@plataforma.local",
		Password:   "Admin12345",
	}); err != nil {
		t.Fatalf("bootstrap platform admin failed: %v", err)
	}

	registerResult, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Pendente",
		Name:       "Novo Admin",
		Email:      "novo@teste.com",
		Password:   "senhaInicial123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	actor := AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@plataforma.local",
	}
	items, err := svc.ListPendingRegistrations(context.Background(), actor)
	if err != nil {
		t.Fatalf("list pending failed: %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected pending registrations")
	}

	approvedUser, err := svc.ApprovePendingRegistration(context.Background(), actor, ApproveRegistrationInput{
		UserID:      registerResult.User.ID,
		NewPassword: "NovaSenhaAprovada123",
	})
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if !approvedUser.IsActive {
		t.Fatalf("expected approved user to be active")
	}

	_, err = svc.Login(context.Background(), LoginInput{
		Email:    "novo@teste.com",
		Password: "NovaSenhaAprovada123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("login after approval failed: %v", err)
	}
}

func TestNonPlatformAdminCannotApproveRegistrations(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: true,
	})

	_, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Pendente",
		Name:       "Novo Admin",
		Email:      "novo@teste.com",
		Password:   "senhaInicial123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err = svc.ListPendingRegistrations(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@tenant.com",
	})
	if err == nil {
		t.Fatalf("expected forbidden list for non-platform admin")
	}
}

func TestPlatformAdminCanLoginWithUsernameAlias(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       7 * 24 * time.Hour,
		PasswordResetTTL:      30 * time.Minute,
		JWTSecret:             "test-secret",
		AppEnv:                "dev",
		PlatformAdminUsername: "admin",
		PlatformAdminEmail:    "admin@plataforma.local",
		RegistrationApproval:  true,
	})

	if err := svc.BootstrapPlatformAdmin(context.Background(), PlatformAdminBootstrapInput{
		TenantName: "Plataforma",
		Name:       "Administrador da Plataforma",
		Email:      "admin@plataforma.local",
		Password:   "Admin12345",
	}); err != nil {
		t.Fatalf("bootstrap platform admin failed: %v", err)
	}

	result, err := svc.Login(context.Background(), LoginInput{
		Email:    "admin",
		Password: "Admin12345",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("login with admin username failed: %v", err)
	}
	if result.User.Email != "admin@plataforma.local" {
		t.Fatalf("unexpected platform admin email: %s", result.User.Email)
	}
}

func TestPlatformAdminCanListTenants(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       7 * 24 * time.Hour,
		PasswordResetTTL:      30 * time.Minute,
		JWTSecret:             "test-secret",
		AppEnv:                "dev",
		PlatformAdminUsername: "admin",
		PlatformAdminEmail:    "admin@plataforma.local",
		RegistrationApproval:  true,
	})

	if err := svc.BootstrapPlatformAdmin(context.Background(), PlatformAdminBootstrapInput{
		TenantName: "Plataforma",
		Name:       "Administrador da Plataforma",
		Email:      "admin@plataforma.local",
		Password:   "Admin12345",
	}); err != nil {
		t.Fatalf("bootstrap platform admin failed: %v", err)
	}

	_, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Exemplo",
		Name:       "Admin Imob",
		Email:      "admin@imob.com",
		Password:   "senhaInicial123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register tenant admin failed: %v", err)
	}

	items, err := svc.ListTenants(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@plataforma.local",
	})
	if err != nil {
		t.Fatalf("list tenants failed: %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected at least one tenant")
	}
}

func TestNonPlatformAdminCannotListTenants(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:        15 * time.Minute,
		RefreshTokenTTL:       7 * 24 * time.Hour,
		PasswordResetTTL:      30 * time.Minute,
		JWTSecret:             "test-secret",
		AppEnv:                "dev",
		PlatformAdminUsername: "admin",
		PlatformAdminEmail:    "admin@plataforma.local",
		RegistrationApproval:  true,
	})

	_, err := svc.ListTenants(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@tenant.com",
	})
	if err == nil {
		t.Fatalf("expected forbidden error")
	}
}

func TestPlatformAdminCanUpdateTenant(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: true,
	})

	if err := svc.BootstrapPlatformAdmin(context.Background(), PlatformAdminBootstrapInput{
		TenantName: "Plataforma",
		Name:       "Administrador da Plataforma",
		Email:      "admin@plataforma.local",
		Password:   "Admin12345",
	}); err != nil {
		t.Fatalf("bootstrap platform admin failed: %v", err)
	}

	registerResult, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Editar",
		TenantCNPJ: "11.111.111/0001-11",
		Name:       "Admin Imob",
		Email:      "admin@editar.com",
		Password:   "senhaInicial123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register tenant admin failed: %v", err)
	}

	updatedTenant, err := svc.UpdateTenant(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@plataforma.local",
	}, UpdateTenantInput{
		TenantID:   registerResult.User.TenantID,
		TenantName: "Imobiliaria Editada",
		TenantCNPJ: "22.222.222/0001-22",
	})
	if err != nil {
		t.Fatalf("update tenant failed: %v", err)
	}
	if updatedTenant.NomeFantasia != "Imobiliaria Editada" {
		t.Fatalf("unexpected tenant name: %s", updatedTenant.NomeFantasia)
	}
	if updatedTenant.CNPJ != "22.222.222/0001-22" {
		t.Fatalf("unexpected tenant cnpj: %s", updatedTenant.CNPJ)
	}
}

func TestNonPlatformAdminCannotUpdateTenant(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: true,
	})

	_, err := svc.UpdateTenant(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@tenant.com",
	}, UpdateTenantInput{
		TenantID:   "tenant-x",
		TenantName: "Nome",
	})
	if err == nil {
		t.Fatalf("expected forbidden error")
	}
}

func TestPlatformAdminCanDeleteTenant(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: true,
	})

	if err := svc.BootstrapPlatformAdmin(context.Background(), PlatformAdminBootstrapInput{
		TenantName: "Plataforma",
		Name:       "Administrador da Plataforma",
		Email:      "admin@plataforma.local",
		Password:   "Admin12345",
	}); err != nil {
		t.Fatalf("bootstrap platform admin failed: %v", err)
	}

	registerResult, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Excluir",
		Name:       "Admin Excluir",
		Email:      "admin@excluir.com",
		Password:   "senhaInicial123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register tenant admin failed: %v", err)
	}

	if err := svc.DeleteTenant(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@plataforma.local",
	}, registerResult.User.TenantID); err != nil {
		t.Fatalf("delete tenant failed: %v", err)
	}

	if _, err := repo.GetTenantByID(context.Background(), registerResult.User.TenantID); err == nil {
		t.Fatalf("expected tenant to be deleted")
	}
}

func TestPlatformAdminCannotDeletePlatformTenant(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: true,
	})

	if err := svc.BootstrapPlatformAdmin(context.Background(), PlatformAdminBootstrapInput{
		TenantName: "Plataforma",
		Name:       "Administrador da Plataforma",
		Email:      "admin@plataforma.local",
		Password:   "Admin12345",
	}); err != nil {
		t.Fatalf("bootstrap platform admin failed: %v", err)
	}

	platformUser, err := repo.GetUserByEmail(context.Background(), "admin@plataforma.local")
	if err != nil {
		t.Fatalf("get platform admin failed: %v", err)
	}

	err = svc.DeleteTenant(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@plataforma.local",
	}, platformUser.TenantID)
	if err == nil {
		t.Fatalf("expected protected platform tenant error")
	}
}

func TestPlatformAdminCanResetTenantAdminPassword(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: false,
	})

	if err := svc.BootstrapPlatformAdmin(context.Background(), PlatformAdminBootstrapInput{
		TenantName: "Plataforma",
		Name:       "Administrador da Plataforma",
		Email:      "admin@plataforma.local",
		Password:   "Admin12345",
	}); err != nil {
		t.Fatalf("bootstrap platform admin failed: %v", err)
	}

	registerResult, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Senha",
		Name:       "Admin Senha",
		Email:      "admin@senha.com",
		Password:   "senhaInicial123",
	}, ClientMetadata{})
	if err != nil {
		t.Fatalf("register tenant admin failed: %v", err)
	}

	if _, err := svc.ResetTenantAdminPassword(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@plataforma.local",
	}, ResetTenantAdminPasswordInput{
		TenantID:    registerResult.User.TenantID,
		NewPassword: "NovaSenha12345",
	}); err != nil {
		t.Fatalf("reset tenant admin password failed: %v", err)
	}

	if _, err := svc.Login(context.Background(), LoginInput{
		Email:    "admin@senha.com",
		Password: "NovaSenha12345",
	}, ClientMetadata{}); err != nil {
		t.Fatalf("login with reset password failed: %v", err)
	}
}

func TestNonPlatformAdminCannotResetTenantAdminPassword(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
		JWTSecret:            "test-secret",
		AppEnv:               "dev",
		PlatformAdminEmail:   "admin@plataforma.local",
		RegistrationApproval: true,
	})

	_, err := svc.ResetTenantAdminPassword(context.Background(), AuthClaims{
		Role:  RoleAdmin,
		Email: "admin@tenant.com",
	}, ResetTenantAdminPasswordInput{
		TenantID:    "tenant-x",
		NewPassword: "NovaSenha12345",
	})
	if err == nil {
		t.Fatalf("expected forbidden error")
	}
}

func TestRegisterTenantAdminRollsBackTenantOnUserCreationConflict(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, ServiceConfig{
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
		PasswordResetTTL: 30 * time.Minute,
		JWTSecret:        "test-secret",
		AppEnv:           "dev",
	})

	if _, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Base",
		Name:       "Admin Base",
		Email:      "admin@base.com",
		Password:   "SenhaBase123",
	}, ClientMetadata{}); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	if _, err := svc.RegisterTenantAdmin(context.Background(), RegisterTenantAdminInput{
		TenantName: "Imobiliaria Orfa",
		Name:       "Admin Orfa",
		Email:      "admin@base.com",
		Password:   "SenhaOrfa123",
	}, ClientMetadata{}); err == nil {
		t.Fatalf("expected conflict on duplicated email")
	}

	// O rollback do tenant evita registros duplicados sem usuario administrador associado.
	if len(repo.tenants) != 1 {
		t.Fatalf("expected single tenant after rollback, got %d", len(repo.tenants))
	}
}
