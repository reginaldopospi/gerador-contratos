package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"gerador-contratos/backend/internal/modules/common"
)

type Service struct {
	repo                  Repository
	tokens                *tokenManager
	accessTokenTTL        time.Duration
	refreshTokenTTL       time.Duration
	passwordResetTTL      time.Duration
	appEnv                string
	platformAdminUsername string
	platformAdminEmail    string
	registrationApproval  bool
	passwordResetNotifier PasswordResetNotifier
}

type ServiceConfig struct {
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	PasswordResetTTL      time.Duration
	JWTSecret             string
	AppEnv                string
	PlatformAdminUsername string
	PlatformAdminEmail    string
	RegistrationApproval  bool
	PasswordResetNotifier PasswordResetNotifier
}

func NewService(repo Repository, cfg ServiceConfig) *Service {
	notifier := cfg.PasswordResetNotifier
	if notifier == nil {
		// O notifier nulo evita quebrar o fluxo em ambientes sem SMTP configurado.
		notifier = NoopPasswordResetNotifier{}
	}
	adminUsername := strings.TrimSpace(strings.ToLower(cfg.PlatformAdminUsername))
	if adminUsername == "" {
		adminUsername = "admin"
	}

	return &Service{
		repo:                  repo,
		tokens:                newTokenManager(cfg.JWTSecret),
		accessTokenTTL:        cfg.AccessTokenTTL,
		refreshTokenTTL:       cfg.RefreshTokenTTL,
		passwordResetTTL:      cfg.PasswordResetTTL,
		appEnv:                cfg.AppEnv,
		platformAdminUsername: adminUsername,
		platformAdminEmail:    strings.TrimSpace(strings.ToLower(cfg.PlatformAdminEmail)),
		registrationApproval:  cfg.RegistrationApproval,
		passwordResetNotifier: notifier,
	}
}

func (s *Service) RegisterTenantAdmin(ctx context.Context, in RegisterTenantAdminInput, metadata ClientMetadata) (*RegisterTenantAdminResult, error) {
	in.TenantName = strings.TrimSpace(in.TenantName)
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	if in.TenantName == "" {
		return nil, common.NewBadRequest("invalid_tenant_name", "nome da imobiliaria e obrigatorio")
	}
	if in.Name == "" {
		return nil, common.NewBadRequest("invalid_name", "nome do usuario e obrigatorio")
	}
	if !isValidEmail(in.Email) {
		return nil, common.NewBadRequest("invalid_email", "email invalido")
	}
	if err := validatePassword(in.Password); err != nil {
		return nil, err
	}

	passwordHash, err := hashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	tenant := Tenant{
		ID:           uuid.NewString(),
		NomeFantasia: in.TenantName,
		CNPJ:         strings.TrimSpace(in.TenantCNPJ),
	}
	if err := s.repo.CreateTenant(ctx, tenant); err != nil {
		return nil, err
	}

	user := User{
		ID:           uuid.NewString(),
		TenantID:     tenant.ID,
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: passwordHash,
		Role:         RoleAdmin,
		// O admin cadastrado por autoatendimento pode ficar pendente ate aprovacao da plataforma.
		IsActive: !s.registrationApproval,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	if s.registrationApproval {
		user.PasswordHash = ""
		return &RegisterTenantAdminResult{
			User:            user,
			PendingApproval: true,
			Message:         "cadastro recebido e aguardando aprovacao administrativa",
		}, nil
	}

	authResult, err := s.newAuthResult(ctx, user, metadata)
	if err != nil {
		return nil, err
	}
	return &RegisterTenantAdminResult{
		User:            authResult.User,
		Tokens:          &authResult.Tokens,
		PendingApproval: false,
		Message:         "cadastro aprovado e acesso liberado",
	}, nil
}

func (s *Service) RegisterUser(ctx context.Context, actor AuthClaims, in RegisterUserInput) (*User, error) {
	if actor.Role != RoleAdmin && actor.Role != RoleGestor {
		return nil, common.NewForbidden("insufficient_permissions", "perfil sem permissao para criar usuarios")
	}

	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Name == "" {
		return nil, common.NewBadRequest("invalid_name", "nome e obrigatorio")
	}
	if !isValidEmail(in.Email) {
		return nil, common.NewBadRequest("invalid_email", "email invalido")
	}
	if err := validatePassword(in.Password); err != nil {
		return nil, err
	}
	if in.Role == "" {
		in.Role = RoleOperador
	}
	if !isValidRole(in.Role) {
		return nil, common.NewBadRequest("invalid_role", "role invalida")
	}

	hash, err := hashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	user := User{
		ID:           uuid.NewString(),
		TenantID:     actor.TenantID,
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: hash,
		Role:         in.Role,
		IsActive:     true,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return &user, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput, metadata ClientMetadata) (*AuthResult, error) {
	identifier := strings.TrimSpace(strings.ToLower(in.Email))
	// Keep password exactly as typed to stay consistent with register/reset hashing behavior.
	password := in.Password
	if identifier == "" || password == "" {
		return nil, common.NewUnauthorized("invalid_credentials", "email ou senha invalidos")
	}

	email := identifier
	if !isValidEmail(email) {
		// Permite login administrativo com username fixo (ex.: "admin") mapeando para o email configurado.
		if identifier == s.platformAdminUsername && s.platformAdminEmail != "" {
			email = s.platformAdminEmail
		} else {
			return nil, common.NewUnauthorized("invalid_credentials", "email ou senha invalidos")
		}
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, common.NewUnauthorized("invalid_credentials", "email ou senha invalidos")
	}
	if !user.IsActive {
		return nil, common.NewForbidden("inactive_user", "usuario inativo")
	}
	if !checkPassword(user.PasswordHash, password) {
		return nil, common.NewUnauthorized("invalid_credentials", "email ou senha invalidos")
	}

	return s.newAuthResult(ctx, *user, metadata)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, metadata ClientMetadata) (*AuthResult, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, common.NewUnauthorized("invalid_refresh_token", "refresh token invalido")
	}

	claims, err := s.tokens.parse(refreshToken)
	if err != nil || claims.Type != "refresh" {
		return nil, common.NewUnauthorized("invalid_refresh_token", "refresh token invalido")
	}

	hash := hashToken(refreshToken)
	session, err := s.repo.GetSessionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if session.RevokedAt != nil || session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, common.NewUnauthorized("expired_refresh_token", "refresh token expirado")
	}

	if err := s.repo.RevokeSession(ctx, session.ID); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, common.NewUnauthorized("invalid_refresh_token", "refresh token invalido")
	}

	return s.newAuthResult(ctx, *user, metadata)
}

func (s *Service) ForgotPassword(ctx context.Context, in ForgotPasswordInput) (string, error) {
	email := strings.TrimSpace(strings.ToLower(in.Email))
	if !isValidEmail(email) {
		return "", nil
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", nil
	}

	rawToken, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}

	entry := PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().UTC().Add(s.passwordResetTTL),
	}
	if err := s.repo.StorePasswordResetToken(ctx, entry); err != nil {
		return "", err
	}

	notification := PasswordResetNotification{
		ToEmail:   user.Email,
		Token:     rawToken,
		ExpiresAt: entry.ExpiresAt,
	}
	if err := s.passwordResetNotifier.SendPasswordReset(ctx, notification); err != nil {
		return "", common.NewInternal("password_reset_notification_failed", "nao foi possivel enviar email de recuperacao")
	}

	if strings.EqualFold(s.appEnv, "dev") {
		return rawToken, nil
	}
	return "", nil
}

func (s *Service) ResetPassword(ctx context.Context, in ResetPasswordInput) error {
	in.Token = strings.TrimSpace(in.Token)
	if in.Token == "" {
		return common.NewBadRequest("invalid_token", "token de recuperacao e obrigatorio")
	}
	if err := validatePassword(in.NewPassword); err != nil {
		return err
	}

	hash := hashToken(in.Token)
	entry, err := s.repo.GetPasswordResetTokenByHash(ctx, hash)
	if err != nil {
		return common.NewBadRequest("invalid_token", "token invalido ou expirado")
	}

	if entry.UsedAt != nil || entry.ExpiresAt.Before(time.Now().UTC()) {
		return common.NewBadRequest("invalid_token", "token invalido ou expirado")
	}

	newHash, err := hashPassword(in.NewPassword)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateUserPassword(ctx, entry.UserID, newHash); err != nil {
		return err
	}
	if err := s.repo.MarkPasswordResetTokenUsed(ctx, entry.ID); err != nil {
		return err
	}
	if err := s.repo.RevokeAllUserSessions(ctx, entry.UserID); err != nil {
		return err
	}

	return nil
}

// IsPlatformAdmin valida se o ator autenticado pode aprovar novos cadastros.
func (s *Service) IsPlatformAdmin(actor AuthClaims) bool {
	if s.platformAdminEmail == "" {
		return false
	}
	return actor.Role == RoleAdmin && strings.EqualFold(strings.TrimSpace(actor.Email), s.platformAdminEmail)
}

// ListPendingRegistrations retorna os cadastros de administradores aguardando aprovacao.
func (s *Service) ListPendingRegistrations(ctx context.Context, actor AuthClaims) ([]PendingRegistration, error) {
	if !s.IsPlatformAdmin(actor) {
		return nil, common.NewForbidden("insufficient_permissions", "somente admin da plataforma pode listar cadastros pendentes")
	}
	return s.repo.ListPendingTenantAdmins(ctx)
}

// ListTenants retorna as imobiliarias cadastradas para a area administrativa.
func (s *Service) ListTenants(ctx context.Context, actor AuthClaims) ([]TenantSummary, error) {
	if !s.IsPlatformAdmin(actor) {
		return nil, common.NewForbidden("insufficient_permissions", "somente admin da plataforma pode listar imobiliarias")
	}
	return s.repo.ListTenants(ctx)
}

// ApprovePendingRegistration ativa o cadastro e permite redefinir senha no momento da aprovacao.
func (s *Service) ApprovePendingRegistration(ctx context.Context, actor AuthClaims, in ApproveRegistrationInput) (*User, error) {
	if !s.IsPlatformAdmin(actor) {
		return nil, common.NewForbidden("insufficient_permissions", "somente admin da plataforma pode aprovar cadastros")
	}

	in.UserID = strings.TrimSpace(in.UserID)
	if in.UserID == "" {
		return nil, common.NewBadRequest("invalid_user_id", "id do usuario e obrigatorio")
	}

	user, err := s.repo.GetUserByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if user.Role != RoleAdmin {
		return nil, common.NewBadRequest("invalid_registration_type", "apenas cadastros de administrador podem ser aprovados")
	}
	if user.IsActive {
		return nil, common.NewConflict("already_approved", "cadastro ja aprovado")
	}

	if strings.TrimSpace(in.NewPassword) != "" {
		if err := validatePassword(in.NewPassword); err != nil {
			return nil, err
		}
		passwordHash, err := hashPassword(in.NewPassword)
		if err != nil {
			return nil, err
		}
		if err := s.repo.UpdateUserPassword(ctx, user.ID, passwordHash); err != nil {
			return nil, err
		}
	}

	if err := s.repo.SetUserActive(ctx, user.ID, true); err != nil {
		return nil, err
	}
	user.IsActive = true
	user.PasswordHash = ""
	return user, nil
}

// BootstrapPlatformAdmin garante a existencia do usuario administrativo da plataforma.
func (s *Service) BootstrapPlatformAdmin(ctx context.Context, in PlatformAdminBootstrapInput) error {
	in.TenantName = strings.TrimSpace(in.TenantName)
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))

	if in.TenantName == "" {
		in.TenantName = "Plataforma"
	}
	if in.Name == "" {
		in.Name = "Administrador da Plataforma"
	}
	if !isValidEmail(in.Email) {
		return common.NewBadRequest("invalid_platform_admin_email", "email do admin da plataforma e invalido")
	}
	if err := validatePassword(in.Password); err != nil {
		return err
	}
	if s.platformAdminEmail != "" && !strings.EqualFold(in.Email, s.platformAdminEmail) {
		return common.NewBadRequest("platform_admin_mismatch", "email informado nao corresponde ao admin da plataforma configurado")
	}

	passwordHash, err := hashPassword(in.Password)
	if err != nil {
		return err
	}

	existingUser, err := s.repo.GetUserByEmail(ctx, in.Email)
	if err == nil {
		if existingUser.Role != RoleAdmin {
			return common.NewConflict("platform_admin_invalid_role", "usuario do admin da plataforma precisa ter perfil admin")
		}
		if err := s.repo.UpdateUserPassword(ctx, existingUser.ID, passwordHash); err != nil {
			return err
		}
		if !existingUser.IsActive {
			if err := s.repo.SetUserActive(ctx, existingUser.ID, true); err != nil {
				return err
			}
		}
		return nil
	}
	if !isUserNotFoundError(err) {
		return err
	}

	tenant := Tenant{
		ID:           uuid.NewString(),
		NomeFantasia: in.TenantName,
	}
	if err := s.repo.CreateTenant(ctx, tenant); err != nil {
		return err
	}

	user := User{
		ID:           uuid.NewString(),
		TenantID:     tenant.ID,
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: passwordHash,
		Role:         RoleAdmin,
		IsActive:     true,
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) ValidateAccessToken(token string) (*AuthClaims, error) {
	claims, err := s.tokens.parse(strings.TrimSpace(token))
	if err != nil || claims.Type != "access" {
		return nil, common.NewUnauthorized("invalid_token", "token invalido")
	}
	return &AuthClaims{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Role:     Role(claims.Role),
		Email:    claims.Email,
	}, nil
}

func (s *Service) newAuthResult(ctx context.Context, user User, metadata ClientMetadata) (*AuthResult, error) {
	accessToken, accessExp, err := s.tokens.createAccessToken(user, s.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("create access token: %w", err)
	}
	refreshToken, refreshExp, err := s.tokens.createRefreshToken(user, s.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	session := Session{
		ID:               uuid.NewString(),
		UserID:           user.ID,
		RefreshTokenHash: hashToken(refreshToken),
		IP:               metadata.IP,
		UserAgent:        metadata.UserAgent,
		ExpiresAt:        refreshExp,
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return &AuthResult{
		User: user,
		Tokens: AuthTokens{
			AccessToken:      accessToken,
			RefreshToken:     refreshToken,
			AccessExpiresAt:  accessExp,
			RefreshExpiresAt: refreshExp,
		},
	}, nil
}

func validatePassword(password string) error {
	p := strings.TrimSpace(password)
	if len(p) < 8 {
		return common.NewBadRequest("weak_password", "senha deve ter no minimo 8 caracteres")
	}
	return nil
}

func isValidRole(role Role) bool {
	switch role {
	case RoleAdmin, RoleGestor, RoleOperador:
		return true
	default:
		return false
	}
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	return strings.Contains(email, "@") && len(email) >= 5
}

func isUserNotFoundError(err error) bool {
	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Code == "user_not_found"
}
