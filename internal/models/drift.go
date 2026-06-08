package models

import (
	"encoding/json"
	"time"
)

type DriftStatus string

const (
	DriftOpen     DriftStatus = "open"
	DriftResolved DriftStatus = "resolved"
	DriftIgnored  DriftStatus = "ignored"
)

type DriftEvent struct {
	ID            string          `json:"id"`
	ServerID      string          `json:"server_id"`
	SettingKey    string          `json:"setting_key"`
	ExpectedValue json.RawMessage `json:"expected_value"`
	ActualValue   json.RawMessage `json:"actual_value"`
	DetectedAt    time.Time       `json:"detected_at"`
	ResolvedAt    *time.Time      `json:"resolved_at,omitempty"`
	Status        DriftStatus     `json:"status"`
}
