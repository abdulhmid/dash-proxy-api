package handler

import (
	"net/http"

	"api-source-proxy/app/dto"
	"api-source-proxy/app/service"
	"api-source-proxy/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		response.ValidationError(w, r, formatValidationErrors(err))
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Success(w, r, "Login successful", resp)
}
