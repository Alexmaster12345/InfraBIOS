package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Alexmaster12345/infrabios/internal/api/response"
	"github.com/Alexmaster12345/infrabios/internal/db"
)

type DriftHandler struct {
	db  *db.DB
	log *zap.Logger
}

func NewDriftHandler(database *db.DB, log *zap.Logger) *DriftHandler {
	return &DriftHandler{db: database, log: log}
}

// GET /api/v1/drift?server_id=&status=
func (h *DriftHandler) List(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("server_id")
	status := r.URL.Query().Get("status")
	events, err := h.db.ListDriftEvents(r.Context(), serverID, status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"events": events, "total": len(events)})
}

// POST /api/v1/drift/{id}/resolve
func (h *DriftHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.ResolveDriftEvent(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"resolved": id})
}

// POST /api/v1/drift/{id}/ignore
func (h *DriftHandler) Ignore(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.IgnoreDriftEvent(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"ignored": id})
}
