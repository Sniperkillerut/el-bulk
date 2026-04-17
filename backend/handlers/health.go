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
	Version string
}

func NewHealthHandler(s *service.HealthService, version string) *HealthHandler {
	return &HealthHandler{Service: s, Version: version}
}

func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !h.Service.IsAvailable() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "degraded", "message": "database connection unavailable", "version": h.Version})
		return
	}

	err := h.Service.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "message": err.Error(), "version": h.Version})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "version": h.Version})
}

func (h *HealthHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !h.Service.IsAvailable() {
		json.NewEncoder(w).Encode(map[string]string{"status": "degraded", "message": "database connection unavailable"})
		return
	}

	stats, err := h.Service.GetStats(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to get stats: %v", err)
		render.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
