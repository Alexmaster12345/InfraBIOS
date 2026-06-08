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

type ChangesHandler struct {
	db  *db.DB
	log *zap.Logger
}

func NewChangesHandler(database *db.DB, log *zap.Logger) *ChangesHandler {
	return &ChangesHandler{db: database, log: log}
}

// POST /api/v1/changes
func (h *ChangesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.RequestedBy == "" {
		response.Error(w, http.StatusUnprocessableEntity, "requested_by is required")
		return
	}
	if len(req.Payload) == 0 {
		req.Payload = json.RawMessage("{}")
	}
	c, err := h.db.CreateChange(r.Context(), uuid.NewString(), &req)
	if err != nil {
		h.log.Error("create change", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.log.Info("change request created", zap.String("id", c.ID), zap.String("type", string(c.Type)))
	response.JSON(w, http.StatusCreated, c)
}

// GET /api/v1/changes?status=
func (h *ChangesHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	changes, err := h.db.ListChanges(r.Context(), status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"changes": changes, "total": len(changes)})
}

// GET /api/v1/changes/{id}
func (h *ChangesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.db.GetChange(r.Context(), id)
	if err != nil || c == nil {
		response.Error(w, http.StatusNotFound, "change not found")
		return
	}
	response.JSON(w, http.StatusOK, c)
}

// POST /api/v1/changes/{id}/review
func (h *ChangesHandler) Review(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req models.ReviewChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.ReviewedBy == "" {
		response.Error(w, http.StatusUnprocessableEntity, "reviewed_by is required")
		return
	}
	if err := h.db.ReviewChange(r.Context(), id, &req); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	status := "rejected"
	if req.Approved {
		status = "approved"
	}
	h.log.Info("change reviewed", zap.String("id", id), zap.String("status", status))
	response.JSON(w, http.StatusOK, map[string]string{"id": id, "status": status})
}
