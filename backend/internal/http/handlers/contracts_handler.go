package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"gerador-contratos/backend/internal/http/httpx"
	"gerador-contratos/backend/internal/http/middleware"
	"gerador-contratos/backend/internal/modules/common"
	"gerador-contratos/backend/internal/modules/contracts"
)

type ContractsHandler struct {
	service *contracts.Service
}

func NewContractsHandler(service *contracts.Service) *ContractsHandler {
	return &ContractsHandler{service: service}
}

func (h *ContractsHandler) RegisterRoutes(r chi.Router, requireAuth func(http.Handler) http.Handler) {
	r.Group(func(sr chi.Router) {
		sr.Use(requireAuth)

		sr.Get("/", h.list)
		sr.Post("/", h.create)
		sr.Post("/preview", h.previewFromData)
		sr.Get("/{contractID}", h.getByID)
		sr.Get("/{contractID}/preview", h.previewLatest)
		sr.Post("/{contractID}/versions", h.addVersion)
		sr.Get("/{contractID}/versions/{versionNumber}", h.getVersion)
	})
}

func (h *ContractsHandler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	filter := contracts.ListContractsFilter{
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
		Query:  strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:  queryInt(r, "limit", 20),
		Offset: queryInt(r, "offset", 0),
	}

	items, err := h.service.ListContracts(r.Context(), claims, filter)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ContractsHandler) create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	var input contracts.CreateContractInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	result, err := h.service.CreateContract(r.Context(), claims, input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}

func (h *ContractsHandler) getByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	contractID := chi.URLParam(r, "contractID")
	result, err := h.service.GetContractDetails(r.Context(), claims, contractID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *ContractsHandler) addVersion(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	contractID := chi.URLParam(r, "contractID")
	var input contracts.AddVersionInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	version, err := h.service.AddVersion(r.Context(), claims, contractID, input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"version": version})
}

func (h *ContractsHandler) getVersion(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	contractID := chi.URLParam(r, "contractID")
	versionNumber := queryParamInt(chi.URLParam(r, "versionNumber"), 0)

	version, err := h.service.GetVersion(r.Context(), claims, contractID, versionNumber)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"version": version})
}

func (h *ContractsHandler) previewLatest(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		httpx.WriteError(w, common.NewUnauthorized("invalid_token", "token invalido"))
		return
	}

	contractID := chi.URLParam(r, "contractID")
	preview, err := h.service.PreviewLatest(r.Context(), claims, contractID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"preview": preview})
}

func (h *ContractsHandler) previewFromData(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Numero string         `json:"numero"`
		Tipo   string         `json:"tipo"`
		Data   map[string]any `json:"data"`
	}
	if err := httpx.DecodeJSON(r, &payload); err != nil {
		httpx.WriteError(w, common.NewBadRequest("invalid_payload", "payload invalido"))
		return
	}

	preview, err := h.service.PreviewFromData(payload.Numero, payload.Tipo, payload.Data)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"preview": preview})
}

func queryInt(r *http.Request, key string, fallback int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func queryParamInt(v string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return n
}
