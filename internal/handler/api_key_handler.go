package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"api-source-proxy/internal/dto"
	"api-source-proxy/internal/service"
	"api-source-proxy/pkg/response"
)

type ApiKeyHandler struct {
	userService *service.UserService
}

func NewApiKeyHandler(userService *service.UserService) *ApiKeyHandler {
	return &ApiKeyHandler{userService: userService}
}

func (h *ApiKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateApiKeyRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		response.ValidationError(w, r, formatValidationErrors(err))
		return
	}

	resp, err := h.userService.CreateApiKey(r.Context(), req)
	if err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Created(w, r, "API key created successfully", resp)
}

func (h *ApiKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.userService.ListApiKeys(r.Context())
	if err != nil {
		response.Error(w, r, http.StatusInternalServerError, "Failed to list API keys")
		return
	}

	response.Success(w, r, "API keys retrieved successfully", keys)
}

func (h *ApiKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, r, http.StatusBadRequest, "API key ID is required")
		return
	}

	if err := h.userService.RevokeApiKey(r.Context(), id); err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Success(w, r, "API key revoked successfully", nil)
}
