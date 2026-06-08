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

type FirmwareHandler struct {
	db  *db.DB
	log *zap.Logger
}

func NewFirmwareHandler(database *db.DB, log *zap.Logger) *FirmwareHandler {
	return &FirmwareHandler{db: database, log: log}
}

// GET /api/v1/servers/{id}/firmware
func (h *FirmwareHandler) List(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	fw, err := h.db.ListFirmware(r.Context(), serverID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"firmware": fw, "total": len(fw)})
}

// POST /api/v1/servers/{id}/firmware
func (h *FirmwareHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	var req models.UpsertFirmwareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	fw, err := h.db.UpsertFirmware(r.Context(), uuid.NewString(), serverID, &req)
	if err != nil {
		h.log.Error("upsert firmware", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, fw)
}

// POST /api/v1/servers/{id}/firmware/{component}/approve
func (h *FirmwareHandler) Approve(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	component := chi.URLParam(r, "component")
	var req struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.db.ApproveFirmware(r.Context(), serverID, component, req.Version); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"approved": component, "version": req.Version})
}
