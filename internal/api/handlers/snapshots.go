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

type SnapshotsHandler struct {
	db  *db.DB
	log *zap.Logger
}

func NewSnapshotsHandler(database *db.DB, log *zap.Logger) *SnapshotsHandler {
	return &SnapshotsHandler{db: database, log: log}
}

// POST /api/v1/servers/{id}/snapshots
func (h *SnapshotsHandler) Create(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	var req models.CreateSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// pull latest settings + firmware as snapshot content
	settings, err := h.db.LatestSettings(r.Context(), serverID)
	if err != nil || settings == nil {
		response.Error(w, http.StatusBadRequest, "no settings collected — cannot snapshot")
		return
	}
	fw, _ := h.db.ListFirmware(r.Context(), serverID)
	fwJSON, _ := json.Marshal(fw)

	snap, err := h.db.CreateSnapshot(r.Context(), uuid.NewString(), serverID, &req, settings.Settings, fwJSON)
	if err != nil {
		h.log.Error("create snapshot", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.log.Info("snapshot taken", zap.String("id", snap.ID), zap.String("server", serverID))
	response.JSON(w, http.StatusCreated, snap)
}

// GET /api/v1/servers/{id}/snapshots
func (h *SnapshotsHandler) List(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	snaps, err := h.db.ListSnapshots(r.Context(), serverID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"snapshots": snaps, "total": len(snaps)})
}

// GET /api/v1/snapshots/{id}
func (h *SnapshotsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	snap, err := h.db.GetSnapshot(r.Context(), id)
	if err != nil || snap == nil {
		response.Error(w, http.StatusNotFound, "snapshot not found")
		return
	}
	response.JSON(w, http.StatusOK, snap)
}

// DELETE /api/v1/snapshots/{id}
func (h *SnapshotsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteSnapshot(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"deleted": id})
}
