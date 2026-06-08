package models

import (
	"encoding/json"
	"time"
)

type ChangeStatus string
type ChangeType string

const (
	ChangeStatusPending  ChangeStatus = "pending"
	ChangeStatusApproved ChangeStatus = "approved"
	ChangeStatusRejected ChangeStatus = "rejected"
	ChangeStatusApplied  ChangeStatus = "applied"
	ChangeStatusFailed   ChangeStatus = "failed"

	ChangeTypeApplyProfile ChangeType = "apply_profile"
	ChangeTypeSetValue     ChangeType = "set_value"
	ChangeTypeRemediate    ChangeType = "remediate"
)

type ChangeRequest struct {
	ID          string          `json:"id"`
	ServerID    *string         `json:"server_id,omitempty"`
	ProfileID   *string         `json:"profile_id,omitempty"`
	RequestedBy string          `json:"requested_by"`
	Type        ChangeType      `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Status      ChangeStatus    `json:"status"`
	RequestedAt time.Time       `json:"requested_at"`
	ReviewedBy  *string         `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time      `json:"reviewed_at,omitempty"`
	AppliedAt   *time.Time      `json:"applied_at,omitempty"`
}

type CreateChangeRequest struct {
	ServerID    *string         `json:"server_id"`
	ProfileID   *string         `json:"profile_id"`
	RequestedBy string          `json:"requested_by"`
	Type        ChangeType      `json:"type"`
	Payload     json.RawMessage `json:"payload"`
}

type ReviewChangeRequest struct {
	Approved   bool   `json:"approved"`
	ReviewedBy string `json:"reviewed_by"`
}
