package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Alexmaster12345/infrabios/internal/db"
	"github.com/Alexmaster12345/infrabios/internal/models"
)

type DriftService struct {
	db  *db.DB
	log *zap.Logger
}

func NewDriftService(database *db.DB, log *zap.Logger) *DriftService {
	return &DriftService{db: database, log: log}
}

// Detect compares the newly submitted settings to the previous snapshot.
// Any key that changed emits a drift event.
func (s *DriftService) Detect(ctx context.Context, serverID string, newSettings json.RawMessage) {
	prev, err := s.db.LatestSettings(ctx, serverID)
	if err != nil || prev == nil {
		return // no baseline — nothing to compare
	}

	var before, after map[string]any
	if err := json.Unmarshal(prev.Settings, &before); err != nil {
		return
	}
	if err := json.Unmarshal(newSettings, &after); err != nil {
		return
	}

	for key, oldVal := range before {
		newVal, exists := after[key]
		if !exists || fmt.Sprintf("%v", oldVal) != fmt.Sprintf("%v", newVal) {
			oldJSON, _ := json.Marshal(oldVal)
			newJSON, _ := json.Marshal(newVal)
			ev := &models.DriftEvent{
				ServerID:      serverID,
				SettingKey:    key,
				ExpectedValue: oldJSON,
				ActualValue:   newJSON,
			}
			if err := s.db.CreateDriftEvent(ctx, uuid.NewString(), ev); err != nil {
				s.log.Error("failed to record drift event", zap.Error(err))
			} else {
				s.log.Warn("drift detected",
					zap.String("server", serverID),
					zap.String("key", key),
					zap.Any("old", oldVal),
					zap.Any("new", newVal),
				)
			}
		}
	}
}
