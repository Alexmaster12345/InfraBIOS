package models

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	ID          string          `json:"id"`
	ServerID    string          `json:"server_id"`
	Settings    json.RawMessage `json:"settings"`
	Firmware    json.RawMessage `json:"firmware"`
	TakenBy     string          `json:"taken_by"`
	TakenAt     time.Time       `json:"taken_at"`
	Description string          `json:"description"`
}

type CreateSnapshotRequest struct {
	TakenBy     string `json:"taken_by"`
	Description string `json:"description"`
}
