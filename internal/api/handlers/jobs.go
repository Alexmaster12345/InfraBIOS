package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Alexmaster12345/infrabios/internal/api/response"
	"github.com/Alexmaster12345/infrabios/internal/db"
	"github.com/Alexmaster12345/infrabios/internal/models"
)

type JobsHandler struct {
	db  *db.DB
	log *zap.Logger
}

func NewJobsHandler(database *db.DB, log *zap.Logger) *JobsHandler {
	return &JobsHandler{db: database, log: log}
}

// POST /api/v1/jobs
func (h *JobsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Type == "" {
		response.Error(w, http.StatusUnprocessableEntity, "type is required")
		return
	}
	if len(req.Target) == 0 {
		req.Target = json.RawMessage("{}")
	}
	job, err := h.db.CreateJob(r.Context(), uuid.NewString(), &req)
	if err != nil {
		h.log.Error("create job", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.log.Info("job queued", zap.String("id", job.ID), zap.String("type", string(job.Type)))
	response.JSON(w, http.StatusCreated, job)
}

// GET /api/v1/jobs?status=
func (h *JobsHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	jobs, err := h.db.ListJobs(r.Context(), status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"jobs": jobs, "total": len(jobs)})
}

// GET /api/v1/jobs/{id}
func (h *JobsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.db.GetJob(r.Context(), id)
	if err != nil || job == nil {
		response.Error(w, http.StatusNotFound, "job not found")
		return
	}
	response.JSON(w, http.StatusOK, job)
}

// POST /api/v1/jobs/{id}/cancel
func (h *JobsHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.CancelJob(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.log.Info("job cancelled", zap.String("id", id))
	response.JSON(w, http.StatusOK, map[string]string{"cancelled": id})
}
