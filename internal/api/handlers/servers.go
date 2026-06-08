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
	"github.com/Alexmaster12345/infrabios/internal/service"
)

type ServersHandler struct {
	db    *db.DB
	drift *service.DriftService
	log   *zap.Logger
}

func NewServersHandler(database *db.DB, drift *service.DriftService, log *zap.Logger) *ServersHandler {
	return &ServersHandler{db: database, drift: drift, log: log}
}

// POST /api/v1/servers
func (h *ServersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Hostname == "" {
		response.Error(w, http.StatusUnprocessableEntity, "hostname is required")
		return
	}
	srv, err := h.db.CreateServer(r.Context(), uuid.NewString(), &req)
	if err != nil {
		h.log.Error("create server", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.log.Info("server registered", zap.String("id", srv.ID), zap.String("hostname", srv.Hostname))
	response.JSON(w, http.StatusCreated, srv)
}

// GET /api/v1/servers
func (h *ServersHandler) List(w http.ResponseWriter, r *http.Request) {
	servers, err := h.db.ListServers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"servers": servers, "total": len(servers)})
}

// GET /api/v1/servers/{id}
func (h *ServersHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	srv, err := h.db.GetServer(r.Context(), id)
	if err != nil || srv == nil {
		response.Error(w, http.StatusNotFound, "server not found")
		return
	}
	response.JSON(w, http.StatusOK, srv)
}

// PUT /api/v1/servers/{id}/profile
func (h *ServersHandler) AssignProfile(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	var req struct {
		ProfileID string `json:"profile_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.db.AssignProfile(r.Context(), serverID, req.ProfileID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"server_id": serverID, "profile_id": req.ProfileID})
}

// DELETE /api/v1/servers/{id}
func (h *ServersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteServer(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"deleted": id})
}

// POST /api/v1/servers/{id}/inventory  (agent endpoint)
func (h *ServersHandler) SubmitInventory(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	var req models.SubmitInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// detect drift before saving new settings
	go h.drift.Detect(r.Context(), serverID, req.Settings)

	settings, err := h.db.SaveSettings(r.Context(), uuid.NewString(), serverID, &req)
	if err != nil {
		h.log.Error("save settings", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.db.UpdateServerSeen(r.Context(), serverID, "", "")
	response.JSON(w, http.StatusCreated, settings)
}

// GET /api/v1/servers/{id}/settings
func (h *ServersHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	settings, err := h.db.LatestSettings(r.Context(), serverID)
	if err != nil || settings == nil {
		response.Error(w, http.StatusNotFound, "no settings collected for this server")
		return
	}
	response.JSON(w, http.StatusOK, settings)
}
