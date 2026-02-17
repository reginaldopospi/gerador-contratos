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

	httpx.WriteJSON(w, http.StatusOK, result)
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

	httpx.WriteJSON(w, http.StatusOK, result)
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
		"user": map[string]any{
			"id":        claims.UserID,
			"tenant_id": claims.TenantID,
			"email":     claims.Email,
			"role":      claims.Role,
		},
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

func requestMetadata(r *http.Request) auth.ClientMetadata {
	return auth.ClientMetadata{
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}
}
