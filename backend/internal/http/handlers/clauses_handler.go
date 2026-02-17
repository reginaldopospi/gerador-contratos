package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gerador-contratos/backend/internal/http/httpx"
	"gerador-contratos/backend/internal/http/middleware"
	"gerador-contratos/backend/internal/modules/clauses"
	"gerador-contratos/backend/internal/modules/common"
)

type ClausesHandler struct {
	service *clauses.Service
}

func NewClausesHandler(service *clauses.Service) *ClausesHandler {
	return &ClausesHandler{service: service}
}

func (h *ClausesHandler) RegisterRoutes(r chi.Router, requireAuth func(http.Handler) http.Handler) {
	r.Group(func(sr chi.Router) {
		sr.Use(requireAuth)
		sr.Get("/", h.list)
		sr.Post("/", h.upsert)
		sr.Delete("/{clauseID}", h.delete)
	})
}

func (h *ClausesHandler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	items, err := h.service.List(r.Context(), claims)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ClausesHandler) upsert(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	var input clauses.UpsertClauseInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	item, err := h.service.Upsert(r.Context(), claims, input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"clause": item})
}

func (h *ClausesHandler) delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	if err := h.service.Delete(r.Context(), claims, chi.URLParam(r, "clauseID")); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "clausula removida com sucesso"})
}
