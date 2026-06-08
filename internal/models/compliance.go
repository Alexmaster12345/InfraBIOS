package models

import (
	"encoding/json"
	"time"
)

type ComplianceViolation struct {
	Key      string          `json:"key"`
	Expected json.RawMessage `json:"expected"`
	Actual   json.RawMessage `json:"actual"`
}

type ComplianceReport struct {
	ID          string                `json:"id"`
	ServerID    string                `json:"server_id"`
	ProfileID   string                `json:"profile_id"`
	Compliant   bool                  `json:"compliant"`
	Violations  []ComplianceViolation `json:"violations"`
	Score       float64               `json:"score"`
	GeneratedAt time.Time             `json:"generated_at"`
}
