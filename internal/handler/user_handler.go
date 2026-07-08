package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/teodosiopiera/api-source-proxy/internal/dto"
	"github.com/teodosiopiera/api-source-proxy/internal/service"
	"github.com/teodosiopiera/api-source-proxy/pkg/response"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		response.ValidationError(w, r, formatValidationErrors(err))
		return
	}

	user, err := h.userService.Create(r.Context(), req)
	if err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Created(w, r, "User created successfully", user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, r, http.StatusBadRequest, "User ID is required")
		return
	}

	var req dto.UpdateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		response.ValidationError(w, r, formatValidationErrors(err))
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), id, req)
	if err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Success(w, r, "User updated successfully", user)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		response.Error(w, r, http.StatusInternalServerError, "Failed to list users")
		return
	}

	response.Success(w, r, "Users retrieved successfully", users)
}
