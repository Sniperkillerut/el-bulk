package handlers

import (
	"net/http"
	"strings"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

func (h *ProductHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	logger.TraceCtx(r.Context(), "Entering ProductHandler.GetRecommendations | IDs: %s", idsParam)

	if idsParam == "" {
		render.Success(w, []interface{}{})
		return
	}

	ids := strings.Split(idsParam, ",")
	recommendations, err := h.Service.GetRecommendations(r.Context(), ids)
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to get recommendations: %v", err)
		render.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
		return
	}

	render.Success(w, recommendations)
}
