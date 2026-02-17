package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gerador-contratos/backend/internal/http/httpx"
	"gerador-contratos/backend/internal/http/middleware"
	"gerador-contratos/backend/internal/modules/brokers"
	"gerador-contratos/backend/internal/modules/common"
)

type BrokersHandler struct {
	service *brokers.Service
}

func NewBrokersHandler(service *brokers.Service) *BrokersHandler {
	return &BrokersHandler{service: service}
}

func (h *BrokersHandler) RegisterRoutes(r chi.Router, requireAuth func(http.Handler) http.Handler) {
	r.Group(func(sr chi.Router) {
		sr.Use(requireAuth)
		sr.Get("/", h.list)
		sr.Post("/", h.create)
		sr.Put("/{brokerID}", h.update)
		sr.Delete("/{brokerID}", h.delete)
	})
}

func (h *BrokersHandler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	result, err := h.service.List(r.Context(), claims)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": result})
}

func (h *BrokersHandler) create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	var input brokers.CreateBrokerInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	created, err := h.service.Create(r.Context(), claims, input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"broker": created})
}

func (h *BrokersHandler) update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	var input brokers.UpdateBrokerInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	updated, err := h.service.Update(r.Context(), claims, chi.URLParam(r, "brokerID"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"broker": updated})
}

func (h *BrokersHandler) delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	if err := h.service.Delete(r.Context(), claims, chi.URLParam(r, "brokerID")); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "corretor removido com sucesso"})
}
