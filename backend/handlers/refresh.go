package handlers

import (
	"net/http"
	"time"

	"context"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

// RefreshHandler handles on-demand and scheduled price refreshes.
type RefreshHandler struct {
	Service *service.RefreshService
}

func NewRefreshHandler(s *service.RefreshService) *RefreshHandler {
	return &RefreshHandler{Service: s}
}

// POST /api/admin/prices/refresh
func (h *RefreshHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering RefreshHandler.Trigger")
	// Manual trigger from price menu syncs everything
	updated, errs := h.Service.RunPriceRefresh(r.Context(), "")
	render.Success(w, map[string]int{"updated": updated, "errors": errs})
}

// StartMidnightScheduler launches a goroutine that runs RunPriceRefresh
// once per day at midnight (server local time).
func StartMidnightScheduler(svc *service.RefreshService) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			sleepDur := time.Until(next)
			logger.Info("⏰ Next price refresh in %s (at %s)",
				sleepDur.Round(time.Minute), next.Format("2006-01-02 15:04"))
			time.Sleep(sleepDur)

			logger.Info("[price-refresh] Starting scheduled midnight refresh...")
			updated, errs := svc.RunPriceRefresh(context.Background(), "")
			logger.Info("[price-refresh] Done: %d updated, %d errors", updated, errs)
		}
	}()
}
