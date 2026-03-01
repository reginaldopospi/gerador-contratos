package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gerador-contratos/backend/internal/http/httpx"
	"gerador-contratos/backend/internal/http/middleware"
	"gerador-contratos/backend/internal/modules/cnpj"
	"gerador-contratos/backend/internal/modules/common"
)

type CNPJHandler struct {
	service *cnpj.Service
}

func NewCNPJHandler(service *cnpj.Service) *CNPJHandler {
	return &CNPJHandler{service: service}
}

func (h *CNPJHandler) RegisterRoutes(r chi.Router, requireAuth func(http.Handler) http.Handler) {
	r.Group(func(sr chi.Router) {
		sr.Use(requireAuth)
		sr.Get("/{cnpj}", h.lookupByCNPJ)
	})
}

func (h *CNPJHandler) lookupByCNPJ(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	company, err := h.service.LookupByCNPJ(r.Context(), claims, chi.URLParam(r, "cnpj"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"company": company})
}
