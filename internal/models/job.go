package models

import (
	"encoding/json"
	"time"
)

type JobStatus string
type JobType string

const (
	JobQueued    JobStatus = "queued"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
	JobCancelled JobStatus = "cancelled"

	JobTypeComplianceScan  JobType = "compliance_scan"
	JobTypeApplyProfile    JobType = "apply_profile"
	JobTypeCollectInventory JobType = "collect_inventory"
	JobTypeRemediate       JobType = "remediate"
)

type Job struct {
	ID          string          `json:"id"`
	Type        JobType         `json:"type"`
	Target      json.RawMessage `json:"target"`
	Status      JobStatus       `json:"status"`
	Progress    int             `json:"progress"`
	Result      json.RawMessage `json:"result,omitempty"`
	CreatedBy   string          `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

type CreateJobRequest struct {
	Type      JobType         `json:"type"`
	Target    json.RawMessage `json:"target"`
	CreatedBy string          `json:"created_by"`
}

// JobTarget describes which servers a job applies to
type JobTarget struct {
	ServerIDs []string `json:"server_ids,omitempty"`
	ProfileID string   `json:"profile_id,omitempty"`
	All       bool     `json:"all,omitempty"`
}
