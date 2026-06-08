package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Alexmaster12345/infrabios/internal/db"
	"github.com/Alexmaster12345/infrabios/internal/models"
)

type ComplianceService struct {
	db *db.DB
}

func NewComplianceService(database *db.DB) *ComplianceService {
	return &ComplianceService{db: database}
}

// ScanServer compares the server's latest BIOS settings against its assigned profile.
// Returns nil if the server has no profile assigned.
func (s *ComplianceService) ScanServer(ctx context.Context, serverID string) (*models.ComplianceReport, error) {
	server, err := s.db.GetServer(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("get server: %w", err)
	}
	if server == nil {
		return nil, fmt.Errorf("server %s not found", serverID)
	}
	if server.ProfileID == nil {
		return nil, fmt.Errorf("server %s has no assigned profile", serverID)
	}

	profile, err := s.db.GetProfile(ctx, *server.ProfileID)
	if err != nil || profile == nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}

	latest, err := s.db.LatestSettings(ctx, serverID)
	if err != nil || latest == nil {
		return nil, fmt.Errorf("no BIOS settings collected for server %s", serverID)
	}

	report, err := s.compare(serverID, *server.ProfileID, profile.Settings, latest.Settings)
	if err != nil {
		return nil, err
	}

	if err := s.db.SaveComplianceReport(ctx, uuid.NewString(), report); err != nil {
		return nil, fmt.Errorf("save report: %w", err)
	}

	// Emit drift events for each violation
	for _, v := range report.Violations {
		ev := &models.DriftEvent{
			ServerID:      serverID,
			SettingKey:    v.Key,
			ExpectedValue: v.Expected,
			ActualValue:   v.Actual,
		}
		_ = s.db.CreateDriftEvent(ctx, uuid.NewString(), ev)
	}

	return report, nil
}

func (s *ComplianceService) compare(serverID, profileID string, profileSettings, actualSettings json.RawMessage) (*models.ComplianceReport, error) {
	var expected map[string]any
	var actual map[string]any
	if err := json.Unmarshal(profileSettings, &expected); err != nil {
		return nil, fmt.Errorf("parse profile settings: %w", err)
	}
	if err := json.Unmarshal(actualSettings, &actual); err != nil {
		return nil, fmt.Errorf("parse actual settings: %w", err)
	}

	var violations []models.ComplianceViolation
	for key, expVal := range expected {
		actVal, exists := actual[key]
		if !exists || fmt.Sprintf("%v", expVal) != fmt.Sprintf("%v", actVal) {
			expJSON, _ := json.Marshal(expVal)
			actJSON, _ := json.Marshal(actVal)
			violations = append(violations, models.ComplianceViolation{
				Key:      key,
				Expected: expJSON,
				Actual:   actJSON,
			})
		}
	}

	total := len(expected)
	passed := total - len(violations)
	score := 100.0
	if total > 0 {
		score = float64(passed) / float64(total) * 100
	}

	return &models.ComplianceReport{
		ServerID:    serverID,
		ProfileID:   profileID,
		Compliant:   len(violations) == 0,
		Violations:  violations,
		Score:       score,
		GeneratedAt: time.Now().UTC(),
	}, nil
}
