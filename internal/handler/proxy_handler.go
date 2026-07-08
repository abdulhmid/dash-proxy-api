package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/teodosiopiera/api-source-proxy/internal/dto"
	"github.com/teodosiopiera/api-source-proxy/internal/model"
	"github.com/teodosiopiera/api-source-proxy/internal/service"
	"github.com/teodosiopiera/api-source-proxy/pkg/response"
)

type ProxyHandler struct {
	proxyService *service.ProxyService
	userService  *service.UserService
}

func NewProxyHandler(proxyService *service.ProxyService, userService *service.UserService) *ProxyHandler {
	return &ProxyHandler{proxyService: proxyService, userService: userService}
}

func (h *ProxyHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	sourceName := chi.URLParam(r, "source")

	msisdn := r.FormValue("msisdn")
	if msisdn == "" {
		msisdn = r.URL.Query().Get("msisdn")
	}
	if msisdn == "" {
		var req dto.ProxyRequest
		if err := decodeJSON(r, &req); err != nil {
			response.Error(w, r, http.StatusBadRequest, "msisdn is required")
			return
		}
		msisdn = req.Msisdn
	}

	if msisdn == "" {
		response.Error(w, r, http.StatusBadRequest, "msisdn is required")
		return
	}

	user := GetUser(r.Context())
	apiKey := GetApiKey(r.Context())
	clientIP := h.proxyService.BuildProxyClientIP(r)
	client := r.Header.Get("X-Client")
	if client == "" {
		client = r.URL.Query().Get("client")
	}

	result, err := h.proxyService.ProxyRequest(r.Context(), sourceName, msisdn, user, apiKey, clientIP, client)
	if err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Success(w, r, "Proxy request successful", result)
}

func (h *ProxyHandler) ProxyTest(w http.ResponseWriter, r *http.Request) {
	sourceName := chi.URLParam(r, "source")

	var testReq dto.ProxyTestRequest
	if err := decodeJSON(r, &testReq); err != nil {
		testReq = dto.ProxyTestRequest{Params: map[string]interface{}{}}
	}

	claims := GetClaims(r.Context())
	if claims == nil {
		response.Error(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	user, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		response.Error(w, r, http.StatusInternalServerError, "User not found")
		return
	}

	clientIP := h.proxyService.BuildProxyClientIP(r)

	result, err := h.proxyService.ProxyRequestWithOverride(r.Context(), sourceName, user, clientIP, claims.Username, testReq.Method, testReq.Params)
	if err != nil {
		response.Success(w, r, "Proxy request completed", map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"result":  nil,
		})
		return
	}

	response.Success(w, r, "Proxy request successful", map[string]interface{}{
		"success": true,
		"error":   nil,
		"result":  result,
	})
}

func (h *ProxyHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.proxyService.GetApiSources(r.Context())
	if err != nil {
		response.Error(w, r, http.StatusInternalServerError, "Failed to list sources")
		return
	}
	response.Success(w, r, "Sources retrieved successfully", sources)
}

func (h *ProxyHandler) CreateSource(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateApiSourceRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		response.ValidationError(w, r, formatValidationErrors(err))
		return
	}

	name := req.Name
	if name == "" {
		name = extractNameFromURL(req.BaseURL)
	}

	authType := req.AuthType
	if authType == "" {
		authType = "custom"
	}
	method := req.Method
	if method == "" {
		method = "POST"
	}

	source := &model.ApiSource{
		Name:           name,
		BaseURL:        req.BaseURL,
		Username:       req.Username,
		AuthType:       authType,
		AuthHeaders:    req.AuthHeaders,
		ExtraParams:    req.ExtraParams,
		AcceptedFields: req.AcceptedFields,
		Method:         method,
		TimeoutMs:      req.TimeoutMs,
		IsActive:    true,
	}

	if err := h.proxyService.CreateApiSource(r.Context(), source); err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Created(w, r, "API source created successfully", source)
}

func (h *ProxyHandler) UpdateSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, r, http.StatusBadRequest, "Source ID is required")
		return
	}

	existing, err := h.proxyService.GetApiSourceByID(r.Context(), id)
	if err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	var req dto.UpdateApiSourceRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.BaseURL != "" {
		existing.BaseURL = req.BaseURL
	}
	if req.Username != "" {
		existing.Username = req.Username
	}
	if req.AuthType != "" {
		existing.AuthType = req.AuthType
	}
	if req.AuthHeaders != "" {
		existing.AuthHeaders = req.AuthHeaders
	}
	if req.ExtraParams != "" {
		existing.ExtraParams = req.ExtraParams
	}
	if req.AcceptedFields != "" {
		existing.AcceptedFields = req.AcceptedFields
	}
	if req.Method != "" {
		existing.Method = req.Method
	}
	if req.TimeoutMs != nil {
		existing.TimeoutMs = *req.TimeoutMs
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.proxyService.UpdateApiSource(r.Context(), existing); err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Success(w, r, "API source updated successfully", existing)
}

func (h *ProxyHandler) DeleteSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, r, http.StatusBadRequest, "Source ID is required")
		return
	}

	if err := h.proxyService.DeleteApiSource(r.Context(), id); err != nil {
		appErr := toAppError(err)
		response.Error(w, r, appErr.Code, appErr.Message)
		return
	}

	response.Success(w, r, "API source deleted successfully", nil)
}
