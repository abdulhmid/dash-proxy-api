package handler

import (
	"net/http"
	"strconv"

	"github.com/teodosiopiera/api-source-proxy/internal/service"
	"github.com/teodosiopiera/api-source-proxy/pkg/response"
)

type LogHandler struct {
	proxyService *service.ProxyService
}

func NewLogHandler(proxyService *service.ProxyService) *LogHandler {
	return &LogHandler{proxyService: proxyService}
}

func (h *LogHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := h.proxyService.ProcessLogFilter(
		r.URL.Query().Get("user_id"),
		r.URL.Query().Get("api_source_name"),
		r.URL.Query().Get("start_date"),
		r.URL.Query().Get("end_date"),
	)

	logs, total, err := h.proxyService.GetLogs(r.Context(), filter, page, limit)
	if err != nil {
		response.Error(w, r, http.StatusInternalServerError, "Failed to retrieve logs")
		return
	}

	response.Paginated(w, r, "Logs retrieved successfully", logs, page, limit, int(total))
}

func (h *LogHandler) UserList(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r.Context())
	if claims == nil {
		response.Error(w, r, http.StatusUnauthorized, "Not authenticated")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := h.proxyService.ProcessLogFilter(
		claims.UserID,
		r.URL.Query().Get("api_source_name"),
		r.URL.Query().Get("start_date"),
		r.URL.Query().Get("end_date"),
	)

	logs, total, err := h.proxyService.GetLogs(r.Context(), filter, page, limit)
	if err != nil {
		response.Error(w, r, http.StatusInternalServerError, "Failed to retrieve logs")
		return
	}

	response.Paginated(w, r, "Logs retrieved successfully", logs, page, limit, int(total))
}
