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

type ProfilesHandler struct {
	db  *db.DB
	log *zap.Logger
}

func NewProfilesHandler(database *db.DB, log *zap.Logger) *ProfilesHandler {
	return &ProfilesHandler{db: database, log: log}
}

// POST /api/v1/profiles
func (h *ProfilesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		response.Error(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	if len(req.Settings) == 0 {
		req.Settings = json.RawMessage("{}")
	}
	p, err := h.db.CreateProfile(r.Context(), uuid.NewString(), &req)
	if err != nil {
		h.log.Error("create profile", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.log.Info("profile created", zap.String("id", p.ID), zap.String("name", p.Name))
	response.JSON(w, http.StatusCreated, p)
}

// GET /api/v1/profiles
func (h *ProfilesHandler) List(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.db.ListProfiles(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"profiles": profiles, "total": len(profiles)})
}

// GET /api/v1/profiles/{id}
func (h *ProfilesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	p, err := h.db.GetProfile(r.Context(), id)
	if err != nil || p == nil {
		response.Error(w, http.StatusNotFound, "profile not found")
		return
	}
	response.JSON(w, http.StatusOK, p)
}

// PUT /api/v1/profiles/{id}
func (h *ProfilesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	p, err := h.db.UpdateProfile(r.Context(), id, &req)
	if err != nil || p == nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, p)
}

// DELETE /api/v1/profiles/{id}
func (h *ProfilesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteProfile(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"deleted": id})
}
