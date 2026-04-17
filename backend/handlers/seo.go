package handlers

import (
	"net/http"

	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

type SeoHandler struct {
	Service *service.SeoService
}

func NewSeoHandler(s *service.SeoService) *SeoHandler {
	return &SeoHandler{Service: s}
}

func (h *SeoHandler) GetSitemapData(w http.ResponseWriter, r *http.Request) {
	data, err := h.Service.GetSitemapData(r.Context())
	if err != nil {
		logger.ErrorCtx(r.Context(), "Failed to get sitemap data: %v", err)
		render.Error(w, "Failed to collect SEO data", http.StatusInternalServerError)
		return
	}

	render.Success(w, data)
}
