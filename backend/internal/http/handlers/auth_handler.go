package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gerador-contratos/backend/internal/http/httpx"
	"gerador-contratos/backend/internal/http/middleware"
	"gerador-contratos/backend/internal/modules/auth"
	"gerador-contratos/backend/internal/modules/common"
)

type AuthHandler struct {
	service *auth.Service
}

func NewAuthHandler(service *auth.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router, requireAuth func(http.Handler) http.Handler) {
	r.Post("/register", h.registerTenantAdmin)
	r.Post("/login", h.login)
	r.Post("/refresh", h.refresh)
	r.Post("/forgot-password", h.forgotPassword)
	r.Post("/reset-password", h.resetPassword)

	r.Group(func(sr chi.Router) {
		sr.Use(requireAuth)
		sr.Get("/me", h.me)
		sr.Post("/users", h.registerUser)
		sr.Get("/tenants", h.listTenants)
		sr.Put("/tenants/{tenantID}", h.updateTenant)
		sr.Delete("/tenants/{tenantID}", h.deleteTenant)
		sr.Post("/tenants/{tenantID}/admin-password", h.resetTenantAdminPassword)
		sr.Get("/pending-registrations", h.listPendingRegistrations)
		sr.Post("/pending-registrations/{userID}/approve", h.approvePendingRegistration)
	})
}

func (h *AuthHandler) registerTenantAdmin(w http.ResponseWriter, r *http.Request) {
	var input auth.RegisterTenantAdminInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	result, err := h.service.RegisterTenantAdmin(r.Context(), input, requestMetadata(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	if result.Tokens != nil {
		// Mantem o payload consistente com login, incluindo o indicador de admin da plataforma.
		httpx.WriteJSON(w, http.StatusCreated, map[string]any{
			"user":             h.authUserPayload(auth.AuthClaims{UserID: result.User.ID, TenantID: result.User.TenantID, Role: result.User.Role, Email: result.User.Email}),
			"tokens":           result.Tokens,
			"pending_approval": result.PendingApproval,
			"message":          result.Message,
		})
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	result, err := h.service.Login(r.Context(), input, requestMetadata(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	h.writeAuthResult(w, http.StatusOK, result)
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	result, err := h.service.Refresh(r.Context(), body.RefreshToken, requestMetadata(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	h.writeAuthResult(w, http.StatusOK, result)
}

func (h *AuthHandler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var input auth.ForgotPasswordInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	devToken, err := h.service.ForgotPassword(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	resp := map[string]any{"message": "se o email existir, enviaremos instrucoes de recuperacao"}
	if devToken != "" {
		resp["dev_reset_token"] = devToken
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var input auth.ResetPasswordInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	if err := h.service.ResetPassword(r.Context(), input); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "senha redefinida com sucesso"})
}

func (h *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"user": h.authUserPayload(claims),
	})
}

func (h *AuthHandler) registerUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	var input auth.RegisterUserInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	user, err := h.service.RegisterUser(r.Context(), claims, input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (h *AuthHandler) listTenants(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	items, err := h.service.ListTenants(r.Context(), claims)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *AuthHandler) updateTenant(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	var body struct {
		TenantName string `json:"tenant_name"`
		TenantCNPJ string `json:"tenant_cnpj"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	tenantID := chi.URLParam(r, "tenantID")
	tenant, err := h.service.UpdateTenant(r.Context(), claims, auth.UpdateTenantInput{
		TenantID:   tenantID,
		TenantName: body.TenantName,
		TenantCNPJ: body.TenantCNPJ,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"tenant": tenant, "message": "imobiliaria atualizada com sucesso"})
}

func (h *AuthHandler) deleteTenant(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	tenantID := chi.URLParam(r, "tenantID")
	if err := h.service.DeleteTenant(r.Context(), claims, tenantID); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"message": "imobiliaria excluida com sucesso"})
}

func (h *AuthHandler) resetTenantAdminPassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	var body struct {
		NewPassword string `json:"new_password"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	tenantID := chi.URLParam(r, "tenantID")
	adminUser, err := h.service.ResetTenantAdminPassword(r.Context(), claims, auth.ResetTenantAdminPasswordInput{
		TenantID:    tenantID,
		NewPassword: body.NewPassword,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"user":    adminUser,
		"message": "senha do administrador redefinida com sucesso",
	})
}

func (h *AuthHandler) listPendingRegistrations(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	items, err := h.service.ListPendingRegistrations(r.Context(), claims)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *AuthHandler) approvePendingRegistration(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	userID := chi.URLParam(r, "userID")
	user, err := h.service.ApprovePendingRegistration(r.Context(), claims, auth.ApproveRegistrationInput{
		UserID: userID,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": user, "message": "cadastro aprovado com sucesso"})
}

func requestMetadata(r *http.Request) auth.ClientMetadata {
	return auth.ClientMetadata{
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}
}

func (h *AuthHandler) writeAuthResult(w http.ResponseWriter, status int, result *auth.AuthResult) {
	claims := auth.AuthClaims{
		UserID:   result.User.ID,
		TenantID: result.User.TenantID,
		Role:     result.User.Role,
		Email:    result.User.Email,
	}
	httpx.WriteJSON(w, status, map[string]any{
		"user":   h.authUserPayload(claims),
		"tokens": result.Tokens,
	})
}

func (h *AuthHandler) authUserPayload(claims auth.AuthClaims) map[string]any {
	return map[string]any{
		"id":                claims.UserID,
		"tenant_id":         claims.TenantID,
		"email":             claims.Email,
		"role":              claims.Role,
		"is_platform_admin": h.service.IsPlatformAdmin(claims),
	}
}
