package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Alexmaster12345/infrabios/internal/api/response"
	"github.com/Alexmaster12345/infrabios/internal/db"
	"github.com/Alexmaster12345/infrabios/internal/service"
)

type ComplianceHandler struct {
	db          *db.DB
	compliance  *service.ComplianceService
	log         *zap.Logger
}

func NewComplianceHandler(database *db.DB, compliance *service.ComplianceService, log *zap.Logger) *ComplianceHandler {
	return &ComplianceHandler{db: database, compliance: compliance, log: log}
}

// POST /api/v1/servers/{id}/compliance/scan
func (h *ComplianceHandler) Scan(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	report, err := h.compliance.ScanServer(r.Context(), serverID)
	if err != nil {
		h.log.Error("compliance scan", zap.String("server", serverID), zap.Error(err))
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, report)
}

// GET /api/v1/servers/{id}/compliance
func (h *ComplianceHandler) Latest(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	report, err := h.db.LatestComplianceReport(r.Context(), serverID)
	if err != nil || report == nil {
		response.Error(w, http.StatusNotFound, "no compliance report found")
		return
	}
	response.JSON(w, http.StatusOK, report)
}

// GET /api/v1/servers/{id}/compliance/history
func (h *ComplianceHandler) History(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	reports, err := h.db.ListComplianceReports(r.Context(), serverID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"reports": reports, "total": len(reports)})
}
