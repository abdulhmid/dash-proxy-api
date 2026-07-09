package handler

import (
	"embed"
	"net/http"
)

//go:embed dashboard.html
var dashboardFS embed.FS

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) Serve(w http.ResponseWriter, r *http.Request) {
	data, err := dashboardFS.ReadFile("dashboard.html")
	if err != nil {
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
