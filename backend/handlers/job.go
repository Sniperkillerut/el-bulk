package handlers

import (
	"net/http"
	"strconv"

	"github.com/el-bulk/backend/service"
	"github.com/el-bulk/backend/utils/render"
	"github.com/go-chi/chi/v5"
)

type JobHandler struct {
	Service *service.JobService
}

func NewJobHandler(s *service.JobService) *JobHandler {
	return &JobHandler{Service: s}
}

// GET /api/admin/jobs/{id}
func (h *JobHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.Service.GetJob(r.Context(), id)
	if err != nil {
		render.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	render.Success(w, job)
}

// GET /api/admin/jobs
func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 10
	}

	jobs, err := h.Service.ListRecent(r.Context(), limit)
	if err != nil {
		render.Error(w, "Failed to list jobs", http.StatusInternalServerError)
		return
	}
	render.Success(w, jobs)
}
