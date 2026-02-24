package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"gerador-contratos/backend/internal/modules/common"
)

type Service struct {
	repo             Repository
	tokens           *tokenManager
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	passwordResetTTL time.Duration
	appEnv           string
}

type ServiceConfig struct {
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	PasswordResetTTL time.Duration
	JWTSecret        string
	AppEnv           string
}

func NewService(repo Repository, cfg ServiceConfig) *Service {
	return &Service{
		repo:             repo,
		tokens:           newTokenManager(cfg.JWTSecret),
		accessTokenTTL:   cfg.AccessTokenTTL,
		refreshTokenTTL:  cfg.RefreshTokenTTL,
		passwordResetTTL: cfg.PasswordResetTTL,
		appEnv:           cfg.AppEnv,
	}
}

func (s *Service) RegisterTenantAdmin(ctx context.Context, in RegisterTenantAdminInput, metadata ClientMetadata) (*AuthResult, error) {
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
		IsActive:     true,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return s.newAuthResult(ctx, user, metadata)
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
	email := strings.TrimSpace(strings.ToLower(in.Email))
	// Keep password exactly as typed to stay consistent with register/reset hashing behavior.
	password := in.Password
	if !isValidEmail(email) || password == "" {
		return nil, common.NewUnauthorized("invalid_credentials", "email ou senha invalidos")
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
