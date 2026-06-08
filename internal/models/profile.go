package models

import (
	"encoding/json"
	"time"
)

type BIOSProfile struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Settings    json.RawMessage `json:"settings"`
	Tags        []string        `json:"tags"`
	CreatedBy   string          `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateProfileRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Settings    json.RawMessage `json:"settings"`
	Tags        []string        `json:"tags"`
	CreatedBy   string          `json:"created_by"`
}

type UpdateProfileRequest struct {
	Description string          `json:"description"`
	Settings    json.RawMessage `json:"settings"`
	Tags        []string        `json:"tags"`
}
