package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

type HealthHandler struct {
	Service *service.HealthService
}

func NewHealthHandler(s *service.HealthService) *HealthHandler {
	return &HealthHandler{Service: s}
}

func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !h.Service.IsAvailable() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "message": "database connection unavailable"})
		return
	}

	err := h.Service.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *HealthHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !h.Service.IsAvailable() {
		json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "message": "database connection unavailable"})
		return
	}

	stats, err := h.Service.GetStats()
	if err != nil {
		logger.Error("Failed to get stats: %v", err)
		render.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
