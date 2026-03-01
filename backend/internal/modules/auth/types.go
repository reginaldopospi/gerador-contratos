package auth

import "time"

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleGestor   Role = "gestor"
	RoleOperador Role = "operador"
)

type Tenant struct {
	ID           string    `json:"id"`
	NomeFantasia string    `json:"nome_fantasia"`
	CNPJ         string    `json:"cnpj,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	RefreshTokenHash string     `json:"-"`
	IP               string     `json:"ip,omitempty"`
	UserAgent        string     `json:"user_agent,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type RegisterTenantAdminInput struct {
	TenantName string `json:"tenant_name"`
	TenantCNPJ string `json:"tenant_cnpj"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

// RegisterTenantAdminResult retorna o status do cadastro inicial da imobiliaria.
type RegisterTenantAdminResult struct {
	User            User        `json:"user"`
	Tokens          *AuthTokens `json:"tokens,omitempty"`
	PendingApproval bool        `json:"pending_approval"`
	Message         string      `json:"message"`
}

type RegisterUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     Role   `json:"role"`
}

// PendingRegistration representa um cadastro aguardando aprovacao da plataforma.
type PendingRegistration struct {
	UserID     string    `json:"user_id"`
	TenantID   string    `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Role       Role      `json:"role"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// TenantSummary apresenta uma imobiliaria cadastrada para o painel administrativo.
type TenantSummary struct {
	TenantID    string    `json:"tenant_id"`
	TenantName  string    `json:"tenant_name"`
	TenantCNPJ  string    `json:"tenant_cnpj,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	AdminUserID string    `json:"admin_user_id,omitempty"`
	AdminEmail  string    `json:"admin_email,omitempty"`
	TotalUsers  int       `json:"total_users"`
	ActiveUsers int       `json:"active_users"`
}

// UpdateTenantInput permite editar dados cadastrais da imobiliaria no painel administrativo.
type UpdateTenantInput struct {
	TenantID   string
	TenantName string `json:"tenant_name"`
	TenantCNPJ string `json:"tenant_cnpj"`
}

// ResetTenantAdminPasswordInput redefine a senha do admin principal da imobiliaria.
type ResetTenantAdminPasswordInput struct {
	TenantID    string
	NewPassword string `json:"new_password"`
}

// ApproveRegistrationInput aprova o cadastro sem alterar senha.
type ApproveRegistrationInput struct {
	UserID string
}

// PlatformAdminBootstrapInput define o usuario administrativo da plataforma.
type PlatformAdminBootstrapInput struct {
	TenantName string
	Name       string
	Email      string
	Password   string
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordInput struct {
	Email string `json:"email"`
}

type ResetPasswordInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type AuthTokens struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type AuthResult struct {
	User   User       `json:"user"`
	Tokens AuthTokens `json:"tokens"`
}

type AuthClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     Role   `json:"role"`
	Email    string `json:"email"`
}

type ClientMetadata struct {
	IP        string
	UserAgent string
}
