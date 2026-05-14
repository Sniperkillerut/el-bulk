package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/logger"
	"github.com/el-bulk/backend/utils/render"
)

// RefreshHandler handles on-demand and scheduled price refreshes.
type RefreshHandler struct {
	Service *service.RefreshService
	Pool    *service.WorkerPool
	Audit   service.Auditer
}

func NewRefreshHandler(s *service.RefreshService, p *service.WorkerPool, a service.Auditer) *RefreshHandler {
	return &RefreshHandler{Service: s, Pool: p, Audit: a}
}

// POST /api/admin/prices/refresh
func (h *RefreshHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering RefreshHandler.Trigger")

	// Submit background job
	job, err := h.Pool.JobService.CreateJob(r.Context(), "price_refresh", nil, nil)
	if err != nil {
		render.Error(w, fmt.Sprintf("failed to create job: %v", err), http.StatusInternalServerError)
		return
	}

	h.Pool.Submit(job)
	if h.Audit != nil {
		h.Audit.LogAction(r.Context(), "TRIGGER_PRICE_REFRESH", "job", job.ID, nil)
	}
	render.Success(w, map[string]interface{}{"status": "queued", "job_id": job.ID})
}

// POST /api/admin/currency/sync
func (h *RefreshHandler) SyncCurrency(w http.ResponseWriter, r *http.Request) {
	logger.TraceCtx(r.Context(), "Entering RefreshHandler.SyncCurrency")
	if err := h.Service.SyncCurrencyRates(r.Context()); err != nil {
		render.Error(w, fmt.Sprintf("failed to sync currency rates: %v", err), http.StatusInternalServerError)
		return
	}
	render.Success(w, map[string]string{"status": "ok"})
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

			logger.Info("[currency-sync] Starting scheduled midnight sync...")
			if err := svc.SyncCurrencyRates(context.Background()); err != nil {
				logger.Error("[currency-sync] Failed: %v", err)
			} else {
				logger.Info("[currency-sync] Done. Rates updated and prices materialized.")
			}

			logger.Info("[price-refresh] Starting scheduled midnight refresh...")
			updated, errs := svc.RunPriceRefresh(context.Background(), "", nil)
			logger.Info("[price-refresh] Done: %d updated, %d errors", updated, errs)
		}
	}()
}
